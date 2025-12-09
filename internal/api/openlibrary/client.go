package openlibrary

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	// BaseURL is the OpenLibrary API base URL
	BaseURL = "https://openlibrary.org"

	// DefaultTimeout for HTTP requests
	DefaultTimeout = 10 * time.Second

	// Default cache TTL in seconds
	CacheTTLSuccess  = 86400 // 24 hours
	CacheTTLNotFound = 3600  // 1 hour

	// UserAgent for OpenLibrary API
	UserAgent = "go-jf-org/1.0 (https://github.com/opd-ai/go-jf-org)"
)

// Client represents an OpenLibrary API client
type Client struct {
	httpClient *http.Client
	cache      *Cache
	baseURL    string
	userAgent  string
}

// Config holds configuration for the OpenLibrary client
type Config struct {
	CacheDir  string
	Timeout   time.Duration
	UserAgent string
}

// NewClient creates a new OpenLibrary API client
func NewClient(config Config) (*Client, error) {
	if config.Timeout == 0 {
		config.Timeout = DefaultTimeout
	}

	if config.UserAgent == "" {
		config.UserAgent = UserAgent
	}

	cache, err := NewCache(config.CacheDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache: %w", err)
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		cache:     cache,
		baseURL:   BaseURL,
		userAgent: config.UserAgent,
	}, nil
}

// get performs a GET request to the OpenLibrary API with caching
func (c *Client) get(endpoint string, params url.Values) ([]byte, error) {
	// Construct URL
	apiURL := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	if params != nil && len(params) > 0 {
		apiURL = fmt.Sprintf("%s?%s", apiURL, params.Encode())
	}

	// Check cache first
	cacheKey := apiURL
	if cached, found := c.cache.Get(cacheKey); found {
		jsonData, err := json.Marshal(cached)
		if err == nil {
			log.Debug().Str("endpoint", endpoint).Msg("Using cached response")
			return jsonData, nil
		}
	}

	// Make HTTP request
	log.Debug().Str("endpoint", endpoint).Msg("Making OpenLibrary API request")
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle HTTP errors
	if resp.StatusCode != http.StatusOK {
		// Try to parse error response
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
			return nil, fmt.Errorf("OpenLibrary API error (%d): %s", resp.StatusCode, errResp.Error)
		}
		return nil, fmt.Errorf("OpenLibrary API error: HTTP %d", resp.StatusCode)
	}

	// Cache successful response
	var data interface{}
	if err := json.Unmarshal(body, &data); err == nil {
		if err := c.cache.Set(cacheKey, data, CacheTTLSuccess); err != nil {
			log.Warn().Err(err).Msg("Failed to cache response")
		}
	}

	return body, nil
}

// Search searches for books by title and/or author
func (c *Client) Search(title string, author string) (*SearchResponse, error) {
	if title == "" && author == "" {
		return nil, fmt.Errorf("title or author is required")
	}

	params := url.Values{}

	// Build query
	var queryParts []string
	if title != "" {
		queryParts = append(queryParts, title)
	}
	if author != "" {
		queryParts = append(queryParts, author)
	}
	params.Set("q", strings.Join(queryParts, " "))
	params.Set("limit", "5")

	body, err := c.get("/search.json", params)
	if err != nil {
		return nil, err
	}

	var response SearchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// GetBookByISBN retrieves book information by ISBN
func (c *Client) GetBookByISBN(isbn string) (*ISBNResponse, error) {
	if isbn == "" {
		return nil, fmt.Errorf("ISBN is required")
	}

	endpoint := fmt.Sprintf("/isbn/%s.json", isbn)
	body, err := c.get(endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response ISBNResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// GetBookDetails retrieves detailed information about a specific book
func (c *Client) GetBookDetails(bookKey string) (*BookDetails, error) {
	if bookKey == "" {
		return nil, fmt.Errorf("book key is required")
	}

	// Ensure key starts with /books/
	if !strings.HasPrefix(bookKey, "/books/") {
		bookKey = "/books/" + bookKey
	}

	endpoint := fmt.Sprintf("%s.json", bookKey)
	body, err := c.get(endpoint, nil)
	if err != nil {
		return nil, err
	}

	var details BookDetails
	if err := json.Unmarshal(body, &details); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &details, nil
}

// GetWorkDetails retrieves detailed information about a work
func (c *Client) GetWorkDetails(workKey string) (*WorkDetails, error) {
	if workKey == "" {
		return nil, fmt.Errorf("work key is required")
	}

	// Ensure key starts with /works/
	if !strings.HasPrefix(workKey, "/works/") {
		workKey = "/works/" + workKey
	}

	endpoint := fmt.Sprintf("%s.json", workKey)
	body, err := c.get(endpoint, nil)
	if err != nil {
		return nil, err
	}

	var details WorkDetails
	if err := json.Unmarshal(body, &details); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &details, nil
}

// GetAuthorDetails retrieves detailed information about an author
func (c *Client) GetAuthorDetails(authorKey string) (*AuthorDetails, error) {
	if authorKey == "" {
		return nil, fmt.Errorf("author key is required")
	}

	// Ensure key starts with /authors/
	if !strings.HasPrefix(authorKey, "/authors/") {
		authorKey = "/authors/" + authorKey
	}

	endpoint := fmt.Sprintf("%s.json", authorKey)
	body, err := c.get(endpoint, nil)
	if err != nil {
		return nil, err
	}

	var details AuthorDetails
	if err := json.Unmarshal(body, &details); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &details, nil
}

// GetCoverURL returns the URL for a book cover image
func (c *Client) GetCoverURL(coverID int, size string) string {
	if coverID == 0 {
		return ""
	}

	// Size can be: S (small), M (medium), L (large)
	if size == "" {
		size = "M"
	}

	return fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-%s.jpg", coverID, size)
}
