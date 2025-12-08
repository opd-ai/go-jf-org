package metadata

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/opd-ai/go-jf-org/internal/util"
	"github.com/opd-ai/go-jf-org/pkg/types"
)

// MovieParser parses movie filenames
type MovieParser interface {
	Parse(filename string) (*types.Metadata, error)
}

type movieParser struct {
	// Pattern to extract title and year: Title.2023.Quality.Source.mkv
	titleYearPattern *regexp.Regexp
	// Pattern for quality tags (1080p, 720p, 4K, etc.)
	qualityPattern *regexp.Regexp
	// Pattern for source tags (BluRay, WEB-DL, etc.)
	sourcePattern *regexp.Regexp
	// Pattern for codec tags (x264, h265, etc.)
	codecPattern *regexp.Regexp
	// Pattern to extract just the year
	yearPattern *regexp.Regexp
}

// NewMovieParser creates a new MovieParser
func NewMovieParser() MovieParser {
	return &movieParser{
		// Capture title (non-greedy) and year
		// Supports years 1850-2199 (extended to cover 21st century beyond 2100)
		titleYearPattern: regexp.MustCompile(`^(.+?)[\[\(._\s]+(18[5-9]\d|19\d{2}|20\d{2}|21\d{2})[\]\)._\s]*`),
		qualityPattern:   regexp.MustCompile(`(?i)(4K|8K|2160p|1080p|720p|480p|UHD|HD)`),
		sourcePattern:    regexp.MustCompile(`(?i)(BluRay|Blu-Ray|BRRip|BDRip|WEB-DL|WEBRip|WEBDL|DVDRip|DVD-Rip|HDTV|PDTV|HDRip)`),
		codecPattern:     regexp.MustCompile(`(?i)(x264|x265|h264|h265|HEVC|AVC|XviD)`),
		yearPattern:      regexp.MustCompile(`[\[\(._\s](18[5-9]\d|19\d{2}|20\d{2}|21\d{2})[\]\)._\s]`),
	}
}

// Parse extracts metadata from a movie filename
func (m *movieParser) Parse(filename string) (*types.Metadata, error) {
	metadata := &types.Metadata{
		MovieMetadata: &types.MovieMetadata{},
	}

	// Remove extension
	name := util.RemoveExtension(filename)

	// Extract title and year
	matches := m.titleYearPattern.FindStringSubmatch(name)
	if len(matches) >= 3 {
		// Clean up title - replace dots and underscores with spaces
		title := util.CleanTitle(matches[1])
		metadata.Title = title

		// Parse year
		year, err := strconv.Atoi(matches[2])
		if err == nil {
			metadata.Year = year
		}
	} else {
		// Try to extract just the year if title+year pattern didn't match
		yearMatches := m.yearPattern.FindStringSubmatch(name)
		if len(yearMatches) >= 2 {
			year, err := strconv.Atoi(yearMatches[1])
			if err == nil {
				metadata.Year = year
			}
		}

		// Use filename as title if no better option
		if metadata.Title == "" {
			metadata.Title = util.CleanTitle(name)
		}
	}

	// Extract quality
	if qualityMatch := m.qualityPattern.FindString(name); qualityMatch != "" {
		metadata.Quality = strings.ToUpper(qualityMatch)
	}

	// Extract source
	if sourceMatch := m.sourcePattern.FindString(name); sourceMatch != "" {
		metadata.Source = sourceMatch
	}

	// Extract codec
	if codecMatch := m.codecPattern.FindString(name); codecMatch != "" {
		metadata.Codec = strings.ToLower(codecMatch)
	}

	return metadata, nil
}
