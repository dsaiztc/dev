package cmd

import (
	"fmt"
	"os"

	"github.com/dsaiztc/dev/internal/worktree"
	"github.com/spf13/cobra"
)

var wtNewCmd = &cobra.Command{
	Use:   "new <branch>",
	Short: "Create a new worktree with a new branch",
	Args:  cobra.ExactArgs(1),
	RunE:  runWtNew,
}

func init() {
	wtCmd.AddCommand(wtNewCmd)
}

func runWtNew(cmd *cobra.Command, args []string) error {
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
