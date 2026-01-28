package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/usescrolls/scribe/internal/scribe"
)

var (
	infoCmd = &cobra.Command{
		Use:   "info <name>",
		Short: "Show detailed information about an installed plugin",
		Long: `Show detailed information about an installed plugin.

Examples:
  scribe info prettier`,
		Args: cobra.ExactArgs(1),
		RunE: runInfo,
	}
)

func runInfo(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Initialize server
	if err := initServer(); err != nil {
		scribe.Logger.Error("failed to initialize", "error", err)
		os.Exit(ExitRegistryError)
	}

	entry, exists := server.GetRegistryEntry(name)
	if !exists {
		if !quiet {
			fmt.Fprintf(os.Stderr, "Plugin '%s' not found\n", name)
		}
		os.Exit(ExitNotFound)
	}

	// Format output
	sourceStr := formatSourceEntry(entry.Source)

	fmt.Printf("Name:        %s\n", entry.Name)
	fmt.Printf("Source:      %s\n", sourceStr)

	if entry.Version != "" {
		fmt.Printf("Version:     %s\n", entry.Version)
	}

	if entry.Category != "" {
		fmt.Printf("Category:    %s\n", entry.Category)
	}

	if entry.Description != "" {
		fmt.Printf("Description: %s\n", entry.Description)
	}

	if entry.Author != nil && entry.Author.Name != "" {
		fmt.Printf("Author:      %s\n", entry.Author.Name)
	}

	fmt.Printf("Installed:   %s\n", entry.InstalledAt.Format("2006-01-02 15:04:05"))

	if len(entry.Tags) > 0 {
		fmt.Printf("Tags:        %s\n", strings.Join(entry.Tags, ", "))
	}

	return nil
}
