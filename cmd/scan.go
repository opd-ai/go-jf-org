package cmd

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/opd-ai/go-jf-org/internal/api/tmdb"
	"github.com/opd-ai/go-jf-org/internal/config"
	"github.com/opd-ai/go-jf-org/internal/scanner"
	"github.com/opd-ai/go-jf-org/pkg/types"
)

var (
	enrichScan bool
)

var scanCmd = &cobra.Command{
	Use:   "scan [directory]",
	Short: "Scan a directory for media files",
	Long: `Scan scans the specified directory (and subdirectories) for media files.

It identifies video, audio, and book files based on their extensions
and reports what it finds. Use --enrich to fetch metadata from TMDB.`,
	Args: cobra.ExactArgs(1),
	RunE: runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)
	scanCmd.Flags().BoolVar(&enrichScan, "enrich", false, "Enrich metadata using TMDB API (requires API key)")
}

func runScan(cmd *cobra.Command, args []string) error {
	scanPath := args[0]

	// Make path absolute
	absPath, err := filepath.Abs(scanPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	log.Info().Str("path", absPath).Msg("Starting scan")

	// Create scanner with configuration
	minSize := int64(10 * 1024 * 1024) // 10MB default
	if cfg.Filters.MinFileSize != "" {
		var err error
		minSize, err = config.ParseSize(cfg.Filters.MinFileSize)
		if err != nil {
			log.Warn().Err(err).Str("config_value", cfg.Filters.MinFileSize).Msg("Failed to parse MinFileSize, using default")
			minSize = 10 * 1024 * 1024
		}
	}

	s := scanner.NewScanner(
		cfg.Filters.VideoExtensions,
		cfg.Filters.AudioExtensions,
		cfg.Filters.BookExtensions,
		minSize,
	)

	// Set up enricher if requested
	var enricher *tmdb.Enricher
	if enrichScan {
		if cfg.APIKeys.TMDB == "" {
			log.Warn().Msg("TMDB API key not configured, skipping enrichment. Set api_keys.tmdb in config.")
		} else {
			client, err := tmdb.NewClient(tmdb.Config{
				APIKey: cfg.APIKeys.TMDB,
			})
			if err != nil {
				log.Warn().Err(err).Msg("Failed to create TMDB client, skipping enrichment")
			} else {
				enricher = tmdb.NewEnricher(client)
				log.Info().Msg("TMDB enrichment enabled")
			}
		}
	}

	// Perform scan
	result, err := s.Scan(absPath)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	// Display results
	fmt.Println()
	fmt.Printf("Scan Results for: %s\n", absPath)
	fmt.Println("=====================================")
	fmt.Printf("Total media files found: %d\n", len(result.Files))

	if len(result.Errors) > 0 {
		fmt.Printf("Errors encountered: %d\n", len(result.Errors))
	}

	fmt.Println()

	// Group by extension for summary
	extMap := make(map[string]int)
	for _, file := range result.Files {
		ext := filepath.Ext(file)
		extMap[ext]++
	}

	if len(extMap) > 0 {
		// Get sorted extensions for deterministic output
		exts := make([]string, 0, len(extMap))
		for ext := range extMap {
			exts = append(exts, ext)
		}
		sort.Strings(exts)

		fmt.Println("Files by extension:")
		for _, ext := range exts {
			fmt.Printf("  %s: %d\n", ext, extMap[ext])
		}
		fmt.Println()
	}

	// List all files if verbose
	if verbose {
		fmt.Println("Files found:")
		for _, file := range result.Files {
			mediaType := s.GetMediaType(file)
			metadata, err := s.GetMetadata(file)

			if err != nil {
				fmt.Printf("  [%s] %s (error parsing metadata: %v)\n", mediaType, file, err)
				continue
			}

			// Enrich metadata if enricher is available
			if enricher != nil && metadata != nil {
				switch mediaType {
				case types.MediaTypeMovie:
					if err := enricher.EnrichMovie(metadata); err != nil {
						log.Debug().Err(err).Str("file", file).Msg("Failed to enrich movie metadata")
					}
				case types.MediaTypeTV:
					if err := enricher.EnrichTVShow(metadata); err != nil {
						log.Debug().Err(err).Str("file", file).Msg("Failed to enrich TV metadata")
					}
				}
			}

			// Display based on media type
			switch mediaType {
			case types.MediaTypeMovie:
				fmt.Printf("  [movie] %s\n", file)
				if metadata.Title != "" {
					fmt.Printf("          Title: %s", metadata.Title)
					if metadata.Year > 0 {
						fmt.Printf(" (%d)", metadata.Year)
					}
					fmt.Println()
				}
				if metadata.Quality != "" || metadata.Source != "" || metadata.Codec != "" {
					fmt.Printf("          ")
					if metadata.Quality != "" {
						fmt.Printf("Quality: %s  ", metadata.Quality)
					}
					if metadata.Source != "" {
						fmt.Printf("Source: %s  ", metadata.Source)
					}
					if metadata.Codec != "" {
						fmt.Printf("Codec: %s", metadata.Codec)
					}
					fmt.Println()
				}
				// Show enriched data if available
				if metadata.MovieMetadata != nil {
					if metadata.MovieMetadata.Plot != "" {
						fmt.Printf("          Plot: %s\n", truncate(metadata.MovieMetadata.Plot, 100))
					}
					if metadata.MovieMetadata.Rating > 0 {
						fmt.Printf("          Rating: %.1f/10\n", metadata.MovieMetadata.Rating)
					}
					if len(metadata.MovieMetadata.Genres) > 0 {
						fmt.Printf("          Genres: %v\n", metadata.MovieMetadata.Genres)
					}
				}
			case types.MediaTypeTV:
				fmt.Printf("  [tv] %s\n", file)
				if metadata.TVMetadata != nil {
					if metadata.TVMetadata.ShowTitle != "" {
						fmt.Printf("          Show: %s  ", metadata.TVMetadata.ShowTitle)
					}
					if metadata.TVMetadata.Season > 0 || metadata.TVMetadata.Episode > 0 {
						fmt.Printf("S%02dE%02d", metadata.TVMetadata.Season, metadata.TVMetadata.Episode)
					}
					if metadata.TVMetadata.EpisodeTitle != "" {
						fmt.Printf("  %s", metadata.TVMetadata.EpisodeTitle)
					}
					fmt.Println()
					// Show enriched data if available
					if metadata.TVMetadata.Plot != "" {
						fmt.Printf("          Plot: %s\n", truncate(metadata.TVMetadata.Plot, 100))
					}
					if metadata.TVMetadata.Rating > 0 {
						fmt.Printf("          Rating: %.1f/10\n", metadata.TVMetadata.Rating)
					}
					if len(metadata.TVMetadata.Genres) > 0 {
						fmt.Printf("          Genres: %v\n", metadata.TVMetadata.Genres)
					}
				}
			default:
				fmt.Printf("  [%s] %s\n", mediaType, file)
				if metadata.Title != "" {
					fmt.Printf("          Title: %s\n", metadata.Title)
				}
			}
		}
		fmt.Println()
	}

	// Display any errors
	if len(result.Errors) > 0 && verbose {
		fmt.Println("Errors:")
		for _, err := range result.Errors {
			fmt.Printf("  %v\n", err)
		}
	}

	return nil
}

// truncate truncates a string to maxLen characters, adding "..." if truncated
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
