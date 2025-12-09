package artwork

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewOpenLibraryDownloader(t *testing.T) {
	tests := []struct {
		name     string
		size     ImageSize
		expected ImageSize
	}{
		{
			name:     "Default size",
			size:     "",
			expected: SizeLarge,
		},
		{
			name:     "Small size",
			size:     SizeSmall,
			expected: SizeSmall,
		},
		{
			name:     "Medium size",
			size:     SizeMedium,
			expected: SizeMedium,
		},
		{
			name:     "Large size",
			size:     SizeLarge,
			expected: SizeLarge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			downloader := NewOpenLibraryDownloader(config, tt.size)

			if downloader == nil {
				t.Fatal("Expected non-nil downloader")
			}
			if downloader.imageSize != tt.expected {
				t.Errorf("Expected size %s, got %s", tt.expected, downloader.imageSize)
			}
		})
	}
}

func TestOpenLibraryDownloader_DownloadBookCoverByISBN(t *testing.T) {
	tests := []struct {
		name        string
		isbn        string
		expectError bool
	}{
		{
			name:        "Empty ISBN",
			isbn:        "",
			expectError: false,
		},
		{
			name:        "Valid ISBN",
			isbn:        "0385472579",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			downloader := NewOpenLibraryDownloader(config, SizeLarge)
			tempDir := t.TempDir()

			err := downloader.DownloadBookCoverByISBN(context.Background(), tt.isbn, tempDir)
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestOpenLibraryDownloader_DownloadBookCoverByOLID(t *testing.T) {
	tests := []struct {
		name        string
		olid        string
		expectError bool
	}{
		{
			name:        "Empty OLID",
			olid:        "",
			expectError: false,
		},
		{
			name:        "Valid OLID",
			olid:        "OL7440033M",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			downloader := NewOpenLibraryDownloader(config, SizeLarge)
			tempDir := t.TempDir()

			err := downloader.DownloadBookCoverByOLID(context.Background(), tt.olid, tempDir)
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestOpenLibraryDownloader_downloadWithFallback(t *testing.T) {
	tests := []struct {
		name        string
		setupServer func() *httptest.Server
		expectError bool
	}{
		{
			name: "404 Not Found",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}))
			},
			expectError: false, // Should not error on 404
		},
		{
			name: "500 Internal Server Error",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
			},
			expectError: true,
		},
		{
			name: "Successful HEAD then download",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method == http.MethodHead {
						w.WriteHeader(http.StatusOK)
					} else {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte("fake image data"))
					}
				}))
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.setupServer()
			defer server.Close()

			config := DefaultConfig()
			downloader := NewOpenLibraryDownloader(config, SizeLarge)
			tempDir := t.TempDir()
			destPath := tempDir + "/cover.jpg"

			err := downloader.downloadWithFallback(context.Background(), server.URL, destPath)
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestOpenLibraryDownloader_getSizeString(t *testing.T) {
	tests := []struct {
		name     string
		size     ImageSize
		expected string
	}{
		{
			name:     "Small size",
			size:     SizeSmall,
			expected: "S",
		},
		{
			name:     "Medium size",
			size:     SizeMedium,
			expected: "M",
		},
		{
			name:     "Large size",
			size:     SizeLarge,
			expected: "L",
		},
		{
			name:     "Original size",
			size:     SizeOriginal,
			expected: "L",
		},
		{
			name:     "Default/empty size",
			size:     "",
			expected: "L",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			downloader := NewOpenLibraryDownloader(config, tt.size)
			
			result := downloader.getSizeString()
			if result != tt.expected {
				t.Errorf("Expected size string %s, got %s", tt.expected, result)
			}
		})
	}
}
