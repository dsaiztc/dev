package repos

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/sahilm/fuzzy"
)

// repoDepth is the expected depth of repos under baseDir: source/org/project.
const repoDepth = 3

// Discover finds all repos at exactly depth 3 (source/org/project) under
// baseDir that contain a .git directory.
func Discover(baseDir string) ([]string, error) {
	var repos []string

	err := walkToDepth(baseDir, baseDir, 0, &repos)
	if err != nil {
		return nil, err
	}

	sort.Strings(repos)
	return repos, nil
}

func walkToDepth(baseDir, currentDir string, depth int, repos *[]string) error {
	entries, err := os.ReadDir(currentDir)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name()[0] == '.' {
			continue
		}

		fullPath := filepath.Join(currentDir, entry.Name())

		if depth+1 == repoDepth {
			// At target depth — check for .git
			gitDir := filepath.Join(fullPath, ".git")
			if _, err := os.Stat(gitDir); err == nil {
				relPath, err := filepath.Rel(baseDir, fullPath)
				if err != nil {
					continue
				}
				*repos = append(*repos, relPath)
			}
		} else {
			// Not deep enough yet — keep descending
			walkToDepth(baseDir, fullPath, depth+1, repos)
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
