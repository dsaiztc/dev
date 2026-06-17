package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/dsaiztc/dev/internal/fuzzy"
	"github.com/dsaiztc/dev/internal/plugin"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dev",
	Short: "A CLI tool for managing development projects",
	Long:  `dev reduces cognitive load when navigating between development projects by enforcing an opinionated directory structure (~/src/<source>/<org>/<project>).`,
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		noColor, _ := cmd.Root().PersistentFlags().GetBool("no-color")
		if noColor || os.Getenv("NO_COLOR") != "" {
			fuzzy.DisableColor()
		}
		return nil
	},
}

const pluginGroupID = "plugins"

// SetVersion sets the version string reported by --version.
func SetVersion(v string) {
	rootCmd.Version = v
}

func init() {
	rootCmd.PersistentFlags().Bool("no-input", false, "disable all interactive prompts (for use in scripts/agents)")
	rootCmd.PersistentFlags().Bool("no-color", false, "disable ANSI color output (also honored via NO_COLOR env var)")
}

// noInput returns true when --no-input was passed or when /dev/tty is unavailable.
func noInput(cmd *cobra.Command) bool {
	v, _ := cmd.Root().PersistentFlags().GetBool("no-input")
	return v
}

func Execute() {
	registerPlugins()

	warnDeprecatedAlias(os.Args[1:], os.Stderr)

	if err := rootCmd.Execute(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// warnDeprecatedAlias prints a deprecation notice to stderr when a renamed
// command is invoked via a deprecated alias. It inspects the first non-flag
// argument (the subcommand name) so the warning fires for every subcommand
// (e.g. "dev wkt rm"), not just the bare alias.
func warnDeprecatedAlias(args []string, stderr io.Writer) {
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			continue
		}
		if a == "wkt" {
			fmt.Fprintln(stderr, `dev: "wkt" is deprecated, use "wt" instead`)
		}
		return // first non-flag arg is the subcommand; stop either way
	}
}

func registerPlugins() {
	builtins := make(map[string]bool)
	for _, c := range rootCmd.Commands() {
		builtins[c.Name()] = true
	}

	plugins := plugin.Discover()
	if len(plugins) == 0 {
		return
	}

	rootCmd.AddGroup(&cobra.Group{
		ID:    pluginGroupID,
		Title: "Plugin Commands:",
	})

	for _, p := range plugins {
		if builtins[p.Name] {
			continue
		}
		rootCmd.AddCommand(&cobra.Command{
			Use:                p.Name,
			Short:              "Plugin: " + p.Path,
			GroupID:            pluginGroupID,
			DisableFlagParsing: true,
			SilenceUsage:       true,
			SilenceErrors:      true,
			RunE: func(_ *cobra.Command, args []string) error {
				return plugin.Run(p, args)
			},
		})
	}
}
