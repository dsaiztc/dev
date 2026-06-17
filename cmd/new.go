package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dsaiztc/dev/internal/config"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Create a new project directory under ~/src/<source>/<org>/<name>",
	Args:  cobra.ExactArgs(1),
	RunE:  runNew,
}

func init() {
	newCmd.Flags().String("source", "", "override default source (e.g. github.com)")
	newCmd.Flags().String("org", "", "override default org (e.g. dsaiztc)")
	rootCmd.AddCommand(newCmd)
}

func runNew(cmd *cobra.Command, args []string) error {
	name := args[0]

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	source, _ := cmd.Flags().GetString("source")
	org, _ := cmd.Flags().GetString("org")

	resolvedSource, resolvedOrg, err := resolveConfig(homeDir, noInput(cmd), source, org, os.Stderr)
	if err != nil {
		return err
	}

	return createProject(homeDir, resolvedSource, resolvedOrg, name, os.Stdout, os.Stderr)
}

// resolveConfig determines the source and org to use. It loads config if
// present; if not, it either prompts interactively or falls back to defaults
// when non-interactive. flagSource/flagOrg override whatever config provides.
func resolveConfig(homeDir string, nonInteractive bool, flagSource, flagOrg string, stderr io.Writer) (source, org string, err error) {
	cfg, loadErr := config.Load()
	if loadErr != nil && !errors.Is(loadErr, os.ErrNotExist) {
		return "", "", fmt.Errorf("could not load config: %w", loadErr)
	}

	if errors.Is(loadErr, os.ErrNotExist) {
		// No config — derive or prompt
		defaultSource := "github.com"
		defaultOrg := filepath.Base(homeDir)

		if nonInteractive || !isTTYAvailable() {
			fmt.Fprintf(stderr, "no config found; using defaults (source=%s, org=%s)\n", defaultSource, defaultOrg)
			cfg = &config.Config{DefaultSource: defaultSource, DefaultOrg: defaultOrg}
		} else {
			cfg, err = promptForConfig(homeDir, stderr)
			if err != nil {
				return "", "", err
			}
			if err := config.Save(cfg); err != nil {
				return "", "", fmt.Errorf("could not save config: %w", err)
			}
			configPath, _ := config.Path()
			fmt.Fprintf(stderr, "config saved to %s\n", configPath)
		}
	}

	source = cfg.DefaultSource
	if flagSource != "" {
		source = flagSource
	}
	org = cfg.DefaultOrg
	if flagOrg != "" {
		org = flagOrg
	}
	return source, org, nil
}

func isTTYAvailable() bool {
	tty, err := os.Open("/dev/tty")
	if err != nil {
		return false
	}
	tty.Close()
	return true
}

// createProject creates the project directory, runs git init, and prints the
// cd command to stdout. It is extracted from runNew for testability.
func createProject(homeDir, source, org, name string, stdout, stderr io.Writer) error {
	targetDir := filepath.Join(homeDir, "src", source, org, name)

	if info, err := os.Stat(targetDir); err == nil && info.IsDir() {
		fmt.Fprintf(stderr, "already exists: %s\n", targetDir)
	} else {
		if err := os.MkdirAll(targetDir, 0o755); err != nil {
			return fmt.Errorf("could not create directory: %w", err)
		}
		fmt.Fprintf(stderr, "created %s\n", targetDir)

		gitCmd := exec.Command("git", "init", targetDir)
		gitCmd.Stdout = stderr
		gitCmd.Stderr = stderr
		if err := gitCmd.Run(); err != nil {
			return fmt.Errorf("git init failed: %w", err)
		}
	}

	fmt.Fprintf(stdout, "cd %s\n", targetDir)
	return nil
}

func promptForConfig(homeDir string, stderr io.Writer) (*config.Config, error) {
	// Read from /dev/tty since stdin is captured by the $() subshell
	tty, err := os.Open("/dev/tty")
	if err != nil {
		return nil, fmt.Errorf("could not open /dev/tty: %w", err)
	}
	defer tty.Close()

	reader := bufio.NewReader(tty)

	fmt.Fprint(stderr, "No config found. Let's set up your defaults.\n")

	defaultOrg := filepath.Base(homeDir)

	fmt.Fprint(stderr, "Default source [github.com]: ")
	source, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("could not read input: %w", err)
	}
	source = strings.TrimSpace(source)
	if source == "" {
		source = "github.com"
	}

	fmt.Fprintf(stderr, "Default org [%s]: ", defaultOrg)
	org, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("could not read input: %w", err)
	}
	org = strings.TrimSpace(org)
	if org == "" {
		org = defaultOrg
	}

	return &config.Config{
		DefaultSource: source,
		DefaultOrg:    org,
	}, nil
}
