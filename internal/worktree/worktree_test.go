package worktree

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestFormatWorktreePath(t *testing.T) {
	tests := []struct {
		name   string
		root   string
		source string
		org    string
		repo   string
		branch string
		want   string
	}{
		{
			name:   "simple branch",
			root:   "/home/user/src__worktrees",
			source: "github.com",
			org:    "dsaiztc",
			repo:   "dev",
			branch: "feature-x",
			want:   "/home/user/src__worktrees/github.com/dsaiztc/dev__feature-x",
		},
		{
			name:   "branch with slashes",
			root:   "/home/user/src__worktrees",
			source: "github.com",
			org:    "dsaiztc",
			repo:   "dev",
			branch: "fix/bug-123",
			want:   "/home/user/src__worktrees/github.com/dsaiztc/dev__fix--bug-123",
		},
		{
			name:   "different source",
			root:   "/tmp/worktrees",
			source: "gitlab.com",
			org:    "team",
			repo:   "service",
			branch: "main-v2",
			want:   "/tmp/worktrees/gitlab.com/team/service__main-v2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatWorktreePath(tt.root, tt.source, tt.org, tt.repo, tt.branch)
			if got != tt.want {
				t.Errorf("FormatWorktreePath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseWorktreeListOutput(t *testing.T) {
	mainPath := "/home/user/src/github.com/dsaiztc/dev"

	tests := []struct {
		name     string
		output   string
		mainPath string
		want     []Worktree
	}{
		{
			name: "single main worktree",
			output: "worktree /home/user/src/github.com/dsaiztc/dev\n" +
				"HEAD abc1234\n" +
				"branch refs/heads/main\n" +
				"\n",
			mainPath: mainPath,
			want: []Worktree{
				{Path: mainPath, Branch: "main", IsMain: true, Commit: "abc1234"},
			},
		},
		{
			name: "main and linked worktrees",
			output: "worktree /home/user/src/github.com/dsaiztc/dev\n" +
				"HEAD abc1234\n" +
				"branch refs/heads/main\n" +
				"\n" +
				"worktree /home/user/src__worktrees/github.com/dsaiztc/dev__feature-x\n" +
				"HEAD def5678\n" +
				"branch refs/heads/feature-x\n" +
				"\n" +
				"worktree /home/user/src__worktrees/github.com/dsaiztc/dev__alpha\n" +
				"HEAD 111222\n" +
				"branch refs/heads/alpha\n" +
				"\n",
			mainPath: mainPath,
			want: []Worktree{
				{Path: mainPath, Branch: "main", IsMain: true, Commit: "abc1234"},
				{Path: "/home/user/src__worktrees/github.com/dsaiztc/dev__alpha", Branch: "alpha", IsMain: false, Commit: "111222"},
				{Path: "/home/user/src__worktrees/github.com/dsaiztc/dev__feature-x", Branch: "feature-x", IsMain: false, Commit: "def5678"},
			},
		},
		{
			name:     "empty output",
			output:   "",
			mainPath: mainPath,
			want:     nil,
		},
		{
			name: "output without trailing newline",
			output: "worktree /home/user/src/github.com/dsaiztc/dev\n" +
				"HEAD abc1234\n" +
				"branch refs/heads/main",
			mainPath: mainPath,
			want: []Worktree{
				{Path: mainPath, Branch: "main", IsMain: true, Commit: "abc1234"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseWorktreeListOutput(tt.output, tt.mainPath)
			if len(got) != len(tt.want) {
				t.Fatalf("ParseWorktreeListOutput() returned %d worktrees, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i].Path != tt.want[i].Path {
					t.Errorf("[%d] Path = %q, want %q", i, got[i].Path, tt.want[i].Path)
				}
				if got[i].Branch != tt.want[i].Branch {
					t.Errorf("[%d] Branch = %q, want %q", i, got[i].Branch, tt.want[i].Branch)
				}
				if got[i].IsMain != tt.want[i].IsMain {
					t.Errorf("[%d] IsMain = %v, want %v", i, got[i].IsMain, tt.want[i].IsMain)
				}
				if got[i].Commit != tt.want[i].Commit {
					t.Errorf("[%d] Commit = %q, want %q", i, got[i].Commit, tt.want[i].Commit)
				}
			}
		})
	}
}

// setupTestRepo creates a temporary git repo that mimics the ~/src/source/org/repo structure.
func setupTestRepo(t *testing.T) (homeDir string, repoPath string) {
	t.Helper()
	homeDir = t.TempDir()
	repoPath = filepath.Join(homeDir, "src", "github.com", "testuser", "testrepo")
	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		t.Fatal(err)
	}

	// git init
	cmd := exec.Command("git", "init", repoPath)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init failed: %v", err)
	}

	// Configure git user for commits
	for _, c := range [][]string{
		{"git", "-C", repoPath, "config", "user.email", "test@test.com"},
		{"git", "-C", repoPath, "config", "user.name", "Test"},
	} {
		if err := exec.Command(c[0], c[1:]...).Run(); err != nil {
			t.Fatalf("%v failed: %v", c, err)
		}
	}

	// Create initial commit
	dummyFile := filepath.Join(repoPath, "README.md")
	if err := os.WriteFile(dummyFile, []byte("# test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, c := range [][]string{
		{"git", "-C", repoPath, "add", "."},
		{"git", "-C", repoPath, "commit", "-m", "initial"},
	} {
		if err := exec.Command(c[0], c[1:]...).Run(); err != nil {
			t.Fatalf("%v failed: %v", c, err)
		}
	}

	return homeDir, repoPath
}

func TestListWorktrees_Integration(t *testing.T) {
	_, repoPath := setupTestRepo(t)

	info := &RepoInfo{MainPath: repoPath}
	worktrees, err := ListWorktrees(info)
	if err != nil {
		t.Fatalf("ListWorktrees: %v", err)
	}

	if len(worktrees) != 1 {
		t.Fatalf("expected 1 worktree, got %d", len(worktrees))
	}
	if !worktrees[0].IsMain {
		t.Error("expected first worktree to be main")
	}
}

func TestCreateAndListWorktrees_Integration(t *testing.T) {
	homeDir, repoPath := setupTestRepo(t)
	wtRoot := filepath.Join(homeDir, "src__worktrees")

	info := &RepoInfo{
		MainPath: repoPath,
		Source:   "github.com",
		Org:      "testuser",
		Repo:     "testrepo",
	}

	// Create worktree using git directly (since CreateWorktree reads config)
	wtPath := FormatWorktreePath(wtRoot, "github.com", "testuser", "testrepo", "feature-a")
	if err := os.MkdirAll(filepath.Dir(wtPath), 0o755); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "worktree", "add", wtPath, "-b", "feature-a")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("git worktree add: %v", err)
	}

	worktrees, err := ListWorktrees(info)
	if err != nil {
		t.Fatalf("ListWorktrees: %v", err)
	}

	if len(worktrees) != 2 {
		t.Fatalf("expected 2 worktrees, got %d", len(worktrees))
	}
	if !worktrees[0].IsMain {
		t.Error("expected first worktree to be main")
	}
	if worktrees[1].Branch != "feature-a" {
		t.Errorf("expected second worktree branch 'feature-a', got %q", worktrees[1].Branch)
	}
}
