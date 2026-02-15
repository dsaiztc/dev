package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dev",
	Short: "A CLI tool for managing development projects",
	Long:  `dev reduces cognitive load when navigating between development projects by enforcing an opinionated directory structure (~/src/<source>/<org>/<project>).`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
