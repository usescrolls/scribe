package scribe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ============================================================================
// ValidateSkill (skills.go)
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

// ============================================================================
// SanitizeName (skills.go)
// ============================================================================

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

func TestBoost_SanitizeName_Long(t *testing.T) {
	long := strings.Repeat("a", 300)
	result := SanitizeName(long)
	if len(result) > 255 {
		t.Errorf("SanitizeName long string: len = %d, want <= 255", len(result))
	}
}

func TestBoost_SanitizeName_Empty(t *testing.T) {
	result := SanitizeName("")
	if result != "" {
		t.Errorf("SanitizeName('') = %q, want ''", result)
	}
}

// ============================================================================
// FilterSkillsByName (skills.go)
// ============================================================================

func TestFilterSkillsByName_MatchesSubset(t *testing.T) {
	skills := []*Skill{
		{Name: "alpha", Description: "A"},
		{Name: "beta", Description: "B"},
		{Name: "gamma", Description: "G"},
	}
	result := FilterSkillsByName(skills, []string{"alpha", "gamma"})
	if len(result) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(result))
	}
	if result[0].Name != "alpha" || result[1].Name != "gamma" {
		t.Errorf("got [%s, %s], want [alpha, gamma]", result[0].Name, result[1].Name)
	}
}

func TestFilterSkillsByName_NoMatches(t *testing.T) {
	skills := []*Skill{
		{Name: "alpha", Description: "A"},
		{Name: "beta", Description: "B"},
	}
	result := FilterSkillsByName(skills, []string{"nonexistent"})
	if len(result) != 0 {
		t.Errorf("expected 0 skills, got %d", len(result))
	}
}

func TestFilterSkillsByName_EmptyNames(t *testing.T) {
	skills := []*Skill{
		{Name: "alpha", Description: "A"},
	}
	result := FilterSkillsByName(skills, []string{})
	if len(result) != 0 {
		t.Errorf("expected 0 skills, got %d", len(result))
	}
}

func TestFilterSkillsByName_EmptySkills(t *testing.T) {
	result := FilterSkillsByName([]*Skill{}, []string{"alpha"})
	if len(result) != 0 {
		t.Errorf("expected 0 skills, got %d", len(result))
	}
}

func TestFilterSkillsByName_NilSkills(t *testing.T) {
	result := FilterSkillsByName(nil, []string{"alpha"})
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestFilterSkillsByName_TrimsWhitespace(t *testing.T) {
	skills := []*Skill{
		{Name: "alpha", Description: "A"},
		{Name: "beta", Description: "B"},
	}
	result := FilterSkillsByName(skills, []string{" alpha ", " beta "})
	if len(result) != 2 {
		t.Errorf("expected 2 skills with trimmed names, got %d", len(result))
	}
}

func TestFilterSkillsByName_AllMatch(t *testing.T) {
	skills := []*Skill{
		{Name: "alpha", Description: "A"},
		{Name: "beta", Description: "B"},
	}
	result := FilterSkillsByName(skills, []string{"alpha", "beta"})
	if len(result) != 2 {
		t.Errorf("expected 2 skills, got %d", len(result))
	}
}

func TestFilterSkillsByName_DuplicateNamesInFilter(t *testing.T) {
	skills := []*Skill{
		{Name: "alpha", Description: "A"},
	}
	result := FilterSkillsByName(skills, []string{"alpha", "alpha"})
	if len(result) != 1 {
		t.Errorf("expected 1 skill (no duplicates), got %d", len(result))
	}
}

func TestFilterSkillsByName_CaseInsensitive(t *testing.T) {
	skills := []*Skill{
		{Name: "alpha", Description: "A"},
		{Name: "beta", Description: "B"},
	}
	result := FilterSkillsByName(skills, []string{"Alpha", "BETA"})
	if len(result) != 2 {
		t.Errorf("expected 2 skills with case-insensitive match, got %d", len(result))
	}
}

func TestSkillInList_CaseInsensitive(t *testing.T) {
	skills := []*Skill{
		{Name: "my-skill", Description: "A"},
	}
	if !skillInList(skills, "My-Skill") {
		t.Error("skillInList should match case-insensitively")
	}
	if !skillInList(skills, "MY-SKILL") {
		t.Error("skillInList should match case-insensitively")
	}
}

func TestParseSkillContent_NormalizesName(t *testing.T) {
	content := "---\nname: My Cool Skill\ndescription: Test\n---\n# Body\n"
	skill, err := ParseSkillContent(content, "/tmp")
	if err != nil {
		t.Fatalf("ParseSkillContent() error: %v", err)
	}
	if skill.Name != "my-cool-skill" {
		t.Errorf("name = %q, want 'my-cool-skill' (normalized)", skill.Name)
	}
}

func TestParseSkillContent_NormalizesUppercase(t *testing.T) {
	content := "---\nname: MySkill\ndescription: Test\n---\n# Body\n"
	skill, err := ParseSkillContent(content, "/tmp")
	if err != nil {
		t.Fatalf("ParseSkillContent() error: %v", err)
	}
	if skill.Name != "myskill" {
		t.Errorf("name = %q, want 'myskill' (normalized)", skill.Name)
	}
}

// ============================================================================
// GetSkillInfo (skills.go)
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
	if info.SourceURL != "" {
		t.Errorf("info.SourceURL = %q, want '' (no meta)", info.SourceURL)
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
			SourceURL:   "https://github.com/octocat/skills",
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
	if info.SourceURL != "https://github.com/octocat/skills" {
		t.Errorf("info.SourceURL = %q, want 'https://github.com/octocat/skills'", info.SourceURL)
	}
	if info.InstalledAt != "2025-01-15T10:00:00Z" {
		t.Errorf("info.InstalledAt = %q, want '2025-01-15T10:00:00Z'", info.InstalledAt)
	}
}

func TestGetSkillInfo_MapsVersionFields(t *testing.T) {
	skill := &Skill{
		Name:        "version-skill",
		Description: "Skill with version fields",
		Meta: &SkillMeta{
			Source:      "octocat/skills",
			SourceType:  "github",
			SourceURL:   "https://github.com/octocat/skills",
			ContentHash: "sha256:abcdef1234567890",
			CommitHash:  "abc1234",
			CommitDate:  "2025-06-15T10:30:00Z",
			InstalledAt: "2025-06-01T12:00:00Z",
			UpdatedAt:   "2025-06-15T14:30:00Z",
		},
	}
	info := GetSkillInfo(skill)

	if info.UpdatedAt != "2025-06-15T14:30:00Z" {
		t.Errorf("info.UpdatedAt = %q, want '2025-06-15T14:30:00Z'", info.UpdatedAt)
	}
	if info.ContentHash != "sha256:abcdef1234567890" {
		t.Errorf("info.ContentHash = %q, want 'sha256:abcdef1234567890'", info.ContentHash)
	}
	if info.CommitHash != "abc1234" {
		t.Errorf("info.CommitHash = %q, want 'abc1234'", info.CommitHash)
	}
	if info.CommitDate != "2025-06-15T10:30:00Z" {
		t.Errorf("info.CommitDate = %q, want '2025-06-15T10:30:00Z'", info.CommitDate)
	}
}

func TestGetSkillInfo_NoMeta_VersionFieldsEmpty(t *testing.T) {
	skill := &Skill{
		Name:        "no-meta-skill",
		Description: "No metadata",
	}
	info := GetSkillInfo(skill)

	if info.UpdatedAt != "" {
		t.Errorf("info.UpdatedAt = %q, want empty", info.UpdatedAt)
	}
	if info.ContentHash != "" {
		t.Errorf("info.ContentHash = %q, want empty", info.ContentHash)
	}
	if info.CommitHash != "" {
		t.Errorf("info.CommitHash = %q, want empty", info.CommitHash)
	}
	if info.CommitDate != "" {
		t.Errorf("info.CommitDate = %q, want empty", info.CommitDate)
	}
}

func TestGetSkillInfo_MapsIsPrivate(t *testing.T) {
	skill := &Skill{
		Name:        "private-skill",
		Description: "Private repo skill",
		Meta: &SkillMeta{
			Source:     "octocat/private-skills",
			SourceType: "github",
			IsPrivate:  true,
		},
	}
	info := GetSkillInfo(skill)
	if !info.IsPrivate {
		t.Error("info.IsPrivate = false, want true")
	}
}

func TestGetSkillInfo_IsPrivateDefaultsFalse(t *testing.T) {
	skill := &Skill{
		Name:        "public-skill",
		Description: "Public repo skill",
		Meta: &SkillMeta{
			Source:     "octocat/public-skills",
			SourceType: "github",
		},
	}
	info := GetSkillInfo(skill)
	if info.IsPrivate {
		t.Error("info.IsPrivate = true, want false for public source")
	}
}

func TestGetSkillInfo_NoMeta_IsPrivateFalse(t *testing.T) {
	skill := &Skill{
		Name:        "no-meta-skill",
		Description: "No metadata",
	}
	info := GetSkillInfo(skill)
	if info.IsPrivate {
		t.Error("info.IsPrivate = true, want false when no meta")
	}
}

// ============================================================================
// GetAgentsWithSkill (skills.go)
// ============================================================================

func TestBoost_GetAgentsWithSkill_NoAgents(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)

	agents := GetAgentsWithSkill("some-skill")
	if len(agents) != 0 {
		t.Errorf("expected 0 agents with no installations, got %d", len(agents))
	}
}

func TestBoost_GetAgentsWithSkill_WithAgent(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)

	// Create claude-code config dir and a skill in its skills dir
	_ = os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0o755)
	skillDir := filepath.Join(tmpDir, ".claude", "skills", "detected-skill")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Skill"), 0o644)

	agents := GetAgentsWithSkill("detected-skill")
	if len(agents) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(agents))
	}
	if agents[0] != "claude-code" {
		t.Errorf("agent = %q, want 'claude-code'", agents[0])
	}
}

// ============================================================================
// skillInList (skills.go)
// ============================================================================

func TestBoost_SkillInList(t *testing.T) {
	skills := []*Skill{
		{Name: "a", Description: "A"},
		{Name: "b", Description: "B"},
	}

	if !skillInList(skills, "a") {
		t.Error("skillInList('a') = false, want true")
	}
	if !skillInList(skills, "b") {
		t.Error("skillInList('b') = false, want true")
	}
	if skillInList(skills, "c") {
		t.Error("skillInList('c') = true, want false")
	}
	if skillInList(nil, "a") {
		t.Error("skillInList(nil, 'a') = true, want false")
	}
}

// ============================================================================
// DiscoverSkills (skills.go)
// ============================================================================

func TestBoost_DiscoverSkillsWithDepth_SkipsDirs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-discover-*")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a skill in node_modules (should be skipped)
	nmDir := filepath.Join(tmpDir, "node_modules", "some-pkg")
	_ = os.MkdirAll(nmDir, 0o755)
	_ = os.WriteFile(filepath.Join(nmDir, "SKILL.md"), []byte("---\nname: hidden\ndescription: hidden\n---\n# Hidden\n"), 0o644)

	// Create a valid skill at root
	_ = os.WriteFile(filepath.Join(tmpDir, "SKILL.md"), []byte("---\nname: visible\ndescription: visible\n---\n# Visible\n"), 0o644)

	skills, err := DiscoverSkillsWithDepth(tmpDir, 5)
	if err != nil {
		t.Fatalf("DiscoverSkillsWithDepth() error: %v", err)
	}

	// Only the root skill should be found, not the one in node_modules
	for _, s := range skills {
		if s.Name == "hidden" {
			t.Error("skill in node_modules should have been skipped")
		}
	}

	found := false
	for _, s := range skills {
		if s.Name == "visible" {
			found = true
		}
	}
	if !found {
		t.Error("root skill not found")
	}
}

func TestBoost_DiscoverSkills_NoSkills(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-discover-*")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	_, err = DiscoverSkills(tmpDir)
	if err != ErrNoSkillsFound {
		t.Errorf("DiscoverSkills(empty) error = %v, want ErrNoSkillsFound", err)
	}
}

func TestBoost_DiscoverSkills_InCommonDirs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-discover-*")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Put skills in the "skills/" common dir
	skillsDir := filepath.Join(tmpDir, "skills", "my-skill")
	_ = os.MkdirAll(skillsDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), []byte("---\nname: common-skill\ndescription: Common\n---\n# Common\n"), 0o644)

	skills, err := DiscoverSkills(tmpDir)
	if err != nil {
		t.Fatalf("DiscoverSkills() error: %v", err)
	}
	if len(skills) < 1 {
		t.Fatal("expected at least 1 skill in common dir")
	}

	found := false
	for _, s := range skills {
		if s.Name == "common-skill" {
			found = true
		}
	}
	if !found {
		t.Error("skill in skills/ dir not discovered")
	}
}

func TestBoost_DiscoverSkills_ClaudeSkillsDir(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "scribe-discover-*")
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create skill in .claude/skills common dir
	claudeSkillDir := filepath.Join(tmpDir, ".claude", "skills", "claude-skill")
	_ = os.MkdirAll(claudeSkillDir, 0o755)
	_ = os.WriteFile(filepath.Join(claudeSkillDir, "SKILL.md"), []byte("---\nname: claude-skill\ndescription: Claude\n---\n# Claude\n"), 0o644)

	skills, err := DiscoverSkills(tmpDir)
	if err != nil {
		t.Fatalf("DiscoverSkills error: %v", err)
	}
	found := false
	for _, s := range skills {
		if s.Name == "claude-skill" {
			found = true
		}
	}
	if !found {
		t.Error("skill in .claude/skills not discovered")
	}
}

func TestBoost_DiscoverSkillsWithDepth_Limit(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "scribe-depth-*")
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a skill at depth 3 (too deep for depth=1)
	deepDir := filepath.Join(tmpDir, "a", "b", "c")
	_ = os.MkdirAll(deepDir, 0o755)
	_ = os.WriteFile(filepath.Join(deepDir, "SKILL.md"), []byte("---\nname: deep\ndescription: Deep\n---\n# D\n"), 0o644)

	// Depth 1 should not find it
	skills, _ := DiscoverSkillsWithDepth(tmpDir, 1)
	for _, s := range skills {
		if s.Name == "deep" {
			t.Error("deep skill should not be found at depth 1")
		}
	}

	// Depth 5 should find it
	skills, _ = DiscoverSkillsWithDepth(tmpDir, 5)
	found := false
	for _, s := range skills {
		if s.Name == "deep" {
			found = true
		}
	}
	if !found {
		t.Error("deep skill should be found at depth 5")
	}
}

func TestDiscoverSkills_CaseInsensitiveFilename(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-discover-case-*")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a skill with standard SKILL.md
	skill1Dir := filepath.Join(tmpDir, "standard")
	_ = os.MkdirAll(skill1Dir, 0o755)
	_ = os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte("---\nname: standard\ndescription: Standard casing\n---\n# Standard\n"), 0o644)

	// Create a skill with all-uppercase SKILL.MD (like better-auth/skills security/)
	skill2Dir := filepath.Join(tmpDir, "uppercase")
	_ = os.MkdirAll(skill2Dir, 0o755)
	_ = os.WriteFile(filepath.Join(skill2Dir, "SKILL.MD"), []byte("---\nname: uppercase\ndescription: Uppercase extension\n---\n# Uppercase\n"), 0o644)

	// Create a skill with lowercase skill.md
	skill3Dir := filepath.Join(tmpDir, "lowercase")
	_ = os.MkdirAll(skill3Dir, 0o755)
	_ = os.WriteFile(filepath.Join(skill3Dir, "skill.md"), []byte("---\nname: lowercase\ndescription: Lowercase filename\n---\n# Lowercase\n"), 0o644)

	skills, err := DiscoverSkills(tmpDir)
	if err != nil {
		t.Fatalf("DiscoverSkills() error: %v", err)
	}

	nameSet := make(map[string]bool)
	for _, s := range skills {
		nameSet[s.Name] = true
	}

	if !nameSet["standard"] {
		t.Error("standard SKILL.md not discovered")
	}
	if !nameSet["uppercase"] {
		t.Error("uppercase SKILL.MD not discovered")
	}
	if !nameSet["lowercase"] {
		t.Error("lowercase skill.md not discovered")
	}
	if len(skills) != 3 {
		t.Errorf("expected 3 skills, got %d", len(skills))
	}
}

func TestDiscoverSkills_CaseInsensitiveRoot(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-discover-root-case-*")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a SKILL.MD at root with uppercase extension
	_ = os.WriteFile(filepath.Join(tmpDir, "SKILL.MD"), []byte("---\nname: root-upper\ndescription: Root uppercase\n---\n# Root\n"), 0o644)

	skills, err := DiscoverSkills(tmpDir)
	if err != nil {
		t.Fatalf("DiscoverSkills() error: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if skills[0].Name != "root-upper" {
		t.Errorf("skill name = %q, want 'root-upper'", skills[0].Name)
	}
}

func TestDiscoverSkills_CaseInsensitiveInCommonDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-discover-common-case-*")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Put a SKILL.MD (uppercase) in the "skills/" common dir
	skillDir := filepath.Join(tmpDir, "skills", "upper-skill")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.MD"), []byte("---\nname: upper-common\ndescription: Uppercase in common dir\n---\n# Upper\n"), 0o644)

	skills, err := DiscoverSkills(tmpDir)
	if err != nil {
		t.Fatalf("DiscoverSkills() error: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if skills[0].Name != "upper-common" {
		t.Errorf("skill name = %q, want 'upper-common'", skills[0].Name)
	}
}

// ============================================================================
// discoverSkillsInDir (skills.go) - private helper
// ============================================================================

func TestBoost_DiscoverSkillsInDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-discover-*")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create nested skills at different depths
	skill1 := filepath.Join(tmpDir, "skill1")
	_ = os.MkdirAll(skill1, 0o755)
	_ = os.WriteFile(filepath.Join(skill1, "SKILL.md"), []byte("---\nname: skill1\ndescription: S1\n---\n# S1\n"), 0o644)

	deepSkill := filepath.Join(tmpDir, "deep", "deeper", "skill2")
	_ = os.MkdirAll(deepSkill, 0o755)
	_ = os.WriteFile(filepath.Join(deepSkill, "SKILL.md"), []byte("---\nname: skill2\ndescription: S2\n---\n# S2\n"), 0o644)

	// Depth 1 should only find skill1
	skills, parseErrs := discoverSkillsInDir(tmpDir, 1)
	if len(parseErrs) > 0 {
		t.Fatalf("discoverSkillsInDir(depth=1) parse errors: %v", parseErrs)
	}
	if len(skills) != 1 {
		t.Errorf("depth 1: expected 1 skill, got %d", len(skills))
	}

	// Depth 5 should find both
	skills, parseErrs = discoverSkillsInDir(tmpDir, 5)
	if len(parseErrs) > 0 {
		t.Fatalf("discoverSkillsInDir(depth=5) parse errors: %v", parseErrs)
	}
	if len(skills) != 2 {
		t.Errorf("depth 5: expected 2 skills, got %d", len(skills))
	}
}

// ============================================================================
// ReadSkill (skills.go)
// ============================================================================

func TestBoost_ReadSkill_WithMeta(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	skillDir := filepath.Join(tmpDir, ".scribe", "scrolls", "read-test")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: read-test\ndescription: Read test\n---\n# Read\n"), 0o644)
	_ = WriteSkillMeta(filepath.Join(skillDir, ".scribe-meta.json"), &SkillMeta{
		Source:      "test/repo",
		SourceType:  "github",
		ContentHash: ComputeContentHash("test"),
		InstalledAt: "2025-01-01T00:00:00Z",
		UpdatedAt:   "2025-01-01T00:00:00Z",
	})

	skill, err := ReadSkill("read-test")
	if err != nil {
		t.Fatalf("ReadSkill() error: %v", err)
	}
	if skill.Name != "read-test" {
		t.Errorf("skill name = %q, want 'read-test'", skill.Name)
	}
	if skill.Meta == nil {
		t.Fatal("skill.Meta is nil")
	}
	if skill.Meta.SourceType != "github" {
		t.Errorf("meta.SourceType = %q, want 'github'", skill.Meta.SourceType)
	}
}

func TestBoost_ReadSkill_NonExistent(t *testing.T) {
	_ = setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	_, err := ReadSkill("nonexistent-skill-xyz")
	if err == nil {
		t.Error("expected error for nonexistent skill")
	}
}

// ============================================================================
// ReadAllSkills / GetAllSkillInfo (skills.go)
// ============================================================================

func TestBoost_ReadAllSkills(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// Create skills
	for _, name := range []string{"read-all-a", "read-all-b"} {
		skillDir := filepath.Join(tmpDir, ".scribe", "scrolls", name)
		_ = os.MkdirAll(skillDir, 0o755)
		_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: "+name+"\ndescription: Test\n---\n# Test\n"), 0o644)
	}

	skills, err := ReadAllSkills()
	if err != nil {
		t.Fatalf("ReadAllSkills() error: %v", err)
	}
	if len(skills) != 2 {
		t.Errorf("ReadAllSkills() returned %d skills, want 2", len(skills))
	}
}

func TestBoost_ReadAllSkills_MixedValidity(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")

	// Valid skill
	s1 := filepath.Join(scrollsDir, "valid")
	_ = os.MkdirAll(s1, 0o755)
	_ = os.WriteFile(filepath.Join(s1, "SKILL.md"), []byte("---\nname: valid\ndescription: Valid\n---\n# V\n"), 0o644)

	// Invalid skill (bad SKILL.md)
	s2 := filepath.Join(scrollsDir, "invalid")
	_ = os.MkdirAll(s2, 0o755)
	_ = os.WriteFile(filepath.Join(s2, "SKILL.md"), []byte("no frontmatter"), 0o644)

	skills, err := ReadAllSkills()
	if err != nil {
		t.Fatalf("ReadAllSkills error: %v", err)
	}
	// Should only get the valid skill
	if len(skills) != 1 {
		t.Errorf("expected 1 valid skill, got %d", len(skills))
	}
}

func TestBoost_GetAllSkillInfo(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// Create a skill
	skillDir := filepath.Join(tmpDir, ".scribe", "scrolls", "info-skill")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: info-skill\ndescription: Info\n---\n# Info\n"), 0o644)

	infos, err := GetAllSkillInfo()
	if err != nil {
		t.Fatalf("GetAllSkillInfo() error: %v", err)
	}
	if len(infos) != 1 {
		t.Fatalf("expected 1 skill info, got %d", len(infos))
	}
	if infos[0].Name != "info-skill" {
		t.Errorf("skill name = %q, want 'info-skill'", infos[0].Name)
	}
}

func TestGetAllSkillInfo_UsesDirNameNotFrontmatterName(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	// Create a skill where directory name differs from frontmatter name
	skillDir := filepath.Join(tmpDir, ".scribe", "scrolls", "dir-name")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: frontmatter-name\ndescription: Mismatched names\n---\n# Content\n"), 0o644)

	infos, err := GetAllSkillInfo()
	if err != nil {
		t.Fatalf("GetAllSkillInfo() error: %v", err)
	}
	if len(infos) != 1 {
		t.Fatalf("expected 1 skill info, got %d", len(infos))
	}
	// SkillInfo.Name must be the directory name, not the frontmatter name,
	// because backend operations (UpdateSkill, RemoveSkill, etc.) resolve by directory name.
	if infos[0].Name != "dir-name" {
		t.Errorf("SkillInfo.Name = %q, want 'dir-name' (directory name), not 'frontmatter-name'", infos[0].Name)
	}
	if infos[0].Description != "Mismatched names" {
		t.Errorf("SkillInfo.Description = %q, want 'Mismatched names'", infos[0].Description)
	}
}

// ============================================================================
// ParseSkillContent (skills.go)
// ============================================================================

func TestBoost_ParseSkillContent_WithMetadata(t *testing.T) {
	content := "---\nname: meta-skill\ndescription: Has metadata\nauthor: test\nversion: 1.0\n---\n# Body\n"
	skill, err := ParseSkillContent(content, "/tmp")
	if err != nil {
		t.Fatalf("ParseSkillContent() error: %v", err)
	}
	if skill.Name != "meta-skill" {
		t.Errorf("name = %q, want 'meta-skill'", skill.Name)
	}
	if skill.Metadata == nil {
		t.Fatal("metadata is nil")
	}
}

func TestBoost_ParseSkillContent_NoFrontmatter(t *testing.T) {
	_, err := ParseSkillContent("# No frontmatter here", "/tmp")
	if err == nil {
		t.Error("expected error for missing frontmatter")
	}
}

func TestBoost_ParseSkillContent_MissingName(t *testing.T) {
	content := "---\ndescription: No name\n---\n# Body\n"
	_, err := ParseSkillContent(content, "/tmp")
	if err != ErrMissingName {
		t.Errorf("error = %v, want ErrMissingName", err)
	}
}

func TestBoost_ParseSkillContent_MissingDescription(t *testing.T) {
	content := "---\nname: no-desc\n---\n# Body\n"
	_, err := ParseSkillContent(content, "/tmp")
	if err != ErrMissingDesc {
		t.Errorf("error = %v, want ErrMissingDesc", err)
	}
}

func TestBoost_ParseSkillContent_InvalidYAML(t *testing.T) {
	content := "---\ninvalid: [\n---\n# Body\n"
	_, err := ParseSkillContent(content, "/tmp")
	if err == nil {
		t.Error("expected error for invalid YAML in frontmatter")
	}
}

func TestBoost_ParseSkillMd_NonExistent(t *testing.T) {
	_, err := ParseSkillMd("/nonexistent/path/SKILL.md")
	if err == nil {
		t.Error("expected error for nonexistent SKILL.md")
	}
}

// ============================================================================
// ListInstalledSkills / SkillExists (skills.go)
// ============================================================================

func TestBoost_ListInstalledSkills_MixedContent(t *testing.T) {
	tmpDir := setupTempHome(t)
	_ = EnsureScribeDirs()

	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")

	// Create a valid skill
	skillDir := filepath.Join(scrollsDir, "valid-skill")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Valid"), 0o644)

	// Create a directory without SKILL.md
	_ = os.MkdirAll(filepath.Join(scrollsDir, "not-a-skill"), 0o755)

	// Create a regular file (not a directory)
	_ = os.WriteFile(filepath.Join(scrollsDir, "random.txt"), []byte("data"), 0o644)

	skills, err := ListInstalledSkills()
	if err != nil {
		t.Fatalf("ListInstalledSkills error: %v", err)
	}
	if len(skills) != 1 {
		t.Errorf("expected 1 skill, got %d", len(skills))
	}
}

func TestBoost_SkillExists_IsolatedHome(t *testing.T) {
	tmpDir := setupTempHome(t)
	_ = EnsureScribeDirs()

	exists, err := SkillExists("nonexistent")
	if err != nil {
		t.Fatalf("SkillExists error: %v", err)
	}
	if exists {
		t.Error("expected false for nonexistent skill")
	}

	// Create skill
	skillDir := filepath.Join(tmpDir, ".scribe", "scrolls", "exists-test")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# E"), 0o644)

	exists, err = SkillExists("exists-test")
	if err != nil {
		t.Fatalf("SkillExists error: %v", err)
	}
	if !exists {
		t.Error("expected true for existing skill")
	}
}
