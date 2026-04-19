package scribe

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSourceQualifier_GitHub(t *testing.T) {
	source := &SourceInfo{Type: "github", Owner: "alice", Repo: "skills"}
	got := SourceQualifier(source)
	if got != "alice-skills" {
		t.Errorf("SourceQualifier() = %q, want %q", got, "alice-skills")
	}
}

func TestSourceQualifier_GitLab(t *testing.T) {
	source := &SourceInfo{Type: "gitlab", Owner: "bob", Repo: "tools"}
	got := SourceQualifier(source)
	if got != "gitlab-bob-tools" {
		t.Errorf("SourceQualifier() = %q, want %q", got, "gitlab-bob-tools")
	}
}

func TestSourceQualifier_GenericGitUsesHost(t *testing.T) {
	source := &SourceInfo{
		Type:  "git",
		Owner: "bob",
		Repo:  "tools",
		URL:   "https://git.example.com/team/tools.git",
	}
	got := SourceQualifier(source)
	if got != "git-example-com-team-tools" {
		t.Errorf("SourceQualifier() = %q, want %q", got, "git-example-com-team-tools")
	}
}

func TestSourceQualifier_Local(t *testing.T) {
	source := &SourceInfo{Type: "local", LocalPath: "/home/user/my-skills"}
	got := SourceQualifier(source)
	if got != "my-skills" {
		t.Errorf("SourceQualifier() = %q, want %q", got, "my-skills")
	}
}

func TestSourceQualifier_URL(t *testing.T) {
	source := &SourceInfo{Type: "url", URL: "https://example.com/skills/pack"}
	got := SourceQualifier(source)
	if got == "" {
		t.Error("SourceQualifier() returned empty for URL source")
	}
}

func TestSourceQualifier_EmptyOwner(t *testing.T) {
	source := &SourceInfo{Type: "github"}
	got := SourceQualifier(source)
	if got != "" {
		t.Errorf("SourceQualifier() = %q, want empty", got)
	}
}

func TestQualifiedName(t *testing.T) {
	got := QualifiedName("alice-skills", "commit")
	if got != "alice-skills--commit" {
		t.Errorf("QualifiedName() = %q, want %q", got, "alice-skills--commit")
	}
}

func TestQualifiedName_EmptyQualifier(t *testing.T) {
	got := QualifiedName("", "commit")
	if got != "commit" {
		t.Errorf("QualifiedName() = %q, want %q", got, "commit")
	}
}

func TestIsQualifiedName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"commit", false},
		{"alice-skills--commit", true},
		{"a--b", true},
	}
	for _, tt := range tests {
		if got := IsQualifiedName(tt.name); got != tt.want {
			t.Errorf("IsQualifiedName(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestSourceQualifierFromMeta(t *testing.T) {
	meta := &SkillMeta{Source: "alice/skills", SourceType: "github"}
	got := SourceQualifierFromMeta(meta)
	if got != "alice-skills" {
		t.Errorf("SourceQualifierFromMeta() = %q, want %q", got, "alice-skills")
	}
}

func TestSourceQualifierFromMeta_GitLab(t *testing.T) {
	meta := &SkillMeta{Source: "alice/skills", SourceType: "gitlab"}
	got := SourceQualifierFromMeta(meta)
	if got != "gitlab-alice-skills" {
		t.Errorf("SourceQualifierFromMeta() = %q, want %q", got, "gitlab-alice-skills")
	}
}

func TestSourceQualifierFromMeta_Nil(t *testing.T) {
	got := SourceQualifierFromMeta(nil)
	if got != "" {
		t.Errorf("SourceQualifierFromMeta(nil) = %q, want empty", got)
	}
}

func TestFrontmatterNameFromStorage(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"commit", "commit"},
		{"alice-skills--commit", "commit"},
		{"a--b--c", "c"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := FrontmatterNameFromStorage(tt.input); got != tt.want {
			t.Errorf("FrontmatterNameFromStorage(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestGetFrontmatterName(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	skillDir := filepath.Join(tmpDir, ".scribe", "scrolls", "my-skill")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"),
		[]byte("---\nname: My Skill\ndescription: test\n---\n# Test\n"), 0o644)

	got, err := GetFrontmatterName("my-skill")
	if err != nil {
		t.Fatalf("GetFrontmatterName() error: %v", err)
	}
	if got != "my-skill" {
		t.Errorf("GetFrontmatterName() = %q, want %q", got, "my-skill")
	}
}
