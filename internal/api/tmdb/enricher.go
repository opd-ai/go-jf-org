package tmdb

import (
	"fmt"
	"strings"

	"github.com/opd-ai/go-jf-org/pkg/types"
	"github.com/rs/zerolog/log"
)

// Enricher enriches metadata using TMDB API
type Enricher struct {
	client *Client
}

// NewEnricher creates a new metadata enricher
func NewEnricher(client *Client) *Enricher {
	return &Enricher{client: client}
}

// EnrichMovie enriches movie metadata with TMDB data
func (e *Enricher) EnrichMovie(metadata *types.Metadata) error {
	if metadata == nil {
		return fmt.Errorf("metadata is nil")
	}

	if metadata.Title == "" {
		return fmt.Errorf("title is required for enrichment")
	}

	// Ensure MovieMetadata exists
	if metadata.MovieMetadata == nil {
		metadata.MovieMetadata = &types.MovieMetadata{}
	}

	log.Debug().
		Str("title", metadata.Title).
		Int("year", metadata.Year).
		Msg("Enriching movie metadata")

	// Search for movie
	searchResp, err := e.client.SearchMovie(metadata.Title, metadata.Year)
	if err != nil {
		return fmt.Errorf("failed to search movie: %w", err)
	}

	if len(searchResp.Results) == 0 {
		log.Warn().
			Str("title", metadata.Title).
			Int("year", metadata.Year).
			Msg("No TMDB results found for movie")
		return nil // Not an error, just no results
	}

	// Use first result (best match)
	movie := searchResp.Results[0]

	// Get detailed information
	details, err := e.client.GetMovieDetails(movie.ID)
	if err != nil {
		log.Warn().Err(err).Int("id", movie.ID).Msg("Failed to get movie details")
		// Use search result data only
		e.applyMovieSearchResult(metadata, &movie)
		return nil
	}

	// Apply enriched metadata
	e.applyMovieDetails(metadata, details)

	log.Info().
		Str("title", metadata.Title).
		Int("tmdb_id", details.ID).
		Str("imdb_id", details.IMDBID).
		Msg("Movie metadata enriched")

	return nil
}

// EnrichTVShow enriches TV show metadata with TMDB data
func (e *Enricher) EnrichTVShow(metadata *types.Metadata) error {
	if metadata == nil {
		return fmt.Errorf("metadata is nil")
	}

	// Ensure TVMetadata exists
	if metadata.TVMetadata == nil {
		metadata.TVMetadata = &types.TVMetadata{}
	}

	if metadata.TVMetadata.ShowTitle == "" && metadata.Title == "" {
		return fmt.Errorf("show name is required for enrichment")
	}

	// Use ShowTitle from TVMetadata, fallback to Title
	showName := metadata.TVMetadata.ShowTitle
	if showName == "" {
		showName = metadata.Title
	}

	log.Debug().
		Str("show", showName).
		Msg("Enriching TV show metadata")

	// Extract year from show name if present
	year := 0
	if metadata.Year > 0 {
		year = metadata.Year
	}

	// Search for TV show
	searchResp, err := e.client.SearchTV(showName, year)
	if err != nil {
		return fmt.Errorf("failed to search TV show: %w", err)
	}

	if len(searchResp.Results) == 0 {
		log.Warn().
			Str("show", showName).
			Msg("No TMDB results found for TV show")
		return nil
	}

	// Use first result
	show := searchResp.Results[0]

	// Get detailed information
	details, err := e.client.GetTVDetails(show.ID)
	if err != nil {
		log.Warn().Err(err).Int("id", show.ID).Msg("Failed to get TV details")
		e.applyTVSearchResult(metadata, &show)
		return nil
	}

	// Apply enriched metadata
	e.applyTVDetails(metadata, details)

	log.Info().
		Str("show", showName).
		Int("tmdb_id", details.ID).
		Msg("TV show metadata enriched")

	return nil
}

// applyMovieSearchResult applies data from search result to metadata
func (e *Enricher) applyMovieSearchResult(metadata *types.Metadata, movie *MovieResult) {
	metadata.MovieMetadata.Plot = movie.Overview
	metadata.MovieMetadata.Rating = movie.VoteAverage

	// Extract year from release date if not already set
	if metadata.Year == 0 && movie.ReleaseDate != "" {
		parts := strings.Split(movie.ReleaseDate, "-")
		if len(parts) > 0 {
			var year int
			fmt.Sscanf(parts[0], "%d", &year)
			if year > 0 {
				metadata.Year = year
			}
		}
	}

	// Store TMDB ID for reference
	metadata.MovieMetadata.TMDBID = movie.ID

	// Build poster URL if available
	if movie.PosterPath != "" {
		metadata.MovieMetadata.PosterURL = fmt.Sprintf("https://image.tmdb.org/t/p/w500%s", movie.PosterPath)
	}
}

// applyMovieDetails applies detailed movie data to metadata
func (e *Enricher) applyMovieDetails(metadata *types.Metadata, details *MovieDetails) {
	// Title (prefer original if set, otherwise use TMDB title)
	if metadata.Title == "" {
		metadata.Title = details.Title
	}

	metadata.MovieMetadata.Plot = details.Overview
	metadata.MovieMetadata.Rating = details.VoteAverage
	metadata.MovieMetadata.TMDBID = details.ID
	metadata.MovieMetadata.IMDBID = details.IMDBID
	metadata.MovieMetadata.Runtime = details.Runtime

	// Extract year from release date
	if details.ReleaseDate != "" {
		parts := strings.Split(details.ReleaseDate, "-")
		if len(parts) > 0 {
			var year int
			fmt.Sscanf(parts[0], "%d", &year)
			if year > 0 && metadata.Year == 0 {
				metadata.Year = year
			}
		}
	}

	// Genres
	if len(details.Genres) > 0 {
		metadata.MovieMetadata.Genres = make([]string, len(details.Genres))
		for i, genre := range details.Genres {
			metadata.MovieMetadata.Genres[i] = genre.Name
		}
	}

	// Poster URL
	if details.PosterPath != "" {
		metadata.MovieMetadata.PosterURL = fmt.Sprintf("https://image.tmdb.org/t/p/w500%s", details.PosterPath)
	}

	// Backdrop URL
	if details.BackdropPath != "" {
		metadata.MovieMetadata.BackdropURL = fmt.Sprintf("https://image.tmdb.org/t/p/w1280%s", details.BackdropPath)
	}

	metadata.MovieMetadata.Tagline = details.Tagline
}

// applyTVSearchResult applies data from TV search result to metadata
func (e *Enricher) applyTVSearchResult(metadata *types.Metadata, show *TVResult) {
	metadata.TVMetadata.Plot = show.Overview
	metadata.TVMetadata.Rating = show.VoteAverage
	metadata.TVMetadata.TMDBID = show.ID

	// Extract year from first air date
	if show.FirstAirDate != "" {
		parts := strings.Split(show.FirstAirDate, "-")
		if len(parts) > 0 {
			var year int
			fmt.Sscanf(parts[0], "%d", &year)
			if year > 0 && metadata.Year == 0 {
				metadata.Year = year
			}
		}
	}

	// Poster URL
	if show.PosterPath != "" {
		metadata.TVMetadata.PosterURL = fmt.Sprintf("https://image.tmdb.org/t/p/w500%s", show.PosterPath)
	}
}

// applyTVDetails applies detailed TV show data to metadata
func (e *Enricher) applyTVDetails(metadata *types.Metadata, details *TVDetails) {
	if metadata.TVMetadata.ShowTitle == "" {
		metadata.TVMetadata.ShowTitle = details.Name
		metadata.Title = details.Name
	}

	metadata.TVMetadata.Plot = details.Overview
	metadata.TVMetadata.Rating = details.VoteAverage
	metadata.TVMetadata.TMDBID = details.ID

	// Extract year from first air date
	if details.FirstAirDate != "" {
		parts := strings.Split(details.FirstAirDate, "-")
		if len(parts) > 0 {
			var year int
			fmt.Sscanf(parts[0], "%d", &year)
			if year > 0 && metadata.Year == 0 {
				metadata.Year = year
			}
		}
	}

	// Genres
	if len(details.Genres) > 0 {
		metadata.TVMetadata.Genres = make([]string, len(details.Genres))
		for i, genre := range details.Genres {
			metadata.TVMetadata.Genres[i] = genre.Name
		}
	}

	// Poster URL
	if details.PosterPath != "" {
		metadata.TVMetadata.PosterURL = fmt.Sprintf("https://image.tmdb.org/t/p/w500%s", details.PosterPath)
	}

	// Backdrop URL
	if details.BackdropPath != "" {
		metadata.TVMetadata.BackdropURL = fmt.Sprintf("https://image.tmdb.org/t/p/w1280%s", details.BackdropPath)
	}

	metadata.TVMetadata.Tagline = details.Tagline
}
