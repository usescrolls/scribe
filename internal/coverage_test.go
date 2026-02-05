package scribe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ============================================================================
// FormatSource (url_scheme.go:173) - public function
// ============================================================================

func TestFormatSource_GitHub(t *testing.T) {
	source := &SourceInfo{Type: "github", Owner: "octocat", Repo: "hello-world"}
	got := FormatSource(source)
	if got != "github:octocat/hello-world" {
		t.Errorf("FormatSource(github) = %q, want %q", got, "github:octocat/hello-world")
	}
}

func TestFormatSource_GitHubWithRef(t *testing.T) {
	source := &SourceInfo{Type: "github", Owner: "octocat", Repo: "hello-world", Ref: "v2.0"}
	got := FormatSource(source)
	if got != "github:octocat/hello-world#v2.0" {
		t.Errorf("FormatSource(github+ref) = %q, want %q", got, "github:octocat/hello-world#v2.0")
	}
}

func TestFormatSource_GitHubWithSubpath(t *testing.T) {
	source := &SourceInfo{Type: "github", Owner: "org", Repo: "mono", Subpath: "packages/skill"}
	got := FormatSource(source)
	if got != "github:org/mono/packages/skill" {
		t.Errorf("FormatSource(github+subpath) = %q, want %q", got, "github:org/mono/packages/skill")
	}
}

func TestFormatSource_GitHubWithRefAndSubpath(t *testing.T) {
	source := &SourceInfo{Type: "github", Owner: "org", Repo: "mono", Ref: "dev", Subpath: "skills/a"}
	got := FormatSource(source)
	if got != "github:org/mono#dev/skills/a" {
		t.Errorf("FormatSource(github+ref+subpath) = %q, want %q", got, "github:org/mono#dev/skills/a")
	}
}

func TestFormatSource_GitLab(t *testing.T) {
	source := &SourceInfo{Type: "gitlab", Owner: "group", Repo: "project"}
	got := FormatSource(source)
	if got != "gitlab:group/project" {
		t.Errorf("FormatSource(gitlab) = %q, want %q", got, "gitlab:group/project")
	}
}

func TestFormatSource_GitLabWithRef(t *testing.T) {
	// FormatSource for gitlab does not append ref (unlike formatSource in meta.go)
	source := &SourceInfo{Type: "gitlab", Owner: "group", Repo: "project", Ref: "v1"}
	got := FormatSource(source)
	// FormatSource in url_scheme.go doesn't include ref for gitlab
	expected := "gitlab:group/project"
	if got != expected {
		t.Errorf("FormatSource(gitlab+ref) = %q, want %q", got, expected)
	}
}

func TestFormatSource_Local(t *testing.T) {
	source := &SourceInfo{Type: "local", LocalPath: "/home/user/my-skill"}
	got := FormatSource(source)
	if got != "local:/home/user/my-skill" {
		t.Errorf("FormatSource(local) = %q, want %q", got, "local:/home/user/my-skill")
	}
}

func TestFormatSource_Zip(t *testing.T) {
	source := &SourceInfo{Type: "zip", URL: "https://example.com/skills.zip"}
	got := FormatSource(source)
	if got != "zip:https://example.com/skills.zip" {
		t.Errorf("FormatSource(zip) = %q, want %q", got, "zip:https://example.com/skills.zip")
	}
}

func TestFormatSource_WellKnown(t *testing.T) {
	source := &SourceInfo{Type: "well-known", URL: "https://example.com/.well-known/skills"}
	got := FormatSource(source)
	if got != "https://example.com/.well-known/skills" {
		t.Errorf("FormatSource(well-known) = %q, want %q", got, "https://example.com/.well-known/skills")
	}
}

func TestFormatSource_Unknown(t *testing.T) {
	source := &SourceInfo{Type: "custom", URL: "https://custom.example.com/something"}
	got := FormatSource(source)
	if got != "https://custom.example.com/something" {
		t.Errorf("FormatSource(unknown) = %q, want %q", got, "https://custom.example.com/something")
	}
}

// ============================================================================
// filterSkillsByName (url_scheme.go:162)
// ============================================================================

func TestFilterSkillsByName_Match(t *testing.T) {
	skills := []*Skill{
		{Name: "alpha", Description: "A"},
		{Name: "beta", Description: "B"},
		{Name: "gamma", Description: "C"},
	}
	result := filterSkillsByName(skills, "beta")
	if len(result) != 1 || result[0].Name != "beta" {
		t.Errorf("filterSkillsByName match: got %d results, want 1 with name 'beta'", len(result))
	}
}

func TestFilterSkillsByName_NoMatch(t *testing.T) {
	skills := []*Skill{
		{Name: "alpha", Description: "A"},
		{Name: "beta", Description: "B"},
	}
	result := filterSkillsByName(skills, "omega")
	if len(result) != 0 {
		t.Errorf("filterSkillsByName no match: got %d results, want 0", len(result))
	}
}

func TestFilterSkillsByName_EmptyList(t *testing.T) {
	result := filterSkillsByName([]*Skill{}, "anything")
	if len(result) != 0 {
		t.Errorf("filterSkillsByName empty list: got %d results, want 0", len(result))
	}
}

func TestFilterSkillsByName_NilList(t *testing.T) {
	result := filterSkillsByName(nil, "anything")
	if len(result) != 0 {
		t.Errorf("filterSkillsByName nil list: got %d results, want 0", len(result))
	}
}

func TestFilterSkillsByName_MultipleMatches(t *testing.T) {
	// Two skills with the same name
	skills := []*Skill{
		{Name: "dup", Description: "First"},
		{Name: "other", Description: "Other"},
		{Name: "dup", Description: "Second"},
	}
	result := filterSkillsByName(skills, "dup")
	if len(result) != 2 {
		t.Errorf("filterSkillsByName multiple matches: got %d results, want 2", len(result))
	}
}

// ============================================================================
// ParseInstallURL edge cases (url_scheme.go:96)
// ============================================================================

func TestParseInstallURL_GitLabSource(t *testing.T) {
	source, skill, err := ParseInstallURL("agenthub://install?source=gitlab&repo=myorg/myproject")
	if err != nil {
		t.Fatalf("ParseInstallURL(gitlab) error: %v", err)
	}
	if source.Type != "gitlab" {
		t.Errorf("Type = %q, want 'gitlab'", source.Type)
	}
	if source.Owner != "myorg" || source.Repo != "myproject" {
		t.Errorf("Owner/Repo = %q/%q, want 'myorg'/'myproject'", source.Owner, source.Repo)
	}
	if source.URL != "https://gitlab.com/myorg/myproject" {
		t.Errorf("URL = %q, want 'https://gitlab.com/myorg/myproject'", source.URL)
	}
	if skill != "" {
		t.Errorf("skill = %q, want empty", skill)
	}
}

func TestParseInstallURL_ZipSource(t *testing.T) {
	zipURL := "https://example.com/archive.zip"
	source, _, err := ParseInstallURL("agenthub://install?source=zip&repo=" + zipURL)
	if err != nil {
		t.Fatalf("ParseInstallURL(zip) error: %v", err)
	}
	if source.Type != "zip" {
		t.Errorf("Type = %q, want 'zip'", source.Type)
	}
	if source.URL != zipURL {
		t.Errorf("URL = %q, want %q", source.URL, zipURL)
	}
}

func TestParseInstallURL_UrlSource(t *testing.T) {
	urlVal := "https://example.com/archive.zip"
	source, _, err := ParseInstallURL("agenthub://install?source=url&repo=" + urlVal)
	if err != nil {
		t.Fatalf("ParseInstallURL(url) error: %v", err)
	}
	if source.Type != "zip" {
		t.Errorf("Type = %q, want 'zip' (url maps to zip)", source.Type)
	}
	if source.URL != urlVal {
		t.Errorf("URL = %q, want %q", source.URL, urlVal)
	}
}

func TestParseInstallURL_MissingRepo(t *testing.T) {
	_, _, err := ParseInstallURL("agenthub://install?source=github")
	if err == nil {
		t.Error("ParseInstallURL missing repo: expected error, got nil")
	}
	if !strings.Contains(err.Error(), "missing 'repo'") {
		t.Errorf("error = %q, want it to mention 'missing repo'", err.Error())
	}
}

func TestParseInstallURL_UnsupportedSourceType(t *testing.T) {
	_, _, err := ParseInstallURL("agenthub://install?source=bitbucket&repo=user/repo")
	if err == nil {
		t.Error("ParseInstallURL unsupported source: expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported source type") {
		t.Errorf("error = %q, want it to mention 'unsupported source type'", err.Error())
	}
}

func TestParseInstallURL_RefParameter(t *testing.T) {
	source, _, err := ParseInstallURL("agenthub://install?repo=user/repo&ref=v2.0.0")
	if err != nil {
		t.Fatalf("ParseInstallURL(ref) error: %v", err)
	}
	if source.Ref != "v2.0.0" {
		t.Errorf("Ref = %q, want 'v2.0.0'", source.Ref)
	}
}

func TestParseInstallURL_WrongScheme(t *testing.T) {
	_, _, err := ParseInstallURL("https://install?repo=user/repo")
	if err == nil {
		t.Error("ParseInstallURL wrong scheme: expected error, got nil")
	}
}

func TestParseInstallURL_InvalidRepoFormat_GitLab(t *testing.T) {
	_, _, err := ParseInstallURL("agenthub://install?source=gitlab&repo=noslash")
	if err == nil {
		t.Error("ParseInstallURL gitlab invalid repo: expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid repo format") {
		t.Errorf("error = %q, want it to mention 'invalid repo format'", err.Error())
	}
}

// ============================================================================
// IsOnboardingCompleted / CompleteOnboarding (onboarding.go:41,50)
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
// DetectSkillConflicts (onboarding.go:107) - pure function
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
// DeleteExistingSkills (onboarding.go:199)
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
// InstallDemoSkill (onboarding.go:211)
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
// moveDir (onboarding.go:267)
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
// expandPath (agents.go:312)
// ============================================================================

func TestExpandPath_Tilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory")
	}
	result := expandPath("~/foo")
	expected := filepath.Join(home, "foo")
	if result != expected {
		t.Errorf("expandPath('~/foo') = %q, want %q", result, expected)
	}
}

func TestExpandPath_AbsolutePath(t *testing.T) {
	result := expandPath("/absolute/path")
	if result != "/absolute/path" {
		t.Errorf("expandPath('/absolute/path') = %q, want '/absolute/path'", result)
	}
}

func TestExpandPath_RelativePath(t *testing.T) {
	result := expandPath("relative/path")
	if result != "relative/path" {
		t.Errorf("expandPath('relative/path') = %q, want 'relative/path'", result)
	}
}

func TestExpandPath_TildeOnly(t *testing.T) {
	// "~" without "/" prefix should not be expanded
	result := expandPath("~")
	if result != "~" {
		t.Errorf("expandPath('~') = %q, want '~' (no expansion for bare tilde)", result)
	}
}

func TestExpandPath_EmptyString(t *testing.T) {
	result := expandPath("")
	if result != "" {
		t.Errorf("expandPath('') = %q, want ''", result)
	}
}

// ============================================================================
// dirExists (agents.go:324)
// ============================================================================

func TestDirExists_ExistingDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	if !dirExists(tmpDir) {
		t.Errorf("dirExists(%q) = false, want true", tmpDir)
	}
}

func TestDirExists_NonExisting(t *testing.T) {
	if dirExists("/tmp/nonexistent-scribe-test-dir-999999") {
		t.Error("dirExists(non-existent) = true, want false")
	}
}

func TestDirExists_File(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "scribe-test-file-*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	_ = tmpFile.Close()
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if dirExists(tmpFile.Name()) {
		t.Errorf("dirExists(file) = true, want false (not a directory)")
	}
}

// ============================================================================
// countSkillsInDir (agents.go:334)
// ============================================================================

func TestCountSkillsInDir_WithSkills(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create 3 skill subdirectories with SKILL.md
	for _, name := range []string{"skill-a", "skill-b", "skill-c"} {
		skillDir := filepath.Join(tmpDir, name)
		_ = os.MkdirAll(skillDir, 0o755)
		_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# "+name), 0o644)
	}

	// Create a directory without SKILL.md (should not count)
	_ = os.MkdirAll(filepath.Join(tmpDir, "not-a-skill"), 0o755)

	// Create a file (should not count)
	_ = os.WriteFile(filepath.Join(tmpDir, "random.txt"), []byte("nope"), 0o644)

	count := countSkillsInDir(tmpDir)
	if count != 3 {
		t.Errorf("countSkillsInDir() = %d, want 3", count)
	}
}

func TestCountSkillsInDir_Empty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	count := countSkillsInDir(tmpDir)
	if count != 0 {
		t.Errorf("countSkillsInDir(empty) = %d, want 0", count)
	}
}

func TestCountSkillsInDir_NonExistent(t *testing.T) {
	count := countSkillsInDir("/tmp/nonexistent-scribe-test-dir-888888")
	if count != 0 {
		t.Errorf("countSkillsInDir(non-existent) = %d, want 0", count)
	}
}

// ============================================================================
// GetAgent / GetAllAgents (agents.go:259,264)
// ============================================================================

func TestGetAgent_Known(t *testing.T) {
	agent := GetAgent("claude-code")
	if agent == nil {
		t.Fatal("GetAgent('claude-code') = nil, want non-nil")
	}
	if agent.DisplayName != "Claude Code" {
		t.Errorf("DisplayName = %q, want 'Claude Code'", agent.DisplayName)
	}
	if agent.GlobalSkillsDir == "" {
		t.Error("GlobalSkillsDir is empty")
	}
}

func TestGetAgent_Unknown(t *testing.T) {
	agent := GetAgent("nonexistent-agent-xyz")
	if agent != nil {
		t.Errorf("GetAgent('nonexistent') = %v, want nil", agent)
	}
}

func TestGetAgent_EmptyString(t *testing.T) {
	agent := GetAgent("")
	if agent != nil {
		t.Errorf("GetAgent('') = %v, want nil", agent)
	}
}

func TestGetAllAgents_ReturnsCopy(t *testing.T) {
	agents1 := GetAllAgents()
	agents2 := GetAllAgents()

	if len(agents1) != len(agents2) {
		t.Fatalf("GetAllAgents() returned different lengths: %d vs %d", len(agents1), len(agents2))
	}

	// Mutating the returned slice should not affect the global
	if len(agents1) > 0 {
		agents1[0].DisplayName = "MUTATED"
		original := GetAgent(agents1[0].ID)
		if original != nil && original.DisplayName == "MUTATED" {
			t.Error("GetAllAgents() did not return a copy; mutation leaked to global")
		}
	}
}

func TestGetAllAgents_NonEmpty(t *testing.T) {
	agents := GetAllAgents()
	if len(agents) == 0 {
		t.Error("GetAllAgents() returned empty slice")
	}
}

// ============================================================================
// ComputeContentHash (meta.go:37)
// ============================================================================

func TestComputeContentHash_Deterministic(t *testing.T) {
	content := "some skill content here"
	h1 := ComputeContentHash(content)
	h2 := ComputeContentHash(content)
	if h1 != h2 {
		t.Errorf("ComputeContentHash not deterministic: %q != %q", h1, h2)
	}
}

func TestComputeContentHash_Prefix(t *testing.T) {
	hash := ComputeContentHash("anything")
	if !strings.HasPrefix(hash, "sha256:") {
		t.Errorf("ComputeContentHash prefix: got %q, want 'sha256:' prefix", hash)
	}
}

func TestComputeContentHash_Length(t *testing.T) {
	hash := ComputeContentHash("test")
	// "sha256:" (7 chars) + 64 hex chars = 71
	if len(hash) != 71 {
		t.Errorf("ComputeContentHash length = %d, want 71", len(hash))
	}
}

func TestComputeContentHash_Different(t *testing.T) {
	h1 := ComputeContentHash("content A")
	h2 := ComputeContentHash("content B")
	if h1 == h2 {
		t.Error("ComputeContentHash: different content produced same hash")
	}
}

func TestComputeContentHash_Empty(t *testing.T) {
	hash := ComputeContentHash("")
	if !strings.HasPrefix(hash, "sha256:") {
		t.Error("ComputeContentHash('') should still produce a valid hash")
	}
	if len(hash) != 71 {
		t.Errorf("ComputeContentHash('') length = %d, want 71", len(hash))
	}
}

// ============================================================================
// NewSkillMeta / UpdateSkillMeta (meta.go:43,90)
// ============================================================================

func TestNewSkillMeta_GitHub(t *testing.T) {
	source := &SourceInfo{
		Type:  "github",
		Owner: "octocat",
		Repo:  "skills",
		URL:   "https://github.com/octocat/skills",
	}
	meta := NewSkillMeta(source, "skills/test", "content body")

	if meta.SourceType != "github" {
		t.Errorf("SourceType = %q, want 'github'", meta.SourceType)
	}
	if meta.Source != "octocat/skills" {
		t.Errorf("Source = %q, want 'octocat/skills'", meta.Source)
	}
	if meta.SourceURL != "https://github.com/octocat/skills" {
		t.Errorf("SourceURL = %q, want URL", meta.SourceURL)
	}
	if meta.SkillPath != "skills/test" {
		t.Errorf("SkillPath = %q, want 'skills/test'", meta.SkillPath)
	}
	if meta.ContentHash == "" {
		t.Error("ContentHash is empty")
	}
	if meta.InstalledAt == "" {
		t.Error("InstalledAt is empty")
	}
	if meta.UpdatedAt == "" {
		t.Error("UpdatedAt is empty")
	}
	if meta.InstalledAt != meta.UpdatedAt {
		t.Error("InstalledAt and UpdatedAt should be equal on creation")
	}
}

func TestNewSkillMeta_NoURL(t *testing.T) {
	source := &SourceInfo{
		Type:      "local",
		LocalPath: "/path/to/skill",
	}
	meta := NewSkillMeta(source, "", "local content")

	if meta.SourceURL != "" {
		t.Errorf("SourceURL = %q, want empty for local source with no URL", meta.SourceURL)
	}
	if meta.SkillPath != "" {
		t.Errorf("SkillPath = %q, want empty", meta.SkillPath)
	}
}

func TestNewSkillMeta_WithRef(t *testing.T) {
	source := &SourceInfo{
		Type:  "github",
		Owner: "user",
		Repo:  "repo",
		Ref:   "v2.0.0",
		URL:   "https://github.com/user/repo",
	}
	meta := NewSkillMeta(source, "", "content")
	// formatSource (private) includes ref only for non main/master
	if !strings.Contains(meta.Source, "v2.0.0") {
		t.Errorf("Source = %q, expected it to contain ref 'v2.0.0'", meta.Source)
	}
}

func TestUpdateSkillMeta_ContentChange(t *testing.T) {
	source := &SourceInfo{Type: "github", Owner: "u", Repo: "r"}
	meta := NewSkillMeta(source, "", "old content")
	oldHash := meta.ContentHash
	oldUpdated := meta.UpdatedAt

	UpdateSkillMeta(meta, "new content")

	if meta.ContentHash == oldHash {
		t.Error("UpdateSkillMeta did not change ContentHash")
	}
	if meta.ContentHash != ComputeContentHash("new content") {
		t.Errorf("UpdateSkillMeta ContentHash = %q, want hash of 'new content'", meta.ContentHash)
	}
	// UpdatedAt should be >= old (could be equal if test runs fast)
	if meta.UpdatedAt < oldUpdated {
		t.Errorf("UpdateSkillMeta: UpdatedAt went backwards: %q < %q", meta.UpdatedAt, oldUpdated)
	}
	// InstalledAt should not change
	if meta.InstalledAt == "" {
		t.Error("UpdateSkillMeta: InstalledAt was cleared")
	}
}

// ============================================================================
// SkillNeedsUpdate (meta.go:171)
// ============================================================================

func TestSkillNeedsUpdate_MatchingHash(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	content := "test skill content"
	meta := &SkillMeta{
		ContentHash: ComputeContentHash(content),
		InstalledAt: "2025-01-01T00:00:00Z",
		UpdatedAt:   "2025-01-01T00:00:00Z",
	}
	metaPath := filepath.Join(tmpDir, MetaFileName)
	_ = WriteSkillMeta(metaPath, meta)

	needsUpdate, err := SkillNeedsUpdate(tmpDir, content)
	if err != nil {
		t.Fatalf("SkillNeedsUpdate() error: %v", err)
	}
	if needsUpdate {
		t.Error("SkillNeedsUpdate() = true for matching hash, want false")
	}
}

func TestSkillNeedsUpdate_DifferentHash(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	meta := &SkillMeta{
		ContentHash: ComputeContentHash("old content"),
		InstalledAt: "2025-01-01T00:00:00Z",
		UpdatedAt:   "2025-01-01T00:00:00Z",
	}
	metaPath := filepath.Join(tmpDir, MetaFileName)
	_ = WriteSkillMeta(metaPath, meta)

	needsUpdate, err := SkillNeedsUpdate(tmpDir, "new content")
	if err != nil {
		t.Fatalf("SkillNeedsUpdate() error: %v", err)
	}
	if !needsUpdate {
		t.Error("SkillNeedsUpdate() = false for different hash, want true")
	}
}

func TestSkillNeedsUpdate_MissingMeta(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// No meta file at all
	needsUpdate, err := SkillNeedsUpdate(tmpDir, "any content")
	if err != nil {
		t.Fatalf("SkillNeedsUpdate() error: %v", err)
	}
	if !needsUpdate {
		t.Error("SkillNeedsUpdate() = false for missing meta, want true")
	}
}

// ============================================================================
// SaveSkillWithMeta (meta.go:142)
// ============================================================================

func TestSaveSkillWithMeta_FullRoundtrip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a source skill directory with SKILL.md
	srcDir := filepath.Join(tmpDir, "src-skill")
	_ = os.MkdirAll(srcDir, 0o755)
	skillContent := "---\nname: test-save\ndescription: Test saving\n---\n# Content\n"
	_ = os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte(skillContent), 0o644)

	skill := &Skill{
		Name:        "test-save",
		Description: "Test saving",
		Path:        srcDir,
		Content:     "# Content",
	}
	source := &SourceInfo{
		Type:  "github",
		Owner: "testuser",
		Repo:  "testrepo",
		URL:   "https://github.com/testuser/testrepo",
	}

	destDir := filepath.Join(tmpDir, "dest-skill")
	err = SaveSkillWithMeta(destDir, skill, source, "skills/test-save")
	if err != nil {
		t.Fatalf("SaveSkillWithMeta() error: %v", err)
	}

	// Verify SKILL.md exists
	if _, err := os.Stat(filepath.Join(destDir, "SKILL.md")); os.IsNotExist(err) {
		t.Error("SaveSkillWithMeta: SKILL.md not created")
	}

	// Verify meta exists
	if _, err := os.Stat(filepath.Join(destDir, ".scribe-meta.json")); os.IsNotExist(err) {
		t.Error("SaveSkillWithMeta: .scribe-meta.json not created")
	}

	// Verify meta contents
	meta, err := ReadSkillMeta(filepath.Join(destDir, ".scribe-meta.json"))
	if err != nil {
		t.Fatalf("ReadSkillMeta() error: %v", err)
	}
	if meta.SourceType != "github" {
		t.Errorf("meta.SourceType = %q, want 'github'", meta.SourceType)
	}
	if meta.SkillPath != "skills/test-save" {
		t.Errorf("meta.SkillPath = %q, want 'skills/test-save'", meta.SkillPath)
	}
}

// ============================================================================
// ValidateSkill / SanitizeName (skills.go:88,102)
// ============================================================================

func TestValidateSkill_Valid(t *testing.T) {
	skill := &Skill{Name: "my-skill", Description: "A valid skill"}
	if err := ValidateSkill(skill); err != nil {
		t.Errorf("ValidateSkill(valid) error: %v", err)
	}
}

func TestValidateSkill_Nil(t *testing.T) {
	if err := ValidateSkill(nil); err != ErrInvalidSkill {
		t.Errorf("ValidateSkill(nil) = %v, want ErrInvalidSkill", err)
	}
}

func TestValidateSkill_MissingName(t *testing.T) {
	skill := &Skill{Description: "Has description but no name"}
	if err := ValidateSkill(skill); err != ErrMissingName {
		t.Errorf("ValidateSkill(no name) = %v, want ErrMissingName", err)
	}
}

func TestValidateSkill_MissingDescription(t *testing.T) {
	skill := &Skill{Name: "has-name"}
	if err := ValidateSkill(skill); err != ErrMissingDesc {
		t.Errorf("ValidateSkill(no desc) = %v, want ErrMissingDesc", err)
	}
}

func TestSanitizeName_Simple(t *testing.T) {
	result := SanitizeName("Hello World")
	if result != "hello-world" {
		t.Errorf("SanitizeName('Hello World') = %q, want 'hello-world'", result)
	}
}

func TestSanitizeName_Underscores(t *testing.T) {
	result := SanitizeName("my_cool_skill")
	if result != "my-cool-skill" {
		t.Errorf("SanitizeName('my_cool_skill') = %q, want 'my-cool-skill'", result)
	}
}

func TestSanitizeName_SpecialChars(t *testing.T) {
	result := SanitizeName("skill@#$%^&*()")
	if result != "skill" {
		t.Errorf("SanitizeName('skill@#$%%^&*()') = %q, want 'skill'", result)
	}
}

func TestSanitizeName_ConsecutiveHyphens(t *testing.T) {
	result := SanitizeName("a - - b")
	if result != "a-b" {
		t.Errorf("SanitizeName('a - - b') = %q, want 'a-b'", result)
	}
}

func TestSanitizeName_LeadingTrailingHyphens(t *testing.T) {
	result := SanitizeName("-leading-trailing-")
	if result != "leading-trailing" {
		t.Errorf("SanitizeName('-leading-trailing-') = %q, want 'leading-trailing'", result)
	}
}

func TestSanitizeName_AllInvalid(t *testing.T) {
	result := SanitizeName("@#$%")
	if result != "" {
		t.Errorf("SanitizeName('@#$%%') = %q, want ''", result)
	}
}

func TestSanitizeName_AlreadyValid(t *testing.T) {
	result := SanitizeName("already-valid-123")
	if result != "already-valid-123" {
		t.Errorf("SanitizeName('already-valid-123') = %q, want 'already-valid-123'", result)
	}
}

// ============================================================================
// GetSkillInfo (skills.go:302)
// ============================================================================

func TestGetSkillInfo_WithoutMeta(t *testing.T) {
	skill := &Skill{
		Name:        "test-skill",
		Description: "A test skill",
	}
	info := GetSkillInfo(skill)
	if info.Name != "test-skill" {
		t.Errorf("info.Name = %q, want 'test-skill'", info.Name)
	}
	if info.Description != "A test skill" {
		t.Errorf("info.Description = %q, want 'A test skill'", info.Description)
	}
	if info.Source != "" {
		t.Errorf("info.Source = %q, want '' (no meta)", info.Source)
	}
	if info.SourceType != "" {
		t.Errorf("info.SourceType = %q, want '' (no meta)", info.SourceType)
	}
	if info.InstalledAt != "" {
		t.Errorf("info.InstalledAt = %q, want '' (no meta)", info.InstalledAt)
	}
	if info.Agents == nil {
		t.Error("info.Agents is nil, want empty slice")
	}
}

func TestGetSkillInfo_WithMeta(t *testing.T) {
	skill := &Skill{
		Name:        "meta-skill",
		Description: "Skill with metadata",
		Meta: &SkillMeta{
			Source:      "octocat/skills",
			SourceType:  "github",
			InstalledAt: "2025-01-15T10:00:00Z",
		},
	}
	info := GetSkillInfo(skill)
	if info.Source != "octocat/skills" {
		t.Errorf("info.Source = %q, want 'octocat/skills'", info.Source)
	}
	if info.SourceType != "github" {
		t.Errorf("info.SourceType = %q, want 'github'", info.SourceType)
	}
	if info.InstalledAt != "2025-01-15T10:00:00Z" {
		t.Errorf("info.InstalledAt = %q, want '2025-01-15T10:00:00Z'", info.InstalledAt)
	}
}

// ============================================================================
// LoadSkillWithMeta / ListSkillsWithMeta (meta.go:96,114)
// ============================================================================

func TestLoadSkillWithMeta_WithMeta(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	skillDir := filepath.Join(tmpDir, "my-skill")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: my-skill\ndescription: Test\n---\n# Body\n"), 0o644)
	meta := &SkillMeta{
		Source:      "local",
		SourceType:  "local",
		ContentHash: ComputeContentHash("test"),
		InstalledAt: "2025-01-01T00:00:00Z",
		UpdatedAt:   "2025-01-01T00:00:00Z",
	}
	_ = WriteSkillMeta(filepath.Join(skillDir, ".scribe-meta.json"), meta)

	skill, err := LoadSkillWithMeta(skillDir)
	if err != nil {
		t.Fatalf("LoadSkillWithMeta() error: %v", err)
	}
	if skill.Name != "my-skill" {
		t.Errorf("skill.Name = %q, want 'my-skill'", skill.Name)
	}
	if skill.Meta == nil {
		t.Fatal("skill.Meta is nil, expected metadata")
	}
	if skill.Meta.SourceType != "local" {
		t.Errorf("skill.Meta.SourceType = %q, want 'local'", skill.Meta.SourceType)
	}
}

func TestLoadSkillWithMeta_WithoutMeta(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	skillDir := filepath.Join(tmpDir, "bare-skill")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: bare-skill\ndescription: No meta\n---\n# Content\n"), 0o644)

	skill, err := LoadSkillWithMeta(skillDir)
	if err != nil {
		t.Fatalf("LoadSkillWithMeta() error: %v", err)
	}
	if skill.Name != "bare-skill" {
		t.Errorf("skill.Name = %q, want 'bare-skill'", skill.Name)
	}
	if skill.Meta != nil {
		t.Errorf("skill.Meta = %v, want nil (no meta file)", skill.Meta)
	}
}

func TestLoadSkillWithMeta_InvalidSkill(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	skillDir := filepath.Join(tmpDir, "bad-skill")
	_ = os.MkdirAll(skillDir, 0o755)
	// No SKILL.md at all
	_, err = LoadSkillWithMeta(skillDir)
	if err == nil {
		t.Error("LoadSkillWithMeta(no SKILL.md) expected error, got nil")
	}
}

func TestListSkillsWithMeta_Mixed(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	scrollsDir := filepath.Join(tmpDir, "scrolls")
	_ = os.MkdirAll(scrollsDir, 0o755)

	// Valid skill with meta
	s1 := filepath.Join(scrollsDir, "skill-one")
	_ = os.MkdirAll(s1, 0o755)
	_ = os.WriteFile(filepath.Join(s1, "SKILL.md"), []byte("---\nname: skill-one\ndescription: First\n---\n# One\n"), 0o644)
	_ = WriteSkillMeta(filepath.Join(s1, ".scribe-meta.json"), &SkillMeta{
		Source: "test", SourceType: "local", ContentHash: "sha256:abc", InstalledAt: "2025-01-01T00:00:00Z", UpdatedAt: "2025-01-01T00:00:00Z",
	})

	// Valid skill without meta
	s2 := filepath.Join(scrollsDir, "skill-two")
	_ = os.MkdirAll(s2, 0o755)
	_ = os.WriteFile(filepath.Join(s2, "SKILL.md"), []byte("---\nname: skill-two\ndescription: Second\n---\n# Two\n"), 0o644)

	// Invalid skill (no frontmatter) - should be skipped
	s3 := filepath.Join(scrollsDir, "bad-skill")
	_ = os.MkdirAll(s3, 0o755)
	_ = os.WriteFile(filepath.Join(s3, "SKILL.md"), []byte("No frontmatter"), 0o644)

	// A regular file (not a dir) - should be skipped
	_ = os.WriteFile(filepath.Join(scrollsDir, "not-a-dir.txt"), []byte("ignored"), 0o644)

	skills, err := ListSkillsWithMeta(scrollsDir)
	if err != nil {
		t.Fatalf("ListSkillsWithMeta() error: %v", err)
	}
	if len(skills) != 2 {
		t.Errorf("ListSkillsWithMeta() returned %d skills, want 2", len(skills))
	}
}

func TestListSkillsWithMeta_NonExistentDir(t *testing.T) {
	skills, err := ListSkillsWithMeta("/tmp/nonexistent-scribe-scrolls-dir-777")
	if err != nil {
		t.Fatalf("ListSkillsWithMeta(nonexistent) error: %v", err)
	}
	if len(skills) != 0 {
		t.Errorf("ListSkillsWithMeta(nonexistent) returned %d skills, want 0", len(skills))
	}
}

func TestListSkillsWithMeta_EmptyDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	skills, err := ListSkillsWithMeta(tmpDir)
	if err != nil {
		t.Fatalf("ListSkillsWithMeta(empty) error: %v", err)
	}
	if len(skills) != 0 {
		t.Errorf("ListSkillsWithMeta(empty) returned %d skills, want 0", len(skills))
	}
}

// ============================================================================
// GetAgentStatus (agents.go:283)
// ============================================================================

func TestGetAgentStatus_EmptyHome(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	InitLoggerCLI(false)

	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	statuses := GetAgentStatus(scrollsDir)

	if len(statuses) != len(AllAgents) {
		t.Errorf("GetAgentStatus() returned %d statuses, want %d", len(statuses), len(AllAgents))
	}

	// With empty HOME, no agents should be installed
	for _, s := range statuses {
		if s.Installed {
			t.Errorf("agent %q reported as installed in empty HOME", s.ID)
		}
		if s.SkillCount != 0 {
			t.Errorf("agent %q has %d skills in empty HOME, want 0", s.ID, s.SkillCount)
		}
	}
}

func TestGetAgentStatus_WithAgent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	InitLoggerCLI(false)

	// Create claude-code config and skills directories
	_ = os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0o755)
	skillsDir := filepath.Join(tmpDir, ".claude", "skills")
	_ = os.MkdirAll(filepath.Join(skillsDir, "test-skill"), 0o755)
	_ = os.WriteFile(filepath.Join(skillsDir, "test-skill", "SKILL.md"), []byte("# Test"), 0o644)

	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	statuses := GetAgentStatus(scrollsDir)

	// Find claude-code status
	var claudeStatus *AgentStatus
	for i := range statuses {
		if statuses[i].ID == "claude-code" {
			claudeStatus = &statuses[i]
			break
		}
	}

	if claudeStatus == nil {
		t.Fatal("claude-code not found in statuses")
	}
	if !claudeStatus.Installed {
		t.Error("claude-code should be installed")
	}
	if claudeStatus.SkillCount != 1 {
		t.Errorf("claude-code SkillCount = %d, want 1", claudeStatus.SkillCount)
	}
}

// ============================================================================
// ExpandAgentPath (agents.go:307) - public wrapper
// ============================================================================

func TestExpandAgentPath_Tilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir")
	}
	result := ExpandAgentPath("~/test/path")
	expected := filepath.Join(home, "test", "path")
	if result != expected {
		t.Errorf("ExpandAgentPath('~/test/path') = %q, want %q", result, expected)
	}
}

func TestExpandAgentPath_Absolute(t *testing.T) {
	result := ExpandAgentPath("/usr/local/bin")
	if result != "/usr/local/bin" {
		t.Errorf("ExpandAgentPath('/usr/local/bin') = %q, want '/usr/local/bin'", result)
	}
}

func TestExpandAgentPath_Relative(t *testing.T) {
	result := ExpandAgentPath("some/relative")
	if result != "some/relative" {
		t.Errorf("ExpandAgentPath('some/relative') = %q, want 'some/relative'", result)
	}
}

// ============================================================================
// ReadSkillMeta / WriteSkillMeta roundtrip (meta.go:12,27)
// ============================================================================

func TestReadWriteSkillMeta_Roundtrip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	metaPath := filepath.Join(tmpDir, ".scribe-meta.json")
	original := &SkillMeta{
		Source:      "myorg/myrepo",
		SourceType:  "github",
		SourceURL:   "https://github.com/myorg/myrepo",
		SkillPath:   "skills/foo",
		ContentHash: "sha256:abcdef1234567890",
		InstalledAt: "2025-06-01T12:00:00Z",
		UpdatedAt:   "2025-06-15T14:30:00Z",
	}

	if err := WriteSkillMeta(metaPath, original); err != nil {
		t.Fatalf("WriteSkillMeta() error: %v", err)
	}

	loaded, err := ReadSkillMeta(metaPath)
	if err != nil {
		t.Fatalf("ReadSkillMeta() error: %v", err)
	}

	if loaded.Source != original.Source {
		t.Errorf("Source = %q, want %q", loaded.Source, original.Source)
	}
	if loaded.SourceType != original.SourceType {
		t.Errorf("SourceType = %q, want %q", loaded.SourceType, original.SourceType)
	}
	if loaded.SourceURL != original.SourceURL {
		t.Errorf("SourceURL = %q, want %q", loaded.SourceURL, original.SourceURL)
	}
	if loaded.SkillPath != original.SkillPath {
		t.Errorf("SkillPath = %q, want %q", loaded.SkillPath, original.SkillPath)
	}
	if loaded.ContentHash != original.ContentHash {
		t.Errorf("ContentHash = %q, want %q", loaded.ContentHash, original.ContentHash)
	}
	if loaded.InstalledAt != original.InstalledAt {
		t.Errorf("InstalledAt = %q, want %q", loaded.InstalledAt, original.InstalledAt)
	}
	if loaded.UpdatedAt != original.UpdatedAt {
		t.Errorf("UpdatedAt = %q, want %q", loaded.UpdatedAt, original.UpdatedAt)
	}
}

func TestReadSkillMeta_NonExistent(t *testing.T) {
	_, err := ReadSkillMeta("/tmp/nonexistent-scribe-meta-file.json")
	if err == nil {
		t.Error("ReadSkillMeta(nonexistent) expected error, got nil")
	}
}

// ============================================================================
// formatSource (meta.go:66) - private function
// ============================================================================

func TestFormatSourcePrivate_GitHubMainRef(t *testing.T) {
	source := &SourceInfo{Type: "github", Owner: "user", Repo: "repo", Ref: "main"}
	result := formatSource(source)
	// main/master refs are excluded by the private formatSource
	if result != "user/repo" {
		t.Errorf("formatSource(github, ref=main) = %q, want 'user/repo'", result)
	}
}

func TestFormatSourcePrivate_GitHubMasterRef(t *testing.T) {
	source := &SourceInfo{Type: "github", Owner: "user", Repo: "repo", Ref: "master"}
	result := formatSource(source)
	if result != "user/repo" {
		t.Errorf("formatSource(github, ref=master) = %q, want 'user/repo'", result)
	}
}

func TestFormatSourcePrivate_GitHubCustomRef(t *testing.T) {
	source := &SourceInfo{Type: "github", Owner: "user", Repo: "repo", Ref: "v1.0"}
	result := formatSource(source)
	if result != "user/repo#v1.0" {
		t.Errorf("formatSource(github, ref=v1.0) = %q, want 'user/repo#v1.0'", result)
	}
}

func TestFormatSourcePrivate_GitLab(t *testing.T) {
	source := &SourceInfo{Type: "gitlab", Owner: "group", Repo: "proj", Ref: "develop"}
	result := formatSource(source)
	if result != "group/proj#develop" {
		t.Errorf("formatSource(gitlab, ref=develop) = %q, want 'group/proj#develop'", result)
	}
}

func TestFormatSourcePrivate_Local(t *testing.T) {
	source := &SourceInfo{Type: "local", LocalPath: "/my/local/path"}
	result := formatSource(source)
	if result != "/my/local/path" {
		t.Errorf("formatSource(local) = %q, want '/my/local/path'", result)
	}
}

func TestFormatSourcePrivate_URL(t *testing.T) {
	source := &SourceInfo{Type: "url", URL: "https://example.com/archive.zip"}
	result := formatSource(source)
	if result != "https://example.com/archive.zip" {
		t.Errorf("formatSource(url) = %q, want URL", result)
	}
}

func TestFormatSourcePrivate_WellKnown(t *testing.T) {
	source := &SourceInfo{Type: "well-known", URL: "https://example.com/.well-known/skills"}
	result := formatSource(source)
	if result != "https://example.com/.well-known/skills" {
		t.Errorf("formatSource(well-known) = %q, want URL", result)
	}
}

func TestFormatSourcePrivate_Unknown(t *testing.T) {
	source := &SourceInfo{Type: "custom", URL: "https://other.com"}
	result := formatSource(source)
	if result != "https://other.com" {
		t.Errorf("formatSource(custom) = %q, want URL", result)
	}
}
