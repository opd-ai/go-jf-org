package cmd

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/opd-ai/go-jf-org/internal/config"
	"github.com/opd-ai/go-jf-org/internal/scanner"
)

var scanCmd = &cobra.Command{
	Use:   "scan [directory]",
	Short: "Scan a directory for media files",
	Long: `Scan scans the specified directory (and subdirectories) for media files.

It identifies video, audio, and book files based on their extensions
and reports what it finds.`,
	Args: cobra.ExactArgs(1),
	RunE: runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)
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
			fmt.Printf("  [%s] %s\n", mediaType, file)
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
