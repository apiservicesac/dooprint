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
