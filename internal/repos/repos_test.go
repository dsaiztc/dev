package repos

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscover(t *testing.T) {
	// Create a temp directory structure
	base := t.TempDir()

	// source/org/project with .git
	repo1 := filepath.Join(base, "github.com", "dsaiztc", "dev")
	os.MkdirAll(filepath.Join(repo1, ".git"), 0o755)

	// source/org/project2 with .git
	repo2 := filepath.Join(base, "github.com", "dsaiztc", "other")
	os.MkdirAll(filepath.Join(repo2, ".git"), 0o755)

	// source/org/project without .git (should not appear)
	noGit := filepath.Join(base, "github.com", "dsaiztc", "nogit")
	os.MkdirAll(noGit, 0o755)

	// nested repo inside a repo (should not appear — parent is a repo)
	nested := filepath.Join(repo1, "vendor", "nested")
	os.MkdirAll(filepath.Join(nested, ".git"), 0o755)

	repos, err := Discover(base)
	if err != nil {
		t.Fatalf("Discover() error: %v", err)
	}

	want := map[string]bool{
		"github.com/dsaiztc/dev":   true,
		"github.com/dsaiztc/other": true,
	}

	if len(repos) != len(want) {
		t.Fatalf("Discover() returned %d repos, want %d: %v", len(repos), len(want), repos)
	}

	for _, r := range repos {
		if !want[r] {
			t.Errorf("unexpected repo: %s", r)
		}
	}
}

func TestDiscover_WrongDepth(t *testing.T) {
	base := t.TempDir()

	// Repo at depth 2 (source/project) — should not be found
	shallow := filepath.Join(base, "github.com", "solo-project")
	os.MkdirAll(filepath.Join(shallow, ".git"), 0o755)

	// Repo at depth 4 (source/org/sub/project) — should not be found
	deep := filepath.Join(base, "gitlab.com", "org", "sub", "project")
	os.MkdirAll(filepath.Join(deep, ".git"), 0o755)

	repos, err := Discover(base)
	if err != nil {
		t.Fatalf("Discover() error: %v", err)
	}

	if len(repos) != 0 {
		t.Errorf("expected no repos at wrong depth, got %v", repos)
	}
}

func TestDiscover_IgnoresParentGit(t *testing.T) {
	base := t.TempDir()

	// Org directory has .git (should be ignored, not at depth 3)
	orgDir := filepath.Join(base, "github.com", "dsaiztc")
	os.MkdirAll(filepath.Join(orgDir, ".git"), 0o755)

	// Repos inside should still be found
	repo := filepath.Join(orgDir, "dev")
	os.MkdirAll(filepath.Join(repo, ".git"), 0o755)

	repos, err := Discover(base)
	if err != nil {
		t.Fatalf("Discover() error: %v", err)
	}

	if len(repos) != 1 || repos[0] != "github.com/dsaiztc/dev" {
		t.Errorf("Discover() = %v, want [github.com/dsaiztc/dev]", repos)
	}
}

func TestFuzzyMatch(t *testing.T) {
	repos := []string{
		"github.com/dsaiztc/dev",
		"github.com/dsaiztc/dotfiles",
		"github.com/apache/kafka",
		"gitlab.com/team/service",
	}

	results := FuzzyMatch(repos, "kafka")
	if len(results) == 0 {
		t.Fatal("FuzzyMatch('kafka') returned no results")
	}
	if results[0] != "github.com/apache/kafka" {
		t.Errorf("FuzzyMatch('kafka')[0] = %q, want 'github.com/apache/kafka'", results[0])
	}

	results = FuzzyMatch(repos, "dev")
	if len(results) == 0 {
		t.Fatal("FuzzyMatch('dev') returned no results")
	}
	if results[0] != "github.com/dsaiztc/dev" {
		t.Errorf("FuzzyMatch('dev')[0] = %q, want 'github.com/dsaiztc/dev'", results[0])
	}
}
