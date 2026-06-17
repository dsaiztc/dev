package cmd

import (
	"github.com/spf13/cobra"
)

var wtCmd = &cobra.Command{
	Use:     "wt",
	Aliases: []string{"wkt"},
	Short:   "Manage git worktrees",
	Long: `Create, navigate, and remove git worktrees organized under ~/src__worktrees/.

Run without a subcommand to navigate to a worktree via the fuzzy finder
(equivalent to "dev wt cd").

"wkt" is a deprecated alias for "wt" and will be removed in a future release.`,
	// Default to "cd" when invoked without a subcommand.
	RunE: runWtCd,
}

func init() {
	rootCmd.AddCommand(wtCmd)
}
