// Package app holds the device logic: listing printers, adding and removing network printers,
// status and troubleshooting data. Both the web interface and the Odoo link use it.
package app

import (
	"fmt"
	"os"
	"time"

	"github.com/apiservicesac/dooprint/internal/agent"
	"github.com/apiservicesac/dooprint/internal/config"
	"github.com/apiservicesac/dooprint/internal/logger"
	"github.com/apiservicesac/dooprint/internal/printer"
	"github.com/apiservicesac/dooprint/internal/update"
	"github.com/apiservicesac/dooprint/internal/util"
)

type Printer struct {
	Name   string `json:"name"`
	Ip     string `json:"ip"`
	Id     string `json:"id"`
	IsLAN  bool   `json:"isLAN"`
	LANIp  string `json:"lanIp,omitempty"`
	Online bool   `json:"online"`
	Type   string `json:"type"`
}

type UnavailablePrinter struct {
	Name     string `json:"name"`
	ErrorMsg string `json:"errorMsg"`
	IsLAN    bool   `json:"isLAN"`
	LANIp    string `json:"lanIp,omitempty"`
}

type Printers struct {
	ErrorMsg            string               `json:"errorMsg"`
	Printers            []Printer            `json:"printers"`
	UnavailablePrinters []UnavailablePrinter `json:"unavailablePrinters"`
}

type Variable struct {
	ServerRunning bool   `json:"serverRunning"`
	Os            string `json:"os"`
	Mode          string `json:"mode"`
	Port          int    `json:"port"`
	Name          string `json:"name"`
	Version       string `json:"version"`
	Address       string `json:"address"`
}

type TroubleshootInfo struct {
	ActiveFirewall string `json:"activeFirewall"`
	FirewallZone   string `json:"firewallZone"`
	Port           int    `json:"port"`
	Subnet         string `json:"subnet"`
	LocalIP        string `json:"localIp"`
	ExecPath       string `json:"execPath"`
}

// restartDelay leaves the answer time to reach the browser before the process ends.
const restartDelay = 1500 * time.Millisecond

// Restart ends the process so it starts again with whatever binary is in place: systemd on
// Linux and the service recovery on Windows bring it back. A non-zero code is what the Windows
// service manager treats as a failure worth restarting.
func Restart() {
	logger.Infof("Restarting the service")
	os.Exit(1)
}

type Service struct {
	Config  *config.Manager
	Manager *printer.Manager
	Port    int
	Running func() bool
	Link    *Link
}

// CheckUpdate is which version is published, next to the running one.
func (s *Service) CheckUpdate() (update.Release, error) {
	return update.Latest(Version)
}

// InstallUpdate puts the latest release in place and restarts the service, which is what makes
// the new version run. The restart waits for the answer to reach the browser.
func (s *Service) InstallUpdate() (update.Release, error) {
	release, err := update.Install(Version)
	if err != nil {
		return release, err
	}
	go func() {
		time.Sleep(restartDelay)
		Restart()
	}()
	return release, nil
}

// LinkStatus describes the Odoo link for the web interface.
func (s *Service) LinkStatus() agent.Status {
	if s.Link == nil {
		return agent.Status{}
	}
	return s.Link.Status()
}

func (s *Service) Variable() Variable {
	running := false
	if s.Running != nil {
		running = s.Running()
	}
	return Variable{
		ServerRunning: running,
		Os:            hostOS(),
		Mode:          s.Config.OdooLink().Mode,
		Port:          s.Port,
		Name:          hostName(),
		Version:       Version,
		Address:       s.Address(),
	}
}

// Address is how this device sees itself on the network; Odoo uses it in local mode.
func (s *Service) Address() string {
	return fmt.Sprintf("%s:%d", util.GetLocalIP(true), s.Port)
}

// PrinterURL is the address of a given printer.
func (s *Service) PrinterURL(id string) string {
	return fmt.Sprintf("%s/p/%s", s.Address(), id)
}

func (s *Service) Printers() Printers {
	printers := make([]Printer, 0)
	unavailable := make([]UnavailablePrinter, 0)
	errorMsg := ""

	usb, err := printer.ListUSBPrinters()
	if err != nil {
		errorMsg = err.Error()
		logger.Errorf("USB printer detection failed: %v", err)
	} else {
		for _, info := range usb.Available {
			printers = append(printers, Printer{
				Id:     info.Id,
				Name:   info.Name,
				Ip:     s.PrinterURL(info.Id),
				Online: true,
				Type:   string(info.Type),
			})
		}
		for _, info := range usb.Unavailable {
			unavailable = append(unavailable, UnavailablePrinter{Name: info.Name, ErrorMsg: info.Error})
		}
	}

	for _, info := range printer.ListLANPrinters(s.Config) {
		printers = append(printers, Printer{
			Id:    info.Id,
			Name:  fmt.Sprintf("Network - %s", info.IP),
			Ip:    s.PrinterURL(info.Id),
			IsLAN: true,
			LANIp: info.IP,
			Type:  string(printer.TypeReceipt),
		})
	}

	return Printers{Printers: printers, UnavailablePrinters: unavailable, ErrorMsg: errorMsg}
}

func (s *Service) AddLANPrinter(ip string) error {
	ip, err := printer.ValidateIPAddress(ip)
	if err != nil {
		return fmt.Errorf("invalid IP address: %s, error: %v", ip, err)
	}
	if err := printer.CheckLANPrinter(ip); err != nil {
		return fmt.Errorf("LAN printer unreachable: %s, error: %v", ip, err)
	}
	if err := s.Config.AddLanEposPrinter(ip); err != nil {
		return fmt.Errorf("failed to save LAN printer: %s, error: %v", ip, err)
	}
	logger.Infof("LAN printer added: %s", ip)
	return nil
}

func (s *Service) RemoveLANPrinter(ip string) error {
	logger.Infof("Removing LAN printer: %s", ip)
	return s.Config.RemoveLANPrinter(ip)
}

func (s *Service) CheckLANPrinterStatus(ip string) bool {
	return printer.CheckLANPrinter(ip) == nil
}

func (s *Service) Troubleshoot() TroubleshootInfo {
	netInfo := util.GetNetworkInfo()
	execPath, _ := os.Executable()
	return TroubleshootInfo{
		ActiveFirewall: netInfo.ActiveFirewall,
		FirewallZone:   netInfo.Zone,
		Port:           s.Port,
		Subnet:         netInfo.Subnet,
		LocalIP:        netInfo.IP,
		ExecPath:       execPath,
	}
}
