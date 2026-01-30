# Contributing to Go SDK

## Development Environment
1. Install [Go](https://go.dev/doc/install).
2. Clone the repository: `git clone https://github.com/username/repo.git`.
3. Install dependencies: `go mod download`.

## Running Tests
- Run all tests: `go test ./...`
- Run with race detection: `go test -race ./...`
- Run with coverage: `go test -cover ./...`

## Code Style
- All code must be formatted with `gofmt`.
- We use `golangci-lint` for static analysis. Run it locally using `golangci-lint run`.

## Pull Request Process
1. Fork the repository and create your branch from `main`.
2. Ensure all tests pass and linting is clean.
3. Update the `CHANGELOG.md` under the `[Unreleased]` section.
4. Submit the PR for review.

## Releasing

We use Git tags for versioning. Tags follow [semantic versioning](https://semver.org/) (e.g., `v1.2.3`).

### Creating a Release
```bash
make release VERSION=v0.1.0
```

This creates a Git tag and pushes it to GitHub.

### Rolling Back a Release
If you need to delete a tag (e.g., tagged the wrong commit):

```bash
make delete-tag VERSION=v0.1.0
```

> **⚠️ Warning:** If the tag has already been fetched by Go's module proxy, users may still have the old version cached. In that case, release a new patch version instead of deleting.