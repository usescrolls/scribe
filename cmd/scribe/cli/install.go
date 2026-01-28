package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/usescrolls/scribe/internal/scribe"
)

var (
	// Install flags
	githubRepo string
	npmPackage string
	gitURL     string
	zipURL     string
	ref        string
	noEnable   bool

	installCmd = &cobra.Command{
		Use:   "install [name]",
		Short: "Install a plugin from a source",
		Long: `Install a plugin from a source.

Exactly one source flag must be specified: --github, --npm, --url, or --zip.

Examples:
  scribe install prettier --github usescrolls/prettier-skill
  scribe install eslint --npm @anthropic/claude-eslint
  scribe install custom --url https://github.com/user/plugin.git
  scribe install tool --zip https://example.com/plugin.zip
  scribe install prettier --github usescrolls/prettier-skill --ref v1.0.0`,
		Args: cobra.ExactArgs(1),
		RunE: runInstall,
	}
)

func init() {
	installCmd.Flags().StringVar(&githubRepo, "github", "", "GitHub repository (owner/repo)")
	installCmd.Flags().StringVar(&npmPackage, "npm", "", "NPM package name")
	installCmd.Flags().StringVar(&gitURL, "url", "", "Git URL")
	installCmd.Flags().StringVar(&zipURL, "zip", "", "Zip file URL")
	installCmd.Flags().StringVar(&ref, "ref", "", "Branch or tag reference")
	installCmd.Flags().BoolVar(&noEnable, "no-enable", false, "Don't auto-enable in Claude settings")
}

func runInstall(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Validate exactly one source flag is set
	sourceCount := 0
	var sourceType string
	if githubRepo != "" {
		sourceCount++
		sourceType = "github"
	}
	if npmPackage != "" {
		sourceCount++
		sourceType = "npm"
	}
	if gitURL != "" {
		sourceCount++
		sourceType = "url"
	}
	if zipURL != "" {
		sourceCount++
		sourceType = "zip"
	}

	if sourceCount == 0 {
		return fmt.Errorf("exactly one source flag is required: --github, --npm, --url, or --zip")
	}
	if sourceCount > 1 {
		return fmt.Errorf("only one source flag can be specified at a time")
	}

	// Build plugin source
	pluginSource := scribe.PluginSource{
		Source: sourceType,
		Ref:    ref,
	}

	switch sourceType {
	case "github":
		pluginSource.Repo = githubRepo
	case "npm":
		pluginSource.Package = npmPackage
	case "url":
		pluginSource.URL = gitURL
	case "zip":
		pluginSource.URL = zipURL
	}

	// Initialize server
	if err := initServer(); err != nil {
		scribe.Logger.Error("failed to initialize", "error", err)
		os.Exit(ExitRegistryError)
	}

	if !quiet {
		fmt.Printf("Installing plugin '%s' from %s...\n", name, formatSource(pluginSource))
	}

	// Resolve the source
	resolvedSource, err := server.ResolveSource(name, pluginSource)
	if err != nil {
		scribe.Logger.Error("failed to resolve source", "name", name, "error", err)
		os.Exit(ExitSourceFailed)
	}

	// Create registry entry
	entry := scribe.RegistryEntry{
		Name:           name,
		Source:         pluginSource,
		ResolvedSource: resolvedSource,
		InstalledAt:    time.Now(),
	}

	// Add to registry
	server.SetRegistryEntry(name, entry)

	// Persist registry
	if err := server.SaveRegistry(); err != nil {
		scribe.Logger.Error("failed to save registry", "error", err)
		os.Exit(ExitRegistryError)
	}

	// Regenerate marketplace.json
	if err := server.GenerateMarketplace(); err != nil {
		scribe.Logger.Error("failed to regenerate marketplace", "error", err)
		os.Exit(ExitRegistryError)
	}

	// Auto-enable unless --no-enable
	if !noEnable {
		if err := server.UpdateClaudeSettings(name, true); err != nil {
			scribe.Logger.Warn("failed to update claude settings", "name", name, "error", err)
		}
	}

	if !quiet {
		fmt.Printf("Plugin '%s' installed successfully\n", name)
	}

	return nil
}

// formatSource returns a human-readable source string
func formatSource(source scribe.PluginSource) string {
	switch source.Source {
	case "github":
		if source.Ref != "" {
			return fmt.Sprintf("github:%s@%s", source.Repo, source.Ref)
		}
		return fmt.Sprintf("github:%s", source.Repo)
	case "npm":
		return fmt.Sprintf("npm:%s", source.Package)
	case "url":
		return fmt.Sprintf("url:%s", source.URL)
	case "zip":
		return fmt.Sprintf("zip:%s", source.URL)
	default:
		return source.Source
	}
}
