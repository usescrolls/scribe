package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	scribe "github.com/usescrolls/scribe/internal"
)

var (
	cacheCmd = &cobra.Command{
		Use:   "cache",
		Short: "Manage the local clone cache",
		Long: `Manage the local cache of cloned repositories.

Scribe caches cloned repositories to speed up subsequent installs, checks,
and updates. Use these commands to inspect or clear the cache.

Examples:
  scribe cache path
  scribe cache clear`,
	}

	cacheClearCmd = &cobra.Command{
		Use:   "clear",
		Short: "Clear the entire clone cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := scribe.ClearCache(); err != nil {
				return fmt.Errorf("failed to clear cache: %w", err)
			}
			if !quiet {
				fmt.Println("Cache cleared")
			}
			return nil
		},
	}

	cachePathCmd = &cobra.Command{
		Use:   "path",
		Short: "Print the cache directory path",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := scribe.GetCacheDir()
			if err != nil {
				return err
			}
			fmt.Println(dir)
			return nil
		},
	}
)

func init() {
	cacheCmd.AddCommand(cacheClearCmd)
	cacheCmd.AddCommand(cachePathCmd)
}
