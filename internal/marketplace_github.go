package scribe

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GitHubMarketplace implements MarketplaceProvider using the GitHub Repository Search API.
// Code Search requires authentication, so we use Repository Search (unauthenticated, 10 req/min).
type GitHubMarketplace struct {
	// baseURL overrides the GitHub API base for testing. Empty = production.
	baseURL string
}

func (g *GitHubMarketplace) ID() string          { return "github" }
func (g *GitHubMarketplace) DisplayName() string { return "GitHub" }

func (g *GitHubMarketplace) Search(query string, page int) (*MarketplaceResult, error) {
	if page < 1 {
		page = 1
	}

	base := g.baseURL
	if base == "" {
		base = "https://api.github.com"
	}

	// Use Repository Search API (works unauthenticated, 10 req/min).
	// Code Search would let us filter by filename:SKILL.md but requires auth.
	// We append "skill" and qualifiers to bias results toward skill repos.
	// When query is empty, we just search with qualifiers to get popular skill repos.
	trimmed := strings.TrimSpace(query)
	enhancedQuery := "skill archived:false fork:false"
	if trimmed != "" {
		enhancedQuery = fmt.Sprintf("%s skill archived:false fork:false", trimmed)
	}
	apiURL := fmt.Sprintf(
		"%s/search/repositories?q=%s&per_page=30&page=%d&sort=stars&order=desc",
		base, url.QueryEscape(enhancedQuery), page,
	)

	req, err := http.NewRequest("GET", apiURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "Scribe-Skills-Manager")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GitHub API request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("GitHub API rate limit exceeded. Please wait a minute and try again")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var ghResp githubRepoSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&ghResp); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub response: %w", err)
	}

	repos := make([]MarketplaceRepo, 0, len(ghResp.Items))
	for _, item := range ghResp.Items {
		repos = append(repos, MarketplaceRepo{
			Owner:       item.Owner.Login,
			Name:        item.Name,
			FullName:    item.FullName,
			Description: item.Description,
			URL:         item.HTMLURL,
			AvatarURL:   item.Owner.AvatarURL,
			Stars:       item.Stars,
			SkillCount:  0, // unknown until install discovers SKILL.md files
			Provider:    "github",
		})
	}

	return &MarketplaceResult{
		Repos:      repos,
		TotalCount: ghResp.TotalCount,
	}, nil
}

// GitHub Repository Search API response types (unexported)

type githubRepoSearchResponse struct {
	TotalCount int              `json:"total_count"`
	Items      []githubRepoItem `json:"items"`
}

type githubRepoItem struct {
	FullName    string          `json:"full_name"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	HTMLURL     string          `json:"html_url"`
	Stars       int             `json:"stargazers_count"`
	Owner       githubOwnerInfo `json:"owner"`
}

type githubOwnerInfo struct {
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
}

// GetRepoReadme fetches the README.md content for a GitHub repository.
// Uses the GitHub Readme API with raw media type to get markdown directly.
func GetRepoReadme(owner, repo string) (string, error) {
	return getRepoReadmeFrom("https://api.github.com", owner, repo)
}

func getRepoReadmeFrom(baseURL, owner, repo string) (string, error) {
	apiURL := fmt.Sprintf("%s/repos/%s/%s/readme",
		baseURL, url.PathEscape(owner), url.PathEscape(repo))

	req, err := http.NewRequest("GET", apiURL, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.raw+json")
	req.Header.Set("User-Agent", "Scribe-Skills-Manager")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("GitHub API request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("no README found for %s/%s", owner, repo)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	return string(body), nil
}
