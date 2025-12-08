package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/opd-ai/go-jf-org/internal/organizer"
	"github.com/opd-ai/go-jf-org/pkg/types"
)

var (
	organizeDest            string
	organizeMediaType       string
	organizeConflictStrategy string
	organizeDryRun          bool
)

var organizeCmd = &cobra.Command{
	Use:   "organize [directory]",
	Short: "Organize media files into Jellyfin-compatible structure",
	Long: `Organize scans the specified directory and moves media files into a
Jellyfin-compatible directory structure with proper naming conventions.

The organize command:
  - Detects media types (movies, TV shows, music, books)
  - Parses metadata from filenames
  - Organizes files into proper directory structures
  - Renames files according to Jellyfin conventions
  - Handles conflicts based on specified strategy

Safety features:
  - Files are moved, never deleted
  - Conflict resolution strategies available
  - Dry-run mode for testing (--dry-run)
  - Validation before operations`,
	Args: cobra.ExactArgs(1),
	RunE: runOrganize,
}

func init() {
	rootCmd.AddCommand(organizeCmd)

	organizeCmd.Flags().StringVarP(&organizeDest, "dest", "d", "", "destination root directory (default from config)")
	organizeCmd.Flags().StringVarP(&organizeMediaType, "type", "t", "", "filter by media type (movie, tv, music, book)")
	organizeCmd.Flags().StringVar(&organizeConflictStrategy, "conflict", "skip", "conflict resolution strategy (skip, rename)")
	organizeCmd.Flags().BoolVar(&organizeDryRun, "dry-run", false, "preview changes without executing")
}

func runOrganize(cmd *cobra.Command, args []string) error {
	scanPath := args[0]

	// Make path absolute
	absPath, err := filepath.Abs(scanPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Determine destination root
	destRoot, err := getDestinationRoot(organizeMediaType, organizeDest)
	if err != nil {
		return err
	}

	// Parse media type filter
	mediaTypeFilter, err := parseMediaTypeFilter(organizeMediaType)
	if err != nil {
		return err
	}

	if organizeDryRun {
		fmt.Println("⚠ DRY-RUN MODE: No files will be moved")
		fmt.Println()
	}

	log.Info().
		Str("path", absPath).
		Str("dest", destRoot).
		Bool("dry_run", organizeDryRun).
		Msg("Starting organization")

	// Create scanner
	s := createScanner()

	// Scan for files
	fmt.Printf("Scanning %s...\n", absPath)
	result, err := s.Scan(absPath)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	if len(result.Files) == 0 {
		fmt.Println("No media files found to organize.")
		return nil
	}

	fmt.Printf("Found %d media files\n\n", len(result.Files))

	// Create organizer
	org := organizer.NewOrganizer(organizeDryRun)

	// Plan organization
	fmt.Println("Planning organization...")
	plans, err := org.PlanOrganization(result.Files, destRoot, mediaTypeFilter)
	if err != nil {
		return fmt.Errorf("failed to plan organization: %w", err)
	}

	if len(plans) == 0 {
		fmt.Println("No files match the criteria for organization.")
		return nil
	}

	fmt.Printf("Planned %d file operations\n\n", len(plans))

	// Validate plans
	validationErrors := org.ValidatePlan(plans)
	if len(validationErrors) > 0 {
		fmt.Printf("⚠ Warning: %d validation errors found:\n", len(validationErrors))
		for _, err := range validationErrors {
			fmt.Printf("  - %v\n", err)
		}
		fmt.Println("\nProceeding with valid files only...")
	}

	// Count by type and conflicts
	movieCount := 0
	tvCount := 0
	musicCount := 0
	bookCount := 0
	conflictCount := 0

	for _, plan := range plans {
		switch plan.MediaType {
		case types.MediaTypeMovie:
			movieCount++
		case types.MediaTypeTV:
			tvCount++
		case types.MediaTypeMusic:
			musicCount++
		case types.MediaTypeBook:
			bookCount++
		}
		if plan.Conflict {
			conflictCount++
		}
	}

	// Display summary
	fmt.Println("Organization Summary:")
	fmt.Println("====================")
	if movieCount > 0 {
		fmt.Printf("Movies: %d\n", movieCount)
	}
	if tvCount > 0 {
		fmt.Printf("TV Shows: %d\n", tvCount)
	}
	if musicCount > 0 {
		fmt.Printf("Music: %d\n", musicCount)
	}
	if bookCount > 0 {
		fmt.Printf("Books: %d\n", bookCount)
	}

	if conflictCount > 0 {
		fmt.Printf("\n⚠ Conflicts: %d (strategy: %s)\n", conflictCount, organizeConflictStrategy)
	}
	fmt.Println()

	// Execute organization
	if organizeDryRun {
		fmt.Println("Simulating file operations...")
	} else {
		fmt.Println("Organizing files...")
	}

	ops, err := org.Execute(plans, organizeConflictStrategy)
	if err != nil {
		return fmt.Errorf("organization failed: %w", err)
	}

	// Count results
	successCount := 0
	failedCount := 0
	skippedCount := len(plans) - len(ops) // Plans that were skipped due to conflicts

	for _, op := range ops {
		if op.Status == types.OperationStatusCompleted {
			successCount++
		} else if op.Status == types.OperationStatusFailed {
			failedCount++
		}
	}

	// Display results
	fmt.Println()
	fmt.Println("Results:")
	fmt.Println("========")
	if organizeDryRun {
		fmt.Printf("Would organize: %d files\n", successCount)
	} else {
		fmt.Printf("✓ Successfully organized: %d files\n", successCount)
	}
	if failedCount > 0 {
		fmt.Printf("✗ Failed: %d files\n", failedCount)
	}
	if skippedCount > 0 {
		fmt.Printf("⊘ Skipped: %d files\n", skippedCount)
	}

	// Display failures if any
	if failedCount > 0 && verbose {
		fmt.Println("\nFailed Operations:")
		for _, op := range ops {
			if op.Status == types.OperationStatusFailed {
				fmt.Printf("  ✗ %s\n", op.Source)
				fmt.Printf("    Error: %v\n", op.Error)
			}
		}
	}

	// Success message
	if successCount > 0 && !organizeDryRun {
		fmt.Printf("\n✓ Organization complete! Files are now in:\n")
		fmt.Printf("  %s\n", destRoot)
	}

	if organizeDryRun {
		fmt.Println("\nTo execute this organization, run the same command without --dry-run")
	}

	return nil
}
