package musicbrainz

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
	// BaseURL is the MusicBrainz API base URL
	BaseURL = "https://musicbrainz.org/ws/2"

	// DefaultTimeout for HTTP requests
	DefaultTimeout = 10 * time.Second

	// Default cache TTL in seconds
	CacheTTLSuccess  = 86400 // 24 hours
	CacheTTLNotFound = 3600  // 1 hour

	// UserAgent for MusicBrainz API (required)
	UserAgent = "go-jf-org/1.0 (https://github.com/opd-ai/go-jf-org)"
)

// Client represents a MusicBrainz API client
type Client struct {
	httpClient  *http.Client
	rateLimiter *RateLimiter
	cache       *Cache
	baseURL     string
	userAgent   string
}

// Config holds configuration for the MusicBrainz client
type Config struct {
	CacheDir  string
	Timeout   time.Duration
	UserAgent string
}

// NewClient creates a new MusicBrainz API client
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
		rateLimiter: NewMusicBrainzRateLimiter(),
		cache:       cache,
		baseURL:     BaseURL,
		userAgent:   config.UserAgent,
	}, nil
}

// get performs a GET request to the MusicBrainz API with rate limiting and caching
func (c *Client) get(endpoint string, params url.Values) ([]byte, error) {
	// Add format parameter for JSON response
	if params == nil {
		params = url.Values{}
	}
	params.Set("fmt", "json")

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
	log.Debug().Str("endpoint", endpoint).Msg("Making MusicBrainz API request")
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// MusicBrainz requires User-Agent header
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
		if err := json.Unmarshal(body, &errResp); err == nil {
			return nil, fmt.Errorf("MusicBrainz API error (%d): %s", resp.StatusCode, errResp.Error)
		}
		return nil, fmt.Errorf("MusicBrainz API error: HTTP %d", resp.StatusCode)
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

// SearchRelease searches for releases (albums) by title and artist
func (c *Client) SearchRelease(title string, artist string) (*SearchReleaseResponse, error) {
	params := url.Values{}

	// Build Lucene query
	query := ""
	if title != "" {
		query += fmt.Sprintf("release:\"%s\"", title)
	}
	if artist != "" {
		if query != "" {
			query += " AND "
		}
		query += fmt.Sprintf("artist:\"%s\"", artist)
	}

	if query == "" {
		return nil, fmt.Errorf("title or artist is required")
	}

	params.Set("query", query)
	params.Set("limit", "5") // Limit results

	body, err := c.get("/release", params)
	if err != nil {
		return nil, err
	}

	var response SearchReleaseResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// GetReleaseDetails retrieves detailed information about a specific release
func (c *Client) GetReleaseDetails(releaseID string) (*ReleaseDetails, error) {
	params := url.Values{}
	params.Set("inc", "artists+labels+recordings") // Include related data

	endpoint := fmt.Sprintf("/release/%s", releaseID)
	body, err := c.get(endpoint, params)
	if err != nil {
		return nil, err
	}

	var details ReleaseDetails
	if err := json.Unmarshal(body, &details); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &details, nil
}

// SearchArtist searches for artists by name
func (c *Client) SearchArtist(name string) (*SearchArtistResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("artist name is required")
	}

	params := url.Values{}
	params.Set("query", fmt.Sprintf("artist:\"%s\"", name))
	params.Set("limit", "5")

	body, err := c.get("/artist", params)
	if err != nil {
		return nil, err
	}

	var response SearchArtistResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// GetArtistDetails retrieves detailed information about a specific artist
func (c *Client) GetArtistDetails(artistID string) (*ArtistDetails, error) {
	params := url.Values{}
	params.Set("inc", "aliases") // Include aliases

	endpoint := fmt.Sprintf("/artist/%s", artistID)
	body, err := c.get(endpoint, params)
	if err != nil {
		return nil, err
	}

	var details ArtistDetails
	if err := json.Unmarshal(body, &details); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &details, nil
}
