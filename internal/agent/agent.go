// Package agent keeps the link with Odoo: it pairs the device, reports what it finds and
// handles the print jobs and commands it receives.
//
// Every call starts here, so it works the same with Odoo on the same network or on a remote
// server: no ports to open and no tunnels in the customer network.
package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/apiservicesac/dooprint/internal/escpos"
	"github.com/apiservicesac/dooprint/internal/logger"
	"github.com/apiservicesac/dooprint/internal/printer"
)

const (
	// HeartbeatInterval is how often the device reports it is alive and refreshes its printers.
	HeartbeatInterval = 30 * time.Second
	requestTimeout    = 30 * time.Second
	jobsPerRequest    = 5
	reconnectDelay    = 5 * time.Second
	forwardTimeout    = 15 * time.Second
	// forwardMaxBody caps the answer of a forwarded request sent back to Odoo.
	forwardMaxBody = 1 << 20
)

// Printers returns the detected printers as they are reported to Odoo.
type Printers func() []map[string]any

// Status is what the web interface shows about the Odoo link.
type Status struct {
	Paired    bool      `json:"paired"`
	OdooURL   string    `json:"odooUrl"`
	BoxName   string    `json:"boxName"`
	Mode      string    `json:"mode"`
	BusState  string    `json:"busState"`
	LastPoll  time.Time `json:"lastPoll"`
	LastError string    `json:"lastError"`
	Printed   int       `json:"printed"`
	Failed    int       `json:"failed"`
}

type Config struct {
	OdooURL string
	Token   string
	Mode    string
	Name    string
}

type Agent struct {
	config      Config
	manager     *printer.Manager
	printers    Printers
	client      *http.Client
	mu          sync.Mutex
	status      Status
	wake        chan struct{}
	busFallback string
	// Restart is called by the restart command sent from Odoo.
	Restart func()
}

type job struct {
	Id      int    `json:"id"`
	Printer string `json:"printer"`
	Payload string `json:"payload"`
}

type command struct {
	Id      int         `json:"id"`
	Name    string      `json:"name"`
	Payload httpRequest `json:"payload"`
}

// httpRequest is what Odoo asks the device to fetch on its network with the "http" command.
type httpRequest struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

func New(config Config, manager *printer.Manager, printers Printers) *Agent {
	return &Agent{
		config:   config,
		manager:  manager,
		printers: printers,
		client:   &http.Client{Timeout: requestTimeout},
		status: Status{
			Paired: true, OdooURL: config.OdooURL, BoxName: config.Name,
			Mode: config.Mode, BusState: "connecting",
		},
		wake: make(chan struct{}, 1),
	}
}

func (a *Agent) Status() Status {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.status
}

// Run serves Odoo until done is closed: the bus announces new jobs and the periodic heartbeat
// keeps the device record up to date and acts as a fallback when the bus is down.
//
// In local mode Odoo sends the jobs itself, so there is no bus to listen to: the heartbeat is
// what keeps the device online in Odoo, refreshes its printers and picks up its commands.
func (a *Agent) Run(done <-chan struct{}) {
	if a.local() {
		logger.Infof("Linked to %s (local mode, heartbeat every %s)", a.config.OdooURL, HeartbeatInterval)
	} else {
		logger.Infof("Linked to %s (bus notifications, heartbeat every %s)", a.config.OdooURL, HeartbeatInterval)
		go a.listenBus(done)
	}

	for {
		a.heartbeat()
		taken := a.poll()

		wait := HeartbeatInterval
		if taken > 0 {
			wait = 0 // there may be more queued
		}
		select {
		case <-done:
			return
		case <-a.wake:
		case <-time.After(wait):
		}
	}
}

// Pair registers the device in Odoo with the pairing token and returns its own token.
func Pair(odooURL, pairingToken string, box map[string]any, printers []map[string]any) (string, error) {
	agent := &Agent{config: Config{OdooURL: odooURL}, client: &http.Client{Timeout: requestTimeout}}
	var reply struct {
		Token string `json:"token"`
		Error string `json:"error"`
	}
	if err := agent.call("/dooprint/register", map[string]any{
		"pairing_token": pairingToken,
		"device":        box,
		"printers":      printers,
	}, &reply); err != nil {
		return "", err
	}
	if reply.Error != "" {
		return "", fmt.Errorf("Odoo rejected the pairing: %s", reply.Error)
	}
	if reply.Token == "" {
		return "", fmt.Errorf("Odoo did not return the device token")
	}
	return reply.Token, nil
}

// local is true when Odoo reaches the device itself instead of the device calling for jobs.
func (a *Agent) local() bool {
	return a.config.Mode == "local"
}

func (a *Agent) heartbeat() {
	var reply struct {
		Name  string `json:"name"`
		Mode  string `json:"mode"`
		Error string `json:"error"`
	}
	if err := a.call("/dooprint/heartbeat", map[string]any{
		"token":    a.config.Token,
		"device":   a.boxInfo(),
		"printers": a.printers(),
	}, &reply); err != nil {
		a.fail(err.Error())
		a.localLink("disconnected")
		return
	}
	if reply.Error != "" {
		a.fail(fmt.Sprintf("Odoo answered %q: check the pairing", reply.Error))
		a.localLink("disconnected")
		return
	}
	a.mu.Lock()
	a.status.LastError = ""
	a.status.BoxName = reply.Name
	a.status.LastPoll = time.Now()
	a.mu.Unlock()
	a.localLink("connected")
}

// localLink reports the state of the link in local mode, where the heartbeat takes the place
// of the bus. In agent mode the bus reports its own state.
func (a *Agent) localLink(state string) {
	if a.local() {
		a.setBusState(state)
	}
}

func (a *Agent) boxInfo() map[string]any {
	return map[string]any{"mode": a.config.Mode, "name": a.config.Name}
}

func (a *Agent) poll() int {
	var reply struct {
		Jobs     []job     `json:"jobs"`
		Commands []command `json:"commands"`
		Error    string    `json:"error"`
	}
	if err := a.call("/dooprint/jobs", map[string]any{
		"token": a.config.Token,
		"limit": jobsPerRequest,
	}, &reply); err != nil {
		a.fail(err.Error())
		return 0
	}
	if reply.Error != "" {
		a.fail(fmt.Sprintf("Odoo answered %q: check the pairing", reply.Error))
		return 0
	}

	for _, j := range reply.Jobs {
		a.print(j)
	}
	for _, c := range reply.Commands {
		a.run(c)
	}
	return len(reply.Jobs) + len(reply.Commands)
}

func (a *Agent) print(j job) {
	errorCode := ""
	data := []byte(j.Payload)
	if strings.Contains(j.Payload, "<epos-print") {
		parsed, err := escpos.ParseXML(data)
		if err != nil {
			errorCode = "SchemaError"
		}
		data = parsed
	}

	if errorCode == "" {
		reply, err := a.manager.WriteAsync(j.Printer, data)
		if err == nil {
			if result := <-reply; !result.OK {
				err = result.Err
			}
		}
		if err != nil {
			errorCode = "EX_BADPORT"
			logger.Errorf("Job %d not printed: %v", j.Id, err)
		}
	}

	a.mu.Lock()
	if errorCode == "" {
		a.status.Printed++
	} else {
		a.status.Failed++
	}
	a.mu.Unlock()

	a.ack("job", j.Id, errorCode, "")
}

func (a *Agent) run(c command) {
	logger.Infof("Command received from Odoo: %s", c.Name)
	switch c.Name {
	case "refresh":
		a.ack("command", c.Id, "", "")
		a.heartbeat()
	case "restart":
		a.ack("command", c.Id, "", "")
		if a.Restart != nil {
			go a.Restart()
		}
	case "http":
		result, err := a.forward(c.Payload)
		if err != nil {
			a.ack("command", c.Id, err.Error(), "")
		} else {
			a.ack("command", c.Id, "", result)
		}
	default:
		a.ack("command", c.Id, "unknown command", "")
	}
}

// forward makes an HTTP request on the local network for Odoo and returns the status and the
// body of the answer as JSON.
func (a *Agent) forward(r httpRequest) (string, error) {
	if r.URL == "" {
		return "", fmt.Errorf("no url")
	}
	method := strings.ToUpper(r.Method)
	if method == "" {
		method = http.MethodGet
	}
	request, err := http.NewRequest(method, r.URL, strings.NewReader(r.Body))
	if err != nil {
		return "", err
	}
	for name, value := range r.Headers {
		request.Header.Set(name, value)
	}
	response, err := (&http.Client{Timeout: forwardTimeout}).Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, forwardMaxBody))
	if err != nil {
		return "", err
	}
	result, err := json.Marshal(map[string]any{"status": response.StatusCode, "body": string(body)})
	return string(result), err
}

func (a *Agent) ack(kind string, id int, errorCode string, result string) {
	var reply struct {
		Error string `json:"error"`
	}
	if err := a.call("/dooprint/ack", map[string]any{
		"token":  a.config.Token,
		"kind":   kind,
		"id":     id,
		"error":  errorCode,
		"result": result,
	}, &reply); err != nil {
		a.fail(fmt.Sprintf("%s %d handled but not confirmed: %v", kind, id, err))
	}
}

func (a *Agent) fail(message string) {
	a.mu.Lock()
	changed := a.status.LastError != message
	a.status.LastError = message
	a.mu.Unlock()
	if changed {
		logger.Warnf("Odoo link: %s", message)
	}
}

// call speaks Odoo JSON-RPC, which is what its controllers expose.
func (a *Agent) call(path string, params map[string]any, out any) error {
	body, err := json.Marshal(map[string]any{"jsonrpc": "2.0", "method": "call", "params": params})
	if err != nil {
		return err
	}

	url := strings.TrimSuffix(a.config.OdooURL, "/") + path
	response, err := a.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("%s answered %s", url, response.Status)
	}

	var envelope struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return fmt.Errorf("unreadable answer from %s", url)
	}
	if envelope.Error != nil {
		return fmt.Errorf("Odoo: %s", envelope.Error.Message)
	}
	return json.Unmarshal(envelope.Result, out)
}
