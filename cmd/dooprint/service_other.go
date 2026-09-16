//go:build !windows

package main

// runAsService is only meaningful on Windows; elsewhere systemd runs the plain process.
func runAsService(port int) bool { return false }
