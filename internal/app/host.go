package app

import (
	"os"
	"runtime"
)

// Version is the device version reported to Odoo. Releases set it from the tag with
// -ldflags "-X github.com/apiservicesac/dooprint/internal/app.Version=1.2.0".
var Version = "1.0"

func hostOS() string { return runtime.GOOS }

func hostName() string {
	name, err := os.Hostname()
	if err != nil || name == "" {
		return "dooprint"
	}
	return name
}
