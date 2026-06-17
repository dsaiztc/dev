package cmd

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestResolveConfig_NonInteractiveUsesDefaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home) // config.Load uses os.UserHomeDir() → $HOME
	var stderr bytes.Buffer
	src, org, err := resolveConfig(home, true, "", "", &stderr)
	if err != nil {
		t.Fatalf("resolveConfig: %v", err)
	}
	if src != "github.com" {
		t.Errorf("source = %q, want %q", src, "github.com")
	}
	wantOrg := filepath.Base(home)
	if org != wantOrg {
		t.Errorf("org = %q, want %q", org, wantOrg)
	}
	// Should mention defaults on stderr
	if stderr.Len() == 0 {
		t.Error("expected stderr to describe defaults being used")
	}
}

func TestResolveConfig_FlagOverrides(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	var stderr bytes.Buffer
	src, org, err := resolveConfig(home, true, "gitlab.com", "myteam", &stderr)
	if err != nil {
		t.Fatalf("resolveConfig: %v", err)
	}
	if src != "gitlab.com" {
		t.Errorf("source = %q, want %q", src, "gitlab.com")
	}
	if org != "myteam" {
		t.Errorf("org = %q, want %q", org, "myteam")
	}
}
