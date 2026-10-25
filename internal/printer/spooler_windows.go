//go:build windows

package printer

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"
)

const (
	printerEnumLocal       = 0x2
	printerEnumConnections = 0x4
	printerInfoLevelPorts  = 5
	documentInfoLevel      = 1
	rawDatatype            = "RAW"
	documentName           = "dooprint"
	usbPortPrefix          = "USB"
)

var (
	winspool         = syscall.NewLazyDLL("winspool.drv")
	procEnumPrinters = winspool.NewProc("EnumPrintersW")
	procOpenPrinter  = winspool.NewProc("OpenPrinterW")
	procClosePrinter = winspool.NewProc("ClosePrinter")
	procStartDoc     = winspool.NewProc("StartDocPrinterW")
	procEndDoc       = winspool.NewProc("EndDocPrinter")
	procStartPage    = winspool.NewProc("StartPagePrinter")
	procEndPage      = winspool.NewProc("EndPagePrinter")
	procWritePrinter = winspool.NewProc("WritePrinter")
)

type printerInfoPorts struct {
	PrinterName              *uint16
	PortName                 *uint16
	Attributes               uint32
	DeviceNotSelectedTimeout uint32
	TransmissionRetryTimeout uint32
}

type documentInfo struct {
	DocName    *uint16
	OutputFile *uint16
	Datatype   *uint16
}

// spoolerJob is one RAW document sent to a printer installed in Windows.
type spoolerJob struct {
	handle syscall.Handle
}

// listSpoolerUSBPrinters returns the names of the Windows printers
// installed on a USB port, which print through the stock usbprint driver.
func listSpoolerUSBPrinters() ([]string, error) {
	flags := uintptr(printerEnumLocal | printerEnumConnections)
	var needed, returned uint32
	procEnumPrinters.Call(flags, 0, printerInfoLevelPorts, 0, 0,
		uintptr(unsafe.Pointer(&needed)), uintptr(unsafe.Pointer(&returned)))
	if needed == 0 {
		return nil, nil
	}

	buffer := make([]byte, needed)
	result, _, err := procEnumPrinters.Call(flags, 0, printerInfoLevelPorts,
		uintptr(unsafe.Pointer(&buffer[0])), uintptr(needed),
		uintptr(unsafe.Pointer(&needed)), uintptr(unsafe.Pointer(&returned)))
	if result == 0 {
		return nil, fmt.Errorf("failed to list Windows printers: %w", err)
	}

	entries := unsafe.Slice((*printerInfoPorts)(unsafe.Pointer(&buffer[0])), returned)
	names := make([]string, 0, returned)
	for _, entry := range entries {
		port := syscall.UTF16ToString(utf16Slice(entry.PortName))
		if strings.HasPrefix(strings.ToUpper(port), usbPortPrefix) {
			names = append(names, syscall.UTF16ToString(utf16Slice(entry.PrinterName)))
		}
	}
	return names, nil
}

// utf16Slice reads a NUL-terminated UTF-16 string that Windows placed
// inside the buffer returned by EnumPrinters.
func utf16Slice(start *uint16) []uint16 {
	if start == nil {
		return nil
	}
	length := 0
	for ptr := unsafe.Pointer(start); *(*uint16)(ptr) != 0; ptr = unsafe.Add(ptr, 2) {
		length++
	}
	return unsafe.Slice(start, length+1)
}

func openSpoolerJob(printerName string) (*spoolerJob, error) {
	name, err := syscall.UTF16PtrFromString(printerName)
	if err != nil {
		return nil, err
	}
	var handle syscall.Handle
	result, _, callErr := procOpenPrinter.Call(uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(&handle)), 0)
	if result == 0 {
		return nil, fmt.Errorf("failed to open Windows printer %q: %w", printerName, callErr)
	}

	job := &spoolerJob{handle: handle}
	docName, _ := syscall.UTF16PtrFromString(documentName)
	datatype, _ := syscall.UTF16PtrFromString(rawDatatype)
	info := documentInfo{DocName: docName, Datatype: datatype}
	result, _, callErr = procStartDoc.Call(uintptr(handle), documentInfoLevel, uintptr(unsafe.Pointer(&info)))
	if result == 0 {
		job.release()
		return nil, fmt.Errorf("failed to start print document on %q: %w", printerName, callErr)
	}
	result, _, callErr = procStartPage.Call(uintptr(handle))
	if result == 0 {
		procEndDoc.Call(uintptr(handle))
		job.release()
		return nil, fmt.Errorf("failed to start print page on %q: %w", printerName, callErr)
	}
	return job, nil
}

func (j *spoolerJob) write(data []byte) error {
	for len(data) > 0 {
		var written uint32
		result, _, err := procWritePrinter.Call(uintptr(j.handle),
			uintptr(unsafe.Pointer(&data[0])), uintptr(len(data)), uintptr(unsafe.Pointer(&written)))
		if result == 0 {
			return fmt.Errorf("failed to write to Windows printer: %w", err)
		}
		data = data[written:]
	}
	return nil
}

// close finishes the document so the spooler sends it to the printer.
func (j *spoolerJob) close() {
	procEndPage.Call(uintptr(j.handle))
	procEndDoc.Call(uintptr(j.handle))
	j.release()
}

func (j *spoolerJob) release() {
	procClosePrinter.Call(uintptr(j.handle))
}
