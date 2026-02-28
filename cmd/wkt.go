package cmd

import (
	"github.com/spf13/cobra"
)

var wktCmd = &cobra.Command{
	Use:   "wkt",
	Short: "Manage git worktrees",
	Long:  `Create, navigate, and remove git worktrees organized under ~/src__worktrees/.`,
}

func init() {
	rootCmd.AddCommand(wktCmd)
}
