package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dsaiztc/dev/internal/fuzzy"
	"github.com/dsaiztc/dev/internal/worktree"
	"github.com/spf13/cobra"
)

var wtRmCmd = &cobra.Command{
	Use:   "rm [branch]",
	Short: "Remove a worktree and its branch",
	Long: `Remove a worktree, its local branch, and its remote branch.

From a linked worktree (no args): removes the current worktree.
From the main worktree: specify a branch name or pick one from the fuzzy finder.

This command deletes the worktree directory, the local branch (git branch -D),
and the remote branch (git push origin --delete). Prompts for confirmation
unless --yes is given.`,
	Args:        cobra.MaximumNArgs(1),
	Annotations: map[string]string{"destructive": "true"},
	RunE:        runWtRm,
}

func init() {
	wtRmCmd.Flags().BoolP("yes", "y", false, "skip confirmation prompt (required when non-interactive)")
	wtRmCmd.Flags().Bool("force", false, "alias for --yes")
	wtCmd.AddCommand(wtRmCmd)
}

func runWtRm(cmd *cobra.Command, args []string) error {
	yes, _ := cmd.Flags().GetBool("yes")
	force, _ := cmd.Flags().GetBool("force")
	skipPrompt := yes || force

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
		// From main worktree with arg: exact then substring match
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

	prompt := fmt.Sprintf("remove worktree %q (branch: %s, path: %s)? [y/N] ", target.Branch, target.Branch, target.Path)
	ok, err := confirmRemoval(skipPrompt, prompt, os.Stderr)
	if err != nil {
		return err
	}
	if !ok {
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

// confirmRemoval handles the confirmation step for destructive operations.
// When skipPrompt is true it proceeds immediately. When no TTY is available
// and skipPrompt is false it returns an error — never silently cancels.
func confirmRemoval(skipPrompt bool, prompt string, stderr io.Writer) (bool, error) {
	if skipPrompt {
		return true, nil
	}

	tty, err := os.Open("/dev/tty")
	if err != nil {
		return false, fmt.Errorf("refusing to remove worktree without --yes: no interactive terminal available")
	}
	defer tty.Close()

	fmt.Fprint(stderr, prompt)
	reader := bufio.NewReader(tty)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("could not read confirmation: %w", err)
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes", nil
}
