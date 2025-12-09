# Short-Term Implementation Plan

This file tracks the immediate next tasks for the go-jf-org project. Once a task is completed, it should be checked off. When all tasks are complete, this file should be deleted.

## Phase 1: Artwork Downloads for Media Files

**Status:** Not Started  
**Priority:** Medium  
**Estimated Effort:** 2-3 days

### Objective
Implement artwork downloading functionality to automatically fetch and save poster images, album covers, and book covers alongside organized media files.

### Tasks

- [ ] **Research and Design** (0.5 days)
  - [ ] Review TMDB API documentation for image downloads
  - [ ] Review Cover Art Archive (MusicBrainz) API for album artwork
  - [ ] Review OpenLibrary API for book covers
  - [ ] Design artwork downloader architecture and interfaces
  - [ ] Plan file storage structure for artwork

- [ ] **Create Core Artwork Package** (0.5 days)
  - [ ] Create `internal/artwork/` package structure
  - [ ] Define `Downloader` interface
  - [ ] Implement base downloader with common HTTP logic
  - [ ] Add error handling and retry mechanism
  - [ ] Implement artwork caching to avoid re-downloads

- [ ] **TMDB Artwork Downloader** (0.5 days)
  - [ ] Implement movie poster downloader
  - [ ] Implement TV show poster downloader
  - [ ] Add backdrop/fanart support
  - [ ] Handle multiple image sizes (w500, w780, original)
  - [ ] Add unit tests with >80% coverage

- [ ] **Cover Art Archive Integration** (0.5 days)
  - [ ] Implement album cover downloader
  - [ ] Handle MusicBrainz release IDs
  - [ ] Add proper rate limiting (1 req/s for MusicBrainz)
  - [ ] Add fallback for missing artwork
  - [ ] Add unit tests

- [ ] **OpenLibrary Book Cover Downloader** (0.5 days)
  - [ ] Implement book cover downloader
  - [ ] Handle ISBN and OpenLibrary IDs
  - [ ] Support different cover sizes (S, M, L)
  - [ ] Add error handling for missing covers
  - [ ] Add unit tests

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
