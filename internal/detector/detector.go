package detector

import (
	"path/filepath"
	"strings"

	"github.com/opd-ai/go-jf-org/internal/util"
	"github.com/opd-ai/go-jf-org/pkg/types"
)

// Detector is responsible for detecting the media type of a file
type Detector interface {
	// Detect determines the media type based on the filename
	Detect(filename string) types.MediaType
}

// detector is the main implementation of Detector
type detector struct {
	movieDetector MovieDetector
	tvDetector    TVDetector
}

// New creates a new Detector instance
func New() Detector {
	return &detector{
		movieDetector: NewMovieDetector(),
		tvDetector:    NewTVDetector(),
	}
}

// Detect determines the media type based on filename patterns
func (d *detector) Detect(filename string) types.MediaType {
	// Get the base filename without path
	base := filepath.Base(filename)
	ext := strings.ToLower(filepath.Ext(base))

	// Determine if it's a video, audio, or book based on extension
	// Then use specific detectors to distinguish between movie/TV, etc.

	// Check if it's a video file
	if isVideoExtension(ext) {
		// Try TV detector first (more specific patterns)
		if d.tvDetector.IsTV(base) {
			return types.MediaTypeTV
		}
		// Try movie detector
		if d.movieDetector.IsMovie(base) {
			return types.MediaTypeMovie
		}
		// If no specific pattern matched, default to movie
		// (most single video files are movies)
		return types.MediaTypeMovie
	}

	// Audio files are music
	if isAudioExtension(ext) {
		return types.MediaTypeMusic
	}

	// Book extensions
	if isBookExtension(ext) {
		return types.MediaTypeBook
	}

	return types.MediaTypeUnknown
}

// Video extensions
var videoExtensions = []string{
	".mkv", ".mp4", ".avi", ".m4v", ".ts", ".webm",
	".mov", ".wmv", ".flv", ".mpg", ".mpeg",
}

// Audio extensions
var audioExtensions = []string{
	".mp3", ".flac", ".m4a", ".ogg", ".opus", ".wav",
	".aac", ".wma", ".ape", ".alac",
}

// Book extensions
var bookExtensions = []string{
	".epub", ".mobi", ".pdf", ".azw3", ".cbz", ".cbr",
}

func isVideoExtension(ext string) bool {
	return util.ContainsExtension(videoExtensions, ext)
}

func isAudioExtension(ext string) bool {
	return util.ContainsExtension(audioExtensions, ext)
}

func isBookExtension(ext string) bool {
	return util.ContainsExtension(bookExtensions, ext)
}
