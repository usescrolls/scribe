package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	scribe "gitlab.com/usescrolls/scribe/internal"
)

var desktopCmd = &cobra.Command{
	Use:   "desktop",
	Short: "Launch the Scribe desktop app",
	Long:  `Launch the installed Scribe desktop app without loading desktop UI libraries in the CLI process.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !quiet {
			fmt.Println("Launching Scribe desktop app...")
		}
		return scribe.LaunchDesktop()
	},
}
