package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	scribe "gitlab.com/usescrolls/scribe/internal"
)

// setupTempHome creates a temp directory, sets HOME to it, initializes the
// logger and scribe directories, and returns a cleanup function.
func setupTempHome(t *testing.T) (dir string, cleanup func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "scribe-ws-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	scribe.InitLoggerCLI(false)
	_ = scribe.EnsureScribeDirs()

	cleanup = func() {
		_ = os.Setenv("HOME", origHome)
		_ = os.RemoveAll(tmpDir)
	}
	return tmpDir, cleanup
}

// installFakeSkill installs a fake skill into the temp HOME's scrolls directory.
// Returns the skill directory path.
func installFakeSkill(t *testing.T, name, description, sourceType, source string) string {
	return installFakeSkillWithBody(t, name, description, sourceType, source, "Test skill content.")
}

func installFakeSkillWithBody(t *testing.T, name, description, sourceType, source, body string) string {
	t.Helper()
	scrollsDir, err := scribe.GetScrollsDir()
	if err != nil {
		t.Fatalf("failed to get scrolls dir: %v", err)
	}
	skillDir := filepath.Join(scrollsDir, name)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("failed to create skill dir: %v", err)
	}

	content := fmt.Sprintf("---\nname: %s\ndescription: %s\n---\n\n# %s\n\n%s\n", name, description, name, body)
	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write SKILL.md: %v", err)
	}

	meta := scribe.NewSkillMeta(&scribe.SourceInfo{
		Type:  sourceType,
		Owner: "testowner",
		Repo:  "testrepo",
		URL:   "https://github.com/testowner/testrepo",
	}, "", content, nil)
	meta.Source = source
	meta.SourceType = sourceType

	metaPath := filepath.Join(skillDir, ".scribe-meta.json")
	if err := scribe.WriteSkillMeta(metaPath, meta); err != nil {
		t.Fatalf("failed to write meta: %v", err)
	}

	return skillDir
}

// captureStdout runs fn() and captures whatever it writes to os.Stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String()
}

// saveAndRestoreFlags saves CLI flag globals and restores them in the cleanup function.
func saveAndRestoreFlags(t *testing.T) {
	t.Helper()
	origQuiet := quiet
	origJSON := jsonOutput
	origNamesOnly := namesOnly
	origListSearch := listSearch
	origInstallSkills := installSkills
	origInstallListOnly := installListOnly
	origInstallYes := installYes
	origInstallAll := installAll
	origUninstallAll := uninstallAll
	origUpdateForce := updateForce
	origDebug := debug
	origWsDesc := workspaceDescription

	t.Cleanup(func() {
		quiet = origQuiet
		jsonOutput = origJSON
		namesOnly = origNamesOnly
		listSearch = origListSearch
		installSkills = origInstallSkills
		installListOnly = origInstallListOnly
		installYes = origInstallYes
		installAll = origInstallAll
		uninstallAll = origUninstallAll
		updateForce = origUpdateForce
		debug = origDebug
		workspaceDescription = origWsDesc
	})
}
