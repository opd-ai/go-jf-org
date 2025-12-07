package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/opd-ai/go-jf-org/pkg/types"
	"github.com/rs/zerolog/log"
)

// Scanner handles file system scanning
type Scanner struct {
	// Extensions maps media types to their file extensions
	videoExtensions []string
	audioExtensions []string
	bookExtensions  []string
	minFileSize     int64
}

// NewScanner creates a new Scanner with the given configuration
func NewScanner(videoExts, audioExts, bookExts []string, minSize int64) *Scanner {
	return &Scanner{
		videoExtensions: normalizeExtensions(videoExts),
		audioExtensions: normalizeExtensions(audioExts),
		bookExtensions:  normalizeExtensions(bookExts),
		minFileSize:     minSize,
	}
}

// ScanResult contains the results of a scan operation
type ScanResult struct {
	Files  []string
	Errors []error
}

// Scan walks the directory tree and returns all media files
func (s *Scanner) Scan(rootPath string) (*ScanResult, error) {
	// Verify the path exists
	info, err := os.Stat(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to access path: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", rootPath)
	}

	result := &ScanResult{
		Files:  make([]string, 0),
		Errors: make([]error, 0),
	}

	log.Info().Str("path", rootPath).Msg("Starting directory scan")

	// Walk the directory tree
	err = filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Warn().Err(err).Str("path", path).Msg("Error accessing path")
			result.Errors = append(result.Errors, fmt.Errorf("error accessing %s: %w", path, err))
			return nil // Continue walking
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Check if file matches our criteria
		if s.isMediaFile(path) {
			// Check file size
			fileInfo, err := d.Info()
			if err != nil {
				log.Warn().Err(err).Str("path", path).Msg("Failed to get file info")
				return nil
			}

			if fileInfo.Size() < s.minFileSize {
				log.Debug().Str("path", path).Int64("size", fileInfo.Size()).Msg("File too small, skipping")
				return nil
			}

			result.Files = append(result.Files, path)
			log.Debug().Str("path", path).Msg("Found media file")
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	log.Info().Int("count", len(result.Files)).Int("errors", len(result.Errors)).Msg("Scan complete")

	return result, nil
}

// isMediaFile checks if a file is a media file based on its extension
func (s *Scanner) isMediaFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))

	return contains(s.videoExtensions, ext) ||
		contains(s.audioExtensions, ext) ||
		contains(s.bookExtensions, ext)
}

// GetMediaType determines the media type based on file extension
func (s *Scanner) GetMediaType(path string) types.MediaType {
	ext := strings.ToLower(filepath.Ext(path))

	if contains(s.videoExtensions, ext) {
		// We'll need more sophisticated detection to distinguish movies from TV
		// For now, return unknown for video files
		return types.MediaTypeUnknown
	}

	if contains(s.audioExtensions, ext) {
		return types.MediaTypeMusic
	}

	if contains(s.bookExtensions, ext) {
		return types.MediaTypeBook
	}

	return types.MediaTypeUnknown
}

// normalizeExtensions ensures all extensions start with a dot and are lowercase
func normalizeExtensions(exts []string) []string {
	normalized := make([]string, len(exts))
	for i, ext := range exts {
		ext = strings.ToLower(ext)
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		normalized[i] = ext
	}
	return normalized
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
