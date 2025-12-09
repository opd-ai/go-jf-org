# Short-Term Implementation Plan

This file tracks the immediate next tasks for the go-jf-org project. Once a task is completed, it should be checked off. When all tasks are complete, this file should be deleted.

## Phase 1: Artwork Downloads for Media Files

**Status:** In Progress - Core downloaders complete  
**Priority:** Medium  
**Estimated Effort:** 2-3 days

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
  - [ ] Add fallback for missing artwork
  - [x] Add unit tests

- [x] **OpenLibrary Book Cover Downloader** (0.5 days)
  - [x] Implement book cover downloader
  - [x] Handle ISBN and OpenLibrary IDs
  - [x] Support different cover sizes (S, M, L)
  - [x] Add error handling for missing covers
  - [x] Add unit tests

- [ ] **CLI Integration** (0.5 days)
  - [ ] Add `--download-artwork` flag to organize command
  - [ ] Add `--artwork-size` flag for image size preference
  - [ ] Integrate artwork download into organization workflow
  - [ ] Add dry-run support for artwork downloads
  - [ ] Add progress indicators for artwork downloads

- [ ] **Testing and Documentation** (0.5 days)
  - [ ] Write integration tests for full workflow
  - [ ] Test error cases and network failures
  - [ ] Create docs/artwork-downloads.md guide
  - [ ] Update README with artwork examples
  - [ ] Update STATUS.md to mark task complete

### Acceptance Criteria

- [ ] All media types (movies, TV, music, books) support artwork downloads
- [ ] Artwork is saved in Jellyfin-compatible locations and formats
- [ ] `--download-artwork` flag works with organize command
- [ ] Dry-run mode shows which artwork would be downloaded
- [ ] All tests pass with >80% code coverage
- [ ] Documentation is complete and accurate
- [ ] No regressions in existing functionality

### Implementation Notes

**Artwork File Locations (Jellyfin Standard):**
- Movies: `Movie Name (Year)/poster.jpg`, `Movie Name (Year)/backdrop.jpg`
- TV Shows: `Show Name/poster.jpg`, `Show Name/Season 01/poster.jpg`
- Music: `Artist/Album (Year)/cover.jpg`
- Books: `Author Last, First/Book Title (Year)/cover.jpg`

**Image Preferences:**
- Default size: Medium/Large (balance quality and disk space)
- Support configurable sizes via `--artwork-size` flag
- Cache downloaded images to avoid re-downloading

**Error Handling:**
- Gracefully handle missing artwork (log warning, don't fail)
- Retry failed downloads up to 3 times with exponential backoff
- Continue organization even if artwork download fails
