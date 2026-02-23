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

	cfg, err := config.Load()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("could not load config: %w", err)
		}
		// Config doesn't exist — prompt user for defaults
		cfg, err = promptForConfig(homeDir)
		if err != nil {
			return err
		}
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("could not save config: %w", err)
		}
		configPath, _ := config.Path()
		fmt.Fprintf(os.Stderr, "config saved to %s\n", configPath)
	}

	source, _ := cmd.Flags().GetString("source")
	if source == "" {
		source = cfg.DefaultSource
	}
	org, _ := cmd.Flags().GetString("org")
	if org == "" {
		org = cfg.DefaultOrg
	}

	return createProject(homeDir, source, org, name, os.Stdout, os.Stderr)
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

func promptForConfig(homeDir string) (*config.Config, error) {
	// Read from /dev/tty since stdin is captured by the $() subshell
	tty, err := os.Open("/dev/tty")
	if err != nil {
		return nil, fmt.Errorf("could not open /dev/tty: %w", err)
	}
	defer tty.Close()

	reader := bufio.NewReader(tty)

	fmt.Fprint(os.Stderr, "No config found. Let's set up your defaults.\n")

	defaultOrg := filepath.Base(homeDir)

	fmt.Fprint(os.Stderr, "Default source [github.com]: ")
	source, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("could not read input: %w", err)
	}
	source = strings.TrimSpace(source)
	if source == "" {
		source = "github.com"
	}

	fmt.Fprintf(os.Stderr, "Default org [%s]: ", defaultOrg)
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
