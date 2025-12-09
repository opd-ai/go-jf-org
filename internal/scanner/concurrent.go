package scanner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/opd-ai/go-jf-org/internal/detector"
	"github.com/opd-ai/go-jf-org/pkg/types"
	"github.com/rs/zerolog/log"
)

// WorkerPool manages concurrent file scanning operations
type WorkerPool struct {
	numWorkers int
	detector   detector.Detector
}

// NewWorkerPool creates a new worker pool for concurrent scanning
func NewWorkerPool(numWorkers int, det detector.Detector) *WorkerPool {
	if numWorkers < 1 {
		numWorkers = 1
	}
	return &WorkerPool{
		numWorkers: numWorkers,
		detector:   det,
	}
}

// FileScanResult represents a single file scan result
type FileScanResult struct {
	Path      string
	Size      int64
	MediaType types.MediaType
	Error     error
}

// ScanConcurrent scans a directory concurrently using worker pools
func (wp *WorkerPool) ScanConcurrent(ctx context.Context, rootPath string, extensions []string) ([]string, []int64, error) {
	// Channel for discovered file paths
	pathChan := make(chan string, 100)

	// Channel for scan results
	resultChan := make(chan FileScanResult, 100)

	// WaitGroup for workers
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < wp.numWorkers; i++ {
		wg.Add(1)
		go wp.worker(ctx, &wg, pathChan, resultChan, extensions)
	}

	// Start directory walker in a separate goroutine
	go func() {
		defer close(pathChan)
		wp.walkDirectory(ctx, rootPath, pathChan)
	}()

	// Start result collector
	paths := make([]string, 0)
	sizes := make([]int64, 0)
	var resultWg sync.WaitGroup
	resultWg.Add(1)

	go func() {
		defer resultWg.Done()
		for result := range resultChan {
			if result.Error != nil {
				log.Debug().Err(result.Error).Str("path", result.Path).Msg("Error processing file")
				continue
			}
			if result.MediaType != types.MediaTypeUnknown {
				paths = append(paths, result.Path)
				sizes = append(sizes, result.Size)
			}
		}
	}()

	// Wait for all workers to finish
	wg.Wait()
	close(resultChan)

	// Wait for result collector (ensures all appends are complete before returning)
	resultWg.Wait()

	return paths, sizes, nil // Safe: all appends completed
}

// worker processes files from the path channel
func (wp *WorkerPool) worker(ctx context.Context, wg *sync.WaitGroup, pathChan <-chan string, resultChan chan<- FileScanResult, extensions []string) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case path, ok := <-pathChan:
			if !ok {
				return
			}

			// Process the file
			result := wp.processFile(path, extensions)

			// Send result (skip if nil)
			if result != nil {
				select {
				case resultChan <- *result:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

// processFile processes a single file and returns a scan result
func (wp *WorkerPool) processFile(path string, extensions []string) *FileScanResult {
	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		return &FileScanResult{Path: path, Error: err}
	}

	// Skip directories
	if info.IsDir() {
		return nil
	}

	// Check extension
	ext := filepath.Ext(path)
	if !containsExtension(ext, extensions) {
		return nil
	}

	// Detect media type
	mediaType := wp.detector.Detect(path)
	if mediaType == types.MediaTypeUnknown {
		log.Debug().Str("path", path).Msg("Unknown media type, skipping")
		return nil
	}

	return &FileScanResult{
		Path:      path,
		Size:      info.Size(),
		MediaType: mediaType,
		Error:     nil,
	}
}

// walkDirectory walks the directory tree and sends paths to the channel
func (wp *WorkerPool) walkDirectory(ctx context.Context, rootPath string, pathChan chan<- string) {
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Debug().Err(err).Str("path", path).Msg("Error accessing path")
			return nil // Continue walking
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Skip hidden files and directories
		name := info.Name()
		if len(name) > 0 && name[0] == '.' && path != rootPath {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Send file paths to channel
		if !info.IsDir() {
			select {
			case pathChan <- path:
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		return nil
	})

	if err != nil && err != context.Canceled {
		log.Debug().Err(err).Str("root", rootPath).Msg("Directory walk error")
	}
}

// containsExtension checks if an extension is in the list (case-insensitive)
func containsExtension(ext string, extensions []string) bool {
	extLower := strings.ToLower(ext)
	for _, e := range extensions {
		if extLower == e {
			return true
		}
	}
	return false
}
