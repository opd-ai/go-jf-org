package jellyfin

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/opd-ai/go-jf-org/pkg/types"
)

// Naming provides Jellyfin-compatible naming conventions for media files
type Naming struct{}

// NewNaming creates a new Naming instance
func NewNaming() *Naming {
	return &Naming{}
}

// GetMovieName returns the Jellyfin-compatible filename for a movie
// Format: "Movie Name (Year).ext"
func (n *Naming) GetMovieName(metadata *types.Metadata, ext string) string {
	if metadata == nil || metadata.Title == "" {
		return ""
	}

	title := SanitizeFilename(metadata.Title)
	
	if metadata.Year > 0 {
		return fmt.Sprintf("%s (%d)%s", title, metadata.Year, ext)
	}
	
	return fmt.Sprintf("%s%s", title, ext)
}

// GetMovieDir returns the Jellyfin-compatible directory name for a movie
// Format: "Movie Name (Year)/"
func (n *Naming) GetMovieDir(metadata *types.Metadata) string {
	if metadata == nil || metadata.Title == "" {
		return ""
	}

	title := SanitizeFilename(metadata.Title)
	
	if metadata.Year > 0 {
		return fmt.Sprintf("%s (%d)", title, metadata.Year)
	}
	
	return title
}

// GetTVShowName returns the Jellyfin-compatible filename for a TV episode
// Format: "Show Name - S##E## - Episode Title.ext"
func (n *Naming) GetTVShowName(metadata *types.Metadata, ext string) string {
	if metadata == nil || metadata.TVMetadata == nil {
		return ""
	}

	tv := metadata.TVMetadata
	show := SanitizeFilename(tv.ShowTitle)
	
	if show == "" {
		return ""
	}

	// Base format: "Show Name - S##E##"
	name := fmt.Sprintf("%s - S%02dE%02d", show, tv.Season, tv.Episode)
	
	// Add episode title if available
	if tv.EpisodeTitle != "" {
		episodeTitle := SanitizeFilename(tv.EpisodeTitle)
		name = fmt.Sprintf("%s - %s", name, episodeTitle)
	}
	
	return name + ext
}

// GetTVShowDir returns the Jellyfin-compatible show directory name
// Format: "Show Name/"
func (n *Naming) GetTVShowDir(metadata *types.Metadata) string {
	if metadata == nil || metadata.TVMetadata == nil {
		return ""
	}

	return SanitizeFilename(metadata.TVMetadata.ShowTitle)
}

// GetTVSeasonDir returns the Jellyfin-compatible season directory name
// Format: "Season ##/" or "Specials/" for season 0
func (n *Naming) GetTVSeasonDir(season int) string {
	if season == 0 {
		return "Specials"
	}
	return fmt.Sprintf("Season %02d", season)
}

// GetMusicDir returns the Jellyfin-compatible music directory structure
// Format: "Artist Name/Album Name (Year)/"
func (n *Naming) GetMusicDir(metadata *types.Metadata) (artist, album string) {
	if metadata == nil || metadata.MusicMetadata == nil {
		return "", ""
	}

	music := metadata.MusicMetadata
	artist = SanitizeFilename(music.Artist)
	if artist == "" {
		artist = "Unknown Artist"
	}

	albumName := SanitizeFilename(music.Album)
	if albumName == "" {
		albumName = "Unknown Album"
	}

	if metadata.Year > 0 {
		album = fmt.Sprintf("%s (%d)", albumName, metadata.Year)
	} else {
		album = albumName
	}

	return artist, album
}

// GetMusicTrackName returns the Jellyfin-compatible track filename
// Format: "## - Track Name.ext"
func (n *Naming) GetMusicTrackName(metadata *types.Metadata, ext string) string {
	if metadata == nil || metadata.MusicMetadata == nil {
		return ""
	}

	music := metadata.MusicMetadata
	title := SanitizeFilename(metadata.Title)
	
	if title == "" {
		title = "Unknown Track"
	}

	if music.TrackNumber > 0 {
		return fmt.Sprintf("%02d - %s%s", music.TrackNumber, title, ext)
	}

	return title + ext
}

// GetBookDir returns the Jellyfin-compatible book directory structure
// Format: "Author Last, First/Book Title (Year)/"
func (n *Naming) GetBookDir(metadata *types.Metadata) (author, book string) {
	if metadata == nil || metadata.BookMetadata == nil {
		return "", ""
	}

	authorName := SanitizeFilename(metadata.BookMetadata.Author)
	if authorName == "" {
		authorName = "Unknown Author"
	}

	// Try to format as "Last, First" if possible
	author = FormatAuthorName(authorName)

	title := SanitizeFilename(metadata.Title)
	if title == "" {
		title = "Unknown Book"
	}

	if metadata.Year > 0 {
		book = fmt.Sprintf("%s (%d)", title, metadata.Year)
	} else {
		book = title
	}

	return author, book
}

// GetBookName returns the Jellyfin-compatible book filename
// Format: "Book Title.ext"
func (n *Naming) GetBookName(metadata *types.Metadata, ext string) string {
	if metadata == nil {
		return ""
	}

	title := SanitizeFilename(metadata.Title)
	if title == "" {
		title = "Unknown Book"
	}

	return title + ext
}

// SanitizeFilename removes or replaces characters that are invalid in filenames
// Replaces <>:"/\|?* and removes leading/trailing dots and spaces
func SanitizeFilename(s string) string {
	// Replace invalid characters with safe alternatives
	replacements := map[rune]string{
		'<':  "",
		'>':  "",
		':':  " -",
		'"':  "'",
		'/':  "-",
		'\\': "-",
		'|':  "-",
		'?':  "",
		'*':  "",
	}

	var result strings.Builder
	for _, r := range s {
		if replacement, found := replacements[r]; found {
			result.WriteString(replacement)
		} else {
			result.WriteRune(r)
		}
	}

	cleaned := result.String()

	// Remove leading/trailing dots and spaces
	cleaned = strings.TrimSpace(cleaned)
	cleaned = strings.Trim(cleaned, ".")

	// Collapse multiple spaces into single space
	spaceRegex := regexp.MustCompile(`\s+`)
	cleaned = spaceRegex.ReplaceAllString(cleaned, " ")

	return cleaned
}

// FormatAuthorName attempts to format author name as "Last, First"
// If already in that format or only one name, returns as-is
func FormatAuthorName(author string) string {
	// Already in "Last, First" format
	if strings.Contains(author, ",") {
		return author
	}

	// Split by whitespace
	parts := strings.Fields(author)
	if len(parts) < 2 {
		// Single name, return as-is
		return author
	}

	// Assume last part is surname
	lastName := parts[len(parts)-1]
	firstName := strings.Join(parts[:len(parts)-1], " ")

	return fmt.Sprintf("%s, %s", lastName, firstName)
}

// BuildFullPath constructs the full path for a media file based on its type and metadata
func (n *Naming) BuildFullPath(destRoot string, mediaType types.MediaType, metadata *types.Metadata, ext string) string {
	switch mediaType {
	case types.MediaTypeMovie:
		dir := n.GetMovieDir(metadata)
		filename := n.GetMovieName(metadata, ext)
		if dir == "" || filename == "" {
			return ""
		}
		return filepath.Join(destRoot, dir, filename)

	case types.MediaTypeTV:
		if metadata.TVMetadata == nil {
			return ""
		}
		showDir := n.GetTVShowDir(metadata)
		seasonDir := n.GetTVSeasonDir(metadata.TVMetadata.Season)
		filename := n.GetTVShowName(metadata, ext)
		if showDir == "" || filename == "" {
			return ""
		}
		return filepath.Join(destRoot, showDir, seasonDir, filename)

	case types.MediaTypeMusic:
		artistDir, albumDir := n.GetMusicDir(metadata)
		filename := n.GetMusicTrackName(metadata, ext)
		if artistDir == "" || filename == "" {
			return ""
		}
		return filepath.Join(destRoot, artistDir, albumDir, filename)

	case types.MediaTypeBook:
		authorDir, bookDir := n.GetBookDir(metadata)
		filename := n.GetBookName(metadata, ext)
		if authorDir == "" || filename == "" {
			return ""
		}
		return filepath.Join(destRoot, authorDir, bookDir, filename)

	default:
		return ""
	}
}
