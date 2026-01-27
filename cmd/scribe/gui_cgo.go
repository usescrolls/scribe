//go:build cgo

package main

import (
	"fmt"
	"time"

	"github.com/getlantern/systray"
	"github.com/usescrolls/scribe/internal/scribe"
)

// RunWithGUI starts the application with the system tray GUI.
// This version is only compiled when CGO is enabled (native builds).
func RunWithGUI() {
	systray.Run(onReady, onExit)
}

func onReady() {
	scribe.Logger.Debug("initializing system tray")

	// Register URL scheme handler (must be on main thread for macOS)
	// This allows receiving agenthub:// URLs while the app is already running
	RegisterURLSchemeHandler()

	// Set the icon (a simple orange circle as PNG)
	systray.SetIcon(scribe.GetIcon())
	systray.SetTitle("")
	systray.SetTooltip("Scribe")

	// Menu items (display-only, no interaction needed)
	systray.AddMenuItem("Scribe v"+scribe.Version, "").Disable()

	systray.AddSeparator()

	mPlugins := systray.AddMenuItem(fmt.Sprintf("Plugins: %d", server.PluginCount()), "Number of installed plugins")
	mPlugins.Disable()

	systray.AddSeparator()

	// Reset option
	mReset := systray.AddMenuItem("Delete Local Data", "Uninstall plugins, clear settings, and delete data")

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Stop the middleware and quit")

	scribe.Logger.Debug("system tray initialized")

	// Update plugin count periodically and handle menu clicks
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				count := server.PluginCount()
				mPlugins.SetTitle(fmt.Sprintf("Plugins: %d", count))
			case <-mReset.ClickedCh:
				scribe.Logger.Info("reset requested from system tray")
				if err := server.FullReset(); err != nil {
					scribe.Logger.Error("failed to perform reset", "error", err)
				}
			case <-mQuit.ClickedCh:
				scribe.Logger.Info("quit requested from system tray")
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	// Cleanup IPC resources (socket on Linux, mutex/pipe on Windows)
	if CleanupIPC != nil {
		CleanupIPC()
	}
	scribe.Logger.Info("Scribe stopped")
}
