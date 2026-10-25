//go:build !windows

package printer

import "errors"

var errSpoolerUnsupported = errors.New("the Windows print spooler is only available on Windows")

type spoolerJob struct{}

func listSpoolerUSBPrinters() ([]string, error) {
	return nil, nil
}

func openSpoolerJob(string) (*spoolerJob, error) {
	return nil, errSpoolerUnsupported
}

func (j *spoolerJob) write([]byte) error {
	return errSpoolerUnsupported
}

func (j *spoolerJob) close() {}
