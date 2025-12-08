package musicbrainz

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestNewClient tests creating a new MusicBrainz client
func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Timeout: 5 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "default timeout",
			config: Config{},
			wantErr: false,
		},
		{
			name: "custom user agent",
			config: Config{
				UserAgent: "test-app/1.0",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use temporary cache directory
			tmpDir, err := os.MkdirTemp("", "mb-cache-test-*")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			tt.config.CacheDir = tmpDir

			client, err := NewClient(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil client")
			}
		})
	}
}

// TestSearchRelease tests searching for releases
func TestSearchRelease(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify User-Agent header
		if r.Header.Get("User-Agent") == "" {
			t.Error("Missing User-Agent header")
		}

		// Return mock search results
		response := SearchReleaseResponse{
			Count: 1,
			Releases: []Release{
				{
					ID:    "test-release-id",
					Title: "Dark Side of the Moon",
					Date:  "1973-03-01",
					ArtistCredit: []ArtistCredit{
						{
							Name: "Pink Floyd",
							Artist: Artist{
								ID:   "artist-id",
								Name: "Pink Floyd",
							},
						},
					},
					ReleaseGroup: ReleaseGroup{
						ID:               "rg-id",
						Title:            "Dark Side of the Moon",
						PrimaryType:      "Album",
						FirstReleaseDate: "1973",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with mock server
	tmpDir, err := os.MkdirTemp("", "mb-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	client, err := NewClient(Config{
		CacheDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Override base URL to use mock server
	client.baseURL = server.URL

	// Test search
	result, err := client.SearchRelease("Dark Side of the Moon", "Pink Floyd")
	if err != nil {
		t.Errorf("SearchRelease() error = %v", err)
		return
	}

	if result.Count != 1 {
		t.Errorf("SearchRelease() count = %d, want 1", result.Count)
	}

	if len(result.Releases) != 1 {
		t.Errorf("SearchRelease() results = %d, want 1", len(result.Releases))
	}

	if result.Releases[0].Title != "Dark Side of the Moon" {
		t.Errorf("SearchRelease() title = %s, want Dark Side of the Moon", result.Releases[0].Title)
	}
}

// TestGetReleaseDetails tests getting release details
func TestGetReleaseDetails(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ReleaseDetails{
			ID:    "test-release-id",
			Title: "Dark Side of the Moon",
			Date:  "1973-03-01",
			ArtistCredit: []ArtistCredit{
				{
					Name: "Pink Floyd",
					Artist: Artist{
						ID:   "artist-id",
						Name: "Pink Floyd",
					},
				},
			},
			ReleaseGroup: ReleaseGroup{
				ID:               "rg-id",
				Title:            "Dark Side of the Moon",
				PrimaryType:      "Album",
				FirstReleaseDate: "1973",
			},
			Media: []Media{
				{
					Format:     "CD",
					TrackCount: 10,
					Position:   1,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	tmpDir, err := os.MkdirTemp("", "mb-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	client, err := NewClient(Config{
		CacheDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	client.baseURL = server.URL

	// Test get details
	details, err := client.GetReleaseDetails("test-release-id")
	if err != nil {
		t.Errorf("GetReleaseDetails() error = %v", err)
		return
	}

	if details.Title != "Dark Side of the Moon" {
		t.Errorf("GetReleaseDetails() title = %s, want Dark Side of the Moon", details.Title)
	}

	if len(details.Media) != 1 {
		t.Errorf("GetReleaseDetails() media count = %d, want 1", len(details.Media))
	}
}

// TestCache tests caching functionality
func TestCache(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mb-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cache, err := NewCache(tmpDir)
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	// Test cache miss
	_, found := cache.Get("test-key")
	if found {
		t.Error("Expected cache miss, got hit")
	}

	// Test cache set
	testData := map[string]string{"test": "data"}
	err = cache.Set("test-key", testData, 3600)
	if err != nil {
		t.Errorf("Set() error = %v", err)
	}

	// Test cache hit
	data, found := cache.Get("test-key")
	if !found {
		t.Error("Expected cache hit, got miss")
	}

	// Verify data
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		t.Error("Cached data has wrong type")
	}
	if dataMap["test"] != "data" {
		t.Errorf("Cached data = %v, want {test: data}", dataMap)
	}

	// Test cache size
	size, err := cache.Size()
	if err != nil {
		t.Errorf("Size() error = %v", err)
	}
	if size != 1 {
		t.Errorf("Size() = %d, want 1", size)
	}

	// Test cache clear
	err = cache.Clear()
	if err != nil {
		t.Errorf("Clear() error = %v", err)
	}

	size, err = cache.Size()
	if err != nil {
		t.Errorf("Size() error = %v", err)
	}
	if size != 0 {
		t.Errorf("Size() after clear = %d, want 0", size)
	}
}

// TestRateLimiter tests rate limiting functionality
func TestRateLimiter(t *testing.T) {
	t.Run("allow consumes token", func(t *testing.T) {
		rl := NewMusicBrainzRateLimiter()

		// First request should be allowed
		if !rl.Allow() {
			t.Error("First request should be allowed")
		}

		// Second immediate request should be denied (1 req/sec limit)
		if rl.Allow() {
			t.Error("Second immediate request should be denied")
		}

		// Check available tokens
		available := rl.Available()
		if available != 0 {
			t.Errorf("Available() = %d, want 0", available)
		}
	})

	t.Run("tokens refill over time", func(t *testing.T) {
		rl := NewMusicBrainzRateLimiter()

		// Consume token
		rl.Allow()

		// Wait for refill
		time.Sleep(1100 * time.Millisecond)

		// Should have token again
		available := rl.Available()
		if available != 1 {
			t.Errorf("Available() after refill = %d, want 1", available)
		}
	})

	t.Run("wait blocks until token available", func(t *testing.T) {
		rl := NewMusicBrainzRateLimiter()

		// Consume initial token
		rl.Allow()

		start := time.Now()
		// Wait should block for ~1 second
		rl.Wait()
		elapsed := time.Since(start)

		if elapsed < 900*time.Millisecond {
			t.Errorf("Wait() elapsed = %v, want >= 900ms", elapsed)
		}
	})
}

// TestCacheExpiration tests cache expiration
func TestCacheExpiration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mb-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cache, err := NewCache(tmpDir)
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	// Set with 1 second TTL
	testData := map[string]string{"test": "data"}
	err = cache.Set("test-key", testData, 1)
	if err != nil {
		t.Errorf("Set() error = %v", err)
	}

	// Should be available immediately
	_, found := cache.Get("test-key")
	if !found {
		t.Error("Expected cache hit")
	}

	// Wait for expiration
	time.Sleep(1200 * time.Millisecond)

	// Should be expired
	_, found = cache.Get("test-key")
	if found {
		t.Error("Expected cache miss after expiration")
	}

	// Cache file should be removed
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read cache dir: %v", err)
	}

	jsonFiles := 0
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".json" {
			jsonFiles++
		}
	}

	if jsonFiles != 0 {
		t.Errorf("Expected 0 cache files after expiration, got %d", jsonFiles)
	}
}
