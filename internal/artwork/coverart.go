package artwork

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	// CoverArtArchiveBaseURL is the base URL for Cover Art Archive API
	CoverArtArchiveBaseURL = "https://coverartarchive.org"

	// CoverArtDefaultTimeout for API requests
	CoverArtDefaultTimeout = 15 * time.Second
)

// CoverArtDownloader handles artwork downloads from Cover Art Archive (MusicBrainz)
type CoverArtDownloader struct {
	*BaseDownloader
	imageSize ImageSize
}

// CoverArtResponse represents the Cover Art Archive API response
type CoverArtResponse struct {
	Images []CoverArtImage `json:"images"`
}

// CoverArtImage represents a single image from the Cover Art Archive
type CoverArtImage struct {
	Types      []string          `json:"types"`
	Front      bool              `json:"front"`
	Back       bool              `json:"back"`
	Image      string            `json:"image"`
	Thumbnails CoverArtThumbnails `json:"thumbnails"`
}

// CoverArtThumbnails contains different sizes of thumbnails
type CoverArtThumbnails struct {
	Small  string `json:"250"`  // 250px
	Medium string `json:"500"`  // 500px
	Large  string `json:"1200"` // 1200px
}

// NewCoverArtDownloader creates a new Cover Art Archive downloader
func NewCoverArtDownloader(config Config, size ImageSize) *CoverArtDownloader {
	if size == "" {
		size = SizeMedium // Default to medium
	}

	// Extend timeout for Cover Art Archive (may be slower)
	if config.Timeout < CoverArtDefaultTimeout {
		config.Timeout = CoverArtDefaultTimeout
	}

	return &CoverArtDownloader{
		BaseDownloader: NewBaseDownloader(config),
		imageSize:      size,
	}
}

// DownloadAlbumCover downloads album cover art for the given MusicBrainz release ID
func (d *CoverArtDownloader) DownloadAlbumCover(ctx context.Context, releaseID, destDir string) error {
	if releaseID == "" {
		log.Debug().Msg("No MusicBrainz release ID available, skipping cover download")
		return nil
	}

	// Construct API URL to get artwork metadata
	apiURL := fmt.Sprintf("%s/release/%s", CoverArtArchiveBaseURL, releaseID)

	// Fetch artwork metadata
	imageURL, err := d.getImageURL(ctx, apiURL)
	if err != nil {
		return fmt.Errorf("failed to get cover art URL: %w", err)
	}

	if imageURL == "" {
		log.Debug().
			Str("releaseID", releaseID).
			Msg("No cover art available for this release")
		return nil
	}

	destPath := filepath.Join(destDir, "cover.jpg")

	log.Info().
		Str("releaseID", releaseID).
		Str("dest", destPath).
		Msg("Downloading album cover")

	return d.DownloadImage(ctx, imageURL, destPath)
}

// getImageURL fetches the Cover Art Archive metadata and extracts the appropriate image URL
func (d *CoverArtDownloader) getImageURL(ctx context.Context, apiURL string) (string, error) {
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	// Make HTTP request
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle 404 - no artwork available
	if resp.StatusCode == http.StatusNotFound {
		log.Debug().Str("url", apiURL).Msg("No cover art found for this release")
		return "", nil
	}

	// Check for successful response
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse JSON response
	var artResp CoverArtResponse
	if err := json.NewDecoder(resp.Body).Decode(&artResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Find the front cover
	for _, img := range artResp.Images {
		if img.Front || contains(img.Types, "Front") {
			return d.selectImageURL(img), nil
		}
	}

	// No front cover found, use first available image
	if len(artResp.Images) > 0 {
		return d.selectImageURL(artResp.Images[0]), nil
	}

	log.Debug().Msg("No images found in Cover Art Archive response")
	return "", nil
}

// selectImageURL selects the appropriate image URL based on size preference
func (d *CoverArtDownloader) selectImageURL(img CoverArtImage) string {
	switch d.imageSize {
	case SizeSmall:
		if img.Thumbnails.Small != "" {
			return img.Thumbnails.Small
		}
	case SizeMedium:
		if img.Thumbnails.Medium != "" {
			return img.Thumbnails.Medium
		}
	case SizeLarge:
		if img.Thumbnails.Large != "" {
			return img.Thumbnails.Large
		}
	case SizeOriginal:
		return img.Image
	}

	// Fallback to original if preferred size not available
	return img.Image
}

// contains checks if a string slice contains a value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
