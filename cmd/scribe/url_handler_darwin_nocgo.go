//go:build darwin && !cgo

package main

import "github.com/usescrolls/scribe/internal/scribe"

func init() {
	// When CGO is disabled on macOS, URL scheme handling is not available.
	// The app will still work but won't receive agenthub:// URLs while running.
	TrySendToRunningInstance = func(url string) bool { return false }
	CleanupIPC = func() {}
}

// RegisterURLSchemeHandler is a no-op when CGO is disabled on macOS.
// URL scheme handling requires Objective-C/Cocoa which needs CGO.
func RegisterURLSchemeHandler() {
	scribe.Logger.Warn("URL scheme handler unavailable (CGO disabled)")
}
