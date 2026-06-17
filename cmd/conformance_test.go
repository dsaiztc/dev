package cmd

// TestAIFirstConformance is the AI-first CLI conformance harness.
// It walks every command in the tree and asserts the standards we've committed
// to. Failures here mean a new or modified command regressed against the
// checklist from steipete's create-cli skill and clig.dev.
//
// To register a command as destructive, add:
//
//	Annotations: map[string]string{"destructive": "true"}
//
// That annotation is the trigger for requiring a --yes flag.

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func init() {
	// Ensure version is set for conformance tests; in production main() does this.
	if rootCmd.Version == "" {
		SetVersion("test")
	}
}

// allCommands returns all commands in the tree (root + every descendant).
func allCommands(root *cobra.Command) []*cobra.Command {
	cmds := []*cobra.Command{root}
	for _, c := range root.Commands() {
		cmds = append(cmds, allCommands(c)...)
	}
	return cmds
}

func TestAIFirstConformance_HelpWorksForEveryCommand(t *testing.T) {
	t.Helper()
	for _, cmd := range allCommands(rootCmd) {
		t.Run(cmd.CommandPath(), func(t *testing.T) {
			// --help must not error and must include "Usage:" in output.
			// We use a fresh output buffer rather than running a subprocess.
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Cobra's help function writes help and returns; it never errors.
			cmd.HelpFunc()(cmd, []string{})

			out := buf.String()
			if !strings.Contains(out, "Usage:") {
				t.Errorf("--help output for %q does not contain 'Usage:':\n%s", cmd.CommandPath(), out)
			}
		})
	}
}

func TestAIFirstConformance_RootHasVersion(t *testing.T) {
	if rootCmd.Version == "" {
		t.Error("rootCmd.Version is empty: wire SetVersion() in main.go or tests")
	}
}

func TestAIFirstConformance_RootHasPersistentFlags(t *testing.T) {
	required := []string{"no-input", "no-color"}
	for _, name := range required {
		if rootCmd.PersistentFlags().Lookup(name) == nil {
			t.Errorf("rootCmd is missing persistent flag --%s", name)
		}
	}
}

func TestAIFirstConformance_DestructiveCommandsHaveYesFlag(t *testing.T) {
	for _, cmd := range allCommands(rootCmd) {
		if cmd.Annotations["destructive"] != "true" {
			continue
		}
		f := cmd.Flags().Lookup("yes")
		if f == nil {
			t.Errorf("command %q is annotated destructive but has no --yes flag", cmd.CommandPath())
		}
	}
}

