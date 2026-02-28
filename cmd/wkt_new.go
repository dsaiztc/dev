package cmd

import (
	"fmt"
	"os"

	"github.com/dsaiztc/dev/internal/worktree"
	"github.com/spf13/cobra"
)

var wktNewCmd = &cobra.Command{
	Use:   "new <branch>",
	Short: "Create a new worktree with a new branch",
	Args:  cobra.ExactArgs(1),
	RunE:  runWktNew,
}

func init() {
	wktCmd.AddCommand(wktNewCmd)
}

func runWktNew(cmd *cobra.Command, args []string) error {
	branchName := args[0]

	repoInfo, err := worktree.DetectCurrentRepo()
	if err != nil {
		return err
	}

	path, err := worktree.CreateWorktree(repoInfo, branchName)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "created worktree for branch %q at %s\n", branchName, path)
	fmt.Printf("cd %s\n", path)
	return nil
}
