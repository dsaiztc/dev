package repos

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/sahilm/fuzzy"
)

const maxDepth = 4

// Discover walks baseDir up to maxDepth levels deep and returns relative paths
// of directories containing a .git directory. It stops descending into repos.
func Discover(baseDir string) ([]string, error) {
	var repos []string

	err := walkDepth(baseDir, baseDir, 0, &repos)
	if err != nil {
		return nil, err
	}

	sort.Strings(repos)
	return repos, nil
}

func walkDepth(baseDir, currentDir string, depth int, repos *[]string) error {
	if depth > maxDepth {
		return nil
	}

	entries, err := os.ReadDir(currentDir)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name()[0] == '.' {
			continue
		}

		fullPath := filepath.Join(currentDir, entry.Name())

		// Check if this dir contains .git
		gitDir := filepath.Join(fullPath, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			relPath, err := filepath.Rel(baseDir, fullPath)
			if err != nil {
				continue
			}
			*repos = append(*repos, relPath)
			// Don't descend into repos
			continue
		}

		// Recurse deeper
		if err := walkDepth(baseDir, fullPath, depth+1, repos); err != nil {
			continue
		}
	}

	return nil
}

// FuzzyMatch matches the query against repo paths and returns results sorted by score.
func FuzzyMatch(repos []string, query string) []string {
	matches := fuzzy.Find(query, repos)
	result := make([]string, len(matches))
	for i, m := range matches {
		result[i] = m.Str
	}
	return result
}
