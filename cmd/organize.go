package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/opd-ai/go-jf-org/internal/artwork"
	"github.com/opd-ai/go-jf-org/internal/organizer"
	"github.com/opd-ai/go-jf-org/internal/safety"
	"github.com/opd-ai/go-jf-org/internal/util"
	"github.com/opd-ai/go-jf-org/pkg/types"
)

var (
	organizeDest            string
	organizeMediaType       string
	organizeConflictStrategy string
	organizeDryRun          bool
	organizeNoTransaction   bool
	organizeCreateNFO       bool
	organizeJSONOutput      bool
	organizeInteractive     bool
	organizeDownloadArtwork bool
	organizeArtworkSize     string
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
	organizeCmd.Flags().StringVar(&organizeConflictStrategy, "conflict", "skip", "conflict resolution strategy (skip, rename, interactive)")
	organizeCmd.Flags().BoolVar(&organizeDryRun, "dry-run", false, "preview changes without executing")
	organizeCmd.Flags().BoolVar(&organizeNoTransaction, "no-transaction", false, "disable transaction logging (not recommended)")
	organizeCmd.Flags().BoolVar(&organizeCreateNFO, "create-nfo", false, "create Jellyfin-compatible NFO metadata files")
	organizeCmd.Flags().BoolVar(&organizeDownloadArtwork, "download-artwork", false, "download poster and cover artwork for media")
	organizeCmd.Flags().StringVar(&organizeArtworkSize, "artwork-size", "medium", "artwork size preference (small, medium, large, original)")
	organizeCmd.Flags().BoolVar(&organizeJSONOutput, "json", false, "output statistics in JSON format")
	organizeCmd.Flags().BoolVar(&organizeInteractive, "interactive", false, "prompt for decisions on conflicts (sets conflict strategy to interactive)")
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
	
	// Handle interactive flag
	if organizeInteractive {
		organizeConflictStrategy = "interactive"
	}
	
	// Validate conflict strategy
	validStrategies := map[string]bool{
		"skip":        true,
		"rename":      true,
		"interactive": true,
	}
	if !validStrategies[organizeConflictStrategy] {
		return fmt.Errorf("invalid conflict strategy: %s (must be skip, rename, or interactive)", organizeConflictStrategy)
	}
	
	// Interactive mode requires TTY
	if organizeConflictStrategy == "interactive" {
		if organizeJSONOutput {
			return fmt.Errorf("interactive mode cannot be used with --json output")
		}
		if organizeDryRun {
			fmt.Println("⚠️  Note: Interactive mode in dry-run will simulate prompts without user input")
			fmt.Println()
		}
	}

	if organizeDryRun && !organizeJSONOutput {
		fmt.Println("⚠ DRY-RUN MODE: No files will be moved")
		fmt.Println()
	}

	log.Info().
		Str("path", absPath).
		Str("dest", destRoot).
		Bool("dry_run", organizeDryRun).
		Msg("Starting organization")

	// Create statistics tracker
	stats := util.NewStatistics()

	// Create scanner
	s := createScanner()

	// Scan for files with progress
	if !organizeJSONOutput {
		fmt.Printf("Scanning %s...\n", absPath)
	}
	scanSpinner := util.NewSpinner("Scanning for media files")
	if !organizeJSONOutput {
		scanSpinner.Start()
	}
	
	scanTimer := stats.NewTimer("scan")
	result, err := s.Scan(absPath)
	scanTimer.Stop()
	
	if !organizeJSONOutput {
		scanSpinner.Stop()
	}
	
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}
	
	stats.Add("files_scanned", len(result.Files))

	if len(result.Files) == 0 {
		fmt.Println("No media files found to organize.")
		return nil
	}

	fmt.Printf("Found %d media files\n\n", len(result.Files))

	// Create organizer with transaction support
	var org *organizer.Organizer
	var tm *safety.TransactionManager
	
	if !organizeNoTransaction && !organizeDryRun {
		logDir, err := safety.GetDefaultLogDir()
		if err != nil {
			log.Warn().Err(err).Msg("Failed to get transaction log directory, proceeding without transactions")
			org = organizer.NewOrganizer(organizeDryRun)
		} else {
			tm, err = safety.NewTransactionManager(logDir)
			if err != nil {
				log.Warn().Err(err).Msg("Failed to initialize transaction manager, proceeding without transactions")
				org = organizer.NewOrganizer(organizeDryRun)
			} else {
				org = organizer.NewOrganizerWithTransactions(organizeDryRun, tm)
			}
		}
	} else {
		org = organizer.NewOrganizer(organizeDryRun)
	}
	
	// Configure NFO generation
	org.SetCreateNFO(organizeCreateNFO)
	
	if organizeCreateNFO {
		log.Info().Msg("NFO file generation enabled")
	}
	
	// Configure artwork downloads
	if organizeDownloadArtwork {
		var artworkSize artwork.ImageSize
		switch organizeArtworkSize {
		case "small":
			artworkSize = artwork.SizeSmall
		case "medium":
			artworkSize = artwork.SizeMedium
		case "large":
			artworkSize = artwork.SizeLarge
		case "original":
			artworkSize = artwork.SizeOriginal
		default:
			artworkSize = artwork.SizeMedium
		}
		org.SetDownloadArtwork(true, artworkSize)
		log.Info().Str("size", organizeArtworkSize).Msg("Artwork download enabled")
	}

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
	if !organizeJSONOutput {
		fmt.Println()
	}

	// Execute organization with progress tracking
	if !organizeJSONOutput {
		if organizeDryRun {
			fmt.Println("Simulating file operations...")
		} else {
			fmt.Println("Organizing files...")
		}
	}
	
	// Handle interactive conflict resolution
	if organizeConflictStrategy == "interactive" && !organizeDryRun {
		plans = handleInteractiveConflicts(plans)
	}

	var ops []types.Operation
	var txnID string
	
	// Use the actual conflict strategy for execution
	// If interactive, conflicts have been resolved, so use "skip" for any remaining
	execStrategy := organizeConflictStrategy
	if organizeConflictStrategy == "interactive" {
		execStrategy = "skip" // Interactive conflicts already resolved
	}

	execTimer := stats.NewTimer("execution")
	if tm != nil {
		txnID, ops, err = org.ExecuteWithTransaction(plans, execStrategy)
		if err != nil {
			execTimer.Stop()
			return fmt.Errorf("organization failed: %w", err)
		}
	} else {
		ops, err = org.Execute(plans, execStrategy)
		if err != nil {
			execTimer.Stop()
			return fmt.Errorf("organization failed: %w", err)
		}
	}
	execTimer.Stop()

	// Count results and update statistics
	successCount := 0
	failedCount := 0
	skippedCount := len(plans) - len(ops) // Plans that were skipped due to conflicts
	var totalBytes int64

	for _, op := range ops {
		if op.Status == types.OperationStatusCompleted {
			successCount++
			// Try to get file size from destination (after move) or source (if still there)
			if info, err := os.Stat(op.Destination); err == nil {
				totalBytes += info.Size()
			} else if info, err := os.Stat(op.Source); err == nil {
				totalBytes += info.Size()
			}
		} else if op.Status == types.OperationStatusFailed {
			failedCount++
			// For failed operations, try to get file size from source
			if info, err := os.Stat(op.Source); err == nil {
				totalBytes += info.Size()
			}
		}
	}
	
	stats.Add("files_organized", successCount)
	stats.Add("files_failed", failedCount)
	stats.Add("files_skipped", skippedCount)
	stats.AddSize("total_bytes", totalBytes)

	// Display results
	if !organizeJSONOutput {
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

	// Display transaction ID if available
	if txnID != "" && !organizeJSONOutput {
		fmt.Printf("\nTransaction ID: %s\n", txnID)
		fmt.Printf("To rollback this operation, run: go-jf-org rollback %s\n", txnID)
	}

	// Success message
	if successCount > 0 && !organizeDryRun && !organizeJSONOutput {
		fmt.Printf("\n✓ Organization complete! Files are now in:\n")
		fmt.Printf("  %s\n", destRoot)
	}

	if organizeDryRun && !organizeJSONOutput {
		fmt.Println("\nTo execute this organization, run the same command without --dry-run")
	}
	
	// Finalize and display statistics
	stats.Finish()
	
	if organizeJSONOutput {
		jsonStr, err := stats.ToJSON()
		if err != nil {
			log.Error().Err(err).Msg("Failed to generate JSON statistics")
		} else {
			fmt.Fprintln(os.Stdout, jsonStr)
		}
	} else {
		// Show summary statistics
		fmt.Println()
		fmt.Printf("Completed in %s\n", util.FormatDuration(stats.Duration))
		if totalBytes > 0 {
			fmt.Printf("Total data processed: %s\n", util.FormatBytes(totalBytes))
		}
	}

	return nil
}

// handleInteractiveConflicts processes plans with conflicts and prompts user for resolution
func handleInteractiveConflicts(plans []organizer.Plan) []organizer.Plan {
	skipAll := false
	result := make([]organizer.Plan, 0, len(plans))
	
	for _, plan := range plans {
		if !plan.Conflict {
			// No conflict, keep as-is
			result = append(result, plan)
			continue
		}
		
		// Skip remaining conflicts if user chose "skip all"
		if skipAll {
			log.Info().Str("file", plan.SourcePath).Msg("Skipping due to 'skip all' choice")
			continue
		}
		
		// Prompt user for resolution
		choice := promptConflictResolution(plan.SourcePath, plan.DestinationPath)
		
		switch choice {
		case "skip":
			log.Info().Str("file", plan.SourcePath).Msg("User chose to skip")
			// Don't add to result - file will be skipped
		case "skip-all":
			skipAll = true
			log.Info().Msg("User chose to skip all remaining conflicts")
			// Don't add to result - file will be skipped
		case "rename":
			// Add suffix to destination
			newPath, err := findAvailableName(plan.DestinationPath)
			if err != nil {
				log.Error().Err(err).Str("file", plan.SourcePath).Msg("Failed to find available name, skipping")
				continue
			}
			plan.DestinationPath = newPath
			plan.Conflict = false // Conflict resolved
			log.Info().Str("file", plan.SourcePath).Str("new_dest", plan.DestinationPath).Msg("User chose to rename")
			result = append(result, plan)
		default:
			log.Warn().Str("file", plan.SourcePath).Str("choice", choice).Msg("Unknown choice, skipping")
		}
	}
	
	return result
}

// findAvailableName finds an available filename by adding -1, -2, etc suffix
func findAvailableName(path string) (string, error) {
	ext := filepath.Ext(path)
	base := path[:len(path)-len(ext)]
	
	for i := 1; i < 1000; i++ {
		newPath := fmt.Sprintf("%s-%d%s", base, i, ext)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath, nil
		}
	}
	
	return "", fmt.Errorf("could not find available filename after 1000 attempts")
}
