package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestWtAliasResolves(t *testing.T) {
	// "dev wkt rm" must still route to the rm subcommand via the alias.
	c, _, err := rootCmd.Find([]string{"wkt", "rm"})
	if err != nil {
		t.Fatalf("Find([wkt rm]): %v", err)
	}
	if c.Name() != "rm" {
		t.Errorf("resolved command = %q, want %q", c.Name(), "rm")
	}
	if c.Parent() != wtCmd {
		t.Errorf("rm parent = %v, want wtCmd", c.Parent())
	}
}

func TestWtDefaultsToCd(t *testing.T) {
	// "dev wt" with no subcommand should be runnable (defaults to cd).
	if wtCmd.RunE == nil {
		t.Fatal("wtCmd.RunE is nil, expected it to default to cd")
	}
}

func TestWarnDeprecatedAlias(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantWarn bool
	}{
		{"bare wkt", []string{"wkt"}, true},
		{"wkt subcommand", []string{"wkt", "rm"}, true},
		{"wkt with leading flag", []string{"--no-color", "wkt", "cd"}, true},
		{"wt is fine", []string{"wt", "rm"}, false},
		{"other command", []string{"cd"}, false},
		{"no args", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			warnDeprecatedAlias(tt.args, &buf)
			gotWarn := strings.Contains(buf.String(), "deprecated")
			if gotWarn != tt.wantWarn {
				t.Errorf("warn = %v (out=%q), want %v", gotWarn, buf.String(), tt.wantWarn)
			}
		})
	}
}
