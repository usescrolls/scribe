package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	scribe "github.com/usescrolls/scribe/internal"
)

var (
	// Global flags
	debug      bool
	jsonOutput bool
	quiet      bool

	// Root command
	rootCmd = &cobra.Command{
		Use:   "scribe",
		Short: "Skills manager for coding agents",
		Long: `Scribe CLI provides intuitive command-line management for skills across 45+ coding agents.

Examples:
  scribe install owner/repo
  scribe list
  scribe uninstall my-skill
  scribe workspace list`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Initialize quiet logger for CLI mode
			// Only shows logs when --debug flag is passed
			scribe.InitLoggerCLI(debug)
			// Ensure scribe directories exist
			if err := scribe.EnsureScribeDirs(); err != nil {
				scribe.Logger.Error("failed to initialize scribe directories", "error", err)
				return err
			}

			// Skip onboarding check for certain commands
			cmdName := cmd.Name()
			skipOnboarding := cmdName == "setup" || cmdName == "version" || cmdName == "help" || cmdName == "scribe"
			if !skipOnboarding && checkOnboarding() {
				return runOnboardingIfNeeded()
			}

			return nil
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
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(cacheCmd)
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
		"check",
		"update",
		"cache",
		"setup",
	}
}
