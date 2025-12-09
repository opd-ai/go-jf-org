package musicbrainz

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
)

// Cache manages local caching of MusicBrainz API responses
type Cache struct {
	dir string
}

// NewCache creates a new cache instance
// Default location: ~/.go-jf-org/cache/musicbrainz/
func NewCache(cacheDir string) (*Cache, error) {
	if cacheDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		cacheDir = filepath.Join(home, ".go-jf-org", "cache", "musicbrainz")
	}

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &Cache{dir: cacheDir}, nil
}

// Get retrieves a cached response if it exists and is not expired
func (c *Cache) Get(key string) (interface{}, bool) {
	filename := c.getCacheFilename(key)

	data, err := os.ReadFile(filename)
	if err != nil {
		// Cache miss - file doesn't exist or can't be read
		return nil, false
	}

	var cached CachedResponse
	if err := json.Unmarshal(data, &cached); err != nil {
		log.Debug().Err(err).Str("file", filename).Msg("Failed to unmarshal cached response")
		return nil, false
	}

	// Check if cache entry has expired
	expiresAt := cached.Timestamp.Add(time.Duration(cached.TTL) * time.Second)
	if time.Now().After(expiresAt) {
		log.Debug().Str("key", key).Msg("Cache entry expired")
		// Remove expired cache file
		if err := os.Remove(filename); err != nil {
			log.Warn().Err(err).Str("file", filename).Msg("Failed to remove expired cache file")
		}
		return nil, false
	}

	log.Debug().Str("key", key).Msg("Cache hit")
	return cached.Data, true
}

// Set stores a response in the cache with the specified TTL
func (c *Cache) Set(key string, data interface{}, ttl int) error {
	cached := CachedResponse{
		Data:      data,
		Timestamp: time.Now(),
		TTL:       ttl,
	}

	jsonData, err := json.MarshalIndent(cached, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	filename := c.getCacheFilename(key)
	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	log.Debug().Str("key", key).Int("ttl", ttl).Msg("Cached response")
	return nil
}

// getCacheFilename generates a cache filename from a key using SHA-256 hash
func (c *Cache) getCacheFilename(key string) string {
	hash := sha256.Sum256([]byte(key))
	hashStr := hex.EncodeToString(hash[:])
	return filepath.Join(c.dir, hashStr+".json")
}

// Clear removes all cached responses
func (c *Cache) Clear() error {
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			if err := os.Remove(filepath.Join(c.dir, entry.Name())); err != nil {
				log.Warn().Err(err).Str("file", entry.Name()).Msg("Failed to remove cache file")
			}
		}
	}

	log.Info().Msg("Cache cleared")
	return nil
}

// Size returns the number of cached entries
func (c *Cache) Size() (int, error) {
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return 0, fmt.Errorf("failed to read cache directory: %w", err)
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			count++
		}
	}

	return count, nil
}
