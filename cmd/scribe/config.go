package main

import (
	_ "embed"
	"log/slog"
	"os"
)

//go:embed icon.png
var iconData []byte

const (
	MarketplaceName  = "scribe"
	MarketplaceOwner = "Scribe"
	Version          = "1.0.0"

	// Directory structure constants
	HubDirName         = ".scribe"
	MarketplaceDirName = ".claude-plugin"
	MarketplaceFile    = "marketplace.json"
	PluginsDirName     = "plugins"
	DataDirName        = "data"
	RegistryFile       = "registry.json"
	OldPluginsFile     = "plugins.json" // For migration
)

// logger is the global structured logger instance
var logger *slog.Logger

// initLogger initializes the structured logger
func initLogger(debug bool) {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewTextHandler(os.Stderr, opts)
	logger = slog.New(handler)
	slog.SetDefault(logger)
}

// getIcon returns the embedded icon PNG
func getIcon() []byte {
	return iconData
}
