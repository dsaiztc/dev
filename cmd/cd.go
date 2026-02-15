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

var cdCmd = &cobra.Command{
	Use:   "cd [query]",
	Short: "Navigate to a project directory",
	Long:  `Without arguments, opens an interactive fuzzy finder. With a query, jumps to the best matching repo.`,
	RunE:  runCD,
}

func init() {
	rootCmd.AddCommand(cdCmd)
}

func runCD(cmd *cobra.Command, args []string) error {
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
		fmt.Fprintf(os.Stderr, "%s\n", selected)
	}

	fullPath := filepath.Join(baseDir, selected)
	fmt.Printf("cd %s\n", fullPath)
	return nil
}
