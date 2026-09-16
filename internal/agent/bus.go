package agent

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/apiservicesac/dooprint/internal/logger"

	"github.com/coder/websocket"
)

// listenBus keeps the Odoo bus open (the same WebSocket its web client uses) and wakes the agent
// up as soon as there is a job. It subscribes to the device token as channel: only this agent
// knows it, so no login is needed.
func (a *Agent) listenBus(done <-chan struct{}) {
	for {
		if err := a.readBus(done); err != nil {
			a.setBusState("disconnected")
			logger.Warnf("Odoo bus: %v (retrying in %s)", err, reconnectDelay)
		}
		select {
		case <-done:
			return
		case <-time.After(reconnectDelay):
		}
	}
}

func (a *Agent) readBus(done <-chan struct{}) error {
	address, origin, err := busAddress(a.config.OdooURL)
	if err != nil {
		return err
	}
	if a.busFallback != "" {
		address = a.busFallback
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		select {
		case <-done:
			cancel()
		case <-ctx.Done():
		}
	}()

	connection, _, err := websocket.Dial(ctx, address, &websocket.DialOptions{
		HTTPHeader: map[string][]string{"Origin": {origin}},
	})
	if err != nil {
		// Odoo with several workers serves the bus on its longpolling port; 8069 only answers when
		// it runs with one. When talking to it directly, without a reverse proxy, switch
		// to 8072.
		if fallback := longpollingAddress(address); fallback != "" && a.busFallback == "" {
			a.busFallback = fallback
			logger.Infof("Odoo bus: trying the longpolling port at %s", fallback)
		}
		return err
	}
	defer connection.CloseNow()

	subscribe, err := json.Marshal(map[string]any{
		"event_name": "subscribe",
		"data":       map[string]any{"channels": []string{a.config.Token}, "last": 0},
	})
	if err != nil {
		return err
	}
	if err := connection.Write(ctx, websocket.MessageText, subscribe); err != nil {
		return err
	}

	a.setBusState("connected")
	logger.Infof("Odoo bus connected: real-time notifications")

	for {
		if _, _, err := connection.Read(ctx); err != nil {
			return err
		}
		// Any message on the channel means there is something to print.
		select {
		case a.wake <- struct{}{}:
		default:
		}
	}
}

// busAddress turns the Odoo URL into the WebSocket one and returns the Origin Odoo expects.
func busAddress(odooURL string) (string, string, error) {
	parsed, err := url.Parse(strings.TrimSuffix(odooURL, "/"))
	if err != nil {
		return "", "", err
	}
	scheme := "ws"
	if parsed.Scheme == "https" {
		scheme = "wss"
	}
	return scheme + "://" + parsed.Host + "/websocket", parsed.Scheme + "://" + parsed.Host, nil
}

// longpollingAddress swaps port 8069 for 8072, the one serving the bus when Odoo runs with
// several workers and no reverse proxy.
func longpollingAddress(address string) string {
	if !strings.Contains(address, ":8069/") {
		return ""
	}
	return strings.Replace(address, ":8069/", ":8072/", 1)
}

func (a *Agent) setBusState(state string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.status.BusState = state
}
