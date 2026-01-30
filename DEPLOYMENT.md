# Deployment Guide

## Prerequisites
- Go 1.21 or higher installed.
- GitHub account with write access to the repository.
- A valid `GITHUB_TOKEN` for automated releases.

## Version Management
This project follows [Semantic Versioning (SemVer)](https://semver.org/).
- Major (v1.0.0): Breaking changes.
- Minor (v0.1.0): New features, non-breaking.
- Patch (v0.0.1): Bug fixes.

## Step-by-Step Publish
1. **Update Changelog**: Ensure `CHANGELOG.md` is updated with the new version.
2. **Tag the Release**: 
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```
3. **Verify Module**: Ensure the module is available via the Go proxy:
   ```bash
   GOPROXY=proxy.golang.org go list -m github.com/username/repo@v1.0.0
   ```

## Troubleshooting
- **Cache Issues**: If the version doesn't appear on pkg.go.dev, use the `go list` command above to trigger a crawl.
- **Tag Mismatch**: Ensure the tag in Git matches the version string in any internal constants.