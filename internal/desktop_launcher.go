package scribe

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// LaunchDesktop starts the installed desktop app without importing desktop UI libraries.
func LaunchDesktop() error {
	manifest, err := ReadInstallManifest()
	if err != nil {
		manifest, err = DefaultSplitInstallManifest(Version)
		if err != nil {
			return err
		}
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		appBundle := manifest.AppBundlePath
		if appBundle == "" {
			appBundle, err = DefaultAppBundlePath()
			if err != nil {
				return err
			}
		}
		if _, err := os.Stat(appBundle); err != nil {
			return fmt.Errorf("desktop app is not installed at %s", appBundle)
		}
		cmd = exec.Command("open", appBundle)
	case "linux":
		desktopPath := manifest.DesktopPath
		if desktopPath == "" {
			desktopPath, err = DefaultDesktopPath()
			if err != nil {
				return err
			}
		}
		if _, err := os.Stat(desktopPath); err != nil {
			return fmt.Errorf("desktop app is not installed at %s", desktopPath)
		}
		cmd = exec.Command(desktopPath)
	default:
		return fmt.Errorf("desktop launch is not supported on %s", runtime.GOOS)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch desktop app: %w", err)
	}
	return nil
}
