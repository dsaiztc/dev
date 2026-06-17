---
description: >
  Read this when cutting a release of the dev CLI — whenever the user says
  "cut a release", "ship it", "tag a version", or asks to update release
  notes. Covers the full release checklist, semver choice, and the
  requirement to keep the README in sync with the released version.
read_when: cutting a release, tagging vX.Y.Z, editing release notes, bumping the version
---

# Releasing

Releases are automated with [GoReleaser](https://goreleaser.com/) and GitHub
Actions: pushing a `v*` tag triggers a build that cross-compiles binaries,
creates a GitHub Release, and updates the Homebrew tap.

## Checklist

1. **Update the README first.** Before tagging, make sure the README reflects
   everything new/changed/removed in this release — new commands, flags,
   behavior changes, removed features. The README must always describe the
   version being released. (See also: keep `--help` text accurate; the
   conformance test in `cmd/conformance_test.go` guards `--help` existence but
   not its content.)
2. Commit all changes and push to `main`.
3. Pick the version per [semver](https://semver.org/):
   - **patch** (`vX.Y.Z`) — bug fixes only, no behavior changes
   - **minor** (`vX.Y.0`) — new features, backwards-compatible
   - **major** (`vX.0.0`) — breaking changes
4. Tag the release: `git tag vX.Y.Z`
5. Push the tag: `git push origin vX.Y.Z`
6. Wait for the GitHub Actions release workflow to finish (builds binaries,
   updates the Homebrew tap). Check it:
   ```bash
   gh run list --workflow=release.yml --limit=1
   gh run watch <run-id> --exit-status   # block until done
   ```
7. Edit the release notes with a summary of the changes:
   ```bash
   gh release edit vX.Y.Z --notes "..."
   ```

## Release notes format

Follow [Keep a Changelog](https://keepachangelog.com/) conventions. Start with a
one-line summary, then use `###` sections as applicable:

- **Added** — new features
- **Changed** — changes to existing functionality
- **Deprecated** — features that will be removed
- **Removed** — features that were removed
- **Fixed** — bug fixes
- **Security** — vulnerability fixes

Only include sections that apply to the release. If a release changes the shell
wrapper (`internal/shell/shell.go`), note in the release that users should
reload it with `eval "$(dev init)"`.
