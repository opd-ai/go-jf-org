package verifier

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"

	"github.com/opd-ai/go-jf-org/pkg/types"
)

// Result represents the outcome of a verification operation
type Result struct {
	Path           string
	TotalDirs      int
	CheckedDirs    int
	Violations     []Violation
	ErrorCount     int
	WarningCount   int
	MediaCounts    map[types.MediaType]int
}

// Verifier performs structure verification on Jellyfin media directories
type Verifier struct {
	movieRules *MovieRules
	tvRules    *TVRules
	musicRules *MusicRules
	bookRules  *BookRules
}

// NewVerifier creates a new verifier instance
func NewVerifier() *Verifier {
	return &Verifier{
		movieRules: &MovieRules{},
		tvRules:    &TVRules{},
		musicRules: &MusicRules{},
		bookRules:  &BookRules{},
	}
}

// VerifyPath verifies a directory structure for Jellyfin compatibility
// mediaType can be specified to verify only specific media types, or empty for all
func (v *Verifier) VerifyPath(rootPath string, mediaType types.MediaType) (*Result, error) {
	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if path exists
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("cannot access path: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", absPath)
	}

	result := &Result{
		Path:        absPath,
		Violations:  []Violation{},
		MediaCounts: make(map[types.MediaType]int),
	}

	log.Info().Str("path", absPath).Msg("Starting verification")

	// If mediaType is specified, verify based on type
	// Otherwise, scan all subdirectories
	if mediaType != "" {
		violations := v.verifyByType(absPath, mediaType)
		result.Violations = append(result.Violations, violations...)
		result.CheckedDirs = 1
	} else {
		// Read top-level directories and infer media type
		violations, checked := v.verifyAllTypes(absPath)
		result.Violations = append(result.Violations, violations...)
		result.CheckedDirs = checked
	}

	// Count violations by severity
	for _, violation := range result.Violations {
		if violation.Severity == SeverityError {
			result.ErrorCount++
		} else {
			result.WarningCount++
		}
		result.MediaCounts[violation.MediaType]++
	}

	result.TotalDirs = result.CheckedDirs

	log.Info().
		Int("checked", result.CheckedDirs).
		Int("errors", result.ErrorCount).
		Int("warnings", result.WarningCount).
		Msg("Verification complete")

	return result, nil
}

// verifyByType verifies a directory as a specific media type
func (v *Verifier) verifyByType(path string, mediaType types.MediaType) []Violation {
	switch mediaType {
	case types.MediaTypeMovie:
		return v.movieRules.VerifyMovie(path)
	case types.MediaTypeTV:
		return v.tvRules.VerifyTVShow(path)
	case types.MediaTypeMusic:
		return v.musicRules.VerifyMusic(path)
	case types.MediaTypeBook:
		return v.bookRules.VerifyBook(path)
	default:
		return []Violation{{
			Severity:   SeverityError,
			Path:       path,
			Message:    fmt.Sprintf("Unknown media type: %s", mediaType),
			Suggestion: "Use movie, tv, music, or book",
		}}
	}
}

// verifyAllTypes scans a root directory and verifies subdirectories
func (v *Verifier) verifyAllTypes(rootPath string) ([]Violation, int) {
	violations := []Violation{}
	checked := 0

	entries, err := os.ReadDir(rootPath)
	if err != nil {
		violations = append(violations, Violation{
			Severity:   SeverityError,
			Path:       rootPath,
			Message:    fmt.Sprintf("Cannot read directory: %v", err),
			Suggestion: "Check directory permissions",
		})
		return violations, 0
	}

	// Iterate through top-level directories
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirPath := filepath.Join(rootPath, entry.Name())
		dirName := entry.Name()

		// Infer media type based on directory structure
		mediaType := v.inferMediaType(dirPath, dirName)

		if mediaType != "" {
			log.Debug().Str("path", dirPath).Str("type", string(mediaType)).Msg("Verifying directory")
			dirViolations := v.verifyByType(dirPath, mediaType)
			violations = append(violations, dirViolations...)
			checked++
		} else {
			// Unknown structure - warning
			violations = append(violations, Violation{
				Severity:   SeverityWarning,
				Path:       dirPath,
				Message:    fmt.Sprintf("Cannot determine media type for directory: %s", dirName),
				Suggestion: "Ensure directory follows Jellyfin naming conventions",
			})
		}
	}

	return violations, checked
}

// inferMediaType attempts to determine media type from directory structure
func (v *Verifier) inferMediaType(dirPath, dirName string) types.MediaType {
	// Check for common patterns
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return ""
	}

	// Check for TV show structure (Season folders)
	for _, entry := range entries {
		if entry.IsDir() {
			subDirName := entry.Name()
			// "Season ##" pattern indicates TV show
			if len(subDirName) >= 8 && subDirName[:6] == "Season" {
				return types.MediaTypeTV
			}
		}
	}

	// Check for movie pattern: "Movie Name (Year)"
	// Movies typically have video files directly in the folder
	hasVideoFile := false
	hasAudioFile := false
	hasBookFile := false

	videoExts := map[string]bool{".mkv": true, ".mp4": true, ".avi": true, ".m4v": true, ".ts": true, ".webm": true}
	audioExts := map[string]bool{".flac": true, ".mp3": true, ".m4a": true, ".ogg": true, ".opus": true, ".wav": true}
	bookExts := map[string]bool{".epub": true, ".mobi": true, ".pdf": true, ".azw3": true, ".cbz": true, ".cbr": true}

	for _, entry := range entries {
		if entry.IsDir() {
			// Subdirectories could indicate music (albums) or books
			continue
		}
		ext := filepath.Ext(entry.Name())
		if videoExts[ext] {
			hasVideoFile = true
		}
		if audioExts[ext] {
			hasAudioFile = true
		}
		if bookExts[ext] {
			hasBookFile = true
		}
	}

	// Determine type based on content
	if hasVideoFile {
		// Could be movie or TV - check directory name for year pattern
		// "Movie Name (Year)" pattern
		if len(dirName) > 6 && dirName[len(dirName)-5] == '(' && dirName[len(dirName)-1] == ')' {
			return types.MediaTypeMovie
		}
		// Default to movie if no season folders
		return types.MediaTypeMovie
	}

	if hasAudioFile {
		return types.MediaTypeMusic
	}

	if hasBookFile {
		return types.MediaTypeBook
	}

	// Cannot determine
	return ""
}

// IsValid returns true if the result has no errors
func (r *Result) IsValid() bool {
	return r.ErrorCount == 0
}

// HasIssues returns true if there are any violations (errors or warnings)
func (r *Result) HasIssues() bool {
	return len(r.Violations) > 0
}
