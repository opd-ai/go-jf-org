# Filename Pattern Examples

This document shows examples of filenames that are successfully parsed by go-jf-org's metadata extraction system.

## Movies

The movie detector and parser recognize various filename patterns commonly used for movie files.

### Supported Patterns

#### Year-based Detection
Movies are identified primarily by year patterns (1850-2100):
- `(YYYY)` - Year in parentheses
- `.YYYY.` - Year with dots
- `[YYYY]` - Year in brackets
- `_YYYY_` - Year with underscores

#### Quality Tags
- `1080p`, `720p`, `480p`, `2160p` - Resolution
- `4K`, `8K`, `UHD`, `HD` - Quality descriptors

#### Source Tags
- `BluRay`, `Blu-Ray`, `BRRip`, `BDRip` - Blu-ray sources
- `WEB-DL`, `WEBRip`, `WEBDL` - Web sources
- `DVDRip`, `DVD-Rip` - DVD sources
- `HDTV`, `PDTV`, `HDRip` - TV sources

#### Codec Tags
- `x264`, `x265`, `h264`, `h265`, `HEVC`, `AVC`, `XviD`

### Examples

| Filename | Detected Metadata |
|----------|-------------------|
| `The.Matrix.1999.1080p.BluRay.x264.mkv` | **Title:** The Matrix (1999)<br>**Quality:** 1080P<br>**Source:** BluRay<br>**Codec:** x264 |
| `Inception (2010).mp4` | **Title:** Inception (2010) |
| `Blade.Runner.2049.2160p.4K.UHD.BluRay.x265.mkv` | **Title:** Blade Runner (2049)<br>**Quality:** 2160P<br>**Source:** BluRay<br>**Codec:** x265 |
| `The.Dark.Knight.[2008].720p.BDRip.mkv` | **Title:** The Dark Knight (2008)<br>**Quality:** 720P<br>**Source:** BDRip |
| `Movie.Title.2023.1080p.WEB-DL.h264.mkv` | **Title:** Movie Title (2023)<br>**Quality:** 1080P<br>**Source:** WEB-DL<br>**Codec:** h264 |
| `Some_Movie_Title_2020_720p.mp4` | **Title:** Some Movie Title (2020)<br>**Quality:** 720P |

## TV Shows

The TV detector and parser recognize episode numbering patterns to identify TV shows.

### Supported Patterns

#### Season/Episode Formats
- `S##E##` - Standard format (S01E01, S1E1, s01e01)
- `##x##` - Alternative format (1x01, 01x05)
- Case insensitive

#### Episode Title Extraction
Extracts episode titles when present between episode number and quality tags.

### Examples

| Filename | Detected Metadata |
|----------|-------------------|
| `Breaking.Bad.S01E01.Pilot.720p.mkv` | **Show:** Breaking Bad<br>**Season:** 1<br>**Episode:** 1<br>**Title:** Pilot |
| `game.of.thrones.s05e09.1080p.mp4` | **Show:** game of thrones<br>**Season:** 5<br>**Episode:** 9 |
| `The.Office.S02E15.The.Big.Job.720p.WEB-DL.mkv` | **Show:** The Office<br>**Season:** 2<br>**Episode:** 15<br>**Title:** The Big Job |
| `Show.Name.1x01.Episode.mkv` | **Show:** Show Name<br>**Season:** 1<br>**Episode:** 1 |
| `Series.Title.01x05.720p.mkv` | **Show:** Series Title<br>**Season:** 1<br>**Episode:** 5 |
| `Anime.Series.S1E5.mkv` | **Show:** Anime Series<br>**Season:** 1<br>**Episode:** 5 |
| `Game.of.Thrones.S05E09.The.Dance.of.Dragons.1080p.mkv` | **Show:** Game of Thrones<br>**Season:** 5<br>**Episode:** 9<br>**Title:** The Dance of Dragons |
| `My_Show_S02E10_720p.mkv` | **Show:** My Show<br>**Season:** 2<br>**Episode:** 10 |

## Music

Music files are identified by their audio file extensions:
- `.mp3`, `.flac`, `.m4a`, `.ogg`, `.opus`, `.wav`, `.aac`, `.wma`, `.ape`, `.alac`

**Note:** Detailed music metadata parsing (artist, album, track) is not yet implemented in this phase.

## Books

Book files are identified by their ebook/document extensions:
- `.epub`, `.mobi`, `.pdf`, `.azw3`, `.cbz`, `.cbr`

**Note:** Detailed book metadata parsing is not yet implemented in this phase.

## Edge Cases Handled

### Special Characters
- Dots (`.`) → Converted to spaces in titles
- Underscores (`_`) → Converted to spaces in titles
- Mixed case → Preserved in titles

### Ambiguous Files
- Video files without clear patterns default to **movie** type
- Files must be larger than the configured minimum size (default: 10MB)

## Testing Your Filenames

Use the scan command with verbose mode to see how your filenames are parsed:

```bash
go-jf-org scan /path/to/media -v
```

Example output:
```
Files found:
  [movie] /tmp/test-media/The.Matrix.1999.1080p.BluRay.x264.mkv
          Title: The Matrix (1999)
          Quality: 1080P  Source: BluRay  Codec: x264
  
  [tv] /tmp/test-media/Breaking.Bad.S01E01.Pilot.720p.mkv
          Show: Breaking Bad  S01E01  Pilot
  
  [music] /tmp/test-media/Pink.Floyd.-.Dark.Side.of.the.Moon.mp3
```

## Future Enhancements

The following features are planned for future releases:
- Music metadata extraction from ID3 tags and filenames
- Book metadata extraction from embedded metadata
- External API enrichment (TMDB, MusicBrainz, OpenLibrary)
- Support for more complex naming patterns
- Multi-part movies and special editions
- Anime-specific patterns
