package config

import (
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	// Sources are the directories to scan for media files
	Sources []string `yaml:"sources"`
	// Destinations for organized media
	Destinations Destinations `yaml:"destinations"`
	// APIKeys for external services
	APIKeys APIKeys `yaml:"api_keys"`
	// Organize settings
	Organize OrganizeSettings `yaml:"organize"`
	// Safety settings
	Safety SafetySettings `yaml:"safety"`
	// Filters for file selection
	Filters FilterSettings `yaml:"filters"`
	// Performance settings
	Performance PerformanceSettings `yaml:"performance"`
}

// Destinations contains paths for different media types
type Destinations struct {
	Movies string `yaml:"movies"`
	TV     string `yaml:"tv"`
	Music  string `yaml:"music"`
	Books  string `yaml:"books"`
}

// APIKeys contains API keys for external services
type APIKeys struct {
	TMDB            string `yaml:"tmdb"`
	MusicBrainzApp  string `yaml:"musicbrainz_app"`
	LastFM          string `yaml:"lastfm"`
	GoogleBooksAPI  string `yaml:"google_books_api"`
}

// OrganizeSettings contains settings for file organization
type OrganizeSettings struct {
	CreateNFO            bool `yaml:"create_nfo"`
	DownloadArtwork      bool `yaml:"download_artwork"`
	NormalizeNames       bool `yaml:"normalize_names"`
	PreserveQualityTags  bool `yaml:"preserve_quality_tags"`
}

// SafetySettings contains safety-related settings
type SafetySettings struct {
	DryRun              bool   `yaml:"dry_run"`
	TransactionLog      bool   `yaml:"transaction_log"`
	LogDirectory        string `yaml:"log_directory"`
	ConflictResolution  string `yaml:"conflict_resolution"` // skip, rename, interactive
	BackupBeforeMove    bool   `yaml:"backup_before_move"`
}

// FilterSettings contains file filtering settings
type FilterSettings struct {
	MinFileSize       string   `yaml:"min_file_size"`
	VideoExtensions   []string `yaml:"video_extensions"`
	AudioExtensions   []string `yaml:"audio_extensions"`
	BookExtensions    []string `yaml:"book_extensions"`
}

// PerformanceSettings contains performance-related settings
type PerformanceSettings struct {
	MaxConcurrentOps int    `yaml:"max_concurrent_operations"`
	APIRateLimit     int    `yaml:"api_rate_limit"`
	CacheTTL         string `yaml:"cache_ttl"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".go-jf-org")
	
	return &Config{
		Sources: []string{},
		Destinations: Destinations{
			Movies: filepath.Join(homeDir, "media", "movies"),
			TV:     filepath.Join(homeDir, "media", "tv"),
			Music:  filepath.Join(homeDir, "media", "music"),
			Books:  filepath.Join(homeDir, "media", "books"),
		},
		APIKeys: APIKeys{
			MusicBrainzApp: "go-jf-org/1.0",
		},
		Organize: OrganizeSettings{
			CreateNFO:           true,
			DownloadArtwork:     true,
			NormalizeNames:      true,
			PreserveQualityTags: true,
		},
		Safety: SafetySettings{
			DryRun:             false,
			TransactionLog:     true,
			LogDirectory:       filepath.Join(configDir, "logs"),
			ConflictResolution: "skip",
			BackupBeforeMove:   false,
		},
		Filters: FilterSettings{
			MinFileSize: "10MB",
			VideoExtensions: []string{
				".mkv", ".mp4", ".avi", ".m4v", ".ts", ".webm",
				".mov", ".wmv", ".flv", ".mpg", ".mpeg",
			},
			AudioExtensions: []string{
				".mp3", ".flac", ".m4a", ".ogg", ".opus", ".wav",
				".aac", ".wma", ".ape", ".alac",
			},
			BookExtensions: []string{
				".epub", ".mobi", ".pdf", ".azw3", ".cbz", ".cbr",
			},
		},
		Performance: PerformanceSettings{
			MaxConcurrentOps: 4,
			APIRateLimit:     40,
			CacheTTL:         "24h",
		},
	}
}
