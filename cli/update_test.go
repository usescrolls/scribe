package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	scribe "gitlab.com/usescrolls/scribe/internal"
)

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

func TestRunUpdateShowsNewSkillsNotice(t *testing.T) {
	tmpDir, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	updateForce = true

	// Create a local git repo with 2 skills (skill-a and skill-b)
	repoDir := filepath.Join(tmpDir, "remote-repo.git")
	skillADir := filepath.Join(repoDir, "skill-a")
	skillBDir := filepath.Join(repoDir, "skill-b")
	_ = os.MkdirAll(skillADir, 0o755)
	_ = os.MkdirAll(skillBDir, 0o755)

	contentA := "---\nname: skill-a\ndescription: Skill A\n---\n\n# Skill A\n\nTest skill content.\n"
	contentB := "---\nname: skill-b\ndescription: Skill B\n---\n\n# Skill B\n\nTest skill content.\n"
	_ = os.WriteFile(filepath.Join(skillADir, "SKILL.md"), []byte(contentA), 0o644)
	_ = os.WriteFile(filepath.Join(skillBDir, "SKILL.md"), []byte(contentB), 0o644)

	// Initialize git repo and commit
	cmds := [][]string{
		{"git", "init", repoDir},
		{"git", "-C", repoDir, "add", "."},
		{"git", "-C", repoDir, "-c", "user.email=test@test.com", "-c", "user.name=Test", "commit", "-m", "init"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git command %v failed: %v\n%s", args, err, out)
		}
	}

	// Install only skill-a, pointing its source to the local git repo
	scrollsDir, _ := scribe.GetScrollsDir()
	skillDir := filepath.Join(scrollsDir, "skill-a")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(contentA), 0o644)

	meta := scribe.NewSkillMeta(&scribe.SourceInfo{
		Type:  "github",
		Owner: "testowner",
		Repo:  "remote-repo",
		URL:   repoDir,
	}, "", contentA, nil)
	meta.Source = "testowner/remote-repo"
	meta.SourceType = "github"
	meta.SourceURL = repoDir
	metaPath := filepath.Join(skillDir, ".scribe-meta.json")
	_ = scribe.WriteSkillMeta(metaPath, meta)

	output := captureStdout(t, func() {
		err := runUpdate(updateCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "other skill(s) available") {
		t.Errorf("expected 'other skill(s) available' notice, got: %s", output)
	}
	if !strings.Contains(output, "testowner/remote-repo") {
		t.Errorf("expected source name in notice, got: %s", output)
	}
}

func TestRunUpdateNewSkillsNoticeQuiet(t *testing.T) {
	tmpDir, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	updateForce = true

	// Create a local git repo with 2 skills (skill-a and skill-b)
	repoDir := filepath.Join(tmpDir, "remote-repo.git")
	skillADir := filepath.Join(repoDir, "skill-a")
	skillBDir := filepath.Join(repoDir, "skill-b")
	_ = os.MkdirAll(skillADir, 0o755)
	_ = os.MkdirAll(skillBDir, 0o755)

	contentA := "---\nname: skill-a\ndescription: Skill A\n---\n\n# Skill A\n\nTest skill content.\n"
	contentB := "---\nname: skill-b\ndescription: Skill B\n---\n\n# Skill B\n\nTest skill content.\n"
	_ = os.WriteFile(filepath.Join(skillADir, "SKILL.md"), []byte(contentA), 0o644)
	_ = os.WriteFile(filepath.Join(skillBDir, "SKILL.md"), []byte(contentB), 0o644)

	// Initialize git repo and commit
	cmds := [][]string{
		{"git", "init", repoDir},
		{"git", "-C", repoDir, "add", "."},
		{"git", "-C", repoDir, "-c", "user.email=test@test.com", "-c", "user.name=Test", "commit", "-m", "init"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git command %v failed: %v\n%s", args, err, out)
		}
	}

	// Install only skill-a, pointing its source to the local git repo
	scrollsDir, _ := scribe.GetScrollsDir()
	skillDir := filepath.Join(scrollsDir, "skill-a")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(contentA), 0o644)

	meta := scribe.NewSkillMeta(&scribe.SourceInfo{
		Type:  "github",
		Owner: "testowner",
		Repo:  "remote-repo",
		URL:   repoDir,
	}, "", contentA, nil)
	meta.Source = "testowner/remote-repo"
	meta.SourceType = "github"
	meta.SourceURL = repoDir
	metaPath := filepath.Join(skillDir, ".scribe-meta.json")
	_ = scribe.WriteSkillMeta(metaPath, meta)

	output := captureStdout(t, func() {
		err := runUpdate(updateCmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if strings.Contains(output, "other skill(s) available") {
		t.Error("quiet mode should suppress new skills notice")
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
