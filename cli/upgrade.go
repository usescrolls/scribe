package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	scribe "gitlab.com/usescrolls/scribe/internal"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade Scribe to the latest version",
	Long: `Upgrade the Scribe binary to the latest release.

This command downloads and replaces the current binary with the latest version.

Examples:
  scribe upgrade`,
	Args: cobra.NoArgs,
	RunE: runUpgrade,
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	if !quiet {
		fmt.Printf("Current version: %s\n", scribe.Version)
		fmt.Printf("Install method:  %s\n", scribe.DetectInstallMethod())
	}

	result, err := scribe.SelfUpdate("")
	if err != nil {
		return err
	}

	if !result.Updated {
		if !quiet {
			fmt.Printf("Already up to date (%s)\n", result.OldVersion)
		}
		return nil
	}

	if !quiet {
		fmt.Printf("Successfully upgraded: %s -> %s\n", result.OldVersion, result.NewVersion)
	}

	return nil
}
