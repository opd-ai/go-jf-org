package openlibrary

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// TestNewClient tests creating a new OpenLibrary client
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
			name:    "default timeout",
			config:  Config{},
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
			tmpDir, err := os.MkdirTemp("", "ol-cache-test-*")
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

// TestSearch tests searching for books
func TestSearch(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify User-Agent header
		if r.Header.Get("User-Agent") == "" {
			t.Error("Missing User-Agent header")
		}

		// Return mock search results
		response := SearchResponse{
			NumFound: 1,
			Start:    0,
			Docs: []BookDoc{
				{
					Key:              "/works/OL27258W",
					Title:            "The Great Gatsby",
					AuthorName:       []string{"F. Scott Fitzgerald"},
					FirstPublishYear: 1925,
					ISBN:             []string{"9780743273565"},
					Publisher:        []string{"Scribner"},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with mock server
	tmpDir, err := os.MkdirTemp("", "ol-cache-test-*")
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
	result, err := client.Search("The Great Gatsby", "F. Scott Fitzgerald")
	if err != nil {
		t.Errorf("Search() error = %v", err)
		return
	}

	if result.NumFound != 1 {
		t.Errorf("Search() numFound = %d, want 1", result.NumFound)
	}

	if len(result.Docs) != 1 {
		t.Errorf("Search() docs = %d, want 1", len(result.Docs))
	}

	if result.Docs[0].Title != "The Great Gatsby" {
		t.Errorf("Search() title = %s, want The Great Gatsby", result.Docs[0].Title)
	}
}

// TestGetBookByISBN tests getting book by ISBN
func TestGetBookByISBN(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ISBNResponse{
			Key:   "/books/OL7353617M",
			Title: "The Great Gatsby",
			Authors: []AuthorRef{
				{Key: "/authors/OL9388A"},
			},
			Publishers:  []string{"Scribner"},
			PublishDate: "April 10, 1925",
			ISBN13:      []string{"9780743273565"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	tmpDir, err := os.MkdirTemp("", "ol-cache-test-*")
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

	// Test get by ISBN
	book, err := client.GetBookByISBN("9780743273565")
	if err != nil {
		t.Errorf("GetBookByISBN() error = %v", err)
		return
	}

	if book.Title != "The Great Gatsby" {
		t.Errorf("GetBookByISBN() title = %s, want The Great Gatsby", book.Title)
	}

	if len(book.Publishers) == 0 || book.Publishers[0] != "Scribner" {
		t.Errorf("GetBookByISBN() publisher = %v, want Scribner", book.Publishers)
	}
}

// TestGetCoverURL tests cover URL generation
func TestGetCoverURL(t *testing.T) {
	client := &Client{}

	tests := []struct {
		name     string
		coverID  int
		size     string
		expected string
	}{
		{
			name:     "medium cover",
			coverID:  12345,
			size:     "M",
			expected: "https://covers.openlibrary.org/b/id/12345-M.jpg",
		},
		{
			name:     "large cover",
			coverID:  67890,
			size:     "L",
			expected: "https://covers.openlibrary.org/b/id/67890-L.jpg",
		},
		{
			name:     "small cover",
			coverID:  11111,
			size:     "S",
			expected: "https://covers.openlibrary.org/b/id/11111-S.jpg",
		},
		{
			name:     "default size",
			coverID:  22222,
			size:     "",
			expected: "https://covers.openlibrary.org/b/id/22222-M.jpg",
		},
		{
			name:     "zero cover ID",
			coverID:  0,
			size:     "M",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := client.GetCoverURL(tt.coverID, tt.size)
			if url != tt.expected {
				t.Errorf("GetCoverURL() = %s, want %s", url, tt.expected)
			}
		})
	}
}

// TestCache tests caching functionality
func TestCache(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ol-cache-test-*")
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

// TestSearchErrors tests error handling
func TestSearchErrors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ol-cache-test-*")
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

	// Test with empty title and author
	_, err = client.Search("", "")
	if err == nil {
		t.Error("Search() with empty params should return error")
	}

	// Test with empty ISBN
	_, err = client.GetBookByISBN("")
	if err == nil {
		t.Error("GetBookByISBN() with empty ISBN should return error")
	}
}
