package main

import (
	"flag"
	"os"
	"runtime"
	"strings"

	"github.com/usescrolls/scribe/internal/scribe"
)

// server is the global server instance
var server *scribe.Server

func main() {
	noGui := flag.Bool("no-gui", false, "Run without system tray icon")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	// Initialize logger
	scribe.InitLogger(*debug)

	scribe.Logger.Info("initializing scribe", "version", scribe.Version, "debug", *debug)

	server = scribe.NewServer()

	// Check for URL scheme argument (passed by OS when opening agenthub:// links)
	args := flag.Args()
	if len(args) > 0 && strings.HasPrefix(args[0], "agenthub://") {
		urlArg := args[0]
		scribe.Logger.Info("URL scheme argument detected", "url", urlArg)

		// On Linux/Windows: try to send URL to running instance before full startup.
		// This allows the new process to exit quickly if another instance handles the URL.
		// macOS doesn't need this - Apple Events handles the "already running" case natively.
		if runtime.GOOS != "darwin" && TrySendToRunningInstance != nil {
			if TrySendToRunningInstance(urlArg) {
				scribe.Logger.Info("URL forwarded to running instance, exiting")
				return
			}
			scribe.Logger.Debug("no running instance found, continuing startup")
		}

		// Initialize directory structure first
		if err := server.Initialize(); err != nil {
			scribe.Logger.Error("failed to initialize directory structure", "error", err)
			os.Exit(1)
		}

		// Load existing registry
		if err := server.Load(); err != nil {
			scribe.Logger.Warn("failed to load existing registry", "error", err)
		}

		// Handle the URL scheme action
		server.HandleURLScheme(urlArg)

		// Regenerate marketplace after any changes
		if err := server.GenerateMarketplace(); err != nil {
			scribe.Logger.Warn("failed to generate marketplace", "error", err)
		}

		scribe.Logger.Info("URL scheme handling completed")
		return
	}

	// Initialize directory structure
	if err := server.Initialize(); err != nil {
		scribe.Logger.Error("failed to initialize directory structure", "error", err)
		os.Exit(1)
	}

	// Migrate from old format if needed
	if err := server.Migrate(); err != nil {
		scribe.Logger.Warn("migration failed", "error", err)
	}

	// Load existing registry
	if err := server.Load(); err != nil {
		scribe.Logger.Warn("failed to load existing registry", "error", err)
	}

	// Generate initial marketplace.json
	if err := server.GenerateMarketplace(); err != nil {
		scribe.Logger.Warn("failed to generate initial marketplace", "error", err)
	}

	if *noGui {
		scribe.Logger.Info("running in headless mode")
		// Register URL scheme handler (IPC server on Linux/Windows)
		RegisterURLSchemeHandler()
		// Block forever - IPC server handles URL scheme requests
		select {}
	} else {
		scribe.Logger.Info("running with system tray")
		// Run with system tray (or headless fallback if CGO disabled)
		RunWithGUI()
	}
}
