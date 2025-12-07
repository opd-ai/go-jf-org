# Metadata Sources

This document describes the external metadata sources that go-jf-org uses to enrich media information.

## Overview

go-jf-org uses external APIs to fetch metadata for movies, TV shows, music, and books. This ensures accurate information and proper organization.

---

## Movies & TV Shows

### The Movie Database (TMDB)

**Website:** https://www.themoviedb.org/  
**API Docs:** https://developers.themoviedb.org/3

#### Features
- Comprehensive movie and TV show database
- Free API with generous rate limits (40 requests per 10 seconds)
- Rich metadata: plot, cast, crew, ratings, genres
- High-quality images (posters, backdrops, logos)
- Multiple languages supported
- External IDs (IMDB, TVDB, etc.)

#### API Key
Free API key required. Sign up at: https://www.themoviedb.org/settings/api

#### Usage in go-jf-org
```go
// Example metadata fetch
tmdb.SearchMovie("The Matrix", 1999)
// Returns: Title, Year, Plot, Cast, Director, Genres, TMDB ID, IMDB ID
```

#### Endpoints Used
- `/search/movie` - Search for movies
- `/search/tv` - Search for TV shows
- `/movie/{id}` - Get movie details
- `/tv/{id}` - Get TV show details
- `/tv/{id}/season/{season_number}` - Get season details
- `/tv/{id}/season/{season_number}/episode/{episode_number}` - Get episode details

#### Rate Limits
- 40 requests per 10 seconds
- 3000 requests per hour (not enforced strictly)

#### Caching Strategy
- Cache successful lookups for 24 hours
- Cache "not found" results for 1 hour
- Store locally to minimize API calls

---

## Music

### MusicBrainz

**Website:** https://musicbrainz.org/  
**API Docs:** https://musicbrainz.org/doc/MusicBrainz_API

#### Features
- Open music encyclopedia
- Free to use (no API key required)
- Comprehensive artist, album, track metadata
- Release dates, labels, genres
- ISRC, barcode, catalog numbers
- Relationships (collaborations, remixes, etc.)

#### Usage Guidelines
- Must identify application in User-Agent header
- Rate limit: 1 request per second (enforced)
- No more than 50 requests from the same IP in a short period

#### Usage in go-jf-org
```go
// Example metadata fetch
musicbrainz.SearchRelease("Dark Side of the Moon", "Pink Floyd")
// Returns: Album Title, Artist, Year, Tracks, Label, MusicBrainz ID
```

#### Endpoints Used
- `/ws/2/release` - Search for releases (albums)
- `/ws/2/artist` - Search for artists
- `/ws/2/recording` - Search for recordings (tracks)
- `/ws/2/release/{id}` - Get release details

#### User-Agent Format
```
go-jf-org/1.0 (https://github.com/opd-ai/go-jf-org)
```

### Cover Art Archive

**Website:** https://coverartarchive.org/  
**API Docs:** https://musicbrainz.org/doc/Cover_Art_Archive/API

#### Features
- Album cover artwork
- Associated with MusicBrainz
- Free to use
- Multiple image sizes available

#### Usage in go-jf-org
```go
// Fetch cover art using MusicBrainz Release ID
coverart.Get(musicbrainzReleaseID)
```

### Last.fm (Optional)

**Website:** https://www.last.fm/api  
**Use Case:** Fallback for album artwork and additional metadata

#### Features
- Album artwork
- Artist information
- Tags and genres
- Similar artists

---

## Books

### OpenLibrary

**Website:** https://openlibrary.org/  
**API Docs:** https://openlibrary.org/developers/api

#### Features
- Free, open book database
- ISBN lookup
- Author, title, publication date
- Cover images
- Book descriptions
- No API key required

#### Usage in go-jf-org
```go
// Example metadata fetch
openlibrary.SearchByISBN("9780743273565")
openlibrary.SearchByTitle("The Great Gatsby", "F. Scott Fitzgerald")
// Returns: Title, Author, Year, Publisher, ISBN, Description, Cover URL
```

#### Endpoints Used
- `/search.json` - Search books
- `/isbn/{isbn}.json` - Get book by ISBN
- `/books/{id}.json` - Get book details
- `/authors/{id}.json` - Get author details

#### Rate Limits
- Reasonable use policy (no hard limit documented)
- Recommend 1 request per second

### Google Books API (Optional)

**Website:** https://developers.google.com/books  
**Use Case:** Fallback for book metadata

#### Features
- Large book database
- Free tier available (no API key for basic use)
- Rich metadata
- Preview availability

---

## Metadata Extraction Pipeline

### Step 1: Local Extraction

Extract metadata from the file itself:

**Video Files:**
- Use `ffprobe` or `mediainfo`
- Extract: resolution, codec, duration, audio tracks, subtitles

**Audio Files:**
- Read ID3 tags or other embedded metadata
- Extract: artist, album, track number, genre, year

**Books:**
- Parse EPUB/MOBI metadata
- Extract: title, author, publisher, ISBN

### Step 2: Filename Parsing

Parse the filename to extract:
- Title
- Year
- Season/Episode (for TV)
- Quality (1080p, 4K, etc.)
- Source (BluRay, WEB-DL, etc.)

### Step 3: External API Lookup

1. **Search** using parsed title and year
2. **Match** results based on similarity
3. **Fetch** full metadata for best match
4. **Enrich** local metadata with API data

### Step 4: Confidence Scoring

Score matches based on:
- Title similarity (fuzzy matching)
- Year match (exact or ±1 year)
- External ID match (if available)

Only use metadata with confidence > 80%.

### Step 5: Caching

Cache all API responses locally:
```
~/.go-jf-org/cache/
├── tmdb/
│   ├── movie_603.json      # The Matrix
│   └── tv_1396.json        # Breaking Bad
├── musicbrainz/
│   └── release_a1ad30cb.json
└── openlibrary/
    └── isbn_9780743273565.json
```

---

## Error Handling

### API Unavailable
- Fall back to local metadata only
- Log warning
- Continue processing

### Rate Limit Exceeded
- Implement exponential backoff
- Queue requests for later
- Use cached data if available

### No Match Found
- Use filename as-is
- Log for manual review
- Allow interactive mode for user input

### Multiple Matches
- Score each match
- Use highest scoring match if confidence > 80%
- Otherwise, prompt user (interactive mode)

---

## Configuration

### API Keys (Optional)

```yaml
api_keys:
  tmdb: "your-tmdb-api-key"
  lastfm: "your-lastfm-api-key"  # optional
  
metadata:
  sources:
    movies: tmdb           # Primary source for movies
    tv: tmdb               # Primary source for TV
    music: musicbrainz     # Primary source for music
    books: openlibrary     # Primary source for books
  
  cache:
    enabled: true
    directory: ~/.go-jf-org/cache
    ttl: 24h               # Time to live for cached data
  
  matching:
    min_confidence: 0.80   # Minimum confidence for auto-match
    fuzzy_threshold: 0.85  # String similarity threshold
```

---

## Offline Mode

go-jf-org can work without external APIs:
- Use only local file metadata
- Parse filenames for basic info
- Skip artwork downloads
- No NFO enrichment

Enable with:
```bash
go-jf-org organize /path --offline
```

Or in config:
```yaml
metadata:
  offline_mode: true
```

---

## Privacy & Data Usage

### Data Sent to APIs
- Movie/TV titles and years (for search)
- Artist/album names (for music search)
- Book titles and ISBNs (for book search)

### Data NOT Sent
- File paths
- User information
- System information
- File hashes or content

### Caching Benefits
- Reduces API calls
- Faster subsequent runs
- Works better with rate limits
- Privacy-friendly (local storage)

---

## API Client Implementation

### Best Practices

1. **Respect Rate Limits**
   - Implement token bucket algorithm
   - Add delays between requests
   - Use exponential backoff on errors

2. **User-Agent Headers**
   - Identify application clearly
   - Include version and contact info
   - Follow API provider guidelines

3. **Error Handling**
   - Graceful degradation
   - Retry with backoff
   - Log failures for review

4. **Caching**
   - Cache all successful responses
   - Implement cache invalidation
   - Handle cache misses gracefully

5. **Testing**
   - Mock API responses in tests
   - Test rate limiting
   - Test error conditions

---

## Example API Calls

### TMDB Movie Search
```bash
curl "https://api.themoviedb.org/3/search/movie?api_key=YOUR_KEY&query=The%20Matrix&year=1999"
```

Response:
```json
{
  "results": [
    {
      "id": 603,
      "title": "The Matrix",
      "release_date": "1999-03-31",
      "overview": "A computer hacker learns...",
      "poster_path": "/f89U3ADr1oiB1s9GkdPOEpXUk5H.jpg"
    }
  ]
}
```

### MusicBrainz Release Search
```bash
curl "https://musicbrainz.org/ws/2/release?query=release:Dark%20Side%20of%20the%20Moon%20AND%20artist:Pink%20Floyd&fmt=json" \
  -H "User-Agent: go-jf-org/1.0"
```

Response:
```json
{
  "releases": [
    {
      "id": "a1ad30cb-b8c4-4d68-9253-15b18fcde1d1",
      "title": "The Dark Side of the Moon",
      "date": "1973-03-01",
      "artist-credit": [{"name": "Pink Floyd"}]
    }
  ]
}
```

### OpenLibrary ISBN Lookup
```bash
curl "https://openlibrary.org/isbn/9780743273565.json"
```

Response:
```json
{
  "title": "The Great Gatsby",
  "authors": [{"key": "/authors/OL102037A"}],
  "publish_date": "2004",
  "publishers": ["Scribner"]
}
```

---

## Future Enhancements

1. **Additional Sources**
   - AniDB for anime
   - Audible for audiobooks
   - Goodreads for book ratings

2. **Metadata Quality**
   - Machine learning for better matching
   - User feedback loop
   - Community corrections

3. **Performance**
   - Parallel API requests
   - Batch lookups
   - Pre-warming cache

4. **Features**
   - Subtitle download (OpenSubtitles)
   - Artwork optimization
   - Multi-language metadata

---

## References

- [TMDB API Documentation](https://developers.themoviedb.org/3)
- [MusicBrainz API Guide](https://musicbrainz.org/doc/MusicBrainz_API)
- [OpenLibrary API](https://openlibrary.org/developers/api)
- [Cover Art Archive](https://coverartarchive.org/)
- [MediaInfo Documentation](https://mediaarea.net/en/MediaInfo)
