# NFO File Generation

go-jf-org can automatically generate Jellyfin-compatible NFO (iNFO) files for your media collection. NFO files are XML files that contain metadata about media items, helping Jellyfin identify and display your content correctly.

## What are NFO Files?

NFO files are Kodi-compatible XML files that Jellyfin uses for metadata. They complement Jellyfin's built-in scrapers and can help with:

- **Accurate identification** - NFO files can help Jellyfin correctly identify media when filenames are ambiguous
- **Offline metadata** - Metadata is available even when external APIs are unavailable
- **Custom metadata** - You can manually edit NFO files to add or correct information
- **Fast scanning** - Jellyfin can use NFO files without querying external services

## Enabling NFO Generation

Use the `--create-nfo` flag with the `organize` or `preview` commands:

```bash
# Organize files and create NFO files
go-jf-org organize /media/unsorted --create-nfo

# Preview what NFO files would be created
go-jf-org preview /media/unsorted --create-nfo
```

## Generated NFO Files

### Movies

For each movie, go-jf-org creates a `movie.nfo` file in the movie's directory:

**Directory Structure:**
```
The Matrix (1999)/
├── The Matrix (1999).mkv
└── movie.nfo
```

**NFO Content:**
```xml
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<movie>
    <title>The Matrix</title>
    <originaltitle>The Matrix</originaltitle>
    <year>1999</year>
</movie>
```

### TV Shows

For TV shows, go-jf-org creates three types of NFO files:

#### 1. tvshow.nfo (Show level)
Created in the show's root directory:

**Directory Structure:**
```
Breaking Bad/
├── tvshow.nfo
└── Season 01/
    └── ...
```

**NFO Content:**
```xml
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<tvshow>
    <title>Breaking Bad</title>
</tvshow>
```

#### 2. season.nfo (Season level)
Created in each season directory:

**Directory Structure:**
```
Breaking Bad/
└── Season 01/
    ├── season.nfo
    ├── Breaking Bad - S01E01.mkv
    └── Breaking Bad - S01E02.mkv
```

**NFO Content:**
```xml
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<season>
    <seasonnumber>1</seasonnumber>
</season>
```

### Music

For music albums, go-jf-org creates an `album.nfo` file in the album's directory:

**Directory Structure:**
```
Pink Floyd/
└── The Dark Side of the Moon (1973)/
    ├── 01 - Speak to Me.flac
    ├── 02 - Breathe.flac
    ├── album.nfo
    └── folder.jpg
```

**NFO Content:**
```xml
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<album>
    <title>The Dark Side of the Moon</title>
    <artist>Pink Floyd</artist>
    <albumartist>Pink Floyd</albumartist>
    <year>1973</year>
</album>
```

### Books

For books, go-jf-org creates a `book.nfo` file in the book's directory:

**Directory Structure:**
```
Fitzgerald, F. Scott/
└── The Great Gatsby (1925)/
    ├── The Great Gatsby.epub
    ├── book.nfo
    └── cover.jpg
```

**NFO Content:**
```xml
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<book>
    <title>The Great Gatsby</title>
    <author>F. Scott Fitzgerald</author>
    <year>1925</year>
</book>
```

## Current Limitations

### Metadata Source

Currently, NFO files are generated using **metadata extracted from filenames only**. This means:

- ✅ Title, year, season, and episode numbers are included
- ❌ Extended metadata (plot, cast, genres) requires filename-based metadata or manual editing
- 🔄 TMDB API integration (planned) will enable rich metadata in future versions

### Supported Media Types

| Media Type | NFO Generation | Status |
|------------|---------------|---------|
| Movies     | ✅ Supported   | movie.nfo created |
| TV Shows   | ✅ Supported   | tvshow.nfo, season.nfo created |
| Music      | ✅ Supported   | album.nfo created |
| Books      | ✅ Supported   | book.nfo created |

## Transaction Support

NFO file creation is **fully integrated with transaction logging**:

- ✅ NFO file operations are logged in transactions
- ✅ Rollback removes created NFO files
- ✅ Dry-run mode previews NFO creation without executing

**Example:**
```bash
# Organize with NFO and transaction logging
$ go-jf-org organize /media/unsorted --create-nfo
✓ Successfully organized: 5 files
  - 2 media files moved
  - 3 NFO files created
Transaction ID: d8f309ee07381295

# Rollback if needed - removes both media files AND NFO files
$ go-jf-org rollback d8f309ee07381295
✓ Rollback completed successfully
  - 2 media files moved back
  - 3 NFO files removed
```

## NFO File Format

go-jf-org generates **Kodi-compatible NFO files** that work with Jellyfin and other Kodi-based media centers. The XML format follows these standards:

- **XML 1.0** with UTF-8 encoding
- **Standalone** XML documents (no DTD)
- **Proper escaping** of special characters (&, <, >, ", ')
- **4-space indentation** for readability

## Manual Editing

You can manually edit generated NFO files to add additional metadata:

### Movie NFO Example
```xml
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<movie>
    <title>The Matrix</title>
    <originaltitle>The Matrix</originaltitle>
    <year>1999</year>
    <!-- You can add these manually: -->
    <plot>A computer hacker learns about the true nature of reality...</plot>
    <director>Lana Wachowski</director>
    <director>Lilly Wachowski</director>
    <genre>Action</genre>
    <genre>Sci-Fi</genre>
    <tmdbid>603</tmdbid>
    <imdbid>tt0133093</imdbid>
</movie>
```

### TV Show NFO Example
```xml
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<tvshow>
    <title>Breaking Bad</title>
    <!-- You can add these manually: -->
    <plot>A chemistry teacher diagnosed with cancer...</plot>
    <genre>Crime</genre>
    <genre>Drama</genre>
    <premiered>2008-01-20</premiered>
    <studio>AMC</studio>
    <tmdbid>1396</tmdbid>
    <tvdbid>81189</tvdbid>
</tvshow>
```

### Music Album NFO Example
```xml
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<album>
    <title>Abbey Road</title>
    <artist>The Beatles</artist>
    <albumartist>The Beatles</albumartist>
    <year>1969</year>
    <!-- You can add these manually: -->
    <genre>Rock</genre>
    <review>One of the most iconic albums of all time...</review>
    <musicbrainzalbumid>1234-5678-abcd</musicbrainzalbumid>
</album>
```

### Book NFO Example
```xml
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<book>
    <title>The Great Gatsby</title>
    <author>F. Scott Fitzgerald</author>
    <year>1925</year>
    <!-- You can add these manually: -->
    <publisher>Charles Scribner's Sons</publisher>
    <isbn>978-0-7432-7356-5</isbn>
    <description>A critique of the American Dream set in the Jazz Age...</description>
    <series>N/A</series>
</book>
```

## Best Practices

1. **Enable NFO generation from the start** - It's easier to create NFO files during organization than to add them later

2. **Use with transaction logging** - Don't use `--no-transaction` when creating NFO files, so you can roll back if needed

3. **Verify with preview** - Use `preview --create-nfo` to see what NFO files will be created before organizing

4. **Combine with API integration** (future) - Once TMDB integration is complete, NFO files will contain rich metadata automatically

## Future Enhancements

The following features are planned for future releases:

- 🔄 **Enhanced API integration** - Automatically populate music NFO files with metadata from MusicBrainz and book NFO files from OpenLibrary
- 🔄 **Episode NFO files** - Generate `<filename>.nfo` for individual episodes with episode-specific metadata
- 🔄 **NFO update mode** - Refresh existing NFO files with new metadata from APIs
- 🔄 **Custom NFO templates** - Allow users to customize NFO file structure

## Troubleshooting

### NFO Files Not Created

**Check:**
- ✅ Did you use the `--create-nfo` flag?
- ✅ Did the file organization succeed?
- ✅ Check logs with `-v` flag for errors

**Example:**
```bash
go-jf-org organize /media/unsorted --create-nfo -v
```

### Invalid XML in NFO Files

This shouldn't happen as go-jf-org generates valid XML, but if you manually edited an NFO file:

**Validation:**
```bash
# Check XML validity
xmllint --noout your-nfo-file.nfo

# Or use an online XML validator
```

### Jellyfin Not Reading NFO Files

**Jellyfin Configuration:**
1. Go to Dashboard → Libraries
2. Edit your library
3. Under "NFO Settings", ensure "Enable Kodi/NFO metadata" is checked
4. Rescan library

## See Also

- [Jellyfin NFO Documentation](https://jellyfin.org/docs/general/server/metadata/nfo/)
- [Kodi NFO Files](https://kodi.wiki/view/NFO_files)
- [Jellyfin Naming Conventions](jellyfin-conventions.md)
- [Metadata Sources](metadata-sources.md)
