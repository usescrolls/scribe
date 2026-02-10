package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"

	scribe "github.com/usescrolls/scribe/internal"
)

var (
	checkCmd = &cobra.Command{
		Use:   "check [skill-name]",
		Short: "Check for skill updates",
		Long: `Check if installed skills have updates available.

Without arguments, checks all installed skills.
With a skill name, checks only that skill.

Examples:
  scribe check
  scribe check react-best-practices`,
		Args: cobra.MaximumNArgs(1),
		RunE: runCheck,
	}
)

// CheckResult represents the result of checking a skill for updates
type CheckResult struct {
	Name        string `json:"name"`
	NeedsUpdate bool   `json:"needsUpdate"`
	Error       string `json:"error,omitempty"`
	CurrentHash string `json:"currentHash,omitempty"`
	RemoteHash  string `json:"remoteHash,omitempty"`
}

func runCheck(cmd *cobra.Command, args []string) error {
	var results []CheckResult

	if len(args) == 1 {
		// Check single skill
		result := checkSkill(args[0])
		results = append(results, result)
	} else {
		// Check all skills
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

		if !quiet {
			fmt.Printf("Checking %d skill(s) for updates...\n\n", len(skillNames))
		}

		for _, name := range skillNames {
			result := checkSkill(name)
			results = append(results, result)
		}
	}

	if jsonOutput {
		return checkOutputJSON(results)
	}

	return checkOutputTable(results)
}

func checkSkill(skillName string) CheckResult {
	result := CheckResult{Name: skillName}

	// Read the skill and its metadata
	skill, err := scribe.ReadSkill(skillName)
	if err != nil {
		result.Error = fmt.Sprintf("failed to read skill: %v", err)
		return result
	}

	if skill.Meta == nil {
		result.Error = "no metadata (manually added skill)"
		return result
	}

	result.CurrentHash = skill.Meta.ContentHash

	// Skip local sources - they can't be checked remotely
	if skill.Meta.SourceType == "local" {
		result.Error = "local source (cannot check for updates)"
		return result
	}

	// Reconstruct source info
	source := reconstructSource(skill.Meta)

	// Fetch remote content
	skills, fetchResult, err := scribe.FetchAndDiscoverSkills(source)
	if err != nil {
		errMsg := fmt.Sprintf("failed to fetch: %v", err)
		if scribe.IsAuthError(err) {
			errMsg += " (auth issue — see 'scribe install --help')"
		}
		result.Error = errMsg
		return result
	}
	if fetchResult != nil {
		defer fetchResult.Cleanup()
	}

	// Find the specific skill in fetched content
	var remoteSkill *scribe.Skill
	for _, s := range skills {
		if s.Name == skillName {
			remoteSkill = s
			break
		}
	}

	if remoteSkill == nil {
		result.Error = "skill not found in remote source"
		return result
	}

	// Read remote SKILL.md content and compute hash
	remoteSkillPath := filepath.Join(remoteSkill.Path, scribe.SkillFileName)
	remoteContent, err := os.ReadFile(remoteSkillPath)
	if err != nil {
		result.Error = fmt.Sprintf("failed to read remote skill: %v", err)
		return result
	}

	result.RemoteHash = scribe.ComputeContentHash(string(remoteContent))
	result.NeedsUpdate = result.CurrentHash != result.RemoteHash

	return result
}

// reconstructSource creates a SourceInfo from SkillMeta
func reconstructSource(meta *scribe.SkillMeta) *scribe.SourceInfo {
	return scribe.ReconstructSource(meta)
}

func checkOutputJSON(results []CheckResult) error {
	outdated := 0
	upToDate := 0
	errors := 0

	for _, r := range results {
		switch {
		case r.Error != "":
			errors++
		case r.NeedsUpdate:
			outdated++
		default:
			upToDate++
		}
	}

	output := struct {
		Results []CheckResult `json:"results"`
		Summary struct {
			Total    int `json:"total"`
			Outdated int `json:"outdated"`
			UpToDate int `json:"upToDate"`
			Errors   int `json:"errors"`
		} `json:"summary"`
	}{
		Results: results,
	}
	output.Summary.Total = len(results)
	output.Summary.Outdated = outdated
	output.Summary.UpToDate = upToDate
	output.Summary.Errors = errors

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func checkOutputTable(results []CheckResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tSTATUS\tCURRENT\tREMOTE")

	outdated := 0
	upToDate := 0
	errors := 0

	for _, r := range results {
		status := "up-to-date"
		current := truncateHash(r.CurrentHash)
		remote := truncateHash(r.RemoteHash)

		switch {
		case r.Error != "":
			status = "error: " + r.Error
			current = "-"
			remote = "-"
			errors++
		case r.NeedsUpdate:
			status = "outdated"
			outdated++
		default:
			upToDate++
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.Name, status, current, remote)
	}

	_ = w.Flush()

	if !quiet {
		fmt.Printf("\n%d skill(s) checked: %d outdated, %d up-to-date", len(results), outdated, upToDate)
		if errors > 0 {
			fmt.Printf(", %d error(s)", errors)
		}
		fmt.Println()
	}

	return nil
}

// truncateHash truncates a hash for display
func truncateHash(hash string) string {
	if hash == "" {
		return "-"
	}
	// Show first 20 chars of hash (sha256:abc123...)
	if len(hash) > 20 {
		return hash[:20] + "..."
	}
	return hash
}
