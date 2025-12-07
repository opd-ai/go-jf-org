# go-jf-org Technical Specification

## 1. ARCHITECTURE OVERVIEW

```
┌─────────────┐
│  CLI Layer  │  (cobra: scan, organize, preview, verify, rollback)
└──────┬──────┘
       │
┌──────▼──────────────────────────────────────────────────────┐
│               Core Business Logic                            │
├─────────────┬──────────┬───────────┬───────────┬────────────┤
│   Scanner   │ Detector │ Metadata  │ Organizer │   Safety   │
│  (traverse) │ (typing) │ (extract) │  (move)   │   (txn)    │
└─────────────┴──────────┴───────────┴───────────┴────────────┘
       │            │           │           │            │
       └────────────┴───────────┴───────────┴────────────┘
                              │
                    ┌─────────▼──────────┐
                    │  Storage & APIs    │
                    │ (fs, TMDB, MusicBz)│
                    └────────────────────┘
```

**Data Flow:**
1. Scanner → walks filesystem, filters by extension
2. Detector → determines media type (movie/tv/music/book)
3. Metadata → parses filename, reads file tags, enriches via API
4. Organizer → builds target path, creates NFO, moves file
5. Safety → logs operation, validates, enables rollback

## 2. DIRECTORY STRUCTURE

```
go-jf-org/
├── cmd/
│   ├── root.go              # CLI setup
│   ├── scan.go              # List detected media
│   ├── organize.go          # Execute organization
│   ├── preview.go           # Dry-run mode
│   ├── verify.go            # Check output structure
│   └── rollback.go          # Undo operations
├── internal/
│   ├── scanner/
│   │   ├── scanner.go       # Walk directories
│   │   └── filter.go        # Extension matching
│   ├── detector/
│   │   ├── detector.go      # Interface
│   │   ├── movie.go         # Movie patterns
│   │   ├── tv.go            # TV patterns (S##E##)
│   │   ├── music.go         # Audio file detection
│   │   └── book.go          # eBook detection
│   ├── metadata/
│   │   ├── parser.go        # Regex-based filename parsing
│   │   ├── extractor.go     # Read embedded metadata
│   │   ├── enricher.go      # API orchestration
│   │   ├── tmdb.go          # TMDB client (movies/TV)
│   │   ├── musicbrainz.go   # MusicBrainz client
│   │   └── openlibrary.go   # OpenLibrary client
│   ├── organizer/
│   │   ├── organizer.go     # Main orchestrator
│   │   ├── namer.go         # Jellyfin naming rules
│   │   ├── nfo.go           # NFO XML generation
│   │   └── mover.go         # Atomic file operations
│   ├── safety/
│   │   ├── transaction.go   # Operation logging (JSON)
│   │   ├── rollback.go      # Reverse operations
│   │   └── validator.go     # Pre-flight checks
│   └── config/
│       ├── config.go        # Viper integration
│       └── defaults.go      # Sane defaults
├── pkg/types/
│   └── media.go             # Shared structs
└── main.go
```

## 3. CORE COMPONENTS

### Scanner
- **Purpose:** Discover media files in source directories
- **Types:** `ScanResult{Files []string, Errors []error}`
- **Functions:**
  - `Walk(path string) ([]string, error)` - recursive traversal
  - `Filter(files []string, exts []string) []string` - extension filtering

### Detector
- **Purpose:** Classify files as movie/TV/music/book
- **Types:** `MediaType int` (const: Movie, TV, Music, Book)
- **Functions:**
  - `Detect(path string) MediaType` - primary detection
  - `IsMovie(filename string) bool` - pattern: `Title.Year.Quality`
  - `IsTV(filename string) bool` - pattern: `S##E##` or `#x#`

### Metadata Parser
- **Purpose:** Extract metadata from filenames and files
- **Types:** `Metadata{Title, Year, Quality, Season, Episode string}`
- **Functions:**
  - `ParseMovie(filename) *MovieMeta` - regex: `(.+)[\. ](\d{4})`
  - `ParseTV(filename) *TVMeta` - regex: `S(\d+)E(\d+)`
  - `ReadTags(path) map[string]string` - ID3/FLAC/EPUB tags

### Metadata Enricher
- **Purpose:** Fetch external metadata (TMDB, MusicBrainz)
- **Functions:**
  - `EnrichMovie(title, year) *MovieDetails` - TMDB search
  - `EnrichTV(show, season, ep) *EpisodeDetails`
  - `EnrichMusic(artist, album) *AlbumDetails` - MusicBrainz

### Organizer
- **Purpose:** Move files to Jellyfin-compatible structure
- **Types:** `Operation{Type, Source, Dest, Status string}`
- **Functions:**
  - `BuildPath(media) string` - construct target path
  - `Move(src, dst) error` - atomic move with validation
  - `GenerateNFO(media) string` - XML for Jellyfin

### Safety
- **Purpose:** Enable rollback and prevent data loss
- **Types:** `Transaction{ID, Ops []Operation, Time}`
- **Functions:**
  - `Log(op) error` - persist to `~/.go-jf-org/txn/`
  - `Rollback(txnID) error` - reverse all operations
  - `Validate(op) error` - check disk space, permissions

## 4. FILE ORGANIZATION STRATEGY

### Movies
```
Input:  The.Matrix.1999.1080p.BluRay.mkv
Output: movies/The Matrix (1999)/The Matrix (1999).mkv
        movies/The Matrix (1999)/movie.nfo
        movies/The Matrix (1999)/poster.jpg
```

### TV Shows
```
Input:  breaking.bad.s01e01.720p.mkv
Output: tv/Breaking Bad/Season 01/Breaking Bad - S01E01 - Pilot.mkv
        tv/Breaking Bad/Season 01/season.nfo
        tv/Breaking Bad/tvshow.nfo
```

### Music
```
Input:  Pink Floyd - Dark Side/01 Speak to Me.mp3
Output: music/Pink Floyd/The Dark Side of the Moon (1973)/01 - Speak to Me.mp3
        music/Pink Floyd/The Dark Side of the Moon (1973)/album.nfo
        music/Pink Floyd/The Dark Side of the Moon (1973)/folder.jpg
```

### Books
```
Input:  gatsby.epub
Output: books/Fitzgerald, F. Scott/The Great Gatsby (1925)/The Great Gatsby.epub
        books/Fitzgerald, F. Scott/The Great Gatsby (1925)/book.nfo
```

### Edge Cases
- **Duplicates:** Append `-1`, `-2` suffix
- **Missing year:** Interactive prompt or skip
- **Special chars:** Sanitize (`:` → `-`, remove `?/\<>|`)

## 5. METADATA HANDLING

### Filename Patterns
```go
// Movies
"(?P<title>.+?)[\. ](?P<year>19\d{2}|20\d{2})[\. ]?(?P<quality>\d{3,4}p)?"

// TV Shows
"(?P<show>.+?)[\. ]S(?P<season>\d{2})E(?P<episode>\d{2})"
"(?P<show>.+?)[\. ](?P<season>\d{1,2})x(?P<episode>\d{2})"

// Music (from ID3 tags primarily)
Artist, Album, Track, Year, Genre

// Books (from EPUB metadata)
Title, Author, Publisher, ISBN
```

### Metadata Sources by Type
| Type  | Filename | Embedded | External API |
|-------|----------|----------|--------------|
| Movie | Title, Year, Quality | Duration, Codec | TMDB (plot, cast) |
| TV    | Show, S##E## | Duration | TMDB (episode title) |
| Music | - | ID3 tags | MusicBrainz (album art) |
| Book  | Title hint | EPUB meta | OpenLibrary (cover) |

### Metadata Injection
- **Video:** Use `ffmpeg` to embed chapters/metadata (optional, v2)
- **Audio:** `github.com/dhowden/tag` for ID3v2 writing
- **Books:** EPUB metadata is read-only, use sidecar NFO

## 6. SAFETY MECHANISMS

### Dry-Run Mode
```bash
go-jf-org organize /media/unsorted --dry-run
# Output: Shows planned operations, no file changes
```

### Transaction Log (JSON)
```json
{
  "id": "txn-2025-12-07-abc123",
  "timestamp": "2025-12-07T15:00:00Z",
  "operations": [
    {"type": "move", "src": "/a/movie.mkv", "dst": "/b/Movie (2023)/Movie (2023).mkv"},
    {"type": "create", "dst": "/b/Movie (2023)/movie.nfo"}
  ],
  "status": "completed"
}
```

### Rollback
```bash
go-jf-org rollback txn-2025-12-07-abc123
# Reverses: moves file back, deletes created NFOs/dirs
```

### Validation Checks
- Source exists and readable
- Destination writable
- Sufficient disk space (file size + 10% buffer)
- No filename conflicts (or resolve via config: skip/rename)

## 7. IMPLEMENTATION PHASES

### Phase 1: Foundation (Week 1-2)
**Deliverables:**
- CLI skeleton (cobra)
- Config loading (viper)
- Scanner implementation
- Basic logging

**Tests:**
- Scan 1000 files in <1s
- Config validation

### Phase 2: Detection & Parsing (Week 3)
**Deliverables:**
- Media type detector
- Regex filename parsers (movie, TV)
- Embedded metadata readers

**Tests:**
- 95%+ accuracy on common patterns
- Handle edge cases (no year, special chars)

### Phase 3: External APIs (Week 4)
**Deliverables:**
- TMDB client (rate-limited: 40/10s)
- MusicBrainz client
- OpenLibrary client
- Response caching (24h TTL)

**Tests:**
- Mock API responses
- Rate limit compliance

### Phase 4: Organization (Week 5)
**Deliverables:**
- Path builder (Jellyfin conventions)
- File mover (atomic operations)
- NFO generator

**Tests:**
- Verify output structure
- No data loss

### Phase 5: Safety (Week 6)
**Deliverables:**
- Transaction logging
- Rollback engine
- Validation pipeline

**Tests:**
- Rollback restores original state
- Handles partial failures

### Phase 6: Polish (Week 7-8)
**Deliverables:**
- Progress bars
- Interactive mode
- Documentation
- Binaries (Linux, macOS, Windows)

## 8. THIRD-PARTY LIBRARIES

| Library | Purpose | Justification |
|---------|---------|---------------|
| `github.com/spf13/cobra` | CLI framework | Industry standard, excellent UX |
| `github.com/spf13/viper` | Config management | YAML/ENV/flags support |
| `github.com/rs/zerolog` | Structured logging | Fast, zero-alloc |
| `github.com/dhowden/tag` | Audio metadata | Pure Go, ID3/FLAC/MP4 |
| `github.com/taylorskalyo/goreader/epub` | EPUB parsing | Lightweight |
| HTTP client (stdlib) | API requests | No external deps needed |
| `github.com/schollz/progressbar/v3` | Progress UI | Simple, customizable |

### Why NOT use:
- **ffmpeg bindings:** Heavy dependency, only for advanced features (Phase 6+)
- **ORMs:** No database needed, use JSON files
- **Web frameworks:** CLI-only for MVP

## JELLYFIN COMPATIBILITY NOTES

### Required Naming Patterns
- Movies: `Name (Year).ext` in `Name (Year)/` directory
- TV: `Show - S##E## - Title.ext` in `Show/Season ##/`
- Music: `## - Track.ext` in `Artist/Album (Year)/`

### NFO Format (Kodi-compatible)
```xml
<movie>
  <title>The Matrix</title>
  <year>1999</year>
  <plot>...</plot>
  <tmdbid>603</tmdbid>
</movie>
```

### Artwork Locations
- `poster.jpg` - Movie/show poster
- `fanart.jpg` - Background art
- `folder.jpg` - Album cover

## SUCCESS METRICS

- **Safety:** Zero file deletions, 100% rollback success
- **Performance:** 100 files/sec organization (SSD)
- **Accuracy:** 95%+ correct metadata matching
- **UX:** Single command for typical use case
