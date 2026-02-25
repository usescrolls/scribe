package scribe

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// DetectInstallMethod returns how Scribe was installed.
// Possible values: "homebrew", "app-bundle", "binary", "dev", "unknown".
func DetectInstallMethod() string {
	if Version == "dev" || strings.HasSuffix(Version, "-dev") {
		return "dev"
	}

	execPath, err := os.Executable()
	if err != nil {
		return "unknown"
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return "unknown"
	}

	return classifyInstallPath(execPath)
}

// classifyInstallPath determines the install method from a resolved executable path.
func classifyInstallPath(execPath string) string {
	lower := strings.ToLower(execPath)

	// Homebrew formula: binary lives inside /Cellar/
	if strings.Contains(lower, "/cellar/") {
		return "homebrew"
	}

	// Homebrew cask: binary lives inside /Caskroom/
	if strings.Contains(lower, "/caskroom/") {
		return "homebrew"
	}

	// macOS .app bundle
	if strings.Contains(execPath, ".app/Contents/MacOS/") {
		// Check if Homebrew manages this .app via a cask
		if isHomebrewCask() {
			return "homebrew"
		}
		return "app-bundle"
	}

	return "binary"
}

// isHomebrewCask checks if a Homebrew Caskroom entry exists for scribe.
func isHomebrewCask() bool {
	caskDirs := []string{
		"/opt/homebrew/Caskroom/scribe",
		"/usr/local/Caskroom/scribe",
	}
	for _, d := range caskDirs {
		if _, err := os.Stat(d); err == nil {
			return true
		}
	}
	return false
}

// expectedAssetName returns the GitHub release asset filename for the current platform.
// Returns "" if the platform is unsupported.
func expectedAssetName() string {
	switch runtime.GOOS {
	case "linux":
		if runtime.GOARCH == "amd64" {
			return "scribe-linux-amd64"
		}
	case "windows":
		if runtime.GOARCH == "amd64" {
			return "scribe-windows-amd64.exe"
		}
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "scribe-darwin-arm64"
		}
		if runtime.GOARCH == "amd64" {
			return "scribe-darwin-amd64"
		}
	}
	return ""
}

// SelfUpdate downloads and installs the latest version of the Scribe binary.
// Pass "" for baseURL to use the production GitHub API.
func SelfUpdate(baseURL string) (*SelfUpdateResult, error) {
	method := DetectInstallMethod()

	switch method {
	case "homebrew":
		return nil, fmt.Errorf("scribe was installed via Homebrew; run 'brew upgrade usescrolls/tap/scribe' instead")
	case "dev":
		return nil, fmt.Errorf("cannot upgrade development builds")
	case "unknown":
		return nil, fmt.Errorf("cannot determine installation method; upgrade manually")
	}

	// Fetch latest release
	release, err := fetchLatestRelease(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(Version, "v")

	if compareSemver(current, latest) >= 0 {
		return &SelfUpdateResult{
			Updated:       false,
			OldVersion:    Version,
			NewVersion:    release.TagName,
			InstallMethod: method,
			Message:       "already up to date",
		}, nil
	}

	// Find matching asset
	assetName := expectedAssetName()
	if assetName == "" {
		return nil, fmt.Errorf("unsupported platform: %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		return nil, fmt.Errorf("release %s does not contain asset %q", release.TagName, assetName)
	}

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("cannot determine executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve executable path: %w", err)
	}

	// Download to temp file in same directory (same filesystem for atomic rename)
	execDir := filepath.Dir(execPath)
	tmpFile, err := os.CreateTemp(execDir, ".scribe-upgrade-*")
	if err != nil {
		return nil, fmt.Errorf("cannot write to %s: %w (try running with appropriate permissions)", execDir, err)
	}
	tmpPath := tmpFile.Name()
	// Clean up temp file on any error path
	success := false
	defer func() {
		if !success {
			_ = os.Remove(tmpPath)
		}
	}()

	// Download the asset
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(downloadURL) //nolint:gosec // URL comes from GitHub API response
	if err != nil {
		_ = tmpFile.Close()
		return nil, fmt.Errorf("failed to download update: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		_ = tmpFile.Close()
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		_ = tmpFile.Close()
		return nil, fmt.Errorf("failed to save update: %w", err)
	}
	_ = tmpFile.Close()

	// Make executable
	if err := os.Chmod(tmpPath, 0o755); err != nil {
		return nil, fmt.Errorf("failed to set permissions: %w", err)
	}

	// Replace the current binary
	if err := replaceBinary(execPath, tmpPath); err != nil {
		return nil, err
	}

	success = true
	return &SelfUpdateResult{
		Updated:       true,
		OldVersion:    Version,
		NewVersion:    release.TagName,
		InstallMethod: method,
		Message:       fmt.Sprintf("upgraded from %s to %s", Version, release.TagName),
	}, nil
}

// replaceBinary atomically replaces the binary at execPath with the new binary at tmpPath.
func replaceBinary(execPath, tmpPath string) error {
	if runtime.GOOS == "windows" {
		// Windows cannot rename over a running executable; move the old one aside first
		oldPath := execPath + ".old"
		_ = os.Remove(oldPath)
		if err := os.Rename(execPath, oldPath); err != nil {
			return fmt.Errorf("failed to move old binary: %w", err)
		}
		if err := os.Rename(tmpPath, execPath); err != nil {
			// Try to restore the old binary
			_ = os.Rename(oldPath, execPath)
			return fmt.Errorf("failed to install new binary: %w", err)
		}
		// Clean up old binary (best-effort, may still be locked)
		_ = os.Remove(oldPath)
		return nil
	}

	// Unix: atomic rename
	if err := os.Rename(tmpPath, execPath); err != nil {
		return fmt.Errorf("failed to replace binary: %w (try running with appropriate permissions)", err)
	}
	return nil
}
