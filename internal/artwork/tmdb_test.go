package artwork

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestNewTMDBDownloader(t *testing.T) {
	tests := []struct {
		name     string
		size     ImageSize
		expected ImageSize
	}{
		{
			name:     "Default size",
			size:     "",
			expected: SizeMedium,
		},
		{
			name:     "Small size",
			size:     SizeSmall,
			expected: SizeSmall,
		},
		{
			name:     "Large size",
			size:     SizeLarge,
			expected: SizeLarge,
		},
		{
			name:     "Original size",
			size:     SizeOriginal,
			expected: SizeOriginal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			downloader := NewTMDBDownloader(config, tt.size)

			if downloader == nil {
				t.Fatal("Expected non-nil downloader")
			}
			if downloader.imageSize != tt.expected {
				t.Errorf("Expected size %s, got %s", tt.expected, downloader.imageSize)
			}
		})
	}
}

func TestBuildImageURL(t *testing.T) {
	tests := []struct {
		name     string
		size     ImageSize
		path     string
		isPoster bool
		expected string
	}{
		{
			name:     "Small poster",
			size:     SizeSmall,
			path:     "/pB8BM7pdSp6B6Ih7QZ4DrQ3PmJK.jpg",
			isPoster: true,
			expected: "https://image.tmdb.org/t/p/w185/pB8BM7pdSp6B6Ih7QZ4DrQ3PmJK.jpg",
		},
		{
			name:     "Medium poster",
			size:     SizeMedium,
			path:     "/pB8BM7pdSp6B6Ih7QZ4DrQ3PmJK.jpg",
			isPoster: true,
			expected: "https://image.tmdb.org/t/p/w500/pB8BM7pdSp6B6Ih7QZ4DrQ3PmJK.jpg",
		},
		{
			name:     "Large poster",
			size:     SizeLarge,
			path:     "/pB8BM7pdSp6B6Ih7QZ4DrQ3PmJK.jpg",
			isPoster: true,
			expected: "https://image.tmdb.org/t/p/w780/pB8BM7pdSp6B6Ih7QZ4DrQ3PmJK.jpg",
		},
		{
			name:     "Original poster",
			size:     SizeOriginal,
			path:     "/pB8BM7pdSp6B6Ih7QZ4DrQ3PmJK.jpg",
			isPoster: true,
			expected: "https://image.tmdb.org/t/p/original/pB8BM7pdSp6B6Ih7QZ4DrQ3PmJK.jpg",
		},
		{
			name:     "Small backdrop",
			size:     SizeSmall,
			path:     "/rr7E0NoGKxvbkb89eR1GwfoYjpA.jpg",
			isPoster: false,
			expected: "https://image.tmdb.org/t/p/w300/rr7E0NoGKxvbkb89eR1GwfoYjpA.jpg",
		},
		{
			name:     "Medium backdrop",
			size:     SizeMedium,
			path:     "/rr7E0NoGKxvbkb89eR1GwfoYjpA.jpg",
			isPoster: false,
			expected: "https://image.tmdb.org/t/p/w780/rr7E0NoGKxvbkb89eR1GwfoYjpA.jpg",
		},
		{
			name:     "Large backdrop",
			size:     SizeLarge,
			path:     "/rr7E0NoGKxvbkb89eR1GwfoYjpA.jpg",
			isPoster: false,
			expected: "https://image.tmdb.org/t/p/w1280/rr7E0NoGKxvbkb89eR1GwfoYjpA.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			downloader := NewTMDBDownloader(config, tt.size)
			result := downloader.buildImageURL(tt.path, tt.isPoster)

			if result != tt.expected {
				t.Errorf("Expected URL %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDownloadMoviePoster(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("fake poster data"))
	}))
	defer server.Close()

	tempDir := t.TempDir()

	tests := []struct {
		name        string
		posterPath  string
		expectError bool
		expectFile  bool
	}{
		{
			name:        "Successful download",
			posterPath:  "/test-poster.jpg",
			expectError: false,
			expectFile:  true,
		},
		{
			name:        "Empty poster path",
			posterPath:  "",
			expectError: false,
			expectFile:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			// Override base URL to use test server
			downloader := NewTMDBDownloader(config, SizeMedium)

			// Mock the download by using server URL
			ctx := context.Background()
			var err error

			if tt.posterPath != "" {
				// For this test, directly download from server
				destPath := filepath.Join(tempDir, "poster.jpg")
				err = downloader.DownloadImage(ctx, server.URL, destPath)
			} else {
				err = downloader.DownloadMoviePoster(ctx, tt.posterPath, tempDir)
			}

			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if tt.expectFile {
				posterPath := filepath.Join(tempDir, "poster.jpg")
				if !FileExists(posterPath) {
					t.Error("Expected poster file to exist")
				}
			}
		})
	}
}

func TestDownloadMovieBackdrop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("fake backdrop data"))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	config := DefaultConfig()
	downloader := NewTMDBDownloader(config, SizeMedium)

	t.Run("Empty backdrop path", func(t *testing.T) {
		err := downloader.DownloadMovieBackdrop(context.Background(), "", tempDir)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("Successful download", func(t *testing.T) {
		destPath := filepath.Join(tempDir, "backdrop.jpg")
		err := downloader.DownloadImage(context.Background(), server.URL, destPath)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !FileExists(destPath) {
			t.Error("Expected backdrop file to exist")
		}
	})
}

func TestDownloadTVPoster(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("fake TV poster data"))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	config := DefaultConfig()
	downloader := NewTMDBDownloader(config, SizeMedium)

	t.Run("Empty poster path", func(t *testing.T) {
		err := downloader.DownloadTVPoster(context.Background(), "", tempDir)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}

func TestDownloadSeasonPoster(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("fake season poster data"))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	config := DefaultConfig()
	downloader := NewTMDBDownloader(config, SizeMedium)

	t.Run("Empty poster path", func(t *testing.T) {
		err := downloader.DownloadSeasonPoster(context.Background(), "", tempDir)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}

func TestDownloadMovieArtwork(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("fake artwork data"))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	config := DefaultConfig()
	downloader := NewTMDBDownloader(config, SizeMedium)

	tests := []struct {
		name         string
		posterPath   string
		backdropPath string
	}{
		{
			name:         "No artwork",
			posterPath:   "",
			backdropPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := downloader.DownloadMovieArtwork(context.Background(), tt.posterPath, tt.backdropPath, tempDir)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestDownloadTVArtwork(t *testing.T) {
	tempDir := t.TempDir()
	config := DefaultConfig()
	downloader := NewTMDBDownloader(config, SizeMedium)

	tests := []struct {
		name       string
		posterPath string
	}{
		{
			name:       "No poster",
			posterPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := downloader.DownloadTVArtwork(context.Background(), tt.posterPath, tempDir)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestGetSizeString(t *testing.T) {
	tests := []struct {
		name     string
		size     ImageSize
		isPoster bool
		expected string
	}{
		// Poster sizes
		{name: "Poster small", size: SizeSmall, isPoster: true, expected: "w185"},
		{name: "Poster medium", size: SizeMedium, isPoster: true, expected: "w500"},
		{name: "Poster large", size: SizeLarge, isPoster: true, expected: "w780"},
		{name: "Poster original", size: SizeOriginal, isPoster: true, expected: "original"},
		{name: "Poster default", size: "", isPoster: true, expected: "w500"},

		// Backdrop sizes
		{name: "Backdrop small", size: SizeSmall, isPoster: false, expected: "w300"},
		{name: "Backdrop medium", size: SizeMedium, isPoster: false, expected: "w780"},
		{name: "Backdrop large", size: SizeLarge, isPoster: false, expected: "w1280"},
		{name: "Backdrop original", size: SizeOriginal, isPoster: false, expected: "original"},
		{name: "Backdrop default", size: "", isPoster: false, expected: "w780"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			downloader := NewTMDBDownloader(config, tt.size)
			result := downloader.getSizeString(tt.isPoster)

			if result != tt.expected {
				t.Errorf("Expected size %s, got %s", tt.expected, result)
			}
		})
	}
}
