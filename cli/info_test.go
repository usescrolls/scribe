package cli

import (
	"encoding/json"
	"strings"
	"testing"

	scribe "gitlab.com/usescrolls/scribe/internal"
)

// ---------------------------------------------------------------------------
// Tests for infoJSON (info.go)
// ---------------------------------------------------------------------------

func TestInfoJSON(t *testing.T) {
	skill := &scribe.Skill{
		Name:        "test-skill",
		Description: "A test skill",
		Metadata:    map[string]any{"version": "1.0"},
		Meta: &scribe.SkillMeta{
			Source:      "owner/repo",
			SourceType:  "github",
			SourceURL:   "https://github.com/owner/repo",
			SkillPath:   "skills/test",
			ContentHash: "sha256:abc123",
			InstalledAt: "2024-06-15T10:00:00Z",
			UpdatedAt:   "2024-07-01T12:00:00Z",
		},
	}
	agents := []string{"claude-code", "cursor"}

	output := captureStdout(t, func() {
		err := infoJSON(skill, agents)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var parsed skillDetailJSON
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v\nOutput: %s", err, output)
	}
	if parsed.Name != "test-skill" {
		t.Errorf("name = %q, expected %q", parsed.Name, "test-skill")
	}
	if parsed.Description != "A test skill" {
		t.Errorf("description = %q, expected %q", parsed.Description, "A test skill")
	}
	if parsed.Source != "owner/repo" {
		t.Errorf("source = %q, expected %q", parsed.Source, "owner/repo")
	}
	if parsed.SourceType != "github" {
		t.Errorf("sourceType = %q, expected %q", parsed.SourceType, "github")
	}
	if parsed.SourceURL != "https://github.com/owner/repo" {
		t.Errorf("sourceUrl = %q, expected %q", parsed.SourceURL, "https://github.com/owner/repo")
	}
	if parsed.ContentHash != "sha256:abc123" {
		t.Errorf("contentHash = %q, expected %q", parsed.ContentHash, "sha256:abc123")
	}
	if parsed.InstalledAt != "2024-06-15T10:00:00Z" {
		t.Errorf("installedAt = %q, expected %q", parsed.InstalledAt, "2024-06-15T10:00:00Z")
	}
	if parsed.UpdatedAt != "2024-07-01T12:00:00Z" {
		t.Errorf("updatedAt = %q, expected %q", parsed.UpdatedAt, "2024-07-01T12:00:00Z")
	}
	if len(parsed.Agents) != 2 {
		t.Errorf("agents count = %d, expected 2", len(parsed.Agents))
	}
}

func TestInfoJSONNoMeta(t *testing.T) {
	skill := &scribe.Skill{
		Name:        "bare-skill",
		Description: "No metadata",
		Meta:        nil,
	}

	output := captureStdout(t, func() {
		err := infoJSON(skill, []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var parsed skillDetailJSON
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if parsed.Name != "bare-skill" {
		t.Errorf("name = %q, expected %q", parsed.Name, "bare-skill")
	}
	if parsed.Source != "" {
		t.Errorf("source should be empty, got %q", parsed.Source)
	}
	if parsed.ContentHash != "" {
		t.Errorf("contentHash should be empty, got %q", parsed.ContentHash)
	}
}

// ---------------------------------------------------------------------------
// Tests for infoTable (info.go)
// ---------------------------------------------------------------------------

func TestInfoTable(t *testing.T) {
	skill := &scribe.Skill{
		Name:        "test-skill",
		Description: "A test skill",
		Meta: &scribe.SkillMeta{
			Source:      "owner/repo",
			SourceType:  "github",
			SourceURL:   "https://github.com/owner/repo",
			ContentHash: "sha256:abc123",
			InstalledAt: "2024-06-15T10:00:00Z",
			UpdatedAt:   "2024-07-01T12:00:00Z",
		},
	}

	output := captureStdout(t, func() {
		err := infoTable(skill, []string{"claude-code"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Name:") {
		t.Error("expected 'Name:' in output")
	}
	if !strings.Contains(output, "test-skill") {
		t.Error("expected 'test-skill' in output")
	}
	if !strings.Contains(output, "Source:") {
		t.Error("expected 'Source:' in output")
	}
	if !strings.Contains(output, "owner/repo") {
		t.Error("expected 'owner/repo' in output")
	}
	if !strings.Contains(output, "Agents:") {
		t.Error("expected 'Agents:' in output")
	}
	if !strings.Contains(output, "claude-code") {
		t.Error("expected 'claude-code' in output")
	}
	if !strings.Contains(output, "Installed:") {
		t.Error("expected 'Installed:' in output")
	}
	if !strings.Contains(output, "2024-06-15") {
		t.Error("expected formatted date in output")
	}
}

func TestInfoTableNoAgents(t *testing.T) {
	skill := &scribe.Skill{
		Name:        "lonely-skill",
		Description: "No agents",
		Meta:        nil,
	}

	output := captureStdout(t, func() {
		err := infoTable(skill, []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "(none)") {
		t.Error("expected '(none)' for agents when no agents present")
	}
}

func TestInfoTableNoMeta(t *testing.T) {
	skill := &scribe.Skill{
		Name:        "no-meta-info",
		Description: "No metadata info test",
		Meta:        nil,
	}

	output := captureStdout(t, func() {
		_ = infoTable(skill, []string{})
	})

	if !strings.Contains(output, "no-meta-info") {
		t.Errorf("expected skill name in output, got: %s", output)
	}
	// Should not have Source: line since meta is nil
	if strings.Contains(output, "Source:") {
		t.Error("should not have Source: line when meta is nil")
	}
	if !strings.Contains(output, "(none)") {
		t.Error("expected '(none)' for agents")
	}
}

// ---------------------------------------------------------------------------
// Tests for runInfo (info.go)
// ---------------------------------------------------------------------------

func TestRunInfoJSONMode(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	jsonOutput = true

	installFakeSkill(t, "info-test-skill", "A test skill for info", "github", "owner/repo")

	output := captureStdout(t, func() {
		err := runInfo(infoCmd, []string{"info-test-skill"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var parsed skillDetailJSON
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v\nOutput: %s", err, output)
	}
	if parsed.Name != "info-test-skill" {
		t.Errorf("name = %q, expected %q", parsed.Name, "info-test-skill")
	}
	if parsed.Source != "owner/repo" {
		t.Errorf("source = %q, expected %q", parsed.Source, "owner/repo")
	}
}

func TestRunInfoTableMode(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()
	saveAndRestoreFlags(t)
	quiet = false
	jsonOutput = false

	installFakeSkill(t, "info-table-skill", "Table info test", "github", "owner/repo")

	output := captureStdout(t, func() {
		err := runInfo(infoCmd, []string{"info-table-skill"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "info-table-skill") {
		t.Errorf("expected skill name in output, got: %s", output)
	}
	if !strings.Contains(output, "Table info test") {
		t.Errorf("expected description in output, got: %s", output)
	}
}
