package plugin

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func createExe(t *testing.T, dir, name string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func createNonExe(t *testing.T, dir, name string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("not executable"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDiscoverFindsPlugins(t *testing.T) {
	dir := t.TempDir()
	createExe(t, dir, "dev-ll")
	createExe(t, dir, "dev-hello")
	createNonExe(t, dir, "dev-noexec")
	createExe(t, dir, "other-tool")
	createExe(t, dir, "dev-") // empty name after prefix

	plugins := DiscoverFromPATH(dir)

	if len(plugins) != 2 {
		t.Fatalf("expected 2 plugins, got %d: %v", len(plugins), plugins)
	}
	if plugins[0].Name != "hello" {
		t.Errorf("expected first plugin 'hello', got %q", plugins[0].Name)
	}
	if plugins[1].Name != "ll" {
		t.Errorf("expected second plugin 'll', got %q", plugins[1].Name)
	}
	if plugins[0].Path != filepath.Join(dir, "dev-hello") {
		t.Errorf("unexpected path: %s", plugins[0].Path)
	}
}

func TestDiscoverSkipsDirectories(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "dev-subdir"), 0o755); err != nil {
		t.Fatal(err)
	}
	createExe(t, dir, "dev-real")

	plugins := DiscoverFromPATH(dir)

	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d: %v", len(plugins), plugins)
	}
	if plugins[0].Name != "real" {
		t.Errorf("expected 'real', got %q", plugins[0].Name)
	}
}

func TestDiscoverDeduplication(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	createExe(t, dir1, "dev-foo")
	createExe(t, dir2, "dev-foo")

	pathEnv := dir1 + string(os.PathListSeparator) + dir2
	plugins := DiscoverFromPATH(pathEnv)

	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}
	if plugins[0].Path != filepath.Join(dir1, "dev-foo") {
		t.Errorf("expected first PATH entry to win, got %s", plugins[0].Path)
	}
}

func TestDiscoverMergesMultipleDirs(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	createExe(t, dir1, "dev-alpha")
	createExe(t, dir2, "dev-beta")

	pathEnv := dir1 + string(os.PathListSeparator) + dir2
	plugins := DiscoverFromPATH(pathEnv)

	if len(plugins) != 2 {
		t.Fatalf("expected 2 plugins, got %d: %v", len(plugins), plugins)
	}
	// Sorted by name
	if plugins[0].Name != "alpha" || plugins[1].Name != "beta" {
		t.Errorf("expected [alpha, beta], got [%s, %s]", plugins[0].Name, plugins[1].Name)
	}
}

func TestDiscoverEmptyPATH(t *testing.T) {
	plugins := DiscoverFromPATH("")
	if plugins != nil {
		t.Errorf("expected nil for empty PATH, got %v", plugins)
	}
}

func TestDiscoverMissingDirectory(t *testing.T) {
	plugins := DiscoverFromPATH("/nonexistent/path/that/should/not/exist")
	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins for missing dir, got %d", len(plugins))
	}
}

func TestRunPassesArgsAndEnv(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "dev-test")
	output := filepath.Join(dir, "output.txt")

	// Script writes args and env vars to a file so we can inspect them.
	err := os.WriteFile(script, []byte("#!/bin/sh\n"+
		"echo \"ARGS=$*\" > "+output+"\n"+
		"echo \"DEV_ROOT=$DEV_ROOT\" >> "+output+"\n"+
		"echo \"DEV_CWD=$DEV_CWD\" >> "+output+"\n",
	), 0o755)
	if err != nil {
		t.Fatal(err)
	}

	p := Plugin{Name: "test", Path: script}
	if err := Run(p, []string{"--foo", "bar"}); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("could not read output: %v", err)
	}
	got := string(data)

	if !strings.Contains(got, "ARGS=--foo bar") {
		t.Errorf("args not passed through, got: %s", got)
	}

	homeDir, _ := os.UserHomeDir()
	expectedRoot := "DEV_ROOT=" + filepath.Join(homeDir, "src")
	if !strings.Contains(got, expectedRoot) {
		t.Errorf("expected %s in output, got: %s", expectedRoot, got)
	}

	cwd, _ := os.Getwd()
	expectedCwd := "DEV_CWD=" + cwd
	if !strings.Contains(got, expectedCwd) {
		t.Errorf("expected %s in output, got: %s", expectedCwd, got)
	}
}

func TestRunExitCode(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "dev-fail")
	err := os.WriteFile(script, []byte("#!/bin/sh\nexit 42\n"), 0o755)
	if err != nil {
		t.Fatal(err)
	}

	p := Plugin{Name: "fail", Path: script}
	err = Run(p, nil)
	if err == nil {
		t.Fatal("expected error from non-zero exit")
	}

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected *exec.ExitError, got %T: %v", err, err)
	}
	if exitErr.ExitCode() != 42 {
		t.Errorf("expected exit code 42, got %d", exitErr.ExitCode())
	}
}

