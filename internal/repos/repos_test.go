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

func TestDiscover_TooDeep(t *testing.T) {
	base := t.TempDir()

	// Create a repo at depth 6 (should not be found, maxDepth=4)
	deep := filepath.Join(base, "a", "b", "c", "d", "e", "f")
	os.MkdirAll(filepath.Join(deep, ".git"), 0o755)

	repos, err := Discover(base)
	if err != nil {
		t.Fatalf("Discover() error: %v", err)
	}

	if len(repos) != 0 {
		t.Errorf("expected no repos at depth > %d, got %v", maxDepth, repos)
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
