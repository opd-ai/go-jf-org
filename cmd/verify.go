package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/opd-ai/go-jf-org/internal/verifier"
	"github.com/opd-ai/go-jf-org/pkg/types"
)

var (
	verifyStrict     bool
	verifyMediaType  string
	verifyJSONOutput bool
)

var verifyCmd = &cobra.Command{
	Use:   "verify [directory]",
	Short: "Verify Jellyfin-compatible directory structure",
	Long: `Verify checks if a directory structure follows Jellyfin naming conventions.

It validates:
- Directory and file naming patterns
- Proper media organization (movies, TV shows, music, books)
- Presence of NFO files (optional but recommended)
- Structural consistency

Use --strict to fail on any violations (exit code 1).
Use --type to verify only specific media types.
Use --json for machine-readable output.`,
	Args: cobra.ExactArgs(1),
	RunE: runVerify,
}

func init() {
	rootCmd.AddCommand(verifyCmd)
	verifyCmd.Flags().BoolVar(&verifyStrict, "strict", false, "Fail with exit code 1 if errors are found")
	verifyCmd.Flags().StringVar(&verifyMediaType, "type", "", "Verify specific media type (movie, tv, music, book)")
	verifyCmd.Flags().BoolVar(&verifyJSONOutput, "json", false, "Output results as JSON")
}

func runVerify(cmd *cobra.Command, args []string) error {
	verifyPath := args[0]

	// Make path absolute
	absPath, err := filepath.Abs(verifyPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	log.Info().Str("path", absPath).Msg("Starting verification")

	// Parse media type if specified
	var mediaType types.MediaType
	if verifyMediaType != "" {
		mediaType = types.MediaType(strings.ToLower(verifyMediaType))
		// Validate media type
		validTypes := map[types.MediaType]bool{
			types.MediaTypeMovie: true,
			types.MediaTypeTV:    true,
			types.MediaTypeMusic: true,
			types.MediaTypeBook:  true,
		}
		if !validTypes[mediaType] {
			return fmt.Errorf("invalid media type: %s (must be movie, tv, music, or book)", verifyMediaType)
		}
	}

	// Create verifier and run verification
	v := verifier.NewVerifier()
	result, err := v.VerifyPath(absPath, mediaType)
	if err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	// Output results
	if verifyJSONOutput {
		return outputJSON(result)
	}

	return outputHuman(result, verifyStrict)
}

// outputJSON outputs results in JSON format
func outputJSON(result *verifier.Result) error {
	output := struct {
		Path         string                       `json:"path"`
		CheckedDirs  int                          `json:"checked_directories"`
		ErrorCount   int                          `json:"error_count"`
		WarningCount int                          `json:"warning_count"`
		MediaCounts  map[types.MediaType]int      `json:"media_counts"`
		Violations   []verifier.Violation         `json:"violations"`
	}{
		Path:         result.Path,
		CheckedDirs:  result.CheckedDirs,
		ErrorCount:   result.ErrorCount,
		WarningCount: result.WarningCount,
		MediaCounts:  result.MediaCounts,
		Violations:   result.Violations,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// outputHuman outputs results in human-readable format
func outputHuman(result *verifier.Result, strict bool) error {
	fmt.Println()
	fmt.Printf("Verification Results for: %s\n", result.Path)
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Directories checked: %d\n", result.CheckedDirs)
	fmt.Printf("Errors:              %d\n", result.ErrorCount)
	fmt.Printf("Warnings:            %d\n", result.WarningCount)
	fmt.Println()

	// Display media type breakdown if available
	if len(result.MediaCounts) > 0 {
		fmt.Println("Issues by media type:")
		for mediaType, count := range result.MediaCounts {
			fmt.Printf("  %s: %d\n", mediaType, count)
		}
		fmt.Println()
	}

	// Display violations
	if len(result.Violations) > 0 {
		fmt.Println("Violations:")
		fmt.Println(strings.Repeat("-", 80))

		// Group by severity
		errors := []verifier.Violation{}
		warnings := []verifier.Violation{}

		for _, v := range result.Violations {
			if v.Severity == verifier.SeverityError {
				errors = append(errors, v)
			} else {
				warnings = append(warnings, v)
			}
		}

		// Display errors first
		if len(errors) > 0 {
			fmt.Println("\nERRORS:")
			for i, v := range errors {
				displayViolation(i+1, v)
			}
		}

		// Display warnings
		if len(warnings) > 0 {
			fmt.Println("\nWARNINGS:")
			for i, v := range warnings {
				displayViolation(i+1, v)
			}
		}

		fmt.Println()
	}

	// Summary
	if result.IsValid() {
		fmt.Println("✓ Structure is valid! No errors found.")
		if result.WarningCount > 0 {
			fmt.Printf("  Note: %d warning(s) detected. These are optional improvements.\n", result.WarningCount)
		}
		return nil
	}

	fmt.Printf("✗ Structure has %d error(s) that should be fixed.\n", result.ErrorCount)

	// Return error in strict mode for errors (not warnings)
	if strict {
		return fmt.Errorf("verification failed with %d error(s)", result.ErrorCount)
	}

	return nil
}

// displayViolation displays a single violation in formatted output
func displayViolation(num int, v verifier.Violation) {
	// Shorten path for display
	displayPath := v.Path
	if len(displayPath) > 70 {
		displayPath = "..." + displayPath[len(displayPath)-67:]
	}

	fmt.Printf("\n%d. [%s] %s\n", num, v.MediaType, displayPath)
	fmt.Printf("   Issue:      %s\n", v.Message)
	if v.Suggestion != "" {
		fmt.Printf("   Suggestion: %s\n", v.Suggestion)
	}
}
