package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

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

	if err := rootCmd.Execute(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
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
