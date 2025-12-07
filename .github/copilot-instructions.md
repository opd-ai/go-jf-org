# Project Overview

`go-jf-org` is a safe, powerful Go CLI tool designed to organize disorganized media collections (movies, TV shows, music, and books) into a clean, Jellyfin-compatible directory structure. The tool extracts metadata from filenames and file content, enriches it using external APIs (TMDB, MusicBrainz, OpenLibrary), and safely moves files without ever deleting anything.

The project prioritizes safety through transaction logging and rollback support, allowing users to preview changes before execution and undo operations if needed. It's built for the Jellyfin community to simplify media library management while maintaining strict adherence to Jellyfin's naming conventions and NFO file requirements.

The target audience includes Jellyfin server administrators and home media enthusiasts who need to organize large collections of unsorted media files. The tool handles edge cases like duplicate files, missing metadata, and special characters while maintaining a minimal configuration approach with sensible defaults.

## Technical Stack

- **Primary Language**: Go 1.24.10
- **CLI Framework**: Cobra v1.10.2 for command-line interface with subcommands (scan, organize, preview, verify, rollback)
- **Configuration**: Viper v1.21.0 for YAML/ENV/flag-based configuration management with hot reload support
- **Logging**: zerolog v1.34.0 for structured, zero-allocation JSON logging with console-friendly output
- **Testing**: Go's built-in testing package with table-driven tests and subtests using `t.Run`
- **Build/Deploy**: Makefile-based build system (`make build`, `make test`, `make lint`), binaries output to `bin/` directory

## Code Assistance Guidelines

1. **Safety-First Development**: Never implement file deletion operations. All file operations must be moves or renames only. Every operation must be logged to a transaction file before execution to enable rollback. Use validation checks (file exists, destination writable, sufficient disk space) before any file operation. Implement dry-run mode for all commands that modify files.

2. **Jellyfin Naming Conventions**: Strictly follow Jellyfin naming patterns - Movies: `Movie Name (Year).ext` in `Movie Name (Year)/` directory; TV Shows: `Show Name - S##E## - Episode Title.ext` in `Show Name/Season ##/`; Music: `## - Track.ext` in `Artist/Album (Year)/`; Books: `Book Title.ext` in `Author Last, First/Book Title (Year)/`. Generate Kodi-compatible NFO XML files for all media types. Sanitize filenames by replacing `<>:"/\|?*` characters and removing leading/trailing dots and spaces.

3. **Table-Driven Testing**: Write all tests using table-driven approach with `t.Run` for subtests. Use `os.MkdirTemp` for temporary directories in tests, never hardcode paths. Test edge cases including missing years, special characters, duplicates, and API failures. Aim for >80% code coverage with 100% coverage for safety-critical code (transaction logging, rollback, validation).

4. **Structured Logging with zerolog**: Use structured logging for all operations: `log.Info().Str("file", filename).Str("type", "movie").Msg("Processing media file")`. Log at appropriate levels: Debug for detailed flow, Info for user-visible operations, Warn for recoverable issues, Error for failures. Include context fields (file paths, operation types, transaction IDs) in all log entries.

5. **CLI Command Structure**: Each command should be in a separate file in `cmd/` directory following the existing pattern (scan.go, root.go). Use Cobra's `PersistentPreRun` for configuration loading and logging setup. Validate arguments using `cobra.ExactArgs` or similar validators. Provide clear error messages that guide users to solutions.

6. **Configuration Management**: Load configuration from `~/.go-jf-org/config.yaml` with environment variable and command-line flag overrides via Viper. Use sensible defaults defined in `internal/config/defaults.go`. Support minimal configuration - tool should work out-of-the-box for common use cases. Validate configuration values at load time with clear error messages.

7. **External API Integration**: Implement rate limiting for all external APIs (TMDB: 40 requests/10s, MusicBrainz: 1 request/s). Cache API responses with 24-hour TTL to reduce unnecessary requests. Handle API failures gracefully - fall back to filename-only metadata if APIs unavailable. Use stdlib `net/http` client with reasonable timeouts (10s for metadata APIs).

## Project Context

- **Domain**: Media library organization with focus on Jellyfin media server compatibility. Key concepts include metadata extraction (from filenames using regex patterns like `(.+)[\. ](\d{4})` for movies and `S(\d+)E(\d+)` for TV shows), NFO file generation (Kodi-compatible XML), and transaction-based safety mechanisms. The tool handles four media types: movies, TV shows, music, and books, each with distinct naming conventions and metadata sources.

- **Architecture**: Standard Go project layout with `cmd/` for CLI commands, `internal/` for private packages (scanner, detector, metadata, organizer, safety, config), and `pkg/types/` for shared types. Data flow: Scanner → Detector → Metadata → Organizer → Safety. The scanner walks filesystems filtering by extension, detector determines media type, metadata extracts and enriches information, organizer builds target paths and moves files, safety logs operations and enables rollback.

- **Key Directories**: 
  - `cmd/` - Cobra-based CLI commands (root.go, scan.go, organize.go, preview.go, verify.go, rollback.go)
  - `internal/scanner/` - File system traversal and extension filtering
  - `internal/config/` - Viper integration and configuration management
  - `pkg/types/` - Shared data structures (MediaType, Metadata, Operation, Transaction)
  - `docs/` - Project documentation including jellyfin-conventions.md, metadata-sources.md, examples.md
  - `test/` - Integration tests and test fixtures
  - `bin/` - Build output directory (gitignored)

- **Configuration**: Default config location is `~/.go-jf-org/config.yaml`. Example config includes source/destination paths, API keys (optional - uses free tier), organize options (create_nfo, download_artwork, normalize_names), and safety settings (dry_run, transaction_log, conflict_resolution: skip|rename|interactive). Configuration can be overridden via environment variables or command-line flags. Minimal config approach - most users can run without any configuration.

## Quality Standards

- **Testing Requirements**: Maintain >80% code coverage using Go's built-in testing package. Write table-driven tests for all business logic functions. Include integration tests for all CLI commands in `test/integration/`. Test edge cases: missing metadata, special characters in filenames, API failures, disk space issues, concurrent operations. Use `os.MkdirTemp` for test isolation. Run tests with `make test` before committing.

- **Code Review Criteria**: All code must pass `gofmt` formatting (run `make fmt`). Follow [Effective Go](https://golang.org/doc/effective_go.html) conventions. Add godoc comments for all exported functions and types. Handle errors explicitly - never ignore errors with `_`. Use meaningful variable names (no single-letter variables except loop counters). Keep functions focused - maximum 50 lines per function. Ensure builds pass with `make build` and linter passes with `make lint` (golangci-lint).

- **Documentation Standards**: Update relevant documentation when changing functionality. Keep TECHNICAL_SPEC.md in sync with architecture changes. Update examples in docs/ when adding new features. Document all CLI flags in command help text. Add inline comments for complex logic (regex patterns, file path construction, NFO XML generation). Maintain changelog entries for user-visible changes.

## Networking Best Practices (for Go projects)

When declaring network variables, always use interface types:
- Never use `net.UDPAddr`, `net.IPAddr`, or `net.TCPAddr`. Use `net.Addr` only instead.
- Never use `net.UDPConn`, use `net.PacketConn` instead
- Never use `net.TCPConn`, use `net.Conn` instead
- Never use `net.UDPListener` or `net.TCPListener`, use `net.Listener` instead
- Never use a type switch or type assertion to convert from an interface type to a concrete type. Use the interface methods instead.

This approach enhances testability and flexibility when working with different network implementations or mocks.

## Development Workflow

- **Building**: Run `make build` to compile binary to `bin/go-jf-org`. Use `make run` to build and execute. Dependencies are managed with `go mod` - run `make mod` to tidy and verify.

- **Testing**: Run `make test` for all tests. Use `go test ./internal/scanner/...` for package-specific tests. Check coverage with `go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out`. Tests should complete in <5 seconds for unit tests, <30 seconds for integration tests.

- **Linting**: Run `make lint` (requires golangci-lint). Install with `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`. Format code with `make fmt` before committing.

- **Commit Guidelines**: Write clear, descriptive commit messages. Format: "Add movie filename parser" with bullet points for details. Reference issues when applicable. Keep commits focused on single changes.

## Metadata Extraction Patterns

- **Movie Regex**: `(?P<title>.+?)[\. ](?P<year>19\d{2}|20\d{2})[\. ]?(?P<quality>\d{3,4}p)?` - Matches "The.Matrix.1999.1080p.BluRay.mkv"
- **TV Show Regex**: `(?P<show>.+?)[\. ]S(?P<season>\d{2})E(?P<episode>\d{2})` - Matches "Breaking.Bad.S01E01.720p.mkv"
- **Valid Years**: 1850-2100 for movies/TV/books, 1900-current for music
- **Extension Categories**: Video (.mkv, .mp4, .avi, .m4v, .ts, .webm), Audio (.flac, .mp3, .m4a, .ogg, .opus, .wav), Books (.epub, .mobi, .pdf, .azw3, .cbz, .cbr)

## File Organization Safety Rules

- **Never Delete**: All operations must be moves or renames. Implement copy-then-verify-then-delete-source pattern if atomic move not possible.
- **Transaction Logging**: Log operations to `~/.go-jf-org/txn/<transaction-id>.json` before execution. Include operation type, source path, destination path, timestamp.
- **Rollback Support**: Every logged operation must be reversible. Test rollback for all operation types.
- **Conflict Resolution**: Implement three strategies - skip (leave existing file), rename (append -1, -2 suffix), interactive (prompt user). Default to skip for safety.
- **Validation Pipeline**: Check source exists and readable, destination writable, sufficient disk space (file size + 10% buffer), no unsafe characters in target path.
