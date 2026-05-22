package scribe

import "testing"

func TestSearchSkillInfoFuzzyMatchesNameDescriptionAndSource(t *testing.T) {
	skills := []SkillInfo{
		{
			Name:        "typescript-patterns",
			Description: "TypeScript tips",
			Source:      "local/path",
			Agents:      []string{"claude-code"},
		},
		{
			Name:        "react-best-practices",
			DisplayName: "react-best-practices",
			Description: "React component patterns",
			Source:      "vercel-labs/skills",
			SourceType:  "github",
			Agents:      []string{"cursor"},
		},
		{
			Name:        "go-idioms",
			Description: "Go patterns",
			Source:      "example/go-skills",
			SourceType:  "github",
			Agents:      []string{"claude-code"},
		},
	}

	tests := []struct {
		name  string
		query string
		want  []string
	}{
		{
			name:  "subsequence name match",
			query: "rct",
			want:  []string{"react-best-practices"},
		},
		{
			name:  "multi-token match",
			query: "react comp",
			want:  []string{"react-best-practices"},
		},
		{
			name:  "source match",
			query: "vercel",
			want:  []string{"react-best-practices"},
		},
		{
			name:  "agent match",
			query: "cursor",
			want:  []string{"react-best-practices"},
		},
		{
			name:  "no match",
			query: "python",
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SearchSkillInfo(skills, tt.query)
			if len(got) != len(tt.want) {
				t.Fatalf("SearchSkillInfo() returned %d results, want %d: %#v", len(got), len(tt.want), got)
			}
			for i, want := range tt.want {
				if got[i].Name != want {
					t.Errorf("result %d = %q, want %q", i, got[i].Name, want)
				}
			}
		})
	}
}

func TestSearchSkillInfoRanksNameAboveDescription(t *testing.T) {
	skills := []SkillInfo{
		{
			Name:        "go-idioms",
			Description: "React migration notes",
		},
		{
			Name:        "react-patterns",
			Description: "Component patterns",
		},
	}

	got := SearchSkillInfo(skills, "react")
	if len(got) != 2 {
		t.Fatalf("SearchSkillInfo() returned %d results, want 2", len(got))
	}
	if got[0].Name != "react-patterns" {
		t.Errorf("first result = %q, want name match first", got[0].Name)
	}
}

func TestSearchSkillInfoRejectsDistantSubsequenceMatches(t *testing.T) {
	skills := []SkillInfo{
		{
			Name:   "typescript-tips",
			Source: "owner/typescript",
		},
		{
			Name:   "react-patterns",
			Source: "owner/react",
		},
	}

	got := SearchSkillInfo(skills, "rct")
	if len(got) != 1 {
		t.Fatalf("SearchSkillInfo() returned %d results, want 1: %#v", len(got), got)
	}
	if got[0].Name != "react-patterns" {
		t.Errorf("result = %q, want react-patterns", got[0].Name)
	}
}

func TestSearchSkillInfoMatchesSkillMdContent(t *testing.T) {
	skills := []SkillInfo{
		{
			Name:        "metadata-only",
			Description: "No body marker here",
			Content:     "---\nname: metadata-only\ndescription: No body marker here\n---\nUse browser automation for visual checks.",
		},
		{
			Name:        "other-skill",
			Description: "Different content",
			Content:     "---\nname: other-skill\ndescription: Different content\n---\nUse database migrations.",
		},
	}

	got := SearchSkillInfo(skills, "browser automation")
	if len(got) != 1 {
		t.Fatalf("SearchSkillInfo() returned %d results, want 1: %#v", len(got), got)
	}
	if got[0].Name != "metadata-only" {
		t.Errorf("result = %q, want metadata-only", got[0].Name)
	}
}

func TestSearchSkillInfoEmptyQueryPreservesInputOrder(t *testing.T) {
	skills := []SkillInfo{
		{Name: "beta"},
		{Name: "alpha"},
	}

	got := SearchSkillInfo(skills, "  ")
	if len(got) != 2 {
		t.Fatalf("SearchSkillInfo() returned %d results, want 2", len(got))
	}
	if got[0].Name != "beta" || got[1].Name != "alpha" {
		t.Errorf("empty query should preserve input order, got %#v", got)
	}
}
