//go:build !darwin && !linux && !windows

package main

func init() {
	// On unsupported platforms, these are no-ops
	TrySendToRunningInstance = func(url string) bool { return false }
	CleanupIPC = func() {}
}

// RegisterURLSchemeHandler is a no-op on unsupported platforms
func RegisterURLSchemeHandler() {
	// URL scheme handling not implemented for this platform
}
