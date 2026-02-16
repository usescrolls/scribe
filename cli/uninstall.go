package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	scribe "github.com/usescrolls/scribe/internal"
)

var (
	uninstallAll bool

	uninstallCmd = &cobra.Command{
		Use:     "uninstall <skill-name>",
		Aliases: []string{"remove", "rm"},
		Short:   "Remove an installed skill",
		Long: `Remove a skill from canonical storage and all agent directories.

Examples:
  scribe uninstall react-best-practices
  scribe rm typescript-patterns
  scribe uninstall --all`,
		Args: cobra.MaximumNArgs(1),
		RunE: runUninstall,
	}
)

func init() {
	uninstallCmd.Flags().BoolVar(&uninstallAll, "all", false, "Remove all installed skills")
}

func runUninstall(cmd *cobra.Command, args []string) error {
	if uninstallAll {
		return runUninstallAll()
	}

	if len(args) == 0 {
		return fmt.Errorf("skill name is required (or use --all to remove all skills)")
	}

	skillName := args[0]

	// Check if skill exists
	exists, err := scribe.SkillExists(skillName)
	if err != nil {
		return fmt.Errorf("failed to check skill: %w", err)
	}
	if !exists {
		return fmt.Errorf("skill '%s' not found", skillName)
	}

	// Check if it's a system skill
	if scribe.IsSystemSkill(skillName) {
		return fmt.Errorf("cannot uninstall system skill '%s'", skillName)
	}

	if !quiet {
		fmt.Printf("Removing skill '%s'...\n", skillName)
	}

	// Remove from all workspaces
	if err := scribe.RemoveSkillFromAllWorkspaces(skillName); err != nil {
		scribe.Logger.Warn("failed to remove from workspaces", "skill", skillName, "error", err)
	}

	// Uninstall the skill (removes symlinks and canonical copy)
	if err := scribe.UninstallSkill(skillName); err != nil {
		return fmt.Errorf("failed to remove skill: %w", err)
	}

	if !quiet {
		fmt.Printf("Skill '%s' removed successfully\n", skillName)
	}

	return nil
}

func runUninstallAll() error {
	skills, err := scribe.ListInstalledSkills()
	if err != nil {
		return fmt.Errorf("failed to list skills: %w", err)
	}

	if len(skills) == 0 {
		if !quiet {
			fmt.Println("No skills installed")
		}
		return nil
	}

	if !quiet {
		fmt.Printf("Removing %d skill(s)...\n", len(skills))
	}

	for _, skillName := range skills {
		// Skip system skills
		if scribe.IsSystemSkill(skillName) {
			if !quiet {
				fmt.Printf("  Skipping system skill %s\n", skillName)
			}
			continue
		}

		// Remove from all workspaces
		_ = scribe.RemoveSkillFromAllWorkspaces(skillName)

		// Uninstall the skill
		if err := scribe.UninstallSkill(skillName); err != nil {
			fmt.Fprintf(os.Stderr, "  x Failed to remove %s: %v\n", skillName, err)
			continue
		}

		if !quiet {
			fmt.Printf("  Removed %s\n", skillName)
		}
	}

	// Rebuild default workspace (should now be empty)
	_ = scribe.RebuildDefaultWorkspace()

	if !quiet {
		fmt.Println("All skills removed")
	}

	return nil
}
