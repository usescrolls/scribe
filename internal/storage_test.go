package scribe

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBoost_GetScribeDir(t *testing.T) {
	tmpDir := setupTempHome(t)
	dir, err := GetScribeDir()
	if err != nil {
		t.Fatalf("GetScribeDir() error: %v", err)
	}
	expected := filepath.Join(tmpDir, ".scribe")
	if dir != expected {
		t.Errorf("GetScribeDir() = %q, want %q", dir, expected)
	}
}

func TestBoost_GetWorkspacesDir(t *testing.T) {
	tmpDir := setupTempHome(t)
	dir, err := GetWorkspacesDir()
	if err != nil {
		t.Fatalf("GetWorkspacesDir() error: %v", err)
	}
	expected := filepath.Join(tmpDir, ".scribe", "workspaces")
	if dir != expected {
		t.Errorf("GetWorkspacesDir() = %q, want %q", dir, expected)
	}
}

func TestBoost_GetConfigPath(t *testing.T) {
	tmpDir := setupTempHome(t)
	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() error: %v", err)
	}
	expected := filepath.Join(tmpDir, ".scribe", "config.json")
	if path != expected {
		t.Errorf("GetConfigPath() = %q, want %q", path, expected)
	}
}

func TestBoost_GetWorkspacePath(t *testing.T) {
	tmpDir := setupTempHome(t)
	path, err := GetWorkspacePath("test-ws")
	if err != nil {
		t.Fatalf("GetWorkspacePath() error: %v", err)
	}
	expected := filepath.Join(tmpDir, ".scribe", "workspaces", "test-ws.json")
	if path != expected {
		t.Errorf("GetWorkspacePath() = %q, want %q", path, expected)
	}
}

func TestBoost_EnsureDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	newDir := filepath.Join(tmpDir, "a", "b", "c")
	err = EnsureDir(newDir)
	if err != nil {
		t.Fatalf("EnsureDir() error: %v", err)
	}
	if !dirExists(newDir) {
		t.Error("directory was not created")
	}

	// Idempotent
	err = EnsureDir(newDir)
	if err != nil {
		t.Fatalf("EnsureDir() second call error: %v", err)
	}
}

func TestBoost_EnsureScribeDirs(t *testing.T) {
	tmpDir := setupTempHome(t)
	err := EnsureScribeDirs()
	if err != nil {
		t.Fatalf("EnsureScribeDirs() error: %v", err)
	}

	// Verify all dirs exist
	dirs := []string{
		filepath.Join(tmpDir, ".scribe"),
		filepath.Join(tmpDir, ".scribe", "scrolls"),
		filepath.Join(tmpDir, ".scribe", "workspaces"),
		filepath.Join(tmpDir, ".scribe", "cache"),
	}
	for _, dir := range dirs {
		if !dirExists(dir) {
			t.Errorf("directory %q was not created", dir)
		}
	}
}

func TestBoost_LoadConfig_Default(t *testing.T) {
	_ = setupTempHome(t)
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}
	if config.ActiveWorkspace != DefaultWorkspaceName {
		t.Errorf("default ActiveWorkspace = %q, want %q", config.ActiveWorkspace, DefaultWorkspaceName)
	}
	if config.OnboardingCompleted {
		t.Error("default OnboardingCompleted should be false")
	}
}

func TestBoost_SaveAndLoadConfig_Roundtrip(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	config := &Config{
		ActiveWorkspace:     "test-ws",
		OnboardingCompleted: true,
	}
	err := SaveConfig(config)
	if err != nil {
		t.Fatalf("SaveConfig() error: %v", err)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}
	if loaded.ActiveWorkspace != "test-ws" {
		t.Errorf("loaded.ActiveWorkspace = %q, want 'test-ws'", loaded.ActiveWorkspace)
	}
	if !loaded.OnboardingCompleted {
		t.Error("loaded.OnboardingCompleted = false, want true")
	}
}

func TestBoost_Config_Roundtrip(t *testing.T) {
	_ = setupTempHome(t)
	_ = EnsureScribeDirs()

	cfg := &Config{
		ActiveWorkspace:     "my-ws",
		OnboardingCompleted: true,
	}
	err := SaveConfig(cfg)
	if err != nil {
		t.Fatalf("SaveConfig error: %v", err)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}
	if loaded.ActiveWorkspace != "my-ws" {
		t.Errorf("ActiveWorkspace = %q, want 'my-ws'", loaded.ActiveWorkspace)
	}
	if !loaded.OnboardingCompleted {
		t.Error("OnboardingCompleted = false, want true")
	}
}

func TestBoost_LoadConfig_InvalidJSON(t *testing.T) {
	tmpDir := setupTempHome(t)
	_ = EnsureScribeDirs()

	configPath := filepath.Join(tmpDir, ".scribe", "config.json")
	_ = os.WriteFile(configPath, []byte("invalid json"), 0o644)

	_, err := LoadConfig()
	if err == nil {
		t.Error("expected error for invalid JSON config")
	}
}
