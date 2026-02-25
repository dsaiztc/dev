package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dsaiztc/dev/internal/repos"
	"github.com/spf13/cobra"
)

var treeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Display a tree view of all repositories",
	Long:  `Shows the directory structure from ~/src/ down to each repository.`,
	RunE:  runTree,
}

func init() {
	rootCmd.AddCommand(treeCmd)
}

func runTree(cmd *cobra.Command, args []string) error {
	// 1. Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	// 2. Build base path
	baseDir := filepath.Join(homeDir, "src")

	// 3. Discover repos
	allRepos, err := repos.Discover(baseDir)
	if err != nil {
		return fmt.Errorf("could not discover repos: %w", err)
	}

	// 4. Handle empty case
	if len(allRepos) == 0 {
		fmt.Fprintf(os.Stderr, "no repos found under %s\n", baseDir)
		return nil
	}

	// 5. Build and render tree
	root := buildTree(allRepos)
	fmt.Printf("~/src/\n")
	printTree(root, "", false)

	return nil
}

// treeNode represents a node in the directory tree
type treeNode struct {
	name     string
	children map[string]*treeNode
}

// newTreeNode creates a new tree node with the given name
func newTreeNode(name string) *treeNode {
	return &treeNode{
		name:     name,
		children: make(map[string]*treeNode),
	}
}

// buildTree constructs a tree structure from a flat list of repository paths
func buildTree(repoPaths []string) *treeNode {
	root := newTreeNode("")

	for _, path := range repoPaths {
		// Split path: "github.com/dsaiztc/dev" → ["github.com", "dsaiztc", "dev"]
		parts := strings.Split(path, string(filepath.Separator))
		current := root

		// Walk/create nodes for each part
		for _, part := range parts {
			if current.children[part] == nil {
				current.children[part] = newTreeNode(part)
			}
			current = current.children[part]
		}
	}

	return root
}

// printTree recursively renders the tree structure with ASCII characters
func printTree(node *treeNode, prefix string, isLast bool) {
	// Get sorted children (alphabetical order)
	keys := make([]string, 0, len(node.children))
	for k := range node.children {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Render each child
	for i, key := range keys {
		child := node.children[key]
		isLastChild := i == len(keys)-1

		// Determine branch character
		branch := "├── "
		if isLastChild {
			branch = "└── "
		}

		// Print node
		fmt.Printf("%s%s%s\n", prefix, branch, child.name)

		// Calculate prefix for children
		var newPrefix string
		if isLastChild {
			newPrefix = prefix + "    " // spaces for last child
		} else {
			newPrefix = prefix + "│   " // vertical bar for non-last
		}

		// Recurse
		printTree(child, newPrefix, isLastChild)
	}
}
