package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	want := &Config{
		DefaultSource: "github.com",
		DefaultOrg:    "testuser",
	}

	if err := SaveTo(want, path); err != nil {
		t.Fatalf("SaveTo: %v", err)
	}

	got, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}

	if got.DefaultSource != want.DefaultSource || got.DefaultOrg != want.DefaultOrg {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, want)
	}
}

func TestLoadNotExist(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "does-not-exist.json")

	_, err := LoadFrom(path)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected os.ErrNotExist, got %v", err)
	}
}

func TestGetWorktreeRoot_Custom(t *testing.T) {
	cfg := &Config{WorktreeRoot: "/custom/worktrees"}
	got := cfg.GetWorktreeRoot()
	if got != "/custom/worktrees" {
		t.Errorf("GetWorktreeRoot() = %q, want %q", got, "/custom/worktrees")
	}
}

func TestGetWorktreeRoot_TildeExpansion(t *testing.T) {
	cfg := &Config{WorktreeRoot: "~/my_worktrees"}
	got := cfg.GetWorktreeRoot()
	homeDir, _ := os.UserHomeDir()
	want := filepath.Join(homeDir, "my_worktrees")
	if got != want {
		t.Errorf("GetWorktreeRoot() = %q, want %q", got, want)
	}
}

func TestGetWorktreeRoot_Default(t *testing.T) {
	cfg := &Config{}
	got := cfg.GetWorktreeRoot()
	homeDir, _ := os.UserHomeDir()
	want := filepath.Join(homeDir, "src__worktrees")
	if got != want {
		t.Errorf("GetWorktreeRoot() = %q, want %q", got, want)
	}
}

func TestWorktreeRootRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	want := &Config{
		DefaultSource: "github.com",
		DefaultOrg:    "testuser",
		WorktreeRoot:  "/my/worktrees",
	}

	if err := SaveTo(want, path); err != nil {
		t.Fatalf("SaveTo: %v", err)
	}

	got, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}

	if got.WorktreeRoot != want.WorktreeRoot {
		t.Errorf("WorktreeRoot = %q, want %q", got.WorktreeRoot, want.WorktreeRoot)
	}
}

func TestWorktreeRootOmittedWhenEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := &Config{DefaultSource: "github.com", DefaultOrg: "testuser"}
	if err := SaveTo(cfg, path); err != nil {
		t.Fatalf("SaveTo: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if strings.Contains(string(data), "worktree_root") {
		t.Errorf("expected worktree_root to be omitted from JSON, got: %s", data)
	}
}

func TestSaveCreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "dir", "config.json")

	cfg := &Config{DefaultSource: "gitlab.com", DefaultOrg: "team"}
	if err := SaveTo(cfg, path); err != nil {
		t.Fatalf("SaveTo: %v", err)
	}

	got, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if got.DefaultSource != "gitlab.com" || got.DefaultOrg != "team" {
		t.Errorf("unexpected config: %+v", got)
	}
}
