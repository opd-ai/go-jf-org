package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/opd-ai/go-jf-org/internal/api/musicbrainz"
	"github.com/opd-ai/go-jf-org/internal/api/openlibrary"
	"github.com/opd-ai/go-jf-org/internal/api/tmdb"
	"github.com/opd-ai/go-jf-org/internal/config"
	"github.com/opd-ai/go-jf-org/internal/scanner"
	"github.com/opd-ai/go-jf-org/internal/util"
	"github.com/opd-ai/go-jf-org/pkg/types"
)

var (
	enrichScan bool
	jsonOutput bool
)

var scanCmd = &cobra.Command{
	Use:   "scan [directory]",
	Short: "Scan a directory for media files",
	Long: `Scan scans the specified directory (and subdirectories) for media files.

It identifies video, audio, and book files based on their extensions
and reports what it finds. Use --enrich to fetch metadata from external APIs
(TMDB for movies/TV, MusicBrainz for music, OpenLibrary for books).`,
	Args: cobra.ExactArgs(1),
	RunE: runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)
	scanCmd.Flags().BoolVar(&enrichScan, "enrich", false, "Enrich metadata using external APIs (TMDB, MusicBrainz, OpenLibrary)")
	scanCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output statistics in JSON format")
}

func runScan(cmd *cobra.Command, args []string) error {
	scanPath := args[0]

	// Make path absolute
	absPath, err := filepath.Abs(scanPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	log.Info().Str("path", absPath).Msg("Starting scan")

	// Create statistics tracker
	stats := util.NewStatistics()

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

	// Set up enrichers if requested
	var tmdbEnricher *tmdb.Enricher
	var mbEnricher *musicbrainz.Enricher
	var olEnricher *openlibrary.Enricher
	
	if enrichScan {
		// Set up TMDB enricher for movies and TV shows
		if cfg.APIKeys.TMDB == "" {
			log.Warn().Msg("TMDB API key not configured, skipping movie/TV enrichment. Set api_keys.tmdb in config.")
		} else {
			client, err := tmdb.NewClient(tmdb.Config{
				APIKey: cfg.APIKeys.TMDB,
			})
			if err != nil {
				log.Warn().Err(err).Msg("Failed to create TMDB client, skipping movie/TV enrichment")
			} else {
				tmdbEnricher = tmdb.NewEnricher(client)
				log.Info().Msg("TMDB enrichment enabled for movies and TV shows")
			}
		}

		// Set up MusicBrainz enricher for music
		mbClient, err := musicbrainz.NewClient(musicbrainz.Config{})
		if err != nil {
			log.Warn().Err(err).Msg("Failed to create MusicBrainz client, skipping music enrichment")
		} else {
			mbEnricher = musicbrainz.NewEnricher(mbClient)
			log.Info().Msg("MusicBrainz enrichment enabled for music")
		}

		// Set up OpenLibrary enricher for books
		olClient, err := openlibrary.NewClient(openlibrary.Config{})
		if err != nil {
			log.Warn().Err(err).Msg("Failed to create OpenLibrary client, skipping book enrichment")
		} else {
			olEnricher = openlibrary.NewEnricher(olClient)
			log.Info().Msg("OpenLibrary enrichment enabled for books")
		}
	}

	// Perform scan with progress tracking
	if !jsonOutput {
		fmt.Printf("Scanning %s...\n", absPath)
	}
	
	scanTimer := stats.NewTimer("scan")
	result, err := s.Scan(absPath)
	scanTimer.Stop()
	
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}
	
	stats.Add("files_found", len(result.Files))
	stats.Add("errors", len(result.Errors))

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
		// Set up progress tracking for metadata enrichment
		var progress *util.ProgressTracker
		if enrichScan && !jsonOutput {
			progress = util.NewProgressTracker(len(result.Files), "Enriching metadata")
			defer progress.Finish()
		}
		
		fmt.Println("Files found:")
		for _, file := range result.Files {
			mediaType := s.GetMediaType(file)
			metadata, err := s.GetMetadata(file)
			
			stats.Increment("files_processed")

			if err != nil {
				fmt.Printf("  [%s] %s (error parsing metadata: %v)\n", mediaType, file, err)
				continue
			}

			// Enrich metadata if enrichers are available
			if metadata != nil {
				var enriched bool
				switch mediaType {
				case types.MediaTypeMovie:
					if tmdbEnricher != nil {
						enrichTimer := stats.NewTimer("enrichment")
						if err := tmdbEnricher.EnrichMovie(metadata); err != nil {
							log.Debug().Err(err).Str("file", file).Msg("Failed to enrich movie metadata")
							stats.Increment("enrichment_failures")
						} else {
							stats.Increment("enrichment_success")
							enriched = true
						}
						enrichTimer.Stop()
					}
				case types.MediaTypeTV:
					if tmdbEnricher != nil {
						enrichTimer := stats.NewTimer("enrichment")
						if err := tmdbEnricher.EnrichTVShow(metadata); err != nil {
							log.Debug().Err(err).Str("file", file).Msg("Failed to enrich TV metadata")
							stats.Increment("enrichment_failures")
						} else {
							stats.Increment("enrichment_success")
							enriched = true
						}
						enrichTimer.Stop()
					}
				case types.MediaTypeMusic:
					if mbEnricher != nil {
						enrichTimer := stats.NewTimer("enrichment")
						if err := mbEnricher.EnrichMusic(metadata); err != nil {
							log.Debug().Err(err).Str("file", file).Msg("Failed to enrich music metadata")
							stats.Increment("enrichment_failures")
						} else {
							stats.Increment("enrichment_success")
							enriched = true
						}
						enrichTimer.Stop()
					}
				case types.MediaTypeBook:
					if olEnricher != nil {
						enrichTimer := stats.NewTimer("enrichment")
						if err := olEnricher.EnrichBook(metadata); err != nil {
							log.Debug().Err(err).Str("file", file).Msg("Failed to enrich book metadata")
							stats.Increment("enrichment_failures")
						} else {
							stats.Increment("enrichment_success")
							enriched = true
						}
						enrichTimer.Stop()
					}
				}
				_ = enriched // Silence unused variable warning
			}
			
			// Update progress if tracking
			if progress != nil {
				progress.Increment()
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
			case types.MediaTypeMusic:
				fmt.Printf("  [music] %s\n", file)
				if metadata.MusicMetadata != nil {
					if metadata.MusicMetadata.Artist != "" {
						fmt.Printf("          Artist: %s\n", metadata.MusicMetadata.Artist)
					}
					if metadata.MusicMetadata.Album != "" {
						fmt.Printf("          Album: %s", metadata.MusicMetadata.Album)
						if metadata.Year > 0 {
							fmt.Printf(" (%d)", metadata.Year)
						}
						fmt.Println()
					}
					if metadata.MusicMetadata.TrackNumber > 0 {
						fmt.Printf("          Track: %d\n", metadata.MusicMetadata.TrackNumber)
					}
					if metadata.MusicMetadata.Genre != "" {
						fmt.Printf("          Genre: %s\n", metadata.MusicMetadata.Genre)
					}
					if metadata.MusicMetadata.MusicBrainzRID != "" {
						fmt.Printf("          MusicBrainz ID: %s\n", metadata.MusicMetadata.MusicBrainzRID)
					}
				}
			case types.MediaTypeBook:
				fmt.Printf("  [book] %s\n", file)
				if metadata.BookMetadata != nil {
					if metadata.BookMetadata.Author != "" {
						fmt.Printf("          Author: %s\n", metadata.BookMetadata.Author)
					}
					if metadata.Title != "" {
						fmt.Printf("          Title: %s", metadata.Title)
						if metadata.Year > 0 {
							fmt.Printf(" (%d)", metadata.Year)
						}
						fmt.Println()
					}
					if metadata.BookMetadata.Publisher != "" {
						fmt.Printf("          Publisher: %s\n", metadata.BookMetadata.Publisher)
					}
					if metadata.BookMetadata.ISBN != "" {
						fmt.Printf("          ISBN: %s\n", metadata.BookMetadata.ISBN)
					}
					if metadata.BookMetadata.Description != "" {
						fmt.Printf("          Description: %s\n", truncate(metadata.BookMetadata.Description, 100))
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

	// Finalize and display statistics
	stats.Finish()
	
	if jsonOutput {
		// Output JSON statistics
		jsonStr, err := stats.ToJSON()
		if err != nil {
			log.Error().Err(err).Msg("Failed to generate JSON statistics")
		} else {
			fmt.Fprintln(os.Stdout, jsonStr)
		}
	} else if !verbose {
		// Show summary statistics for non-verbose mode
		fmt.Println()
		fmt.Printf("Scan completed in %s\n", util.FormatDuration(stats.Duration))
		if enrichScan {
			enrichSuccess := stats.Get("enrichment_success")
			enrichFailed := stats.Get("enrichment_failures")
			if enrichSuccess > 0 || enrichFailed > 0 {
				fmt.Printf("Enrichment: %d successful, %d failed\n", enrichSuccess, enrichFailed)
			}
		}
	}

	return nil
}

// truncate truncates a string to maxLen characters, adding "..." if truncated
func truncate(s string, maxLen int) string {
	if maxLen < 3 {
		maxLen = 3
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
