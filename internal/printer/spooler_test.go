package printer

import "testing"

func TestSpoolerPrinterIDRoundTrip(t *testing.T) {
	name, ok := DecodeSpoolerPrinterID(EncodeSpoolerPrinterID("PrinterPOS-80"))
	if !ok || name != "PrinterPOS-80" {
		t.Fatalf("got %q, %v", name, ok)
	}
}

func TestSpoolerPrinterIDRejectsOtherKinds(t *testing.T) {
	if _, ok := DecodeSpoolerPrinterID(EncodeLANPrinterID("192.168.0.9")); ok {
		t.Fatal("a LAN id must not decode as a spooler id")
	}
}

func TestWithoutUnsupportedKeepsOtherErrors(t *testing.T) {
	kept := withoutUnsupported([]UnavailableInfo{
		{Name: "a", Error: "libusb: not supported [code -12]"},
		{Name: "b", Error: "libusb: access denied [code -3]"},
	})
	if len(kept) != 1 || kept[0].Name != "b" {
		t.Fatalf("got %v", kept)
	}
}

func TestDevicePrinterIDRoundTrip(t *testing.T) {
	path := `\\?\usb#vid_0456&pid_0808#5&1a2b3c#{28d78fad-5a12-11d1-ae5b-0000f803a8c2}`
	decoded, ok := DecodeDevicePrinterID(EncodeDevicePrinterID(path))
	if !ok || decoded != path {
		t.Fatalf("got %q, %v", decoded, ok)
	}
}

func TestDevicePrinterIDRejectsOtherKinds(t *testing.T) {
	if _, ok := DecodeDevicePrinterID(EncodeSpoolerPrinterID("x")); ok {
		t.Fatal("a spooler id must not decode as a device id")
	}
}

func TestVidPidInPath(t *testing.T) {
	match := vidPidInPath.FindStringSubmatch(`\\?\usb#vid_0456&pid_0808#5&1a2b3c#{guid}`)
	if match == nil || match[1] != "0456" || match[2] != "0808" {
		t.Fatalf("got %v", match)
	}
}
