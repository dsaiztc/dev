package repourl

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

// RepoPath represents the parsed components of a git repository URL.
type RepoPath struct {
	Source  string // e.g. "github.com"
	Org     string // e.g. "dsaiztc" or "gitlab-org/subgroup"
	Project string // e.g. "dev"
}

// FullPath returns the full relative path: source/org/project.
func (r RepoPath) FullPath() string {
	return path.Join(r.Source, r.Org, r.Project)
}

// Parse parses a git URL into its RepoPath components.
// Supports SSH (git@host:org/repo.git), HTTPS (https://host/org/repo.git),
// and SSH with scheme (ssh://git@host/org/repo.git).
func Parse(rawURL string) (RepoPath, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return RepoPath{}, fmt.Errorf("empty URL")
	}

	var host, repoPath string

	if strings.Contains(rawURL, "://") {
		// HTTPS or ssh:// scheme
		u, err := url.Parse(rawURL)
		if err != nil {
			return RepoPath{}, fmt.Errorf("invalid URL: %w", err)
		}
		host = u.Hostname()
		repoPath = strings.TrimPrefix(u.Path, "/")
	} else if strings.Contains(rawURL, ":") {
		// SCP-style SSH: git@host:org/repo.git
		parts := strings.SplitN(rawURL, ":", 2)
		hostPart := parts[0]
		repoPath = parts[1]
		// Strip user@ prefix
		if idx := strings.Index(hostPart, "@"); idx != -1 {
			host = hostPart[idx+1:]
		} else {
			host = hostPart
		}
	} else {
		return RepoPath{}, fmt.Errorf("unrecognized URL format: %s", rawURL)
	}

	if host == "" {
		return RepoPath{}, fmt.Errorf("could not determine host from URL: %s", rawURL)
	}

	// Strip trailing .git
	repoPath = strings.TrimSuffix(repoPath, ".git")
	// Strip leading/trailing slashes
	repoPath = strings.Trim(repoPath, "/")

	if repoPath == "" {
		return RepoPath{}, fmt.Errorf("could not determine repo path from URL: %s", rawURL)
	}

	// Split into segments; last segment is project, everything before is org
	segments := strings.Split(repoPath, "/")
	if len(segments) < 2 {
		return RepoPath{}, fmt.Errorf("URL must contain at least org/project: %s", rawURL)
	}

	project := segments[len(segments)-1]
	org := strings.Join(segments[:len(segments)-1], "/")

	return RepoPath{
		Source:  host,
		Org:     org,
		Project: project,
	}, nil
}
