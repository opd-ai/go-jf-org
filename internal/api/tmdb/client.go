package tmdb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	// BaseURL is the TMDB API base URL
	BaseURL = "https://api.themoviedb.org/3"

	// DefaultTimeout for HTTP requests
	DefaultTimeout = 10 * time.Second

	// Default cache TTL in seconds
	CacheTTLSuccess = 86400 // 24 hours
	CacheTTLNotFound = 3600 // 1 hour
)

// Client represents a TMDB API client
type Client struct {
	apiKey      string
	httpClient  *http.Client
	rateLimiter *RateLimiter
	cache       *Cache
	baseURL     string
}

// Config holds configuration for the TMDB client
type Config struct {
	APIKey   string
	CacheDir string
	Timeout  time.Duration
}

// NewClient creates a new TMDB API client
func NewClient(config Config) (*Client, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("TMDB API key is required")
	}

	if config.Timeout == 0 {
		config.Timeout = DefaultTimeout
	}

	cache, err := NewCache(config.CacheDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache: %w", err)
	}

	return &Client{
		apiKey: config.APIKey,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		rateLimiter: NewTMDBRateLimiter(),
		cache:       cache,
		baseURL:     BaseURL,
	}, nil
}

// get performs a GET request to the TMDB API with rate limiting and caching
func (c *Client) get(endpoint string, params url.Values) ([]byte, error) {
	// Add API key to parameters
	if params == nil {
		params = url.Values{}
	}
	params.Set("api_key", c.apiKey)

	// Construct URL
	apiURL := fmt.Sprintf("%s%s?%s", c.baseURL, endpoint, params.Encode())

	// Check cache first
	cacheKey := apiURL
	if cached, found := c.cache.Get(cacheKey); found {
		jsonData, err := json.Marshal(cached)
		if err == nil {
			log.Debug().Str("endpoint", endpoint).Msg("Using cached response")
			return jsonData, nil
		}
	}

	// Rate limiting - wait for token
	log.Debug().Str("endpoint", endpoint).Msg("Waiting for rate limiter")
	c.rateLimiter.Wait()

	// Make HTTP request
	log.Debug().Str("endpoint", endpoint).Msg("Making TMDB API request")
	resp, err := c.httpClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for error responses
	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			return nil, fmt.Errorf("TMDB API error (%d): %s", errResp.StatusCode, errResp.StatusMessage)
		}
		return nil, fmt.Errorf("TMDB API returned status %d", resp.StatusCode)
	}

	// Cache successful response
	var data interface{}
	if err := json.Unmarshal(body, &data); err == nil {
		if err := c.cache.Set(cacheKey, data, CacheTTLSuccess); err != nil {
			log.Warn().Err(err).Str("endpoint", endpoint).Msg("Failed to cache TMDB response")
		}
	}

	return body, nil
}

// SearchMovie searches for movies by title and optional year
func (c *Client) SearchMovie(title string, year int) (*SearchMovieResponse, error) {
	params := url.Values{}
	params.Set("query", title)
	if year > 0 {
		params.Set("year", fmt.Sprintf("%d", year))
	}

	body, err := c.get("/search/movie", params)
	if err != nil {
		return nil, err
	}

	var result SearchMovieResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse movie search response: %w", err)
	}

	log.Info().
		Str("title", title).
		Int("year", year).
		Int("results", len(result.Results)).
		Msg("Movie search completed")

	return &result, nil
}

// GetMovieDetails retrieves detailed information for a movie by ID
func (c *Client) GetMovieDetails(movieID int) (*MovieDetails, error) {
	endpoint := fmt.Sprintf("/movie/%d", movieID)

	body, err := c.get(endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result MovieDetails
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse movie details response: %w", err)
	}

	log.Info().
		Int("id", movieID).
		Str("title", result.Title).
		Msg("Movie details retrieved")

	return &result, nil
}

// SearchTV searches for TV shows by name and optional year
func (c *Client) SearchTV(name string, year int) (*SearchTVResponse, error) {
	params := url.Values{}
	params.Set("query", name)
	if year > 0 {
		params.Set("first_air_date_year", fmt.Sprintf("%d", year))
	}

	body, err := c.get("/search/tv", params)
	if err != nil {
		return nil, err
	}

	var result SearchTVResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse TV search response: %w", err)
	}

	log.Info().
		Str("name", name).
		Int("year", year).
		Int("results", len(result.Results)).
		Msg("TV search completed")

	return &result, nil
}

// GetTVDetails retrieves detailed information for a TV show by ID
func (c *Client) GetTVDetails(tvID int) (*TVDetails, error) {
	endpoint := fmt.Sprintf("/tv/%d", tvID)

	body, err := c.get(endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result TVDetails
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse TV details response: %w", err)
	}

	log.Info().
		Int("id", tvID).
		Str("name", result.Name).
		Msg("TV details retrieved")

	return &result, nil
}

// ClearCache clears all cached TMDB responses
func (c *Client) ClearCache() error {
	return c.cache.Clear()
}

// GetCacheSize returns the number of cached entries
func (c *Client) GetCacheSize() (int, error) {
	return c.cache.Size()
}
