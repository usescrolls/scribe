package scribe

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestClassifyInstallPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		// macOS .app bundle
		{"/Applications/Scribe.app/Contents/MacOS/scribe", "app-bundle"},
		{"/Users/me/Applications/Scribe.app/Contents/MacOS/scribe", "app-bundle"},
		// Standalone binary
		{"/usr/local/bin/scribe", "binary"},
		{"/home/user/.local/bin/scribe", "binary"},
		{"/home/user/scribe", "binary"},
		{`C:\Users\me\scribe.exe`, "binary"},
		{`C:\Program Files\Scribe\scribe.exe`, "binary"},
	}

	for _, tt := range tests {
		got := classifyInstallPath(tt.path)
		if got != tt.want {
			t.Errorf("classifyInstallPath(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestDetectInstallMethod_Dev(t *testing.T) {
	tests := []struct {
		version string
	}{
		{"dev"},
		{"1.17.0-dev"},
		{"0.0.0-dev"},
	}
	for _, tt := range tests {
		old := Version
		Version = tt.version
		if got := DetectInstallMethod(); got != "dev" {
			t.Errorf("DetectInstallMethod() with Version=%q = %q, want %q", tt.version, got, "dev")
		}
		Version = old
	}
}

func TestExpectedAssetName(t *testing.T) {
	name := expectedAssetName()

	switch runtime.GOOS {
	case "linux":
		if runtime.GOARCH == "amd64" && name != "scribe-desktop-linux-amd64" {
			t.Errorf("expectedAssetName() = %q, want %q", name, "scribe-desktop-linux-amd64")
		}
	case "darwin":
		if runtime.GOARCH == "arm64" && name != "scribe-desktop-darwin-arm64" {
			t.Errorf("expectedAssetName() = %q, want %q", name, "scribe-desktop-darwin-arm64")
		}
		if runtime.GOARCH == "amd64" && name != "scribe-desktop-darwin-amd64" {
			t.Errorf("expectedAssetName() = %q, want %q", name, "scribe-desktop-darwin-amd64")
		}
	case "windows":
		if runtime.GOARCH == "amd64" && name != "scribe-windows-amd64.exe" {
			t.Errorf("expectedAssetName() = %q, want %q", name, "scribe-windows-amd64.exe")
		}
	}

	// Just ensure it's non-empty on common platforms
	if (runtime.GOOS == "linux" || runtime.GOOS == "darwin" || runtime.GOOS == "windows") && runtime.GOARCH == "amd64" {
		if name == "" {
			t.Error("expectedAssetName() returned empty for a supported platform")
		}
	}
}

func TestExpectedComponentAssetNames(t *testing.T) {
	names, err := expectedComponentAssetNames()
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		if err != nil {
			t.Fatalf("expectedComponentAssetNames() unexpected error: %v", err)
		}
	}

	switch runtime.GOOS {
	case "linux", "darwin":
		if names[InstallComponentCLI] == "" {
			t.Error("missing CLI asset name")
		}
		if names[InstallComponentDesktop] == "" {
			t.Error("missing desktop asset name")
		}
	case "windows":
		if names[InstallComponentBinary] == "" {
			t.Error("missing Windows binary asset name")
		}
	}
}

func TestSelfUpdate_DevBuild(t *testing.T) {
	old := Version
	Version = "dev"
	defer func() { Version = old }()

	_, err := SelfUpdate("")
	if err == nil {
		t.Fatal("expected error for dev build")
	}
	if got := err.Error(); got != "cannot upgrade development builds" {
		t.Errorf("unexpected error: %q", got)
	}
}

func TestSelfUpdate_AlreadyUpToDate(t *testing.T) {
	setSelfUpdateTestHome(t)

	assetName := expectedAssetName()
	if assetName == "" {
		t.Skip("unsupported platform for self-update test")
	}

	srv := newSelfUpdateServer(t, "v1.0.0", assetName)
	defer srv.Close()

	old := Version
	Version = "1.0.0"
	defer func() { Version = old }()

	result, err := SelfUpdate(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Updated {
		t.Error("expected Updated to be false")
	}
	if result.Message != "already up to date" {
		t.Errorf("Message = %q, want %q", result.Message, "already up to date")
	}
}

func TestSelfUpdate_MissingAsset(t *testing.T) {
	setSelfUpdateTestHome(t)

	// Server returns a release with no matching asset
	srv := newSelfUpdateServer(t, "v99.0.0", "scribe-nonexistent-platform")
	defer srv.Close()

	old := Version
	Version = "1.0.0"
	defer func() { Version = old }()

	_, err := SelfUpdate(srv.URL)
	if err == nil {
		t.Fatal("expected error for missing asset")
	}
}

func TestSelfUpdate_PairedComponentsHappyPath(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("paired CLI/Desktop self-update only applies to Linux and macOS")
	}
	setSelfUpdateTestHome(t)

	names, err := expectedComponentAssetNames()
	if err != nil {
		t.Skipf("unsupported platform for paired update test: %v", err)
	}

	installDir := t.TempDir()
	cliPath := filepath.Join(installDir, "scribe")
	desktopPath := filepath.Join(installDir, "scribe-desktop")
	if err := os.WriteFile(cliPath, []byte("old cli"), 0o755); err != nil {
		t.Fatalf("failed to write CLI binary: %v", err)
	}
	if err := os.WriteFile(desktopPath, []byte("old desktop"), 0o755); err != nil {
		t.Fatalf("failed to write desktop binary: %v", err)
	}

	srv := newSelfUpdateServerWithAssets(t, "v2.0.0", map[string]string{
		names[InstallComponentCLI]:     "new cli",
		names[InstallComponentDesktop]: "new desktop",
	}, nil)
	defer srv.Close()
	writeTestInstallManifest(t, "1.0.0", cliPath, desktopPath, srv.URL)

	old := Version
	Version = "9.0.0"
	defer func() { Version = old }()

	result, err := SelfUpdate("")
	if err != nil {
		t.Fatalf("SelfUpdate() error = %v", err)
	}
	if !result.Updated {
		t.Fatal("expected update to run")
	}
	if result.OldVersion != "1.0.0" {
		t.Fatalf("OldVersion = %q, want manifest version 1.0.0", result.OldVersion)
	}
	if result.NewVersion != "v2.0.0" {
		t.Fatalf("NewVersion = %q, want v2.0.0", result.NewVersion)
	}
	assertContainsAllComponents(t, result.Components, []string{InstallComponentCLI, InstallComponentDesktop})
	assertFileContent(t, cliPath, "new cli")
	assertFileContent(t, desktopPath, "new desktop")

	manifest, err := ReadInstallManifest()
	if err != nil {
		t.Fatalf("ReadInstallManifest() error = %v", err)
	}
	if manifest.Version != "2.0.0" {
		t.Fatalf("manifest Version = %q, want 2.0.0", manifest.Version)
	}
	if manifest.PublicDownloadBase != srv.URL {
		t.Fatalf("manifest PublicDownloadBase = %q, want %q", manifest.PublicDownloadBase, srv.URL)
	}
}

func TestSelfUpdate_PairedComponentsCreatesMissingDesktopTarget(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("paired CLI/Desktop self-update only applies to Linux and macOS")
	}
	setSelfUpdateTestHome(t)

	names, err := expectedComponentAssetNames()
	if err != nil {
		t.Skipf("unsupported platform for paired update test: %v", err)
	}

	installDir := t.TempDir()
	cliPath := filepath.Join(installDir, "scribe")
	desktopPath := filepath.Join(installDir, "missing-desktop-dir", "scribe-desktop")
	if err := os.WriteFile(cliPath, []byte("old combined binary"), 0o755); err != nil {
		t.Fatalf("failed to write legacy binary: %v", err)
	}

	srv := newSelfUpdateServerWithAssets(t, "v2.0.0", map[string]string{
		names[InstallComponentCLI]:     "new cli",
		names[InstallComponentDesktop]: "new desktop",
	}, nil)
	defer srv.Close()
	writeTestInstallManifest(t, "1.0.0", cliPath, desktopPath, srv.URL)

	old := Version
	Version = "9.0.0"
	defer func() { Version = old }()

	result, err := SelfUpdate("")
	if err != nil {
		t.Fatalf("SelfUpdate() error = %v", err)
	}
	if !result.Updated {
		t.Fatal("expected update to run")
	}
	assertContainsAllComponents(t, result.Components, []string{InstallComponentCLI, InstallComponentDesktop})
	assertFileContent(t, cliPath, "new cli")
	assertFileContent(t, desktopPath, "new desktop")
	assertNoUpgradeTemps(t, installDir, filepath.Dir(desktopPath))
}

func TestSelfUpdate_PairedDownloadFailureDoesNotReplace(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("paired CLI/Desktop self-update only applies to Linux and macOS")
	}
	setSelfUpdateTestHome(t)

	names, err := expectedComponentAssetNames()
	if err != nil {
		t.Skipf("unsupported platform for paired update test: %v", err)
	}

	installDir := t.TempDir()
	cliPath := filepath.Join(installDir, "scribe")
	desktopPath := filepath.Join(installDir, "scribe-desktop")
	if err := os.WriteFile(cliPath, []byte("old cli"), 0o755); err != nil {
		t.Fatalf("failed to write CLI binary: %v", err)
	}
	if err := os.WriteFile(desktopPath, []byte("old desktop"), 0o755); err != nil {
		t.Fatalf("failed to write desktop binary: %v", err)
	}

	srv := newSelfUpdateServerWithAssets(t, "v2.0.0", map[string]string{
		names[InstallComponentCLI]:     "new cli",
		names[InstallComponentDesktop]: "new desktop",
	}, map[string]int{
		names[InstallComponentDesktop]: http.StatusInternalServerError,
	})
	defer srv.Close()
	writeTestInstallManifest(t, "1.0.0", cliPath, desktopPath, srv.URL)

	old := Version
	Version = "9.0.0"
	defer func() { Version = old }()

	_, err = SelfUpdate("")
	if err == nil {
		t.Fatal("expected download failure")
	}
	assertFileContent(t, cliPath, "old cli")
	assertFileContent(t, desktopPath, "old desktop")
	assertNoUpgradeTemps(t, installDir)
}

func TestReplaceBinary(t *testing.T) {
	tmpDir := t.TempDir()
	fakeBinary := []byte("#!/bin/sh\necho upgraded\n")

	fakeExec := filepath.Join(tmpDir, "scribe")
	if runtime.GOOS == "windows" {
		fakeExec = filepath.Join(tmpDir, "scribe.exe")
	}
	if err := os.WriteFile(fakeExec, []byte("old binary"), 0o755); err != nil {
		t.Fatalf("failed to create fake executable: %v", err)
	}

	tmpNew := filepath.Join(tmpDir, ".scribe-upgrade-test")
	if err := os.WriteFile(tmpNew, fakeBinary, 0o755); err != nil {
		t.Fatalf("failed to create temp binary: %v", err)
	}

	if err := replaceBinary(fakeExec, tmpNew); err != nil {
		t.Fatalf("replaceBinary failed: %v", err)
	}

	// Verify the binary was replaced
	content, err := os.ReadFile(fakeExec)
	if err != nil {
		t.Fatalf("failed to read replaced binary: %v", err)
	}
	if string(content) != string(fakeBinary) {
		t.Errorf("binary content = %q, want %q", string(content), string(fakeBinary))
	}

	// Verify the temp file was removed (renamed into place)
	if _, err := os.Stat(tmpNew); !os.IsNotExist(err) {
		t.Error("expected temp file to be removed after rename")
	}
}

func TestReplaceUpdateAssets_RollsBackCompletedComponents(t *testing.T) {
	tmpDir := t.TempDir()
	cliPath := filepath.Join(tmpDir, "scribe")
	desktopPath := filepath.Join(tmpDir, "scribe-desktop")
	cliStagedPath := filepath.Join(tmpDir, ".scribe-upgrade-cli")
	missingDesktopStagedPath := filepath.Join(tmpDir, ".scribe-upgrade-desktop-missing")

	if err := os.WriteFile(cliPath, []byte("old cli"), 0o755); err != nil {
		t.Fatalf("failed to write CLI binary: %v", err)
	}
	if err := os.WriteFile(desktopPath, []byte("old desktop"), 0o755); err != nil {
		t.Fatalf("failed to write desktop binary: %v", err)
	}
	if err := os.WriteFile(cliStagedPath, []byte("new cli"), 0o755); err != nil {
		t.Fatalf("failed to write staged CLI binary: %v", err)
	}

	err := replaceUpdateAssets([]updateAsset{
		{
			Component:   InstallComponentCLI,
			DownloadURL: cliStagedPath,
			TargetPath:  cliPath,
		},
		{
			Component:   InstallComponentDesktop,
			DownloadURL: missingDesktopStagedPath,
			TargetPath:  desktopPath,
		},
	})
	if err == nil {
		t.Fatal("expected replace failure")
	}
	assertFileContent(t, cliPath, "old cli")
	assertFileContent(t, desktopPath, "old desktop")
	assertNoUpgradeTemps(t, tmpDir)
}

func TestReplaceUpdateAssets_RollsBackCreatedComponent(t *testing.T) {
	tmpDir := t.TempDir()
	cliPath := filepath.Join(tmpDir, "new-dir", "scribe")
	desktopPath := filepath.Join(tmpDir, "scribe-desktop")
	cliStagedPath := filepath.Join(tmpDir, ".scribe-upgrade-cli")
	missingDesktopStagedPath := filepath.Join(tmpDir, ".scribe-upgrade-desktop-missing")

	if err := os.WriteFile(desktopPath, []byte("old desktop"), 0o755); err != nil {
		t.Fatalf("failed to write desktop binary: %v", err)
	}
	if err := os.WriteFile(cliStagedPath, []byte("new cli"), 0o755); err != nil {
		t.Fatalf("failed to write staged CLI binary: %v", err)
	}

	err := replaceUpdateAssets([]updateAsset{
		{
			Component:   InstallComponentCLI,
			DownloadURL: cliStagedPath,
			TargetPath:  cliPath,
		},
		{
			Component:   InstallComponentDesktop,
			DownloadURL: missingDesktopStagedPath,
			TargetPath:  desktopPath,
		},
	})
	if err == nil {
		t.Fatal("expected replace failure")
	}
	if _, err := os.Stat(cliPath); !os.IsNotExist(err) {
		t.Fatalf("created CLI target still exists after rollback: %v", err)
	}
	assertFileContent(t, desktopPath, "old desktop")
	assertNoUpgradeTemps(t, tmpDir, filepath.Dir(cliPath))
}

func TestReplaceBinary_InvalidPath(t *testing.T) {
	tmpDir := t.TempDir()
	fakeExec := filepath.Join(tmpDir, "scribe")
	if err := os.WriteFile(fakeExec, []byte("old"), 0o755); err != nil {
		t.Fatalf("failed to create fake executable: %v", err)
	}

	// Replacing with a non-existent temp file should fail
	err := replaceBinary(fakeExec, filepath.Join(tmpDir, "nonexistent"))
	if err == nil {
		t.Error("expected error when temp file does not exist")
	}
}

func TestSelfUpdate_DevSuffixBuild(t *testing.T) {
	old := Version
	Version = "1.17.0-dev"
	defer func() { Version = old }()

	_, err := SelfUpdate("")
	if err == nil {
		t.Fatal("expected error for dev-suffixed build")
	}
	if got := err.Error(); got != "cannot upgrade development builds" {
		t.Errorf("unexpected error: %q", got)
	}
}

func setSelfUpdateTestHome(t *testing.T) {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)
}

func writeTestInstallManifest(t *testing.T, version, cliPath, desktopPath, publicDownloadBase string) {
	t.Helper()
	if err := WriteInstallManifest(&InstallManifest{
		Version:             version,
		Channel:             "stable",
		CLIPath:             cliPath,
		DesktopPath:         desktopPath,
		InstalledComponents: []string{InstallComponentCLI, InstallComponentDesktop},
		PublicDownloadBase:  publicDownloadBase,
	}); err != nil {
		t.Fatalf("WriteInstallManifest() error = %v", err)
	}
}

// newSelfUpdateServer creates a test server that returns a release with one asset.
func newSelfUpdateServer(t *testing.T, tagName, assetName string) *httptest.Server {
	t.Helper()
	return newSelfUpdateServerWithAssets(t, tagName, map[string]string{assetName: "fake binary"}, nil)
}

func newSelfUpdateServerWithAssets(t *testing.T, tagName string, assets map[string]string, failures map[string]int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Serve download requests
		if strings.HasPrefix(r.URL.Path, "/download/") {
			assetName := strings.TrimPrefix(r.URL.Path, "/download/")
			if statusCode := failures[assetName]; statusCode != 0 {
				w.WriteHeader(statusCode)
				_, _ = w.Write([]byte("download failed"))
				return
			}
			content, ok := assets[assetName]
			if !ok {
				http.NotFound(w, r)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(content))
			return
		}
		releaseAssets := make([]releaseManifestAsset, 0, len(assets))
		for assetName := range assets {
			releaseAssets = append(releaseAssets, releaseManifestAsset{
				Name:               assetName,
				BrowserDownloadURL: "http://" + r.Host + "/download/" + assetName,
			})
		}
		rel := releaseManifest{
			TagName: tagName,
			HTMLURL: "https://gitlab.com/usescrolls/scribe/-/releases/" + tagName,
			Assets:  releaseAssets,
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(rel)
	}))
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	if string(got) != want {
		t.Fatalf("%s content = %q, want %q", path, string(got), want)
	}
}

func assertContainsAllComponents(t *testing.T, got, want []string) {
	t.Helper()
	seen := map[string]bool{}
	for _, component := range got {
		seen[component] = true
	}
	for _, component := range want {
		if !seen[component] {
			t.Fatalf("components = %#v, missing %q", got, component)
		}
	}
}

func assertNoUpgradeTemps(t *testing.T, dirs ...string) {
	t.Helper()
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("failed to read %s: %v", dir, err)
		}
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".scribe-upgrade-") || strings.HasPrefix(entry.Name(), ".scribe-backup-") {
				t.Fatalf("unexpected temporary update file left behind: %s", filepath.Join(dir, entry.Name()))
			}
		}
	}
}
