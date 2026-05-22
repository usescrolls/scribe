package scribe

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// ============================================================================
// IsSystemSkill
// ============================================================================

func TestIsSystemSkill(t *testing.T) {
	if !IsSystemSkill("scribe-cli") {
		t.Error("IsSystemSkill('scribe-cli') should be true")
	}
	if IsSystemSkill("scribe-welcome") {
		t.Error("IsSystemSkill('scribe-welcome') should be false")
	}
	if IsSystemSkill("random-skill") {
		t.Error("IsSystemSkill('random-skill') should be false")
	}
	if IsSystemSkill("") {
		t.Error("IsSystemSkill('') should be false")
	}
}

// ============================================================================
// EnsureSystemSkill
// ============================================================================

func TestEnsureSystemSkill(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// First call: creates the skill
	err = EnsureSystemSkill()
	if err != nil {
		t.Fatalf("EnsureSystemSkill() error: %v", err)
	}

	// Verify SKILL.md exists with correct content
	skillPath := filepath.Join(tmpDir, ".scribe", "scrolls", SystemSkillName, SkillFileName)
	content, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("failed to read SKILL.md: %v", err)
	}
	if string(content) != SystemSkillContent {
		t.Error("SKILL.md content doesn't match SystemSkillContent")
	}

	// Verify metadata exists
	metaPath := filepath.Join(tmpDir, ".scribe", "scrolls", SystemSkillName, MetaFileName)
	meta, err := ReadSkillMeta(metaPath)
	if err != nil {
		t.Fatalf("failed to read meta: %v", err)
	}
	if meta.SourceType != "system" {
		t.Errorf("meta.SourceType = %q, want 'system'", meta.SourceType)
	}
	if meta.Source != "scribe" {
		t.Errorf("meta.Source = %q, want 'scribe'", meta.Source)
	}
	if meta.ContentHash != ComputeContentHash(SystemSkillContent) {
		t.Errorf("meta.ContentHash mismatch")
	}
}

func TestEnsureSystemSkill_Idempotent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// First call
	_ = EnsureSystemSkill()

	metaPath := filepath.Join(tmpDir, ".scribe", "scrolls", SystemSkillName, MetaFileName)
	meta1, _ := ReadSkillMeta(metaPath)
	originalInstalledAt := meta1.InstalledAt

	// Second call should be a no-op
	err = EnsureSystemSkill()
	if err != nil {
		t.Fatalf("second EnsureSystemSkill() error: %v", err)
	}

	meta2, _ := ReadSkillMeta(metaPath)
	if meta2.InstalledAt != originalInstalledAt {
		t.Error("InstalledAt should not change on idempotent call")
	}
}

func TestEnsureSystemSkill_UpdatesContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// Write a stale version of the skill
	skillDir := filepath.Join(tmpDir, ".scribe", "scrolls", SystemSkillName)
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, SkillFileName), []byte("old content"), 0o644)

	metaPath := filepath.Join(skillDir, MetaFileName)
	oldMeta := &SkillMeta{
		Source:      "scribe",
		SourceType:  "system",
		ContentHash: ComputeContentHash("old content"),
		InstalledAt: "2024-01-01T00:00:00Z",
		UpdatedAt:   "2024-01-01T00:00:00Z",
	}
	_ = WriteSkillMeta(metaPath, oldMeta)

	// EnsureSystemSkill should detect the mismatch and update
	err = EnsureSystemSkill()
	if err != nil {
		t.Fatalf("EnsureSystemSkill() error: %v", err)
	}

	// Verify content was updated
	content, _ := os.ReadFile(filepath.Join(skillDir, SkillFileName))
	if string(content) != SystemSkillContent {
		t.Error("SKILL.md should be updated to current SystemSkillContent")
	}

	// Verify installedAt is preserved
	meta, _ := ReadSkillMeta(metaPath)
	if meta.InstalledAt != "2024-01-01T00:00:00Z" {
		t.Errorf("InstalledAt should be preserved, got %q", meta.InstalledAt)
	}

	// Verify updatedAt changed
	if meta.UpdatedAt == "2024-01-01T00:00:00Z" {
		t.Error("UpdatedAt should be updated")
	}
}

func TestEnsureSystemSkill_RepairsMetadataWithoutContentChange(t *testing.T) {
	tmpDir := setupTempHome(t)

	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	skillDir := filepath.Join(tmpDir, ".scribe", "scrolls", SystemSkillName)
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, SkillFileName), []byte(SystemSkillContent), 0o644)
	_ = os.WriteFile(filepath.Join(skillDir, MetaFileName), []byte{}, 0o644)

	err := EnsureSystemSkill()
	if err != nil {
		t.Fatalf("EnsureSystemSkill() error: %v", err)
	}

	meta, err := ReadSkillMeta(filepath.Join(skillDir, MetaFileName))
	if err != nil {
		t.Fatalf("failed to read repaired meta: %v", err)
	}
	if meta.Source != "scribe" || meta.SourceType != "system" {
		t.Fatalf("meta source = %q/%q, want scribe/system", meta.Source, meta.SourceType)
	}
	if meta.ContentHash != ComputeContentHash(SystemSkillContent) {
		t.Errorf("meta.ContentHash = %q, want current system skill hash", meta.ContentHash)
	}
}

func TestReadSkill_RepairsCorruptSystemSkill(t *testing.T) {
	tmpDir := setupTempHome(t)

	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	skillDir := filepath.Join(tmpDir, ".scribe", "scrolls", SystemSkillName)
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, SkillFileName), []byte{}, 0o644)
	_ = os.WriteFile(filepath.Join(skillDir, MetaFileName), []byte{}, 0o644)

	skill, err := ReadSkill(SystemSkillName)
	if err != nil {
		t.Fatalf("ReadSkill(%q) error: %v", SystemSkillName, err)
	}
	if skill.Name != SystemSkillName {
		t.Fatalf("skill.Name = %q, want %q", skill.Name, SystemSkillName)
	}
	if skill.Meta == nil || skill.Meta.SourceType != "system" {
		t.Fatalf("skill.Meta = %#v, want repaired system metadata", skill.Meta)
	}

	content, err := os.ReadFile(filepath.Join(skillDir, SkillFileName))
	if err != nil {
		t.Fatalf("read repaired system skill: %v", err)
	}
	if string(content) != SystemSkillContent {
		t.Error("ReadSkill should repair zero-byte system skill content")
	}
}

// ============================================================================
// UninstallSkill guard
// ============================================================================

func TestUninstallSkill_RejectsSystemSkill(t *testing.T) {
	err := UninstallSkill(SystemSkillName)
	if err == nil {
		t.Fatal("UninstallSkill should reject system skill")
	}
	if !strings.Contains(err.Error(), "cannot uninstall system skill") {
		t.Errorf("error = %q, want 'cannot uninstall system skill'", err)
	}
}

// ============================================================================
// RemoveSkillFromWorkspace guard
// ============================================================================

func TestRemoveSkillFromWorkspace_RejectsSystemSkill(t *testing.T) {
	err := RemoveSkillFromWorkspace(SystemSkillName, "default")
	if err == nil {
		t.Fatal("RemoveSkillFromWorkspace should reject system skill")
	}
	if !strings.Contains(err.Error(), "cannot remove system skill") {
		t.Errorf("error = %q, want 'cannot remove system skill'", err)
	}
}

// ============================================================================
// RemoveSkillFromAllWorkspaces guard
// ============================================================================

func TestRemoveSkillFromAllWorkspaces_IgnoresSystemSkill(t *testing.T) {
	// Should silently return nil for system skills
	err := RemoveSkillFromAllWorkspaces(SystemSkillName)
	if err != nil {
		t.Fatalf("RemoveSkillFromAllWorkspaces(system skill) should return nil, got: %v", err)
	}
}

// ============================================================================
// UpdateSkill guard
// ============================================================================

func TestUpdateSkill_RejectsSystemSkill(t *testing.T) {
	_, err := UpdateSkill(SystemSkillName, false)
	if err == nil {
		t.Fatal("UpdateSkill should reject system skill")
	}
	if !strings.Contains(err.Error(), "cannot update system skill") {
		t.Errorf("error = %q, want 'cannot update system skill'", err)
	}
}

// ============================================================================
// CheckSkillForUpdate guard
// ============================================================================

func TestCheckSkillForUpdate_RejectsSystemSkill(t *testing.T) {
	result := CheckSkillForUpdate(SystemSkillName)
	if result.Error == "" {
		t.Fatal("CheckSkillForUpdate should return error for system skill")
	}
	if !strings.Contains(result.Error, "system skill") {
		t.Errorf("error = %q, want to contain 'system skill'", result.Error)
	}
}

// ============================================================================
// GetSkillInfo sets IsSystem
// ============================================================================

func TestGetSkillInfo_SetsIsSystem(t *testing.T) {
	// System skill
	systemSkill := &Skill{
		Name:        SystemSkillName,
		Description: "System skill",
	}
	info := GetSkillInfo(systemSkill)
	if !info.IsSystem {
		t.Error("GetSkillInfo should set IsSystem=true for system skill")
	}

	// Normal skill
	normalSkill := &Skill{
		Name:        "normal-skill",
		Description: "Normal skill",
	}
	info = GetSkillInfo(normalSkill)
	if info.IsSystem {
		t.Error("GetSkillInfo should set IsSystem=false for normal skill")
	}
}

// ============================================================================
// Workspace injection
// ============================================================================

func TestGetWorkspace_InjectsSystemSkill(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	_ = EnsureScribeDirs()

	// Create workspace without system skill
	ws := &Workspace{
		Name:   "test-ws",
		Skills: []string{"skill-a", "skill-b"},
	}
	_ = CreateWorkspace(ws)

	// Read it back - should have system skill injected
	loaded, err := GetWorkspace("test-ws")
	if err != nil {
		t.Fatalf("GetWorkspace error: %v", err)
	}

	hasSystem := slices.Contains(loaded.Skills, SystemSkillName)
	if !hasSystem {
		t.Errorf("workspace should contain system skill, got %v", loaded.Skills)
	}

	// System skill should be first
	if loaded.Skills[0] != SystemSkillName {
		t.Errorf("system skill should be first, got %q", loaded.Skills[0])
	}
}

func TestCreateDefaultWorkspace_InjectsSystemSkill(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	_ = EnsureScribeDirs()

	ws := createDefaultWorkspace()
	hasSystem := slices.Contains(ws.Skills, SystemSkillName)
	if !hasSystem {
		t.Errorf("default workspace should contain system skill, got %v", ws.Skills)
	}
}

// ============================================================================
// Save-time stripping
// ============================================================================

func TestSaveWorkspace_StripsSystemSkills(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	_ = EnsureScribeDirs()

	// Save a workspace that includes system skill
	ws := &Workspace{
		Name:   "strip-test",
		Skills: []string{SystemSkillName, "skill-a", "skill-b"},
	}
	err = saveWorkspace(ws)
	if err != nil {
		t.Fatalf("saveWorkspace error: %v", err)
	}

	// Read the raw JSON from disk - system skill should NOT be there
	wsPath, _ := GetWorkspacePath("strip-test")
	data, err := os.ReadFile(wsPath)
	if err != nil {
		t.Fatalf("failed to read workspace file: %v", err)
	}

	var rawWs Workspace
	_ = json.Unmarshal(data, &rawWs)

	for _, s := range rawWs.Skills {
		if s == SystemSkillName {
			t.Error("system skill should be stripped from persisted workspace JSON")
		}
	}
	if len(rawWs.Skills) != 2 {
		t.Errorf("persisted skills count = %d, want 2", len(rawWs.Skills))
	}

	// But reading via GetWorkspace should re-inject it
	loaded, _ := GetWorkspace("strip-test")
	if len(loaded.Skills) != 3 {
		t.Errorf("loaded skills count = %d, want 3 (2 test + 1 system)", len(loaded.Skills))
	}
}

// ============================================================================
// SystemSkillContent validity
// ============================================================================

func TestSystemSkillContent_HasValidFrontmatter(t *testing.T) {
	skill, err := ParseSkillContent(SystemSkillContent, "")
	if err != nil {
		t.Fatalf("SystemSkillContent failed to parse: %v", err)
	}
	if skill.Name != SystemSkillName {
		t.Errorf("skill name = %q, want %q", skill.Name, SystemSkillName)
	}
	if skill.Description == "" {
		t.Error("skill description should not be empty")
	}
}

// ============================================================================
// injectSystemSkills helper
// ============================================================================

func TestInjectSystemSkills(t *testing.T) {
	// Empty list
	result := injectSystemSkills([]string{})
	if len(result) != 1 || result[0] != SystemSkillName {
		t.Errorf("empty injection = %v, want [%s]", result, SystemSkillName)
	}

	// Already present
	result = injectSystemSkills([]string{SystemSkillName, "other"})
	if len(result) != 2 {
		t.Errorf("already present = %v, want length 2", result)
	}

	// Not present - should prepend
	result = injectSystemSkills([]string{"skill-a", "skill-b"})
	if len(result) != 3 {
		t.Errorf("injection = %v, want length 3", result)
	}
	if result[0] != SystemSkillName {
		t.Errorf("first element = %q, want %q", result[0], SystemSkillName)
	}
}

// ============================================================================
// SyncWorkspace correctness with system skill
// ============================================================================

func TestSyncWorkspace_SystemSkillNotInDiff(t *testing.T) {
	// Both workspaces have system skill via injection - diff should not include it
	current := &Workspace{
		Name:   "current",
		Skills: []string{SystemSkillName, "skill-a"},
	}
	target := &Workspace{
		Name:   "target",
		Skills: []string{SystemSkillName, "skill-b"},
	}

	toRemove := skillDiff(current.Skills, target.Skills)
	toAdd := skillDiff(target.Skills, current.Skills)

	for _, s := range toRemove {
		if s == SystemSkillName {
			t.Error("system skill should not appear in toRemove diff")
		}
	}
	for _, s := range toAdd {
		if s == SystemSkillName {
			t.Error("system skill should not appear in toAdd diff")
		}
	}
}
