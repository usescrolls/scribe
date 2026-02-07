package scribe

import (
	"os"
	"path/filepath"
	"testing"
)

// ============================================================================
// IsOnboardingCompleted / CompleteOnboarding (onboarding.go)
// ============================================================================

func TestIsOnboardingCompleted_Default(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	InitLoggerCLI(false)

	completed, err := IsOnboardingCompleted()
	if err != nil {
		t.Fatalf("IsOnboardingCompleted() error: %v", err)
	}
	if completed {
		t.Error("IsOnboardingCompleted() = true for fresh install, want false")
	}
}

func TestCompleteOnboarding(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	InitLoggerCLI(false)

	// Complete onboarding
	if err := CompleteOnboarding(); err != nil {
		t.Fatalf("CompleteOnboarding() error: %v", err)
	}

	// Verify
	completed, err := IsOnboardingCompleted()
	if err != nil {
		t.Fatalf("IsOnboardingCompleted() error: %v", err)
	}
	if !completed {
		t.Error("IsOnboardingCompleted() = false after CompleteOnboarding(), want true")
	}
}

func TestCompleteOnboarding_Idempotent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	InitLoggerCLI(false)

	// Complete twice
	if err := CompleteOnboarding(); err != nil {
		t.Fatalf("CompleteOnboarding() first call error: %v", err)
	}
	if err := CompleteOnboarding(); err != nil {
		t.Fatalf("CompleteOnboarding() second call error: %v", err)
	}

	completed, err := IsOnboardingCompleted()
	if err != nil {
		t.Fatalf("IsOnboardingCompleted() error: %v", err)
	}
	if !completed {
		t.Error("IsOnboardingCompleted() = false after double CompleteOnboarding()")
	}
}

// ============================================================================
// DetectSkillConflicts (onboarding.go)
// ============================================================================

func TestDetectSkillConflicts_NoConflicts(t *testing.T) {
	skills := []ExistingSkillInfo{
		{Name: "skill-a", Path: "/a", AgentID: "claude-code"},
		{Name: "skill-b", Path: "/b", AgentID: "cursor"},
	}
	conflicts := DetectSkillConflicts(skills)
	if len(conflicts) != 0 {
		t.Errorf("DetectSkillConflicts no conflicts: got %d, want 0", len(conflicts))
	}
}

func TestDetectSkillConflicts_WithConflicts(t *testing.T) {
	skills := []ExistingSkillInfo{
		{Name: "shared-skill", Path: "/claude/shared-skill", AgentID: "claude-code", AgentName: "Claude Code"},
		{Name: "unique-skill", Path: "/cursor/unique-skill", AgentID: "cursor", AgentName: "Cursor"},
		{Name: "shared-skill", Path: "/cursor/shared-skill", AgentID: "cursor", AgentName: "Cursor"},
	}
	conflicts := DetectSkillConflicts(skills)
	if len(conflicts) != 1 {
		t.Fatalf("DetectSkillConflicts: got %d conflicts, want 1", len(conflicts))
	}
	if conflicts[0].Name != "shared-skill" {
		t.Errorf("conflict name = %q, want 'shared-skill'", conflicts[0].Name)
	}
	if len(conflicts[0].Sources) != 2 {
		t.Errorf("conflict sources = %d, want 2", len(conflicts[0].Sources))
	}
}

func TestDetectSkillConflicts_Empty(t *testing.T) {
	conflicts := DetectSkillConflicts([]ExistingSkillInfo{})
	if len(conflicts) != 0 {
		t.Errorf("DetectSkillConflicts empty: got %d, want 0", len(conflicts))
	}
}

func TestDetectSkillConflicts_AllConflicting(t *testing.T) {
	skills := []ExistingSkillInfo{
		{Name: "dup", Path: "/a/dup", AgentID: "agent-a"},
		{Name: "dup", Path: "/b/dup", AgentID: "agent-b"},
		{Name: "dup", Path: "/c/dup", AgentID: "agent-c"},
	}
	conflicts := DetectSkillConflicts(skills)
	if len(conflicts) != 1 {
		t.Fatalf("DetectSkillConflicts all conflicting: got %d, want 1", len(conflicts))
	}
	if len(conflicts[0].Sources) != 3 {
		t.Errorf("conflict sources = %d, want 3", len(conflicts[0].Sources))
	}
}

// ============================================================================
// DeleteExistingSkills (onboarding.go)
// ============================================================================

func TestDeleteExistingSkills(t *testing.T) {
	InitLoggerCLI(false)

	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create skill directories
	skillA := filepath.Join(tmpDir, "skill-a")
	skillB := filepath.Join(tmpDir, "skill-b")
	_ = os.MkdirAll(skillA, 0o755)
	_ = os.MkdirAll(skillB, 0o755)
	_ = os.WriteFile(filepath.Join(skillA, "SKILL.md"), []byte("# A"), 0o644)
	_ = os.WriteFile(filepath.Join(skillB, "SKILL.md"), []byte("# B"), 0o644)

	skills := []ExistingSkillInfo{
		{Name: "skill-a", Path: skillA, AgentID: "test"},
		{Name: "skill-b", Path: skillB, AgentID: "test"},
	}

	err = DeleteExistingSkills(skills)
	if err != nil {
		t.Fatalf("DeleteExistingSkills() error: %v", err)
	}

	// Verify directories are gone
	if _, err := os.Stat(skillA); !os.IsNotExist(err) {
		t.Error("skill-a directory still exists after deletion")
	}
	if _, err := os.Stat(skillB); !os.IsNotExist(err) {
		t.Error("skill-b directory still exists after deletion")
	}
}

func TestDeleteExistingSkills_Empty(t *testing.T) {
	InitLoggerCLI(false)
	err := DeleteExistingSkills([]ExistingSkillInfo{})
	if err != nil {
		t.Fatalf("DeleteExistingSkills(empty) error: %v", err)
	}
}

func TestDeleteExistingSkills_AlreadyGone(t *testing.T) {
	InitLoggerCLI(false)

	// Path that does not exist - RemoveAll on non-existent path is not an error
	skills := []ExistingSkillInfo{
		{Name: "ghost", Path: "/tmp/nonexistent-scribe-test-path-12345", AgentID: "test"},
	}
	err := DeleteExistingSkills(skills)
	if err != nil {
		t.Fatalf("DeleteExistingSkills(non-existent) error: %v", err)
	}
}

// ============================================================================
// InstallDemoSkill (onboarding.go)
// ============================================================================

func TestInstallDemoSkill(t *testing.T) {
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

	err = InstallDemoSkill()
	if err != nil {
		t.Fatalf("InstallDemoSkill() error: %v", err)
	}

	// Verify SKILL.md was created
	skillPath := filepath.Join(tmpDir, ".scribe", "scrolls", "scribe-welcome", "SKILL.md")
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		t.Error("InstallDemoSkill() did not create SKILL.md")
	}

	// Verify meta was created
	metaPath := filepath.Join(tmpDir, ".scribe", "scrolls", "scribe-welcome", ".scribe-meta.json")
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Error("InstallDemoSkill() did not create .scribe-meta.json")
	}

	// Verify content matches DemoSkillContent
	content, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("failed to read SKILL.md: %v", err)
	}
	if string(content) != DemoSkillContent {
		t.Error("InstallDemoSkill() SKILL.md content does not match DemoSkillContent")
	}
}

func TestInstallDemoSkill_Idempotent(t *testing.T) {
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

	// Install twice - should not error
	if err := InstallDemoSkill(); err != nil {
		t.Fatalf("InstallDemoSkill() first call error: %v", err)
	}
	if err := InstallDemoSkill(); err != nil {
		t.Fatalf("InstallDemoSkill() second call error: %v", err)
	}
}

// ============================================================================
// moveDir (onboarding.go)
// ============================================================================

func TestMoveDir_Rename(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	src := filepath.Join(tmpDir, "source")
	dst := filepath.Join(tmpDir, "destination")
	_ = os.MkdirAll(src, 0o755)
	_ = os.WriteFile(filepath.Join(src, "file.txt"), []byte("hello"), 0o644)

	err = moveDir(src, dst)
	if err != nil {
		t.Fatalf("moveDir() error: %v", err)
	}

	// Source should be gone
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Error("moveDir: source directory still exists")
	}

	// Destination should exist with the file
	data, err := os.ReadFile(filepath.Join(dst, "file.txt"))
	if err != nil {
		t.Fatalf("moveDir: failed to read moved file: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("moveDir: file content = %q, want 'hello'", string(data))
	}
}

func TestMoveDir_WithSubdirectories(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	src := filepath.Join(tmpDir, "src")
	dst := filepath.Join(tmpDir, "dst")
	_ = os.MkdirAll(filepath.Join(src, "sub", "deep"), 0o755)
	_ = os.WriteFile(filepath.Join(src, "sub", "deep", "data.txt"), []byte("nested"), 0o644)

	err = moveDir(src, dst)
	if err != nil {
		t.Fatalf("moveDir() error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dst, "sub", "deep", "data.txt"))
	if err != nil {
		t.Fatalf("failed to read nested file: %v", err)
	}
	if string(data) != "nested" {
		t.Errorf("nested file content = %q, want 'nested'", string(data))
	}
}

// ============================================================================
// DetectExistingSkills (onboarding.go)
// ============================================================================

func TestBoost_DetectExistingSkills_NoAgentsInstalled(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)

	skills, err := DetectExistingSkills()
	if err != nil {
		t.Fatalf("DetectExistingSkills() error: %v", err)
	}
	if len(skills) != 0 {
		t.Errorf("expected 0 skills with no agents, got %d", len(skills))
	}
}

func TestBoost_DetectExistingSkills_WithSkills(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)

	// Create claude-code agent dir with skills
	claudeSkillsDir := filepath.Join(tmpDir, ".claude", "skills")
	skill1Dir := filepath.Join(claudeSkillsDir, "my-skill")
	_ = os.MkdirAll(skill1Dir, 0o755)
	_ = os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte("---\nname: my-skill\ndescription: My\n---\n# My\n"), 0o644)

	// Create a directory without SKILL.md (should be ignored)
	noSkillDir := filepath.Join(claudeSkillsDir, "not-a-skill")
	_ = os.MkdirAll(noSkillDir, 0o755)

	// Create a file in skills dir (not a directory, should be ignored)
	_ = os.WriteFile(filepath.Join(claudeSkillsDir, "random.txt"), []byte("nope"), 0o644)

	skills, err := DetectExistingSkills()
	if err != nil {
		t.Fatalf("DetectExistingSkills() error: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if skills[0].Name != "my-skill" {
		t.Errorf("skill name = %q, want 'my-skill'", skills[0].Name)
	}
	if skills[0].AgentID != "claude-code" {
		t.Errorf("agent ID = %q, want 'claude-code'", skills[0].AgentID)
	}
	if skills[0].IsGitRepo {
		t.Error("expected IsGitRepo=false")
	}
}

func TestBoost_DetectExistingSkills_WithGitRepo(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)

	// Create a skill with a .git directory
	claudeSkillsDir := filepath.Join(tmpDir, ".claude", "skills")
	skillDir := filepath.Join(claudeSkillsDir, "git-skill")
	_ = os.MkdirAll(filepath.Join(skillDir, ".git"), 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: git-skill\ndescription: Git\n---\n# Git\n"), 0o644)

	skills, err := DetectExistingSkills()
	if err != nil {
		t.Fatalf("DetectExistingSkills() error: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if !skills[0].IsGitRepo {
		t.Error("expected IsGitRepo=true for skill with .git dir")
	}
}

// ============================================================================
// ImportExistingSkills (onboarding.go)
// ============================================================================

func TestBoost_ImportExistingSkills_SingleSkill(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// Create a skill in a "source" location
	srcDir := filepath.Join(tmpDir, "agent-skills", "import-skill")
	_ = os.MkdirAll(srcDir, 0o755)
	_ = os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte("---\nname: import-skill\ndescription: Import\n---\n# Import\n"), 0o644)

	skills := []ExistingSkillInfo{
		{
			Name:      "import-skill",
			Path:      srcDir,
			AgentID:   "claude-code",
			AgentName: "Claude Code",
		},
	}

	err := ImportExistingSkills(skills)
	if err != nil {
		t.Fatalf("ImportExistingSkills() error: %v", err)
	}

	// Verify skill was moved to scrolls dir
	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	importedDir := filepath.Join(scrollsDir, "import-skill")
	if _, err := os.Stat(filepath.Join(importedDir, "SKILL.md")); err != nil {
		t.Error("SKILL.md not found in scrolls dir after import")
	}

	// Verify metadata was created
	metaPath := filepath.Join(importedDir, ".scribe-meta.json")
	if _, err := os.Stat(metaPath); err != nil {
		t.Error("metadata not created after import")
	}

	// Verify source was removed (moved)
	if _, err := os.Stat(srcDir); !os.IsNotExist(err) {
		t.Error("source directory still exists after import (should have been moved)")
	}
}

func TestBoost_ImportExistingSkills_DuplicateSkipped(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// Create two skills with the same name from different agents
	src1 := filepath.Join(tmpDir, "agent1", "dup-skill")
	src2 := filepath.Join(tmpDir, "agent2", "dup-skill")
	_ = os.MkdirAll(src1, 0o755)
	_ = os.MkdirAll(src2, 0o755)
	_ = os.WriteFile(filepath.Join(src1, "SKILL.md"), []byte("---\nname: dup-skill\ndescription: First\n---\n# First\n"), 0o644)
	_ = os.WriteFile(filepath.Join(src2, "SKILL.md"), []byte("---\nname: dup-skill\ndescription: Second\n---\n# Second\n"), 0o644)

	skills := []ExistingSkillInfo{
		{Name: "dup-skill", Path: src1, AgentID: "agent-a"},
		{Name: "dup-skill", Path: src2, AgentID: "agent-b"},
	}

	err := ImportExistingSkills(skills)
	if err != nil {
		t.Fatalf("ImportExistingSkills() error: %v", err)
	}

	// Only the first one should be imported
	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	importedDir := filepath.Join(scrollsDir, "dup-skill")
	if _, err := os.Stat(filepath.Join(importedDir, "SKILL.md")); err != nil {
		t.Error("SKILL.md not found after import of first duplicate")
	}

	// Second source should still exist (it was skipped)
	if _, err := os.Stat(src2); os.IsNotExist(err) {
		t.Error("second duplicate source was removed, but should have been skipped")
	}
}

func TestBoost_ImportExistingSkills_EmptyList(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	err := ImportExistingSkills([]ExistingSkillInfo{})
	if err != nil {
		t.Fatalf("ImportExistingSkills(empty) error: %v", err)
	}
}

// ============================================================================
// ImportSelectedSkills (onboarding.go)
// ============================================================================

func TestBoost_ImportSelectedSkills_NoMatchingPaths(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// Call with paths that don't match any detected skills
	err := ImportSelectedSkills([]string{"/nonexistent/path/1", "/nonexistent/path/2"})
	if err != nil {
		t.Fatalf("ImportSelectedSkills() error: %v", err)
	}
}

func TestBoost_ImportSelectedSkills_WithMatchingPaths(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// Create claude-code agent dir with skills
	claudeSkillsDir := filepath.Join(tmpDir, ".claude", "skills")
	skill1Dir := filepath.Join(claudeSkillsDir, "select-skill")
	_ = os.MkdirAll(skill1Dir, 0o755)
	_ = os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte("---\nname: select-skill\ndescription: Selected\n---\n# Selected\n"), 0o644)

	// Also create a skill we won't select
	skill2Dir := filepath.Join(claudeSkillsDir, "skip-skill")
	_ = os.MkdirAll(skill2Dir, 0o755)
	_ = os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte("---\nname: skip-skill\ndescription: Skipped\n---\n# Skipped\n"), 0o644)

	// Only import the first skill
	err := ImportSelectedSkills([]string{skill1Dir})
	if err != nil {
		t.Fatalf("ImportSelectedSkills() error: %v", err)
	}

	// Verify selected skill was imported
	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	if _, err := os.Stat(filepath.Join(scrollsDir, "select-skill", "SKILL.md")); err != nil {
		t.Error("selected skill not imported")
	}
}
