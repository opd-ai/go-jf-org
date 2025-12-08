# TMDB API Integration Guide

## Overview

The go-jf-org tool now includes TMDB (The Movie Database) API integration to enrich movie and TV show metadata. This provides accurate plot summaries, ratings, genres, and artwork URLs from the authoritative TMDB database.

## Setup

### 1. Get a TMDB API Key

1. Create a free account at [https://www.themoviedb.org/](https://www.themoviedb.org/)
2. Navigate to Settings → API
3. Request an API key (choose "Developer" option)
4. Copy your API key (v3 auth)

### 2. Configure go-jf-org

Add your API key to the configuration file (`~/.go-jf-org/config.yaml`):

```yaml
api_keys:
  tmdb: "your-api-key-here"
```

Or set it as an environment variable:

```bash
export GO_JF_ORG_API_KEYS_TMDB="your-api-key-here"
```

## Usage

### Scan with Enrichment

Use the `--enrich` flag with the scan command to fetch metadata from TMDB:

```bash
# Basic scan with enrichment
go-jf-org scan /media/unsorted --enrich -v

# Output example:
Files found:
  [movie] /media/The.Matrix.1999.1080p.mkv
          Title: The Matrix (1999)
          Quality: 1080P  Source: BluRay  Codec: x264
          Plot: A computer hacker learns from mysterious rebels about the true nature of his reality...
          Rating: 8.7/10
          Genres: [Action Science Fiction]

  [tv] /media/Breaking.Bad.S01E01.720p.mkv
          Show: Breaking Bad  S01E01  Pilot
          Plot: High school chemistry teacher Walter White's life is suddenly transformed...
          Rating: 9.2/10
          Genres: [Drama Crime]
```

### Organize with Enriched Metadata

When organizing files, enriched metadata improves NFO file quality:

```bash
# Organize with TMDB enrichment (planned feature)
go-jf-org organize /media/unsorted --dest /media/jellyfin --enrich --create-nfo
```

## Features

### Automatic Metadata Enrichment

When enrichment is enabled, go-jf-org:

1. **Parses** the filename to extract title, year, season, episode
2. **Searches** TMDB for the best match based on parsed data
3. **Fetches** detailed metadata including:
   - Plot synopsis
   - Ratings and vote counts
   - Genres
   - Runtime (movies)
   - IMDB ID (movies)
   - Poster and backdrop image URLs
   - Original titles
   - Release dates

4. **Enriches** the local metadata with TMDB data
5. **Caches** the response locally for 24 hours

### Rate Limiting

The TMDB client implements automatic rate limiting:

- **Limit**: 40 requests per 10 seconds (TMDB's limit)
- **Algorithm**: Token bucket with automatic refill
- **Behavior**: Requests are automatically queued when limit is reached

You don't need to worry about rate limiting - the tool handles it automatically!

### Caching

All TMDB API responses are cached locally:

- **Location**: `~/.go-jf-org/cache/tmdb/`
- **TTL**: 24 hours for successful responses, 1 hour for "not found"
- **Benefits**:
  - Faster subsequent scans
  - Reduced API usage
  - Works offline for cached entries
  - Privacy-friendly (local storage)

#### Cache Management

```bash
# Clear cache (future feature)
go-jf-org cache clear --tmdb

# View cache statistics (future feature)
go-jf-org cache stats
```

## Configuration Options

### API Rate Limiting

Adjust the rate limit in your config (default matches TMDB's limit):

```yaml
performance:
  api_rate_limit: 40  # Requests per 10 seconds
```

### Cache TTL

Configure how long to cache responses:

```yaml
performance:
  cache_ttl: 24h  # Can use: 1h, 24h, 7d, etc.
```

## Error Handling

### Missing API Key

If the TMDB API key is not configured:

```
WARN: TMDB API key not configured, skipping enrichment. Set api_keys.tmdb in config.
```

Files will still be scanned with filename-only metadata.

### API Errors

If TMDB returns an error (invalid key, rate limit, etc.):

```
WARN: Failed to create TMDB client, skipping enrichment
```

The tool gracefully falls back to filename-only metadata.

### No Match Found

If TMDB doesn't have a match for a file:

```
WARN: No TMDB results found for movie: "Some Obscure Film"
```

The filename-parsed metadata is still used.

## Privacy & Data Usage

### What is Sent to TMDB

- Movie/TV titles (from parsed filenames)
- Years (from parsed filenames)

### What is NOT Sent

- File paths
- User information
- System information
- File hashes or content
- Personal data

All API requests use HTTPS and follow TMDB's terms of service.

## API Limits & Pricing

TMDB offers a **free API** with generous limits:

- **Rate Limit**: 40 requests per 10 seconds
- **Daily Limit**: No hard daily limit (be reasonable)
- **Cost**: Free forever
- **Attribution**: TMDB logo should be displayed if you publish metadata publicly

See [TMDB API Terms](https://www.themoviedb.org/documentation/api/terms-of-use) for details.

## Advanced Usage

### Offline Mode

Disable API calls and use only cached data:

```bash
# Planned feature
go-jf-org scan /media/unsorted --offline -v
```

### Custom Cache Directory

```bash
# Set via environment variable
export GO_JF_ORG_CACHE_DIR=/custom/cache/path
go-jf-org scan /media/unsorted --enrich -v
```

## Troubleshooting

### "Invalid API key" Error

**Problem**: TMDB returns "Invalid API key: You must be granted a valid key"

**Solution**:
1. Verify your API key is correct in the config file
2. Ensure you copied the v3 API key (not v4 access token)
3. Check that your API key is approved (may take a few minutes after creation)

### "Rate limit exceeded" Error

**Problem**: Too many requests sent too quickly

**Solution**:
- The tool should handle this automatically with built-in rate limiting
- If you see this, try increasing the delay between requests
- Check if other tools are using the same API key

### Cache Issues

**Problem**: Old/incorrect data being returned

**Solution**:
```bash
# Clear the cache directory
rm -rf ~/.go-jf-org/cache/tmdb/

# Or wait 24 hours for cache to expire
```

### No Results Found

**Problem**: TMDB can't find a match for your file

**Solution**:
- Check that the filename includes the year (e.g., "Movie (1999)")
- Verify the title spelling matches TMDB's database
- Try searching TMDB directly to confirm the movie/show exists
- The tool will still use filename-parsed metadata

## Technical Details

### API Endpoints Used

The TMDB client uses these endpoints:

- `GET /search/movie` - Search for movies by title and year
- `GET /movie/{id}` - Get detailed movie information
- `GET /search/tv` - Search for TV shows by name
- `GET /tv/{id}` - Get detailed TV show information

### Response Caching

Cache files are stored as JSON:

```
~/.go-jf-org/cache/tmdb/
├── a1b2c3d4e5f6.json  # Cached movie search
├── f6e5d4c3b2a1.json  # Cached movie details
└── ...
```

Each file contains:
```json
{
  "data": { ... },      // TMDB response data
  "timestamp": "...",   // When cached
  "ttl": 86400          // Time to live in seconds
}
```

### Rate Limiter Algorithm

Token bucket algorithm:

```
Capacity: 40 tokens
Refill: 40 tokens every 10 seconds (4 tokens/second)

Request flow:
1. Check if tokens available
2. If yes: consume token, allow request
3. If no: wait for refill, then allow request
```

## Future Enhancements

Planned features for TMDB integration:

- [ ] Artwork downloading (posters, backdrops)
- [ ] Enhanced NFO generation with TMDB data
- [ ] Interactive match selection when multiple results found
- [ ] Episode-level metadata for TV shows
- [ ] Cast and crew information
- [ ] Trailer URLs
- [ ] Collection information (for movie series)
- [ ] Multi-language metadata support

## Related Documentation

- [Metadata Sources](metadata-sources.md) - Overview of all metadata APIs
- [NFO Files](nfo-files.md) - NFO file generation
- [Configuration](../config.example.yaml) - Full configuration reference

## Support

For TMDB-specific issues:

- Check TMDB API status: [https://status.themoviedb.org/](https://status.themoviedb.org/)
- TMDB API documentation: [https://developers.themoviedb.org/3](https://developers.themoviedb.org/3)
- TMDB support forum: [https://www.themoviedb.org/talk](https://www.themoviedb.org/talk)

For go-jf-org issues:

- GitHub Issues: [https://github.com/opd-ai/go-jf-org/issues](https://github.com/opd-ai/go-jf-org/issues)
- Discussions: [https://github.com/opd-ai/go-jf-org/discussions](https://github.com/opd-ai/go-jf-org/discussions)
