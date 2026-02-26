package plugin

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const prefix = "dev-"

// Plugin represents an external command found on PATH.
type Plugin struct {
	Name string // subcommand name (e.g. "ll" for dev-ll)
	Path string // absolute path to the executable
}

// Discover scans PATH for executables matching dev-* and returns them sorted
// by name. The first match in PATH wins when duplicates exist.
func Discover() []Plugin {
	return DiscoverFromPATH(os.Getenv("PATH"))
}

// DiscoverFromPATH is like Discover but takes an explicit PATH string.
func DiscoverFromPATH(pathEnv string) []Plugin {
	if pathEnv == "" {
		return nil
	}

	seen := make(map[string]bool)
	var plugins []Plugin

	for _, dir := range filepath.SplitList(pathEnv) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if !strings.HasPrefix(name, prefix) {
				continue
			}
			subName := name[len(prefix):]
			if subName == "" {
				continue
			}
			if seen[subName] {
				continue
			}

			fullPath := filepath.Join(dir, name)
			info, err := e.Info()
			if err != nil {
				continue
			}
			if info.Mode()&0o111 == 0 {
				continue
			}

			seen[subName] = true
			plugins = append(plugins, Plugin{Name: subName, Path: fullPath})
		}
	}

	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Name < plugins[j].Name
	})
	return plugins
}

// Run executes the plugin as a child process, inheriting stdin/stdout/stderr.
// It sets DEV_ROOT and DEV_CWD environment variables.
func Run(p Plugin, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not determine working directory: %w", err)
	}

	cmd := exec.Command(p.Path, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		"DEV_ROOT="+filepath.Join(homeDir, "src"),
		"DEV_CWD="+cwd,
	)
	return cmd.Run()
}
