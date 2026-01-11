# Release Process

This project uses [GoReleaser](https://goreleaser.com/) and GitHub Actions to automate releases.

## Creating a New Release

To create a new release, you simply need to push a valid tag to the repository.

1.  **Ensure you are on the `main` branch and valid state:**
    ```bash
    git checkout main
    git pull
    ```

2.  **Tag the release:**
    Use semantic versioning (e.g., `v1.0.0`, `v1.0.1`, `v1.1.0`).
    ```bash
    git tag -a v1.0.0 -m "Release v1.0.0"
    ```

3.  **Push the tag:**
    ```bash
    git push origin v1.0.0
    ```

## What Happens Next?

The GitHub Action defined in `.github/workflows/release.yml` will trigger:
1.  Check out the code.
2.  Install Go.
3.  Run GoReleaser.
4.  GoReleaser will:
    -   Build binaries for Linux, macOS, and Windows.
    -   Create archives.
    -   Generate checksums.
    -   Create a GitHub Release with the artifacts and changelog.

## Verifying the Release

After pushing the tag, visit the [GitHub Actions](https://github.com/maltehedderich/rename-ai/actions) page to see the workflow progress. Once completed, the new release will appear in the [Releases](https://github.com/maltehedderich/rename-ai/releases) section.
