package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dsaiztc/dev/internal/repourl"
	"github.com/spf13/cobra"
)

var cloneCmd = &cobra.Command{
	Use:   "clone <url>",
	Short: "Clone a git repository into ~/src/<source>/<org>/<project>",
	Args:  cobra.ExactArgs(1),
	RunE:  runClone,
}

func init() {
	rootCmd.AddCommand(cloneCmd)
}

func runClone(cmd *cobra.Command, args []string) error {
	parsed, err := repourl.Parse(args[0])
	if err != nil {
		return fmt.Errorf("invalid repository URL: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	targetDir := filepath.Join(homeDir, "src", parsed.FullPath())

	// Check if target already exists
	if info, err := os.Stat(targetDir); err == nil && info.IsDir() {
		gitDir := filepath.Join(targetDir, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			fmt.Fprintf(os.Stderr, "already cloned at %s\n", targetDir)
			fmt.Println(targetDir)
			return nil
		}
		return fmt.Errorf("directory %s already exists but is not a git repository", targetDir)
	}

	// Create parent directory
	parentDir := filepath.Dir(targetDir)
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		return fmt.Errorf("could not create directory %s: %w", parentDir, err)
	}

	// Run git clone
	gitCmd := exec.Command("git", "clone", args[0], targetDir)
	gitCmd.Stdin = os.Stdin
	gitCmd.Stdout = os.Stderr
	gitCmd.Stderr = os.Stderr

	fmt.Fprintf(os.Stderr, "cloning into %s\n", targetDir)
	if err := gitCmd.Run(); err != nil {
		// Clean up partial directory on failure
		os.RemoveAll(targetDir)
		return fmt.Errorf("git clone failed: %w", err)
	}

	// Print target path to stdout (for shell wrapper to eval)
	fmt.Println(targetDir)
	return nil
}
