package repourl

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    RepoPath
		wantErr bool
	}{
		{
			name:  "SSH SCP-style",
			input: "git@github.com:dsaiztc/dev.git",
			want:  RepoPath{Source: "github.com", Org: "dsaiztc", Project: "dev"},
		},
		{
			name:  "HTTPS with .git",
			input: "https://github.com/dsaiztc/dev.git",
			want:  RepoPath{Source: "github.com", Org: "dsaiztc", Project: "dev"},
		},
		{
			name:  "HTTPS without .git",
			input: "https://github.com/dsaiztc/dev",
			want:  RepoPath{Source: "github.com", Org: "dsaiztc", Project: "dev"},
		},
		{
			name:  "SSH with scheme",
			input: "ssh://git@github.com/dsaiztc/dev.git",
			want:  RepoPath{Source: "github.com", Org: "dsaiztc", Project: "dev"},
		},
		{
			name:  "SSH with port",
			input: "ssh://git@github.com:22/dsaiztc/dev.git",
			want:  RepoPath{Source: "github.com", Org: "dsaiztc", Project: "dev"},
		},
		{
			name:  "GitLab nested groups",
			input: "git@gitlab.com:gitlab-org/subgroup/project.git",
			want:  RepoPath{Source: "gitlab.com", Org: "gitlab-org/subgroup", Project: "project"},
		},
		{
			name:  "HTTPS GitLab nested groups",
			input: "https://gitlab.com/gitlab-org/subgroup/deep/project.git",
			want:  RepoPath{Source: "gitlab.com", Org: "gitlab-org/subgroup/deep", Project: "project"},
		},
		{
			name:  "Custom host",
			input: "git@git.company.com:team/service.git",
			want:  RepoPath{Source: "git.company.com", Org: "team", Project: "service"},
		},
		{
			name:    "empty URL",
			input:   "",
			wantErr: true,
		},
		{
			name:    "no org/project",
			input:   "https://github.com/",
			wantErr: true,
		},
		{
			name:    "only project, no org",
			input:   "https://github.com/project",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse(%q) expected error, got %+v", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("Parse(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}

func TestRepoPath_FullPath(t *testing.T) {
	rp := RepoPath{Source: "github.com", Org: "dsaiztc", Project: "dev"}
	want := "github.com/dsaiztc/dev"
	if got := rp.FullPath(); got != want {
		t.Errorf("FullPath() = %q, want %q", got, want)
	}
}
