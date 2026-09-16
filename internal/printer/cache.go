package printer

import (
	"sort"
	"strings"
	"sync"
)

type printerCache struct {
	mu                        sync.RWMutex
	lastSnapshot              string
	cachedPrinters            []Info
	cachedUnavailablePrinters []UnavailableInfo
}

var usbCache = &printerCache{}

func (c *printerCache) HasChanged(keys []string) bool {
	snap := buildSnapshot(keys)
	c.mu.RLock()
	defer c.mu.RUnlock()
	return snap != c.lastSnapshot
}

func (c *printerCache) Get() []Info {
	c.mu.RLock()
	defer c.mu.RUnlock()

	printers := make([]Info, len(c.cachedPrinters))
	copy(printers, c.cachedPrinters)

	return printers
}

func (c *printerCache) Update(keys []string, printers []Info, unavailablePrinters []UnavailableInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lastSnapshot = buildSnapshot(keys)
	c.cachedPrinters = printers
	c.cachedUnavailablePrinters = unavailablePrinters
}

func buildSnapshot(keys []string) string {
	sorted := make([]string, len(keys))
	copy(sorted, keys)
	sort.Strings(sorted)
	return strings.Join(sorted, "|")
}

func (c *printerCache) HasUnavailable() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cachedUnavailablePrinters) > 0
}
