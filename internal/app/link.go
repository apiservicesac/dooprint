package app

import (
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/apiservicesac/dooprint/internal/agent"
	"github.com/apiservicesac/dooprint/internal/config"
	"github.com/apiservicesac/dooprint/internal/logger"
)

// Link manages the Odoo link: it pairs, starts and stops the agent without restarting the
// service, and keeps everything in the configuration so it survives restarts.
type Link struct {
	service *Service
	mu      sync.Mutex
	worker  *agent.Agent
	done    chan struct{}
}

func NewLink(service *Service) *Link {
	return &Link{service: service}
}

// Start runs the agent when the device is already paired. It also runs in local mode, where
// it only reports the device is alive and picks up its commands.
func (l *Link) Start() {
	link := l.service.Config.OdooLink()
	if link.Token == "" {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.startLocked(link)
}

func (l *Link) startLocked(link config.OdooLink) {
	l.stopLocked()
	l.worker = agent.New(agent.Config{
		OdooURL: link.URL,
		Token:   link.Token,
		Mode:    link.Mode,
		Name:    link.Name,
	}, l.service.Manager, l.reportedPrinters)
	// The restart command from Odoo ends the process, and the service manager starts it again.
	l.worker.Restart = func() {
		logger.Infof("Restart requested from Odoo")
		Restart()
	}
	l.done = make(chan struct{})
	go l.worker.Run(l.done)
}

func (l *Link) stopLocked() {
	if l.done != nil {
		close(l.done)
		l.done = nil
	}
	l.worker = nil
}

// Pair pairs the device with Odoo and starts the link. `pairing` is the string Odoo shows when
// connecting a device: its address with the token, e.g. https://odoo.example.com?token=…&db_name=…
func (l *Link) Pair(pairing, name, mode string) error {
	odooURL, pairingToken, err := parsePairing(pairing)
	if err != nil {
		return err
	}
	if mode != "local" {
		mode = "agent"
	}
	if name == "" {
		name = hostName()
	}

	token, err := agent.Pair(odooURL, pairingToken, map[string]any{
		"identifier": l.service.Config.Identifier(),
		"name":       name,
		"mode":       mode,
		"address":    l.service.Address(),
		"version":    Version,
		"os":         hostOS(),
	}, l.reportedPrinters())
	if err != nil {
		return err
	}

	link := config.OdooLink{URL: odooURL, Token: token, Name: name, Mode: mode}
	if err := l.service.Config.SetOdooLink(link); err != nil {
		return err
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.startLocked(link)
	return nil
}

// parsePairing splits the Odoo address and the token out of the pairing string.
func parsePairing(pairing string) (string, string, error) {
	parsed, err := url.Parse(strings.TrimSpace(pairing))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", "", fmt.Errorf("invalid pairing token: copy it whole from Odoo")
	}
	token := parsed.Query().Get("token")
	if token == "" {
		return "", "", fmt.Errorf("the pairing token has no token: copy it whole from Odoo")
	}
	return parsed.Scheme + "://" + parsed.Host + strings.TrimSuffix(parsed.Path, "/"), token, nil
}

// Unpair forgets the pairing and stops the agent.
func (l *Link) Unpair() error {
	l.mu.Lock()
	l.stopLocked()
	l.mu.Unlock()
	return l.service.Config.SetOdooLink(config.OdooLink{})
}

// Status describes the link for the web interface.
func (l *Link) Status() agent.Status {
	link := l.service.Config.OdooLink()
	l.mu.Lock()
	worker := l.worker
	l.mu.Unlock()

	if worker != nil {
		return worker.Status()
	}
	return agent.Status{
		Paired:   link.Token != "",
		OdooURL:  link.URL,
		BoxName:  link.Name,
		Mode:     link.Mode,
		BusState: "off",
	}
}

// reportedPrinters is what Odoo is told about the detected printers.
func (l *Link) reportedPrinters() []map[string]any {
	reported := make([]map[string]any, 0)
	for _, p := range l.service.Printers().Printers {
		connection := "usb"
		if p.IsLAN {
			connection = "network"
		}
		reported = append(reported, map[string]any{
			"id": p.Id, "name": p.Name, "type": p.Type, "connection": connection, "ip": p.LANIp,
		})
	}
	return reported
}
