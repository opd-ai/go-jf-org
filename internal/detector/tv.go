package detector

import (
	"regexp"
	"strings"
)

// TVDetector detects if a video file is a TV show
type TVDetector interface {
	IsTV(filename string) bool
}

type tvDetector struct {
	// Standard TV patterns: S01E01, S1E1, etc.
	seasonEpisodePattern *regexp.Regexp
	// Alternative pattern: 1x01, 01x01, etc.
	altSeasonEpisodePattern *regexp.Regexp
	// Episode pattern without season: E01, E1, etc. (less reliable)
	episodeOnlyPattern *regexp.Regexp
}

// NewTVDetector creates a new TVDetector
func NewTVDetector() TVDetector {
	return &tvDetector{
		// Match patterns like S01E01, S1E1, s01e01 (case insensitive)
		seasonEpisodePattern: regexp.MustCompile(`(?i)s\d{1,4}e\d{1,4}`),
		// Match patterns like 1x01, 01x01 (more strict - needs digit before x)
		altSeasonEpisodePattern: regexp.MustCompile(`(?i)\d{1,4}x\d{1,4}`),
		// Match E01, E1, etc. (less reliable, used as secondary check)
		episodeOnlyPattern: regexp.MustCompile(`(?i)\.e\d{1,4}[\.\s-]`),
	}
}

// IsTV returns true if the filename appears to be a TV show episode
func (t *tvDetector) IsTV(filename string) bool {
	name := strings.ToLower(filename)

	// Check for standard season/episode pattern (S01E01)
	if t.seasonEpisodePattern.MatchString(name) {
		return true
	}

	// Check for alternative pattern (1x01)
	if t.altSeasonEpisodePattern.MatchString(name) {
		return true
	}

	// Check for episode-only pattern (less reliable)
	// Only return true if we also find TV-related keywords
	if t.episodeOnlyPattern.MatchString(name) {
		// Additional validation - look for common TV show indicators
		tvIndicators := []string{
			"episode", "season", "series",
			"hdtv", "pdtv", // TV-specific sources
		}
		for _, indicator := range tvIndicators {
			if strings.Contains(name, indicator) {
				return true
			}
		}
	}

	return false
}
