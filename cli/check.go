package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	scribe "gitlab.com/usescrolls/scribe/internal"
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

func runCheck(cmd *cobra.Command, args []string) error {
	var results []scribe.CheckResult

	if len(args) == 1 {
		// Check single skill
		result := scribe.CheckSkillForUpdate(args[0])
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
			result := scribe.CheckSkillForUpdate(name)
			results = append(results, result)
		}
	}

	if jsonOutput {
		return checkOutputJSON(results)
	}

	return checkOutputTable(results)
}

func checkOutputJSON(results []scribe.CheckResult) error {
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
		Results []scribe.CheckResult `json:"results"`
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

func checkOutputTable(results []scribe.CheckResult) error {
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
