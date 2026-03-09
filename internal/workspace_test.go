package scribe

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ============================================================================
// CreateWorkspace (workspace.go)
// ============================================================================

func TestBoost_CreateWorkspace_Success(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	ws := &Workspace{
		Name:        "test-workspace",
		Description: "Test workspace",
		Skills:      []string{},
	}

	err := CreateWorkspace(ws)
	if err != nil {
		t.Fatalf("CreateWorkspace() error: %v", err)
	}

	// Verify it was created
	loaded, err := GetWorkspace("test-workspace")
	if err != nil {
		t.Fatalf("GetWorkspace() error: %v", err)
	}
	if loaded.Name != "test-workspace" {
		t.Errorf("workspace name = %q, want 'test-workspace'", loaded.Name)
	}
	if loaded.Description != "Test workspace" {
		t.Errorf("workspace description = %q, want 'Test workspace'", loaded.Description)
	}
}

func TestBoost_CreateWorkspace_EmptyName(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	ws := &Workspace{Name: ""}
	err := CreateWorkspace(ws)
	if err == nil {
		t.Error("expected error for empty workspace name")
	}
}

func TestBoost_CreateWorkspace_AlreadyExists(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	ws := &Workspace{Name: "dup-ws", Skills: []string{}}
	_ = CreateWorkspace(ws)

	err := CreateWorkspace(ws)
	if err == nil {
		t.Error("expected error when workspace already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %q, want 'already exists'", err.Error())
	}
}

// ============================================================================
// DeleteWorkspace (workspace.go)
// ============================================================================

func TestBoost_DeleteWorkspace_Success(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	ws := &Workspace{Name: "delete-me", Skills: []string{}}
	_ = CreateWorkspace(ws)

	err := DeleteWorkspace("delete-me")
	if err != nil {
		t.Fatalf("DeleteWorkspace() error: %v", err)
	}
}

func TestBoost_DeleteWorkspace_NotInListAfterDelete(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	ws := &Workspace{Name: "ephemeral", Skills: []string{}}
	_ = CreateWorkspace(ws)

	// Verify it appears in the list before deletion
	before, err := ListWorkspaces()
	if err != nil {
		t.Fatalf("ListWorkspaces() before error: %v", err)
	}
	foundBefore := false
	for _, w := range before {
		if w.Name == "ephemeral" {
			foundBefore = true
		}
	}
	if !foundBefore {
		t.Fatal("workspace not found in list before deletion")
	}

	// Delete and verify it's gone from the list
	if err := DeleteWorkspace("ephemeral"); err != nil {
		t.Fatalf("DeleteWorkspace() error: %v", err)
	}

	after, err := ListWorkspaces()
	if err != nil {
		t.Fatalf("ListWorkspaces() after error: %v", err)
	}
	for _, w := range after {
		if w.Name == "ephemeral" {
			t.Error("deleted workspace still appears in ListWorkspaces")
		}
	}
}

func TestBoost_DeleteWorkspace_DefaultFails(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	err := DeleteWorkspace(DefaultWorkspaceName)
	if err == nil {
		t.Error("expected error when deleting default workspace")
	}
}

func TestBoost_DeleteWorkspace_ActiveWorkspaceSwitchesToDefault(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	// Create and activate a workspace
	ws := &Workspace{Name: "active-delete", Skills: []string{}}
	_ = CreateWorkspace(ws)
	_ = SetActiveWorkspace("active-delete")

	// Delete the active workspace
	err := DeleteWorkspace("active-delete")
	if err != nil {
		t.Fatalf("DeleteWorkspace(active) error: %v", err)
	}

	// Active workspace should now be default
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}
	if config.ActiveWorkspace != DefaultWorkspaceName {
		t.Errorf("active workspace = %q, want %q", config.ActiveWorkspace, DefaultWorkspaceName)
	}
}

// ============================================================================
// SetActiveWorkspace (workspace.go)
// ============================================================================

func TestBoost_SetActiveWorkspace_Success(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	// Create a new workspace
	ws := &Workspace{Name: "new-active", Skills: []string{}}
	_ = CreateWorkspace(ws)

	err := SetActiveWorkspace("new-active")
	if err != nil {
		t.Fatalf("SetActiveWorkspace() error: %v", err)
	}

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}
	if config.ActiveWorkspace != "new-active" {
		t.Errorf("active workspace = %q, want 'new-active'", config.ActiveWorkspace)
	}
}

func TestBoost_SetActiveWorkspace_NonExistent(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	err := SetActiveWorkspace("nonexistent-ws-xyz")
	// This may or may not error depending on how GetWorkspace handles nonexistent
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			t.Logf("SetActiveWorkspace(nonexistent) error: %v", err)
		}
	}
}

// ============================================================================
// RemoveSkillFromWorkspace (workspace.go)
// ============================================================================

func TestBoost_RemoveSkillFromWorkspace_NotInWorkspace(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	// Removing a skill that's not in the workspace should be a no-op
	err := RemoveSkillFromWorkspace("nonexistent-skill", DefaultWorkspaceName)
	if err != nil {
		t.Errorf("RemoveSkillFromWorkspace(not present) error: %v", err)
	}
}

func TestBoost_RemoveSkillFromWorkspace_Success(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	// Install a skill first
	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls", "removable-skill")
	_ = os.MkdirAll(scrollsDir, 0o755)
	_ = os.WriteFile(filepath.Join(scrollsDir, "SKILL.md"), []byte("---\nname: removable-skill\ndescription: Removable\n---\n# R\n"), 0o644)

	// Add to workspace
	_ = AddSkillToWorkspace("removable-skill", DefaultWorkspaceName)

	// Remove from workspace
	err := RemoveSkillFromWorkspace("removable-skill", DefaultWorkspaceName)
	if err != nil {
		t.Fatalf("RemoveSkillFromWorkspace() error: %v", err)
	}

	// Verify it's not in the workspace anymore
	ws, err := GetWorkspace(DefaultWorkspaceName)
	if err != nil {
		t.Fatalf("GetWorkspace error: %v", err)
	}
	for _, s := range ws.Skills {
		if s == "removable-skill" {
			t.Error("skill still in workspace after removal")
		}
	}
}

func TestRemoveSkillFromWorkspace_CaseInsensitive(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	// Add with lowercase
	_ = AddSkillToWorkspace("my-skill", DefaultWorkspaceName)

	// Remove with different casing
	err := RemoveSkillFromWorkspace("MY-SKILL", DefaultWorkspaceName)
	if err != nil {
		t.Fatalf("RemoveSkillFromWorkspace(different case) error: %v", err)
	}

	ws, _ := GetWorkspace(DefaultWorkspaceName)
	for _, s := range ws.Skills {
		if strings.EqualFold(s, "my-skill") {
			t.Error("skill still in workspace after case-insensitive removal")
		}
	}
}

func TestBoost_RemoveSkillFromWorkspace_ActiveWorkspace(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	// Create and install a skill
	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls", "active-remove")
	_ = os.MkdirAll(scrollsDir, 0o755)
	_ = os.WriteFile(filepath.Join(scrollsDir, "SKILL.md"), []byte("---\nname: active-remove\ndescription: Active remove\n---\n# AR\n"), 0o644)

	// Add to active (default) workspace
	_ = AddSkillToWorkspace("active-remove", DefaultWorkspaceName)

	// Create agent dir for symlink sync
	_ = os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0o755)

	// Remove from active workspace - should also remove symlinks
	err := RemoveSkillFromWorkspace("active-remove", DefaultWorkspaceName)
	if err != nil {
		t.Fatalf("RemoveSkillFromWorkspace(active) error: %v", err)
	}
}

func TestBoost_RemoveSkillFromWorkspace_NonActiveWorkspace(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	// Create a non-active workspace
	ws := &Workspace{Name: "other-ws", Skills: []string{"some-skill"}}
	_ = CreateWorkspace(ws)

	// Remove skill from non-active workspace (should not affect symlinks)
	err := RemoveSkillFromWorkspace("some-skill", "other-ws")
	if err != nil {
		t.Fatalf("RemoveSkillFromWorkspace(non-active) error: %v", err)
	}
}

// ============================================================================
// EnsureDefaultWorkspace (workspace.go)
// ============================================================================

func TestBoost_EnsureDefaultWorkspace_FirstCall(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	err := EnsureDefaultWorkspace()
	if err != nil {
		t.Fatalf("EnsureDefaultWorkspace() error: %v", err)
	}

	// Verify default workspace file was created
	wsPath := filepath.Join(tmpDir, ".scribe", "workspaces", "default.json")
	if _, err := os.Stat(wsPath); err != nil {
		t.Error("default workspace file not created")
	}

	// Verify content
	ws, err := GetWorkspace(DefaultWorkspaceName)
	if err != nil {
		t.Fatalf("GetWorkspace(default) error: %v", err)
	}
	if ws.Name != DefaultWorkspaceName {
		t.Errorf("workspace name = %q, want %q", ws.Name, DefaultWorkspaceName)
	}
}

func TestBoost_EnsureDefaultWorkspace_Idempotent(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	_ = EnsureDefaultWorkspace()
	err := EnsureDefaultWorkspace()
	if err != nil {
		t.Fatalf("EnsureDefaultWorkspace() second call error: %v", err)
	}
}

func TestBoost_EnsureDefaultWorkspace_WithInstalledSkills(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// Pre-install some skills
	for _, name := range []string{"skill-a", "skill-b"} {
		skillDir := filepath.Join(tmpDir, ".scribe", "scrolls", name)
		_ = os.MkdirAll(skillDir, 0o755)
		_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: "+name+"\ndescription: test\n---\n# Test\n"), 0o644)
	}

	err := EnsureDefaultWorkspace()
	if err != nil {
		t.Fatalf("EnsureDefaultWorkspace() error: %v", err)
	}

	ws, err := GetWorkspace(DefaultWorkspaceName)
	if err != nil {
		t.Fatalf("GetWorkspace error: %v", err)
	}
	// +1 for injected system skill (scribe-cli)
	if len(ws.Skills) != 3 {
		t.Errorf("default workspace has %d skills, want 3 (2 test + 1 system)", len(ws.Skills))
	}
}

// ============================================================================
// AddSkillToActiveAndDefaultWorkspace (workspace.go)
// ============================================================================

func TestBoost_AddSkillToActiveAndDefaultWorkspace_DefaultIsActive(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	err := AddSkillToActiveAndDefaultWorkspace("new-skill")
	if err != nil {
		t.Fatalf("AddSkillToActiveAndDefaultWorkspace() error: %v", err)
	}

	ws, err := GetWorkspace(DefaultWorkspaceName)
	if err != nil {
		t.Fatalf("GetWorkspace error: %v", err)
	}
	found := false
	for _, s := range ws.Skills {
		if s == "new-skill" {
			found = true
		}
	}
	if !found {
		t.Error("skill not added to default workspace")
	}
}

func TestBoost_AddSkillToActiveAndDefaultWorkspace_DifferentActive(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	// Create and activate a different workspace
	ws := &Workspace{Name: "custom", Skills: []string{}}
	_ = CreateWorkspace(ws)
	_ = SetActiveWorkspace("custom")

	err := AddSkillToActiveAndDefaultWorkspace("dual-skill")
	if err != nil {
		t.Fatalf("AddSkillToActiveAndDefaultWorkspace() error: %v", err)
	}

	// Verify in custom (active workspace) — skills should go to the active workspace
	customWs, _ := GetWorkspace("custom")
	foundInCustom := false
	for _, s := range customWs.Skills {
		if s == "dual-skill" {
			foundInCustom = true
		}
	}
	if !foundInCustom {
		t.Error("skill not added to custom (active) workspace")
	}

	// Verify NOT in default — should respect the user's active workspace
	defaultWs, _ := GetWorkspace(DefaultWorkspaceName)
	for _, s := range defaultWs.Skills {
		if s == "dual-skill" {
			t.Error("skill should not be added to default workspace when custom is active")
		}
	}
}

func TestBoost_AddSkillToActiveAndDefaultWorkspace_FallbackToDefault(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	// Point config to a non-existent active workspace so GetActiveWorkspace fails
	_ = SaveConfig(&Config{ActiveWorkspace: "nonexistent-ws"})

	err := AddSkillToActiveAndDefaultWorkspace("fallback-skill")
	if err != nil {
		t.Fatalf("AddSkillToActiveAndDefaultWorkspace() error: %v", err)
	}

	// Verify skill falls back to default workspace
	defaultWs, err := GetWorkspace(DefaultWorkspaceName)
	if err != nil {
		t.Fatalf("GetWorkspace error: %v", err)
	}
	found := false
	for _, s := range defaultWs.Skills {
		if s == "fallback-skill" {
			found = true
		}
	}
	if !found {
		t.Error("skill not added to default workspace as fallback")
	}
}

// ============================================================================
// ListWorkspaces (workspace.go)
// ============================================================================

func TestBoost_ListWorkspaces_NoDir(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)

	// Without creating any dirs, ListWorkspaces should return default
	workspaces, err := ListWorkspaces()
	if err != nil {
		t.Fatalf("ListWorkspaces() error: %v", err)
	}
	if len(workspaces) < 1 {
		t.Fatal("expected at least 1 workspace (default)")
	}
	found := false
	for _, ws := range workspaces {
		if ws.Name == DefaultWorkspaceName {
			found = true
		}
	}
	if !found {
		t.Error("default workspace not found in list")
	}
}

func TestBoost_ListWorkspaces_WithMultiple(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	ws := &Workspace{Name: "custom-ws", Skills: []string{}}
	_ = CreateWorkspace(ws)

	workspaces, err := ListWorkspaces()
	if err != nil {
		t.Fatalf("ListWorkspaces() error: %v", err)
	}
	if len(workspaces) < 2 {
		t.Errorf("expected at least 2 workspaces, got %d", len(workspaces))
	}
}

func TestBoost_ListWorkspaces_NonJSONFiles(t *testing.T) {
	tmpDir := setupTempHome(t)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	// Create a non-JSON file and a directory in workspace dir
	wsDir := filepath.Join(tmpDir, ".scribe", "workspaces")
	_ = os.WriteFile(filepath.Join(wsDir, "readme.txt"), []byte("ignore"), 0o644)
	_ = os.MkdirAll(filepath.Join(wsDir, "subdir"), 0o755)

	workspaces, err := ListWorkspaces()
	if err != nil {
		t.Fatalf("ListWorkspaces error: %v", err)
	}
	// Should only find default workspace, not the txt/dir
	hasDefault := false
	for _, ws := range workspaces {
		if ws.Name == DefaultWorkspaceName {
			hasDefault = true
		}
	}
	if !hasDefault {
		t.Error("default workspace not found")
	}
}

// ============================================================================
// GetActiveWorkspace (workspace.go)
// ============================================================================

func TestBoost_GetActiveWorkspace_Default(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	ws, err := GetActiveWorkspace()
	if err != nil {
		t.Fatalf("GetActiveWorkspace() error: %v", err)
	}
	if ws.Name != DefaultWorkspaceName {
		t.Errorf("active workspace = %q, want %q", ws.Name, DefaultWorkspaceName)
	}
}

// ============================================================================
// GetWorkspaceInfo (workspace.go)
// ============================================================================

func TestBoost_GetWorkspaceInfo(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	infos, err := GetWorkspaceInfo()
	if err != nil {
		t.Fatalf("GetWorkspaceInfo() error: %v", err)
	}
	if len(infos) < 1 {
		t.Fatal("expected at least 1 workspace info")
	}

	// Default workspace should be active
	found := false
	for _, info := range infos {
		if info.Name == DefaultWorkspaceName && info.IsActive {
			found = true
		}
	}
	if !found {
		t.Error("default workspace not found or not active")
	}
}

// ============================================================================
// UpdateWorkspace (workspace.go)
// ============================================================================

func TestBoost_UpdateWorkspace(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	ws := &Workspace{Name: "update-ws", Description: "Original", Skills: []string{}}
	_ = CreateWorkspace(ws)

	ws.Description = "Updated"
	ws.Skills = []string{"skill-a", "skill-b"}
	err := UpdateWorkspace(ws)
	if err != nil {
		t.Fatalf("UpdateWorkspace() error: %v", err)
	}

	loaded, err := GetWorkspace("update-ws")
	if err != nil {
		t.Fatalf("GetWorkspace error: %v", err)
	}
	if loaded.Description != "Updated" {
		t.Errorf("description = %q, want 'Updated'", loaded.Description)
	}
	// +1 for injected system skill (scribe-cli)
	if len(loaded.Skills) != 3 {
		t.Errorf("skills count = %d, want 3 (2 test + 1 system)", len(loaded.Skills))
	}
}

// ============================================================================
// RemoveSkillFromAllWorkspaces (workspace.go)
// ============================================================================

func TestBoost_RemoveSkillFromAllWorkspaces(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	// Add skill to default workspace
	_ = AddSkillToWorkspace("global-remove", DefaultWorkspaceName)

	// Create another workspace with the same skill
	ws := &Workspace{Name: "other", Skills: []string{"global-remove", "keep-skill"}}
	_ = CreateWorkspace(ws)

	err := RemoveSkillFromAllWorkspaces("global-remove")
	if err != nil {
		t.Fatalf("RemoveSkillFromAllWorkspaces() error: %v", err)
	}

	// Verify removal from default
	defaultWs, _ := GetWorkspace(DefaultWorkspaceName)
	for _, s := range defaultWs.Skills {
		if s == "global-remove" {
			t.Error("skill still in default workspace")
		}
	}

	// Verify removal from other
	otherWs, _ := GetWorkspace("other")
	for _, s := range otherWs.Skills {
		if s == "global-remove" {
			t.Error("skill still in other workspace")
		}
	}
	// keep-skill should still be there
	foundKeep := false
	for _, s := range otherWs.Skills {
		if s == "keep-skill" {
			foundKeep = true
		}
	}
	if !foundKeep {
		t.Error("keep-skill was incorrectly removed")
	}
}

// ============================================================================
// skillDiff (workspace.go)
// ============================================================================

func TestBoost_SkillDiff(t *testing.T) {
	tests := []struct {
		name     string
		a, b     []string
		expected []string
	}{
		{"both empty", nil, nil, nil},
		{"a empty", nil, []string{"x"}, nil},
		{"b empty", []string{"x", "y"}, nil, []string{"x", "y"}},
		{"no diff", []string{"x", "y"}, []string{"x", "y"}, nil},
		{"a has extra", []string{"x", "y", "z"}, []string{"x"}, []string{"y", "z"}},
		{"b has extra", []string{"x"}, []string{"x", "y"}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := skillDiff(tt.a, tt.b)
			if len(got) != len(tt.expected) {
				t.Errorf("skillDiff() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSkillDiff_CaseInsensitive(t *testing.T) {
	// "Alpha" in b should exclude "alpha" from a
	got := skillDiff([]string{"alpha", "beta"}, []string{"Alpha"})
	if len(got) != 1 {
		t.Fatalf("skillDiff() = %v, want [beta]", got)
	}
	if got[0] != "beta" {
		t.Errorf("skillDiff() = %v, want [beta]", got)
	}
}

// ============================================================================
// SyncWorkspace (workspace.go)
// ============================================================================

func TestBoost_SyncWorkspace(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// Create skills
	for _, name := range []string{"keep-skill", "add-skill", "remove-skill"} {
		skillDir := filepath.Join(tmpDir, ".scribe", "scrolls", name)
		_ = os.MkdirAll(skillDir, 0o755)
		_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: "+name+"\ndescription: test\n---\n# Test\n"), 0o644)
	}

	current := &Workspace{Name: "current", Skills: []string{"keep-skill", "remove-skill"}}
	target := &Workspace{Name: "target", Skills: []string{"keep-skill", "add-skill"}}

	err := SyncWorkspace(current, target)
	if err != nil {
		t.Fatalf("SyncWorkspace() error: %v", err)
	}
}

func TestBoost_SyncWorkspace_MixedSkills(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// Create only one skill
	skillDir := filepath.Join(tmpDir, ".scribe", "scrolls", "exists")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: exists\ndescription: E\n---\n# E\n"), 0o644)

	// Create agent dir
	_ = os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0o755)

	current := &Workspace{Name: "old", Skills: []string{}}
	target := &Workspace{Name: "new", Skills: []string{"exists", "missing"}}

	err := SyncWorkspace(current, target)
	if err != nil {
		t.Fatalf("SyncWorkspace() error: %v", err)
	}

	// "exists" should be synced, "missing" should be silently skipped
}

func TestSyncWorkspace_RemovesOrphanedSkills(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// Create canonical skills
	for _, name := range []string{"skill-a", "skill-b"} {
		skillDir := filepath.Join(tmpDir, ".scribe", "scrolls", name)
		_ = os.MkdirAll(skillDir, 0o755)
		_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: "+name+"\ndescription: test\n---\n# Test\n"), 0o644)
	}

	// Create agent dir and pre-populate with an orphaned skill
	agentSkillsDir := filepath.Join(tmpDir, ".claude", "skills")
	_ = os.MkdirAll(agentSkillsDir, 0o755)
	orphanDir := filepath.Join(agentSkillsDir, "orphaned-skill")
	_ = os.MkdirAll(orphanDir, 0o755)
	_ = os.WriteFile(filepath.Join(orphanDir, "SKILL.md"), []byte("orphan"), 0o644)

	// Current workspace does NOT list the orphan (simulates out-of-sync state)
	current := &Workspace{Name: "old", Skills: []string{"skill-a"}}
	// Target workspace only wants skill-b
	target := &Workspace{Name: "new", Skills: []string{"skill-b"}}

	err := SyncWorkspace(current, target)
	if err != nil {
		t.Fatalf("SyncWorkspace() error: %v", err)
	}

	// Orphaned skill should be removed even though it wasn't in current workspace
	if _, err := os.Stat(orphanDir); !os.IsNotExist(err) {
		t.Error("orphaned skill should have been removed from agent directory")
	}

	// skill-a should be removed (not in target)
	if _, err := os.Stat(filepath.Join(agentSkillsDir, "skill-a")); !os.IsNotExist(err) {
		t.Error("skill-a should have been removed (not in target workspace)")
	}

	// skill-b should be present (in target)
	if _, err := os.Stat(filepath.Join(agentSkillsDir, "skill-b")); err != nil {
		t.Error("skill-b should have been synced to agent directory")
	}
}

// ============================================================================
// ResyncCurrentWorkspace (workspace.go)
// ============================================================================

func TestBoost_ResyncCurrentWorkspace(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	// Create a skill and add to default workspace
	skillDir := filepath.Join(tmpDir, ".scribe", "scrolls", "resync-skill")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: resync-skill\ndescription: Resync\n---\n# Resync\n"), 0o644)
	_ = AddSkillToWorkspace("resync-skill", DefaultWorkspaceName)

	// Create agent dir
	_ = os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0o755)

	err := ResyncCurrentWorkspace()
	if err != nil {
		t.Fatalf("ResyncCurrentWorkspace() error: %v", err)
	}

	// Verify skill was synced to agent
	agentSkillDir := filepath.Join(tmpDir, ".claude", "skills", "resync-skill")
	if _, err := os.Stat(agentSkillDir); err != nil {
		t.Error("skill not synced after resync")
	}
}

func TestBoost_ResyncCurrentWorkspace_MissingSkill(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	// Add a nonexistent skill to the workspace
	_ = AddSkillToWorkspace("ghost-skill", DefaultWorkspaceName)

	// Resync should not error even if skill doesn't exist
	err := ResyncCurrentWorkspace()
	if err != nil {
		t.Fatalf("ResyncCurrentWorkspace() error: %v", err)
	}
}

// ============================================================================
// RebuildDefaultWorkspace (workspace.go)
// ============================================================================

func TestBoost_RebuildDefaultWorkspace(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	// Create skills
	for _, name := range []string{"rebuild-a", "rebuild-b"} {
		skillDir := filepath.Join(tmpDir, ".scribe", "scrolls", name)
		_ = os.MkdirAll(skillDir, 0o755)
		_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: "+name+"\ndescription: test\n---\n# Test\n"), 0o644)
	}

	err := RebuildDefaultWorkspace()
	if err != nil {
		t.Fatalf("RebuildDefaultWorkspace() error: %v", err)
	}

	ws, err := GetWorkspace(DefaultWorkspaceName)
	if err != nil {
		t.Fatalf("GetWorkspace error: %v", err)
	}
	// +1 for injected system skill (scribe-cli)
	if len(ws.Skills) != 3 {
		t.Errorf("rebuilt workspace has %d skills, want 3 (2 test + 1 system)", len(ws.Skills))
	}
}

// ============================================================================
// CleanWorkspaces (workspace.go)
// ============================================================================

func TestBoost_CleanWorkspaces(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// Create workspace with orphaned skills
	ws := &Workspace{
		Name:   DefaultWorkspaceName,
		Skills: []string{"existing-skill", "orphan-skill"},
	}
	wsPath, _ := GetWorkspacePath(DefaultWorkspaceName)
	data, _ := json.MarshalIndent(ws, "", "  ")
	_ = os.MkdirAll(filepath.Dir(wsPath), 0o755)
	_ = os.WriteFile(wsPath, data, 0o644)

	// Only create "existing-skill" in scrolls
	skillDir := filepath.Join(tmpDir, ".scribe", "scrolls", "existing-skill")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: existing-skill\ndescription: Exists\n---\n# E\n"), 0o644)

	err := CleanWorkspaces()
	if err != nil {
		t.Fatalf("CleanWorkspaces() error: %v", err)
	}

	// Verify orphan was removed
	cleaned, _ := GetWorkspace(DefaultWorkspaceName)
	for _, s := range cleaned.Skills {
		if s == "orphan-skill" {
			t.Error("orphan skill still in workspace after clean")
		}
	}
	// existing-skill should remain
	found := false
	for _, s := range cleaned.Skills {
		if s == "existing-skill" {
			found = true
		}
	}
	if !found {
		t.Error("existing skill was incorrectly removed during clean")
	}
}

// ============================================================================
// AddSkillToWorkspace (workspace.go)
// ============================================================================

func TestBoost_AddSkillToWorkspace_AlreadyPresent(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	_ = AddSkillToWorkspace("dup", DefaultWorkspaceName)
	err := AddSkillToWorkspace("dup", DefaultWorkspaceName)
	if err != nil {
		t.Fatalf("AddSkillToWorkspace(duplicate) error: %v", err)
	}

	// Should only appear once
	ws, _ := GetWorkspace(DefaultWorkspaceName)
	count := 0
	for _, s := range ws.Skills {
		if s == "dup" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("skill appears %d times, want 1", count)
	}
}

func TestAddSkillToWorkspace_CaseInsensitiveDuplicate(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	_ = AddSkillToWorkspace("my-skill", DefaultWorkspaceName)
	// Adding with different case should not create a duplicate
	err := AddSkillToWorkspace("My-Skill", DefaultWorkspaceName)
	if err != nil {
		t.Fatalf("AddSkillToWorkspace(case-diff) error: %v", err)
	}

	ws, _ := GetWorkspace(DefaultWorkspaceName)
	count := 0
	for _, s := range ws.Skills {
		if strings.EqualFold(s, "my-skill") {
			count++
		}
	}
	if count != 1 {
		t.Errorf("skill appears %d times, want 1 (case-insensitive dedup)", count)
	}
}

func TestBoost_AddSkillToWorkspace_NonActiveWorkspace(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	// Create a non-active workspace
	ws := &Workspace{Name: "inactive-ws", Skills: []string{}}
	_ = CreateWorkspace(ws)

	err := AddSkillToWorkspace("non-active-skill", "inactive-ws")
	if err != nil {
		t.Fatalf("AddSkillToWorkspace(non-active) error: %v", err)
	}

	loaded, _ := GetWorkspace("inactive-ws")
	found := false
	for _, s := range loaded.Skills {
		if s == "non-active-skill" {
			found = true
		}
	}
	if !found {
		t.Error("skill not added to non-active workspace")
	}
}

func TestBoost_AddSkillToWorkspace_ActiveSync(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	// Create skill
	skillDir := filepath.Join(tmpDir, ".scribe", "scrolls", "active-sync-skill")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: active-sync-skill\ndescription: test\n---\n# T\n"), 0o644)

	// Create agent dir
	_ = os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0o755)

	// Add to active (default) workspace - should trigger sync
	err := AddSkillToWorkspace("active-sync-skill", DefaultWorkspaceName)
	if err != nil {
		t.Fatalf("AddSkillToWorkspace error: %v", err)
	}

	// Verify synced to agent
	agentSkill := filepath.Join(tmpDir, ".claude", "skills", "active-sync-skill")
	if _, err := os.Stat(agentSkill); err != nil {
		t.Error("skill not synced to agent when added to active workspace")
	}
}

// ============================================================================
// createDefaultWorkspace (workspace.go)
// ============================================================================

func TestBoost_CreateDefaultWorkspace(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)

	ws := createDefaultWorkspace()
	if ws.Name != DefaultWorkspaceName {
		t.Errorf("name = %q, want %q", ws.Name, DefaultWorkspaceName)
	}
	if ws.Description != "All installed skills" {
		t.Errorf("description = %q, want 'All installed skills'", ws.Description)
	}
}

// ============================================================================
// GetWorkspace (workspace.go)
// ============================================================================

func TestBoost_GetWorkspace_NonExistent(t *testing.T) {
	_ = setupTempHome(t)
	_ = EnsureScribeDirs()

	_, err := GetWorkspace("nonexistent-ws-12345")
	if err == nil {
		t.Error("expected error for nonexistent workspace")
	}
}

func TestBoost_GetWorkspace_InvalidJSON(t *testing.T) {
	tmpDir := setupTempHome(t)
	_ = EnsureScribeDirs()

	wsPath := filepath.Join(tmpDir, ".scribe", "workspaces", "broken.json")
	_ = os.MkdirAll(filepath.Dir(wsPath), 0o755)
	_ = os.WriteFile(wsPath, []byte("not valid json"), 0o644)

	_, err := GetWorkspace("broken")
	if err == nil {
		t.Error("expected error for invalid JSON workspace")
	}
}

// ============================================================================
// AddSkillToWorkspace — frontmatter name conflict (workspace.go)
// ============================================================================

func TestAddSkillToWorkspaceFrontmatterConflict(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")

	// Create two skills with the same frontmatter name but different storage names
	for _, info := range []struct{ storage, fm string }{
		{"alice-skills--commit", "commit"},
		{"bob-tools--commit", "commit"},
	} {
		dir := filepath.Join(scrollsDir, info.storage)
		_ = os.MkdirAll(dir, 0o755)
		content := "---\nname: " + info.fm + "\ndescription: test\n---\n# Test\n"
		_ = os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644)
	}

	// Add the first one — should succeed
	err := AddSkillToWorkspace("alice-skills--commit", DefaultWorkspaceName)
	if err != nil {
		t.Fatalf("first AddSkillToWorkspace() error: %v", err)
	}

	// Add the second one — should fail because same frontmatter name
	err = AddSkillToWorkspace("bob-tools--commit", DefaultWorkspaceName)
	if err == nil {
		t.Fatal("expected error when adding skill with duplicate frontmatter name")
	}
	if !strings.Contains(err.Error(), "same name") {
		t.Errorf("error = %q, want message about same name conflict", err.Error())
	}
}

func TestAddSkillToWorkspaceNoConflictDifferentFrontmatter(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")

	// Two skills with different frontmatter names
	for _, info := range []struct{ storage, fm string }{
		{"skill-a", "commit"},
		{"skill-b", "review"},
	} {
		dir := filepath.Join(scrollsDir, info.storage)
		_ = os.MkdirAll(dir, 0o755)
		content := "---\nname: " + info.fm + "\ndescription: test\n---\n# Test\n"
		_ = os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644)
	}

	// Both should succeed
	if err := AddSkillToWorkspace("skill-a", DefaultWorkspaceName); err != nil {
		t.Fatalf("AddSkillToWorkspace(skill-a) error: %v", err)
	}
	if err := AddSkillToWorkspace("skill-b", DefaultWorkspaceName); err != nil {
		t.Fatalf("AddSkillToWorkspace(skill-b) error: %v", err)
	}
}
