# Jellyfin Naming Conventions

This document provides detailed Jellyfin naming conventions that go-jf-org follows when organizing media files.

## Table of Contents
- [Movies](#movies)
- [TV Shows](#tv-shows)
- [Music](#music)
- [Books](#books)
- [NFO Files](#nfo-files)

---

## Movies

### Basic Naming

**Format:** `Movie Name (Year).ext`

**Examples:**
```
The Matrix (1999).mkv
Inception (2010).mp4
Avengers Endgame (2019).mkv
```

### Directory Structure
```
movies/
└── Movie Name (Year)/
    ├── Movie Name (Year).mkv
    ├── movie.nfo
    ├── poster.jpg
    └── fanart.jpg
```

### Special Cases

#### Multiple Versions
When you have multiple versions of the same movie:
```
Movie Name (Year)/
├── Movie Name (Year) - 1080p.mkv
├── Movie Name (Year) - 4K.mkv
└── Movie Name (Year) - Director's Cut.mkv
```

#### Multi-Part Movies
```
Movie Name (Year)/
├── Movie Name (Year) - Part 1.mkv
├── Movie Name (Year) - Part 2.mkv
└── movie.nfo
```

#### 3D Movies
```
Movie Name (Year)/
└── Movie Name (Year) [3D].mkv
```

### Valid Characters
- Use standard ASCII characters
- Replace special characters: `:` `?` `/` `\` with `-` or remove
- Remove leading/trailing dots and spaces

---

## TV Shows

### Basic Naming

**Format:** `Show Name - S##E## - Episode Title.ext`

**Examples:**
```
Breaking Bad - S01E01 - Pilot.mkv
Game of Thrones - S05E09 - The Dance of Dragons.mkv
The Office - S02E13 - The Secret.mkv
```

### Directory Structure
```
tv/
└── Show Name/
    ├── Season 01/
    │   ├── Show Name - S01E01 - Episode Title.mkv
    │   ├── Show Name - S01E02 - Episode Title.mkv
    │   └── season.nfo
    ├── Season 02/
    │   ├── Show Name - S02E01 - Episode Title.mkv
    │   └── season.nfo
    └── tvshow.nfo
```

### Special Cases

#### Multi-Episode Files
When a single file contains multiple episodes:
```
Show Name - S01E01-E02 - Double Episode.mkv
Show Name - S01E03-E04.mkv
```

#### Specials
Special episodes go in Season 00:
```
Show Name/
└── Season 00/
    ├── Show Name - S00E01 - Behind the Scenes.mkv
    ├── Show Name - S00E02 - Deleted Scenes.mkv
    └── season.nfo
```

#### Anime
Absolute episode numbering is also supported:
```
Anime Name/
└── Season 01/
    ├── Anime Name - 001.mkv
    ├── Anime Name - 002.mkv
    └── ...
```

#### Date-Based Episodes
For shows like daily shows:
```
Show Name/
└── Season 2025/
    ├── Show Name - 2025-01-15.mkv
    └── Show Name - 2025-01-16.mkv
```

---

## Music

### Basic Structure

**Format:**
```
Artist Name/
└── Album Name (Year)/
    ├── 01 - Track Name.ext
    ├── 02 - Track Name.ext
    ├── album.nfo
    └── folder.jpg
```

### Examples

#### Single Artist Album
```
Pink Floyd/
└── The Dark Side of the Moon (1973)/
    ├── 01 - Speak to Me.flac
    ├── 02 - Breathe.flac
    ├── 03 - On the Run.flac
    ├── album.nfo
    └── folder.jpg
```

#### Compilation Albums
```
Various Artists/
└── Greatest Hits of the 80s (2020)/
    ├── 01 - Artist A - Song Title.mp3
    ├── 02 - Artist B - Song Title.mp3
    └── album.nfo
```

#### Multi-Disc Albums
```
Artist Name/
└── Album Name (Year)/
    ├── Disc 1/
    │   ├── 01 - Track.mp3
    │   └── 02 - Track.mp3
    └── Disc 2/
        ├── 01 - Track.mp3
        └── 02 - Track.mp3
```

### Track Numbering
- Always use two-digit track numbers: `01`, `02`, etc.
- For albums with 100+ tracks: `001`, `002`, etc.

---

## Books

### Basic Structure

**Format:**
```
Author Last Name, First Name/
└── Book Title (Year)/
    ├── Book Title.epub
    ├── book.nfo
    └── cover.jpg
```

### Examples

#### Single Book
```
Fitzgerald, F. Scott/
└── The Great Gatsby (1925)/
    ├── The Great Gatsby.epub
    ├── book.nfo
    └── cover.jpg
```

#### Book Series
```
Rowling, J.K./
└── Harry Potter/
    ├── 01 - Harry Potter and the Philosopher's Stone (1997).epub
    ├── 02 - Harry Potter and the Chamber of Secrets (1998).epub
    ├── 03 - Harry Potter and the Prisoner of Azkaban (1999).epub
    └── series.nfo
```

#### Multiple Formats
```
Author Name/
└── Book Title (Year)/
    ├── Book Title.epub
    ├── Book Title.mobi
    ├── Book Title.pdf
    └── book.nfo
```

---

## NFO Files

NFO (iNFO) files are XML files containing metadata for media items. Jellyfin uses these for metadata enrichment.

### Movie NFO Example

**File:** `movie.nfo`
```xml
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<movie>
    <title>The Matrix</title>
    <originaltitle>The Matrix</originaltitle>
    <sorttitle>Matrix</sorttitle>
    <year>1999</year>
    <plot>A computer hacker learns about the true nature of reality...</plot>
    <tagline>Welcome to the Real World</tagline>
    <runtime>136</runtime>
    <mpaa>R</mpaa>
    <genre>Action</genre>
    <genre>Sci-Fi</genre>
    <studio>Warner Bros.</studio>
    <director>Lana Wachowski</director>
    <director>Lilly Wachowski</director>
    <actor>
        <name>Keanu Reeves</name>
        <role>Neo</role>
    </actor>
    <tmdbid>603</tmdbid>
    <imdbid>tt0133093</imdbid>
</movie>
```

### TV Show NFO Example

**File:** `tvshow.nfo`
```xml
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<tvshow>
    <title>Breaking Bad</title>
    <plot>A chemistry teacher diagnosed with cancer...</plot>
    <genre>Crime</genre>
    <genre>Drama</genre>
    <premiered>2008-01-20</premiered>
    <studio>AMC</studio>
    <actor>
        <name>Bryan Cranston</name>
        <role>Walter White</role>
    </actor>
    <tvdbid>81189</tvdbid>
    <tmdbid>1396</tmdbid>
</tvshow>
```

**File:** `season.nfo` (in season folder)
```xml
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<season>
    <seasonnumber>1</seasonnumber>
</season>
```

### Album NFO Example

**File:** `album.nfo`
```xml
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<album>
    <title>The Dark Side of the Moon</title>
    <artist>Pink Floyd</artist>
    <year>1973</year>
    <genre>Progressive Rock</genre>
    <review>One of the best-selling albums of all time...</review>
    <label>Harvest</label>
    <musicbrainzalbumid>a1ad30cb-b8c4-4d68-9253-15b18fcde1d1</musicbrainzalbumid>
</album>
```

### Book NFO Example

**File:** `book.nfo`
```xml
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<book>
    <title>The Great Gatsby</title>
    <author>F. Scott Fitzgerald</author>
    <year>1925</year>
    <plot>The story of the mysteriously wealthy Jay Gatsby...</plot>
    <genre>Fiction</genre>
    <genre>Classic</genre>
    <publisher>Charles Scribner's Sons</publisher>
    <isbn>9780743273565</isbn>
    <goodreadsid>4671</goodreadsid>
</book>
```

---

## Common Patterns

### File Extensions

#### Video
- `.mkv` - Matroska (preferred)
- `.mp4` - MPEG-4
- `.avi` - Audio Video Interleave
- `.m4v` - iTunes Video
- `.ts` - Transport Stream
- `.webm` - WebM

#### Audio
- `.flac` - FLAC (lossless, preferred)
- `.mp3` - MP3
- `.m4a` - AAC/ALAC
- `.ogg` - Ogg Vorbis
- `.opus` - Opus
- `.wav` - Waveform Audio

#### Books
- `.epub` - EPUB (preferred)
- `.mobi` - Kindle
- `.pdf` - PDF
- `.azw3` - Kindle Format 8
- `.cbz` - Comic Book Archive (ZIP)
- `.cbr` - Comic Book Archive (RAR)

### Character Sanitization

Replace or remove these characters in filenames:
- `<` `>` `:` `"` `/` `\` `|` `?` `*` → Remove or replace with `-`
- Leading/trailing spaces → Remove
- Leading/trailing dots → Remove
- Multiple spaces → Replace with single space
- ` & ` → `and` (optional)

### Year Extraction

Valid year formats:
- `(1999)` - Preferred
- `[1999]`
- `.1999.`
- ` 1999 `

Years must be between 1850-2100 for movies, TV, books.
Years 1900-current year for music.

---

## Tools & Validation

### Jellyfin Server Scanning

After organizing, Jellyfin will scan libraries and match files based on:
1. Folder/file names
2. NFO files (if present)
3. External metadata providers (TMDB, TVDB, MusicBrainz)

### Testing Conventions

Use these test files to verify naming:
```bash
# Movie
Movie Name (2023).mkv

# TV Episode
Show Name - S01E01 - Episode Title.mkv

# Music Track
01 - Track Name.mp3

# Book
Book Title.epub
```

Jellyfin should automatically detect and properly display these files.

---

## References

- [Jellyfin Official Documentation](https://jellyfin.org/docs/)
- [Jellyfin Movie Naming](https://jellyfin.org/docs/general/server/media/movies)
- [Jellyfin TV Naming](https://jellyfin.org/docs/general/server/media/shows)
- [Jellyfin Music Naming](https://jellyfin.org/docs/general/server/media/music)
- [Kodi NFO Files](https://kodi.wiki/view/NFO_files) (Jellyfin compatible)
