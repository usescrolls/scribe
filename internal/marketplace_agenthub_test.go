package scribe

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAgentHubMarketplace_IDAndDisplayName(t *testing.T) {
	a := &AgentHubMarketplace{}
	if a.ID() != "agenthub" {
		t.Errorf("ID() = %q, want 'agenthub'", a.ID())
	}
	if a.DisplayName() != "skills.sh" {
		t.Errorf("DisplayName() = %q, want 'skills.sh'", a.DisplayName())
	}
}

func makeAgentHubPlugin(name, author, repoSlug, desc, category string, downloads int, verified bool) agentHubPlugin {
	return agentHubPlugin{
		Name:        name,
		Description: desc,
		Author:      author,
		RepoSlug:    repoSlug,
		Category:    category,
		Downloads:   downloads,
		Verified:    verified,
		Icon:        "box",
	}
}

func newTestAgentHub(t *testing.T, resp agentHubPluginsResponse) (provider *AgentHubMarketplace, cleanup func()) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	return &AgentHubMarketplace{baseURL: server.URL}, server.Close
}

func newTestAgentHubStatus(t *testing.T, statusCode int) (provider *AgentHubMarketplace, cleanup func()) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
	}))
	return &AgentHubMarketplace{baseURL: server.URL}, server.Close
}

func TestAgentHubMarketplace_ParsesSinglePlugin(t *testing.T) {
	a, cleanup := newTestAgentHub(t, agentHubPluginsResponse{
		Plugins: []agentHubPlugin{makeAgentHubPlugin("my-skill", "alice", "skills-repo", "A cool skill", "skill", 42, true)},
		HasMore: false,
		Total:   1,
	})
	defer cleanup()

	result, err := a.Search("alice", 1)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(result.Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(result.Repos))
	}
	repo := result.Repos[0]
	if repo.Owner != "alice" {
		t.Errorf("Owner = %q, want 'alice'", repo.Owner)
	}
	if repo.Name != "my-skill" {
		t.Errorf("Name = %q, want 'my-skill'", repo.Name)
	}
	if repo.FullName != "alice/skills-repo" {
		t.Errorf("FullName = %q, want 'alice/skills-repo'", repo.FullName)
	}
	if repo.Description != "A cool skill" {
		t.Errorf("Description = %q, want 'A cool skill'", repo.Description)
	}
	if repo.URL != "https://github.com/alice/skills-repo" {
		t.Errorf("URL = %q, want 'https://github.com/alice/skills-repo'", repo.URL)
	}
	if repo.SkillCount != 1 {
		t.Errorf("SkillCount = %d, want 1", repo.SkillCount)
	}
	if repo.Provider != "agenthub" {
		t.Errorf("Provider = %q, want 'agenthub'", repo.Provider)
	}
	if repo.AvatarURL != "https://github.com/alice.png" {
		t.Errorf("AvatarURL = %q, want 'https://github.com/alice.png'", repo.AvatarURL)
	}
	if repo.Downloads != 42 {
		t.Errorf("Downloads = %d, want 42", repo.Downloads)
	}
	if !repo.Verified {
		t.Error("Verified = false, want true")
	}
	if repo.Category != "skill" {
		t.Errorf("Category = %q, want 'skill'", repo.Category)
	}
}

func TestAgentHubMarketplace_EmptyQueryHitsAPI(t *testing.T) {
	a, cleanup := newTestAgentHub(t, agentHubPluginsResponse{
		Plugins: []agentHubPlugin{makeAgentHubPlugin("test", "bob", "test-repo", "", "skill", 0, false)},
		Total:   1,
	})
	defer cleanup()

	result, err := a.Search("", 1)
	if err != nil {
		t.Fatalf("empty query should not error: %v", err)
	}
	if len(result.Repos) != 1 {
		t.Errorf("expected 1 repo, got %d", len(result.Repos))
	}
}

func TestAgentHubMarketplace_EmptyResults(t *testing.T) {
	a, cleanup := newTestAgentHub(t, agentHubPluginsResponse{Plugins: []agentHubPlugin{}, Total: 0})
	defer cleanup()

	result, err := a.Search("nonexistent", 1)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(result.Repos) != 0 {
		t.Errorf("expected 0 repos, got %d", len(result.Repos))
	}
	if result.TotalCount != 0 {
		t.Errorf("TotalCount = %d, want 0", result.TotalCount)
	}
}

func TestAgentHubMarketplace_PaginationOffset(t *testing.T) {
	var capturedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(agentHubPluginsResponse{Plugins: []agentHubPlugin{}, Total: 0})
	}))
	defer server.Close()

	a := &AgentHubMarketplace{baseURL: server.URL}
	_, _ = a.Search("test", 2)

	if !strings.Contains(capturedPath, "offset=30") {
		t.Errorf("expected offset=30 for page 2, got URL: %s", capturedPath)
	}
	if !strings.Contains(capturedPath, "limit=30") {
		t.Errorf("expected limit=30, got URL: %s", capturedPath)
	}
}

func TestAgentHubMarketplace_PageClampedToOne(t *testing.T) {
	a, cleanup := newTestAgentHub(t, agentHubPluginsResponse{Plugins: []agentHubPlugin{}, Total: 0})
	defer cleanup()

	result, err := a.Search("test", 0)
	if err != nil {
		t.Fatalf("Search with page=0 returned error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestAgentHubMarketplace_SkipsPluginsWithoutRepoSlug(t *testing.T) {
	a, cleanup := newTestAgentHub(t, agentHubPluginsResponse{
		Plugins: []agentHubPlugin{
			makeAgentHubPlugin("valid", "alice", "repo", "Valid", "skill", 0, false),
			{Name: "invalid", Author: "bob", RepoSlug: "", Description: "Missing repo slug"},
			{Name: "also-invalid", Author: "", RepoSlug: "repo", Description: "Missing author"},
		},
		Total: 3,
	})
	defer cleanup()

	result, err := a.Search("", 1)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(result.Repos) != 1 {
		t.Fatalf("expected 1 valid repo, got %d", len(result.Repos))
	}
	if result.Repos[0].Name != "valid" {
		t.Errorf("expected 'valid' plugin, got %q", result.Repos[0].Name)
	}
}

func TestAgentHubMarketplace_RateLimitReturnsError(t *testing.T) {
	a, cleanup := newTestAgentHubStatus(t, http.StatusTooManyRequests)
	defer cleanup()

	_, err := a.Search("test", 1)
	if err == nil {
		t.Fatal("expected error for 429 response")
	}
	if got := err.Error(); got != "AgentHub API rate limit exceeded. Please wait and try again" {
		t.Errorf("error = %q, want rate limit message", got)
	}
}

func TestAgentHubMarketplace_ServerErrorReturnsStatus(t *testing.T) {
	a, cleanup := newTestAgentHubStatus(t, http.StatusInternalServerError)
	defer cleanup()

	_, err := a.Search("test", 1)
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	if got := err.Error(); got != "AgentHub API returned status 500" {
		t.Errorf("error = %q, want 'AgentHub API returned status 500'", got)
	}
}

func TestAgentHubMarketplace_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	a := &AgentHubMarketplace{baseURL: server.URL}
	_, err := a.Search("test", 1)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestGetMarketplaceProviders_ContainsAgentHub(t *testing.T) {
	providers := GetMarketplaceProviders()

	found := false
	for _, p := range providers {
		if p.ID == "agenthub" {
			found = true
			if p.DisplayName != "skills.sh" {
				t.Errorf("AgentHub provider DisplayName = %q, want 'skills.sh'", p.DisplayName)
			}
		}
	}
	if !found {
		t.Fatal("GetMarketplaceProviders() does not contain 'agenthub'")
	}
}
