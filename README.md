# go-jf-org

> A safe, powerful Go CLI tool to organize disorganized media files into a Jellyfin-compatible structure

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)

## Overview

`go-jf-org` automatically organizes your messy media collections (movies, TV shows, music, books) into a clean, Jellyfin-compatible directory structure. It extracts metadata from filenames and files, enriches it with external APIs, and safely moves files without ever deleting anything.

### Key Features

- 🔒 **100% Safe** - Never deletes files, only moves/renames
- 🎬 **Multi-Media Support** - Movies, TV shows, music, and books
- 🤖 **Smart Detection** - Extracts metadata from filenames and file content
- 🌐 **API Integration** - Enriches metadata using TMDB, MusicBrainz, OpenLibrary
- 📝 **NFO Generation** - Creates Jellyfin-compatible NFO files
- 🎨 **Artwork Download** - Fetches posters, fanart, and album covers
- 🔄 **Rollback Support** - Transaction log allows undoing operations
- 👁️ **Preview Mode** - Dry-run to see changes before executing
- ⚙️ **Minimal Config** - Works out-of-the-box with sensible defaults

## Quick Start

```bash
# Preview what will be organized
go-jf-org preview /path/to/unsorted/media

# Organize files
go-jf-org organize /path/to/unsorted/media

# Verify organized structure
go-jf-org verify /media/jellyfin/movies
```

## Documentation

- **[📋 Implementation Plan](IMPLEMENTATION_PLAN.md)** - Complete architecture and development roadmap
- **[📖 Jellyfin Conventions](docs/jellyfin-conventions.md)** - Naming standards for all media types
- **[🔍 Metadata Sources](docs/metadata-sources.md)** - External APIs and extraction strategies
- **[💡 Usage Examples](docs/examples.md)** - Practical examples and common scenarios

## What It Does

### Before
```
/downloads/
├── The.Matrix.1999.1080p.BluRay.x264.mkv
├── breaking.bad.s01e01.720p.mkv
├── Pink Floyd - Dark Side of the Moon/
│   └── 01 Speak to Me.mp3
└── The.Great.Gatsby.epub
```

### After
```
/media/jellyfin/
├── movies/
│   └── The Matrix (1999)/
│       ├── The Matrix (1999).mkv
│       ├── movie.nfo
│       └── poster.jpg
├── tv/
│   └── Breaking Bad/
│       ├── Season 01/
│       │   ├── Breaking Bad - S01E01 - Pilot.mkv
│       │   └── season.nfo
│       └── tvshow.nfo
├── music/
│   └── Pink Floyd/
│       └── The Dark Side of the Moon (1973)/
│           ├── 01 - Speak to Me.mp3
│           ├── album.nfo
│           └── folder.jpg
└── books/
    └── Fitzgerald, F. Scott/
        └── The Great Gatsby (1925)/
            ├── The Great Gatsby.epub
            └── book.nfo
```

## Installation

### From Source
```bash
git clone https://github.com/opd-ai/go-jf-org.git
cd go-jf-org
make build
sudo make install
```

### Using Go
```bash
go install github.com/opd-ai/go-jf-org@latest
```

## Configuration

Create default configuration:
```bash
go-jf-org config init
```

Configuration file: `~/.go-jf-org/config.yaml`

```yaml
sources:
  - /media/unsorted

destinations:
  movies: /media/jellyfin/movies
  tv: /media/jellyfin/tv
  music: /media/jellyfin/music
  books: /media/jellyfin/books

api_keys:
  tmdb: "your-api-key"  # Optional, uses free tier

organize:
  create_nfo: true
  download_artwork: true
  normalize_names: true

safety:
  dry_run: false
  transaction_log: true
  conflict_resolution: skip  # skip | rename | interactive
```

## Usage Examples

### Scan Directory
```bash
# See what media files are detected
go-jf-org scan /media/unsorted
```

### Preview Changes
```bash
# Dry-run to see what will happen
go-jf-org preview /media/unsorted
```

### Organize Media
```bash
# Organize all media types
go-jf-org organize /media/unsorted

# Organize only movies
go-jf-org organize /media/unsorted --type movie

# Interactive mode for ambiguous files
go-jf-org organize /media/unsorted --interactive
```

### Verify Structure
```bash
# Check if structure is Jellyfin-compatible
go-jf-org verify /media/jellyfin/movies --strict
```

### Rollback
```bash
# Undo an organization operation
go-jf-org rollback <transaction-id>
```

## Safety Features

### Transaction Logging
Every operation is logged before execution:
```json
{
  "transaction_id": "abc-123",
  "operations": [
    {
      "type": "move",
      "source": "/downloads/movie.mkv",
      "destination": "/media/movies/Movie (2023)/Movie (2023).mkv"
    }
  ]
}
```

### Rollback Support
All operations can be reversed:
```bash
go-jf-org rollback abc-123
```

### Validation Checks
Before any operation:
- ✓ Source file exists and is readable
- ✓ Destination is writable
- ✓ Sufficient disk space
- ✓ No conflicts (or resolve per strategy)

### Conflict Resolution
- **Skip** - Don't overwrite existing files
- **Rename** - Add suffix (-1, -2, etc.)
- **Interactive** - Ask user for decision

## Supported Media Types

### Movies
- **Formats:** MKV, MP4, AVI, M4V, TS, WebM
- **Metadata:** TMDB
- **Convention:** `Movie Name (Year).ext`

### TV Shows
- **Formats:** MKV, MP4, AVI, M4V, TS, WebM
- **Metadata:** TMDB
- **Convention:** `Show Name - S##E## - Episode Title.ext`

### Music
- **Formats:** FLAC, MP3, M4A, OGG, Opus, WAV
- **Metadata:** MusicBrainz, ID3 tags
- **Convention:** `Artist/Album (Year)/## - Track.ext`

### Books
- **Formats:** EPUB, MOBI, PDF, AZW3, CBZ, CBR
- **Metadata:** OpenLibrary, embedded metadata
- **Convention:** `Author/Book Title (Year)/Book Title.ext`

## Development Status

This project is currently in the **planning phase**. See [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md) for the complete development roadmap.

### Planned Phases
1. **Foundation** - Project setup, CLI framework, file scanning
2. **Metadata Extraction** - Filename parsing, API integration
3. **File Organization** - Moving/renaming, NFO generation
4. **Safety & Transactions** - Logging, rollback, validation
5. **Polish** - Documentation, optimization, release
6. **Advanced Features** - Web UI, watch mode, plugins

## Contributing

Contributions are welcome! Please see the implementation plan for areas where help is needed.

```bash
# Fork the repository
# Create a feature branch
git checkout -b feature/amazing-feature

# Make your changes
# Test thoroughly
make test

# Commit and push
git commit -m "Add amazing feature"
git push origin feature/amazing-feature

# Open a Pull Request
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Jellyfin](https://jellyfin.org) - Free media server
- [TMDB](https://www.themoviedb.org) - Movie and TV metadata
- [MusicBrainz](https://musicbrainz.org) - Music metadata
- [OpenLibrary](https://openlibrary.org) - Book metadata

## Support

- 📖 [Documentation](docs/)
- 🐛 [Issue Tracker](https://github.com/opd-ai/go-jf-org/issues)
- 💬 [Discussions](https://github.com/opd-ai/go-jf-org/discussions)

---

**Made with ❤️ for the Jellyfin community**
