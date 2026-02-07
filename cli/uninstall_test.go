package cli

import (
	"strings"
	"testing"

	scribe "github.com/usescrolls/scribe/internal"
)

// ---------------------------------------------------------------------------
// Tests for runUninstall (uninstall.go)
// ---------------------------------------------------------------------------

func TestRunUninstallNoArgs(t *testing.T) {
	origQuiet := quiet
	origUninstallAll := uninstallAll
	defer func() {
		quiet = origQuiet
		uninstallAll = origUninstallAll
	}()
	quiet = true
	uninstallAll = false

	err := runUninstall(uninstallCmd, []string{})
	if err == nil {
		t.Fatal("expected error when no args and --all not set")
	}
	if !strings.Contains(err.Error(), "skill name is required") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRunUninstallNonexistentSkill(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	origQuiet := quiet
	origUninstallAll := uninstallAll
	defer func() {
		quiet = origQuiet
		uninstallAll = origUninstallAll
	}()
	quiet = true
	uninstallAll = false

	err := runUninstall(uninstallCmd, []string{"nonexistent-skill"})
	if err == nil {
		t.Fatal("expected error for nonexistent skill")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestRunUninstallExistingSkill(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	uninstallAll = false

	_ = scribe.EnsureDefaultWorkspace()
	installFakeSkill(t, "to-remove", "Skill to remove", "github", "owner/repo")

	// Add to default workspace
	_ = scribe.AddSkillToWorkspace("to-remove", "default")

	output := captureStdout(t, func() {
		err := runUninstall(uninstallCmd, []string{"to-remove"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "removed successfully") {
		t.Errorf("expected success message, got: %s", output)
	}

	exists, _ := scribe.SkillExists("to-remove")
	if exists {
		t.Error("skill should have been removed")
	}
}

func TestRunUninstallQuiet(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	uninstallAll = false

	installFakeSkill(t, "quiet-remove", "Quiet remove test", "github", "owner/repo")

	output := captureStdout(t, func() {
		err := runUninstall(uninstallCmd, []string{"quiet-remove"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if strings.Contains(output, "removed successfully") {
		t.Error("quiet mode should suppress success message")
	}
}

func TestRunUninstallVerbosePath(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	uninstallAll = false

	_ = scribe.EnsureDefaultWorkspace()
	installFakeSkill(t, "verbose-rm", "Verbose remove", "github", "owner/repo")

	output := captureStdout(t, func() {
		err := runUninstall(uninstallCmd, []string{"verbose-rm"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Removing skill 'verbose-rm'") {
		t.Errorf("expected 'Removing skill' message, got: %s", output)
	}
	if !strings.Contains(output, "removed successfully") {
		t.Errorf("expected 'removed successfully', got: %s", output)
	}
}

// ---------------------------------------------------------------------------
// Tests for runUninstallAll (uninstall.go)
// ---------------------------------------------------------------------------

func TestRunUninstallAllNoSkills(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false

	output := captureStdout(t, func() {
		err := runUninstallAll()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "No skills installed") {
		t.Errorf("expected 'No skills installed', got: %s", output)
	}
}

func TestRunUninstallAllNoSkillsQuiet(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true

	output := captureStdout(t, func() {
		err := runUninstallAll()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if strings.Contains(output, "No skills installed") {
		t.Error("quiet mode should suppress message")
	}
}

func TestRunUninstallAllWithSkills(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false

	_ = scribe.EnsureDefaultWorkspace()
	installFakeSkill(t, "skill-one", "First skill", "github", "owner/repo1")
	installFakeSkill(t, "skill-two", "Second skill", "github", "owner/repo2")
	_ = scribe.AddSkillToWorkspace("skill-one", "default")
	_ = scribe.AddSkillToWorkspace("skill-two", "default")

	output := captureStdout(t, func() {
		err := runUninstallAll()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Removing 2 skill(s)") {
		t.Errorf("expected 'Removing 2 skill(s)', got: %s", output)
	}
	if !strings.Contains(output, "All skills removed") {
		t.Errorf("expected 'All skills removed', got: %s", output)
	}

	// Verify all skills removed
	skills, _ := scribe.ListInstalledSkills()
	if len(skills) != 0 {
		t.Errorf("expected 0 skills, got %d", len(skills))
	}
}

func TestRunUninstallAllQuiet(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true

	installFakeSkill(t, "quiet-all-skill", "Quiet skill", "github", "owner/repo")

	output := captureStdout(t, func() {
		err := runUninstallAll()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if strings.Contains(output, "Removing") {
		t.Error("quiet mode should suppress messages")
	}
}

func TestRunUninstallViaAllFlag(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	uninstallAll = true

	installFakeSkill(t, "flag-all-skill", "Flag all skill", "github", "owner/repo")

	err := runUninstall(uninstallCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	exists, _ := scribe.SkillExists("flag-all-skill")
	if exists {
		t.Error("skill should have been removed via --all flag")
	}
}

func TestRunUninstallAllVerboseIndividual(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false

	_ = scribe.EnsureDefaultWorkspace()
	installFakeSkill(t, "ua-skill-1", "Uninstall all 1", "github", "o/r1")
	installFakeSkill(t, "ua-skill-2", "Uninstall all 2", "github", "o/r2")

	output := captureStdout(t, func() {
		err := runUninstallAll()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Removed ua-skill-1") {
		t.Errorf("expected individual removal message for skill-1, got: %s", output)
	}
	if !strings.Contains(output, "Removed ua-skill-2") {
		t.Errorf("expected individual removal message for skill-2, got: %s", output)
	}
}
