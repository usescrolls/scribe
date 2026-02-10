package scribe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ============================================================================
// ComputeContentHash (meta.go)
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
// NewSkillMeta / UpdateSkillMeta (meta.go)
// ============================================================================

func TestNewSkillMeta_GitHub(t *testing.T) {
	source := &SourceInfo{
		Type:  "github",
		Owner: "octocat",
		Repo:  "skills",
		URL:   "https://github.com/octocat/skills",
	}
	meta := NewSkillMeta(source, "skills/test", "content body", nil)

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
	meta := NewSkillMeta(source, "", "local content", nil)

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
	meta := NewSkillMeta(source, "", "content", nil)
	// formatSource (private) includes ref only for non main/master
	if !strings.Contains(meta.Source, "v2.0.0") {
		t.Errorf("Source = %q, expected it to contain ref 'v2.0.0'", meta.Source)
	}
}

func TestUpdateSkillMeta_ContentChange(t *testing.T) {
	source := &SourceInfo{Type: "github", Owner: "u", Repo: "r"}
	meta := NewSkillMeta(source, "", "old content", nil)
	oldHash := meta.ContentHash
	oldUpdated := meta.UpdatedAt

	UpdateSkillMeta(meta, "new content", nil)

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
// SkillNeedsUpdate (meta.go)
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
// SaveSkillWithMeta (meta.go)
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
	err = SaveSkillWithMeta(destDir, skill, source, "skills/test-save", nil)
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

func TestBoost_SaveSkillWithMeta_MissingSource(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "scribe-save-*")
	defer func() { _ = os.RemoveAll(tmpDir) }()

	skill := &Skill{
		Name:        "bad-skill",
		Description: "Missing source",
		Path:        "/nonexistent/path",
	}
	source := &SourceInfo{Type: "local", LocalPath: "/nonexistent"}

	err := SaveSkillWithMeta(filepath.Join(tmpDir, "output"), skill, source, "", nil)
	if err == nil {
		t.Error("expected error when source SKILL.md doesn't exist")
	}
}

// ============================================================================
// LoadSkillWithMeta / ListSkillsWithMeta (meta.go)
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
// ReadSkillMeta / WriteSkillMeta roundtrip (meta.go)
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

func TestBoost_ReadSkillMeta_InvalidJSON(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "scribe-meta-*")
	defer func() { _ = os.RemoveAll(tmpDir) }()

	metaPath := filepath.Join(tmpDir, ".scribe-meta.json")
	_ = os.WriteFile(metaPath, []byte("not valid json"), 0o644)

	_, err := ReadSkillMeta(metaPath)
	if err == nil {
		t.Error("expected error for invalid JSON meta")
	}
}

func TestBoost_SkillMeta_AllFields(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "scribe-meta-*")
	defer func() { _ = os.RemoveAll(tmpDir) }()

	meta := &SkillMeta{
		Source:      "org/repo#branch",
		SourceType:  "github",
		SourceURL:   "https://github.com/org/repo",
		SkillPath:   "skills/my-skill",
		ContentHash: "sha256:abcdef",
		InstalledAt: "2025-01-01T00:00:00Z",
		UpdatedAt:   "2025-06-01T00:00:00Z",
	}

	path := filepath.Join(tmpDir, ".scribe-meta.json")
	_ = WriteSkillMeta(path, meta)

	loaded, err := ReadSkillMeta(path)
	if err != nil {
		t.Fatalf("ReadSkillMeta error: %v", err)
	}
	if loaded.SourceURL != meta.SourceURL {
		t.Errorf("SourceURL = %q, want %q", loaded.SourceURL, meta.SourceURL)
	}
	if loaded.SkillPath != meta.SkillPath {
		t.Errorf("SkillPath = %q, want %q", loaded.SkillPath, meta.SkillPath)
	}
}

// ============================================================================
// formatSource (meta.go) - private function
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

func TestFormatSourcePrivate_Bitbucket(t *testing.T) {
	source := &SourceInfo{Type: "bitbucket", Owner: "myteam", Repo: "myrepo"}
	result := formatSource(source)
	if result != "myteam/myrepo" {
		t.Errorf("formatSource(bitbucket) = %q, want 'myteam/myrepo'", result)
	}
}

func TestFormatSourcePrivate_BitbucketWithRef(t *testing.T) {
	source := &SourceInfo{Type: "bitbucket", Owner: "myteam", Repo: "myrepo", Ref: "develop"}
	result := formatSource(source)
	if result != "myteam/myrepo#develop" {
		t.Errorf("formatSource(bitbucket+ref) = %q, want 'myteam/myrepo#develop'", result)
	}
}

func TestFormatSourcePrivate_BitbucketMainRef(t *testing.T) {
	source := &SourceInfo{Type: "bitbucket", Owner: "myteam", Repo: "myrepo", Ref: "main"}
	result := formatSource(source)
	if result != "myteam/myrepo" {
		t.Errorf("formatSource(bitbucket+main) = %q, want 'myteam/myrepo' (main stripped)", result)
	}
}

func TestBoost_FormatSourcePrivate_Zip(t *testing.T) {
	source := &SourceInfo{Type: "zip", URL: "https://example.com/archive.zip"}
	meta := NewSkillMeta(source, "", "content", nil)
	if meta.Source != "https://example.com/archive.zip" {
		t.Errorf("meta.Source = %q, want URL", meta.Source)
	}
}

// ============================================================================
// NewSkillMeta / UpdateSkillMeta with GitCommitInfo (meta.go)
// ============================================================================

func TestNewSkillMeta_WithGitInfo(t *testing.T) {
	source := &SourceInfo{
		Type:  "github",
		Owner: "octocat",
		Repo:  "skills",
		URL:   "https://github.com/octocat/skills",
	}
	gitInfo := &GitCommitInfo{
		Hash: "abc1234",
		Date: "2025-06-15T10:30:00Z",
	}
	meta := NewSkillMeta(source, "skills/test", "content body", gitInfo)

	if meta.CommitHash != "abc1234" {
		t.Errorf("CommitHash = %q, want 'abc1234'", meta.CommitHash)
	}
	if meta.CommitDate != "2025-06-15T10:30:00Z" {
		t.Errorf("CommitDate = %q, want '2025-06-15T10:30:00Z'", meta.CommitDate)
	}
	// Other fields should still be set
	if meta.SourceType != "github" {
		t.Errorf("SourceType = %q, want 'github'", meta.SourceType)
	}
	if meta.ContentHash == "" {
		t.Error("ContentHash is empty")
	}
}

func TestNewSkillMeta_NilGitInfo(t *testing.T) {
	source := &SourceInfo{
		Type:      "local",
		LocalPath: "/path/to/skill",
	}
	meta := NewSkillMeta(source, "", "content", nil)

	if meta.CommitHash != "" {
		t.Errorf("CommitHash = %q, want empty for nil gitInfo", meta.CommitHash)
	}
	if meta.CommitDate != "" {
		t.Errorf("CommitDate = %q, want empty for nil gitInfo", meta.CommitDate)
	}
}

func TestUpdateSkillMeta_WithGitInfo(t *testing.T) {
	source := &SourceInfo{Type: "github", Owner: "u", Repo: "r"}
	meta := NewSkillMeta(source, "", "old content", nil)

	// Initially no commit info
	if meta.CommitHash != "" {
		t.Errorf("initial CommitHash = %q, want empty", meta.CommitHash)
	}

	gitInfo := &GitCommitInfo{
		Hash: "def5678",
		Date: "2025-07-01T12:00:00Z",
	}
	UpdateSkillMeta(meta, "new content", gitInfo)

	if meta.CommitHash != "def5678" {
		t.Errorf("CommitHash = %q, want 'def5678'", meta.CommitHash)
	}
	if meta.CommitDate != "2025-07-01T12:00:00Z" {
		t.Errorf("CommitDate = %q, want '2025-07-01T12:00:00Z'", meta.CommitDate)
	}
	if meta.ContentHash != ComputeContentHash("new content") {
		t.Errorf("ContentHash not updated to new content hash")
	}
}

func TestUpdateSkillMeta_NilGitInfo_PreservesExisting(t *testing.T) {
	source := &SourceInfo{Type: "github", Owner: "u", Repo: "r"}
	gitInfo := &GitCommitInfo{Hash: "abc1234", Date: "2025-06-15T10:00:00Z"}
	meta := NewSkillMeta(source, "", "content", gitInfo)

	// Update with nil gitInfo — existing commit fields should remain unchanged
	UpdateSkillMeta(meta, "new content", nil)

	if meta.CommitHash != "abc1234" {
		t.Errorf("CommitHash = %q, want 'abc1234' (should be preserved)", meta.CommitHash)
	}
	if meta.CommitDate != "2025-06-15T10:00:00Z" {
		t.Errorf("CommitDate = %q, want preserved value", meta.CommitDate)
	}
}

func TestSaveSkillWithMeta_WithGitInfo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a source skill directory with SKILL.md
	srcDir := filepath.Join(tmpDir, "src-skill")
	_ = os.MkdirAll(srcDir, 0o755)
	_ = os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte("---\nname: git-save\ndescription: Test git save\n---\n# Content\n"), 0o644)

	skill := &Skill{
		Name:        "git-save",
		Description: "Test git save",
		Path:        srcDir,
	}
	source := &SourceInfo{
		Type:  "github",
		Owner: "testuser",
		Repo:  "testrepo",
		URL:   "https://github.com/testuser/testrepo",
	}
	gitInfo := &GitCommitInfo{Hash: "fed9876", Date: "2025-08-01T08:00:00Z"}

	destDir := filepath.Join(tmpDir, "dest-skill")
	err = SaveSkillWithMeta(destDir, skill, source, "skills/git-save", gitInfo)
	if err != nil {
		t.Fatalf("SaveSkillWithMeta() error: %v", err)
	}

	meta, err := ReadSkillMeta(filepath.Join(destDir, ".scribe-meta.json"))
	if err != nil {
		t.Fatalf("ReadSkillMeta() error: %v", err)
	}
	if meta.CommitHash != "fed9876" {
		t.Errorf("meta.CommitHash = %q, want 'fed9876'", meta.CommitHash)
	}
	if meta.CommitDate != "2025-08-01T08:00:00Z" {
		t.Errorf("meta.CommitDate = %q, want '2025-08-01T08:00:00Z'", meta.CommitDate)
	}
}

func TestReadWriteSkillMeta_CommitFields_Roundtrip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	metaPath := filepath.Join(tmpDir, ".scribe-meta.json")
	original := &SkillMeta{
		Source:      "user/repo",
		SourceType:  "github",
		ContentHash: "sha256:abc",
		CommitHash:  "abc1234",
		CommitDate:  "2025-06-15T10:30:00Z",
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

	if loaded.CommitHash != "abc1234" {
		t.Errorf("CommitHash = %q, want 'abc1234'", loaded.CommitHash)
	}
	if loaded.CommitDate != "2025-06-15T10:30:00Z" {
		t.Errorf("CommitDate = %q, want '2025-06-15T10:30:00Z'", loaded.CommitDate)
	}
}

func TestReadWriteSkillMeta_IsPrivate_Roundtrip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	metaPath := filepath.Join(tmpDir, ".scribe-meta.json")
	original := &SkillMeta{
		Source:      "user/private-repo",
		SourceType:  "github",
		ContentHash: "sha256:abc",
		IsPrivate:   true,
		InstalledAt: "2025-06-01T12:00:00Z",
		UpdatedAt:   "2025-06-01T12:00:00Z",
	}

	if err := WriteSkillMeta(metaPath, original); err != nil {
		t.Fatalf("WriteSkillMeta() error: %v", err)
	}

	loaded, err := ReadSkillMeta(metaPath)
	if err != nil {
		t.Fatalf("ReadSkillMeta() error: %v", err)
	}

	if !loaded.IsPrivate {
		t.Error("IsPrivate = false after roundtrip, want true")
	}
}

func TestReadWriteSkillMeta_IsPrivateFalse_OmittedFromJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	metaPath := filepath.Join(tmpDir, ".scribe-meta.json")
	original := &SkillMeta{
		Source:      "user/public-repo",
		SourceType:  "github",
		ContentHash: "sha256:abc",
		IsPrivate:   false,
		InstalledAt: "2025-06-01T12:00:00Z",
		UpdatedAt:   "2025-06-01T12:00:00Z",
	}

	if err := WriteSkillMeta(metaPath, original); err != nil {
		t.Fatalf("WriteSkillMeta() error: %v", err)
	}

	// Verify the JSON doesn't contain isPrivate when false (omitempty)
	data, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if strings.Contains(string(data), "isPrivate") {
		t.Error("JSON contains 'isPrivate' field when false; expected omitempty to exclude it")
	}

	// Still roundtrips correctly
	loaded, err := ReadSkillMeta(metaPath)
	if err != nil {
		t.Fatalf("ReadSkillMeta() error: %v", err)
	}
	if loaded.IsPrivate {
		t.Error("IsPrivate = true after roundtrip of false value")
	}
}
