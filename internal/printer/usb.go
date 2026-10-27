package printer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/apiservicesac/dooprint/internal/logger"

	"github.com/google/gousb"
)

// Connection is the way dooprint talks to a USB printer.
type Connection string

const (
	ConnectionLibusb  Connection = "libusb"  // straight to the device, needs a WinUSB or Linux/macOS driver
	ConnectionSpooler Connection = "spooler" // through a printer installed in Windows
	ConnectionDevice  Connection = "device"  // straight to the Windows usbprint device, no printer installed
)

// Info describes a USB printer that was successfully probed.
type Info struct {
	Id         string
	Name       string
	Type       Type
	Connection Connection
}

// UnavailableInfo describes a device that looks like a printer but could
// not be opened, usually because of permissions or a driver holding it.
type UnavailableInfo struct {
	Name  string
	Error string
}

type Printers struct {
	Available   []Info
	Unavailable []UnavailableInfo
}

// EndpointInfo locates the bulk OUT endpoint used to write to a device.
type EndpointInfo struct {
	config           int
	iFace            int
	alternateSetting int
	outEndpoint      int
}

// LibUsbPrinter is the raw identity read off a device during a scan.
type LibUsbPrinter struct {
	Serial   string
	Path     string
	Name     string
	VidPid   string
	DeviceId DeviceID
}

func ListUSBPrinters() (*Printers, error) {
	logger.Debug("Starting USB printer detection")
	ctx := gousb.NewContext()
	defer func(ctx *gousb.Context) {
		_ = ctx.Close()
	}(ctx)

	var keys []string
	// First list all without opening devices, to avoid permission errors on some platforms
	var descriptors []gousb.DeviceDesc
	_, err := openDevices(ctx, func(desc *gousb.DeviceDesc) bool {
		if _, supported := findPrinterEndpoint(desc); supported {
			descriptors = append(descriptors, *desc)
			keys = append(keys, fingerprintKey(desc))
		}
		return false
	})

	if err != nil {
		return nil, fmt.Errorf("failed to open USB devices for listing: %w", err)
	}

	if !usbCache.HasChanged(keys) && !usbCache.HasUnavailable() {
		logger.Debugf("USB unchanged → using cache")
		cached := &Printers{Available: usbCache.Get()}
		addWindowsPrinters(cached)
		return cached, nil
	}

	logger.Debugf("USB changed → rescanning devices")

	result := &Printers{
		Available:   make([]Info, 0),
		Unavailable: make([]UnavailableInfo, 0),
	}
	for _, desc := range descriptors {
		info, err := GetPrinterInfo(ctx, &desc)
		if err != nil {
			// Device is not accessible, likely due to permissions / drivers.
			vid := fmt.Sprintf("%04X", uint16(desc.Vendor))
			pid := fmt.Sprintf("%04X", uint16(desc.Product))
			result.Unavailable = append(result.Unavailable, UnavailableInfo{
				Name:  getPrinterFriendlyName(vid, pid),
				Error: err.Error(),
			})
		} else if info != nil {
			id, err := encodePrinterID(info)
			if err != nil {
				logger.Errorf("failed to encode printer ID: %v", err)
				continue
			}
			result.Available = append(result.Available, Info{
				Id:         id,
				Name:       info.Name,
				Type:       getPrinterType(info.VidPid),
				Connection: ConnectionLibusb,
			})
		}
	}

	libusbAvailable := result.Available
	addWindowsPrinters(result)
	usbCache.Update(keys, libusbAvailable, result.Unavailable)
	return result, nil
}

func GetPrinterInfo(ctx *gousb.Context, descToFind *gousb.DeviceDesc) (*LibUsbPrinter, error) {
	logger.Debugf("Attempting to get info for USB device: Bus %d, Address %d, Vendor %04X, Product %04X", descToFind.Bus, descToFind.Address, uint16(descToFind.Vendor), uint16(descToFind.Product))
	var found bool
	devices, err := openDevices(ctx, func(desc *gousb.DeviceDesc) bool {
		if found {
			return false
		}
		if descToFind.Bus != desc.Bus || descToFind.Address != desc.Address ||
			descToFind.Vendor != desc.Vendor || descToFind.Product != desc.Product {
			return false
		}
		found = true
		return true
	})

	if err != nil {
		return nil, fmt.Errorf("failed to open USB device for info retrieval: %w", err)
	}

	if len(devices) == 0 {
		return nil, nil
	}

	defer func() {
		for _, d := range devices {
			_ = d.Close()
		}
	}()

	device := devices[0]
	deviceID := getPrinterDeviceID(device)
	info := LibUsbPrinter{}
	productName, _ := device.Product()
	vendorName, _ := device.Manufacturer()
	serial, _ := device.SerialNumber()

	if productName == "" {
		if mdl, ok := deviceID["MDL"]; ok {
			productName = mdl
		} else {
			productName = fmt.Sprintf("PID: %04X", uint16(descToFind.Product))
		}
	}

	if vendorName == "" {
		if vendor, ok := deviceID["MFG"]; ok {
			vendorName = vendor
		} else {
			vendorName = fmt.Sprintf("VID: %04X", uint16(descToFind.Vendor))
		}
	}

	info.Name = fmt.Sprintf("%s %s", vendorName, productName)
	info.Serial = serial
	info.Path = pathToString(descToFind)
	info.VidPid = fmt.Sprintf("%04X:%04X", uint16(descToFind.Vendor), uint16(descToFind.Product))
	info.DeviceId = deviceID
	logger.Debugf("USB printer: %s (Serial: %s)", info.Name, info.Serial)
	return &info, nil
}

func fingerprintKey(desc *gousb.DeviceDesc) string {
	return fmt.Sprintf("%d-%d-%04X:%04X-%s",
		desc.Bus,
		desc.Address,
		desc.Vendor,
		desc.Product,
		pathToString(desc),
	)
}

func findPrinterEndpoint(desc *gousb.DeviceDesc) (EndpointInfo, bool) {
	missingBulkOut := false
	logger.Debugf("Finding printer endpoint for device: %v", desc)
	for cfgNum, cfg := range desc.Configs {
		for _, iFace := range cfg.Interfaces {
			for _, alt := range iFace.AltSettings {
				if alt.Class != gousb.ClassPrinter && !isKnownPrinter(desc) {
					continue
				}
				if epNum, ok := matchBulkOutEndpoint(alt); ok {
					return EndpointInfo{
						config:           cfgNum,
						iFace:            iFace.Number,
						alternateSetting: alt.Alternate,
						outEndpoint:      epNum,
					}, true
				} else {
					missingBulkOut = true
				}
			}
		}
	}
	if missingBulkOut {
		logger.Warnf("Printer device rejected during endpoint matching: VID=%s, PID=%s, reason=no bulk OUT endpoint found", desc.Vendor, desc.Product)
	}
	return EndpointInfo{}, false
}

func matchBulkOutEndpoint(alt gousb.InterfaceSetting) (int, bool) {
	for _, ep := range alt.Endpoints {
		if ep.Direction == gousb.EndpointDirectionOut &&
			ep.TransferType == gousb.TransferTypeBulk {
			return ep.Number, true
		}
	}
	return 0, false
}

// addWindowsPrinters adds the USB printers libusb cannot open because Windows
// binds them to its stock usbprint driver. Those print through the Windows
// printer when one is installed, and straight to the usbprint device when
// Windows never created one.
func addWindowsPrinters(result *Printers) {
	names, err := listSpoolerUSBPrinters()
	if err != nil {
		logger.Warnf("Windows spooler listing failed: %v", err)
	}
	for _, name := range names {
		result.Available = append(result.Available, Info{
			Id:         EncodeSpoolerPrinterID(name),
			Name:       name,
			Type:       TypeReceipt,
			Connection: ConnectionSpooler,
		})
	}

	added := len(names)
	if added == 0 {
		added = addUSBPrintDevices(result)
	}
	if added > 0 {
		result.Unavailable = withoutUnsupported(result.Unavailable)
	}
}

var vidPidInPath = regexp.MustCompile(`(?i)vid_([0-9a-f]{4})&pid_([0-9a-f]{4})`)

// addUSBPrintDevices adds each connected usbprint device and returns how many.
func addUSBPrintDevices(result *Printers) int {
	paths, err := listUSBPrintDevices()
	if err != nil {
		logger.Warnf("USB printer device listing failed: %v", err)
		return 0
	}
	for _, path := range paths {
		logger.Infof("USB printer device %s", path)
		name := "USB printer"
		if match := vidPidInPath.FindStringSubmatch(path); match != nil {
			name = getPrinterFriendlyName(strings.ToUpper(match[1]), strings.ToUpper(match[2]))
		}
		result.Available = append(result.Available, Info{
			Id:         EncodeDevicePrinterID(path),
			Name:       name,
			Type:       TypeReceipt,
			Connection: ConnectionDevice,
		})
	}
	return len(paths)
}

func withoutUnsupported(devices []UnavailableInfo) []UnavailableInfo {
	remaining := make([]UnavailableInfo, 0, len(devices))
	for _, device := range devices {
		if !strings.Contains(device.Error, "not supported") {
			remaining = append(remaining, device)
		}
	}
	return remaining
}
