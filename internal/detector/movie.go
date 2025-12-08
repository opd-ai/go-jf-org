package detector

import (
	"regexp"
	"strings"

	"github.com/opd-ai/go-jf-org/internal/util"
)

// MovieDetector detects if a video file is a movie
type MovieDetector interface {
	IsMovie(filename string) bool
}

type movieDetector struct {
	yearPattern *regexp.Regexp
}

// NewMovieDetector creates a new MovieDetector
func NewMovieDetector() MovieDetector {
	return &movieDetector{
		// Match year patterns like (2023), .2023., [2023], _2023_, or at end of string
		// Years from 1850-2199 (extended to cover 21st century beyond 2100)
		yearPattern: regexp.MustCompile(`[\[\(._\s](18[5-9]\d|19\d{2}|20\d{2}|21\d{2})(?:[\]\)._\s]|$)`),
	}
}

// IsMovie returns true if the filename appears to be a movie
func (m *movieDetector) IsMovie(filename string) bool {
	// Remove extension for analysis
	name := util.RemoveExtension(filename)
	name = strings.ToLower(name)

	// If it has a year pattern, it's likely a movie
	if m.yearPattern.MatchString(name) {
		// Make sure it doesn't have TV show patterns
		// (year alone is not definitive, but it's a strong indicator for movies)
		return true
	}

	// Check for common movie quality/source tags
	// These are more common in movies than TV shows
	movieTags := []string{
		"bluray", "blu-ray", "brrip", "bdrip",
		"webrip", "web-dl", "webdl",
		"dvdrip", "dvd-rip",
		"hdrip", "hdtv",
		"1080p", "720p", "2160p", "4k",
		"x264", "x265", "h264", "h265", "hevc",
	}

	for _, tag := range movieTags {
		if strings.Contains(name, tag) {
			return true
		}
	}

	// If no specific indicators, we can't definitively say it's a movie
	return false
}
