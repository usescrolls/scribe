package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	scribe "github.com/usescrolls/scribe/internal"
)

var (
	// Install flags
	installSkills   string
	installListOnly bool
	installYes      bool
	installAll      bool

	installCmd = &cobra.Command{
		Use:   "install <source>",
		Short: "Install skills from a source",
		Long: `Install skills from various sources.

All skills are installed globally and managed by Scribe across all detected
coding agents.

Sources can be:
  owner/repo                    GitHub shorthand
  https://github.com/owner/repo Full GitHub URL
  ./local/path                  Local directory
  https://example.com           Well-known endpoint

Examples:
  scribe install vercel-labs/agent-skills
  scribe install https://github.com/owner/repo
  scribe install ./my-skills
  scribe install owner/repo --list`,
		Args: cobra.ExactArgs(1),
		RunE: runInstall,
	}
)

func init() {
	installCmd.Flags().StringVarP(&installSkills, "skill", "s", "", "Select specific skills to install (comma-separated)")
	installCmd.Flags().BoolVarP(&installListOnly, "list", "l", false, "List available skills without installing")
	installCmd.Flags().BoolVarP(&installYes, "yes", "y", false, "Skip interactive prompts")
	installCmd.Flags().BoolVar(&installAll, "all", false, "Install all skills to all detected agents")
}

func runInstall(cmd *cobra.Command, args []string) error {
	sourceArg := args[0]

	// Parse the source
	source, err := scribe.ParseSourceString(sourceArg)
	if err != nil {
		return fmt.Errorf("invalid source: %w", err)
	}

	// Create CLI progress emitter
	emit := func(e scribe.ProgressEvent) {
		if !quiet {
			fmt.Printf("  %s\n", e.Message)
		}
	}

	if !quiet {
		fmt.Printf("Fetching skills from %s...\n", formatSourceInfo(source))
	}

	// Fetch and discover skills
	skills, fetchResult, err := scribe.FetchAndDiscoverSkills(source, emit)
	if err != nil {
		if scribe.IsAuthError(err) {
			fmt.Fprintf(os.Stderr, "Hint: %s\n", scribe.AuthHintMessage())
		}
		return fmt.Errorf("failed to fetch skills: %w", err)
	}
	if fetchResult != nil {
		defer fetchResult.Cleanup()
	}

	if len(skills) == 0 {
		return fmt.Errorf("no skills found in source")
	}

	// List only mode
	if installListOnly {
		fmt.Printf("\nFound %d skill(s):\n", len(skills))
		for _, skill := range skills {
			fmt.Printf("  - %s - %s\n", skill.Name, skill.Description)
		}
		return nil
	}

	// Filter skills if --skill flag is provided
	if installSkills != "" {
		skillNames := strings.Split(installSkills, ",")
		skills = scribe.FilterSkillsByName(skills, skillNames)
		if len(skills) == 0 {
			return fmt.Errorf("no matching skills found for: %s", installSkills)
		}
	}

	// Detect installed agents
	installedAgents := scribe.DetectInstalledAgents()
	if len(installedAgents) == 0 {
		return fmt.Errorf("no coding agents detected. Please install at least one agent (Claude Code, Cursor, etc.)")
	}

	if !quiet {
		fmt.Printf("\nFound %d skill(s) to install:\n", len(skills))
		for _, skill := range skills {
			fmt.Printf("  - %s - %s\n", skill.Name, skill.Description)
		}

		fmt.Printf("\nDetected %d agent(s):\n", len(installedAgents))
		for _, agent := range installedAgents {
			fmt.Printf("  - %s\n", agent.DisplayName)
		}
		fmt.Println()
	}

	// Ensure Scribe directories exist
	if err := scribe.EnsureScribeDirs(); err != nil {
		return fmt.Errorf("failed to create Scribe directories: %w", err)
	}

	// Ensure default workspace exists
	if err := scribe.EnsureDefaultWorkspace(); err != nil {
		scribe.Logger.Warn("failed to ensure default workspace", "error", err)
	}

	// Extract git commit info from fetched repo
	gitInfo := scribe.GetHeadCommitInfo(fetchResult.ContentDir)

	// Install each skill
	opts := scribe.InstallOptions{
		Yes:       installYes,
		IsPrivate: fetchResult.IsPrivate,
	}

	successCount := 0
	for i, skill := range skills {
		emit(scribe.ProgressEvent{
			Phase:   "install",
			Step:    "start",
			Message: fmt.Sprintf("Installing %s (%d/%d)...", skill.Name, i+1, len(skills)),
			Detail:  skill.Name,
		})

		if err := scribe.InstallSkill(skill, source, opts, gitInfo, emit); err != nil {
			fmt.Fprintf(os.Stderr, "  x Failed to install %s: %v\n", skill.Name, err)
			continue
		}

		// Add to workspaces
		if err := scribe.AddSkillToActiveAndDefaultWorkspace(skill.Name); err != nil {
			scribe.Logger.Warn("failed to add to workspace", "skill", skill.Name, "error", err)
		}

		emit(scribe.ProgressEvent{
			Phase:   "install",
			Step:    "done",
			Message: fmt.Sprintf("Installed %s", skill.Name),
			Detail:  skill.Name,
		})
		successCount++
	}

	if !quiet {
		fmt.Printf("\nInstalled %d/%d skill(s)\n", successCount, len(skills))
	}

	return nil
}
