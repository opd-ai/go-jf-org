package verifier

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/opd-ai/go-jf-org/pkg/types"
)

// Severity represents the severity level of a violation
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

// Violation represents a single verification rule violation
type Violation struct {
	Severity   Severity
	Path       string
	Message    string
	Suggestion string
	MediaType  types.MediaType
}

// MovieRules contains verification rules for movie directories
type MovieRules struct{}

// VerifyMovie checks if a movie directory follows Jellyfin conventions
func (r *MovieRules) VerifyMovie(dirPath string) []Violation {
	violations := []Violation{}
	
	// Extract directory name
	dirName := filepath.Base(dirPath)
	
	// Check directory naming: "Movie Name (Year)"
	moviePattern := regexp.MustCompile(`^(.+?)\s+\((\d{4})\)$`)
	if !moviePattern.MatchString(dirName) {
		violations = append(violations, Violation{
			Severity:   SeverityError,
			Path:       dirPath,
			MediaType:  types.MediaTypeMovie,
			Message:    fmt.Sprintf("Directory name does not match Jellyfin convention: %s", dirName),
			Suggestion: "Rename to format: 'Movie Name (YYYY)'",
		})
		return violations
	}
	
	// Extract expected movie name from directory
	expectedName := dirName // Full "Movie Name (Year)"
	
	// Check for video files
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		violations = append(violations, Violation{
			Severity:   SeverityError,
			Path:       dirPath,
			MediaType:  types.MediaTypeMovie,
			Message:    fmt.Sprintf("Cannot read directory: %v", err),
			Suggestion: "Check directory permissions",
		})
		return violations
	}
	
	videoExtensions := map[string]bool{
		".mkv": true, ".mp4": true, ".avi": true,
		".m4v": true, ".ts": true, ".webm": true,
	}
	
	var videoFiles []string
	var hasNFO bool
	
	for _, entry := range entries {
		if entry.IsDir() {
			// Subdirectories are not expected in movie folders
			violations = append(violations, Violation{
				Severity:   SeverityWarning,
				Path:       filepath.Join(dirPath, entry.Name()),
				MediaType:  types.MediaTypeMovie,
				Message:    "Unexpected subdirectory in movie folder",
				Suggestion: "Movies should have a flat structure",
			})
			continue
		}
		
		fileName := entry.Name()
		ext := strings.ToLower(filepath.Ext(fileName))
		
		if videoExtensions[ext] {
			videoFiles = append(videoFiles, fileName)
			
			// Check if video file follows naming convention
			nameWithoutExt := strings.TrimSuffix(fileName, ext)
			// Allow optional quality/version suffixes: "Movie Name (Year) - 1080p.mkv"
			if !strings.HasPrefix(nameWithoutExt, expectedName) {
				violations = append(violations, Violation{
					Severity:   SeverityWarning,
					Path:       filepath.Join(dirPath, fileName),
					MediaType:  types.MediaTypeMovie,
					Message:    fmt.Sprintf("Video file name doesn't match directory: %s", fileName),
					Suggestion: fmt.Sprintf("Rename to: %s%s", expectedName, ext),
				})
			}
		} else if strings.ToLower(fileName) == "movie.nfo" {
			hasNFO = true
		}
	}
	
	// Check for at least one video file
	if len(videoFiles) == 0 {
		violations = append(violations, Violation{
			Severity:   SeverityError,
			Path:       dirPath,
			MediaType:  types.MediaTypeMovie,
			Message:    "No video files found in movie directory",
			Suggestion: "Add a video file or remove empty directory",
		})
	}
	
	// NFO is optional but recommended
	if !hasNFO && len(videoFiles) > 0 {
		violations = append(violations, Violation{
			Severity:   SeverityWarning,
			Path:       dirPath,
			MediaType:  types.MediaTypeMovie,
			Message:    "Missing movie.nfo file",
			Suggestion: "Generate NFO file with: go-jf-org organize --create-nfo",
		})
	}
	
	return violations
}

// TVRules contains verification rules for TV show directories
type TVRules struct{}

// VerifyTVShow checks if a TV show directory follows Jellyfin conventions
func (r *TVRules) VerifyTVShow(showPath string) []Violation {
	violations := []Violation{}
	
	showName := filepath.Base(showPath)
	
	// Read show directory
	entries, err := os.ReadDir(showPath)
	if err != nil {
		violations = append(violations, Violation{
			Severity:   SeverityError,
			Path:       showPath,
			MediaType:  types.MediaTypeTV,
			Message:    fmt.Sprintf("Cannot read directory: %v", err),
			Suggestion: "Check directory permissions",
		})
		return violations
	}
	
	seasonPattern := regexp.MustCompile(`^Season\s+(\d{2})$`)
	var seasonDirs []string
	var hasShowNFO bool
	
	for _, entry := range entries {
		if entry.IsDir() {
			dirName := entry.Name()
			if seasonPattern.MatchString(dirName) || dirName == "Specials" {
				seasonDirs = append(seasonDirs, dirName)
				// Verify season directory
				seasonViolations := r.verifySeason(filepath.Join(showPath, dirName), showName)
				violations = append(violations, seasonViolations...)
			} else {
				violations = append(violations, Violation{
					Severity:   SeverityWarning,
					Path:       filepath.Join(showPath, dirName),
					MediaType:  types.MediaTypeTV,
					Message:    fmt.Sprintf("Unexpected directory: %s", dirName),
					Suggestion: "TV show directories should contain 'Season ##' folders",
				})
			}
		} else if strings.ToLower(entry.Name()) == "tvshow.nfo" {
			hasShowNFO = true
		}
	}
	
	// Check for at least one season
	if len(seasonDirs) == 0 {
		violations = append(violations, Violation{
			Severity:   SeverityError,
			Path:       showPath,
			MediaType:  types.MediaTypeTV,
			Message:    "No season directories found",
			Suggestion: "Create directories named 'Season 01', 'Season 02', etc.",
		})
	}
	
	// NFO is optional but recommended
	if !hasShowNFO {
		violations = append(violations, Violation{
			Severity:   SeverityWarning,
			Path:       showPath,
			MediaType:  types.MediaTypeTV,
			Message:    "Missing tvshow.nfo file",
			Suggestion: "Generate NFO file with: go-jf-org organize --create-nfo",
		})
	}
	
	return violations
}

// verifySeason checks a single season directory
func (r *TVRules) verifySeason(seasonPath, showName string) []Violation {
	violations := []Violation{}
	
	seasonDir := filepath.Base(seasonPath)
	
	entries, err := os.ReadDir(seasonPath)
	if err != nil {
		violations = append(violations, Violation{
			Severity:   SeverityError,
			Path:       seasonPath,
			MediaType:  types.MediaTypeTV,
			Message:    fmt.Sprintf("Cannot read season directory: %v", err),
			Suggestion: "Check directory permissions",
		})
		return violations
	}
	
	videoExtensions := map[string]bool{
		".mkv": true, ".mp4": true, ".avi": true,
		".m4v": true, ".ts": true, ".webm": true,
	}
	
	// Expected pattern: "Show Name - S##E## - Episode Title.ext"
	episodePattern := regexp.MustCompile(`^(.+?)\s+-\s+S(\d{2})E(\d{2})(?:\s+-\s+(.+?))?(?:\s+-\s+\d{3,4}p)?\.(.+)$`)
	
	var videoFiles []string
	var hasSeasonNFO bool
	
	for _, entry := range entries {
		if entry.IsDir() {
			violations = append(violations, Violation{
				Severity:   SeverityWarning,
				Path:       filepath.Join(seasonPath, entry.Name()),
				MediaType:  types.MediaTypeTV,
				Message:    "Unexpected subdirectory in season folder",
				Suggestion: "Episode files should be placed directly in season folders",
			})
			continue
		}
		
		fileName := entry.Name()
		ext := strings.ToLower(filepath.Ext(fileName))
		
		if videoExtensions[ext] {
			videoFiles = append(videoFiles, fileName)
			
			// Verify episode naming
			if !episodePattern.MatchString(fileName) {
				violations = append(violations, Violation{
					Severity:   SeverityWarning,
					Path:       filepath.Join(seasonPath, fileName),
					MediaType:  types.MediaTypeTV,
					Message:    fmt.Sprintf("Episode file doesn't match naming convention: %s", fileName),
					Suggestion: fmt.Sprintf("Rename to format: '%s - S##E## - Episode Title%s'", showName, ext),
				})
			}
		} else if strings.ToLower(fileName) == "season.nfo" {
			hasSeasonNFO = true
		}
	}
	
	if len(videoFiles) == 0 {
		violations = append(violations, Violation{
			Severity:   SeverityWarning,
			Path:       seasonPath,
			MediaType:  types.MediaTypeTV,
			Message:    fmt.Sprintf("No episode files found in %s", seasonDir),
			Suggestion: "Add episode files or remove empty season directory",
		})
	}
	
	// Season NFO is optional
	if !hasSeasonNFO && len(videoFiles) > 0 {
		violations = append(violations, Violation{
			Severity:   SeverityWarning,
			Path:       seasonPath,
			MediaType:  types.MediaTypeTV,
			Message:    "Missing season.nfo file",
			Suggestion: "Generate NFO file with: go-jf-org organize --create-nfo",
		})
	}
	
	return violations
}

// MusicRules contains verification rules for music directories
type MusicRules struct{}

// VerifyMusic checks if a music directory follows Jellyfin conventions
func (r *MusicRules) VerifyMusic(artistPath string) []Violation {
	violations := []Violation{}
	
	// Read artist directory
	entries, err := os.ReadDir(artistPath)
	if err != nil {
		violations = append(violations, Violation{
			Severity:   SeverityError,
			Path:       artistPath,
			MediaType:  types.MediaTypeMusic,
			Message:    fmt.Sprintf("Cannot read directory: %v", err),
			Suggestion: "Check directory permissions",
		})
		return violations
	}
	
	// Expected: "Album Name (Year)" directories
	albumPattern := regexp.MustCompile(`^(.+?)\s+\((\d{4})\)$`)
	var albumDirs []string
	
	for _, entry := range entries {
		if entry.IsDir() {
			dirName := entry.Name()
			if albumPattern.MatchString(dirName) {
				albumDirs = append(albumDirs, dirName)
				// Could verify album structure, but keeping it simple for now
			} else {
				violations = append(violations, Violation{
					Severity:   SeverityWarning,
					Path:       filepath.Join(artistPath, dirName),
					MediaType:  types.MediaTypeMusic,
					Message:    fmt.Sprintf("Album directory doesn't match convention: %s", dirName),
					Suggestion: "Rename to format: 'Album Name (YYYY)'",
				})
			}
		}
	}
	
	if len(albumDirs) == 0 {
		violations = append(violations, Violation{
			Severity:   SeverityWarning,
			Path:       artistPath,
			MediaType:  types.MediaTypeMusic,
			Message:    "No album directories found",
			Suggestion: "Create directories named 'Album Name (YYYY)'",
		})
	}
	
	return violations
}

// BookRules contains verification rules for book directories
type BookRules struct{}

// VerifyBook checks if a book directory follows Jellyfin conventions
func (r *BookRules) VerifyBook(authorPath string) []Violation {
	violations := []Violation{}
	
	// Read author directory
	entries, err := os.ReadDir(authorPath)
	if err != nil {
		violations = append(violations, Violation{
			Severity:   SeverityError,
			Path:       authorPath,
			MediaType:  types.MediaTypeBook,
			Message:    fmt.Sprintf("Cannot read directory: %v", err),
			Suggestion: "Check directory permissions",
		})
		return violations
	}
	
	bookExtensions := map[string]bool{
		".epub": true, ".mobi": true, ".pdf": true,
		".azw3": true, ".cbz": true, ".cbr": true,
	}
	
	// Expected: "Book Title (Year)" directories or direct book files
	bookPattern := regexp.MustCompile(`^(.+?)\s+\((\d{4})\)$`)
	var bookDirs []string
	var bookFiles []string
	
	for _, entry := range entries {
		if entry.IsDir() {
			dirName := entry.Name()
			if bookPattern.MatchString(dirName) {
				bookDirs = append(bookDirs, dirName)
			} else {
				violations = append(violations, Violation{
					Severity:   SeverityWarning,
					Path:       filepath.Join(authorPath, dirName),
					MediaType:  types.MediaTypeBook,
					Message:    fmt.Sprintf("Book directory doesn't match convention: %s", dirName),
					Suggestion: "Rename to format: 'Book Title (YYYY)'",
				})
			}
		} else {
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if bookExtensions[ext] {
				bookFiles = append(bookFiles, entry.Name())
			}
		}
	}
	
	if len(bookDirs) == 0 && len(bookFiles) == 0 {
		violations = append(violations, Violation{
			Severity:   SeverityWarning,
			Path:       authorPath,
			MediaType:  types.MediaTypeBook,
			Message:    "No book files or directories found",
			Suggestion: "Add book files in directories named 'Book Title (YYYY)'",
		})
	}
	
	return violations
}
