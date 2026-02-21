package scribe

import (
	"io"
	"log/slog"
	"os"
)

// Version is overridden at build time via ldflags
var Version = "dev"

const (
	MarketplaceName  = "scribe"
	MarketplaceOwner = "Scribe"

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

// LogWriter is the global rotating log writer (for cleanup on shutdown)
var LogWriter *RotatingLogWriter

// InitLogger initializes the structured logger for GUI mode.
// Writes to both stderr and ~/.scribe/scribe.log.
func InitLogger(debug bool) {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var output io.Writer = os.Stderr
	lw, err := NewRotatingLogWriter()
	if err == nil {
		LogWriter = lw
		output = io.MultiWriter(os.Stderr, lw)
	}

	handler := slog.NewTextHandler(output, opts)
	Logger = slog.New(handler).With("source", "internal")
	slog.SetDefault(Logger)
}

// InitLoggerCLI initializes a quiet logger for CLI mode.
// Always writes to ~/.scribe/scribe.log at Info level.
// Writes to stderr only when debug=true.
func InitLoggerCLI(debug bool) {
	var output io.Writer = os.Stderr
	opts := &slog.HandlerOptions{
		Level: slog.LevelError + 1, // Suppress all by default
	}

	lw, err := NewRotatingLogWriter()
	if err == nil {
		LogWriter = lw
		if debug {
			opts.Level = slog.LevelDebug
			output = io.MultiWriter(os.Stderr, lw)
		} else {
			opts.Level = slog.LevelInfo
			output = lw
		}
	} else if debug {
		opts.Level = slog.LevelDebug
	}

	handler := slog.NewTextHandler(output, opts)
	Logger = slog.New(handler).With("source", "internal")
	slog.SetDefault(Logger)
}

// CloseLogger closes the log file writer. Call on application shutdown.
func CloseLogger() {
	if LogWriter != nil {
		_ = LogWriter.Close()
	}
}
