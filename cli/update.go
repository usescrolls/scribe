package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	scribe "github.com/usescrolls/scribe/internal"
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
				result := checkSkill(name)
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
	}

	return nil
}

func updateSkill(skillName string, force bool) error {
	// Read existing skill
	skill, err := scribe.ReadSkill(skillName)
	if err != nil {
		return fmt.Errorf("skill not found: %w", err)
	}

	if skill.Meta == nil {
		return fmt.Errorf("skill has no metadata, cannot update")
	}

	// Skip local sources
	if skill.Meta.SourceType == "local" {
		return fmt.Errorf("local source, cannot update")
	}

	// Reconstruct source
	source := reconstructSource(skill.Meta)

	if !quiet {
		fmt.Printf("Updating %s from %s...\n", skillName, formatSourceInfo(source))
	}

	// Fetch new content
	skills, fetchResult, err := scribe.FetchAndDiscoverSkills(source)
	if err != nil {
		if scribe.IsAuthError(err) {
			fmt.Fprintf(os.Stderr, "Hint: %s\n", scribe.AuthHintMessage())
		}
		return fmt.Errorf("failed to fetch from source: %w", err)
	}
	if fetchResult != nil {
		defer fetchResult.Cleanup()
	}

	// Find the specific skill
	var newSkill *scribe.Skill
	for _, s := range skills {
		if s.Name == skillName {
			newSkill = s
			break
		}
	}

	if newSkill == nil {
		return fmt.Errorf("skill not found in source")
	}

	// Read new content
	newSkillPath := filepath.Join(newSkill.Path, scribe.SkillFileName)
	newContent, err := os.ReadFile(newSkillPath)
	if err != nil {
		return fmt.Errorf("failed to read new skill content: %w", err)
	}

	// Check if update is needed (unless force)
	skillDir, err := scribe.GetSkillDir(skillName)
	if err != nil {
		return fmt.Errorf("failed to get skill directory: %w", err)
	}

	needsUpdate, err := scribe.SkillNeedsUpdate(skillDir, string(newContent))
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !needsUpdate && !force {
		if !quiet {
			fmt.Printf("  %s is already up-to-date\n", skillName)
		}
		return nil
	}

	oldHash := skill.Meta.ContentHash
	newHash := scribe.ComputeContentHash(string(newContent))

	// Copy new skill content to canonical location
	if err := copySkillDir(newSkill.Path, skillDir); err != nil {
		return fmt.Errorf("failed to update skill files: %w", err)
	}

	// Update metadata
	metaPath, err := scribe.GetMetaPath(skillName)
	if err != nil {
		return fmt.Errorf("failed to get meta path: %w", err)
	}

	meta, err := scribe.ReadSkillMeta(metaPath)
	if err != nil {
		// Create new meta if not found
		meta = scribe.NewSkillMeta(source, skill.Meta.SkillPath, string(newContent), scribe.GetHeadCommitInfo(fetchResult.ContentDir))
	} else {
		scribe.UpdateSkillMeta(meta, string(newContent), scribe.GetHeadCommitInfo(fetchResult.ContentDir))
	}

	if err := scribe.WriteSkillMeta(metaPath, meta); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	// Re-sync to all agents
	agents := scribe.DetectInstalledAgents()
	agentIDs := make([]string, len(agents))
	for i, a := range agents {
		agentIDs[i] = a.ID
	}

	if err := scribe.SyncSkillToAgents(skillName, agentIDs); err != nil {
		scribe.Logger.Warn("failed to sync to agents", "skill", skillName, "error", err)
	}

	if !quiet {
		fmt.Printf("  %s -> %s\n", truncateHash(oldHash), truncateHash(newHash))
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

// copySkillDir copies a skill directory to the canonical location
func copySkillDir(src, dst string) error {
	// Remove existing destination
	if err := os.RemoveAll(dst); err != nil {
		return fmt.Errorf("failed to remove existing: %w", err)
	}

	// Create destination directory
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Copy files
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copySkillDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	if err := os.WriteFile(dst, content, info.Mode()); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
