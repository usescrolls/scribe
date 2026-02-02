package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/usescrolls/scribe/internal"
)

var (
	// Global flags
	debug      bool
	jsonOutput bool
	quiet      bool

	// Shared server instance
	server *scribe.Server

	// Root command
	rootCmd = &cobra.Command{
		Use:   "scribe",
		Short: "Package manager for Claude Code plugins",
		Long: `Scribe CLI provides intuitive command-line package management for Claude Code plugins.

Examples:
  scribe install prettier --github usescrolls/prettier-skill
  scribe list
  scribe uninstall prettier`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Initialize quiet logger for CLI mode
			// Only shows logs when --debug flag is passed
			scribe.InitLoggerCLI(debug)
			// Initialize server
			server = scribe.NewServer()
		},
	}
)

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format (where applicable)")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-essential output")

	// Subcommands
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(workspaceCmd)
}

// Execute runs the CLI
func Execute() int {
	if err := rootCmd.Execute(); err != nil {
		if !quiet {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		}
		return ExitError
	}
	return ExitSuccess
}

// initServer initializes and loads the server state
// Called by commands that need access to the registry
func initServer() error {
	if err := server.Initialize(); err != nil {
		return err
	}
	if err := server.Load(); err != nil {
		scribe.Logger.Warn("failed to load existing registry", "error", err)
	}
	return nil
}

// CLICommands returns the list of known CLI commands for detection
func CLICommands() []string {
	return []string{
		"install",
		"uninstall", "remove", "rm",
		"list", "ls",
		"info",
		"version",
		"help",
		"workspace",
	}
}
