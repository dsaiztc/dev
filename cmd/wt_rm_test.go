package cmd

import (
	"bytes"
	"testing"
)

func TestConfirmRemoval(t *testing.T) {
	tests := []struct {
		name       string
		skipPrompt bool
		wantOK     bool
		wantErr    bool
	}{
		{
			name:       "yes flag skips prompt and proceeds",
			skipPrompt: true,
			wantOK:     true,
		},
		{
			name:       "no TTY and no --yes returns error",
			skipPrompt: false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			// /dev/tty is not available in test environments, so the
			// non-interactive path always triggers when skipPrompt=false.
			ok, err := confirmRemoval(tt.skipPrompt, "remove? [y/N] ", &stderr)
			if (err != nil) != tt.wantErr {
				t.Errorf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
		})
	}
}

func TestWtRmHasYesFlag(t *testing.T) {
	f := wtRmCmd.Flags().Lookup("yes")
	if f == nil {
		t.Fatal("--yes flag not registered on wt rm")
	}
	if f.Shorthand != "y" {
		t.Errorf("shorthand = %q, want %q", f.Shorthand, "y")
	}

	fForce := wtRmCmd.Flags().Lookup("force")
	if fForce == nil {
		t.Fatal("--force flag not registered on wt rm")
	}
}

func TestWtRmIsAnnotatedDestructive(t *testing.T) {
	if wtRmCmd.Annotations["destructive"] != "true" {
		t.Error(`wtRmCmd.Annotations["destructive"] != "true"`)
	}
}
