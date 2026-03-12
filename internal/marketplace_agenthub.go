package scribe

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// AgentHubMarketplace implements MarketplaceProvider using the useScrolls AgentHub API.
// Each result is a single skill (unlike GitHub which returns repos with potentially many skills).
type AgentHubMarketplace struct {
	// baseURL overrides the API base for testing. Empty = production.
	baseURL string
}

func (a *AgentHubMarketplace) ID() string          { return "agenthub" }
func (a *AgentHubMarketplace) DisplayName() string { return "skills.sh" }

const agentHubPageSize = 30

func (a *AgentHubMarketplace) Search(query string, page int) (*MarketplaceResult, error) {
	if page < 1 {
		page = 1
	}

	base := a.baseURL
	if base == "" {
		base = "https://usescrolls.com/api"
	}

	offset := (page - 1) * agentHubPageSize
	trimmed := strings.TrimSpace(query)

	apiURL := fmt.Sprintf("%s/plugins?limit=%d&offset=%d", base, agentHubPageSize, offset)
	if trimmed != "" {
		apiURL += "&search=" + url.QueryEscape(trimmed)
	}

	req, err := http.NewRequest("GET", apiURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Scribe-Skills-Manager")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("AgentHub API request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("AgentHub API rate limit exceeded. Please wait and try again")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AgentHub API returned status %d", resp.StatusCode)
	}

	var ahResp agentHubPluginsResponse
	if err := json.NewDecoder(resp.Body).Decode(&ahResp); err != nil {
		return nil, fmt.Errorf("failed to parse AgentHub response: %w", err)
	}

	repos := make([]MarketplaceRepo, 0, len(ahResp.Plugins))
	for _, p := range ahResp.Plugins {
		if p.RepoSlug == "" || p.Author == "" {
			continue
		}
		repos = append(repos, MarketplaceRepo{
			Owner:       p.Author,
			Name:        p.Name,
			FullName:    p.Author + "/" + p.RepoSlug,
			Description: p.Description,
			URL:         "https://github.com/" + p.Author + "/" + p.RepoSlug,
			AvatarURL:   "https://github.com/" + p.Author + ".png",
			Stars:       0,
			SkillCount:  1,
			Provider:    "agenthub",
			Downloads:   p.Downloads,
			Verified:    p.Verified,
			Category:    p.Category,
		})
	}

	return &MarketplaceResult{
		Repos:      repos,
		TotalCount: ahResp.Total,
	}, nil
}

// GetSkillAudits fetches vulnerability audit results for a specific AgentHub skill.
func (a *AgentHubMarketplace) GetSkillAudits(authorName, repoSlug, name string) (*SkillAuditResult, error) {
	base := a.baseURL
	if base == "" {
		base = "https://usescrolls.com/api"
	}

	apiURL := fmt.Sprintf("%s/plugins/%s/%s/%s/audits",
		base,
		url.PathEscape(authorName),
		url.PathEscape(repoSlug),
		url.PathEscape(name),
	)

	req, err := http.NewRequest("GET", apiURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to build audits request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Scribe-Skills-Manager")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("AgentHub audits request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return &SkillAuditResult{Audits: []SkillAudit{}}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AgentHub audits API returned status %d", resp.StatusCode)
	}

	var auditsResp struct {
		Audits []SkillAudit `json:"audits"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&auditsResp); err != nil {
		return nil, fmt.Errorf("failed to parse audits response: %w", err)
	}

	if auditsResp.Audits == nil {
		auditsResp.Audits = []SkillAudit{}
	}

	return &SkillAuditResult{Audits: auditsResp.Audits}, nil
}

// AgentHub API response types (unexported)

type agentHubPluginsResponse struct {
	Plugins []agentHubPlugin `json:"plugins"`
	HasMore bool             `json:"hasMore"`
	Total   int              `json:"total"`
}

type agentHubPlugin struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Author      string `json:"author"`
	RepoSlug    string `json:"repoSlug"`
	Category    string `json:"category"`
	Downloads   int    `json:"downloads"`
	Verified    bool   `json:"verified"`
	Icon        string `json:"icon"`
}
