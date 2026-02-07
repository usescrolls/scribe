package scribe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseSourceString(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectType     string
		expectOwner    string
		expectRepo     string
		expectRef      string
		expectSubpath  string
		expectURL      string
		expectError    bool
		expectErrorMsg string
	}{
		{
			name:        "github shorthand",
			input:       "owner/repo",
			expectType:  "github",
			expectOwner: "owner",
			expectRepo:  "repo",
			expectURL:   "https://github.com/owner/repo",
		},
		{
			name:        "github shorthand with branch",
			input:       "owner/repo#develop",
			expectType:  "github",
			expectOwner: "owner",
			expectRepo:  "repo",
			expectRef:   "develop",
		},
		{
			name:          "github shorthand with subpath",
			input:         "owner/repo/skills/react",
			expectType:    "github",
			expectOwner:   "owner",
			expectRepo:    "repo",
			expectSubpath: "skills/react",
		},
		{
			name:          "github shorthand with subpath and ref",
			input:         "owner/repo/skills/react#v2.0",
			expectType:    "github",
			expectOwner:   "owner",
			expectRepo:    "repo",
			expectSubpath: "skills/react",
			expectRef:     "v2.0",
		},
		{
			name:        "github full url",
			input:       "https://github.com/owner/repo",
			expectType:  "github",
			expectOwner: "owner",
			expectRepo:  "repo",
		},
		{
			name:        "github url with .git suffix",
			input:       "https://github.com/owner/repo.git",
			expectType:  "github",
			expectOwner: "owner",
			expectRepo:  "repo",
		},
		{
			name:          "github url with subpath no tree",
			input:         "https://github.com/owner/repo/skills/react",
			expectType:    "github",
			expectOwner:   "owner",
			expectRepo:    "repo",
			expectSubpath: "skills/react",
		},
		{
			name:          "github url with tree/branch",
			input:         "https://github.com/owner/repo/tree/main/skills",
			expectType:    "github",
			expectOwner:   "owner",
			expectRepo:    "repo",
			expectRef:     "main",
			expectSubpath: "skills",
		},
		{
			name:        "gitlab url",
			input:       "https://gitlab.com/owner/repo",
			expectType:  "gitlab",
			expectOwner: "owner",
			expectRepo:  "repo",
		},
		{
			name:        "gitlab url with .git suffix",
			input:       "https://gitlab.com/owner/repo.git",
			expectType:  "gitlab",
			expectOwner: "owner",
			expectRepo:  "repo",
		},
		{
			name:       "zip url",
			input:      "https://example.com/skills.zip",
			expectType: "zip",
			expectURL:  "https://example.com/skills.zip",
		},
		{
			name:       "well-known url",
			input:      "https://example.com/skills",
			expectType: "well-known",
			expectURL:  "https://example.com/skills",
		},
		{
			name:       "local relative path",
			input:      "./my-skills",
			expectType: "local",
		},
		{
			name:       "local absolute path",
			input:      "/tmp/my-skills",
			expectType: "local",
		},
		{
			name:       "local tilde path",
			input:      "~/my-skills",
			expectType: "local",
		},
		{
			name:           "empty string",
			input:          "",
			expectError:    true,
			expectErrorMsg: "empty",
		},
		{
			name:           "whitespace only",
			input:          "   ",
			expectError:    true,
			expectErrorMsg: "empty",
		},
		{
			name:           "single word no slash",
			input:          "noslash",
			expectError:    true,
			expectErrorMsg: "invalid",
		},
		{
			name:        "input with leading/trailing whitespace",
			input:       "  owner/repo  ",
			expectType:  "github",
			expectOwner: "owner",
			expectRepo:  "repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, err := ParseSourceString(tt.input)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.expectErrorMsg != "" && !strings.Contains(err.Error(), tt.expectErrorMsg) {
					t.Errorf("error = %q, should contain %q", err.Error(), tt.expectErrorMsg)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if source.Type != tt.expectType {
				t.Errorf("Type = %q, want %q", source.Type, tt.expectType)
			}
			if tt.expectOwner != "" && source.Owner != tt.expectOwner {
				t.Errorf("Owner = %q, want %q", source.Owner, tt.expectOwner)
			}
			if tt.expectRepo != "" && source.Repo != tt.expectRepo {
				t.Errorf("Repo = %q, want %q", source.Repo, tt.expectRepo)
			}
			if tt.expectRef != "" && source.Ref != tt.expectRef {
				t.Errorf("Ref = %q, want %q", source.Ref, tt.expectRef)
			}
			if tt.expectSubpath != "" && source.Subpath != tt.expectSubpath {
				t.Errorf("Subpath = %q, want %q", source.Subpath, tt.expectSubpath)
			}
			if tt.expectURL != "" && source.URL != tt.expectURL {
				t.Errorf("URL = %q, want %q", source.URL, tt.expectURL)
			}
		})
	}
}

func TestParseSourceStringLocalPathIsAbsolute(t *testing.T) {
	source, err := ParseSourceString("./relative")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if source.Type != "local" {
		t.Fatalf("Type = %q, want 'local'", source.Type)
	}
	// LocalPath should be absolute regardless of input
	if !strings.HasPrefix(source.LocalPath, "/") {
		t.Errorf("LocalPath = %q, expected absolute path", source.LocalPath)
	}
}

// Integration test: ParseSourceString → FetchAndDiscoverSkills → InstallSkill
func TestInstallFromLocalSource_SingleSkill(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	// Create a local source with one skill
	srcDir := filepath.Join(tmpDir, "my-skills")
	_ = os.MkdirAll(srcDir, 0o755)
	_ = os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte(
		"---\nname: from-source-test\ndescription: Test install from source\n---\n# Test\n",
	), 0o644)

	// Parse
	source, err := ParseSourceString(srcDir)
	if err != nil {
		t.Fatalf("ParseSourceString error: %v", err)
	}
	if source.Type != "local" {
		t.Fatalf("Type = %q, want 'local'", source.Type)
	}

	// Fetch & discover
	skills, fetchResult, err := FetchAndDiscoverSkills(source)
	if fetchResult != nil {
		defer fetchResult.Cleanup()
	}
	if err != nil {
		t.Fatalf("FetchAndDiscoverSkills error: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("discovered %d skills, want 1", len(skills))
	}
	if skills[0].Name != "from-source-test" {
		t.Errorf("skill name = %q, want 'from-source-test'", skills[0].Name)
	}

	// Install
	err = InstallSkill(skills[0], source, InstallOptions{})
	if err != nil {
		t.Fatalf("InstallSkill error: %v", err)
	}

	// Verify
	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	if _, err := os.Stat(filepath.Join(scrollsDir, "from-source-test", "SKILL.md")); err != nil {
		t.Error("SKILL.md not found in installed location")
	}
	if _, err := os.Stat(filepath.Join(scrollsDir, "from-source-test", ".scribe-meta.json")); err != nil {
		t.Error("metadata not found in installed location")
	}
}

func TestInstallFromLocalSource_MultipleSkills(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()
	_ = EnsureDefaultWorkspace()

	// Create a local source with multiple skills
	srcDir := filepath.Join(tmpDir, "multi-skills")
	for _, name := range []string{"skill-alpha", "skill-beta", "skill-gamma"} {
		skillDir := filepath.Join(srcDir, "skills", name)
		_ = os.MkdirAll(skillDir, 0o755)
		_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(
			"---\nname: "+name+"\ndescription: "+name+" description\n---\n# "+name+"\n",
		), 0o644)
	}

	// Parse
	source, err := ParseSourceString(srcDir)
	if err != nil {
		t.Fatalf("ParseSourceString error: %v", err)
	}

	// Fetch & discover
	skills, fetchResult, err := FetchAndDiscoverSkills(source)
	if fetchResult != nil {
		defer fetchResult.Cleanup()
	}
	if err != nil {
		t.Fatalf("FetchAndDiscoverSkills error: %v", err)
	}
	if len(skills) != 3 {
		t.Fatalf("discovered %d skills, want 3", len(skills))
	}

	// Install all
	installed := []string{}
	for _, skill := range skills {
		if err := InstallSkill(skill, source, InstallOptions{}); err != nil {
			t.Errorf("InstallSkill(%s) error: %v", skill.Name, err)
			continue
		}
		if err := AddSkillToActiveAndDefaultWorkspace(skill.Name); err != nil {
			t.Errorf("AddSkillToWorkspace(%s) error: %v", skill.Name, err)
		}
		installed = append(installed, skill.Name)
	}

	if len(installed) != 3 {
		t.Fatalf("installed %d skills, want 3", len(installed))
	}

	// Verify each skill exists in scrolls
	scrollsDir := filepath.Join(tmpDir, ".scribe", "scrolls")
	for _, name := range installed {
		if _, err := os.Stat(filepath.Join(scrollsDir, name, "SKILL.md")); err != nil {
			t.Errorf("skill %q not found in scrolls", name)
		}
	}

	// Verify skills were added to default workspace
	ws, err := GetWorkspace("default")
	if err != nil {
		t.Fatalf("GetWorkspace error: %v", err)
	}
	for _, name := range installed {
		found := false
		for _, s := range ws.Skills {
			if s == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("skill %q not in default workspace", name)
		}
	}
}

func TestInstallFromLocalSource_NoSkillsFound(t *testing.T) {
	setupTempHome(t)
	InitLoggerCLI(false)

	// Create an empty directory
	emptyDir, err := os.MkdirTemp("", "scribe-empty-*")
	if err != nil {
		t.Fatalf("MkdirTemp error: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(emptyDir) })

	source, err := ParseSourceString(emptyDir)
	if err != nil {
		t.Fatalf("ParseSourceString error: %v", err)
	}

	_, fetchResult, err := FetchAndDiscoverSkills(source)
	if fetchResult != nil {
		defer fetchResult.Cleanup()
	}
	if err == nil {
		t.Error("expected error for empty source, got nil")
	}
}
