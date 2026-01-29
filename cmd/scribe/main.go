//go:build nowails

package main

import (
	"flag"
	"os"
	"runtime"
	"strings"

	"github.com/usescrolls/scribe/cmd/scribe/cli"
	"github.com/usescrolls/scribe/internal/scribe"
)

// server is the global server instance
var server *scribe.Server

func main() {
	// Detection order:
	// 1. agenthub:// URL → URL scheme handler
	// 2. Known CLI command → CLI mode
	// 3. No arguments or --no-gui/--debug → GUI mode

	// Check for CLI command before flag parsing
	if len(os.Args) > 1 {
		firstArg := os.Args[1]

		// 1. Check for URL scheme (existing behavior)
		if strings.HasPrefix(firstArg, "agenthub://") {
			handleURLScheme(firstArg)
			return
		}

		// 2. Check for CLI commands
		if isCLICommand(firstArg) {
			exitCode := cli.Execute()
			os.Exit(exitCode)
		}
	}

	// 3. GUI mode (existing behavior)
	runGUIMode()
}

// isCLICommand checks if the argument is a known CLI command
func isCLICommand(arg string) bool {
	// Skip if it looks like a flag
	if strings.HasPrefix(arg, "-") {
		return false
	}

	commands := cli.CLICommands()
	for _, cmd := range commands {
		if arg == cmd {
			return true
		}
	}
	return false
}

// handleURLScheme handles agenthub:// URLs (existing behavior)
func handleURLScheme(urlArg string) {
	// Initialize with debug=false for URL scheme handling (no way to pass --debug flag)
	scribe.InitLogger(false)
	scribe.Logger.Info("URL scheme argument detected", "url", urlArg)

	server = scribe.NewServer()

	// On Linux/Windows: try to send URL to running instance before full startup.
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
}

// runGUIMode runs the application in GUI mode (existing behavior)
func runGUIMode() {
	noGui := flag.Bool("no-gui", false, "Run without system tray icon")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	// Initialize logger
	scribe.InitLogger(*debug)

	scribe.Logger.Info("initializing scribe", "version", scribe.Version, "debug", *debug)

	server = scribe.NewServer()

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
