package scribe

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestInstallManifest_ReadWriteRoundTrip(t *testing.T) {
	home := setInstallManifestTestHome(t)

	manifest := &InstallManifest{
		Version:             "1.2.3",
		Channel:             "stable",
		CLIPath:             filepath.Join(home, ".local", "bin", "scribe"),
		DesktopPath:         filepath.Join(home, ".local", "lib", "scribe", "scribe-desktop"),
		AppBundlePath:       "",
		InstalledComponents: []string{InstallComponentCLI, InstallComponentDesktop},
		PublicDownloadBase:  "https://example.com/scribe",
	}
	if err := WriteInstallManifest(manifest); err != nil {
		t.Fatalf("WriteInstallManifest() error = %v", err)
	}

	got, err := ReadInstallManifest()
	if err != nil {
		t.Fatalf("ReadInstallManifest() error = %v", err)
	}
	if !reflect.DeepEqual(got, manifest) {
		t.Fatalf("ReadInstallManifest() = %#v, want %#v", got, manifest)
	}

	manifestPath, err := GetInstallManifestPath()
	if err != nil {
		t.Fatalf("GetInstallManifestPath() error = %v", err)
	}
	if manifestPath != filepath.Join(home, ".scribe", InstallManifestFileName) {
		t.Fatalf("manifest path = %q, want path under test HOME", manifestPath)
	}
}

func TestDefaultInstallPathsByPlatform(t *testing.T) {
	home := filepath.Join(string(filepath.Separator), "home", "scribe")

	tests := []struct {
		name              string
		goos              string
		wantCLIPath       string
		wantDesktopPath   string
		wantAppBundlePath string
		wantComponents    []string
	}{
		{
			name:              "linux",
			goos:              "linux",
			wantCLIPath:       filepath.Join(home, ".local", "bin", "scribe"),
			wantDesktopPath:   filepath.Join(home, ".local", "lib", "scribe", "scribe-desktop"),
			wantAppBundlePath: "",
			wantComponents:    []string{InstallComponentCLI, InstallComponentDesktop},
		},
		{
			name:              "darwin",
			goos:              "darwin",
			wantCLIPath:       filepath.Join(home, ".local", "bin", "scribe"),
			wantDesktopPath:   filepath.Join(home, "Applications", "Scribe.app", "Contents", "MacOS", "scribe-desktop"),
			wantAppBundlePath: filepath.Join(home, "Applications", "Scribe.app"),
			wantComponents:    []string{InstallComponentCLI, InstallComponentDesktop},
		},
		{
			name:              "windows",
			goos:              "windows",
			wantCLIPath:       filepath.Join(home, "AppData", "Local", "Scribe", "scribe.exe"),
			wantDesktopPath:   "",
			wantAppBundlePath: "",
			wantComponents:    []string{InstallComponentBinary},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := defaultCLIPathFor(home, tt.goos); got != tt.wantCLIPath {
				t.Fatalf("defaultCLIPathFor() = %q, want %q", got, tt.wantCLIPath)
			}
			if got := defaultDesktopPathFor(home, tt.goos); got != tt.wantDesktopPath {
				t.Fatalf("defaultDesktopPathFor() = %q, want %q", got, tt.wantDesktopPath)
			}
			if got := defaultAppBundlePathFor(home, tt.goos); got != tt.wantAppBundlePath {
				t.Fatalf("defaultAppBundlePathFor() = %q, want %q", got, tt.wantAppBundlePath)
			}
			if got := defaultInstalledComponentsFor(tt.goos, tt.wantDesktopPath); !reflect.DeepEqual(got, tt.wantComponents) {
				t.Fatalf("defaultInstalledComponentsFor() = %#v, want %#v", got, tt.wantComponents)
			}
		})
	}
}

func TestInstallManifestForSelfUpdate_MissingManifestUsesExecutableAndDefaults(t *testing.T) {
	home := setInstallManifestTestHome(t)
	oldVersion := Version
	Version = "1.2.3"
	t.Cleanup(func() { Version = oldVersion })

	manifest, err := installManifestForSelfUpdate()
	if err != nil {
		t.Fatalf("installManifestForSelfUpdate() error = %v", err)
	}
	if manifest.Version != "1.2.3" {
		t.Fatalf("Version = %q, want 1.2.3", manifest.Version)
	}

	execPath, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable() error = %v", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		t.Fatalf("EvalSymlinks() error = %v", err)
	}

	switch runtime.GOOS {
	case "darwin":
		if filepath.Base(execPath) == "scribe-desktop" {
			if manifest.DesktopPath != execPath {
				t.Fatalf("DesktopPath = %q, want executable path %q", manifest.DesktopPath, execPath)
			}
		} else if manifest.CLIPath != execPath {
			t.Fatalf("CLIPath = %q, want executable path %q", manifest.CLIPath, execPath)
		}
		if manifest.AppBundlePath != filepath.Join(home, "Applications", "Scribe.app") {
			t.Fatalf("AppBundlePath = %q, want default app bundle path", manifest.AppBundlePath)
		}
	case "linux":
		if filepath.Base(execPath) == "scribe-desktop" {
			if manifest.DesktopPath != execPath {
				t.Fatalf("DesktopPath = %q, want executable path %q", manifest.DesktopPath, execPath)
			}
		} else if manifest.CLIPath != execPath {
			t.Fatalf("CLIPath = %q, want executable path %q", manifest.CLIPath, execPath)
		}
		if !componentInstalled(manifest, InstallComponentCLI) || !componentInstalled(manifest, InstallComponentDesktop) {
			t.Fatalf("InstalledComponents = %#v, want cli and desktop", manifest.InstalledComponents)
		}
	case "windows":
		if manifest.CLIPath != execPath {
			t.Fatalf("CLIPath = %q, want executable path %q", manifest.CLIPath, execPath)
		}
		if !reflect.DeepEqual(manifest.InstalledComponents, []string{InstallComponentBinary}) {
			t.Fatalf("InstalledComponents = %#v, want binary", manifest.InstalledComponents)
		}
	}
}

func setInstallManifestTestHome(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)
	return tmpDir
}
