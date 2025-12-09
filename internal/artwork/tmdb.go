package artwork

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

const (
	// TMDBImageBaseURL is the base URL for TMDB images
	TMDBImageBaseURL = "https://image.tmdb.org/t/p/"
)

// ImageSize represents artwork image size preference
type ImageSize string

const (
	// Image sizes for different use cases
	SizeSmall    ImageSize = "small"
	SizeMedium   ImageSize = "medium"
	SizeLarge    ImageSize = "large"
	SizeOriginal ImageSize = "original"
)

// TMDBDownloader handles artwork downloads from TMDB
type TMDBDownloader struct {
	*BaseDownloader
	imageSize ImageSize
}

// NewTMDBDownloader creates a new TMDB artwork downloader
func NewTMDBDownloader(config Config, size ImageSize) *TMDBDownloader {
	if size == "" {
		size = SizeMedium // Default to medium
	}

	return &TMDBDownloader{
		BaseDownloader: NewBaseDownloader(config),
		imageSize:      size,
	}
}

// DownloadMoviePoster downloads a movie poster to the specified directory
func (d *TMDBDownloader) DownloadMoviePoster(ctx context.Context, posterPath, destDir string) error {
	if posterPath == "" {
		log.Debug().Msg("No poster path available, skipping poster download")
		return nil
	}

	imageURL := d.buildImageURL(posterPath, true)
	destPath := filepath.Join(destDir, "poster.jpg")

	log.Info().
		Str("url", imageURL).
		Str("dest", destPath).
		Msg("Downloading movie poster")

	return d.DownloadImage(ctx, imageURL, destPath)
}

// DownloadMovieBackdrop downloads a movie backdrop to the specified directory
func (d *TMDBDownloader) DownloadMovieBackdrop(ctx context.Context, backdropPath, destDir string) error {
	if backdropPath == "" {
		log.Debug().Msg("No backdrop path available, skipping backdrop download")
		return nil
	}

	imageURL := d.buildImageURL(backdropPath, false)
	destPath := filepath.Join(destDir, "backdrop.jpg")

	log.Info().
		Str("url", imageURL).
		Str("dest", destPath).
		Msg("Downloading movie backdrop")

	return d.DownloadImage(ctx, imageURL, destPath)
}

// DownloadTVPoster downloads a TV show poster to the specified directory
func (d *TMDBDownloader) DownloadTVPoster(ctx context.Context, posterPath, destDir string) error {
	if posterPath == "" {
		log.Debug().Msg("No poster path available, skipping TV poster download")
		return nil
	}

	imageURL := d.buildImageURL(posterPath, true)
	destPath := filepath.Join(destDir, "poster.jpg")

	log.Info().
		Str("url", imageURL).
		Str("dest", destPath).
		Msg("Downloading TV show poster")

	return d.DownloadImage(ctx, imageURL, destPath)
}

// DownloadSeasonPoster downloads a TV season poster to the specified directory
func (d *TMDBDownloader) DownloadSeasonPoster(ctx context.Context, posterPath, seasonDir string) error {
	if posterPath == "" {
		log.Debug().Msg("No season poster path available, skipping season poster download")
		return nil
	}

	imageURL := d.buildImageURL(posterPath, true)
	destPath := filepath.Join(seasonDir, "poster.jpg")

	log.Info().
		Str("url", imageURL).
		Str("dest", destPath).
		Msg("Downloading season poster")

	return d.DownloadImage(ctx, imageURL, destPath)
}

// DownloadMovieArtwork downloads all available artwork for a movie
func (d *TMDBDownloader) DownloadMovieArtwork(ctx context.Context, posterPath, backdropPath, destDir string) error {
	var errors []error

	// Download movie poster
	if posterPath != "" {
		if err := d.DownloadMoviePoster(ctx, posterPath, destDir); err != nil {
			log.Warn().Err(err).Msg("Failed to download movie poster")
			errors = append(errors, err)
		}
	}

	// Download movie backdrop
	if backdropPath != "" {
		if err := d.DownloadMovieBackdrop(ctx, backdropPath, destDir); err != nil {
			log.Warn().Err(err).Msg("Failed to download movie backdrop")
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		log.Warn().
			Int("failed", len(errors)).
			Msg("Some artwork downloads failed")
		// Don't return error - partial success is acceptable
	}

	return nil
}

// DownloadTVArtwork downloads all available artwork for a TV show
func (d *TMDBDownloader) DownloadTVArtwork(ctx context.Context, posterPath, destDir string) error {
	// Download TV show poster
	if posterPath != "" {
		if err := d.DownloadTVPoster(ctx, posterPath, destDir); err != nil {
			log.Warn().Err(err).Msg("Failed to download TV show poster")
			// Don't return error - partial success is acceptable
		}
	}

	return nil
}

// buildImageURL constructs the full TMDB image URL
func (d *TMDBDownloader) buildImageURL(path string, isPoster bool) string {
	sizeStr := d.getSizeString(isPoster)
	return fmt.Sprintf("%s%s%s", TMDBImageBaseURL, sizeStr, path)
}

// getSizeString returns the appropriate size string for TMDB API
func (d *TMDBDownloader) getSizeString(isPoster bool) string {
	if isPoster {
		// Poster sizes: w92, w154, w185, w342, w500, w780, original
		switch d.imageSize {
		case SizeSmall:
			return "w185"
		case SizeMedium:
			return "w500"
		case SizeLarge:
			return "w780"
		case SizeOriginal:
			return "original"
		default:
			return "w500"
		}
	} else {
		// Backdrop sizes: w300, w780, w1280, original
		switch d.imageSize {
		case SizeSmall:
			return "w300"
		case SizeMedium:
			return "w780"
		case SizeLarge:
			return "w1280"
		case SizeOriginal:
			return "original"
		default:
			return "w780"
		}
	}
}
