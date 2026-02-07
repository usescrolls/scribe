package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	scribe "github.com/usescrolls/scribe/internal"
)

// ---------------------------------------------------------------------------
// Tests for copyFile and copySkillDir (update.go)
// ---------------------------------------------------------------------------

func TestCopyFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-copyfile-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	srcPath := filepath.Join(tmpDir, "source.txt")
	dstPath := filepath.Join(tmpDir, "dest.txt")

	content := "Hello, this is test content for copyFile."
	if err := os.WriteFile(srcPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	if err := copyFile(srcPath, dstPath); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	// Verify content
	data, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}
	if string(data) != content {
		t.Errorf("copied content = %q, expected %q", string(data), content)
	}

	// Verify permissions match
	srcInfo, _ := os.Stat(srcPath)
	dstInfo, _ := os.Stat(dstPath)
	if srcInfo.Mode() != dstInfo.Mode() {
		t.Errorf("file mode mismatch: src=%v, dst=%v", srcInfo.Mode(), dstInfo.Mode())
	}
}

func TestCopyFileNonExistentSource(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-copyfile-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	err = copyFile(filepath.Join(tmpDir, "nonexistent"), filepath.Join(tmpDir, "dest"))
	if err == nil {
		t.Error("expected error copying nonexistent file, got nil")
	}
}

func TestCopyFileDifferentPermissions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-copyfile-perms-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	srcPath := filepath.Join(tmpDir, "executable.sh")
	dstPath := filepath.Join(tmpDir, "copy.sh")

	_ = os.WriteFile(srcPath, []byte("#!/bin/bash\necho hello"), 0o755)

	if err := copyFile(srcPath, dstPath); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	info, err := os.Stat(dstPath)
	if err != nil {
		t.Fatalf("failed to stat dest: %v", err)
	}
	if info.Mode() != 0o755 {
		t.Errorf("expected mode 0755, got %v", info.Mode())
	}
}

func TestCopyFileToNonexistentDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-copyfile-nodir-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	srcPath := filepath.Join(tmpDir, "source.txt")
	_ = os.WriteFile(srcPath, []byte("content"), 0o644)

	// copyFile in update.go doesn't create parent dirs, so writing to a deeply
	// nested path should fail
	dstPath := filepath.Join(tmpDir, "deep", "nested", "dest.txt")
	err = copyFile(srcPath, dstPath)
	if err == nil {
		t.Error("expected error writing to nonexistent parent directory")
	}
}

func TestCopySkillDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-copydir-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	srcDir := filepath.Join(tmpDir, "source-skill")
	dstDir := filepath.Join(tmpDir, "dest-skill")

	// Create source directory structure
	if err := os.MkdirAll(filepath.Join(srcDir, "subdir"), 0o755); err != nil {
		t.Fatalf("failed to create source dirs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte("# Test Skill"), 0o644); err != nil {
		t.Fatalf("failed to write SKILL.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "subdir", "extra.txt"), []byte("extra content"), 0o644); err != nil {
		t.Fatalf("failed to write extra.txt: %v", err)
	}

	if err := copySkillDir(srcDir, dstDir); err != nil {
		t.Fatalf("copySkillDir failed: %v", err)
	}

	// Verify SKILL.md was copied
	data, err := os.ReadFile(filepath.Join(dstDir, "SKILL.md"))
	if err != nil {
		t.Fatalf("failed to read copied SKILL.md: %v", err)
	}
	if string(data) != "# Test Skill" {
		t.Errorf("SKILL.md content = %q, expected %q", string(data), "# Test Skill")
	}

	// Verify subdirectory and file were copied
	data, err = os.ReadFile(filepath.Join(dstDir, "subdir", "extra.txt"))
	if err != nil {
		t.Fatalf("failed to read copied subdir/extra.txt: %v", err)
	}
	if string(data) != "extra content" {
		t.Errorf("extra.txt content = %q, expected %q", string(data), "extra content")
	}
}

func TestCopySkillDirOverwrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-copydir-overwrite-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	srcDir := filepath.Join(tmpDir, "source-skill")
	dstDir := filepath.Join(tmpDir, "dest-skill")

	// Create source
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte("new content"), 0o644); err != nil {
		t.Fatalf("failed to write source SKILL.md: %v", err)
	}

	// Create pre-existing destination with old content
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dstDir, "SKILL.md"), []byte("old content"), 0o644); err != nil {
		t.Fatalf("failed to write dest SKILL.md: %v", err)
	}

	// copySkillDir removes existing dest first
	if err := copySkillDir(srcDir, dstDir); err != nil {
		t.Fatalf("copySkillDir failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dstDir, "SKILL.md"))
	if err != nil {
		t.Fatalf("failed to read copied SKILL.md: %v", err)
	}
	if string(data) != "new content" {
		t.Errorf("content = %q, expected %q", string(data), "new content")
	}
}

func TestCopySkillDirNonexistentSource(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-copydir-noexist-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	err = copySkillDir(filepath.Join(tmpDir, "nonexistent"), filepath.Join(tmpDir, "dest"))
	if err == nil {
		t.Error("expected error copying nonexistent source directory")
	}
}

func TestCopySkillDirEmptySource(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-copydir-empty-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	srcDir := filepath.Join(tmpDir, "empty-src")
	dstDir := filepath.Join(tmpDir, "empty-dst")
	_ = os.MkdirAll(srcDir, 0o755)

	err = copySkillDir(srcDir, dstDir)
	if err != nil {
		t.Fatalf("unexpected error copying empty dir: %v", err)
	}

	// Verify destination exists and is empty
	entries, err := os.ReadDir(dstDir)
	if err != nil {
		t.Fatalf("failed to read dest dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty dest dir, got %d entries", len(entries))
	}
}

func TestCopySkillDirDeepNesting(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-copydir-deep-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	srcDir := filepath.Join(tmpDir, "deep-src")
	dstDir := filepath.Join(tmpDir, "deep-dst")

	// Create deeply nested structure
	deepPath := filepath.Join(srcDir, "a", "b", "c")
	_ = os.MkdirAll(deepPath, 0o755)
	_ = os.WriteFile(filepath.Join(srcDir, "top.txt"), []byte("top"), 0o644)
	_ = os.WriteFile(filepath.Join(srcDir, "a", "mid.txt"), []byte("mid"), 0o644)
	_ = os.WriteFile(filepath.Join(deepPath, "deep.txt"), []byte("deep"), 0o644)

	if err := copySkillDir(srcDir, dstDir); err != nil {
		t.Fatalf("copySkillDir failed: %v", err)
	}

	// Verify all files were copied
	for _, rel := range []string{"top.txt", "a/mid.txt", "a/b/c/deep.txt"} {
		data, err := os.ReadFile(filepath.Join(dstDir, rel))
		if err != nil {
			t.Errorf("failed to read %s: %v", rel, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("file %s is empty", rel)
		}
	}
}

// ---------------------------------------------------------------------------
// Tests for runUpdate (update.go)
// ---------------------------------------------------------------------------

func TestRunUpdateNoSkillsInstalled(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	updateForce = false

	output := captureStdout(t, func() {
		err := runUpdate(updateCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "No skills installed") {
		t.Errorf("expected 'No skills installed', got: %s", output)
	}
}

func TestRunUpdateNoSkillsInstalledQuiet(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	updateForce = false

	output := captureStdout(t, func() {
		err := runUpdate(updateCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if strings.Contains(output, "No skills installed") {
		t.Error("quiet mode should suppress message")
	}
}

func TestRunUpdateAllUpToDate(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	updateForce = false

	// Install a local skill (can't be checked for updates)
	installFakeSkill(t, "local-update", "Local update test", "local", "/some/path")

	output := captureStdout(t, func() {
		err := runUpdate(updateCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	// Local skills have errors in checkSkill, so nothing to update
	if !strings.Contains(output, "All skills are up-to-date") {
		t.Errorf("expected 'All skills are up-to-date', got: %s", output)
	}
}

func TestRunUpdateForceWithLocalSkills(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	updateForce = true

	// Local skill - updateSkill will fail with "local source, cannot update"
	installFakeSkill(t, "force-local", "Force local", "local", "/some/path")

	captureStdout(t, func() {
		err := runUpdate(updateCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	// Should not panic, just have failures in output
}

func TestRunUpdateSingleSkillNonexistent(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	updateForce = false

	captureStdout(t, func() {
		err := runUpdate(updateCmd, []string{"nonexistent-skill"})
		if err != nil {
			t.Errorf("unexpected error from runUpdate: %v", err)
		}
	})
}

func TestRunUpdateSingleLocalSkill(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	updateForce = true

	installFakeSkill(t, "single-update", "Single update", "local", "/local/path")

	captureStdout(t, func() {
		err := runUpdate(updateCmd, []string{"single-update"})
		// Should not return an error even though individual update fails
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestRunUpdateForceAllLocalSkills(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	updateForce = true

	installFakeSkill(t, "force-update-1", "Force 1", "local", "/path1")
	installFakeSkill(t, "force-update-2", "Force 2", "local", "/path2")

	output := captureStdout(t, func() {
		err := runUpdate(updateCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Updating 2 skill(s)") {
		t.Errorf("expected 'Updating 2 skill(s)', got: %s", output)
	}
	// Both should fail since they are local sources
	if !strings.Contains(output, "0/2 skill(s) updated") {
		t.Errorf("expected '0/2 skill(s) updated', got: %s", output)
	}
}

func TestRunUpdateVerboseChecking(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	updateForce = false

	// Install two local skills (local skills have errors in check, so 0 to update)
	installFakeSkill(t, "verbose-check-1", "Verbose check 1", "local", "/local/1")
	installFakeSkill(t, "verbose-check-2", "Verbose check 2", "local", "/local/2")

	output := captureStdout(t, func() {
		err := runUpdate(updateCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Checking 2 skill(s) for updates") {
		t.Errorf("expected checking message, got: %s", output)
	}
	if !strings.Contains(output, "All skills are up-to-date") {
		t.Errorf("expected all up-to-date message, got: %s", output)
	}
}

func TestRunUpdateVerboseAllUpToDateQuiet(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	updateForce = false

	installFakeSkill(t, "quiet-check", "Quiet check", "local", "/local")

	output := captureStdout(t, func() {
		err := runUpdate(updateCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if strings.Contains(output, "Checking") {
		t.Error("quiet mode should suppress checking message")
	}
	if strings.Contains(output, "All skills are up-to-date") {
		t.Error("quiet mode should suppress up-to-date message")
	}
}

func TestRunUpdateForceVerbose(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	updateForce = true

	installFakeSkill(t, "force-v-1", "Force verbose 1", "local", "/local")

	output := captureStdout(t, func() {
		err := runUpdate(updateCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	// Force mode should show "Updating N skill(s)"
	if !strings.Contains(output, "Updating 1 skill(s)") {
		t.Errorf("expected 'Updating 1 skill(s)', got: %s", output)
	}
}

// ---------------------------------------------------------------------------
// Tests for updateSkill (update.go)
// ---------------------------------------------------------------------------

func TestUpdateSkillNonexistent(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true

	err := updateSkill("nonexistent", false)
	if err == nil {
		t.Error("expected error for nonexistent skill")
	}
	if !strings.Contains(err.Error(), "skill not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateSkillNoMeta(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true

	// Install a skill without metadata
	scrollsDir, _ := scribe.GetScrollsDir()
	skillDir := filepath.Join(scrollsDir, "no-meta-update")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: no-meta-update\ndescription: No meta\n---\n\nContent\n"), 0o644)

	err := updateSkill("no-meta-update", false)
	if err == nil {
		t.Error("expected error for skill with no metadata")
	}
	if !strings.Contains(err.Error(), "no metadata") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateSkillLocalSource(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true

	installFakeSkill(t, "local-no-update", "Local no update", "local", "/some/path")

	err := updateSkill("local-no-update", false)
	if err == nil {
		t.Error("expected error for local source skill")
	}
	if !strings.Contains(err.Error(), "local source") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateSkillGitHubFetchFails(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true

	installFakeSkill(t, "github-update-skill", "GitHub update test", "github", "nonexistent-owner/nonexistent-repo")

	err := updateSkill("github-update-skill", false)
	if err == nil {
		t.Error("expected error for nonexistent github repo fetch")
	}
	if !strings.Contains(err.Error(), "failed to fetch") {
		t.Errorf("expected 'failed to fetch' error, got: %v", err)
	}
}

func TestUpdateSkillGitHubFetchFailsVerbose(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false

	installFakeSkill(t, "verbose-github-update", "Verbose github update", "github", "nonexistent-owner/nonexistent-repo")

	output := captureStdout(t, func() {
		err := updateSkill("verbose-github-update", false)
		if err == nil {
			t.Error("expected error for nonexistent github repo")
		}
	})

	if !strings.Contains(output, "Updating verbose-github-update") {
		t.Errorf("expected 'Updating' message, got: %s", output)
	}
}
