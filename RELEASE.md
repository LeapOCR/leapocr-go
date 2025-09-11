# Release Process

This document outlines the process for creating a new release of the LeapOCR Go SDK.

## Versioning

This project follows [Semantic Versioning (SemVer)](https://semver.org/). When creating a new release, it's important to choose the correct version number based on the changes being introduced.

- **`MAJOR` version (e.g., `1.2.3` -> `2.0.0`)**: For incompatible API changes.
- **`MINOR` version (e.g., `1.2.3` -> `1.3.0`)**: For adding functionality in a backward-compatible manner.
- **`PATCH` version (e.g., `1.2.3` -> `1.2.4`)**: For backward-compatible bug fixes.

## Release Process

Follow these steps to create a new release:

### Step 1: Update the Version

1.  **Choose the new version number** based on the changes you've made.
2.  **Update the `version.go` file** with the new version number. For example, to release version `v0.1.0`, change the file to:

    ```go
    package ocr

    const Version = "0.1.0"
    ```

### Step 2: Commit the Version Change

Commit the change to `version.go` and push it to the `main` branch.

```bash
git add version.go
git commit -m "chore: bump version to 0.1.0"
git push origin main
```

### Step 3: Create and Push a Git Tag

This is the most important step. Go modules are versioned using git tags. To publish a new version, you need to create a git tag with the version number (prefixed with `v`) and push it to the repository.

```bash
# Create the tag
git tag v0.1.0

# Push the tag to GitHub
git push origin v0.1.0
```

### Step 4: Verify the Release

Pushing a new tag will trigger the `release.yml` GitHub Actions workflow. This workflow will automatically:

1.  Run the tests.
2.  Create a new GitHub release with an automated changelog.
3.  Publish the Go documentation.

You can monitor the progress of the workflow on the "Actions" tab of the GitHub repository. Once the workflow is complete, you should see a new release on the "Releases" page.

## Post-Release

Once the release is published, the Go module proxy will automatically pick up the new version. Users will then be able to get the new version of the SDK by running:

```bash
go get github.com/leapocr/go-sdk@v0.1.0
```

Or, to get the latest version:

```bash
go get -u github.com/leapocr/go-sdk
```
