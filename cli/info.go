package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	scribe "gitlab.com/usescrolls/scribe/internal"
)

var (
	infoCmd = &cobra.Command{
		Use:   "info <name>",
		Short: "Show detailed information about an installed skill",
		Long: `Show detailed information about an installed skill.

Examples:
  scribe info react-best-practices
  scribe info typescript-patterns --json`,
		Args: cobra.ExactArgs(1),
		RunE: runInfo,
	}
)

func runInfo(cmd *cobra.Command, args []string) error {
	name := args[0]

	skill, err := scribe.ReadSkill(name)
	if err != nil {
		if !quiet {
			fmt.Fprintf(os.Stderr, "Skill '%s' not found\n", name)
		}
		os.Exit(ExitNotFound)
	}

	agents := scribe.GetAgentsWithSkill(name)

	if jsonOutput {
		return infoJSON(skill, agents)
	}

	return infoTable(skill, agents)
}

type skillDetailJSON struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Source      string         `json:"source,omitempty"`
	SourceType  string         `json:"sourceType,omitempty"`
	SourceURL   string         `json:"sourceUrl,omitempty"`
	SkillPath   string         `json:"skillPath,omitempty"`
	ContentHash string         `json:"contentHash,omitempty"`
	InstalledAt string         `json:"installedAt,omitempty"`
	UpdatedAt   string         `json:"updatedAt,omitempty"`
	Agents      []string       `json:"agents"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

func infoJSON(skill *scribe.Skill, agents []string) error {
	output := skillDetailJSON{
		Name:        skill.Name,
		Description: skill.Description,
		Agents:      agents,
		Metadata:    skill.Metadata,
	}

	if skill.Meta != nil {
		output.Source = skill.Meta.Source
		output.SourceType = skill.Meta.SourceType
		output.SourceURL = skill.Meta.SourceURL
		output.SkillPath = skill.Meta.SkillPath
		output.ContentHash = skill.Meta.ContentHash
		output.InstalledAt = skill.Meta.InstalledAt
		output.UpdatedAt = skill.Meta.UpdatedAt
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func infoTable(skill *scribe.Skill, agents []string) error {
	fmt.Printf("Name:         %s\n", skill.Name)
	fmt.Printf("Description:  %s\n", skill.Description)

	if skill.Meta != nil {
		if skill.Meta.Source != "" {
			fmt.Printf("Source:       %s\n", skill.Meta.Source)
		}
		if skill.Meta.SourceType != "" {
			fmt.Printf("Source Type:  %s\n", skill.Meta.SourceType)
		}
		if skill.Meta.SourceURL != "" {
			fmt.Printf("Source URL:   %s\n", skill.Meta.SourceURL)
		}
		if skill.Meta.InstalledAt != "" {
			fmt.Printf("Installed:    %s\n", formatInstalledAt(skill.Meta.InstalledAt))
		}
		if skill.Meta.UpdatedAt != "" {
			fmt.Printf("Updated:      %s\n", formatInstalledAt(skill.Meta.UpdatedAt))
		}
		if skill.Meta.ContentHash != "" {
			fmt.Printf("Content Hash: %s\n", skill.Meta.ContentHash)
		}
	}

	if len(agents) > 0 {
		fmt.Printf("Agents:       %s\n", strings.Join(agents, ", "))
	} else {
		fmt.Printf("Agents:       (none)\n")
	}

	return nil
}
