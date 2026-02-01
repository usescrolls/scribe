package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/usescrolls/scribe/internal"
)

var (
	// List flags
	namesOnly bool

	listCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List installed plugins",
		Long: `List installed plugins.

Examples:
  scribe list
  scribe ls --json
  scribe list --names-only`,
		Args: cobra.NoArgs,
		RunE: runList,
	}
)

func init() {
	listCmd.Flags().BoolVar(&namesOnly, "names-only", false, "Print only plugin names, one per line")
}

func runList(cmd *cobra.Command, args []string) error {
	// Initialize server
	if err := initServer(); err != nil {
		scribe.Logger.Error("failed to initialize", "error", err)
		os.Exit(ExitRegistryError)
	}

	plugins := server.GetAllPlugins()

	// Sort by name for consistent output
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Name < plugins[j].Name
	})

	if jsonOutput {
		return listJSON(plugins)
	}

	if namesOnly {
		return listNamesOnly(plugins)
	}

	return listTable(plugins)
}

func listJSON(plugins []scribe.RegistryEntry) error {
	output := struct {
		Plugins []pluginJSON `json:"plugins"`
		Count   int          `json:"count"`
	}{
		Plugins: make([]pluginJSON, 0, len(plugins)),
		Count:   len(plugins),
	}

	for _, p := range plugins {
		output.Plugins = append(output.Plugins, pluginJSON{
			Name:        p.Name,
			Source:      p.Source,
			Version:     p.Version,
			InstalledAt: p.InstalledAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

type pluginJSON struct {
	Name        string             `json:"name"`
	Source      scribe.PluginSource `json:"source"`
	Version     string             `json:"version,omitempty"`
	InstalledAt string             `json:"installedAt"`
}

func listNamesOnly(plugins []scribe.RegistryEntry) error {
	for _, p := range plugins {
		fmt.Println(p.Name)
	}
	return nil
}

func listTable(plugins []scribe.RegistryEntry) error {
	if len(plugins) == 0 {
		if !quiet {
			fmt.Println("No plugins installed")
		}
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tSOURCE\tVERSION\tINSTALLED")

	for _, p := range plugins {
		sourceStr := formatSourceEntry(p.Source)
		version := p.Version
		if version == "" {
			version = "-"
		}
		installed := p.InstalledAt.Format("2006-01-02")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.Name, sourceStr, version, installed)
	}

	w.Flush()

	if !quiet {
		fmt.Printf("\n%d plugin(s) installed\n", len(plugins))
	}

	return nil
}

// formatSourceEntry returns a human-readable source string for table display
func formatSourceEntry(source scribe.PluginSource) string {
	switch source.Source {
	case "github":
		return fmt.Sprintf("github:%s", source.Repo)
	case "url", "git":
		return fmt.Sprintf("url:%s", source.URL)
	case "zip":
		return fmt.Sprintf("zip:%s", source.URL)
	default:
		return source.Source
	}
}
