# Contributing to go-jf-org

Thank you for your interest in contributing to go-jf-org! This document provides guidelines and information for contributors.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)

## Getting Started

Before you begin:

1. Read the [Implementation Plan](IMPLEMENTATION_PLAN.md) to understand the project architecture
2. Check [existing issues](https://github.com/opd-ai/go-jf-org/issues) to see what needs work
3. Read the [Jellyfin Conventions](docs/jellyfin-conventions.md) to understand naming standards

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Git
- Make (optional, but recommended)

### Clone and Build

```bash
# Clone the repository
git clone https://github.com/opd-ai/go-jf-org.git
cd go-jf-org

# Install dependencies
go mod download

# Build the project
make build

# Run tests
make test

# Run the binary
./bin/go-jf-org
```

## Project Structure

```
go-jf-org/
├── cmd/                    # CLI commands (Cobra-based)
├── internal/               # Internal packages
│   ├── scanner/           # File system scanning
│   ├── detector/          # Media type detection
│   ├── metadata/          # Metadata extraction
│   ├── organizer/         # File organization
│   ├── safety/            # Safety mechanisms
│   ├── jellyfin/          # Jellyfin-specific logic
│   └── config/            # Configuration management
├── pkg/                   # Public packages
│   └── types/            # Shared types
├── test/                  # Integration tests
├── docs/                  # Documentation
└── main.go               # Application entry point
```

## Development Workflow

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
```

Use descriptive branch names:
- `feature/add-music-support` for new features
- `fix/movie-parser-bug` for bug fixes
- `docs/update-readme` for documentation

### 2. Make Changes

- Write clean, idiomatic Go code
- Follow the coding standards (see below)
- Add tests for new functionality
- Update documentation as needed

### 3. Test Your Changes

```bash
# Run all tests
make test

# Run specific package tests
go test ./internal/scanner/...

# Run with coverage
go test -cover ./...

# Build to ensure no compilation errors
make build
```

### 4. Commit Your Changes

Write clear, descriptive commit messages:

```bash
git commit -m "Add movie filename parser

- Parse movie titles with year
- Extract quality and source info
- Add tests for edge cases"
```

### 5. Push and Create PR

```bash
git push origin feature/your-feature-name
```

Then create a Pull Request on GitHub.

## Coding Standards

### Go Code Style

Follow standard Go conventions:

- Use `gofmt` to format code (run `make fmt`)
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use meaningful variable and function names
- Add comments for exported functions and types

Example:

```go
// ParseMovieFilename extracts metadata from a movie filename.
// It returns the title, year, and quality information.
//
// Example:
//   ParseMovieFilename("The.Matrix.1999.1080p.BluRay.mkv")
//   // Returns: title="The Matrix", year=1999, quality="1080p"
func ParseMovieFilename(filename string) (*MovieInfo, error) {
    // Implementation
}
```

### Error Handling

Always handle errors explicitly:

```go
// Good
file, err := os.Open(path)
if err != nil {
    return fmt.Errorf("failed to open file %s: %w", path, err)
}
defer file.Close()

// Bad
file, _ := os.Open(path)  // Never ignore errors
```

### Logging

Use structured logging:

```go
log.Info().
    Str("file", filename).
    Str("type", "movie").
    Msg("Processing media file")
```

## Testing

### Unit Tests

Write unit tests for all new functions:

```go
func TestParseMovieFilename(t *testing.T) {
    tests := []struct {
        name     string
        filename string
        want     *MovieInfo
        wantErr  bool
    }{
        {
            name:     "standard format",
            filename: "The.Matrix.1999.1080p.BluRay.mkv",
            want: &MovieInfo{
                Title:   "The Matrix",
                Year:    1999,
                Quality: "1080p",
            },
            wantErr: false,
        },
        // Add more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseMovieFilename(tt.filename)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseMovieFilename() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("ParseMovieFilename() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Integration Tests

Place integration tests in `test/integration/`:

```go
func TestFullOrganizationWorkflow(t *testing.T) {
    // Setup test environment
    // Run organization
    // Verify results
}
```

### Test Coverage

Aim for:
- Minimum 80% code coverage
- 100% coverage for critical safety code

Check coverage:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Or use the Makefile target
make coverage
```

## CI/CD Pipeline

### Automated Checks

Every pull request triggers automated CI/CD pipelines:

1. **Test Suite** - Runs on Linux, macOS, and Windows with multiple Go versions
2. **Linting** - Code quality checks with golangci-lint
3. **Build** - Multi-platform builds (Linux, macOS, Windows on amd64/arm64)
4. **Coverage** - Code coverage reporting and tracking

### Running CI Checks Locally

Before pushing, run the same checks that CI will run:

```bash
# Run all CI checks
make ci

# Individual checks
make test      # Run tests
make lint      # Run linter
make build     # Build binary
make coverage  # Generate coverage report
```

### GitHub Actions Workflows

The project uses the following workflows:

- **CI** (`.github/workflows/ci.yml`) - Runs tests, linting, builds, and coverage reporting
- **Release** (`.github/workflows/release.yml`) - Automated releases on tags

### Release Process

Releases are automated via GitHub Actions:

1. Update version in `main.go`
2. Commit and push changes
3. Create and push a version tag:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```
4. GitHub Actions automatically:
   - Runs full test suite
   - Builds multi-platform binaries
   - Creates compressed archives
   - Generates checksums
   - Creates GitHub release with artifacts

## Pull Request Process

### Before Submitting

- [ ] Code builds without errors (`make build`)
- [ ] All tests pass (`make test`)
- [ ] Code is formatted (`make fmt`)
- [ ] Linter passes (`make lint`)
- [ ] Documentation updated
- [ ] Commit messages are clear

### PR Guidelines

1. **Title**: Clear, concise description
   - Good: "Add TMDB API client for movie metadata"
   - Bad: "Fixed stuff"

2. **Description**: Explain what and why
   - What changes were made
   - Why they were necessary
   - Link to related issues

3. **Size**: Keep PRs focused
   - One feature/fix per PR
   - Break large changes into smaller PRs

4. **Tests**: Include relevant tests
   - Unit tests for new functions
   - Integration tests for workflows

### Review Process

1. Automated checks must pass (CI/CD)
2. At least one maintainer approval required
3. Address review comments
4. Maintainer will merge when ready

## Areas Where Help Is Needed

See [Implementation Plan](IMPLEMENTATION_PLAN.md) for detailed phases.

Current priorities:

### Phase 1: Foundation
- [ ] CLI framework with Cobra
- [ ] Configuration management with Viper
- [ ] File system scanner
- [ ] Logging infrastructure

### Phase 2: Metadata Extraction
- [ ] Filename parsers (movies, TV, music, books)
- [ ] Media type detector
- [ ] TMDB API client
- [ ] MusicBrainz API client
- [ ] OpenLibrary API client

### Phase 3: File Organization
- [ ] Jellyfin naming convention implementation
- [ ] NFO file generator
- [ ] File mover/organizer
- [ ] Conflict resolution

### Phase 4: Safety
- [ ] Transaction logging
- [ ] Rollback mechanism
- [ ] Validation checks

## Getting Help

- **Questions**: Open a [Discussion](https://github.com/opd-ai/go-jf-org/discussions)
- **Bugs**: Open an [Issue](https://github.com/opd-ai/go-jf-org/issues)
- **Ideas**: Open a [Feature Request](https://github.com/opd-ai/go-jf-org/issues/new)

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers
- Focus on constructive feedback
- Help others learn

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

**Thank you for contributing to go-jf-org!** 🎉
