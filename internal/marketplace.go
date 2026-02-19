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
	Stars       int    `json:"stars"`
	SkillCount  int    `json:"skillCount"`
	Provider    string `json:"provider"`
}

// MarketplaceResult holds the search results from a marketplace provider
type MarketplaceResult struct {
	Repos      []MarketplaceRepo `json:"repos"`
	TotalCount int               `json:"totalCount"`
}

// MarketplaceProviderInfo is the frontend-friendly representation of a provider
type MarketplaceProviderInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

// provider registry
var marketplaceProviders = map[string]MarketplaceProvider{}

func init() {
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
