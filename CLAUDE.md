# dev

Go CLI tool for project navigation. Uses Cobra for commands, Bubbletea for the fuzzy finder TUI.

## Docs

Detailed guides live in `docs/`, each with front-matter describing when to read it. Read the relevant doc before doing the work it covers:

- [`docs/releasing.md`](docs/releasing.md) — **read before cutting a release** (tagging a version, editing release notes, "ship it"). Covers the release checklist, semver choice, and the requirement to update the README to match the release before tagging.

## Releasing

When cutting a release, follow [`docs/releasing.md`](docs/releasing.md). Key rule: **update the README to reflect the release before tagging** — it must always describe the released version.
