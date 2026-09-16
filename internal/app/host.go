package app

import (
	"os"
	"runtime"
)

// Version is the device version reported to Odoo.
const Version = "1.0"

func hostOS() string { return runtime.GOOS }

func hostName() string {
	name, err := os.Hostname()
	if err != nil || name == "" {
		return "dooprint"
	}
	return name
}
