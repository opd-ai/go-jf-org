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
	previewDest             string
	previewMediaType        string
	previewConflictStrategy string
	previewCreateNFO        bool
)

var previewCmd = &cobra.Command{
	Use:   "preview [directory]",
	Short: "Preview file organization without making changes",
	Long: `Preview shows what files would be organized and where they would be moved
without actually performing any operations. This is useful for verifying
the organization plan before executing it.

The preview command is equivalent to running organize with --dry-run flag.`,
	Args: cobra.ExactArgs(1),
	RunE: runPreview,
}

func init() {
	rootCmd.AddCommand(previewCmd)

	previewCmd.Flags().StringVarP(&previewDest, "dest", "d", "", "destination root directory (default from config)")
	previewCmd.Flags().StringVarP(&previewMediaType, "type", "t", "", "filter by media type (movie, tv, music, book)")
	previewCmd.Flags().StringVar(&previewConflictStrategy, "conflict", "skip", "conflict resolution strategy (skip, rename, interactive)")
	previewCmd.Flags().BoolVar(&previewCreateNFO, "create-nfo", false, "preview NFO file creation")
}

func runPreview(cmd *cobra.Command, args []string) error {
	scanPath := args[0]

	// Make path absolute
	absPath, err := filepath.Abs(scanPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Determine destination root
	destRoot, err := getDestinationRoot(previewMediaType, previewDest)
	if err != nil {
		return err
	}

	// Parse media type filter
	mediaTypeFilter, err := parseMediaTypeFilter(previewMediaType)
	if err != nil {
		return err
	}

	log.Info().Str("path", absPath).Str("dest", destRoot).Msg("Starting preview")

	// Create scanner
	s := createScanner()

	// Scan for files
	result, err := s.Scan(absPath)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	if len(result.Files) == 0 {
		fmt.Println("No media files found to organize.")
		return nil
	}

	// Create organizer in dry-run mode
	org := organizer.NewOrganizer(true)
	org.SetCreateNFO(previewCreateNFO)

	// Plan organization
	plans, err := org.PlanOrganization(result.Files, destRoot, mediaTypeFilter)
	if err != nil {
		return fmt.Errorf("failed to plan organization: %w", err)
	}

	if len(plans) == 0 {
		fmt.Println("No files match the criteria for organization.")
		return nil
	}

	// Validate plans
	validationErrors := org.ValidatePlan(plans)
	if len(validationErrors) > 0 {
		fmt.Printf("\n⚠ Warning: %d validation errors found:\n", len(validationErrors))
		for _, err := range validationErrors {
			fmt.Printf("  - %v\n", err)
		}
		fmt.Println()
	}

	// Display preview
	fmt.Printf("\nOrganization Preview\n")
	fmt.Printf("====================\n")
	fmt.Printf("Source: %s\n", absPath)
	fmt.Printf("Destination: %s\n", destRoot)
	if mediaTypeFilter != types.MediaTypeUnknown {
		fmt.Printf("Filter: %s only\n", mediaTypeFilter)
	}
	fmt.Printf("Conflict Strategy: %s\n", previewConflictStrategy)
	fmt.Printf("\nFiles to organize: %d\n\n", len(plans))

	// Group by media type for display
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
		fmt.Printf("\n⚠ Conflicts detected: %d files\n", conflictCount)
	}

	// Display detailed plan if verbose
	if verbose {
		fmt.Println("\nDetailed Plan:")
		fmt.Println("==============")
		for i, plan := range plans {
			fmt.Printf("\n%d. [%s] %s\n", i+1, plan.MediaType, filepath.Base(plan.SourcePath))
			fmt.Printf("   From: %s\n", plan.SourcePath)
			fmt.Printf("   To:   %s\n", plan.DestinationPath)
			if plan.Conflict {
				fmt.Printf("   ⚠ CONFLICT: %s\n", plan.ConflictReason)
				if previewConflictStrategy == "rename" {
					fmt.Printf("   → Will be renamed with suffix\n")
				} else {
					fmt.Printf("   → Will be skipped\n")
				}
			}
		}
	} else {
		fmt.Println("\nUse -v/--verbose to see detailed file-by-file plan")
	}

	fmt.Printf("\nTo execute this plan, run:\n")
	cmdArgs := fmt.Sprintf("  go-jf-org organize %s --dest %s", absPath, destRoot)
	if previewMediaType != "" {
		cmdArgs += fmt.Sprintf(" --type %s", previewMediaType)
	}
	if previewConflictStrategy != "skip" {
		cmdArgs += fmt.Sprintf(" --conflict %s", previewConflictStrategy)
	}
	fmt.Println(cmdArgs)

	return nil
}
