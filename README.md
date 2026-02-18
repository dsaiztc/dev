# dev

A CLI tool that reduces cognitive load when navigating between development projects. It enforces an opinionated directory structure (`~/src/<source>/<org>/<project>`) and provides fast navigation.

## Install

### Homebrew

```bash
brew tap dsaiztc/tap
brew install dev
```

### From source

```bash
go install github.com/dsaiztc/dev@latest
```

### For development

See [Development](#development).

## Upgrade

### Homebrew

```bash
brew upgrade dev
```

### From source

```bash
go install github.com/dsaiztc/dev@latest
```

## Setup

Add to your `~/.zshrc` or `~/.bashrc`:

```bash
eval "$(dev init)"
```

## Commands

### `dev clone <url>`

Clones a git repository into `~/src/<source>/<org>/<project>`.

```bash
dev clone git@github.com:dsaiztc/dotfiles.git
# → clones to ~/src/github.com/dsaiztc/dotfiles
```

Supports SSH, HTTPS, and `ssh://` URLs. If the repo is already cloned, it prints the path and exits.

### `dev cd [query]`

Navigates to a project directory.

```bash
dev cd kafka        # fuzzy matches → cd ~/src/github.com/apache/kafka
dev cd              # opens interactive fuzzy finder
```

### `dev init`

Prints the shell wrapper function. The wrapper intercepts `cd` and `clone` to eval their stdout, enabling actual directory changes in the parent shell.

## Development

### Prerequisites

- [Go](https://go.dev/dl/) 1.25+

### Setup

Clone the repo into the expected directory structure and install dependencies:

```bash
git clone git@github.com:dsaiztc/dev.git ~/src/github.com/dsaiztc/dev
cd ~/src/github.com/dsaiztc/dev
go mod download
```

### Running tests

```bash
go test ./...
```

### Build and run

Install the binary from local source to `$GOPATH/bin`:

```bash
go install .
```

After that, `dev` in your shell reflects the local code. Re-run `go install .` after each change to rebuild.

To run a command without installing:

```bash
go run . <command> [args]   # e.g. go run . clone git@github.com:foo/bar.git
```

Note: commands that depend on the shell wrapper (`dev cd`, `dev clone`) need the full installed binary to work correctly via `eval "$(dev init)"`.

### Contributing

1. Fork the repo and create a branch
2. Make changes and add tests where appropriate
3. Run `go test ./...` to verify everything passes
4. Open a pull request

## Releasing

Releases are automated with [GoReleaser](https://goreleaser.com/) and GitHub Actions. Pushing a tag triggers a build that cross-compiles binaries, creates a GitHub Release, and updates the Homebrew formula.

```bash
git tag v0.2.0
git push origin v0.2.0
```

### Setup (one-time)

1. Create a `dsaiztc/homebrew-tap` repo on GitHub
2. Create a fine-grained Personal Access Token with Contents read/write access to `homebrew-tap`
3. Add the token as a repository secret named `HOMEBREW_TAP_TOKEN` in `dsaiztc/dev`

## Directory Structure

All repositories are organized under `~/src/`:

```
~/src/
  github.com/
    dsaiztc/
      dev/
      dotfiles/
    apache/
      kafka/
  gitlab.com/
    team/
      service/
```
