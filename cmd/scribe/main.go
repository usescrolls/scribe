package main

import (
	"flag"
	"os"
	"runtime"
	"strings"
)

func main() {
	noGui := flag.Bool("no-gui", false, "Run without system tray icon")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	// Initialize logger
	initLogger(*debug)

	logger.Info("initializing scribe", "version", Version, "debug", *debug)

	server = NewServer()

	// Check for URL scheme argument (passed by OS when opening agenthub:// links)
	args := flag.Args()
	if len(args) > 0 && strings.HasPrefix(args[0], "agenthub://") {
		urlArg := args[0]
		logger.Info("URL scheme argument detected", "url", urlArg)

		// On Linux/Windows: try to send URL to running instance before full startup.
		// This allows the new process to exit quickly if another instance handles the URL.
		// macOS doesn't need this - Apple Events handles the "already running" case natively.
		if runtime.GOOS != "darwin" && TrySendToRunningInstance != nil {
			if TrySendToRunningInstance(urlArg) {
				logger.Info("URL forwarded to running instance, exiting")
				return
			}
			logger.Debug("no running instance found, continuing startup")
		}

		// Initialize directory structure first
		if err := server.Initialize(); err != nil {
			logger.Error("failed to initialize directory structure", "error", err)
			os.Exit(1)
		}

		// Load existing registry
		if err := server.Load(); err != nil {
			logger.Warn("failed to load existing registry", "error", err)
		}

		// Handle the URL scheme action
		handleURLScheme(urlArg)

		// Regenerate marketplace after any changes
		if err := server.GenerateMarketplace(); err != nil {
			logger.Warn("failed to generate marketplace", "error", err)
		}

		logger.Info("URL scheme handling completed")
		return
	}

	// Initialize directory structure
	if err := server.Initialize(); err != nil {
		logger.Error("failed to initialize directory structure", "error", err)
		os.Exit(1)
	}

	// Migrate from old format if needed
	if err := server.Migrate(); err != nil {
		logger.Warn("migration failed", "error", err)
	}

	// Load existing registry
	if err := server.Load(); err != nil {
		logger.Warn("failed to load existing registry", "error", err)
	}

	// Generate initial marketplace.json
	if err := server.GenerateMarketplace(); err != nil {
		logger.Warn("failed to generate initial marketplace", "error", err)
	}

	if *noGui {
		logger.Info("running in headless mode")
		// Register URL scheme handler (IPC server on Linux/Windows)
		RegisterURLSchemeHandler()
		// Block forever - IPC server handles URL scheme requests
		select {}
	} else {
		logger.Info("running with system tray")
		// Run with system tray (or headless fallback if CGO disabled)
		RunWithGUI()
	}
}
