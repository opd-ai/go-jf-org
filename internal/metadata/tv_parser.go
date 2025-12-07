package metadata

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/opd-ai/go-jf-org/pkg/types"
)

// TVParser parses TV show filenames
type TVParser interface {
	Parse(filename string) (*types.Metadata, error)
}

type tvParser struct {
	// Pattern for S01E01 format
	seasonEpisodePattern *regexp.Regexp
	// Pattern for 1x01 format
	altPattern *regexp.Regexp
	// Pattern to extract show name before season/episode
	showNamePattern *regexp.Regexp
}

// NewTVParser creates a new TVParser
func NewTVParser() TVParser {
	return &tvParser{
		// Capture season and episode numbers from S##E## pattern
		seasonEpisodePattern: regexp.MustCompile(`(?i)S(\d{1,4})E(\d{1,4})`),
		// Capture season and episode from ##x## pattern
		altPattern: regexp.MustCompile(`(?i)(\d{1,4})x(\d{1,4})`),
		// Capture everything before the season/episode pattern as show name
		showNamePattern: regexp.MustCompile(`^(.+?)[\._\s-]+(?i)(?:S?\d{1,4}[xE]\d{1,4})`),
	}
}

// Parse extracts metadata from a TV show filename
func (t *tvParser) Parse(filename string) (*types.Metadata, error) {
	metadata := &types.Metadata{
		TVMetadata: &types.TVMetadata{},
	}

	name := removeExtension(filename)

	// Extract season and episode numbers
	var season, episode int
	var err error

	// Try standard S01E01 pattern first
	matches := t.seasonEpisodePattern.FindStringSubmatch(name)
	if len(matches) >= 3 {
		season, err = strconv.Atoi(matches[1])
		if err == nil {
			metadata.TVMetadata.Season = season
		}
		episode, err = strconv.Atoi(matches[2])
		if err == nil {
			metadata.TVMetadata.Episode = episode
		}
	} else {
		// Try alternative 1x01 pattern
		altMatches := t.altPattern.FindStringSubmatch(name)
		if len(altMatches) >= 3 {
			season, err = strconv.Atoi(altMatches[1])
			if err == nil {
				metadata.TVMetadata.Season = season
			}
			episode, err = strconv.Atoi(altMatches[2])
			if err == nil {
				metadata.TVMetadata.Episode = episode
			}
		}
	}

	// Extract show name (everything before the season/episode pattern)
	showMatches := t.showNamePattern.FindStringSubmatch(name)
	if len(showMatches) >= 2 {
		showName := showMatches[1]
		// Clean up show name
		showName = strings.ReplaceAll(showName, ".", " ")
		showName = strings.ReplaceAll(showName, "_", " ")
		showName = strings.TrimSpace(showName)
		metadata.TVMetadata.ShowTitle = showName
		metadata.Title = showName
	}

	// Try to extract episode title (text after episode number before quality tags)
	// This is more complex and optional for now
	episodeTitlePattern := regexp.MustCompile(`(?i)S?\d{1,4}[xE]\d{1,4}[\.\s-]+(.+?)[\.\s-]+(?:\d{3,4}p|BluRay|WEB|HDTV|x26[45])`)
	episodeMatches := episodeTitlePattern.FindStringSubmatch(name)
	if len(episodeMatches) >= 2 {
		episodeTitle := episodeMatches[1]
		episodeTitle = strings.ReplaceAll(episodeTitle, ".", " ")
		episodeTitle = strings.ReplaceAll(episodeTitle, "_", " ")
		episodeTitle = strings.TrimSpace(episodeTitle)
		metadata.TVMetadata.EpisodeTitle = episodeTitle
	}

	return metadata, nil
}
