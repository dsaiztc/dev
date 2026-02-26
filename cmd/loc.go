package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dsaiztc/dev/internal/fuzzy"
	"github.com/dsaiztc/dev/internal/repos"
	"github.com/spf13/cobra"
)

var locCmd = &cobra.Command{
	Use:   "loc [query]",
	Short: "Locate and print the full path to a repository",
	Long:  `Without arguments, opens an interactive fuzzy finder. With a query, prints the path to the best matching repo.`,
	RunE:  runLoc,
}

func init() {
	rootCmd.AddCommand(locCmd)
}

func runLoc(cmd *cobra.Command, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	baseDir := filepath.Join(homeDir, "src")

	allRepos, err := repos.Discover(baseDir)
	if err != nil {
		return fmt.Errorf("could not discover repos: %w", err)
	}

	if len(allRepos) == 0 {
		return fmt.Errorf("no repos found under %s", baseDir)
	}

	var selected string

	if len(args) == 0 {
		// Interactive fuzzy finder
		selected, err = fuzzy.Run(allRepos)
		if err != nil {
			return err
		}
		if selected == "" {
			return nil // User cancelled
		}
	} else {
		// Fuzzy match with query
		query := strings.Join(args, " ")
		matches := repos.FuzzyMatch(allRepos, query)
		if len(matches) == 0 {
			return fmt.Errorf("no repos matching %q", query)
		}
		selected = matches[0]
	}

	fullPath := filepath.Join(baseDir, selected)
	fmt.Println(fullPath)
	return nil
}
