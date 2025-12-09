package tmdb

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				APIKey:  "test-api-key",
				Timeout: 10 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "missing API key",
			config: Config{
				Timeout: 10 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "default timeout",
			config: Config{
				APIKey: "test-api-key",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use temp directory for cache
			tmpDir := t.TempDir()
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

func TestSearchMovie(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/search/movie" {
			t.Errorf("Expected path /search/movie, got %s", r.URL.Path)
		}

		query := r.URL.Query()
		if query.Get("query") == "" {
			t.Error("Missing query parameter")
		}

		// Return mock response
		response := SearchMovieResponse{
			Page: 1,
			Results: []MovieResult{
				{
					ID:          603,
					Title:       "The Matrix",
					ReleaseDate: "1999-03-31",
					Overview:    "A computer hacker learns about the true nature of his reality.",
					VoteAverage: 8.7,
				},
			},
			TotalPages:   1,
			TotalResults: 1,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with mock server
	tmpDir := t.TempDir()
	client, err := NewClient(Config{
		APIKey:   "test-key",
		CacheDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	client.baseURL = server.URL

	tests := []struct {
		name    string
		title   string
		year    int
		wantErr bool
		wantLen int
	}{
		{
			name:    "search with title and year",
			title:   "The Matrix",
			year:    1999,
			wantErr: false,
			wantLen: 1,
		},
		{
			name:    "search with title only",
			title:   "The Matrix",
			year:    0,
			wantErr: false,
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.SearchMovie(tt.title, tt.year)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchMovie() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(result.Results) != tt.wantLen {
					t.Errorf("SearchMovie() got %d results, want %d", len(result.Results), tt.wantLen)
				}
				if result.Results[0].Title != "The Matrix" {
					t.Errorf("SearchMovie() got title %s, want The Matrix", result.Results[0].Title)
				}
			}
		})
	}
}

func TestSearchTV(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search/tv" {
			t.Errorf("Expected path /search/tv, got %s", r.URL.Path)
		}

		response := SearchTVResponse{
			Page: 1,
			Results: []TVResult{
				{
					ID:           1396,
					Name:         "Breaking Bad",
					FirstAirDate: "2008-01-20",
					Overview:     "A high school chemistry teacher turned methamphetamine producer.",
					VoteAverage:  9.2,
				},
			},
			TotalPages:   1,
			TotalResults: 1,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	client, err := NewClient(Config{
		APIKey:   "test-key",
		CacheDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	client.baseURL = server.URL

	result, err := client.SearchTV("Breaking Bad", 2008)
	if err != nil {
		t.Fatalf("SearchTV() error = %v", err)
	}

	if len(result.Results) != 1 {
		t.Errorf("SearchTV() got %d results, want 1", len(result.Results))
	}

	if result.Results[0].Name != "Breaking Bad" {
		t.Errorf("SearchTV() got name %s, want Breaking Bad", result.Results[0].Name)
	}
}

func TestGetMovieDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/movie/603" {
			t.Errorf("Expected path /movie/603, got %s", r.URL.Path)
		}

		response := MovieDetails{
			ID:          603,
			Title:       "The Matrix",
			ReleaseDate: "1999-03-31",
			Runtime:     136,
			Overview:    "A computer hacker learns about the true nature of his reality.",
			IMDBID:      "tt0133093",
			Genres: []Genre{
				{ID: 28, Name: "Action"},
				{ID: 878, Name: "Science Fiction"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	client, err := NewClient(Config{
		APIKey:   "test-key",
		CacheDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	client.baseURL = server.URL

	details, err := client.GetMovieDetails(603)
	if err != nil {
		t.Fatalf("GetMovieDetails() error = %v", err)
	}

	if details.Title != "The Matrix" {
		t.Errorf("GetMovieDetails() got title %s, want The Matrix", details.Title)
	}

	if details.Runtime != 136 {
		t.Errorf("GetMovieDetails() got runtime %d, want 136", details.Runtime)
	}

	if len(details.Genres) != 2 {
		t.Errorf("GetMovieDetails() got %d genres, want 2", len(details.Genres))
	}
}

func TestCache(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewCache(tmpDir)
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	t.Run("set and get", func(t *testing.T) {
		key := "test-key"
		data := map[string]string{"test": "data"}
		ttl := 3600

		err := cache.Set(key, data, ttl)
		if err != nil {
			t.Errorf("Set() error = %v", err)
		}

		retrieved, found := cache.Get(key)
		if !found {
			t.Error("Get() cache miss, expected hit")
		}

		// Type assertion
		retrievedMap, ok := retrieved.(map[string]interface{})
		if !ok {
			t.Fatal("Get() returned wrong type")
		}

		if retrievedMap["test"] != "data" {
			t.Errorf("Get() returned %v, want map[test:data]", retrievedMap)
		}
	})

	t.Run("expired cache", func(t *testing.T) {
		key := "expired-key"
		data := map[string]string{"test": "data"}
		ttl := 1 // 1 second

		err := cache.Set(key, data, ttl)
		if err != nil {
			t.Errorf("Set() error = %v", err)
		}

		// Wait for expiration
		time.Sleep(2 * time.Second)

		_, found := cache.Get(key)
		if found {
			t.Error("Get() cache hit, expected miss for expired entry")
		}
	})

	t.Run("cache miss", func(t *testing.T) {
		_, found := cache.Get("non-existent-key")
		if found {
			t.Error("Get() cache hit, expected miss")
		}
	})

	t.Run("cache size", func(t *testing.T) {
		cache.Set("key1", "data1", 3600)
		cache.Set("key2", "data2", 3600)

		size, err := cache.Size()
		if err != nil {
			t.Errorf("Size() error = %v", err)
		}

		// At least 2 entries (may be more from previous tests)
		if size < 2 {
			t.Errorf("Size() = %d, want >= 2", size)
		}
	})

	t.Run("clear cache", func(t *testing.T) {
		err := cache.Clear()
		if err != nil {
			t.Errorf("Clear() error = %v", err)
		}

		size, _ := cache.Size()
		if size != 0 {
			t.Errorf("Size() after Clear() = %d, want 0", size)
		}
	})
}

func TestRateLimiter(t *testing.T) {
	t.Run("basic allow", func(t *testing.T) {
		rl := NewRateLimiter(10, 10, 1*time.Second)

		// Should allow first 10 requests
		for i := 0; i < 10; i++ {
			if !rl.Allow() {
				t.Errorf("Allow() returned false on request %d, expected true", i+1)
			}
		}

		// 11th request should be denied
		if rl.Allow() {
			t.Error("Allow() returned true when rate limit exceeded")
		}
	})

	t.Run("refill tokens", func(t *testing.T) {
		rl := NewRateLimiter(5, 5, 100*time.Millisecond)

		// Consume all tokens
		for i := 0; i < 5; i++ {
			rl.Allow()
		}

		// Wait for refill
		time.Sleep(150 * time.Millisecond)

		// Should have tokens again
		if !rl.Allow() {
			t.Error("Allow() returned false after refill period")
		}
	})

	t.Run("available tokens", func(t *testing.T) {
		rl := NewRateLimiter(10, 10, 1*time.Second)

		available := rl.Available()
		if available != 10 {
			t.Errorf("Available() = %d, want 10", available)
		}

		rl.Allow()
		rl.Allow()

		available = rl.Available()
		if available != 8 {
			t.Errorf("Available() after 2 requests = %d, want 8", available)
		}
	})

	t.Run("TMDB rate limiter", func(t *testing.T) {
		rl := NewTMDBRateLimiter()

		// Should start with 40 tokens
		available := rl.Available()
		if available != 40 {
			t.Errorf("NewTMDBRateLimiter() available = %d, want 40", available)
		}
	})
}

func TestAPIErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return error response
		w.WriteHeader(http.StatusUnauthorized)
		response := ErrorResponse{
			StatusCode:    7,
			StatusMessage: "Invalid API key",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	client, _ := NewClient(Config{
		APIKey:   "invalid-key",
		CacheDir: tmpDir,
	})
	client.baseURL = server.URL

	_, err := client.SearchMovie("Test", 2000)
	if err == nil {
		t.Error("SearchMovie() expected error, got nil")
	}
}

func TestCacheWithRealDirectory(t *testing.T) {
	tmpCacheDir := t.TempDir()

	cache, err := NewCache(tmpCacheDir)
	if err != nil {
		t.Fatalf("NewCache() with temp dir error = %v", err)
	}

	// Verify directory exists
	if _, err := os.Stat(tmpCacheDir); os.IsNotExist(err) {
		t.Error("Cache directory was not created")
	}

	// Test basic operations
	err = cache.Set("test", "data", 3600)
	if err != nil {
		t.Errorf("Set() error = %v", err)
	}

	_, found := cache.Get("test")
	if !found {
		t.Error("Get() cache miss after Set()")
	}
}
