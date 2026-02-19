# dev

Go CLI tool for project navigation. Uses Cobra for commands, Bubbletea for the fuzzy finder TUI.

## Release workflow

When cutting a new release:

1. Commit all changes and push to main
2. Tag the release: `git tag vX.Y.Z`
3. Push the tag: `git push origin vX.Y.Z`
4. Wait for the GitHub Actions workflow to complete (builds binaries, updates Homebrew tap)
5. Edit the release notes on GitHub with a summary of the changes:
   ```
   gh release edit vX.Y.Z --notes "..."
   ```

## Project page

The GitHub Pages site is served from the `docs/` directory. Update `docs/index.html` when adding new features or making significant changes to keep the landing page current.
