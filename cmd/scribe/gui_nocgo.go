//go:build !cgo

package main

import (
	"os"
	"os/signal"
	"syscall"
)

// RunWithGUI falls back to headless mode when CGO is disabled.
// This happens during cross-compilation since systray requires CGO.
func RunWithGUI() {
	logger.Warn("system tray unavailable (CGO disabled), running in headless mode")
	runHeadless()
}

func runHeadless() {
	// Register URL scheme handler (IPC server on Linux/Windows)
	RegisterURLSchemeHandler()

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		logger.Info("shutdown signal received")
		if CleanupIPC != nil {
			CleanupIPC()
		}
		os.Exit(0)
	}()

	// Block forever - IPC server handles URL scheme requests
	select {}
}
