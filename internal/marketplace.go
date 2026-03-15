package scribe

import "fmt"

// MarketplaceProvider defines the interface for marketplace search providers.
// Implement this interface to add support for additional providers (GitLab, Bitbucket, etc.).
type MarketplaceProvider interface {
	ID() string
	DisplayName() string
	Search(query string, page int) (*MarketplaceResult, error)
}

// MarketplaceRepo represents a repository found in the marketplace
type MarketplaceRepo struct {
	Owner       string `json:"owner"`
	Name        string `json:"name"`
	FullName    string `json:"fullName"`
	Description string `json:"description"`
	URL         string `json:"url"`
	AvatarURL   string `json:"avatarUrl"`
	Stars       int    `json:"stars"`
	SkillCount  int    `json:"skillCount"`
	Provider    string `json:"provider"`
	Downloads   int    `json:"downloads,omitempty"`
	Verified    bool   `json:"verified,omitempty"`
	Category    string `json:"category,omitempty"`
}

// MarketplaceResult holds the search results from a marketplace provider
type MarketplaceResult struct {
	Repos      []MarketplaceRepo `json:"repos"`
	TotalCount int               `json:"totalCount"`
}

// SkillAudit represents a single security audit result for a marketplace skill
type SkillAudit struct {
	Provider string `json:"provider"` // Audit tool provider name (e.g., "vibesafe")
	Label    string `json:"label"`    // Human-readable audit label
	Result   string `json:"result"`   // Result status: "Pass", "Warn", "Fail", etc.
}

// AuditAlert represents a single alert from a security audit provider
type AuditAlert struct {
	Type        string  `json:"type"`
	Severity    string  `json:"severity"`
	File        *string `json:"file"`
	Description string  `json:"description"`
	Confidence  *int    `json:"confidence"`
}

// AuditDetail contains detailed security audit information for a specific provider
type AuditDetail struct {
	Provider     string            `json:"provider"`
	Result       string            `json:"result"`
	RiskLevel    *string           `json:"riskLevel"`
	AnalysisHTML *string           `json:"analysisHtml"`
	AnalyzedAt   *string           `json:"analyzedAt"`
	Alerts       []AuditAlert      `json:"alerts"`
	Metadata     map[string]string `json:"metadata"`
}

// SkillAuditResult is the response from fetching audits for a skill
type SkillAuditResult struct {
	Audits       []SkillAudit  `json:"audits"`
	AuditDetails []AuditDetail `json:"auditDetails"`
}

// MarketplaceProviderInfo is the frontend-friendly representation of a provider
type MarketplaceProviderInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

// provider registry
var marketplaceProviders = map[string]MarketplaceProvider{}

func init() {
	registerMarketplaceProvider(&AgentHubMarketplace{})
	registerMarketplaceProvider(&GitHubMarketplace{})
}

func registerMarketplaceProvider(p MarketplaceProvider) {
	marketplaceProviders[p.ID()] = p
}

// GetMarketplaceProviders returns info about all registered providers
func GetMarketplaceProviders() []MarketplaceProviderInfo {
	out := make([]MarketplaceProviderInfo, 0, len(marketplaceProviders))
	for _, p := range marketplaceProviders {
		out = append(out, MarketplaceProviderInfo{
			ID:          p.ID(),
			DisplayName: p.DisplayName(),
		})
	}
	return out
}

// SearchMarketplace delegates a search to the specified provider
func SearchMarketplace(providerID, query string, page int) (*MarketplaceResult, error) {
	p, ok := marketplaceProviders[providerID]
	if !ok {
		return nil, fmt.Errorf("unknown marketplace provider: %s", providerID)
	}
	return p.Search(query, page)
}

// GetSkillAudits fetches vulnerability audits for an AgentHub skill.
func GetSkillAudits(authorName, repoSlug, name string) (*SkillAuditResult, error) {
	p, ok := marketplaceProviders["agenthub"]
	if !ok {
		return nil, fmt.Errorf("agenthub provider not registered")
	}
	ah, ok := p.(*AgentHubMarketplace)
	if !ok {
		return nil, fmt.Errorf("agenthub provider has unexpected type")
	}
	return ah.GetSkillAudits(authorName, repoSlug, name)
}
