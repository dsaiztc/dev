package worktree

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dsaiztc/dev/internal/config"
)

// RepoInfo holds information about the current repository and its worktree context.
type RepoInfo struct {
	MainPath    string // absolute path to the main worktree
	Source      string // e.g. "github.com"
	Org         string // e.g. "dsaiztc"
	Repo        string // e.g. "dev"
	CurrentPath string // absolute path to the current worktree (may equal MainPath)
	IsLinked    bool   // true if current directory is inside a linked worktree
}

// Worktree represents a single git worktree entry.
type Worktree struct {
	Path   string
	Branch string
	IsMain bool
	Commit string
}

// DetectCurrentRepo determines the repository context from the current directory.
func DetectCurrentRepo() (*RepoInfo, error) {
	// Get the current worktree root
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return nil, fmt.Errorf("not inside a git repository: %w", err)
	}
	currentPath := strings.TrimSpace(string(out))

	// Check if .git is a file (linked worktree) or directory (main worktree)
	gitPath := filepath.Join(currentPath, ".git")
	info, err := os.Lstat(gitPath)
	if err != nil {
		return nil, fmt.Errorf("could not stat .git: %w", err)
	}

	var mainPath string
	isLinked := !info.IsDir()

	if isLinked {
		// .git is a file — parse it to find main worktree
		data, err := os.ReadFile(gitPath)
		if err != nil {
			return nil, fmt.Errorf("could not read .git file: %w", err)
		}
		// Format: "gitdir: /path/to/.git/worktrees/<name>"
		line := strings.TrimSpace(string(data))
		if !strings.HasPrefix(line, "gitdir: ") {
			return nil, fmt.Errorf("unexpected .git file format: %s", line)
		}
		gitDir := strings.TrimPrefix(line, "gitdir: ")
		// Resolve relative paths
		if !filepath.IsAbs(gitDir) {
			gitDir = filepath.Join(currentPath, gitDir)
		}
		gitDir = filepath.Clean(gitDir)
		// Navigate up from .git/worktrees/<name> to get .git, then parent is main worktree
		// gitDir = /path/to/main/.git/worktrees/<name>
		mainGitDir := filepath.Dir(filepath.Dir(gitDir)) // .git
		mainPath = filepath.Dir(mainGitDir)               // main worktree root
	} else {
		mainPath = currentPath
	}

	// Extract source/org/repo from main worktree path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not determine home directory: %w", err)
	}
	srcDir := filepath.Join(homeDir, "src")
	relPath, err := filepath.Rel(srcDir, mainPath)
	if err != nil || strings.HasPrefix(relPath, "..") {
		return nil, fmt.Errorf("main worktree %s is not under %s", mainPath, srcDir)
	}

	parts := strings.Split(relPath, string(filepath.Separator))
	if len(parts) < 3 {
		return nil, fmt.Errorf("unexpected repo path structure: %s", relPath)
	}

	return &RepoInfo{
		MainPath:    mainPath,
		Source:      parts[0],
		Org:         parts[1],
		Repo:        parts[2],
		CurrentPath: currentPath,
		IsLinked:    isLinked,
	}, nil
}

// ListWorktrees returns all worktrees for the given repo.
func ListWorktrees(repoInfo *RepoInfo) ([]Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = repoInfo.MainPath
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git worktree list failed: %w", err)
	}
	return ParseWorktreeListOutput(string(out), repoInfo.MainPath), nil
}

// ParseWorktreeListOutput parses the porcelain output of `git worktree list --porcelain`.
func ParseWorktreeListOutput(output string, mainPath string) []Worktree {
	var worktrees []Worktree
	var current Worktree

	// Resolve symlinks for reliable comparison
	resolvedMain, err := filepath.EvalSymlinks(mainPath)
	if err != nil {
		resolvedMain = filepath.Clean(mainPath)
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "worktree "):
			current = Worktree{Path: strings.TrimPrefix(line, "worktree ")}
		case strings.HasPrefix(line, "HEAD "):
			current.Commit = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch "):
			ref := strings.TrimPrefix(line, "branch ")
			current.Branch = strings.TrimPrefix(ref, "refs/heads/")
		case line == "":
			if current.Path != "" {
				resolvedPath, err := filepath.EvalSymlinks(current.Path)
				if err != nil {
					resolvedPath = filepath.Clean(current.Path)
				}
				current.IsMain = resolvedPath == resolvedMain
				worktrees = append(worktrees, current)
			}
			current = Worktree{}
		}
	}
	// Handle last entry if output doesn't end with blank line
	if current.Path != "" {
		resolvedPath, err := filepath.EvalSymlinks(current.Path)
		if err != nil {
			resolvedPath = filepath.Clean(current.Path)
		}
		current.IsMain = resolvedPath == resolvedMain
		worktrees = append(worktrees, current)
	}

	// Sort: main first, then alphabetically by branch
	sort.SliceStable(worktrees, func(i, j int) bool {
		if worktrees[i].IsMain != worktrees[j].IsMain {
			return worktrees[i].IsMain
		}
		return worktrees[i].Branch < worktrees[j].Branch
	})

	return worktrees
}

// CreateWorktree creates a new worktree with a new branch.
func CreateWorktree(repoInfo *RepoInfo, branchName string) (string, error) {
	root, err := GetWorktreeRoot()
	if err != nil {
		return "", err
	}

	targetPath := FormatWorktreePath(root, repoInfo.Source, repoInfo.Org, repoInfo.Repo, branchName)

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return "", fmt.Errorf("could not create worktree parent directory: %w", err)
	}

	cmd := exec.Command("git", "worktree", "add", targetPath, "-b", branchName)
	cmd.Dir = repoInfo.MainPath
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git worktree add failed: %w", err)
	}

	return targetPath, nil
}

// RemoveWorktree removes a linked worktree, its local branch, and remote branch (best-effort).
// Returns a cdPath if the caller should change directory (e.g., when removing the current worktree).
func RemoveWorktree(repoInfo *RepoInfo, wt Worktree) (string, error) {
	if wt.IsMain {
		return "", fmt.Errorf("cannot remove the main worktree")
	}

	// If cwd is inside the worktree being removed, we need to cd elsewhere
	var cdPath string
	cwd, err := os.Getwd()
	if err == nil {
		rel, err := filepath.Rel(wt.Path, cwd)
		if err == nil && !strings.HasPrefix(rel, "..") {
			cdPath = repoInfo.MainPath
		}
	}

	// Remove worktree
	cmd := exec.Command("git", "worktree", "remove", wt.Path, "--force")
	cmd.Dir = repoInfo.MainPath
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git worktree remove failed: %w", err)
	}

	// Delete local branch (best-effort)
	delBranch := exec.Command("git", "branch", "-D", wt.Branch)
	delBranch.Dir = repoInfo.MainPath
	_ = delBranch.Run()

	// Delete remote branch (best-effort)
	delRemote := exec.Command("git", "push", "origin", "--delete", wt.Branch)
	delRemote.Dir = repoInfo.MainPath
	_ = delRemote.Run()

	// Clean up empty parent directories in worktree root
	cleanEmptyParents(wt.Path)

	return cdPath, nil
}

// cleanEmptyParents removes empty parent directories up to but not including the worktree root.
func cleanEmptyParents(path string) {
	root, err := GetWorktreeRoot()
	if err != nil {
		return
	}
	dir := filepath.Dir(path)
	for dir != root && strings.HasPrefix(dir, root) {
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) > 0 {
			break
		}
		os.Remove(dir)
		dir = filepath.Dir(dir)
	}
}

// GetWorktreeRoot returns the worktree root directory from config or the default.
func GetWorktreeRoot() (string, error) {
	cfg, err := config.Load()
	if err != nil {
		if os.IsNotExist(err) {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("could not determine home directory: %w", err)
			}
			return filepath.Join(homeDir, "src__worktrees"), nil
		}
		return "", fmt.Errorf("could not load config: %w", err)
	}
	return cfg.GetWorktreeRoot(), nil
}

// FormatWorktreePath constructs the worktree directory path.
// Slashes in branch names are replaced with dashes to keep paths flat.
func FormatWorktreePath(root, source, org, repo, branch string) string {
	safeBranch := strings.ReplaceAll(branch, "/", "-")
	return filepath.Join(root, source, org, repo+"__"+safeBranch)
}
