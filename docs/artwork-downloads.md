# Artwork Downloads

This guide explains how to use the artwork download feature in go-jf-org to automatically fetch and save poster images, album covers, and book covers alongside your organized media files.

## Overview

The artwork download feature integrates with the organization workflow to automatically download artwork for your media files. It supports:

- **Movies**: Poster and backdrop images from TMDB
- **TV Shows**: Show poster images from TMDB
- **Music**: Album cover art from Cover Art Archive (MusicBrainz)
- **Books**: Book covers from OpenLibrary

## Basic Usage

Enable artwork downloads with the `--download-artwork` flag when organizing files:

```bash
# Basic artwork download
./bin/go-jf-org organize /path/to/media --dest /organized --download-artwork

# With metadata enrichment (required for artwork)
./bin/go-jf-org organize /path/to/media --dest /organized --enrich --download-artwork

# Specify artwork size preference
./bin/go-jf-org organize /path/to/media --dest /organized --enrich --download-artwork --artwork-size large

# Preview what artwork would be downloaded (dry-run)
./bin/go-jf-org organize /path/to/media --dest /organized --enrich --download-artwork --dry-run
```

## Artwork Size Options

Use the `--artwork-size` flag to control the resolution of downloaded artwork:

- `small` - Smallest files, lowest resolution (good for storage-constrained systems)
- `medium` - Balanced quality and file size (default)
- `large` - High quality, larger file size
- `original` - Original resolution, largest file size

### Size Mappings by Media Type

**Movies (TMDB):**
- **Posters**: small=w185, medium=w500, large=w780, original=original
- **Backdrops**: small=w300, medium=w780, large=w1280, original=original

**TV Shows (TMDB):**
- **Posters**: small=w185, medium=w500, large=w780, original=original

**Music (Cover Art Archive):**
- **Album Covers**: small=250px, medium=500px, large=1200px, original=full resolution

**Books (OpenLibrary):**
- **Book Covers**: small=S, medium=M, large/original=L

## File Locations

Artwork files are saved in Jellyfin-compatible locations:

### Movies
```
Movie Name (Year)/
├── Movie Name (Year).mkv
├── poster.jpg           # Movie poster
└── backdrop.jpg         # Movie backdrop/fanart
```

### TV Shows
```
Show Name/
├── poster.jpg           # Show poster
├── Season 01/
│   ├── Show Name - S01E01 - Episode Title.mkv
│   └── ...
└── Season 02/
    └── ...
```

### Music
```
Artist/
└── Album (Year)/
    ├── 01 - Track.flac
    ├── 02 - Track.flac
    └── cover.jpg        # Album cover art
```

### Books
```
Author Last, First/
└── Book Title (Year)/
    ├── Book Title.epub
    └── cover.jpg        # Book cover
```

## Requirements

### Metadata Enrichment

Artwork downloads require metadata enrichment to be enabled. The organizer uses metadata from external APIs to locate artwork:

- **Movies/TV**: Uses TMDB API to get poster and backdrop URLs
- **Music**: Uses MusicBrainz release ID to find album art
- **Books**: Uses ISBN to locate book covers

Always use `--enrich` flag with `--download-artwork`:

```bash
./bin/go-jf-org organize /path/to/media --dest /organized --enrich --download-artwork
```

### API Access

- **TMDB**: Free tier works well for movies and TV shows
- **MusicBrainz/Cover Art Archive**: Free, no API key required (1 req/s rate limit)
- **OpenLibrary**: Free, no API key required

Configure TMDB API key in `~/.go-jf-org/config.yaml`:

```yaml
api:
  tmdb:
    api_key: "your-api-key-here"
```

## Behavior

### Success Cases

- Artwork is downloaded after successfully moving/organizing the media file
- If artwork already exists, it is not re-downloaded (unless using `--force`)
- Partial success is acceptable - organization continues even if artwork download fails
- Artwork operations are logged to transaction logs for rollback support

### Failure Cases

The organizer handles artwork download failures gracefully:

- **Missing artwork**: Logged as a warning, organization continues
- **Network errors**: Retried up to 3 times with exponential backoff
- **API failures**: Logged as a warning, organization continues
- **Missing metadata**: Artwork download is skipped

**Important**: Artwork download failures never cause the organization process to fail.

### Dry-Run Mode

Use `--dry-run` to preview artwork downloads without downloading:

```bash
./bin/go-jf-org organize /path/to/media --dest /organized --enrich --download-artwork --dry-run
```

This shows:
- Which files would be organized
- Which artwork files would be downloaded
- Download destinations and sources

### Transaction Logging

Artwork downloads are included in transaction logs:

- Each downloaded artwork file is logged as a `create_file` operation
- Source contains the URL or identifier (TMDB poster path, MusicBrainz ID, ISBN)
- Destination contains the local file path
- Rollback will remove downloaded artwork files

Example transaction entry:
```json
{
  "type": "create_file",
  "source": "/w500/abc123.jpg",
  "destination": "/organized/Movie (2023)/poster.jpg",
  "status": "completed"
}
```

## Examples

### Basic Movie Organization with Artwork

```bash
# Organize movies with artwork
./bin/go-jf-org organize /media/unsorted/movies \
  --dest /media/jellyfin/movies \
  --type movie \
  --enrich \
  --download-artwork \
  --artwork-size medium

# Expected output structure:
# /media/jellyfin/movies/
# ├── The Matrix (1999)/
# │   ├── The Matrix (1999).mkv
# │   ├── poster.jpg
# │   └── backdrop.jpg
# └── Inception (2010)/
#     ├── Inception (2010).mkv
#     ├── poster.jpg
#     └── backdrop.jpg
```

### TV Show Organization with Artwork

```bash
# Organize TV shows with artwork
./bin/go-jf-org organize /media/unsorted/tv \
  --dest /media/jellyfin/tv \
  --type tv \
  --enrich \
  --download-artwork

# Expected output structure:
# /media/jellyfin/tv/
# └── Breaking Bad/
#     ├── poster.jpg
#     ├── Season 01/
#     │   ├── Breaking Bad - S01E01 - Episode Title.mkv
#     │   └── Breaking Bad - S01E02 - Episode Title.mkv
#     └── Season 02/
#         └── ...
```

### Music Organization with Album Art

```bash
# Organize music with album art
./bin/go-jf-org organize /media/unsorted/music \
  --dest /media/jellyfin/music \
  --type music \
  --enrich \
  --download-artwork \
  --artwork-size large

# Expected output structure:
# /media/jellyfin/music/
# └── Pink Floyd/
#     └── The Dark Side of the Moon (1973)/
#         ├── 01 - Speak to Me.flac
#         ├── 02 - Breathe.flac
#         └── cover.jpg
```

### Book Organization with Covers

```bash
# Organize books with covers
./bin/go-jf-org organize /media/unsorted/books \
  --dest /media/jellyfin/books \
  --type book \
  --enrich \
  --download-artwork

# Expected output structure:
# /media/jellyfin/books/
# └── Asimov, Isaac/
#     └── Foundation (1951)/
#         ├── Foundation.epub
#         └── cover.jpg
```

### Preview Before Downloading

```bash
# Dry-run to see what would be downloaded
./bin/go-jf-org organize /media/unsorted \
  --dest /media/jellyfin \
  --enrich \
  --download-artwork \
  --dry-run

# Review the plan, then execute
./bin/go-jf-org organize /media/unsorted \
  --dest /media/jellyfin \
  --enrich \
  --download-artwork
```

## Performance Considerations

### Rate Limiting

The artwork downloaders respect API rate limits:

- **TMDB**: 40 requests per 10 seconds
- **MusicBrainz/Cover Art Archive**: 1 request per second (strictly enforced)
- **OpenLibrary**: No official limit, but downloads are throttled

### Caching

- Downloaded artwork is not re-downloaded if it already exists
- Use `--force` flag to override (not yet implemented in CLI)

### Network Usage

Approximate artwork sizes:

- **Small**: 50-150 KB per image
- **Medium**: 150-400 KB per image
- **Large**: 400-1000 KB per image
- **Original**: 1-5 MB per image

For 100 movies with `medium` size:
- Posters: ~25 MB
- Backdrops: ~35 MB
- Total: ~60 MB

## Troubleshooting

### No Artwork Downloaded

**Possible causes:**
1. Metadata enrichment not enabled (use `--enrich`)
2. No metadata found (API couldn't locate the title)
3. No artwork available for the media item
4. Network connectivity issues

**Solutions:**
- Ensure `--enrich` flag is used
- Check logs for warnings about missing artwork
- Verify API configuration and connectivity
- Use verbose mode (`-v`) for detailed logging

### Partial Artwork Missing

Some artwork may not be available for all media:
- Not all movies have backdrop images
- Some albums may not have cover art in Cover Art Archive
- Books may not have ISBN or covers in OpenLibrary

This is normal and expected.

### Download Errors

If downloads fail repeatedly:
1. Check network connectivity
2. Verify API keys are correct
3. Check API rate limits
4. Review logs for specific error messages

Use verbose logging to diagnose:
```bash
./bin/go-jf-org organize /path/to/media \
  --dest /organized \
  --enrich \
  --download-artwork \
  -v
```

## Configuration

Configure artwork settings in `~/.go-jf-org/config.yaml`:

```yaml
organize:
  download_artwork: true
  artwork_size: medium
  
api:
  tmdb:
    api_key: "your-api-key"
    # poster and backdrop downloads use TMDB image CDN
  
  musicbrainz:
    # Cover Art Archive downloads use MusicBrainz release IDs
    rate_limit: 1  # requests per second
  
  openlibrary:
    # Book cover downloads use OpenLibrary Covers API
```

## Related Documentation

- [Metadata Sources](metadata-sources.md) - Details about external APIs
- [Jellyfin Conventions](jellyfin-conventions.md) - File and directory naming standards
- [NFO Files](nfo-files.md) - Metadata file generation
- [Examples](examples.md) - More usage examples

## Limitations

Current limitations:

1. **No season posters**: Only show-level posters are downloaded for TV shows
2. **No episode thumbnails**: Individual episode artwork not yet supported
3. **No progress indicators**: Artwork downloads don't show individual progress (coming soon)
4. **No force flag in CLI**: Can't force re-download existing artwork via CLI yet

These features may be added in future releases.
