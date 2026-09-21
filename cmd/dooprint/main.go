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
	"github.com/apiservicesac/dooprint/internal/update"
	"github.com/apiservicesac/dooprint/internal/webui"
)

const defaultPort = 4547

func main() {
	port := flag.Int("port", defaultPort, "port of the ePOS server and the web interface")
	dataDir := flag.String("data-dir", "", "folder for the configuration (default: the user config folder)")
	printerIP := flag.String("printer-id", "", "print the id and address of a network printer and exit")
	version := flag.Bool("version", false, "print the version and exit")
	flag.Parse()

	if *version {
		fmt.Println(app.Version)
		return
	}
	if *printerIP != "" {
		fmt.Printf("id:        %s\n", printer.EncodeLANPrinterID(*printerIP))
		fmt.Printf("address: <this-computer-ip>:%d/p/%s\n", *port, printer.EncodeLANPrinterID(*printerIP))
		return
	}
	config.Dir = *dataDir

	// Under the Windows service manager the service handler drives start and stop.
	if runAsService(*port) {
		return
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan struct{})
	go func() {
		<-stop
		close(done)
	}()
	if err := run(*port, done); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// run serves until done is closed.
func run(port int, done <-chan struct{}) error {
	// No logger.InitLogger(): logs go to standard output (journald with systemd) instead of a file.
	cfg, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("cannot open the configuration: %w", err)
	}
	if err := cfg.Load(); err != nil {
		logger.Warnf("configuration not read: %v", err)
	}

	// A Windows update leaves the replaced executable behind; it can go now that it is not running.
	update.CleanOld()

	manager := printer.NewManager()
	service := &app.Service{Config: cfg, Manager: manager, Port: port}
	service.Link = app.NewLink(service)
	srv := server.New(port, manager, webui.Register(service, manager))
	service.Running = srv.Running

	// server.New listens in the background and stops on its own if the port is taken.
	time.Sleep(time.Second)
	if !srv.Running() {
		return fmt.Errorf("cannot listen on port %d", port)
	}
	logger.Infof("dooprint %s listening on 0.0.0.0:%d (web interface at http://<ip>:%d)", app.Version, port, port)

	// When the device is already paired in agent mode, the Odoo link starts on its own.
	service.Link.Start()

	<-done
	if err := srv.Stop(); err != nil {
		return fmt.Errorf("error while stopping: %w", err)
	}
	return nil
}
