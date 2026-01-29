//go:build nowails

package main

import (
	"fmt"
	"os"

	"github.com/usescrolls/scribe/internal/scribe"
)

// RunWithGUI is a fallback for CLI-only builds without Wails
func RunWithGUI() {
	scribe.Logger.Info("GUI mode disabled (built with nowails tag)")
	fmt.Fprintln(os.Stderr, "This build does not include GUI support.")
	fmt.Fprintln(os.Stderr, "Use CLI commands: scribe install, scribe list, scribe uninstall")
	os.Exit(1)
}
