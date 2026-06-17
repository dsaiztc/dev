package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/dsaiztc/dev/internal/plugin"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dev",
	Short: "A CLI tool for managing development projects",
	Long:  `dev reduces cognitive load when navigating between development projects by enforcing an opinionated directory structure (~/src/<source>/<org>/<project>).`,
}

const pluginGroupID = "plugins"

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
