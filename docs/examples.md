# Usage Examples

This document provides practical examples of using go-jf-org to organize media files.

## Table of Contents
- [Quick Start](#quick-start)
- [Movies](#movies)
- [TV Shows](#tv-shows)
- [Music](#music)
- [Books](#books)
- [Advanced Scenarios](#advanced-scenarios)

---

## Quick Start

### Installation

```bash
# Download and install
go install github.com/opd-ai/go-jf-org@latest

# Or build from source
git clone https://github.com/opd-ai/go-jf-org.git
cd go-jf-org
make build
```

### Initial Configuration

```bash
# Create default configuration
go-jf-org config init

# Edit configuration (optional)
nano ~/.go-jf-org/config.yaml
```

### Basic Workflow

```bash
# 1. Scan directory to see what will be organized
go-jf-org scan /media/unsorted

# 2. Preview changes (dry-run)
go-jf-org preview /media/unsorted

# 3. Organize files
go-jf-org organize /media/unsorted
```

---

## Movies

### Example 1: Simple Movie Organization

**Input:**
```
/downloads/
├── The.Matrix.1999.1080p.BluRay.x264.mkv
├── Inception.2010.720p.WEB-DL.mkv
└── Interstellar (2014).mp4
```

**Command:**
```bash
go-jf-org organize /downloads --type movie --dest /media/jellyfin/movies
```

**Output:**
```
/media/jellyfin/movies/
├── The Matrix (1999)/
│   ├── The Matrix (1999).mkv
│   ├── movie.nfo
│   └── poster.jpg
├── Inception (2010)/
│   ├── Inception (2010).mkv
│   ├── movie.nfo
│   └── poster.jpg
└── Interstellar (2014)/
    ├── Interstellar (2014).mp4
    ├── movie.nfo
    └── poster.jpg
```

### Example 2: Movies with Multiple Versions

**Input:**
```
/downloads/
├── blade.runner.1982.theatrical.1080p.mkv
├── blade.runner.1982.final.cut.2160p.mkv
└── blade.runner.1982.directors.cut.1080p.mkv
```

**Command:**
```bash
go-jf-org organize /downloads --type movie
```

**Output:**
```
/media/movies/
└── Blade Runner (1982)/
    ├── Blade Runner (1982) - Theatrical.mkv
    ├── Blade Runner (1982) - Final Cut.mkv
    ├── Blade Runner (1982) - Director's Cut.mkv
    └── movie.nfo
```

### Example 3: Preview Before Organizing

**Command:**
```bash
go-jf-org preview /downloads --type movie
```

**Output:**
```
Preview: Movie Organization
===========================

Source: /downloads/The.Matrix.1999.1080p.BluRay.x264.mkv
  → Destination: /media/movies/The Matrix (1999)/The Matrix (1999).mkv
  Metadata: Title="The Matrix", Year=1999, Quality=1080p
  
Source: /downloads/Inception.2010.720p.WEB-DL.mkv
  → Destination: /media/movies/Inception (2010)/Inception (2010).mkv
  Metadata: Title="Inception", Year=2010, Quality=720p

Total: 2 movies
Operations: 2 moves, 2 NFO files, 4 images
```

---

## TV Shows

### Example 1: Single Season

**Input:**
```
/downloads/
├── breaking.bad.s01e01.720p.mkv
├── breaking.bad.s01e02.720p.mkv
├── breaking.bad.s01e03.720p.mkv
└── breaking.bad.s01e04.720p.mkv
```

**Command:**
```bash
go-jf-org organize /downloads --type tv --dest /media/jellyfin/tv
```

**Output:**
```
/media/jellyfin/tv/
└── Breaking Bad/
    ├── Season 01/
    │   ├── Breaking Bad - S01E01 - Pilot.mkv
    │   ├── Breaking Bad - S01E02 - Cat's in the Bag.mkv
    │   ├── Breaking Bad - S01E03 - And the Bag's in the River.mkv
    │   ├── Breaking Bad - S01E04 - Cancer Man.mkv
    │   └── season.nfo
    └── tvshow.nfo
```

### Example 2: Multiple Seasons

**Input:**
```
/downloads/
├── got.s05e01.1080p.mkv
├── got.s05e02.1080p.mkv
├── got.s06e01.1080p.mkv
└── got.s06e02.1080p.mkv
```

**Command:**
```bash
go-jf-org organize /downloads --type tv
```

**Output:**
```
/media/tv/
└── Game of Thrones/
    ├── Season 05/
    │   ├── Game of Thrones - S05E01 - The Wars to Come.mkv
    │   ├── Game of Thrones - S05E02 - The House of Black and White.mkv
    │   └── season.nfo
    ├── Season 06/
    │   ├── Game of Thrones - S06E01 - The Red Woman.mkv
    │   ├── Game of Thrones - S06E02 - Home.mkv
    │   └── season.nfo
    └── tvshow.nfo
```

### Example 3: TV Show with Specials

**Input:**
```
/downloads/
├── doctor.who.s00e01.the.five.doctors.mkv
├── doctor.who.s11e01.the.woman.who.fell.to.earth.mkv
```

**Command:**
```bash
go-jf-org organize /downloads --type tv
```

**Output:**
```
/media/tv/
└── Doctor Who/
    ├── Season 00/
    │   ├── Doctor Who - S00E01 - The Five Doctors.mkv
    │   └── season.nfo
    ├── Season 11/
    │   ├── Doctor Who - S11E01 - The Woman Who Fell to Earth.mkv
    │   └── season.nfo
    └── tvshow.nfo
```

---

## Music

### Example 1: Single Album

**Input:**
```
/downloads/
└── Pink Floyd - The Dark Side of the Moon/
    ├── 01 Speak to Me.mp3
    ├── 02 Breathe.mp3
    ├── 03 On the Run.mp3
    └── 04 Time.mp3
```

**Command:**
```bash
go-jf-org organize /downloads --type music --dest /media/jellyfin/music
```

**Output:**
```
/media/jellyfin/music/
└── Pink Floyd/
    └── The Dark Side of the Moon (1973)/
        ├── 01 - Speak to Me.mp3
        ├── 02 - Breathe.mp3
        ├── 03 - On the Run.mp3
        ├── 04 - Time.mp3
        ├── album.nfo
        └── folder.jpg
```

### Example 2: Multiple Artists

**Input:**
```
/downloads/
├── Beatles - Abbey Road/
│   └── 01 Come Together.flac
├── Led Zeppelin - IV/
│   └── 01 Black Dog.flac
└── Queen - A Night at the Opera/
    └── 01 Bohemian Rhapsody.flac
```

**Command:**
```bash
go-jf-org organize /downloads --type music
```

**Output:**
```
/media/music/
├── The Beatles/
│   └── Abbey Road (1969)/
│       ├── 01 - Come Together.flac
│       ├── album.nfo
│       └── folder.jpg
├── Led Zeppelin/
│   └── Led Zeppelin IV (1971)/
│       ├── 01 - Black Dog.flac
│       ├── album.nfo
│       └── folder.jpg
└── Queen/
    └── A Night at the Opera (1975)/
        ├── 01 - Bohemian Rhapsody.flac
        ├── album.nfo
        └── folder.jpg
```

### Example 3: Compilation Album

**Input:**
```
/downloads/
└── Various Artists - Now That's What I Call Music 50/
    ├── 01 Artist A - Song One.mp3
    ├── 02 Artist B - Song Two.mp3
    └── 03 Artist C - Song Three.mp3
```

**Output:**
```
/media/music/
└── Various Artists/
    └── Now That's What I Call Music 50 (2020)/
        ├── 01 - Artist A - Song One.mp3
        ├── 02 - Artist B - Song Two.mp3
        ├── 03 - Artist C - Song Three.mp3
        └── album.nfo
```

---

## Books

### Example 1: Individual Books

**Input:**
```
/downloads/
├── The Great Gatsby - F Scott Fitzgerald.epub
├── 1984.George.Orwell.mobi
└── To Kill a Mockingbird.epub
```

**Command:**
```bash
go-jf-org organize /downloads --type book --dest /media/jellyfin/books
```

**Output:**
```
/media/jellyfin/books/
├── Fitzgerald, F. Scott/
│   └── The Great Gatsby (1925)/
│       ├── The Great Gatsby.epub
│       ├── book.nfo
│       └── cover.jpg
├── Orwell, George/
│   └── 1984 (1949)/
│       ├── 1984.mobi
│       ├── book.nfo
│       └── cover.jpg
└── Lee, Harper/
    └── To Kill a Mockingbird (1960)/
        ├── To Kill a Mockingbird.epub
        ├── book.nfo
        └── cover.jpg
```

### Example 2: Book Series

**Input:**
```
/downloads/
├── Harry Potter 1 - Philosopher's Stone.epub
├── Harry Potter 2 - Chamber of Secrets.epub
└── Harry Potter 3 - Prisoner of Azkaban.epub
```

**Command:**
```bash
go-jf-org organize /downloads --type book --series
```

**Output:**
```
/media/books/
└── Rowling, J.K./
    └── Harry Potter/
        ├── 01 - Harry Potter and the Philosopher's Stone (1997).epub
        ├── 02 - Harry Potter and the Chamber of Secrets (1998).epub
        ├── 03 - Harry Potter and the Prisoner of Azkaban (1999).epub
        └── series.nfo
```

---

## Advanced Scenarios

### Example 1: Mixed Media Directory

**Input:**
```
/downloads/
├── movie.2023.1080p.mkv
├── show.s01e01.mkv
├── Artist - Album/
│   └── 01 - Song.mp3
└── Book Title.epub
```

**Command:**
```bash
# Organize all types at once
go-jf-org organize /downloads --all-types
```

**Output:**
Files organized into appropriate media type directories:
- Movie → `/media/movies/`
- TV → `/media/tv/`
- Music → `/media/music/`
- Book → `/media/books/`

### Example 2: Dry Run with Verbose Output

**Command:**
```bash
go-jf-org organize /downloads --dry-run --verbose
```

**Output:**
```
[INFO] Scanning directory: /downloads
[INFO] Found 42 media files
[INFO] Detecting media types...
[INFO] Detected: 15 movies, 20 TV episodes, 5 music tracks, 2 books

[DRY RUN] No files will be moved

Movie: The.Matrix.1999.1080p.BluRay.mkv
  Detected as: Movie
  Parsed: title="The Matrix", year=1999, quality="1080p", source="BluRay"
  TMDB Match: ID=603, Title="The Matrix", Year=1999, Confidence=100%
  → Would move to: /media/movies/The Matrix (1999)/The Matrix (1999).mkv
  → Would create: movie.nfo
  → Would download: poster.jpg, fanart.jpg

TV Show: breaking.bad.s01e01.720p.mkv
  Detected as: TV Show
  Parsed: show="Breaking Bad", season=1, episode=1, quality="720p"
  TMDB Match: ID=1396, Title="Breaking Bad", Confidence=98%
  Episode: "Pilot"
  → Would move to: /media/tv/Breaking Bad/Season 01/Breaking Bad - S01E01 - Pilot.mkv
  → Would create: tvshow.nfo, season.nfo

[DRY RUN SUMMARY]
Total files: 42
  - 15 movies → 15 directories
  - 20 TV episodes → 3 shows, 5 seasons
  - 5 music tracks → 2 albums
  - 2 books → 2 authors

No files were actually moved or modified.
Run without --dry-run to perform these operations.
```

### Example 3: Interactive Mode

**Command:**
```bash
go-jf-org organize /downloads --interactive
```

**Output:**
```
Found ambiguous movie: "movie.mkv"

Possible matches:
1. The Matrix (1999)
2. The Matrix Reloaded (2003)
3. The Matrix Revolutions (2003)
4. Skip this file
5. Enter custom title

Select option [1-5]: 1

Selected: The Matrix (1999)
✓ Organized to: /media/movies/The Matrix (1999)/The Matrix (1999).mkv
```

### Example 4: Rollback Operation

**Command:**
```bash
# Organize files
go-jf-org organize /downloads
# Transaction ID: abc123-456def-789ghi

# Something went wrong, rollback
go-jf-org rollback abc123-456def-789ghi
```

**Output:**
```
Rolling back transaction: abc123-456def-789ghi
✓ Moved file back: /media/movies/Movie/Movie.mkv → /downloads/movie.mkv
✓ Removed directory: /media/movies/Movie/
✓ Removed NFO: /media/movies/Movie/movie.nfo

Rollback complete. All files restored to original location.
```

### Example 5: Verify Jellyfin Structure

**Command:**
```bash
go-jf-org verify /media/jellyfin/movies --strict
```

**Output:**
```
Verifying: /media/jellyfin/movies

✓ The Matrix (1999)/The Matrix (1999).mkv
  - Correct naming convention
  - NFO file present
  - Artwork present (poster.jpg, fanart.jpg)
  
✗ Inception (2010)/inception.mkv
  - ISSUE: Incorrect filename (should be "Inception (2010).mkv")
  - NFO file missing
  
✗ Interstellar/Interstellar.mkv
  - ISSUE: Missing year in directory name
  - ISSUE: Missing year in filename
  
Summary:
  Total: 3 movies
  Correct: 1 (33%)
  Issues: 2 (67%)
  
Run with --fix to automatically correct issues.
```

### Example 6: Custom Configuration

**Configuration:** `~/.go-jf-org/config.yaml`
```yaml
sources:
  - /downloads/complete
  - /media/staging

destinations:
  movies: /mnt/storage/jellyfin/movies
  tv: /mnt/storage/jellyfin/tv
  music: /mnt/storage/jellyfin/music
  books: /mnt/storage/jellyfin/books

organize:
  create_nfo: true
  download_artwork: true
  normalize_names: true
  preserve_quality_tags: true

safety:
  dry_run: false
  transaction_log: true
  conflict_resolution: rename

api_keys:
  tmdb: "your-api-key-here"
```

**Command:**
```bash
# Use configured sources
go-jf-org organize --use-config
```

### Example 7: Batch Processing with Progress

**Command:**
```bash
go-jf-org organize /media/unsorted --batch-size 100
```

**Output:**
```
Organizing media files...

[====================>                    ] 50% (50/100)
Processing: show.s03e15.mkv
ETA: 2m 30s

Current Stats:
  Processed: 50 files
  Successful: 48
  Skipped: 2
  Failed: 0
  
API Calls: 45 (TMDB)
Cache Hits: 30
```

---

## Troubleshooting

### Common Issues

#### Issue: File Not Detected
```bash
# Check what was detected
go-jf-org scan /path/to/file --verbose

# Force media type
go-jf-org organize /path/to/file --type movie
```

#### Issue: Wrong Metadata
```bash
# Clear cache and retry
go-jf-org organize /path --clear-cache

# Use offline mode (filename only)
go-jf-org organize /path --offline
```

#### Issue: Permission Denied
```bash
# Check file permissions
ls -la /path/to/file

# Run with sudo (not recommended)
# Better: fix permissions
chmod 644 /path/to/file
chown user:user /path/to/file
```

---

## Tips & Best Practices

1. **Always Preview First**
   ```bash
   go-jf-org preview /path
   ```

2. **Use Dry Run for Safety**
   ```bash
   go-jf-org organize /path --dry-run
   ```

3. **Enable Transaction Logging**
   - Allows rollback if needed
   - Configured in `~/.go-jf-org/config.yaml`

4. **Use Interactive Mode for Ambiguous Files**
   ```bash
   go-jf-org organize /path --interactive
   ```

5. **Keep API Keys Secure**
   - Store in config file (not in scripts)
   - Use environment variables
   ```bash
   export TMDB_API_KEY="your-key"
   go-jf-org organize /path
   ```

6. **Regular Verification**
   ```bash
   # Weekly cron job
   0 2 * * 0 go-jf-org verify /media/jellyfin --strict
   ```

---

## References

- [Configuration Guide](../README.md#configuration)
- [Jellyfin Naming Conventions](jellyfin-conventions.md)
- [Metadata Sources](metadata-sources.md)
