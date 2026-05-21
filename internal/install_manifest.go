package scribe

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	InstallManifestFileName = "install.json"

	InstallComponentCLI     = "cli"
	InstallComponentDesktop = "desktop"
	InstallComponentBinary  = "binary"
)

// InstallManifest records where the installer placed Scribe components.
// It lets the CLI and desktop app upgrade the installed product as one unit.
type InstallManifest struct {
	Version             string   `json:"version"`
	Channel             string   `json:"channel,omitempty"`
	CLIPath             string   `json:"cliPath,omitempty"`
	DesktopPath         string   `json:"desktopPath,omitempty"`
	AppBundlePath       string   `json:"appBundlePath,omitempty"`
	InstalledComponents []string `json:"installedComponents"`
	PublicDownloadBase  string   `json:"publicDownloadBase,omitempty"`
}

func GetInstallManifestPath() (string, error) {
	scribeDir, err := GetScribeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(scribeDir, InstallManifestFileName), nil
}

func ReadInstallManifest() (*InstallManifest, error) {
	path, err := GetInstallManifestPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var manifest InstallManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse install manifest: %w", err)
	}
	return &manifest, nil
}

func WriteInstallManifest(manifest *InstallManifest) error {
	if manifest == nil {
		return errors.New("install manifest is nil")
	}
	if err := EnsureScribeDirs(); err != nil {
		return err
	}

	path, err := GetInstallManifestPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode install manifest: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write install manifest: %w", err)
	}
	return nil
}

func DefaultCLIPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return defaultCLIPathFor(home, runtime.GOOS), nil
}

func defaultCLIPathFor(home, goos string) string {
	if goos == "windows" {
		return filepath.Join(home, "AppData", "Local", "Scribe", "scribe.exe")
	}
	return filepath.Join(home, ".local", "bin", "scribe")
}

func DefaultAppBundlePath() (string, error) {
	if runtime.GOOS != "darwin" {
		return "", nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return defaultAppBundlePathFor(home, runtime.GOOS), nil
}

func defaultAppBundlePathFor(home, goos string) string {
	if goos != "darwin" {
		return ""
	}
	return filepath.Join(home, "Applications", "Scribe.app")
}

func DefaultDesktopPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return defaultDesktopPathFor(home, runtime.GOOS), nil
}

func defaultDesktopPathFor(home, goos string) string {
	switch goos {
	case "darwin":
		appBundle := defaultAppBundlePathFor(home, goos)
		return filepath.Join(appBundle, "Contents", "MacOS", "scribe-desktop")
	case "linux":
		return filepath.Join(home, ".local", "lib", "scribe", "scribe-desktop")
	default:
		return ""
	}
}

func DefaultSplitInstallManifest(version string) (*InstallManifest, error) {
	cliPath, err := DefaultCLIPath()
	if err != nil {
		return nil, err
	}
	desktopPath, err := DefaultDesktopPath()
	if err != nil {
		return nil, err
	}
	appBundlePath, err := DefaultAppBundlePath()
	if err != nil {
		return nil, err
	}

	return &InstallManifest{
		Version:             version,
		Channel:             "stable",
		CLIPath:             cliPath,
		DesktopPath:         desktopPath,
		AppBundlePath:       appBundlePath,
		InstalledComponents: defaultInstalledComponentsFor(runtime.GOOS, desktopPath),
		PublicDownloadBase:  PublicDownloadBase,
	}, nil
}

func defaultInstalledComponentsFor(goos, desktopPath string) []string {
	if goos == "windows" {
		return []string{InstallComponentBinary}
	}
	components := []string{InstallComponentCLI}
	if desktopPath != "" {
		components = append(components, InstallComponentDesktop)
	}
	return components
}

func componentInstalled(manifest *InstallManifest, component string) bool {
	if manifest == nil {
		return false
	}
	for _, installed := range manifest.InstalledComponents {
		if installed == component {
			return true
		}
	}
	return false
}
