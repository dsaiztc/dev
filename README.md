# dev

A CLI tool that reduces cognitive load when navigating between development projects. It enforces an opinionated directory structure (`~/src/<source>/<org>/<project>`) and provides fast navigation.

## Install

```bash
go install github.com/dsaiztc/dev@latest
```

Then add the shell wrapper to your `~/.zshrc` or `~/.bashrc`:

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
