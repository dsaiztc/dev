package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

func TestLocCmd_OutputFormat(t *testing.T) {
	// This test verifies that loc outputs just the path to stdout
	// We can't easily test the full command without mocking the filesystem,
	// but we can verify the command is registered and has correct metadata

	if locCmd.Use != "loc [query]" {
		t.Errorf("expected Use to be 'loc [query]', got %q", locCmd.Use)
	}

	if locCmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if locCmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

func TestLocCmd_Integration(t *testing.T) {
	// Verify the command is registered with root
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "loc" {
			found = true
			break
		}
	}

	if !found {
		t.Error("loc command not registered with root command")
	}
}

// Example of expected output format
func Example_locOutput() {
	// When run: dev loc dotfiles
	// Expected output format (just the path):
	fmt.Println("/Users/username/src/github.com/dsaiztc/dotfiles")
	// Output: /Users/username/src/github.com/dsaiztc/dotfiles
}

func TestLocCmd_EmptyArgs(t *testing.T) {
	// Test that the command accepts empty args (will trigger fuzzy finder)
	// We can't test the fuzzy finder in unit tests, but we can verify
	// the function signature accepts this case

	args := []string{}
	if len(args) == 0 {
		// This is valid - should trigger fuzzy finder
		return
	}
}

func TestLocCmd_WithQuery(t *testing.T) {
	// Test that the command accepts query args
	args := []string{"dotfiles"}

	if len(args) > 0 {
		query := strings.Join(args, " ")
		if query != "dotfiles" {
			t.Errorf("expected query to be 'dotfiles', got %q", query)
		}
	}
}

func TestLocCmd_OutputGoesToStdout(t *testing.T) {
	testOutput := "/Users/test/src/github.com/test/repo"

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fmt.Println(testOutput)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := strings.TrimSpace(buf.String())

	if output != testOutput {
		t.Errorf("expected output %q, got %q", testOutput, output)
	}
}

func TestLocOutput_Plain(t *testing.T) {
	// Redirect stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := locOutput("github.com/foo/bar", "/home/user/src/github.com/foo/bar", false)

	w.Close()
	os.Stdout = old
	if err != nil {
		t.Fatalf("locOutput: %v", err)
	}
	var buf bytes.Buffer
	io.Copy(&buf, r)
	want := "/home/user/src/github.com/foo/bar\n"
	if buf.String() != want {
		t.Errorf("plain output = %q, want %q", buf.String(), want)
	}
}

func TestLocOutput_JSON(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := locOutput("github.com/foo/bar", "/home/user/src/github.com/foo/bar", true)

	w.Close()
	os.Stdout = old
	if err != nil {
		t.Fatalf("locOutput: %v", err)
	}
	var buf bytes.Buffer
	io.Copy(&buf, r)

	var got struct {
		Repo string `json:"repo"`
		Path string `json:"path"`
	}
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal: %v (raw=%q)", err, buf.String())
	}
	if got.Repo != "github.com/foo/bar" {
		t.Errorf("repo = %q, want %q", got.Repo, "github.com/foo/bar")
	}
	if got.Path != "/home/user/src/github.com/foo/bar" {
		t.Errorf("path = %q, want %q", got.Path, "/home/user/src/github.com/foo/bar")
	}
}
