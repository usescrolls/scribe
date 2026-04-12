package cli

import (
	"encoding/json"
	"strings"
	"testing"

	scribe "gitlab.com/usescrolls/scribe/internal"
)

// ---------------------------------------------------------------------------
// Tests for runWorkspaceList (workspace.go)
// ---------------------------------------------------------------------------

func TestRunWorkspaceList(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	origQuiet := quiet
	origJSON := jsonOutput
	defer func() {
		quiet = origQuiet
		jsonOutput = origJSON
	}()
	quiet = false
	jsonOutput = false

	output := captureStdout(t, func() {
		err := runWorkspaceList(workspaceListCmd, []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Workspaces") {
		t.Errorf("expected 'Workspaces' heading, got: %s", output)
	}
	if !strings.Contains(output, "default") {
		t.Errorf("expected 'default' workspace in listing, got: %s", output)
	}
}

func TestRunWorkspaceListJSON(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	origJSON := jsonOutput
	defer func() { jsonOutput = origJSON }()
	jsonOutput = true

	output := captureStdout(t, func() {
		err := runWorkspaceList(workspaceListCmd, []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var parsed []scribe.WorkspaceInfo
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v\nOutput: %s", err, output)
	}
	if len(parsed) == 0 {
		t.Error("expected at least one workspace (default)")
	}

	foundDefault := false
	for _, ws := range parsed {
		if ws.Name == "default" {
			foundDefault = true
			break
		}
	}
	if !foundDefault {
		t.Error("expected default workspace in JSON output")
	}
}

func TestRunWorkspaceListWithMultiple(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	jsonOutput = false
	workspaceDescription = "Dev workspace"

	_ = scribe.EnsureDefaultWorkspace()
	_ = runWorkspaceCreate(workspaceCreateCmd, []string{"dev"})

	output := captureStdout(t, func() {
		err := runWorkspaceList(workspaceListCmd, []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	// default should have the active marker (*)
	if !strings.Contains(output, "* default") {
		t.Errorf("expected '* default' (active marker) in output, got: %s", output)
	}
	if !strings.Contains(output, "dev") {
		t.Errorf("expected 'dev' workspace in output, got: %s", output)
	}
	if !strings.Contains(output, "Dev workspace") {
		t.Errorf("expected description in output, got: %s", output)
	}
}

func TestRunWorkspaceListDescriptionFormat(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	jsonOutput = false

	_ = scribe.EnsureDefaultWorkspace()

	// Create workspace with no description
	workspaceDescription = ""
	_ = runWorkspaceCreate(workspaceCreateCmd, []string{"no-desc-ws"})

	output := captureStdout(t, func() {
		err := runWorkspaceList(workspaceListCmd, []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	// "no-desc-ws" line should not have " - " since no description
	lines := strings.SplitSeq(output, "\n")
	for line := range lines {
		if strings.Contains(line, "no-desc-ws") {
			if strings.Contains(line, " - ") {
				t.Errorf("workspace without description should not have ' - ' separator: %s", line)
			}
			break
		}
	}
}

// ---------------------------------------------------------------------------
// Tests for runWorkspaceCreate (workspace.go)
// ---------------------------------------------------------------------------

func TestRunWorkspaceCreate(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	origQuiet := quiet
	defer func() { quiet = origQuiet }()
	quiet = false

	// Reset the flag variable
	workspaceDescription = ""

	output := captureStdout(t, func() {
		err := runWorkspaceCreate(workspaceCreateCmd, []string{"test-workspace"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "test-workspace") {
		t.Errorf("expected 'test-workspace' in output, got: %s", output)
	}
	if !strings.Contains(output, "created") {
		t.Errorf("expected 'created' in output, got: %s", output)
	}

	// Verify it actually exists
	ws, err := scribe.GetWorkspace("test-workspace")
	if err != nil {
		t.Fatalf("failed to get workspace: %v", err)
	}
	if ws.Name != "test-workspace" {
		t.Errorf("workspace name = %q, expected %q", ws.Name, "test-workspace")
	}
}

func TestRunWorkspaceCreateDuplicate(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	origQuiet := quiet
	defer func() { quiet = origQuiet }()
	quiet = true

	workspaceDescription = ""

	// Create first time
	err := runWorkspaceCreate(workspaceCreateCmd, []string{"dup-workspace"})
	if err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	// Create again should fail
	err = runWorkspaceCreate(workspaceCreateCmd, []string{"dup-workspace"})
	if err == nil {
		t.Error("expected error creating duplicate workspace")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}
}

func TestRunWorkspaceCreateWithDescription(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true
	workspaceDescription = "My test workspace"

	err := runWorkspaceCreate(workspaceCreateCmd, []string{"described-ws"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ws, err := scribe.GetWorkspace("described-ws")
	if err != nil {
		t.Fatalf("failed to get workspace: %v", err)
	}
	if ws.Description != "My test workspace" {
		t.Errorf("expected description 'My test workspace', got: %s", ws.Description)
	}
}

// ---------------------------------------------------------------------------
// Tests for runWorkspaceCurrent (workspace.go)
// ---------------------------------------------------------------------------

func TestRunWorkspaceCurrent(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	origQuiet := quiet
	origJSON := jsonOutput
	defer func() {
		quiet = origQuiet
		jsonOutput = origJSON
	}()
	quiet = false
	jsonOutput = false

	// Ensure default workspace is saved
	_ = scribe.EnsureDefaultWorkspace()

	output := captureStdout(t, func() {
		err := runWorkspaceCurrent(workspaceCurrentCmd, []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Active workspace:") {
		t.Errorf("expected 'Active workspace:' in output, got: %s", output)
	}
	if !strings.Contains(output, "default") {
		t.Errorf("expected 'default' workspace, got: %s", output)
	}
}

func TestRunWorkspaceCurrentJSON(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	origJSON := jsonOutput
	defer func() { jsonOutput = origJSON }()
	jsonOutput = true

	_ = scribe.EnsureDefaultWorkspace()

	output := captureStdout(t, func() {
		err := runWorkspaceCurrent(workspaceCurrentCmd, []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var parsed scribe.WorkspaceInfo
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v\nOutput: %s", err, output)
	}
	if parsed.Name != "default" {
		t.Errorf("workspace name = %q, expected %q", parsed.Name, "default")
	}
	if !parsed.IsActive {
		t.Error("expected isActive to be true")
	}
}

func TestRunWorkspaceCurrentWithDescription(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	jsonOutput = false
	workspaceDescription = "My custom workspace"

	_ = scribe.EnsureDefaultWorkspace()
	_ = runWorkspaceCreate(workspaceCreateCmd, []string{"custom-ws"})
	_ = scribe.SetActiveWorkspace("custom-ws")

	output := captureStdout(t, func() {
		err := runWorkspaceCurrent(workspaceCurrentCmd, []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "custom-ws") {
		t.Errorf("expected 'custom-ws' in output, got: %s", output)
	}
	if !strings.Contains(output, "Description:") {
		t.Errorf("expected 'Description:' in output, got: %s", output)
	}
}

func TestRunWorkspaceCurrentWithSkills(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	jsonOutput = false

	_ = scribe.EnsureDefaultWorkspace()
	installFakeSkill(t, "ws-skill-1", "WS Skill", "github", "owner/repo")
	_ = scribe.AddSkillToWorkspace("ws-skill-1", "default")

	output := captureStdout(t, func() {
		err := runWorkspaceCurrent(workspaceCurrentCmd, []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "ws-skill-1") {
		t.Errorf("expected 'ws-skill-1' in skill listing, got: %s", output)
	}
	if !strings.Contains(output, "Skills (") {
		t.Errorf("expected 'Skills (' count in output, got: %s", output)
	}
}

// ---------------------------------------------------------------------------
// Tests for runWorkspaceDelete (workspace.go)
// ---------------------------------------------------------------------------

func TestRunWorkspaceDelete(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	origQuiet := quiet
	defer func() { quiet = origQuiet }()
	quiet = false

	workspaceDescription = ""

	// Create a workspace to delete
	err := runWorkspaceCreate(workspaceCreateCmd, []string{"to-delete"})
	if err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	output := captureStdout(t, func() {
		err = runWorkspaceDelete(workspaceDeleteCmd, []string{"to-delete"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", output)
	}
}

func TestRunWorkspaceDeleteDefault(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	origQuiet := quiet
	defer func() { quiet = origQuiet }()
	quiet = true

	_ = scribe.EnsureDefaultWorkspace()

	err := runWorkspaceDelete(workspaceDeleteCmd, []string{"default"})
	if err == nil {
		t.Error("expected error deleting default workspace")
	}
	if !strings.Contains(err.Error(), "cannot delete") {
		t.Errorf("expected 'cannot delete' error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Tests for runWorkspaceUse (workspace.go)
// ---------------------------------------------------------------------------

func TestRunWorkspaceUseAlreadyActive(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false

	_ = scribe.EnsureDefaultWorkspace()

	output := captureStdout(t, func() {
		err := runWorkspaceUse(workspaceUseCmd, []string{"default"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Already using workspace") {
		t.Errorf("expected 'Already using workspace' message, got: %s", output)
	}
}

func TestRunWorkspaceUseAlreadyActiveQuiet(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true

	_ = scribe.EnsureDefaultWorkspace()

	output := captureStdout(t, func() {
		err := runWorkspaceUse(workspaceUseCmd, []string{"default"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if strings.Contains(output, "Already using") {
		t.Error("quiet mode should suppress message")
	}
}

func TestRunWorkspaceUseSwitchWorkspace(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	workspaceDescription = ""

	_ = scribe.EnsureDefaultWorkspace()

	// Create a new workspace
	err := runWorkspaceCreate(workspaceCreateCmd, []string{"my-workspace"})
	if err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	output := captureStdout(t, func() {
		err = runWorkspaceUse(workspaceUseCmd, []string{"my-workspace"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Switching from") {
		t.Errorf("expected 'Switching from' message, got: %s", output)
	}

	// Verify active workspace changed
	config, _ := scribe.LoadConfig()
	if config.ActiveWorkspace != "my-workspace" {
		t.Errorf("expected active workspace to be 'my-workspace', got: %s", config.ActiveWorkspace)
	}
}

func TestRunWorkspaceUseNonexistent(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true

	_ = scribe.EnsureDefaultWorkspace()

	err := runWorkspaceUse(workspaceUseCmd, []string{"nonexistent-ws"})
	if err == nil {
		t.Error("expected error for nonexistent workspace")
	}
}

func TestRunWorkspaceUseVerboseOutput(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	workspaceDescription = ""

	_ = scribe.EnsureDefaultWorkspace()
	_ = runWorkspaceCreate(workspaceCreateCmd, []string{"target-ws"})

	output := captureStdout(t, func() {
		err := runWorkspaceUse(workspaceUseCmd, []string{"target-ws"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Switching from") {
		t.Errorf("expected 'Switching from', got: %s", output)
	}
	if !strings.Contains(output, "Active workspace: target-ws") {
		t.Errorf("expected 'Active workspace: target-ws', got: %s", output)
	}
}

// ---------------------------------------------------------------------------
// Tests for runWorkspaceAdd (workspace.go)
// ---------------------------------------------------------------------------

func TestRunWorkspaceAddExistingSkill(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false

	_ = scribe.EnsureDefaultWorkspace()
	installFakeSkill(t, "addable-skill", "Addable skill", "github", "owner/repo")

	output := captureStdout(t, func() {
		err := runWorkspaceAdd(workspaceAddCmd, []string{"addable-skill"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Added") {
		t.Errorf("expected 'Added' in output, got: %s", output)
	}
}

func TestRunWorkspaceAddQuiet(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true

	_ = scribe.EnsureDefaultWorkspace()
	installFakeSkill(t, "quiet-add-skill", "Quiet add", "github", "owner/repo")

	output := captureStdout(t, func() {
		err := runWorkspaceAdd(workspaceAddCmd, []string{"quiet-add-skill"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if strings.Contains(output, "Added") {
		t.Error("quiet mode should suppress output")
	}
}

func TestRunWorkspaceAddNonexistentSkill(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true

	_ = scribe.EnsureDefaultWorkspace()

	err := runWorkspaceAdd(workspaceAddCmd, []string{"nonexistent-skill"})
	if err == nil {
		t.Error("expected error for nonexistent skill")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Tests for runWorkspaceRemove (workspace.go)
// ---------------------------------------------------------------------------

func TestRunWorkspaceRemoveSkill(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false

	_ = scribe.EnsureDefaultWorkspace()
	installFakeSkill(t, "removable-skill", "Removable skill", "github", "owner/repo")
	_ = scribe.AddSkillToWorkspace("removable-skill", "default")

	output := captureStdout(t, func() {
		err := runWorkspaceRemove(workspaceRemoveCmd, []string{"removable-skill"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Removed") {
		t.Errorf("expected 'Removed' in output, got: %s", output)
	}
}

func TestRunWorkspaceRemoveQuiet(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true

	_ = scribe.EnsureDefaultWorkspace()
	installFakeSkill(t, "quiet-rm-skill", "Quiet remove", "github", "owner/repo")
	_ = scribe.AddSkillToWorkspace("quiet-rm-skill", "default")

	output := captureStdout(t, func() {
		err := runWorkspaceRemove(workspaceRemoveCmd, []string{"quiet-rm-skill"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if strings.Contains(output, "Removed") {
		t.Error("quiet mode should suppress output")
	}
}

func TestRunWorkspaceRemoveNonexistentSkill(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = true

	_ = scribe.EnsureDefaultWorkspace()

	// Removing a skill that doesn't exist in the workspace should not error
	err := runWorkspaceRemove(workspaceRemoveCmd, []string{"not-in-workspace"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
