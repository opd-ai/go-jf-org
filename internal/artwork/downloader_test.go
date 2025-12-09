package artwork

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Timeout != DefaultTimeout {
		t.Errorf("Expected timeout %v, got %v", DefaultTimeout, config.Timeout)
	}
	if config.MaxRetries != DefaultMaxRetries {
		t.Errorf("Expected max retries %d, got %d", DefaultMaxRetries, config.MaxRetries)
	}
	if config.RetryDelay != DefaultRetryDelay {
		t.Errorf("Expected retry delay %v, got %v", DefaultRetryDelay, config.RetryDelay)
	}
	if config.Force {
		t.Error("Expected Force to be false by default")
	}
}

func TestNewBaseDownloader(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name:   "Default config",
			config: DefaultConfig(),
		},
		{
			name:   "Zero values get defaults",
			config: Config{},
		},
		{
			name: "Custom values preserved",
			config: Config{
				Timeout:    5 * time.Second,
				MaxRetries: 5,
				RetryDelay: 2 * time.Second,
				Force:      true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			downloader := NewBaseDownloader(tt.config)
			if downloader == nil {
				t.Fatal("Expected non-nil downloader")
			}
			if downloader.httpClient == nil {
				t.Error("Expected non-nil HTTP client")
			}
		})
	}
}

func TestDownloadImage(t *testing.T) {
	// Create test server that returns a sample image
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("fake image data"))
	}))
	defer server.Close()

	tempDir := t.TempDir()

	tests := []struct {
		name        string
		setupFile   bool // Pre-create file to test skip logic
		force       bool
		expectError bool
	}{
		{
			name:        "Successful download",
			setupFile:   false,
			force:       false,
			expectError: false,
		},
		{
			name:        "Skip existing file",
			setupFile:   true,
			force:       false,
			expectError: false,
		},
		{
			name:        "Force re-download existing file",
			setupFile:   true,
			force:       true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			destPath := filepath.Join(tempDir, tt.name+".jpg")

			if tt.setupFile {
				if err := os.WriteFile(destPath, []byte("existing data"), 0644); err != nil {
					t.Fatalf("Failed to setup test file: %v", err)
				}
			}

			config := DefaultConfig()
			config.Force = tt.force
			downloader := NewBaseDownloader(config)

			ctx := context.Background()
			err := downloader.DownloadImage(ctx, server.URL, destPath)

			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify file exists after download
			if !tt.expectError {
				if !FileExists(destPath) {
					t.Error("Expected file to exist after download")
				}
			}
		})
	}
}

func TestDownloadImageWithRetries(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success after retry"))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "retry-test.jpg")

	config := DefaultConfig()
	config.RetryDelay = 10 * time.Millisecond // Speed up test
	downloader := NewBaseDownloader(config)

	ctx := context.Background()
	err := downloader.DownloadImage(ctx, server.URL, destPath)

	if err != nil {
		t.Errorf("Expected success after retries, got error: %v", err)
	}

	if attempts < 2 {
		t.Errorf("Expected at least 2 attempts, got %d", attempts)
	}

	if !FileExists(destPath) {
		t.Error("Expected file to exist after successful retry")
	}
}

func TestDownloadImageFailure(t *testing.T) {
	tests := []struct {
		name        string
		setupServer func() *httptest.Server
	}{
		{
			name: "404 Not Found",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}))
			},
		},
		{
			name: "500 Internal Server Error",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
			},
		},
		{
			name: "Empty response body",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					// No body written
				}))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.setupServer()
			defer server.Close()

			tempDir := t.TempDir()
			destPath := filepath.Join(tempDir, "failed.jpg")

			config := DefaultConfig()
			config.RetryDelay = 1 * time.Millisecond // Speed up test
			downloader := NewBaseDownloader(config)

			ctx := context.Background()
			err := downloader.DownloadImage(ctx, server.URL, destPath)

			if err == nil {
				t.Error("Expected error but got nil")
			}
		})
	}
}

func TestDownloadImageContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data"))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "cancelled.jpg")

	config := DefaultConfig()
	downloader := NewBaseDownloader(config)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := downloader.DownloadImage(ctx, server.URL, destPath)

	if err == nil {
		t.Error("Expected context cancellation error")
	}
}

func TestFileExists(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		setup    func() string
		expected bool
	}{
		{
			name: "File exists with content",
			setup: func() string {
				path := filepath.Join(tempDir, "exists.txt")
				os.WriteFile(path, []byte("content"), 0644)
				return path
			},
			expected: true,
		},
		{
			name: "File exists but empty",
			setup: func() string {
				path := filepath.Join(tempDir, "empty.txt")
				os.WriteFile(path, []byte{}, 0644)
				return path
			},
			expected: false,
		},
		{
			name: "File does not exist",
			setup: func() string {
				return filepath.Join(tempDir, "nonexistent.txt")
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			result := FileExists(path)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
