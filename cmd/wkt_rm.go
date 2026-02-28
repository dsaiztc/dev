package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/dsaiztc/dev/internal/fuzzy"
	"github.com/dsaiztc/dev/internal/worktree"
	"github.com/spf13/cobra"
)

var wktRmCmd = &cobra.Command{
	Use:   "rm [branch]",
	Short: "Remove a worktree and its branch",
	Long:  `From a linked worktree: removes the current worktree. From the main worktree: removes the specified or selected worktree.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runWktRm,
}

func init() {
	wktCmd.AddCommand(wktRmCmd)
}

func runWktRm(cmd *cobra.Command, args []string) error {
	repoInfo, err := worktree.DetectCurrentRepo()
	if err != nil {
		return err
	}

	worktrees, err := worktree.ListWorktrees(repoInfo)
	if err != nil {
		return err
	}

	var target worktree.Worktree
	var found bool

	if repoInfo.IsLinked {
		// From linked worktree: remove the current one
		for _, wt := range worktrees {
			if wt.Path == repoInfo.CurrentPath {
				target = wt
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("could not find current worktree in list")
		}
	} else if len(args) == 1 {
		// From main worktree with arg: fuzzy match against branch names
		query := args[0]
		for _, wt := range worktrees {
			if wt.IsMain {
				continue
			}
			if wt.Branch == query {
				target = wt
				found = true
				break
			}
		}
		if !found {
			// Try substring/fuzzy match
			var candidates []string
			branchMap := make(map[string]worktree.Worktree)
			for _, wt := range worktrees {
				if wt.IsMain {
					continue
				}
				candidates = append(candidates, wt.Branch)
				branchMap[wt.Branch] = wt
			}
			if len(candidates) == 0 {
				return fmt.Errorf("no linked worktrees to remove")
			}
			for _, c := range candidates {
				if strings.Contains(c, query) {
					target = branchMap[c]
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("no worktree matching %q", query)
			}
		}
	} else {
		// From main worktree without args: fuzzy finder
		var items []string
		pathMap := make(map[string]worktree.Worktree)
		for _, wt := range worktrees {
			if wt.IsMain {
				continue
			}
			items = append(items, wt.Branch)
			pathMap[wt.Branch] = wt
		}
		if len(items) == 0 {
			return fmt.Errorf("no linked worktrees to remove")
		}

		selected, err := fuzzy.Run(items)
		if err != nil {
			return err
		}
		if selected == "" {
			return nil
		}
		target = pathMap[selected]
	}

	// Confirm removal
	fmt.Fprintf(os.Stderr, "remove worktree %q (branch: %s, path: %s)? [y/N] ", target.Branch, target.Branch, target.Path)
	if !confirmFromTTY() {
		fmt.Fprintln(os.Stderr, "cancelled")
		return nil
	}

	cdPath, err := worktree.RemoveWorktree(repoInfo, target)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "removed worktree %q\n", target.Branch)
	if cdPath != "" {
		fmt.Printf("cd %s\n", cdPath)
	}
	return nil
}

func confirmFromTTY() bool {
	tty, err := os.Open("/dev/tty")
	if err != nil {
		return false
	}
	defer tty.Close()

	reader := bufio.NewReader(tty)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes"
}
