package cmd

import (
	"fmt"

	"github.com/opd-ai/go-jf-org/internal/scanner"
	"github.com/opd-ai/go-jf-org/pkg/types"
)

// getDestinationRoot determines the destination directory for organization
// based on the provided dest flag or config based on media type
func getDestinationRoot(mediaType string, dest string) (string, error) {
	if dest != "" {
		return dest, nil
	}
	
	// Try to get from config based on media type
	if mediaType == "movie" && cfg.Destinations.Movies != "" {
		return cfg.Destinations.Movies, nil
	} else if mediaType == "tv" && cfg.Destinations.TV != "" {
		return cfg.Destinations.TV, nil
	} else if mediaType == "music" && cfg.Destinations.Music != "" {
		return cfg.Destinations.Music, nil
	} else if mediaType == "book" && cfg.Destinations.Books != "" {
		return cfg.Destinations.Books, nil
	}
	
	return "", fmt.Errorf("destination directory required (use --dest or configure in config file)")
}

// parseMediaTypeFilter converts a string media type to a MediaType enum
func parseMediaTypeFilter(mediaType string) (types.MediaType, error) {
	if mediaType == "" {
		return types.MediaTypeUnknown, nil
	}
	
	switch mediaType {
	case "movie":
		return types.MediaTypeMovie, nil
	case "tv":
		return types.MediaTypeTV, nil
	case "music":
		return types.MediaTypeMusic, nil
	case "book":
		return types.MediaTypeBook, nil
	default:
		return types.MediaTypeUnknown, fmt.Errorf("invalid media type: %s (must be movie, tv, music, or book)", mediaType)
	}
}

// Minimum file size for scanning (10MB)
const minFileSize = 10 * 1024 * 1024

// createScanner creates a new scanner with configuration from cfg
func createScanner() *scanner.Scanner {
	return scanner.NewScanner(
		cfg.Filters.VideoExtensions,
		cfg.Filters.AudioExtensions,
		cfg.Filters.BookExtensions,
		minFileSize,
	)
}
