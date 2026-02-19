package scribe

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ============================================================================
// Provider registry (marketplace.go)
// ============================================================================

func TestGetMarketplaceProviders_ContainsGitHub(t *testing.T) {
	providers := GetMarketplaceProviders()

	found := false
	for _, p := range providers {
		if p.ID == "github" {
			found = true
			if p.DisplayName != "GitHub" {
				t.Errorf("GitHub provider DisplayName = %q, want 'GitHub'", p.DisplayName)
			}
		}
	}
	if !found {
		t.Fatal("GetMarketplaceProviders() does not contain 'github'")
	}
}

func TestSearchMarketplace_UnknownProvider(t *testing.T) {
	_, err := SearchMarketplace("nonexistent", "test", 1)
	if err == nil {
		t.Fatal("SearchMarketplace with unknown provider should return error")
	}
}

func TestSearchMarketplace_DelegatesToProvider(t *testing.T) {
	result, err := SearchMarketplace("github", "", 1)
	if err != nil {
		t.Fatalf("SearchMarketplace with empty query returned error: %v", err)
	}
	if len(result.Repos) != 0 {
		t.Errorf("expected 0 repos for empty query, got %d", len(result.Repos))
	}
}

// ============================================================================
// GitHubMarketplace (marketplace_github.go)
// ============================================================================

func TestGitHubMarketplace_IDAndDisplayName(t *testing.T) {
	g := &GitHubMarketplace{}
	if g.ID() != "github" {
		t.Errorf("ID() = %q, want 'github'", g.ID())
	}
	if g.DisplayName() != "GitHub" {
		t.Errorf("DisplayName() = %q, want 'GitHub'", g.DisplayName())
	}
}

func TestGitHubMarketplace_EmptyQuery(t *testing.T) {
	g := &GitHubMarketplace{}
	result, err := g.Search("", 1)
	if err != nil {
		t.Fatalf("empty query should not error: %v", err)
	}
	if result == nil || len(result.Repos) != 0 {
		t.Error("empty query should return empty result")
	}
}

func TestGitHubMarketplace_WhitespaceQuery(t *testing.T) {
	g := &GitHubMarketplace{}
	result, err := g.Search("   ", 1)
	if err != nil {
		t.Fatalf("whitespace query should not error: %v", err)
	}
	if result == nil || len(result.Repos) != 0 {
		t.Error("whitespace query should return empty result")
	}
}

// ============================================================================
// Search() via httptest server
// ============================================================================

func makeRepoItem(fullName, owner, name, desc string, stars int) githubRepoItem {
	return githubRepoItem{
		FullName:    fullName,
		Name:        name,
		Description: desc,
		HTMLURL:     "https://github.com/" + fullName,
		Stars:       stars,
		Owner:       githubOwnerInfo{Login: owner},
	}
}

func newTestGitHub(t *testing.T, resp githubRepoSearchResponse) (provider *GitHubMarketplace, cleanup func()) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	return &GitHubMarketplace{baseURL: server.URL}, server.Close
}

func newTestGitHubStatus(t *testing.T, statusCode int) (provider *GitHubMarketplace, cleanup func()) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
	}))
	return &GitHubMarketplace{baseURL: server.URL}, server.Close
}

func TestGitHubMarketplace_ParsesSingleRepo(t *testing.T) {
	g, cleanup := newTestGitHub(t, githubRepoSearchResponse{
		TotalCount: 1,
		Items:      []githubRepoItem{makeRepoItem("alice/skills", "alice", "skills", "My awesome skills", 42)},
	})
	defer cleanup()

	result, err := g.Search("alice", 1)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(result.Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(result.Repos))
	}
	repo := result.Repos[0]
	if repo.FullName != "alice/skills" {
		t.Errorf("FullName = %q, want 'alice/skills'", repo.FullName)
	}
	if repo.Owner != "alice" {
		t.Errorf("Owner = %q, want 'alice'", repo.Owner)
	}
	if repo.Stars != 42 {
		t.Errorf("Stars = %d, want 42", repo.Stars)
	}
	if repo.SkillCount != 0 {
		t.Errorf("SkillCount = %d, want 0", repo.SkillCount)
	}
	if repo.Provider != "github" {
		t.Errorf("Provider = %q, want 'github'", repo.Provider)
	}
}

func TestGitHubMarketplace_MultipleRepos(t *testing.T) {
	g, cleanup := newTestGitHub(t, githubRepoSearchResponse{
		TotalCount: 3,
		Items: []githubRepoItem{
			makeRepoItem("charlie/first", "charlie", "first", "", 100),
			makeRepoItem("alice/second", "alice", "second", "", 50),
			makeRepoItem("bob/third", "bob", "third", "", 10),
		},
	})
	defer cleanup()

	result, err := g.Search("test", 1)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(result.Repos) != 3 {
		t.Fatalf("expected 3 repos, got %d", len(result.Repos))
	}
	expected := []string{"charlie/first", "alice/second", "bob/third"}
	for i, name := range expected {
		if result.Repos[i].FullName != name {
			t.Errorf("repo[%d] = %q, want %q", i, result.Repos[i].FullName, name)
		}
	}
}

func TestGitHubMarketplace_EmptyResults(t *testing.T) {
	g, cleanup := newTestGitHub(t, githubRepoSearchResponse{TotalCount: 0, Items: []githubRepoItem{}})
	defer cleanup()

	result, err := g.Search("nonexistent", 1)
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

func TestGitHubMarketplace_TotalCount(t *testing.T) {
	g, cleanup := newTestGitHub(t, githubRepoSearchResponse{
		TotalCount: 150,
		Items:      []githubRepoItem{makeRepoItem("alice/skills", "alice", "skills", "", 10)},
	})
	defer cleanup()

	result, err := g.Search("test", 1)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if result.TotalCount != 150 {
		t.Errorf("TotalCount = %d, want 150", result.TotalCount)
	}
}

func TestGitHubMarketplace_DescriptionAndURL(t *testing.T) {
	g, cleanup := newTestGitHub(t, githubRepoSearchResponse{
		TotalCount: 1,
		Items:      []githubRepoItem{makeRepoItem("org/cool-skills", "org", "cool-skills", "A collection of useful skills", 99)},
	})
	defer cleanup()

	result, err := g.Search("cool", 1)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(result.Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(result.Repos))
	}
	repo := result.Repos[0]
	if repo.Description != "A collection of useful skills" {
		t.Errorf("Description = %q, want 'A collection of useful skills'", repo.Description)
	}
	if repo.URL != "https://github.com/org/cool-skills" {
		t.Errorf("URL = %q, want 'https://github.com/org/cool-skills'", repo.URL)
	}
}

func TestGitHubMarketplace_RateLimitReturnsError(t *testing.T) {
	g, cleanup := newTestGitHubStatus(t, http.StatusForbidden)
	defer cleanup()

	_, err := g.Search("test", 1)
	if err == nil {
		t.Fatal("expected error for 403 response")
	}
	if got := err.Error(); got != "GitHub API rate limit exceeded. Please wait a minute and try again" {
		t.Errorf("error = %q, want rate limit message", got)
	}
}

func TestGitHubMarketplace_TooManyRequestsReturnsError(t *testing.T) {
	g, cleanup := newTestGitHubStatus(t, http.StatusTooManyRequests)
	defer cleanup()

	_, err := g.Search("test", 1)
	if err == nil {
		t.Fatal("expected error for 429 response")
	}
	if got := err.Error(); got != "GitHub API rate limit exceeded. Please wait a minute and try again" {
		t.Errorf("error = %q, want rate limit message", got)
	}
}

func TestGitHubMarketplace_ServerErrorReturnsStatus(t *testing.T) {
	g, cleanup := newTestGitHubStatus(t, http.StatusInternalServerError)
	defer cleanup()

	_, err := g.Search("test", 1)
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	if got := err.Error(); got != "GitHub API returned status 500" {
		t.Errorf("error = %q, want 'GitHub API returned status 500'", got)
	}
}

func TestGitHubMarketplace_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	g := &GitHubMarketplace{baseURL: server.URL}
	_, err := g.Search("test", 1)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestGitHubMarketplace_PageClampedToOne(t *testing.T) {
	g, cleanup := newTestGitHub(t, githubRepoSearchResponse{TotalCount: 0, Items: []githubRepoItem{}})
	defer cleanup()

	// page=0 should not error (gets clamped to 1 internally)
	result, err := g.Search("test", 0)
	if err != nil {
		t.Fatalf("Search with page=0 returned error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}
