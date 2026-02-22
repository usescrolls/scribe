package scribe

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestClassifyInstallPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		// Homebrew formula (Intel)
		{"/usr/local/Cellar/scribe/1.0.0/bin/scribe", "homebrew"},
		// Homebrew formula (Apple Silicon)
		{"/opt/homebrew/Cellar/scribe/1.0.0/bin/scribe", "homebrew"},
		// Homebrew cask (direct Caskroom path)
		{"/opt/homebrew/Caskroom/scribe/1.0.0/Scribe.app/Contents/MacOS/scribe", "homebrew"},
		{"/usr/local/Caskroom/scribe/1.0.0/Scribe.app/Contents/MacOS/scribe", "homebrew"},
		// macOS .app bundle (manual drag-and-drop, no Caskroom)
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
		if runtime.GOARCH == "amd64" && name != "scribe-linux-amd64" {
			t.Errorf("expectedAssetName() = %q, want %q", name, "scribe-linux-amd64")
		}
	case "darwin":
		if runtime.GOARCH == "arm64" && name != "scribe-darwin-arm64" {
			t.Errorf("expectedAssetName() = %q, want %q", name, "scribe-darwin-arm64")
		}
		if runtime.GOARCH == "amd64" && name != "scribe-darwin-amd64" {
			t.Errorf("expectedAssetName() = %q, want %q", name, "scribe-darwin-amd64")
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

// newSelfUpdateServer creates a test server that returns a release with one asset.
func newSelfUpdateServer(t *testing.T, tagName, assetName string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/usescrolls/scribe/releases/latest" {
			http.NotFound(w, r)
			return
		}
		release := ghRelease{
			TagName: tagName,
			HTMLURL: "https://github.com/usescrolls/scribe/releases/tag/" + tagName,
			Assets: []ghReleaseAsset{
				{Name: assetName, BrowserDownloadURL: "http://" + r.Host + "/download/" + assetName},
			},
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(release)
	}))
}
