// dooprint: print service for Odoo, running without a desktop window.
//
// It starts the ePOS server and the configuration web interface on the same port:
//
//	POST /p/<id>/cgi-bin/epos/service.cgi   network or USB printer by id
//	POST /cgi-bin/epos/service.cgi          first USB printer found
//	GET  /                                  web interface
//	GET  /api/...                           used by the web interface
//
// Network printers need no previous setup: the id carries the IP address.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/apiservicesac/dooprint/internal/app"
	"github.com/apiservicesac/dooprint/internal/config"
	"github.com/apiservicesac/dooprint/internal/logger"
	"github.com/apiservicesac/dooprint/internal/printer"
	"github.com/apiservicesac/dooprint/internal/server"
	"github.com/apiservicesac/dooprint/internal/webui"
)

const defaultPort = 4547

func main() {
	port := flag.Int("port", defaultPort, "port of the ePOS server and the web interface")
	printerIP := flag.String("printer-id", "", "print the id and address of a network printer and exit")
	flag.Parse()

	if *printerIP != "" {
		fmt.Printf("id:        %s\n", printer.EncodeLANPrinterID(*printerIP))
		fmt.Printf("address: <this-computer-ip>:%d/p/%s\n", *port, printer.EncodeLANPrinterID(*printerIP))
		return
	}

	// No logger.InitLogger(): logs go to standard output (journald with systemd, the event
	// viewer with NSSM on Windows) instead of a file.
	cfg, err := config.NewManager()
	if err != nil {
		logger.Fatalf("cannot open the configuration: %v", err)
	}
	if err := cfg.Load(); err != nil {
		logger.Warnf("configuration not read: %v", err)
	}

	manager := printer.NewManager()
	service := &app.Service{Config: cfg, Manager: manager, Port: *port}
	service.Link = app.NewLink(service)
	srv := server.New(*port, manager, webui.Register(service, manager))
	service.Running = srv.Running

	// server.New listens in the background and stops on its own if the port is taken.
	time.Sleep(time.Second)
	if !srv.Running() {
		logger.Fatalf("cannot listen on port %d", *port)
	}
	logger.Infof("dooprint listening on 0.0.0.0:%d (web interface at http://<ip>:%d)", *port, *port)

	// When the device is already paired in agent mode, the Odoo link starts on its own.
	service.Link.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	if err := srv.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "error while stopping: %v\n", err)
		os.Exit(1)
	}
}
