# Implementation Plan for go-jf-org

**Version:** 1.0  
**Date:** 2025-12-07

## Overview

go-jf-org is a Go CLI application designed to organize disorganized media files (movies, TV shows, music, books) into a Jellyfin-compatible directory structure. The application focuses on safety, metadata extraction, and minimal user configuration.

## Core Requirements

- **Never delete files** - only rename and move operations
- Extract metadata from files and filenames
- Inject missing metadata when possible
- Normalize filenames to Jellyfin conventions
- Minimal configuration required from users
- Safe operation with rollback capabilities

---

## 1. Architecture & Packages

### 1.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────┐
│                   CLI Interface                      │
│              (cobra/viper based)                     │
└────────────────┬────────────────────────────────────┘
                 │
┌────────────────┴────────────────────────────────────┐
│              Command Layer                           │
│  (scan, organize, preview, verify, config)          │
└────────────────┬────────────────────────────────────┘
                 │
┌────────────────┴────────────────────────────────────┐
│              Core Business Logic                     │
├─────────────────────────────────────────────────────┤
│  Scanner │ Detector │ Metadata │ Organizer │ Safe   │
└────────────────┬────────────────────────────────────┘
                 │
┌────────────────┴────────────────────────────────────┐
│           External Services & Storage                │
│  (TMDB, MusicBrainz, OpenLibrary, FileSystem)       │
└─────────────────────────────────────────────────────┘
```

### 1.2 Package Structure

```
go-jf-org/
├── cmd/                        # CLI commands
│   ├── root.go                 # Root command setup
│   ├── scan.go                 # Scan media directories
│   ├── organize.go             # Organize files
│   ├── preview.go              # Preview changes without execution
│   ├── verify.go               # Verify organized structure
│   └── config.go               # Configuration management
│
├── internal/
│   ├── scanner/                # File system scanning
│   │   ├── scanner.go          # Main scanner logic
│   │   ├── filter.go           # File filtering by type
│   │   └── walker.go           # Directory traversal
│   │
│   ├── detector/               # Media type detection
│   │   ├── detector.go         # Main detector interface
│   │   ├── movie.go            # Movie detection
│   │   ├── tv.go               # TV show detection
│   │   ├── music.go            # Music detection
│   │   └── book.go             # Book detection
│   │
│   ├── metadata/               # Metadata extraction & enrichment
│   │   ├── extractor.go        # Extract from filename/file
│   │   ├── enricher.go         # Enrich with external APIs
│   │   ├── parser.go           # Parse filenames
│   │   ├── tmdb.go             # TMDB API client
│   │   ├── musicbrainz.go      # MusicBrainz API client
│   │   ├── openlibrary.go      # OpenLibrary API client
│   │   └── models.go           # Metadata models
│   │
│   ├── organizer/              # File organization logic
│   │   ├── organizer.go        # Main organizer
│   │   ├── movie.go            # Movie organization
│   │   ├── tv.go               # TV show organization
│   │   ├── music.go            # Music organization
│   │   ├── book.go             # Book organization
│   │   └── namer.go            # Filename normalization
│   │
│   ├── safety/                 # Safety mechanisms
│   │   ├── transaction.go      # Transaction log
│   │   ├── rollback.go         # Rollback operations
│   │   ├── validator.go        # Pre-operation validation
│   │   └── backup.go           # Backup strategies
│   │
│   ├── jellyfin/               # Jellyfin-specific logic
│   │   ├── naming.go           # Naming conventions
│   │   ├── structure.go        # Directory structure
│   │   └── nfo.go              # NFO file generation
│   │
│   └── config/                 # Configuration management
│       ├── config.go           # Config struct and loading
│       ├── defaults.go         # Default values
│       └── validate.go         # Config validation
│
├── pkg/                        # Public packages (if needed)
│   └── types/                  # Shared types
│       ├── media.go            # Media types
│       └── operation.go        # Operation types
│
├── test/                       # Integration tests
│   ├── fixtures/               # Test fixtures
│   └── integration/            # Integration test suites
│
├── docs/                       # Documentation
│   ├── jellyfin-conventions.md
│   ├── metadata-sources.md
│   └── examples.md
│
├── go.mod
├── go.sum
├── main.go                     # Application entry point
├── Makefile                    # Build automation
└── README.md
```

---

## 2. Directory Structure

### 2.1 Input Structure (Disorganized)

The tool should handle various input structures:

```
/media/unsorted/
├── MovieName.2023.1080p.BluRay.x264.mkv
├── show.s01e01.720p.mkv
├── Artist - Album (2020)/
│   └── 01 - Song.mp3
└── Some.Book.by.Author.epub
```

### 2.2 Output Structure (Jellyfin-Compatible)

#### Movies
```
/media/movies/
├── Movie Name (2023)/
│   ├── Movie Name (2023).mkv
│   ├── movie.nfo
│   ├── poster.jpg
│   └── fanart.jpg
```

#### TV Shows
```
/media/tv/
├── Show Name/
│   ├── Season 01/
│   │   ├── Show Name - S01E01 - Episode Title.mkv
│   │   ├── Show Name - S01E02 - Episode Title.mkv
│   │   └── season.nfo
│   ├── Season 02/
│   └── tvshow.nfo
```

#### Music
```
/media/music/
├── Artist Name/
│   ├── Album Name (2020)/
│   │   ├── 01 - Track Name.mp3
│   │   ├── 02 - Track Name.mp3
│   │   ├── album.nfo
│   │   └── folder.jpg
│   └── artist.nfo
```

#### Books
```
/media/books/
├── Author Name/
│   ├── Book Title (2020)/
│   │   ├── Book Title.epub
│   │   ├── book.nfo
│   │   └── cover.jpg
```

---

## 3. Jellyfin Naming Conventions

### 3.1 Movies

**Format:** `Movie Name (Year).ext`

Examples:
- `The Matrix (1999).mkv`
- `Inception (2010).mp4`

**Directory:** `Movie Name (Year)/`

**Special Cases:**
- Multiple versions: `Movie Name (Year) - 1080p.mkv`, `Movie Name (Year) - 4K.mkv`
- Parts: `Movie Name (Year) - Part 1.mkv`, `Movie Name (Year) - Part 2.mkv`

### 3.2 TV Shows

**Format:** `Show Name - S##E## - Episode Title.ext`

Examples:
- `Breaking Bad - S01E01 - Pilot.mkv`
- `Game of Thrones - S05E09 - The Dance of Dragons.mkv`

**Multi-Episode Format:** `Show Name - S##E##-E## - Episode Title.ext`

Example:
- `Show Name - S01E01-E02 - Double Episode.mkv`

**Directory Structure:**
```
Show Name/
├── Season 01/
├── Season 02/
└── Specials/  (Season 00)
```

### 3.3 Music

**Album Format:**
```
Artist Name/
└── Album Name (Year)/
    └── ## - Track Name.ext
```

**Compilation:**
```
Various Artists/
└── Album Name (Year)/
    └── ## - Artist - Track Name.ext
```

### 3.4 Books

**Format:**
```
Author Name/
└── Book Title (Year)/
    └── Book Title.ext
```

**Series:**
```
Author Name/
└── Series Name/
    ├── 01 - Book Title (Year).ext
    └── 02 - Book Title (Year).ext
```

---

## 4. Metadata Extraction Strategy

### 4.1 Extraction Pipeline

```
File Input
    ↓
┌───────────────────────┐
│ Filename Parsing      │ → Extract: title, year, quality, codec, etc.
└──────────┬────────────┘
           ↓
┌───────────────────────┐
│ File Metadata Reading │ → Extract: duration, resolution, audio, etc.
└──────────┬────────────┘
           ↓
┌───────────────────────┐
│ Media Type Detection  │ → Determine: movie, TV, music, book
└──────────┬────────────┘
           ↓
┌───────────────────────┐
│ External API Lookup   │ → Enrich: plot, cast, artwork, etc.
└──────────┬────────────┘
           ↓
┌───────────────────────┐
│ Metadata Normalization│ → Standardize and validate
└───────────────────────┘
```

### 4.2 Filename Parsing Patterns

#### Movies
- Pattern: `{title} ({year})` or `{title}.{year}.{quality}.{source}.{codec}`
- Examples:
  - `The.Matrix.1999.1080p.BluRay.x264.mkv`
  - `Inception (2010).mp4`
  - `movie.2023.4k.webrip.h265.mkv`

#### TV Shows
- Pattern: `{show}.S##E##` or `{show}.{season}x{episode}`
- Examples:
  - `breaking.bad.s01e01.720p.mkv`
  - `Show.Name.1x01.Episode.Title.mkv`
  - `Series - 01x01 - Title.mp4`

#### Music
- Pattern: `{artist} - {album}` / `{track} - {title}`
- Examples:
  - `Artist Name - Album Name (2020)/01 - Song Title.mp3`
  - `01. Track Name.flac`

#### Books
- Pattern: `{title} - {author}` or `{author} - {title}`
- Examples:
  - `The Great Gatsby - F. Scott Fitzgerald.epub`
  - `Fitzgerald, F. Scott - The Great Gatsby.mobi`

### 4.3 External Metadata Sources

#### Movies & TV
- **Primary:** TMDB (The Movie Database)
  - Free API with generous limits
  - Comprehensive movie/TV metadata
  - Artwork and cast information
  
- **Fallback:** OMDB, TVDB

#### Music
- **Primary:** MusicBrainz
  - Free, open database
  - Comprehensive music metadata
  - Artist, album, track information

- **Artwork:** Cover Art Archive, Last.fm

#### Books
- **Primary:** OpenLibrary
  - Free API
  - ISBN lookup
  - Book metadata

- **Fallback:** Google Books API

### 4.4 File Metadata Extraction

Use Go libraries for reading embedded metadata:

- **Video:** `ffprobe` or `github.com/3d0c/gmf` (FFmpeg bindings)
  - Resolution, codec, duration, audio tracks, subtitles
  
- **Audio:** `github.com/dhowden/tag`
  - ID3 tags, album, artist, track number
  
- **Books:** `github.com/taylorskalyo/goreader/epub`
  - Embedded metadata from EPUB/MOBI files

---

## 5. Safety Mechanisms

### 5.1 Core Safety Principles

1. **Never Delete:** Only rename and move operations
2. **Transaction Log:** Record all operations before execution
3. **Dry-Run Mode:** Preview changes without execution
4. **Validation:** Pre-flight checks before operations
5. **Rollback:** Ability to undo operations
6. **Conflict Resolution:** Handle naming conflicts safely

### 5.2 Transaction System

#### Transaction Log Format (JSON)
```json
{
  "transaction_id": "uuid",
  "timestamp": "2025-12-07T10:30:00Z",
  "operations": [
    {
      "type": "move",
      "source": "/media/unsorted/movie.mkv",
      "destination": "/media/movies/Movie Name (2023)/Movie Name (2023).mkv",
      "status": "pending",
      "metadata": {
        "media_type": "movie",
        "title": "Movie Name",
        "year": 2023
      }
    }
  ],
  "status": "pending"
}
```

#### Transaction Lifecycle
1. **Plan:** Create transaction with all operations
2. **Validate:** Check permissions, disk space, conflicts
3. **Execute:** Perform operations, update status
4. **Commit:** Mark transaction as complete
5. **Rollback (if needed):** Reverse operations

### 5.3 Validation Checks

Before any operation:
- ✓ Source file exists and is readable
- ✓ Destination directory is writable
- ✓ Sufficient disk space available
- ✓ No filename conflicts (or resolve with strategy)
- ✓ Source and destination on same filesystem (for atomic moves)

### 5.4 Conflict Resolution Strategies

When destination file already exists:
1. **Skip:** Don't overwrite, log conflict
2. **Rename:** Add suffix (e.g., `-1`, `-2`)
3. **Interactive:** Prompt user (for CLI mode)
4. **Compare:** Check if files are identical (hash-based)

### 5.5 Rollback Implementation

```go
type Operation struct {
    Type        string // "move", "rename", "create_dir"
    Source      string
    Destination string
    Executed    bool
}

func Rollback(ops []Operation) error {
    // Reverse order
    for i := len(ops) - 1; i >= 0; i-- {
        op := ops[i]
        if op.Executed {
            // Move back to original location
            os.Rename(op.Destination, op.Source)
        }
    }
}
```

### 5.6 Safety Configuration

```yaml
safety:
  dry_run: false              # Preview mode
  create_backup: false        # Copy before move
  transaction_log: true       # Enable logging
  conflict_resolution: "skip" # skip | rename | interactive
  max_file_size: 50GB        # Safety limit
  allowed_extensions:         # Whitelist
    - .mkv
    - .mp4
    - .avi
    - .mp3
    - .flac
    - .epub
```

---

## 6. Implementation Phases

### Phase 1: Foundation (Week 1-2)

**Goal:** Set up project infrastructure and core utilities

**Tasks:**
- [x] Initialize Go module (`go mod init`)
- [ ] Set up project structure (packages, directories)
- [ ] Implement configuration system (Viper)
- [ ] Create CLI framework (Cobra)
- [ ] Implement file system scanner
- [ ] Add basic logging (structured logging with zerolog)
- [ ] Write unit tests for utilities

**Deliverables:**
- Working Go module with dependencies
- Basic CLI with `scan` command
- Configuration file loading
- File system traversal

**Dependencies:**
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration
- `github.com/rs/zerolog` - Logging

### Phase 2: Metadata Extraction (Week 3-4)

**Goal:** Parse filenames and extract metadata

**Tasks:**
- [ ] Implement filename parser with regex patterns
- [ ] Create media type detector (movie, TV, music, book)
- [ ] Integrate TMDB API client
- [ ] Integrate MusicBrainz API client
- [ ] Integrate OpenLibrary API client
- [ ] Build metadata enrichment pipeline
- [ ] Add caching for API responses
- [ ] Write comprehensive tests

**Deliverables:**
- Working filename parser for all media types
- External API integration
- Metadata extraction pipeline
- Cache system for API responses

**Dependencies:**
- External APIs: TMDB, MusicBrainz, OpenLibrary
- `github.com/patrickmn/go-cache` - In-memory cache

### Phase 3: File Organization (Week 5-6)

**Goal:** Implement file organization logic

**Tasks:**
- [ ] Implement Jellyfin naming conventions
- [ ] Create directory structure builder
- [ ] Build file renamer/mover
- [ ] Implement NFO file generation
- [ ] Add conflict resolution
- [ ] Create `organize` command
- [ ] Add `preview` command (dry-run)
- [ ] Write integration tests

**Deliverables:**
- File organization engine
- Preview functionality
- NFO file generation
- Working `organize` command

### Phase 4: Safety & Transactions (Week 7)

**Goal:** Implement safety mechanisms

**Tasks:**
- [ ] Create transaction logging system
- [ ] Implement rollback functionality
- [ ] Add validation checks
- [ ] Build backup mechanism (optional)
- [ ] Create `verify` command
- [ ] Add interactive conflict resolution
- [ ] Write safety tests

**Deliverables:**
- Transaction log system
- Rollback capability
- Verification tools
- Safety validations

### Phase 5: Polish & Documentation (Week 8)

**Goal:** Finalize and document

**Tasks:**
- [ ] Write comprehensive README
- [ ] Add usage examples
- [ ] Create documentation site
- [ ] Implement progress indicators
- [ ] Add statistics reporting
- [ ] Optimize performance
- [ ] Create release builds
- [ ] Final testing and bug fixes

**Deliverables:**
- Complete documentation
- Release-ready binary
- Example configurations
- User guide

### Phase 6: Advanced Features (Future)

**Optional enhancements:**
- [ ] Web UI for batch processing
- [ ] Watch mode (monitor directories)
- [ ] Subtitle handling
- [ ] Artwork download and management
- [ ] Multi-language support
- [ ] Plugin system for custom metadata sources
- [ ] Database for tracking organized files

---

## 7. Configuration Example

### Default Configuration File (`~/.go-jf-org/config.yaml`)

```yaml
# Media source directories
sources:
  - /media/unsorted
  - /downloads/complete

# Media destination directories
destinations:
  movies: /media/jellyfin/movies
  tv: /media/jellyfin/tv
  music: /media/jellyfin/music
  books: /media/jellyfin/books

# Metadata API keys (optional, uses free tier if not provided)
api_keys:
  tmdb: ""
  musicbrainz_app: "go-jf-org/1.0"

# Organization options
organize:
  create_nfo: true           # Generate NFO files
  download_artwork: true     # Download posters/fanart
  normalize_names: true      # Clean up filenames
  preserve_quality_tags: true # Keep quality info (1080p, etc.)

# Safety settings
safety:
  dry_run: false
  transaction_log: true
  log_directory: ~/.go-jf-org/logs
  conflict_resolution: skip  # skip | rename | interactive
  
# File filters
filters:
  min_file_size: 10MB       # Ignore small files
  video_extensions:
    - .mkv
    - .mp4
    - .avi
    - .mov
  audio_extensions:
    - .mp3
    - .flac
    - .m4a
  book_extensions:
    - .epub
    - .mobi
    - .pdf

# Performance
performance:
  max_concurrent_operations: 4
  api_rate_limit: 40         # requests per 10 seconds
  cache_ttl: 24h
```

---

## 8. CLI Usage Examples

### Basic Scan
```bash
# Scan directory for media files
go-jf-org scan /media/unsorted

# Scan with verbose output
go-jf-org scan /media/unsorted -v
```

### Preview Organization
```bash
# Preview changes without making them
go-jf-org preview /media/unsorted

# Preview for specific media type
go-jf-org preview /media/unsorted --type movie
```

### Organize Files
```bash
# Organize all media files
go-jf-org organize /media/unsorted

# Organize only movies
go-jf-org organize /media/unsorted --type movie

# Organize with custom destination
go-jf-org organize /media/unsorted --dest /media/jellyfin/movies

# Organize in dry-run mode
go-jf-org organize /media/unsorted --dry-run
```

### Verify Structure
```bash
# Verify organized media structure
go-jf-org verify /media/jellyfin/movies

# Verify with Jellyfin compatibility check
go-jf-org verify /media/jellyfin --strict
```

### Configuration
```bash
# Show current configuration
go-jf-org config show

# Set TMDB API key
go-jf-org config set api_keys.tmdb "your-api-key"

# Initialize default config
go-jf-org config init
```

---

## 9. Testing Strategy

### Unit Tests
- Parser logic (filename patterns)
- Metadata extraction
- Naming convention functions
- Validation logic

### Integration Tests
- Full organization workflow
- API integration (with mocks)
- Transaction system
- Rollback functionality

### Test Fixtures
```
test/fixtures/
├── movies/
│   ├── movie.2023.1080p.mkv
│   └── Another.Movie.2022.720p.mp4
├── tv/
│   ├── show.s01e01.mkv
│   └── series.1x02.mp4
├── music/
│   └── Artist - Album/01 - Song.mp3
└── books/
    └── Book.Title.epub
```

### CI/CD Pipeline
- Lint: `golangci-lint`
- Test: `go test ./...`
- Coverage: Minimum 80%
- Build: Multi-platform binaries (Linux, macOS, Windows)

---

## 10. Success Metrics

### Functionality
- ✓ Successfully detects 95%+ of common media file patterns
- ✓ Zero data loss (no file deletions)
- ✓ Handles 1000+ files without issues
- ✓ Rollback works 100% of the time

### Performance
- Scan: 1000 files/second
- Organize: 100 files/second (network-dependent)
- API calls: Respects rate limits
- Memory: < 100MB for typical workload

### User Experience
- Minimal configuration required
- Clear progress indicators
- Helpful error messages
- Comprehensive documentation

---

## 11. Potential Challenges & Solutions

### Challenge 1: Ambiguous Filenames
**Problem:** `Movie.mkv` - no year, quality, or clear title
**Solution:** 
- Interactive mode for ambiguous cases
- Use file metadata (date created)
- Allow manual override via config

### Challenge 2: API Rate Limits
**Problem:** TMDB has 40 requests/10 seconds limit
**Solution:**
- Implement caching layer
- Rate limiting with backoff
- Batch processing with delays
- Local metadata database

### Challenge 3: Filesystem Differences
**Problem:** Different filesystems (ext4, NTFS, APFS) have different limitations
**Solution:**
- Sanitize filenames for cross-platform compatibility
- Check filesystem type before operations
- Handle case-sensitive vs case-insensitive FS

### Challenge 4: Large File Operations
**Problem:** Moving 50GB+ files can take time and fail
**Solution:**
- Atomic operations when possible
- Progress indicators for large files
- Verify moves with checksums
- Transaction logging for partial failures

---

## 12. Future Enhancements

1. **Web Interface:** Browser-based UI for batch processing
2. **Watch Mode:** Automatically organize new files
3. **Duplicate Detection:** Find and handle duplicate media
4. **Quality Upgrade:** Replace lower quality with higher quality
5. **Subtitle Management:** Download and organize subtitles
6. **Custom Plugins:** User-defined metadata sources
7. **Cloud Support:** Integration with cloud storage (Rclone)
8. **Database Backend:** Track all organized files
9. **Multi-User Support:** Different profiles/configurations
10. **REST API:** Programmatic access to functionality

---

## Appendix A: Dependencies

### Core Dependencies
```go
require (
    github.com/spf13/cobra v1.8.0      // CLI framework
    github.com/spf13/viper v1.18.2     // Configuration
    github.com/rs/zerolog v1.31.0      // Logging
)
```

### Metadata & Media
```go
require (
    github.com/dhowden/tag v0.0.0-20230630033851-978a0926ee25  // Audio metadata
    github.com/3d0c/gmf v0.0.0-20220906170454-518f98f98b6f      // Video metadata
    github.com/taylorskalyo/goreader v0.0.0-20230626212555-e7f5644f8115 // EPUB reader
)
```

### Utilities
```go
require (
    github.com/patrickmn/go-cache v2.1.0+incompatible  // Caching
    github.com/google/uuid v1.5.0                       // UUID generation
    gopkg.in/yaml.v3 v3.0.1                            // YAML parsing
)
```

---

## Appendix B: References

### Jellyfin Documentation
- [Jellyfin Naming Conventions](https://jellyfin.org/docs/general/server/media/movies)
- [TV Show Naming](https://jellyfin.org/docs/general/server/media/shows)
- [Music Library](https://jellyfin.org/docs/general/server/media/music)

### Metadata APIs
- [TMDB API Docs](https://developers.themoviedb.org/3)
- [MusicBrainz API](https://musicbrainz.org/doc/MusicBrainz_API)
- [OpenLibrary API](https://openlibrary.org/developers/api)

### Go Libraries
- [Cobra (CLI)](https://github.com/spf13/cobra)
- [Viper (Config)](https://github.com/spf13/viper)
- [zerolog (Logging)](https://github.com/rs/zerolog)

---

**End of Implementation Plan**
