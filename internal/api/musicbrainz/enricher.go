package musicbrainz

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/opd-ai/go-jf-org/pkg/types"
	"github.com/rs/zerolog/log"
)

const (
	// MinMusicYear is the earliest valid year for recorded music
	MinMusicYear = 1900
	// MaxMusicYear is the latest valid year (future releases)
	MaxMusicYear = 2100
)

// Enricher enriches metadata using MusicBrainz API
type Enricher struct {
	client *Client
}

// NewEnricher creates a new metadata enricher
func NewEnricher(client *Client) *Enricher {
	return &Enricher{client: client}
}

// EnrichMusic enriches music metadata with MusicBrainz data
func (e *Enricher) EnrichMusic(metadata *types.Metadata) error {
	if metadata == nil {
		return fmt.Errorf("metadata is nil")
	}

	// Ensure MusicMetadata exists
	if metadata.MusicMetadata == nil {
		metadata.MusicMetadata = &types.MusicMetadata{}
	}

	// Need at least album name or artist
	album := metadata.MusicMetadata.Album
	artist := metadata.MusicMetadata.Artist
	if metadata.MusicMetadata.AlbumArtist != "" {
		artist = metadata.MusicMetadata.AlbumArtist
	}

	if album == "" && artist == "" {
		return fmt.Errorf("album name or artist is required for enrichment")
	}

	log.Debug().
		Str("album", album).
		Str("artist", artist).
		Msg("Enriching music metadata")

	// Search for release
	searchResp, err := e.client.SearchRelease(album, artist)
	if err != nil {
		return fmt.Errorf("failed to search release: %w", err)
	}

	if searchResp.Count == 0 {
		log.Warn().
			Str("album", album).
			Str("artist", artist).
			Msg("No MusicBrainz results found for release")
		return nil // Not an error, just no results
	}

	// Use first result (best match)
	release := searchResp.Releases[0]

	// Get detailed information
	details, err := e.client.GetReleaseDetails(release.ID)
	if err != nil {
		log.Warn().Err(err).Str("id", release.ID).Msg("Failed to get release details")
		// Use search result data only
		e.applyReleaseSearchResult(metadata, &release)
		return nil
	}

	// Apply enriched metadata
	e.applyReleaseDetails(metadata, details)

	log.Info().
		Str("album", metadata.MusicMetadata.Album).
		Str("artist", metadata.MusicMetadata.Artist).
		Str("musicbrainz_id", details.ID).
		Msg("Music metadata enriched")

	return nil
}

// applyReleaseSearchResult applies metadata from a release search result
func (e *Enricher) applyReleaseSearchResult(metadata *types.Metadata, release *Release) {
	// Set album title
	if metadata.MusicMetadata.Album == "" {
		metadata.MusicMetadata.Album = release.Title
	}

	// Set artist from artist credit
	if len(release.ArtistCredit) > 0 && metadata.MusicMetadata.Artist == "" {
		metadata.MusicMetadata.Artist = release.ArtistCredit[0].Artist.Name
		metadata.MusicMetadata.AlbumArtist = release.ArtistCredit[0].Artist.Name
	}

	// Set year from release date
	if metadata.Year == 0 && release.Date != "" {
		year := e.extractYear(release.Date)
		if year > 0 {
			metadata.Year = year
		}
	}

	// Set MusicBrainz ID
	metadata.MusicMetadata.MusicBrainzRID = release.ID

	// Set release group ID if available
	if release.ReleaseGroup.ID != "" {
		metadata.MusicMetadata.MusicBrainzID = release.ReleaseGroup.ID
	}
}

// applyReleaseDetails applies metadata from detailed release information
func (e *Enricher) applyReleaseDetails(metadata *types.Metadata, details *ReleaseDetails) {
	// Set album title
	if metadata.MusicMetadata.Album == "" {
		metadata.MusicMetadata.Album = details.Title
	}
	if metadata.Title == "" {
		metadata.Title = details.Title
	}

	// Set artist from artist credit
	if len(details.ArtistCredit) > 0 {
		if metadata.MusicMetadata.Artist == "" {
			metadata.MusicMetadata.Artist = details.ArtistCredit[0].Artist.Name
		}
		if metadata.MusicMetadata.AlbumArtist == "" {
			metadata.MusicMetadata.AlbumArtist = details.ArtistCredit[0].Artist.Name
		}
	}

	// Set year from release date
	if metadata.Year == 0 && details.Date != "" {
		year := e.extractYear(details.Date)
		if year > 0 {
			metadata.Year = year
		}
	}

	// Set MusicBrainz IDs
	metadata.MusicMetadata.MusicBrainzRID = details.ID
	if details.ReleaseGroup.ID != "" {
		metadata.MusicMetadata.MusicBrainzID = details.ReleaseGroup.ID
	}

	// Set genre from release group primary type
	if metadata.MusicMetadata.Genre == "" && details.ReleaseGroup.PrimaryType != "" {
		metadata.MusicMetadata.Genre = details.ReleaseGroup.PrimaryType
	}

	log.Debug().
		Str("album", metadata.MusicMetadata.Album).
		Str("artist", metadata.MusicMetadata.Artist).
		Int("year", metadata.Year).
		Str("musicbrainz_rid", metadata.MusicMetadata.MusicBrainzRID).
		Msg("Applied MusicBrainz metadata")
}

// extractYear extracts year from date string (YYYY-MM-DD or YYYY)
func (e *Enricher) extractYear(dateStr string) int {
	if dateStr == "" {
		return 0
	}

	// Try to extract year from YYYY-MM-DD format
	parts := strings.Split(dateStr, "-")
	if len(parts) > 0 {
		year, err := strconv.Atoi(parts[0])
		if err == nil && year >= MinMusicYear && year <= MaxMusicYear {
			return year
		}
	}

	return 0
}
