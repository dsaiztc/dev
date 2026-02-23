package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateProject_NewDir(t *testing.T) {
	home := t.TempDir()
	var stdout, stderr bytes.Buffer

	err := createProject(home, "github.com", "testuser", "my-project", &stdout, &stderr)
	if err != nil {
		t.Fatalf("createProject: %v", err)
	}

	targetDir := filepath.Join(home, "src", "github.com", "testuser", "my-project")

	// Directory should exist
	info, err := os.Stat(targetDir)
	if err != nil {
		t.Fatalf("expected directory to exist: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("expected a directory")
	}

	// Should have been git init'd
	gitDir := filepath.Join(targetDir, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		t.Errorf("expected .git directory: %v", err)
	}

	// stdout should contain the cd command
	wantCD := "cd " + targetDir + "\n"
	if stdout.String() != wantCD {
		t.Errorf("stdout = %q, want %q", stdout.String(), wantCD)
	}

	// stderr should mention "created"
	if !strings.Contains(stderr.String(), "created") {
		t.Errorf("stderr = %q, expected it to contain 'created'", stderr.String())
	}
}

func TestCreateProject_ExistingDir(t *testing.T) {
	home := t.TempDir()
	targetDir := filepath.Join(home, "src", "github.com", "testuser", "existing")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := createProject(home, "github.com", "testuser", "existing", &stdout, &stderr)
	if err != nil {
		t.Fatalf("createProject: %v", err)
	}

	// stdout should still have the cd command
	wantCD := "cd " + targetDir + "\n"
	if stdout.String() != wantCD {
		t.Errorf("stdout = %q, want %q", stdout.String(), wantCD)
	}

	// stderr should mention "already exists"
	if !strings.Contains(stderr.String(), "already exists") {
		t.Errorf("stderr = %q, expected it to contain 'already exists'", stderr.String())
	}

	// Should NOT have .git (we didn't init it, and existing dirs are left alone)
	gitDir := filepath.Join(targetDir, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		t.Error("expected no .git directory for pre-existing dir")
	}
}

func TestCreateProject_NestedPath(t *testing.T) {
	home := t.TempDir()
	var stdout, stderr bytes.Buffer

	err := createProject(home, "gitlab.com", "myteam", "deep-project", &stdout, &stderr)
	if err != nil {
		t.Fatalf("createProject: %v", err)
	}

	targetDir := filepath.Join(home, "src", "gitlab.com", "myteam", "deep-project")
	if _, err := os.Stat(targetDir); err != nil {
		t.Fatalf("expected directory to exist: %v", err)
	}
}
