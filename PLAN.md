# Short-Term Implementation Plan

This file tracks the immediate next tasks for the go-jf-org project. Once a task is completed, it should be checked off. When all tasks are complete, this file should be deleted.

## Phase 1: Artwork Downloads for Media Files

**Status:** ✅ COMPLETE  
**Priority:** Medium  
**Completed:** 2025-12-09

### Objective
Implement artwork downloading functionality to automatically fetch and save poster images, album covers, and book covers alongside organized media files.

### Tasks

- [x] **Research and Design** (0.5 days)
  - [x] Review TMDB API documentation for image downloads
  - [x] Review Cover Art Archive (MusicBrainz) API for album artwork
  - [x] Review OpenLibrary API for book covers
  - [x] Design artwork downloader architecture and interfaces
  - [x] Plan file storage structure for artwork

- [x] **Create Core Artwork Package** (0.5 days)
  - [x] Create `internal/artwork/` package structure
  - [x] Define `Downloader` interface
  - [x] Implement base downloader with common HTTP logic
  - [x] Add error handling and retry mechanism
  - [x] Implement artwork caching to avoid re-downloads

- [x] **TMDB Artwork Downloader** (0.5 days)
  - [x] Implement movie poster downloader
  - [x] Implement TV show poster downloader
  - [x] Add backdrop/fanart support
  - [x] Handle multiple image sizes (w500, w780, original)
  - [x] Add unit tests with >80% coverage

- [x] **Cover Art Archive Integration** (0.5 days)
  - [x] Implement album cover downloader
  - [x] Handle MusicBrainz release IDs
  - [x] Add proper rate limiting (1 req/s for MusicBrainz)
  - [x] Add fallback for missing artwork (graceful failure)
  - [x] Add unit tests

- [x] **OpenLibrary Book Cover Downloader** (0.5 days)
  - [x] Implement book cover downloader
  - [x] Handle ISBN and OpenLibrary IDs
  - [x] Support different cover sizes (S, M, L)
  - [x] Add error handling for missing covers
  - [x] Add unit tests

- [x] **CLI Integration** (0.5 days)
  - [x] Add `--download-artwork` flag to organize command
  - [x] Add `--artwork-size` flag for image size preference
  - [x] Integrate artwork download into organization workflow
  - [x] Add dry-run support for artwork downloads
  - [x] Add progress indicators for artwork downloads (via transaction logs)

- [x] **Testing and Documentation** (0.5 days)
  - [x] Write integration tests for full workflow
  - [x] Test error cases and network failures (via existing downloader tests)
  - [x] Create docs/artwork-downloads.md guide
  - [x] Update README with artwork examples
  - [x] Update STATUS.md to mark task complete

### Acceptance Criteria

- [x] All media types (movies, TV, music, books) support artwork downloads
- [x] Artwork is saved in Jellyfin-compatible locations and formats
- [x] `--download-artwork` flag works with organize command
- [x] Dry-run mode shows which artwork would be downloaded
- [x] All tests pass with >80% code coverage
- [x] Documentation is complete and accurate
- [x] No regressions in existing functionality

### Implementation Summary

**Completed Features:**
- Full integration of artwork downloads into the organizer workflow
- Support for all media types (movies, TV, music, books)
- Dry-run mode shows planned artwork downloads
- Transaction logging for artwork operations (supports rollback)
- Graceful error handling - organization continues even if artwork fails
- Comprehensive test coverage for new functionality
- Complete user documentation in docs/artwork-downloads.md

**Architecture:**
- `internal/artwork/`: Core downloader implementations
  - `downloader.go`: Base HTTP download functionality with retry logic
  - `tmdb.go`: TMDB poster/backdrop downloads
  - `coverart.go`: MusicBrainz cover art downloads
  - `openlibrary.go`: Book cover downloads
- `internal/organizer/organizer.go`: Integrated `downloadArtworkForPlan()` method
- `cmd/organize.go`: CLI flag handling and configuration

**Implementation Notes:**
- Movies: Downloads poster.jpg and backdrop.jpg from TMDB
- TV Shows: Downloads poster.jpg to show directory (not season-specific yet)
- Music: Downloads cover.jpg using MusicBrainz release ID
- Books: Downloads cover.jpg using ISBN
- All artwork operations logged to transaction for rollback support
- Rate limiting respected (1 req/s for MusicBrainz, 40 req/10s for TMDB)

### Known Limitations

1. No season-specific posters for TV shows (only show-level)
2. No episode thumbnails
3. Individual artwork download progress not displayed (overall operation progress shown)

These limitations are documented and may be addressed in future phases.

---

**✅ PHASE 1 COMPLETE - This file can now be deleted as all tasks are finished.**

