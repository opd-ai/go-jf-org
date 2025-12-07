package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	// Sources are the directories to scan for media files
	Sources []string `yaml:"sources" mapstructure:"sources"`
	// Destinations for organized media
	Destinations Destinations `yaml:"destinations" mapstructure:"destinations"`
	// APIKeys for external services
	APIKeys APIKeys `yaml:"api_keys" mapstructure:"api_keys"`
	// Organize settings
	Organize OrganizeSettings `yaml:"organize" mapstructure:"organize"`
	// Safety settings
	Safety SafetySettings `yaml:"safety" mapstructure:"safety"`
	// Filters for file selection
	Filters FilterSettings `yaml:"filters" mapstructure:"filters"`
	// Performance settings
	Performance PerformanceSettings `yaml:"performance" mapstructure:"performance"`
}

// Destinations contains paths for different media types
type Destinations struct {
	Movies string `yaml:"movies" mapstructure:"movies"`
	TV     string `yaml:"tv" mapstructure:"tv"`
	Music  string `yaml:"music" mapstructure:"music"`
	Books  string `yaml:"books" mapstructure:"books"`
}

// APIKeys contains API keys for external services
type APIKeys struct {
	TMDB           string `yaml:"tmdb" mapstructure:"tmdb"`
	MusicBrainzApp string `yaml:"musicbrainz_app" mapstructure:"musicbrainz_app"`
	LastFM         string `yaml:"lastfm" mapstructure:"lastfm"`
	GoogleBooksAPI string `yaml:"google_books_api" mapstructure:"google_books_api"`
}

// OrganizeSettings contains settings for file organization
type OrganizeSettings struct {
	CreateNFO           bool `yaml:"create_nfo" mapstructure:"create_nfo"`
	DownloadArtwork     bool `yaml:"download_artwork" mapstructure:"download_artwork"`
	NormalizeNames      bool `yaml:"normalize_names" mapstructure:"normalize_names"`
	PreserveQualityTags bool `yaml:"preserve_quality_tags" mapstructure:"preserve_quality_tags"`
}

// SafetySettings contains safety-related settings
type SafetySettings struct {
	DryRun             bool   `yaml:"dry_run" mapstructure:"dry_run"`
	TransactionLog     bool   `yaml:"transaction_log" mapstructure:"transaction_log"`
	LogDirectory       string `yaml:"log_directory" mapstructure:"log_directory"`
	ConflictResolution string `yaml:"conflict_resolution" mapstructure:"conflict_resolution"` // skip, rename, interactive
	BackupBeforeMove   bool   `yaml:"backup_before_move" mapstructure:"backup_before_move"`
}

// FilterSettings contains file filtering settings
type FilterSettings struct {
	MinFileSize     string   `yaml:"min_file_size" mapstructure:"min_file_size"`
	VideoExtensions []string `yaml:"video_extensions" mapstructure:"video_extensions"`
	AudioExtensions []string `yaml:"audio_extensions" mapstructure:"audio_extensions"`
	BookExtensions  []string `yaml:"book_extensions" mapstructure:"book_extensions"`
}

// PerformanceSettings contains performance-related settings
type PerformanceSettings struct {
	MaxConcurrentOps int    `yaml:"max_concurrent_operations" mapstructure:"max_concurrent_operations"`
	APIRateLimit     int    `yaml:"api_rate_limit" mapstructure:"api_rate_limit"`
	CacheTTL         string `yaml:"cache_ttl" mapstructure:"cache_ttl"`
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

// Load loads configuration from file and environment variables
func Load(cfgFile string) (*Config, error) {
	// Set defaults
	setDefaults()

	// Set config file path
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in standard locations
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}

		viper.AddConfigPath(filepath.Join(home, ".go-jf-org"))
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	// Read environment variables
	viper.SetEnvPrefix("GO_JF_ORG")
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		// Check if it's a file not found error (config is optional)
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found in search paths, that's okay
		} else if os.IsNotExist(err) {
			// Specific config file not found, that's okay too
		} else {
			// Other errors (permission denied, parse errors) should be returned
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	// Unmarshal into Config struct
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Apply defaults for empty slices (viper doesn't unmarshal defaults for slices properly)
	defaults := DefaultConfig()
	if len(cfg.Filters.VideoExtensions) == 0 {
		cfg.Filters.VideoExtensions = defaults.Filters.VideoExtensions
	}
	if len(cfg.Filters.AudioExtensions) == 0 {
		cfg.Filters.AudioExtensions = defaults.Filters.AudioExtensions
	}
	if len(cfg.Filters.BookExtensions) == 0 {
		cfg.Filters.BookExtensions = defaults.Filters.BookExtensions
	}
	if cfg.Filters.MinFileSize == "" {
		cfg.Filters.MinFileSize = defaults.Filters.MinFileSize
	}

	// Apply other defaults for empty strings
	if cfg.Safety.LogDirectory == "" {
		cfg.Safety.LogDirectory = defaults.Safety.LogDirectory
	}
	// Only apply conflict resolution default if it wasn't loaded
	if cfg.Safety.ConflictResolution == "" {
		cfg.Safety.ConflictResolution = defaults.Safety.ConflictResolution
	}
	if cfg.APIKeys.MusicBrainzApp == "" {
		cfg.APIKeys.MusicBrainzApp = defaults.APIKeys.MusicBrainzApp
	}
	if cfg.Performance.CacheTTL == "" {
		cfg.Performance.CacheTTL = defaults.Performance.CacheTTL
	}
	if cfg.Performance.MaxConcurrentOps == 0 {
		cfg.Performance.MaxConcurrentOps = defaults.Performance.MaxConcurrentOps
	}
	if cfg.Performance.APIRateLimit == 0 {
		cfg.Performance.APIRateLimit = defaults.Performance.APIRateLimit
	}

	return &cfg, nil
}

// setDefaults sets default values for viper
func setDefaults() {
	defaults := DefaultConfig()

	viper.SetDefault("organize.create_nfo", defaults.Organize.CreateNFO)
	viper.SetDefault("organize.download_artwork", defaults.Organize.DownloadArtwork)
	viper.SetDefault("organize.normalize_names", defaults.Organize.NormalizeNames)
	viper.SetDefault("organize.preserve_quality_tags", defaults.Organize.PreserveQualityTags)

	viper.SetDefault("safety.dry_run", defaults.Safety.DryRun)
	viper.SetDefault("safety.transaction_log", defaults.Safety.TransactionLog)
	viper.SetDefault("safety.log_directory", defaults.Safety.LogDirectory)
	viper.SetDefault("safety.conflict_resolution", defaults.Safety.ConflictResolution)
	viper.SetDefault("safety.backup_before_move", defaults.Safety.BackupBeforeMove)

	viper.SetDefault("filters.min_file_size", defaults.Filters.MinFileSize)
	viper.SetDefault("filters.video_extensions", defaults.Filters.VideoExtensions)
	viper.SetDefault("filters.audio_extensions", defaults.Filters.AudioExtensions)
	viper.SetDefault("filters.book_extensions", defaults.Filters.BookExtensions)

	viper.SetDefault("performance.max_concurrent_operations", defaults.Performance.MaxConcurrentOps)
	viper.SetDefault("performance.api_rate_limit", defaults.Performance.APIRateLimit)
	viper.SetDefault("performance.cache_ttl", defaults.Performance.CacheTTL)

	viper.SetDefault("api_keys.musicbrainz_app", defaults.APIKeys.MusicBrainzApp)
}

// ParseSize converts a size string (e.g., "10MB", "1GB") to bytes
func ParseSize(sizeStr string) (int64, error) {
	if sizeStr == "" {
		return 0, fmt.Errorf("empty size string")
	}

	// Regular expression to parse size with optional unit
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*([KMGT]?B)?$`)
	matches := re.FindStringSubmatch(strings.ToUpper(strings.TrimSpace(sizeStr)))

	if matches == nil {
		return 0, fmt.Errorf("invalid size format: %s", sizeStr)
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size value: %s", matches[1])
	}

	unit := matches[2]
	var multiplier int64 = 1

	switch unit {
	case "KB":
		multiplier = 1024
	case "MB":
		multiplier = 1024 * 1024
	case "GB":
		multiplier = 1024 * 1024 * 1024
	case "TB":
		multiplier = 1024 * 1024 * 1024 * 1024
	case "B", "":
		multiplier = 1
	}

	return int64(value * float64(multiplier)), nil
}
