package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	scribe "gitlab.com/usescrolls/scribe/internal"
)

var (
	updateForce bool

	updateCmd = &cobra.Command{
		Use:   "update [skill-name]",
		Short: "Update installed skills",
		Long: `Update skills to their latest versions.

Without arguments, updates all outdated skills.
With a skill name, updates only that skill.

Examples:
  scribe update
  scribe update react-best-practices
  scribe update --force`,
		Args: cobra.MaximumNArgs(1),
		RunE: runUpdate,
	}
)

func init() {
	updateCmd.Flags().BoolVarP(&updateForce, "force", "f", false, "Force update even if up-to-date")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	var skillsToUpdate []string

	if len(args) == 1 {
		// Update single skill
		skillsToUpdate = []string{args[0]}
	} else {
		// Find all outdated skills
		skillNames, err := scribe.ListInstalledSkills()
		if err != nil {
			return fmt.Errorf("failed to list skills: %w", err)
		}

		if len(skillNames) == 0 {
			if !quiet {
				fmt.Println("No skills installed")
			}
			return nil
		}

		if updateForce {
			// Force update all skills
			skillsToUpdate = skillNames
		} else {
			// Check which skills need updates
			if !quiet {
				fmt.Printf("Checking %d skill(s) for updates...\n", len(skillNames))
			}

			for _, name := range skillNames {
				result := scribe.CheckSkillForUpdate(name)
				if result.Error == "" && result.NeedsUpdate {
					skillsToUpdate = append(skillsToUpdate, name)
				}
			}

			if len(skillsToUpdate) == 0 {
				if !quiet {
					fmt.Println("All skills are up-to-date")
				}
				return nil
			}
		}
	}

	if !quiet {
		fmt.Printf("\nUpdating %d skill(s)...\n\n", len(skillsToUpdate))
	}

	successCount := 0
	for _, skillName := range skillsToUpdate {
		if err := updateSkill(skillName, updateForce); err != nil {
			fmt.Fprintf(os.Stderr, "  x Failed to update %s: %v\n", skillName, err)
			continue
		}
		successCount++
	}

	if !quiet {
		fmt.Printf("\n%d/%d skill(s) updated\n", successCount, len(skillsToUpdate))

		// Check for new available skills across all sources
		sourceResults := scribe.CheckAllSourcesForUpdates()
		for source, r := range sourceResults {
			if len(r.NewAvailableSkills) > 0 {
				fmt.Printf("\n%d other skill(s) available from %s — run `scribe install %s` to review\n", len(r.NewAvailableSkills), source, source)
			}
		}
	}

	return nil
}

func updateSkill(skillName string, force bool) error {
	if !quiet {
		// Show source info before updating
		skill, err := scribe.ReadSkill(skillName)
		if err == nil && skill.Meta != nil && skill.Meta.SourceType != "local" {
			source := scribe.ReconstructSource(skill.Meta)
			fmt.Printf("Updating %s from %s...\n", skillName, formatSourceInfo(source))
		}
	}

	result, err := scribe.UpdateSkill(skillName, force)
	if err != nil {
		if scribe.IsAuthError(err) {
			fmt.Fprintf(os.Stderr, "Hint: %s\n", scribe.AuthHintMessage())
		}
		return err
	}

	if !quiet {
		if !result.Updated {
			fmt.Printf("  %s is already up-to-date\n", skillName)
			return nil
		}
		fmt.Printf("  %s -> %s\n", truncateHash(result.OldHash), truncateHash(result.NewHash))
		agents := scribe.DetectInstalledAgents()
		if len(agents) > 0 {
			fmt.Printf("  Synced to: ")
			for i, a := range agents {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Print(a.ID)
			}
			fmt.Println()
		}
	}

	return nil
}
