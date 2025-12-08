package scanner

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/opd-ai/go-jf-org/internal/detector"
	"github.com/opd-ai/go-jf-org/internal/metadata"
	"github.com/opd-ai/go-jf-org/pkg/types"
	"github.com/rs/zerolog/log"
)

// Scanner handles file system scanning
type Scanner struct {
	// File extension lists for different media types
	videoExtensions []string
	audioExtensions []string
	bookExtensions  []string
	minFileSize     int64
	// Detector for determining media type
	detector detector.Detector
	// Parser for extracting metadata
	parser metadata.Parser
	// Number of workers for concurrent scanning (0 = auto-detect)
	numWorkers int
}

// NewScanner creates a new Scanner with the given configuration
func NewScanner(videoExts, audioExts, bookExts []string, minSize int64) *Scanner {
	return &Scanner{
		videoExtensions: normalizeExtensions(videoExts),
		audioExtensions: normalizeExtensions(audioExts),
		bookExtensions:  normalizeExtensions(bookExts),
		minFileSize:     minSize,
		detector:        detector.New(),
		parser:          metadata.NewParser(),
		numWorkers:      0, // Auto-detect
	}
}

// SetNumWorkers sets the number of concurrent workers (0 = auto-detect based on CPU count)
func (s *Scanner) SetNumWorkers(n int) {
	s.numWorkers = n
}

// ScanResult contains the results of a scan operation
type ScanResult struct {
	// Files is a list of absolute paths to media files that match the scan criteria
	Files []string
	// Errors is a collection of non-fatal errors encountered during the scan
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
				result.Errors = append(result.Errors, fmt.Errorf("failed to get file info for %s: %w", path, err))
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

// ScanConcurrent walks the directory tree concurrently and returns all media files
// This method uses worker pools for better performance on large directories
func (s *Scanner) ScanConcurrent(ctx context.Context, rootPath string) (*ScanResult, error) {
	// Verify the path exists
	info, err := os.Stat(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to access path: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", rootPath)
	}

	// Determine number of workers
	numWorkers := s.numWorkers
	if numWorkers <= 0 {
		numWorkers = runtime.NumCPU()
	}

	log.Info().Str("path", rootPath).Int("workers", numWorkers).Msg("Starting concurrent directory scan")

	// Combine all extensions
	allExtensions := make([]string, 0, len(s.videoExtensions)+len(s.audioExtensions)+len(s.bookExtensions))
	allExtensions = append(allExtensions, s.videoExtensions...)
	allExtensions = append(allExtensions, s.audioExtensions...)
	allExtensions = append(allExtensions, s.bookExtensions...)

	// Create worker pool and scan
	pool := NewWorkerPool(numWorkers, s.detector)
	paths, sizes, err := pool.ScanConcurrent(ctx, rootPath, allExtensions)
	if err != nil {
		return nil, fmt.Errorf("concurrent scan failed: %w", err)
	}

	// Filter by file size
	result := &ScanResult{
		Files:  make([]string, 0, len(paths)),
		Errors: make([]error, 0),
	}

	for i, path := range paths {
		if sizes[i] >= s.minFileSize {
			result.Files = append(result.Files, path)
		} else {
			log.Debug().Str("path", path).Int64("size", sizes[i]).Msg("File too small, skipping")
		}
	}

	log.Info().Int("count", len(result.Files)).Int("workers", numWorkers).Msg("Concurrent scan complete")

	return result, nil
}

// isMediaFile checks if a file is a media file based on its extension
func (s *Scanner) isMediaFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))

	return contains(s.videoExtensions, ext) ||
		contains(s.audioExtensions, ext) ||
		contains(s.bookExtensions, ext)
}

// GetMediaType determines the media type based on file extension and filename patterns
func (s *Scanner) GetMediaType(path string) types.MediaType {
	return s.detector.Detect(path)
}

// GetMetadata extracts metadata from a file
func (s *Scanner) GetMetadata(path string) (*types.Metadata, error) {
	mediaType := s.GetMediaType(path)
	return s.parser.Parse(filepath.Base(path), mediaType)
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
