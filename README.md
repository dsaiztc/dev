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

Then add to your `~/.zshrc` or `~/.bashrc`:

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
