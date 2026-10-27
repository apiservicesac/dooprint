//go:build !windows

package printer

import "errors"

var errWindowsOnly = errors.New("only available on Windows")

func listSpoolerUSBPrinters() ([]string, error) {
	return nil, nil
}

func writeToSpooler(string, []byte) error {
	return errWindowsOnly
}

func listUSBPrintDevices() ([]string, error) {
	return nil, nil
}

func writeToDevice(string, []byte) error {
	return errWindowsOnly
}
