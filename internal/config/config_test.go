package config

import (
	"errors"
	"os"
	"path/filepath"
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
