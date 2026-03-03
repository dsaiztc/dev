package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds user defaults for the dev CLI.
type Config struct {
	DefaultSource string `json:"default_source"`
	DefaultOrg    string `json:"default_org"`
	WorktreeRoot  string `json:"worktree_root,omitempty"`
}

// GetWorktreeRoot returns the configured worktree root or the default ~/src__worktrees.
// Expands a leading ~ to the user's home directory.
func (c *Config) GetWorktreeRoot() string {
	homeDir, _ := os.UserHomeDir()
	if c.WorktreeRoot != "" {
		if strings.HasPrefix(c.WorktreeRoot, "~/") {
			return filepath.Join(homeDir, c.WorktreeRoot[2:])
		}
		return c.WorktreeRoot
	}
	return filepath.Join(homeDir, "src__worktrees")
}

// Path returns the config file path (~/.config/dev/config.json).
func Path() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(homeDir, ".config", "dev", "config.json"), nil
}

// Load reads the config file and returns the parsed Config.
// Returns a wrapped os.ErrNotExist if the file does not exist.
func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}
	return LoadFrom(path)
}

// LoadFrom reads a config from the given path.
func LoadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("could not parse config: %w", err)
	}
	return &cfg, nil
}

// Save writes the config to the default config file path.
func Save(cfg *Config) error {
	path, err := Path()
	if err != nil {
		return err
	}
	return SaveTo(cfg, path)
}

// SaveTo writes the config to the given path, creating parent directories as needed.
func SaveTo(cfg *Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal config: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("could not write config: %w", err)
	}
	return nil
}
