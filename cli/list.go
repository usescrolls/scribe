package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	scribe "github.com/usescrolls/scribe/internal"
)

var (
	// List flags
	namesOnly bool

	listCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List installed skills",
		Long: `List installed skills.

Examples:
  scribe list
  scribe ls --json
  scribe list --names-only`,
		Args: cobra.NoArgs,
		RunE: runList,
	}
)

func init() {
	listCmd.Flags().BoolVar(&namesOnly, "names-only", false, "Print only skill names, one per line")
}

func runList(cmd *cobra.Command, args []string) error {
	skills, err := scribe.GetAllSkillInfo()
	if err != nil {
		scribe.Logger.Error("failed to get skills", "error", err)
		os.Exit(ExitRegistryError)
	}

	// Sort by name for consistent output
	sort.Slice(skills, func(i, j int) bool {
		return skills[i].Name < skills[j].Name
	})

	if jsonOutput {
		return listSkillsJSON(skills)
	}

	if namesOnly {
		return listSkillsNamesOnly(skills)
	}

	return listSkillsTable(skills)
}

type skillJSON struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Source      string   `json:"source"`
	SourceType  string   `json:"sourceType"`
	InstalledAt string   `json:"installedAt"`
	Agents      []string `json:"agents"`
}

func listSkillsJSON(skills []scribe.SkillInfo) error {
	output := struct {
		Skills []skillJSON `json:"skills"`
		Count  int         `json:"count"`
	}{
		Skills: make([]skillJSON, 0, len(skills)),
		Count:  len(skills),
	}

	for _, s := range skills {
		output.Skills = append(output.Skills, skillJSON{
			Name:        s.Name,
			Description: s.Description,
			Source:      s.Source,
			SourceType:  s.SourceType,
			InstalledAt: s.InstalledAt,
			Agents:      s.Agents,
		})
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func listSkillsNamesOnly(skills []scribe.SkillInfo) error {
	for _, s := range skills {
		fmt.Println(s.Name)
	}
	return nil
}

func listSkillsTable(skills []scribe.SkillInfo) error {
	if len(skills) == 0 {
		if !quiet {
			fmt.Println("No skills installed")
		}
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tDESCRIPTION\tSOURCE\tINSTALLED\tAGENTS")

	for _, s := range skills {
		desc := truncateString(s.Description, 40)
		source := formatSkillSource(s)
		installed := formatInstalledAt(s.InstalledAt)
		agents := strings.Join(s.Agents, ", ")
		if agents == "" {
			agents = "-"
		}
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", s.Name, desc, source, installed, agents)
	}

	_ = w.Flush()

	if !quiet {
		fmt.Printf("\n%d skill(s) installed\n", len(skills))
	}

	return nil
}
