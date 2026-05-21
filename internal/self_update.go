package scribe

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type updateAsset struct {
	Component   string
	Name        string
	DownloadURL string
	TargetPath  string
}

// DetectInstallMethod returns how Scribe was installed.
// Possible values: "app-bundle", "binary", "dev", "unknown".
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
	// macOS .app bundle
	if strings.Contains(execPath, ".app/Contents/MacOS/") {
		return "app-bundle"
	}

	return "binary"
}

// expectedAssetName returns the release asset filename for the current platform.
// Returns "" if the platform is unsupported.
func expectedAssetName() string {
	switch runtime.GOOS {
	case "linux":
		if runtime.GOARCH == "amd64" {
			return "scribe-desktop-linux-amd64"
		}
		if runtime.GOARCH == "arm64" {
			return "scribe-desktop-linux-arm64"
		}
	case "windows":
		if runtime.GOARCH == "amd64" {
			return "scribe-windows-amd64.exe"
		}
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "scribe-desktop-darwin-arm64"
		}
		if runtime.GOARCH == "amd64" {
			return "scribe-desktop-darwin-amd64"
		}
	}
	return ""
}

func expectedComponentAssetNames() (map[string]string, error) {
	switch runtime.GOOS {
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			return map[string]string{
				InstallComponentCLI:     "scribe-cli-linux-amd64",
				InstallComponentDesktop: "scribe-desktop-linux-amd64",
			}, nil
		case "arm64":
			return map[string]string{
				InstallComponentCLI:     "scribe-cli-linux-arm64",
				InstallComponentDesktop: "scribe-desktop-linux-arm64",
			}, nil
		}
	case "darwin":
		switch runtime.GOARCH {
		case "arm64":
			return map[string]string{
				InstallComponentCLI:     "scribe-cli-darwin-arm64",
				InstallComponentDesktop: "scribe-desktop-darwin-arm64",
			}, nil
		case "amd64":
			return map[string]string{
				InstallComponentCLI:     "scribe-cli-darwin-amd64",
				InstallComponentDesktop: "scribe-desktop-darwin-amd64",
			}, nil
		}
	case "windows":
		if runtime.GOARCH == "amd64" {
			return map[string]string{
				InstallComponentBinary: "scribe-windows-amd64.exe",
			}, nil
		}
	}
	return nil, fmt.Errorf("unsupported platform: %s/%s", runtime.GOOS, runtime.GOARCH)
}

// SelfUpdate downloads and installs the latest version of all installed Scribe components.
// Pass "" for baseURL to use PublicDownloadBase.
func SelfUpdate(baseURL string) (*SelfUpdateResult, error) {
	method := DetectInstallMethod()

	switch method {
	case "dev":
		return nil, fmt.Errorf("cannot upgrade development builds")
	case "unknown":
		return nil, fmt.Errorf("cannot determine installation method; upgrade manually")
	}

	manifest, err := installManifestForSelfUpdate()
	if err != nil {
		return nil, err
	}
	if baseURL == "" && manifest.PublicDownloadBase != "" {
		baseURL = manifest.PublicDownloadBase
	}

	// Fetch latest release
	rel, err := fetchLatestRelease(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}

	latest := strings.TrimPrefix(rel.TagName, "v")
	currentVersion := Version
	if manifest.Version != "" {
		currentVersion = manifest.Version
	}
	current := strings.TrimPrefix(currentVersion, "v")

	if compareSemver(current, latest) >= 0 {
		return &SelfUpdateResult{
			Updated:       false,
			OldVersion:    currentVersion,
			NewVersion:    rel.TagName,
			InstallMethod: method,
			Message:       "already up to date",
		}, nil
	}

	updateAssets, err := resolveUpdateAssets(rel, manifest)
	if err != nil {
		return nil, err
	}

	staged, err := downloadUpdateAssets(updateAssets)
	if err != nil {
		return nil, err
	}
	success := false
	defer func() {
		if !success {
			for _, asset := range staged {
				_ = os.Remove(asset.DownloadURL)
			}
		}
	}()

	if err := replaceUpdateAssets(staged); err != nil {
		return nil, err
	}
	success = true

	manifest.Version = strings.TrimPrefix(rel.TagName, "v")
	if baseURL != "" {
		manifest.PublicDownloadBase = strings.TrimSuffix(baseURL, "/")
	}
	_ = WriteInstallManifest(manifest)

	components := make([]string, 0, len(staged))
	for _, asset := range staged {
		components = append(components, asset.Component)
	}

	return &SelfUpdateResult{
		Updated:       true,
		OldVersion:    currentVersion,
		NewVersion:    rel.TagName,
		InstallMethod: method,
		Components:    components,
		Message:       fmt.Sprintf("upgraded from %s to %s", currentVersion, rel.TagName),
	}, nil
}

func installManifestForSelfUpdate() (*InstallManifest, error) {
	manifest, err := ReadInstallManifest()
	if err == nil {
		return manifest, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	manifest, err = DefaultSplitInstallManifest(Version)
	if err != nil {
		return nil, err
	}

	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("cannot determine executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve executable path: %w", err)
	}

	switch runtime.GOOS {
	case "darwin":
		if strings.Contains(execPath, ".app/Contents/MacOS/") {
			manifest.DesktopPath = execPath
		} else {
			manifest.CLIPath = execPath
		}
	case "linux":
		if filepath.Base(execPath) == "scribe-desktop" {
			manifest.DesktopPath = execPath
		} else {
			manifest.CLIPath = execPath
		}
	case "windows":
		manifest.CLIPath = execPath
	}

	return manifest, nil
}

func resolveUpdateAssets(rel *release, manifest *InstallManifest) ([]updateAsset, error) {
	expectedNames, err := expectedComponentAssetNames()
	if err != nil {
		return nil, err
	}

	assetURLs := make(map[string]string, len(rel.Assets))
	for _, asset := range rel.Assets {
		assetURLs[asset.Name] = asset.DownloadURL
	}

	targets := componentTargets(manifest)
	if len(targets) == 0 {
		return nil, fmt.Errorf("cannot determine installed Scribe components")
	}

	updateAssets := make([]updateAsset, 0, len(targets))
	for _, component := range orderedUpdateComponents(targets) {
		targetPath := targets[component]
		assetName := expectedNames[component]
		if assetName == "" {
			return nil, fmt.Errorf("no release asset mapping for %s on %s/%s", component, runtime.GOOS, runtime.GOARCH)
		}
		downloadURL := assetURLs[assetName]
		if downloadURL == "" {
			return nil, fmt.Errorf("release %s does not contain asset %q", rel.TagName, assetName)
		}
		updateAssets = append(updateAssets, updateAsset{
			Component:   component,
			Name:        assetName,
			DownloadURL: downloadURL,
			TargetPath:  targetPath,
		})
	}
	return updateAssets, nil
}

func orderedUpdateComponents(targets map[string]string) []string {
	order := []string{InstallComponentBinary, InstallComponentCLI, InstallComponentDesktop}
	components := make([]string, 0, len(targets))
	for _, component := range order {
		if _, ok := targets[component]; ok {
			components = append(components, component)
		}
	}
	return components
}

func componentTargets(manifest *InstallManifest) map[string]string {
	targets := map[string]string{}
	if runtime.GOOS == "windows" {
		if manifest.CLIPath != "" {
			targets[InstallComponentBinary] = manifest.CLIPath
		}
		return targets
	}

	inferComponents := len(manifest.InstalledComponents) == 0
	if (inferComponents || componentInstalled(manifest, InstallComponentCLI)) && manifest.CLIPath != "" {
		targets[InstallComponentCLI] = manifest.CLIPath
	}
	if (inferComponents || componentInstalled(manifest, InstallComponentDesktop)) && manifest.DesktopPath != "" {
		targets[InstallComponentDesktop] = manifest.DesktopPath
	}
	return targets
}

func downloadUpdateAssets(updateAssets []updateAsset) ([]updateAsset, error) {
	staged := make([]updateAsset, 0, len(updateAssets))
	client := &http.Client{Timeout: 120 * time.Second}

	for _, asset := range updateAssets {
		targetDir := filepath.Dir(asset.TargetPath)
		if err := os.MkdirAll(targetDir, 0o755); err != nil {
			cleanupStagedUpdateAssets(staged)
			return nil, fmt.Errorf("cannot create %s: %w", targetDir, err)
		}
		tmpFile, err := os.CreateTemp(targetDir, ".scribe-upgrade-*")
		if err != nil {
			cleanupStagedUpdateAssets(staged)
			return nil, fmt.Errorf("cannot write to %s: %w (try running with appropriate permissions)", targetDir, err)
		}
		tmpPath := tmpFile.Name()

		resp, err := client.Get(asset.DownloadURL) //nolint:gosec // URL comes from trusted release metadata
		if err != nil {
			_ = tmpFile.Close()
			_ = os.Remove(tmpPath)
			cleanupStagedUpdateAssets(staged)
			return nil, fmt.Errorf("failed to download %s update: %w", asset.Component, err)
		}

		if resp.StatusCode != http.StatusOK {
			_ = tmpFile.Close()
			_ = resp.Body.Close()
			_ = os.Remove(tmpPath)
			cleanupStagedUpdateAssets(staged)
			return nil, fmt.Errorf("%s download failed with status %d", asset.Component, resp.StatusCode)
		}

		if _, err := io.Copy(tmpFile, resp.Body); err != nil {
			_ = tmpFile.Close()
			_ = resp.Body.Close()
			_ = os.Remove(tmpPath)
			cleanupStagedUpdateAssets(staged)
			return nil, fmt.Errorf("failed to save %s update: %w", asset.Component, err)
		}
		_ = resp.Body.Close()
		_ = tmpFile.Close()

		if err := os.Chmod(tmpPath, 0o755); err != nil {
			_ = os.Remove(tmpPath)
			cleanupStagedUpdateAssets(staged)
			return nil, fmt.Errorf("failed to set %s permissions: %w", asset.Component, err)
		}

		asset.DownloadURL = tmpPath
		staged = append(staged, asset)
	}

	return staged, nil
}

func cleanupStagedUpdateAssets(staged []updateAsset) {
	for _, asset := range staged {
		_ = os.Remove(asset.DownloadURL)
	}
}

type replacedUpdateAsset struct {
	Component   string
	TargetPath  string
	BackupPath  string
	HadPrevious bool
}

func replaceUpdateAssets(staged []updateAsset) error {
	replaced := make([]replacedUpdateAsset, 0, len(staged))

	for _, asset := range staged {
		backupPath, hadPrevious, err := replaceWithBackup(asset.TargetPath, asset.DownloadURL)
		if err != nil {
			rollbackReplacedAssets(replaced)
			return fmt.Errorf("failed to replace %s component: %w", asset.Component, err)
		}
		replaced = append(replaced, replacedUpdateAsset{
			Component:   asset.Component,
			TargetPath:  asset.TargetPath,
			BackupPath:  backupPath,
			HadPrevious: hadPrevious,
		})
	}

	for _, asset := range replaced {
		if asset.HadPrevious {
			_ = os.Remove(asset.BackupPath)
		}
	}
	return nil
}

func replaceWithBackup(targetPath, stagedPath string) (backupPath string, hadPrevious bool, err error) {
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", false, fmt.Errorf("failed to prepare target directory: %w", err)
	}

	if _, err := os.Stat(targetPath); err != nil {
		if !os.IsNotExist(err) {
			return "", false, fmt.Errorf("failed to inspect old binary: %w", err)
		}
		if err := os.Rename(stagedPath, targetPath); err != nil {
			return "", false, fmt.Errorf("failed to install new binary: %w", err)
		}
		return "", false, nil
	}

	backupFile, err := os.CreateTemp(targetDir, ".scribe-backup-*")
	if err != nil {
		return "", true, fmt.Errorf("failed to prepare backup: %w", err)
	}
	backupPath = backupFile.Name()
	_ = backupFile.Close()
	_ = os.Remove(backupPath)

	if err := os.Rename(targetPath, backupPath); err != nil {
		return "", true, fmt.Errorf("failed to move old binary: %w", err)
	}
	if err := os.Rename(stagedPath, targetPath); err != nil {
		_ = os.Rename(backupPath, targetPath)
		return "", true, fmt.Errorf("failed to install new binary: %w", err)
	}
	return backupPath, true, nil
}

func rollbackReplacedAssets(replaced []replacedUpdateAsset) {
	for i := len(replaced) - 1; i >= 0; i-- {
		asset := replaced[i]
		_ = os.Remove(asset.TargetPath)
		if asset.HadPrevious {
			_ = os.Rename(asset.BackupPath, asset.TargetPath)
		}
	}
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
