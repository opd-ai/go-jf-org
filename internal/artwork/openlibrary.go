package artwork

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

const (
	// OpenLibraryCoversBaseURL is the base URL for OpenLibrary Covers API
	OpenLibraryCoversBaseURL = "https://covers.openlibrary.org/b"
)

// OpenLibraryDownloader handles artwork downloads from OpenLibrary
type OpenLibraryDownloader struct {
	*BaseDownloader
	imageSize ImageSize
}

// NewOpenLibraryDownloader creates a new OpenLibrary cover downloader
func NewOpenLibraryDownloader(config Config, size ImageSize) *OpenLibraryDownloader {
	if size == "" {
		size = SizeLarge // Default to large for books
	}

	return &OpenLibraryDownloader{
		BaseDownloader: NewBaseDownloader(config),
		imageSize:      size,
	}
}

// DownloadBookCoverByISBN downloads book cover by ISBN
func (d *OpenLibraryDownloader) DownloadBookCoverByISBN(ctx context.Context, isbn, destDir string) error {
	if isbn == "" {
		log.Debug().Msg("No ISBN available, skipping book cover download")
		return nil
	}

	sizeStr := d.getSizeString()
	imageURL := fmt.Sprintf("%s/isbn/%s-%s.jpg?default=false", OpenLibraryCoversBaseURL, isbn, sizeStr)
	destPath := filepath.Join(destDir, "cover.jpg")

	log.Info().
		Str("isbn", isbn).
		Str("size", sizeStr).
		Str("dest", destPath).
		Msg("Downloading book cover")

	// Try to download with ?default=false to get 404 if no cover exists
	err := d.downloadWithFallback(ctx, imageURL, destPath)
	if err != nil {
		log.Warn().
			Err(err).
			Str("isbn", isbn).
			Msg("Failed to download book cover")
		return nil // Don't fail if cover is missing
	}

	return nil
}

// DownloadBookCoverByOLID downloads book cover by OpenLibrary ID
func (d *OpenLibraryDownloader) DownloadBookCoverByOLID(ctx context.Context, olid, destDir string) error {
	if olid == "" {
		log.Debug().Msg("No OpenLibrary ID available, skipping book cover download")
		return nil
	}

	sizeStr := d.getSizeString()
	imageURL := fmt.Sprintf("%s/olid/%s-%s.jpg?default=false", OpenLibraryCoversBaseURL, olid, sizeStr)
	destPath := filepath.Join(destDir, "cover.jpg")

	log.Info().
		Str("olid", olid).
		Str("size", sizeStr).
		Str("dest", destPath).
		Msg("Downloading book cover")

	err := d.downloadWithFallback(ctx, imageURL, destPath)
	if err != nil {
		log.Warn().
			Err(err).
			Str("olid", olid).
			Msg("Failed to download book cover")
		return nil // Don't fail if cover is missing
	}

	return nil
}

// downloadWithFallback tries to download, handling 404 gracefully
func (d *OpenLibraryDownloader) downloadWithFallback(ctx context.Context, imageURL, destPath string) error {
	// Check if image exists first
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, imageURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create HEAD request: %w", err)
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("HEAD request failed: %w", err)
	}
	defer resp.Body.Close()

	// If 404, cover doesn't exist
	if resp.StatusCode == http.StatusNotFound {
		log.Debug().Str("url", imageURL).Msg("No cover available")
		return nil
	}

	// If not OK, return error
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Download the image
	return d.DownloadImage(ctx, imageURL, destPath)
}

// getSizeString returns the appropriate size string for OpenLibrary API
func (d *OpenLibraryDownloader) getSizeString() string {
	switch d.imageSize {
	case SizeSmall:
		return "S"
	case SizeMedium:
		return "M"
	case SizeLarge, SizeOriginal:
		return "L"
	default:
		return "L"
	}
}
