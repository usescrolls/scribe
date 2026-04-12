package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	scribe "gitlab.com/usescrolls/scribe/internal"
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

			// Ensure system skill is installed and up to date
			if err := scribe.EnsureSystemSkill(); err != nil {
				scribe.Logger.Warn("failed to ensure system skill", "error", err)
			}

			// Skip onboarding check for certain commands
			cmdName := cmd.Name()
			skipOnboarding := cmdName == "setup" || cmdName == "version" || cmdName == "help" || cmdName == "scribe"
			if !skipOnboarding && checkOnboarding() {
				return runOnboardingIfNeeded()
			}

			// For existing users who completed onboarding before T&C were added,
			// prompt them to accept terms before proceeding.
			if !skipOnboarding {
				if err := checkAndPromptTerms(); err != nil {
					return err
				}
			}

			// Check for app updates (once per day, non-blocking)
			if !quiet && !skipOnboarding {
				checkForAppUpdate()
			}

			return nil
		},
	}
)

// checkForAppUpdate checks for a newer version of Scribe once per day.
// Prints a single-line message if an update is available.
// Never returns an error — failures are silently logged.
func checkForAppUpdate() {
	if scribe.Version == "dev" {
		return
	}

	if !scribe.ShouldCheckForUpdate(24 * time.Hour) {
		return
	}

	info, err := scribe.CheckForUpdate("")
	if err != nil {
		scribe.Logger.Debug("update check failed", "error", err)
		return
	}

	// Record that we checked, regardless of result
	if err := scribe.SetLastUpdateCheck(); err != nil {
		scribe.Logger.Debug("failed to record update check time", "error", err)
	}

	if info.UpdateAvailable {
		method := scribe.DetectInstallMethod()
		releaseURL := info.ReleaseURL
		if releaseURL == "" {
			releaseURL = scribe.PublicDownloadBase
		}
		switch method {
		case "binary", "app-bundle":
			fmt.Fprintf(os.Stderr,
				"\n  A new version of Scribe is available: %s (current: %s)\n  Run: scribe upgrade\n\n",
				info.LatestVersion, info.CurrentVersion)
		default:
			fmt.Fprintf(os.Stderr,
				"\n  A new version of Scribe is available: %s (current: %s)\n  %s\n\n",
				info.LatestVersion, info.CurrentVersion, releaseURL)
		}
	}
}

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
	rootCmd.AddCommand(upgradeCmd)
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
		"upgrade",
	}
}
