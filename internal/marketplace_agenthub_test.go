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

// --- GetSkillAudits tests ---

func TestAgentHubMarketplace_GetSkillAudits_ParsesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/plugins/alice/repo/my-skill/audits" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"audits": []map[string]string{
				{"provider": "vibesafe", "label": "VibeSafe Audit", "result": "Pass"},
				{"provider": "snyk", "label": "Snyk", "result": "Warn"},
			},
		})
	}))
	defer server.Close()

	a := &AgentHubMarketplace{baseURL: server.URL}
	result, err := a.GetSkillAudits("alice", "repo", "my-skill")
	if err != nil {
		t.Fatalf("GetSkillAudits returned error: %v", err)
	}
	if len(result.Audits) != 2 {
		t.Fatalf("expected 2 audits, got %d", len(result.Audits))
	}
	if result.Audits[0].Provider != "vibesafe" {
		t.Errorf("audit[0].Provider = %q, want 'vibesafe'", result.Audits[0].Provider)
	}
	if result.Audits[0].Result != "Pass" {
		t.Errorf("audit[0].Result = %q, want 'Pass'", result.Audits[0].Result)
	}
	if result.Audits[1].Result != "Warn" {
		t.Errorf("audit[1].Result = %q, want 'Warn'", result.Audits[1].Result)
	}
}

func TestAgentHubMarketplace_GetSkillAudits_404ReturnsEmpty(t *testing.T) {
	a, cleanup := newTestAgentHubStatus(t, http.StatusNotFound)
	defer cleanup()

	result, err := a.GetSkillAudits("alice", "repo", "nonexistent")
	if err != nil {
		t.Fatalf("expected no error for 404, got: %v", err)
	}
	if len(result.Audits) != 0 {
		t.Errorf("expected empty audits for 404, got %d", len(result.Audits))
	}
}

func TestAgentHubMarketplace_GetSkillAudits_ServerError(t *testing.T) {
	a, cleanup := newTestAgentHubStatus(t, http.StatusInternalServerError)
	defer cleanup()

	_, err := a.GetSkillAudits("alice", "repo", "my-skill")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestAgentHubMarketplace_GetSkillAudits_NullAuditsReturnsEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"audits": null}`))
	}))
	defer server.Close()

	a := &AgentHubMarketplace{baseURL: server.URL}
	result, err := a.GetSkillAudits("alice", "repo", "my-skill")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Audits == nil {
		t.Fatal("expected non-nil Audits slice")
	}
	if len(result.Audits) != 0 {
		t.Errorf("expected 0 audits, got %d", len(result.Audits))
	}
}

func TestAgentHubMarketplace_GetSkillAudits_ParsesAuditDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"audits": []map[string]string{
				{"provider": "agent-trust-hub", "label": "Agent Trust Hub", "result": "Pass"},
				{"provider": "socket", "label": "Socket", "result": "Warn"},
			},
			"auditDetails": []map[string]any{
				{
					"provider":     "agent-trust-hub",
					"result":       "Pass",
					"riskLevel":    "Low",
					"analysisHtml": "<p>No issues found</p>",
					"analyzedAt":   "2026-03-10",
					"alerts":       nil,
					"metadata":     map[string]string{"Version": "1.2.0"},
				},
				{
					"provider":  "socket",
					"result":    "Warn",
					"riskLevel": nil,
					"alerts": []map[string]any{
						{
							"type":        "Obfuscated Code",
							"severity":    "HIGH",
							"file":        "src/index.js",
							"description": "Detected obfuscated code patterns",
							"confidence":  85,
						},
					},
					"metadata": map[string]string{"Dependencies": "12"},
				},
			},
		})
	}))
	defer server.Close()

	a := &AgentHubMarketplace{baseURL: server.URL}
	result, err := a.GetSkillAudits("alice", "repo", "my-skill")
	if err != nil {
		t.Fatalf("GetSkillAudits returned error: %v", err)
	}
	if len(result.Audits) != 2 {
		t.Fatalf("expected 2 audits, got %d", len(result.Audits))
	}
	if len(result.AuditDetails) != 2 {
		t.Fatalf("expected 2 audit details, got %d", len(result.AuditDetails))
	}

	// Check agent-trust-hub detail
	ath := result.AuditDetails[0]
	if ath.Provider != "agent-trust-hub" {
		t.Errorf("detail[0].Provider = %q, want 'agent-trust-hub'", ath.Provider)
	}
	if ath.RiskLevel == nil || *ath.RiskLevel != "Low" {
		t.Errorf("detail[0].RiskLevel = %v, want 'Low'", ath.RiskLevel)
	}
	if ath.AnalysisHTML == nil || *ath.AnalysisHTML != "<p>No issues found</p>" {
		t.Errorf("detail[0].AnalysisHTML = %v, want '<p>No issues found</p>'", ath.AnalysisHTML)
	}
	if ath.Metadata == nil || ath.Metadata["Version"] != "1.2.0" {
		t.Errorf("detail[0].Metadata = %v, want Version=1.2.0", ath.Metadata)
	}

	// Check socket detail with alerts
	sock := result.AuditDetails[1]
	if sock.Provider != "socket" {
		t.Errorf("detail[1].Provider = %q, want 'socket'", sock.Provider)
	}
	if len(sock.Alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(sock.Alerts))
	}
	alert := sock.Alerts[0]
	if alert.Type != "Obfuscated Code" {
		t.Errorf("alert.Type = %q, want 'Obfuscated Code'", alert.Type)
	}
	if alert.Severity != "HIGH" {
		t.Errorf("alert.Severity = %q, want 'HIGH'", alert.Severity)
	}
	if alert.File == nil || *alert.File != "src/index.js" {
		t.Errorf("alert.File = %v, want 'src/index.js'", alert.File)
	}
	if alert.Confidence == nil || *alert.Confidence != 85 {
		t.Errorf("alert.Confidence = %v, want 85", alert.Confidence)
	}
}

func TestAgentHubMarketplace_GetSkillAudits_NullAuditDetailsReturnsEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"audits": [], "auditDetails": null}`))
	}))
	defer server.Close()

	a := &AgentHubMarketplace{baseURL: server.URL}
	result, err := a.GetSkillAudits("alice", "repo", "my-skill")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AuditDetails == nil {
		t.Fatal("expected non-nil AuditDetails slice")
	}
	if len(result.AuditDetails) != 0 {
		t.Errorf("expected 0 audit details, got %d", len(result.AuditDetails))
	}
}

func TestAgentHubMarketplace_GetSkillAudits_MissingAuditDetailsField(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"audits": [{"provider": "vibesafe", "label": "VibeSafe", "result": "Pass"}]}`))
	}))
	defer server.Close()

	a := &AgentHubMarketplace{baseURL: server.URL}
	result, err := a.GetSkillAudits("alice", "repo", "my-skill")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Audits) != 1 {
		t.Errorf("expected 1 audit, got %d", len(result.Audits))
	}
	if result.AuditDetails == nil {
		t.Fatal("expected non-nil AuditDetails slice when field missing")
	}
	if len(result.AuditDetails) != 0 {
		t.Errorf("expected 0 audit details, got %d", len(result.AuditDetails))
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
