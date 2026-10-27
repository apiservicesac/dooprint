//go:build windows

package printer

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/apiservicesac/dooprint/internal/logger"

	"golang.org/x/sys/windows"
)

const (
	presentDeviceInterfaces = windows.DIGCF_PRESENT | windows.DIGCF_DEVICEINTERFACE
	interfaceDetailHeader   = 4
)

var (
	setupapi                       = syscall.NewLazyDLL("setupapi.dll")
	procSetupDiEnumInterfaces      = setupapi.NewProc("SetupDiEnumDeviceInterfaces")
	procSetupDiGetInterfaceDetailW = setupapi.NewProc("SetupDiGetDeviceInterfaceDetailW")

	// GUID_DEVINTERFACE_USBPRINT: every USB device bound to the stock usbprint driver.
	usbPrintGUID = windows.GUID{
		Data1: 0x28d78fad, Data2: 0x5a12, Data3: 0x11d1,
		Data4: [8]byte{0xae, 0x5b, 0x00, 0x00, 0xf8, 0x03, 0xa8, 0xc2},
	}
)

type deviceInterfaceData struct {
	Size      uint32
	ClassGUID windows.GUID
	Flags     uint32
	Reserved  uintptr
}

// detailSize is the cbSize Windows expects for the variable-length detail
// structure: it depends on the pointer size, not on the path length.
func detailSize() uint32 {
	if unsafe.Sizeof(uintptr(0)) == 8 {
		return 8
	}
	return 6
}

// listUSBPrintDevices returns the device path of every connected USB printer
// bound to usbprint, whether or not Windows created a printer for it.
func listUSBPrintDevices() ([]string, error) {
	set, err := windows.SetupDiGetClassDevsEx(&usbPrintGUID, "", 0, presentDeviceInterfaces, 0, "")
	if err != nil {
		return nil, fmt.Errorf("failed to list USB printer devices: %w", err)
	}
	defer windows.SetupDiDestroyDeviceInfoList(set)

	var paths []string
	for index := uint32(0); ; index++ {
		data := deviceInterfaceData{Size: uint32(unsafe.Sizeof(deviceInterfaceData{}))}
		result, _, _ := procSetupDiEnumInterfaces.Call(uintptr(set), 0,
			uintptr(unsafe.Pointer(&usbPrintGUID)), uintptr(index), uintptr(unsafe.Pointer(&data)))
		if result == 0 {
			break
		}
		path, err := interfacePath(set, &data)
		if err != nil {
			logger.Warnf("Skipping USB printer device: %v", err)
			continue
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func interfacePath(set windows.DevInfo, data *deviceInterfaceData) (string, error) {
	var required uint32
	procSetupDiGetInterfaceDetailW.Call(uintptr(set), uintptr(unsafe.Pointer(data)), 0, 0,
		uintptr(unsafe.Pointer(&required)), 0)
	if required <= interfaceDetailHeader {
		return "", fmt.Errorf("device interface without a path")
	}

	buffer := make([]byte, required)
	*(*uint32)(unsafe.Pointer(&buffer[0])) = detailSize()
	result, _, err := procSetupDiGetInterfaceDetailW.Call(uintptr(set), uintptr(unsafe.Pointer(data)),
		uintptr(unsafe.Pointer(&buffer[0])), uintptr(required), uintptr(unsafe.Pointer(&required)), 0)
	if result == 0 {
		return "", fmt.Errorf("failed to read the device path: %w", err)
	}
	return windows.UTF16PtrToString((*uint16)(unsafe.Pointer(&buffer[interfaceDetailHeader]))), nil
}

// writeToDevice opens the usbprint device for the length of one job. Opening
// it per job leaves it free for Windows and other programs in between.
func writeToDevice(devicePath string, data []byte) error {
	name, err := windows.UTF16PtrFromString(devicePath)
	if err != nil {
		return err
	}
	handle, err := windows.CreateFile(name, windows.GENERIC_WRITE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, 0, 0)
	if err != nil {
		return fmt.Errorf("failed to open USB printer %q: %w", devicePath, err)
	}
	defer windows.CloseHandle(handle)

	for len(data) > 0 {
		var written uint32
		if err := windows.WriteFile(handle, data, &written, nil); err != nil {
			return fmt.Errorf("failed to write to USB printer %q: %w", devicePath, err)
		}
		data = data[written:]
	}
	return nil
}
