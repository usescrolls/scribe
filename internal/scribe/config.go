package scribe

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

// Logger is the global structured logger instance
var Logger *slog.Logger

// InitLogger initializes the structured logger
func InitLogger(debug bool) {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewTextHandler(os.Stderr, opts)
	Logger = slog.New(handler)
	slog.SetDefault(Logger)
}

// GetIcon returns the embedded icon PNG
func GetIcon() []byte {
	return iconData
}
