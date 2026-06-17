package cmd

import (
	"fmt"
	"os"

	"github.com/dsaiztc/dev/internal/fuzzy"
	"github.com/dsaiztc/dev/internal/worktree"
	"github.com/spf13/cobra"
)

var wtCdCmd = &cobra.Command{
	Use:   "cd [-]",
	Short: "Navigate to a worktree via fuzzy finder",
	Long: `Opens a fuzzy finder to navigate between worktrees of the current repository.

Use "-" to return to the previous directory (like "cd -").`,
	RunE: runWtCd,
}

func init() {
	wtCmd.AddCommand(wtCdCmd)
}

func runWtCd(cmd *cobra.Command, args []string) error {
	// "dev wt cd -" returns to the previous directory, mirroring shell "cd -".
	if len(args) == 1 && args[0] == "-" {
		fmt.Println("cd -")
		return nil
	}

	repoInfo, err := worktree.DetectCurrentRepo()
	if err != nil {
		return err
	}

	worktrees, err := worktree.ListWorktrees(repoInfo)
	if err != nil {
		return err
	}

	if len(worktrees) == 0 {
		return fmt.Errorf("no worktrees found")
	}

	// Build display items: branch name with (main) annotation
	items := make([]string, len(worktrees))
	pathMap := make(map[string]string, len(worktrees))
	for i, wt := range worktrees {
		label := wt.Branch
		if wt.IsMain {
			label += " (main)"
		}
		items[i] = label
		pathMap[label] = wt.Path
	}

	selected, err := fuzzy.Run(items)
	if err != nil {
		return err
	}
	if selected == "" {
		return nil
	}

	path := pathMap[selected]
	fmt.Fprintf(os.Stderr, "%s\n", selected)
	fmt.Printf("cd %s\n", path)
	return nil
}
