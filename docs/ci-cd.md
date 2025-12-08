# CI/CD Pipeline Documentation

## Overview

The go-jf-org project uses GitHub Actions for continuous integration and deployment. The pipeline ensures code quality, runs comprehensive tests across multiple platforms, and automates the release process.

## Workflows

### 1. CI Workflow (`.github/workflows/ci.yml`)

Runs on every push and pull request to `main` and `develop` branches.

**Jobs:**

#### Test Job
- **Platforms:** Ubuntu, macOS, Windows
- **Go Versions:** 1.21, 1.22, 1.23
- **Actions:**
  - Checkout code
  - Setup Go environment
  - Download and verify dependencies
  - Run tests with race detection and coverage
  - Upload coverage to Codecov (Ubuntu + Go 1.23 only)

#### Lint Job
- **Platform:** Ubuntu (latest)
- **Go Version:** 1.23
- **Actions:**
  - Checkout code
  - Setup Go environment
  - Run golangci-lint with comprehensive checks

#### Build Job
- **Platform:** Ubuntu (latest)
- **Go Version:** 1.23
- **Target Platforms:**
  - Linux (amd64, arm64)
  - macOS/Darwin (amd64, arm64)
  - Windows (amd64)
- **Actions:**
  - Build binaries for all platforms
  - Upload artifacts for download

### 2. Release Workflow (`.github/workflows/release.yml`)

Triggered when a version tag (e.g., `v1.0.0`) is pushed.

**Actions:**
- Checkout code with full history
- Run full test suite
- Build binaries for all platforms:
  - Linux: amd64, arm64
  - macOS: amd64 (Intel), arm64 (Apple Silicon)
  - Windows: amd64
- Create compressed archives (.tar.gz for Unix, .zip for Windows)
- Generate SHA256 checksums
- Generate changelog from git commits
- Create GitHub release with all artifacts

## Linter Configuration

The project uses golangci-lint with the following enabled linters:

### Code Quality
- `errcheck` - Checks for unchecked errors
- `gosimple` - Simplifies code
- `govet` - Standard Go vet checks
- `ineffassign` - Detects ineffectual assignments
- `staticcheck` - Advanced static analysis
- `unused` - Detects unused code
- `misspell` - Spell checking

### Best Practices
- `revive` - Configurable linting rules
- `gosec` - Security-focused checks
- `bodyclose` - HTTP response body closure
- `noctx` - Context usage in HTTP requests
- `unconvert` - Unnecessary type conversions

See `.golangci.yml` for full configuration.

## Running Locally

### Prerequisites

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Verify installation
golangci-lint --version
```

### Commands

```bash
# Run all CI checks (recommended before pushing)
make ci

# Individual checks
make test       # Run tests with race detection and coverage
make lint       # Run golangci-lint
make build      # Build for current platform
make coverage   # Generate coverage report

# Multi-platform builds
make build-all  # Build for all platforms

# Create release archives
make release    # Build all platforms + create archives
```

## Coverage Requirements

- **Overall:** Minimum 80% coverage
- **Critical Code:** 100% coverage for safety mechanisms
  - Transaction logging
  - Rollback functionality
  - File validation

Current coverage: **>85%** across all packages

## Build Artifacts

### CI Builds
Artifacts are available for download from the Actions tab for 7 days:
- `go-jf-org-linux-amd64`
- `go-jf-org-linux-arm64`
- `go-jf-org-darwin-amd64`
- `go-jf-org-darwin-arm64`
- `go-jf-org-windows-amd64`

### Release Builds
Official releases include:
- Compressed archives (.tar.gz, .zip)
- SHA256 checksums (`checksums.txt`)
- Generated changelog
- Pre-release flag for alpha/beta/rc versions

## Badge Status

The README includes the following CI/CD badges:

- **CI Status** - Overall CI pipeline status
- **Code Coverage** - Current test coverage percentage
- **Go Report Card** - Overall code quality grade
- **Latest Release** - Current stable version

## Troubleshooting

### Failed Tests

If tests fail in CI but pass locally:
1. Check the Go version (CI tests on 1.21, 1.22, 1.23)
2. Run with race detector: `go test -race ./...`
3. Check platform-specific issues (Windows vs Unix)

### Linter Failures

If linter passes locally but fails in CI:
1. Ensure you're using the latest golangci-lint version
2. Run with same timeout: `golangci-lint run --timeout=5m`
3. Check `.golangci.yml` for disabled checks

### Build Failures

If builds fail for specific platforms:
1. Use `GOOS` and `GOARCH` environment variables locally:
   ```bash
   GOOS=windows GOARCH=amd64 go build .
   ```
2. Check for platform-specific code or dependencies
3. Ensure no CGO dependencies for cross-compilation

## Security

### Secrets Management
- No hardcoded API keys or credentials
- Use GitHub Secrets for sensitive data
- GITHUB_TOKEN is automatically provided

### Dependency Scanning
- Go module verification in CI
- Regular dependency updates
- Security scanning via gosec linter

## Future Enhancements

Potential additions to the CI/CD pipeline:

- [ ] Automated dependency updates (Dependabot)
- [ ] Performance regression testing
- [ ] Integration test suite
- [ ] Deployment to package registries (Homebrew, apt, etc.)
- [ ] Automated changelog generation (git-chglog)
- [ ] Semantic versioning enforcement
- [ ] Docker image builds and publishing

## Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [golangci-lint Linters](https://golangci-lint.run/usage/linters/)
- [Codecov Documentation](https://docs.codecov.com/)
- [Go Testing Best Practices](https://golang.org/doc/code.html#Testing)
