package artwork

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	// DefaultTimeout for HTTP requests
	DefaultTimeout = 30 * time.Second

	// DefaultMaxRetries for failed downloads
	DefaultMaxRetries = 3

	// DefaultRetryDelay base delay for exponential backoff
	DefaultRetryDelay = 1 * time.Second
)

// Config holds configuration for artwork downloaders
type Config struct {
	Timeout    time.Duration
	MaxRetries int
	RetryDelay time.Duration
	Force      bool // Force re-download even if file exists
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		Timeout:    DefaultTimeout,
		MaxRetries: DefaultMaxRetries,
		RetryDelay: DefaultRetryDelay,
		Force:      false,
	}
}

// BaseDownloader provides common HTTP download functionality
type BaseDownloader struct {
	httpClient *http.Client
	config     Config
}

// NewBaseDownloader creates a new base downloader
func NewBaseDownloader(config Config) *BaseDownloader {
	if config.Timeout == 0 {
		config.Timeout = DefaultTimeout
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = DefaultMaxRetries
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = DefaultRetryDelay
	}

	return &BaseDownloader{
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		config: config,
	}
}

// DownloadImage downloads an image from the given URL to the destination path
// It handles retries, checks if file exists, and validates the downloaded image
func (d *BaseDownloader) DownloadImage(ctx context.Context, imageURL, destPath string) error {
	// Check if file already exists and skip if not forcing
	if !d.config.Force {
		if info, err := os.Stat(destPath); err == nil && info.Size() > 0 {
			log.Debug().
				Str("path", destPath).
				Int64("size", info.Size()).
				Msg("Artwork already exists, skipping download")
			return nil
		}
	}

	// Ensure destination directory exists
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Retry logic with exponential backoff
	var lastErr error
	for attempt := 0; attempt < d.config.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := d.config.RetryDelay * time.Duration(1<<uint(attempt-1))
			log.Debug().
				Str("url", imageURL).
				Int("attempt", attempt+1).
				Dur("delay", delay).
				Msg("Retrying artwork download after delay")

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := d.downloadOnce(ctx, imageURL, destPath)
		if err == nil {
			log.Info().
				Str("url", imageURL).
				Str("dest", destPath).
				Msg("Artwork downloaded successfully")
			return nil
		}

		lastErr = err
		log.Warn().
			Err(err).
			Str("url", imageURL).
			Int("attempt", attempt+1).
			Msg("Artwork download failed")
	}

	return fmt.Errorf("failed after %d attempts: %w", d.config.MaxRetries, lastErr)
}

// downloadOnce performs a single download attempt
func (d *BaseDownloader) downloadOnce(ctx context.Context, imageURL, destPath string) error {
	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Make HTTP request
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for successful response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp(filepath.Dir(destPath), "artwork-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath) // Clean up on error

	// Copy response to temp file
	written, err := io.Copy(tmpFile, resp.Body)
	if err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write image data: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Validate downloaded file
	if written == 0 {
		return fmt.Errorf("downloaded file is empty")
	}

	// Move temp file to final destination
	if err := os.Rename(tmpPath, destPath); err != nil {
		return fmt.Errorf("failed to move file to destination: %w", err)
	}

	log.Debug().
		Str("path", destPath).
		Int64("size", written).
		Msg("Image file written successfully")

	return nil
}

// FileExists checks if a file exists and has non-zero size
func FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Size() > 0
}
