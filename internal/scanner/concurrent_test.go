package scanner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/opd-ai/go-jf-org/internal/detector"
)

func TestWorkerPool_ScanConcurrent(t *testing.T) {
	tests := []struct {
		name       string
		numWorkers int
		files      map[string]string // filename -> content
		extensions []string
		wantCount  int
	}{
		{
			name:       "single worker",
			numWorkers: 1,
			files: map[string]string{
				"movie1.mkv": "",
				"movie2.mp4": "",
				"show1.mkv":  "",
			},
			extensions: []string{".mkv", ".mp4"},
			wantCount:  3,
		},
		{
			name:       "multiple workers",
			numWorkers: 4,
			files: map[string]string{
				"movie1.mkv": "",
				"movie2.mp4": "",
				"movie3.avi": "",
				"movie4.mkv": "",
				"show1.mkv":  "",
				"show2.mp4":  "",
			},
			extensions: []string{".mkv", ".mp4", ".avi"},
			wantCount:  6,
		},
		{
			name:       "filter by extension",
			numWorkers: 2,
			files: map[string]string{
				"movie1.mkv": "",
				"movie2.txt": "",
				"movie3.mp4": "",
				"readme.md":  "",
			},
			extensions: []string{".mkv", ".mp4"},
			wantCount:  2,
		},
		{
			name:       "zero workers defaults to 1",
			numWorkers: 0,
			files: map[string]string{
				"movie1.mkv": "",
			},
			extensions: []string{".mkv"},
			wantCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tempDir := t.TempDir()

			// Create test files
			for filename := range tt.files {
				path := filepath.Join(tempDir, filename)
				if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			// Create worker pool
			det := detector.New()
			pool := NewWorkerPool(tt.numWorkers, det)

			// Scan concurrently
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			paths, sizes, err := pool.ScanConcurrent(ctx, tempDir, tt.extensions)
			if err != nil {
				t.Fatalf("ScanConcurrent() error = %v", err)
			}

			if len(paths) != tt.wantCount {
				t.Errorf("ScanConcurrent() got %d files, want %d", len(paths), tt.wantCount)
			}

			if len(sizes) != len(paths) {
				t.Errorf("Mismatch between paths (%d) and sizes (%d)", len(paths), len(sizes))
			}

			// Verify all results have required fields
			for i, path := range paths {
				if path == "" {
					t.Error("File has empty path")
				}
				if sizes[i] == 0 {
					t.Error("File has zero size")
				}
			}
		})
	}
}

func TestWorkerPool_ContextCancellation(t *testing.T) {
	// Create temp directory with many files
	tempDir := t.TempDir()
	for i := 0; i < 100; i++ {
		filename := fmt.Sprintf("movie%03d.mkv", i)
		path := filepath.Join(tempDir, filename)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Create worker pool
	det := detector.New()
	pool := NewWorkerPool(4, det)

	// Create context that cancels immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Scan should handle cancellation gracefully
	_, _, err := pool.ScanConcurrent(ctx, tempDir, []string{".mkv"})

	// We expect either no error (completed before cancel) or context canceled error
	if err != nil && err != context.Canceled {
		t.Errorf("Expected nil or context.Canceled, got %v", err)
	}
}

func TestWorkerPool_NonExistentDirectory(t *testing.T) {
	det := detector.New()
	pool := NewWorkerPool(2, det)

	ctx := context.Background()
	paths, _, err := pool.ScanConcurrent(ctx, "/non/existent/path", []string{".mkv"})

	// Should return empty results, not crash
	if err != nil {
		t.Errorf("Expected no error for non-existent directory, got %v", err)
	}
	if len(paths) != 0 {
		t.Errorf("Expected 0 results, got %d", len(paths))
	}
}

func TestWorkerPool_HiddenFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create hidden file
	hiddenPath := filepath.Join(tempDir, ".hidden.mkv")
	if err := os.WriteFile(hiddenPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create hidden file: %v", err)
	}

	// Create normal file
	normalPath := filepath.Join(tempDir, "visible.mkv")
	if err := os.WriteFile(normalPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create normal file: %v", err)
	}

	det := detector.New()
	pool := NewWorkerPool(2, det)

	ctx := context.Background()
	paths, _, err := pool.ScanConcurrent(ctx, tempDir, []string{".mkv"})
	if err != nil {
		t.Fatalf("ScanConcurrent() error = %v", err)
	}

	// Should only find the visible file
	if len(paths) != 1 {
		t.Errorf("Expected 1 file (hidden should be skipped), got %d", len(paths))
	}

	if len(paths) > 0 && filepath.Base(paths[0]) != "visible.mkv" {
		t.Errorf("Expected visible.mkv, got %s", filepath.Base(paths[0]))
	}
}

func BenchmarkWorkerPool_Sequential(b *testing.B) {
	benchmarkWorkerPool(b, 1)
}

func BenchmarkWorkerPool_Parallel2(b *testing.B) {
	benchmarkWorkerPool(b, 2)
}

func BenchmarkWorkerPool_Parallel4(b *testing.B) {
	benchmarkWorkerPool(b, 4)
}

func BenchmarkWorkerPool_Parallel8(b *testing.B) {
	benchmarkWorkerPool(b, 8)
}

func benchmarkWorkerPool(b *testing.B, numWorkers int) {
	// Create temp directory with test files
	tempDir := b.TempDir()
	for i := 0; i < 100; i++ {
		filename := fmt.Sprintf("movie%03d.mkv", i)
		path := filepath.Join(tempDir, filename)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			b.Fatalf("Failed to create test file: %v", err)
		}
	}

	det := detector.New()
	pool := NewWorkerPool(numWorkers, det)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = pool.ScanConcurrent(ctx, tempDir, []string{".mkv"})
	}
}
