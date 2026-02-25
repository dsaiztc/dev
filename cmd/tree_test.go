package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

func TestBuildTree_SingleRepo(t *testing.T) {
	repos := []string{"github.com/dsaiztc/dev"}

	root := buildTree(repos)

	if root == nil {
		t.Fatal("expected non-nil root")
	}

	// Check first level: github.com
	if len(root.children) != 1 {
		t.Errorf("expected 1 child at root, got %d", len(root.children))
	}

	githubNode, ok := root.children["github.com"]
	if !ok {
		t.Fatal("expected github.com node")
	}

	// Check second level: dsaiztc
	if len(githubNode.children) != 1 {
		t.Errorf("expected 1 child under github.com, got %d", len(githubNode.children))
	}

	orgNode, ok := githubNode.children["dsaiztc"]
	if !ok {
		t.Fatal("expected dsaiztc node")
	}

	// Check third level: dev
	if len(orgNode.children) != 1 {
		t.Errorf("expected 1 child under dsaiztc, got %d", len(orgNode.children))
	}

	devNode, ok := orgNode.children["dev"]
	if !ok {
		t.Fatal("expected dev node")
	}

	// Dev should be a leaf
	if len(devNode.children) != 0 {
		t.Errorf("expected dev to be leaf, got %d children", len(devNode.children))
	}
}

func TestBuildTree_MultipleRepos(t *testing.T) {
	repos := []string{
		"github.com/dsaiztc/dev",
		"github.com/dsaiztc/dotfiles",
		"github.com/another-org/project",
		"gitlab.com/team/service",
	}

	root := buildTree(repos)

	// Check sources
	if len(root.children) != 2 {
		t.Errorf("expected 2 sources, got %d", len(root.children))
	}

	// Check github.com structure
	githubNode := root.children["github.com"]
	if githubNode == nil {
		t.Fatal("expected github.com node")
	}
	if len(githubNode.children) != 2 {
		t.Errorf("expected 2 orgs under github.com, got %d", len(githubNode.children))
	}

	// Check dsaiztc org
	dsaizNode := githubNode.children["dsaiztc"]
	if dsaizNode == nil {
		t.Fatal("expected dsaiztc node")
	}
	if len(dsaizNode.children) != 2 {
		t.Errorf("expected 2 repos under dsaiztc, got %d", len(dsaizNode.children))
	}

	// Check repos exist
	if dsaizNode.children["dev"] == nil {
		t.Error("expected dev repo")
	}
	if dsaizNode.children["dotfiles"] == nil {
		t.Error("expected dotfiles repo")
	}

	// Check another-org
	anotherNode := githubNode.children["another-org"]
	if anotherNode == nil {
		t.Fatal("expected another-org node")
	}
	if len(anotherNode.children) != 1 {
		t.Errorf("expected 1 repo under another-org, got %d", len(anotherNode.children))
	}

	// Check gitlab.com
	gitlabNode := root.children["gitlab.com"]
	if gitlabNode == nil {
		t.Fatal("expected gitlab.com node")
	}
	if len(gitlabNode.children) != 1 {
		t.Errorf("expected 1 org under gitlab.com, got %d", len(gitlabNode.children))
	}
}

func TestBuildTree_Sorting(t *testing.T) {
	// Create repos in non-alphabetical order
	repos := []string{
		"github.com/zebra/project",
		"github.com/alpha/project",
		"github.com/beta/project",
	}

	root := buildTree(repos)
	githubNode := root.children["github.com"]

	// Get keys in order they appear
	keys := make([]string, 0, len(githubNode.children))
	for k := range githubNode.children {
		keys = append(keys, k)
	}

	// Note: The map iteration order is not guaranteed, but printTree sorts them
	// This test just verifies all orgs are present
	if len(keys) != 3 {
		t.Errorf("expected 3 orgs, got %d", len(keys))
	}

	expectedOrgs := map[string]bool{
		"alpha": true,
		"beta":  true,
		"zebra": true,
	}

	for _, key := range keys {
		if !expectedOrgs[key] {
			t.Errorf("unexpected org: %s", key)
		}
	}
}

func TestPrintTree_Format(t *testing.T) {
	repos := []string{
		"github.com/org1/repo1",
		"github.com/org1/repo2",
		"github.com/org2/repo3",
	}

	root := buildTree(repos)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printTree(root, "", false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check for ASCII tree characters
	if !strings.Contains(output, "├──") {
		t.Error("expected output to contain ├──")
	}
	if !strings.Contains(output, "└──") {
		t.Error("expected output to contain └──")
	}
	if !strings.Contains(output, "│") {
		t.Error("expected output to contain │")
	}

	// Check structure elements are present
	expectedElements := []string{
		"github.com",
		"org1",
		"org2",
		"repo1",
		"repo2",
		"repo3",
	}

	for _, elem := range expectedElements {
		if !strings.Contains(output, elem) {
			t.Errorf("expected output to contain %s", elem)
		}
	}

	// Verify alphabetical ordering (org1 before org2)
	org1Index := strings.Index(output, "org1")
	org2Index := strings.Index(output, "org2")
	if org1Index == -1 || org2Index == -1 {
		t.Fatal("could not find org1 or org2 in output")
	}
	if org1Index >= org2Index {
		t.Error("expected org1 to appear before org2")
	}

	// Verify repo ordering under org1 (repo1 before repo2)
	repo1Index := strings.Index(output, "repo1")
	repo2Index := strings.Index(output, "repo2")
	if repo1Index == -1 || repo2Index == -1 {
		t.Fatal("could not find repo1 or repo2 in output")
	}
	if repo1Index >= repo2Index {
		t.Error("expected repo1 to appear before repo2")
	}
}

func TestBuildTree_Empty(t *testing.T) {
	repos := []string{}

	root := buildTree(repos)

	if root == nil {
		t.Fatal("expected non-nil root")
	}

	if len(root.children) != 0 {
		t.Errorf("expected empty root, got %d children", len(root.children))
	}
}

func TestPrintTree_EmptyNode(t *testing.T) {
	root := newTreeNode("")

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printTree(root, "", false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Empty tree should produce no output
	if output != "" {
		t.Errorf("expected empty output, got: %s", output)
	}
}

func TestTreeNode_Creation(t *testing.T) {
	node := newTreeNode("test")

	if node.name != "test" {
		t.Errorf("expected name 'test', got '%s'", node.name)
	}

	if node.children == nil {
		t.Error("expected non-nil children map")
	}

	if len(node.children) != 0 {
		t.Errorf("expected empty children map, got %d children", len(node.children))
	}
}

func TestPrintTree_ComplexHierarchy(t *testing.T) {
	// Test that the tree renders correctly with proper indentation
	repos := []string{
		"source1.com/org1/repo1",
		"source1.com/org1/repo2",
		"source2.com/org2/repo3",
	}

	root := buildTree(repos)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printTree(root, "", false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Expected structure (alphabetically sorted):
	// ├── source1.com
	// │   └── org1
	// │       ├── repo1
	// │       └── repo2
	// └── source2.com
	//     └── org2
	//         └── repo3

	if len(lines) != 7 {
		t.Logf("Output:\n%s", output)
		t.Errorf("expected 7 lines, got %d", len(lines))
	}

	// Check that source1 comes before source2
	source1Line := -1
	source2Line := -1
	for i, line := range lines {
		if strings.Contains(line, "source1.com") {
			source1Line = i
		}
		if strings.Contains(line, "source2.com") {
			source2Line = i
		}
	}

	if source1Line == -1 || source2Line == -1 {
		t.Fatal("could not find source lines")
	}

	if source1Line >= source2Line {
		t.Error("expected source1.com to appear before source2.com")
	}
}

// Example test for the full command (integration style)
func ExampleprintTree() {
	repos := []string{
		"github.com/dsaiztc/dev",
		"github.com/dsaiztc/dotfiles",
	}

	root := buildTree(repos)
	fmt.Println("~/src/")
	printTree(root, "", false)

	// Output:
	// ~/src/
	// └── github.com
	//     └── dsaiztc
	//         ├── dev
	//         └── dotfiles
}
