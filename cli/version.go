package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	scribe "gitlab.com/usescrolls/scribe/internal"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  `Show version information.`,
		Args:  cobra.NoArgs,
		Run:   runVersion,
	}
)

func runVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("scribe version %s\n", scribe.Version)
}
