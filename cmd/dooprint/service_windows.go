//go:build windows

package main

import (
	"github.com/apiservicesac/dooprint/internal/config"
	"github.com/apiservicesac/dooprint/internal/logger"

	"golang.org/x/sys/windows/svc"
)

const serviceName = "dooprint"

// runAsService runs dooprint under the Windows service manager when it started it, and reports
// false otherwise so it runs as a normal console program.
func runAsService(port int) bool {
	isService, err := svc.IsWindowsService()
	if err != nil || !isService {
		return false
	}
	// A service has no console: keep the logs in files next to the configuration.
	logger.InitLogger(config.Dir)
	if err := svc.Run(serviceName, &handler{port: port}); err != nil {
		logger.Fatalf("service failed: %v", err)
	}
	return true
}

type handler struct {
	port int
}

func (h *handler) Execute(_ []string, requests <-chan svc.ChangeRequest, status chan<- svc.Status) (bool, uint32) {
	status <- svc.Status{State: svc.StartPending}

	done := make(chan struct{})
	failed := make(chan error, 1)
	go func() { failed <- run(h.port, done) }()

	status <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	for {
		select {
		case err := <-failed:
			// It could not start (for example the port is taken): exit with an error so the
			// recovery options of the service apply.
			logger.Errorf("dooprint stopped: %v", err)
			return true, 1
		case request := <-requests:
			switch request.Cmd {
			case svc.Interrogate:
				status <- request.CurrentStatus
			case svc.Stop, svc.Shutdown:
				status <- svc.Status{State: svc.StopPending}
				close(done)
				<-failed
				return false, 0
			}
		}
	}
}
