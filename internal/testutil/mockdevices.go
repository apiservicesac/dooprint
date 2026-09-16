package testutil

import (
	"github.com/google/gousb"
)

// MockAltWithBulkOut returns an InterfaceSetting containing a Bulk IN (ep 1) and Bulk OUT (ep 2) endpoint.
func MockAltWithBulkOut() gousb.InterfaceSetting {
	return gousb.InterfaceSetting{
		Endpoints: map[gousb.EndpointAddress]gousb.EndpointDesc{
			gousb.EndpointAddress(1): {
				Number:       1,
				Direction:    gousb.EndpointDirectionIn,
				TransferType: gousb.TransferTypeBulk,
			},
			gousb.EndpointAddress(2): {
				Number:       2,
				Direction:    gousb.EndpointDirectionOut,
				TransferType: gousb.TransferTypeBulk,
			},
		},
	}
}

// MockAltInterrupt returns an InterfaceSetting containing only an Interrupt OUT endpoint.
func MockAltInterrupt() gousb.InterfaceSetting {
	return gousb.InterfaceSetting{
		Endpoints: map[gousb.EndpointAddress]gousb.EndpointDesc{
			gousb.EndpointAddress(3): {
				Number:       3,
				Direction:    gousb.EndpointDirectionOut,
				TransferType: gousb.TransferTypeInterrupt,
			},
		},
	}
}

// MockEpsonPrinterDesc returns a standard USB printer class DeviceDesc (Epson 0x04B8:0x0202).
func MockEpsonPrinterDesc() *gousb.DeviceDesc {
	return &gousb.DeviceDesc{
		Bus:     1,
		Address: 2,
		Vendor:  0x04B8,
		Product: 0x0202,
		Path:    []int{1, 2},
		Configs: map[int]gousb.ConfigDesc{
			1: {
				Number: 1,
				Interfaces: []gousb.InterfaceDesc{
					{
						Number: 0,
						AltSettings: []gousb.InterfaceSetting{
							{
								Number:    0,
								Alternate: 0,
								Class:     gousb.ClassPrinter,
								Endpoints: map[gousb.EndpointAddress]gousb.EndpointDesc{
									gousb.EndpointAddress(1): {
										Number:       1,
										Direction:    gousb.EndpointDirectionOut,
										TransferType: gousb.TransferTypeBulk,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// MockZebraPrinterDesc returns a known vendor-specific class DeviceDesc (Zebra 0x0A5F:0x0187).
func MockZebraPrinterDesc() *gousb.DeviceDesc {
	return &gousb.DeviceDesc{
		Bus:     1,
		Address: 3,
		Vendor:  0x0A5F,
		Product: 0x0187,
		Path:    []int{1, 3},
		Configs: map[int]gousb.ConfigDesc{
			1: {
				Number: 1,
				Interfaces: []gousb.InterfaceDesc{
					{
						Number: 0,
						AltSettings: []gousb.InterfaceSetting{
							{
								Number:    0,
								Alternate: 0,
								Class:     gousb.ClassVendorSpec,
								Endpoints: map[gousb.EndpointAddress]gousb.EndpointDesc{
									gousb.EndpointAddress(2): {
										Number:       2,
										Direction:    gousb.EndpointDirectionOut,
										TransferType: gousb.TransferTypeBulk,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// MockMassStorageDesc returns a non-printer DeviceDesc (ClassMassStorage 0x1234:0x5678).
func MockMassStorageDesc() *gousb.DeviceDesc {
	return &gousb.DeviceDesc{
		Bus:     1,
		Address: 4,
		Vendor:  0x1234,
		Product: 0x5678,
		Configs: map[int]gousb.ConfigDesc{
			1: {
				Number: 1,
				Interfaces: []gousb.InterfaceDesc{
					{
						Number: 0,
						AltSettings: []gousb.InterfaceSetting{
							{
								Class: gousb.ClassMassStorage,
								Endpoints: map[gousb.EndpointAddress]gousb.EndpointDesc{
									gousb.EndpointAddress(1): {
										Number:       1,
										Direction:    gousb.EndpointDirectionOut,
										TransferType: gousb.TransferTypeBulk,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// MockMissingBulkPrinterDesc returns a printer class DeviceDesc missing a Bulk OUT endpoint.
func MockMissingBulkPrinterDesc() *gousb.DeviceDesc {
	return &gousb.DeviceDesc{
		Vendor:  0x04B8,
		Product: 0x0001,
		Configs: map[int]gousb.ConfigDesc{
			1: {
				Number: 1,
				Interfaces: []gousb.InterfaceDesc{
					{
						Number: 0,
						AltSettings: []gousb.InterfaceSetting{
							{
								Class: gousb.ClassPrinter,
								Endpoints: map[gousb.EndpointAddress]gousb.EndpointDesc{
									gousb.EndpointAddress(1): {
										Number:       1,
										Direction:    gousb.EndpointDirectionIn,
										TransferType: gousb.TransferTypeBulk,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// MockMultiInterfacePrinterDesc returns a DeviceDesc where interface 0 is HID and interface 1 is a Printer with Bulk OUT.
func MockMultiInterfacePrinterDesc() *gousb.DeviceDesc {
	return &gousb.DeviceDesc{
		Vendor:  0x04B8,
		Product: 0x0202,
		Configs: map[int]gousb.ConfigDesc{
			1: {
				Number: 1,
				Interfaces: []gousb.InterfaceDesc{
					{
						Number: 0,
						AltSettings: []gousb.InterfaceSetting{
							{
								Class: gousb.ClassHID,
							},
						},
					},
					{
						Number: 1,
						AltSettings: []gousb.InterfaceSetting{
							{
								Class: gousb.ClassPrinter,
								Endpoints: map[gousb.EndpointAddress]gousb.EndpointDesc{
									gousb.EndpointAddress(2): {
										Number:       2,
										Direction:    gousb.EndpointDirectionOut,
										TransferType: gousb.TransferTypeBulk,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
