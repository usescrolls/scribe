package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/usescrolls/scribe/internal"
)

var (
	// Uninstall flags
	uninstallAll bool

	uninstallCmd = &cobra.Command{
		Use:     "uninstall <name>",
		Aliases: []string{"remove", "rm"},
		Short:   "Remove an installed plugin",
		Long: `Remove an installed plugin.

Examples:
  scribe uninstall prettier
  scribe rm prettier
  scribe uninstall --all`,
		Args: func(cmd *cobra.Command, args []string) error {
			if uninstallAll {
				if len(args) > 0 {
					return fmt.Errorf("--all flag cannot be used with plugin name")
				}
				return nil
			}
			if len(args) != 1 {
				return fmt.Errorf("requires exactly one plugin name (or use --all)")
			}
			return nil
		},
		RunE: runUninstall,
	}
)

func init() {
	uninstallCmd.Flags().BoolVar(&uninstallAll, "all", false, "Remove all installed plugins")
}

func runUninstall(cmd *cobra.Command, args []string) error {
	// Initialize server
	if err := initServer(); err != nil {
		scribe.Logger.Error("failed to initialize", "error", err)
		os.Exit(ExitRegistryError)
	}

	if uninstallAll {
		return runUninstallAll()
	}

	name := args[0]
	return runUninstallSingle(name)
}

func runUninstallSingle(name string) error {
	// Check if plugin exists
	if _, exists := server.GetRegistryEntry(name); !exists {
		if !quiet {
			fmt.Fprintf(os.Stderr, "Plugin '%s' not found\n", name)
		}
		os.Exit(ExitNotFound)
	}

	if !quiet {
		fmt.Printf("Uninstalling plugin '%s'...\n", name)
	}

	if err := server.UninstallPlugin(name); err != nil {
		if strings.Contains(err.Error(), "not found") {
			os.Exit(ExitNotFound)
		}
		scribe.Logger.Error("failed to uninstall plugin", "error", err)
		os.Exit(ExitRegistryError)
	}

	if !quiet {
		fmt.Printf("Plugin '%s' uninstalled successfully\n", name)
	}

	return nil
}

func runUninstallAll() error {
	count := server.PluginCount()
	if count == 0 {
		if !quiet {
			fmt.Println("No plugins installed")
		}
		return nil
	}

	if !quiet {
		fmt.Printf("Uninstalling all %d plugin(s)...\n", count)
	}

	if err := server.UninstallAllPlugins(); err != nil {
		scribe.Logger.Error("failed to uninstall all plugins", "error", err)
		os.Exit(ExitRegistryError)
	}

	if !quiet {
		fmt.Printf("All plugins uninstalled successfully\n")
	}

	return nil
}
