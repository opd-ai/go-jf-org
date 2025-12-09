package verifier

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/opd-ai/go-jf-org/pkg/types"
)

// TestMovieRules_VerifyMovie tests movie directory verification
func TestMovieRules_VerifyMovie(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(string) error
		expectedErrors int
		expectedWarns  int
	}{
		{
			name: "valid movie directory",
			setupFunc: func(dir string) error {
				movieDir := filepath.Join(dir, "The Matrix (1999)")
				if err := os.Mkdir(movieDir, 0755); err != nil {
					return err
				}
				// Create video file
				videoFile := filepath.Join(movieDir, "The Matrix (1999).mkv")
				return os.WriteFile(videoFile, []byte("fake video"), 0644)
			},
			expectedErrors: 0,
			expectedWarns:  1, // Missing NFO
		},
		{
			name: "valid movie with NFO",
			setupFunc: func(dir string) error {
				movieDir := filepath.Join(dir, "Inception (2010)")
				if err := os.Mkdir(movieDir, 0755); err != nil {
					return err
				}
				videoFile := filepath.Join(movieDir, "Inception (2010).mp4")
				if err := os.WriteFile(videoFile, []byte("fake video"), 0644); err != nil {
					return err
				}
				nfoFile := filepath.Join(movieDir, "movie.nfo")
				return os.WriteFile(nfoFile, []byte("<movie></movie>"), 0644)
			},
			expectedErrors: 0,
			expectedWarns:  0,
		},
		{
			name: "invalid directory name",
			setupFunc: func(dir string) error {
				movieDir := filepath.Join(dir, "InvalidMovieName")
				return os.Mkdir(movieDir, 0755)
			},
			expectedErrors: 1,
			expectedWarns:  0,
		},
		{
			name: "no video files",
			setupFunc: func(dir string) error {
				movieDir := filepath.Join(dir, "Empty Movie (2023)")
				return os.Mkdir(movieDir, 0755)
			},
			expectedErrors: 1,
			expectedWarns:  0,
		},
		{
			name: "wrong video filename",
			setupFunc: func(dir string) error {
				movieDir := filepath.Join(dir, "Test Movie (2020)")
				if err := os.Mkdir(movieDir, 0755); err != nil {
					return err
				}
				videoFile := filepath.Join(movieDir, "wrong_name.mkv")
				return os.WriteFile(videoFile, []byte("fake video"), 0644)
			},
			expectedErrors: 0,
			expectedWarns:  2, // Wrong filename + missing NFO
		},
		{
			name: "unexpected subdirectory",
			setupFunc: func(dir string) error {
				movieDir := filepath.Join(dir, "Movie With Extras (2021)")
				if err := os.Mkdir(movieDir, 0755); err != nil {
					return err
				}
				videoFile := filepath.Join(movieDir, "Movie With Extras (2021).mkv")
				if err := os.WriteFile(videoFile, []byte("fake video"), 0644); err != nil {
					return err
				}
				extrasDir := filepath.Join(movieDir, "Extras")
				return os.Mkdir(extrasDir, 0755)
			},
			expectedErrors: 0,
			expectedWarns:  2, // Subdirectory + missing NFO
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir := t.TempDir()

			// Setup test structure
			if err := tt.setupFunc(tmpDir); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			// Get the created subdirectory
			entries, err := os.ReadDir(tmpDir)
			if err != nil {
				t.Fatalf("Failed to read temp dir: %v", err)
			}

			if len(entries) == 0 {
				t.Fatal("No subdirectories created")
			}

			moviePath := filepath.Join(tmpDir, entries[0].Name())

			// Run verification
			rules := &MovieRules{}
			violations := rules.VerifyMovie(moviePath)

			// Count errors and warnings
			errorCount := 0
			warnCount := 0
			for _, v := range violations {
				if v.Severity == SeverityError {
					errorCount++
				} else {
					warnCount++
				}
			}

			if errorCount != tt.expectedErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectedErrors, errorCount)
				for _, v := range violations {
					if v.Severity == SeverityError {
						t.Logf("  Error: %s - %s", v.Path, v.Message)
					}
				}
			}

			if warnCount != tt.expectedWarns {
				t.Errorf("Expected %d warnings, got %d", tt.expectedWarns, warnCount)
				for _, v := range violations {
					if v.Severity == SeverityWarning {
						t.Logf("  Warning: %s - %s", v.Path, v.Message)
					}
				}
			}
		})
	}
}

// TestTVRules_VerifyTVShow tests TV show directory verification
func TestTVRules_VerifyTVShow(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(string) error
		expectedErrors int
		expectedWarns  int
	}{
		{
			name: "valid TV show structure",
			setupFunc: func(dir string) error {
				showDir := filepath.Join(dir, "Breaking Bad")
				if err := os.Mkdir(showDir, 0755); err != nil {
					return err
				}
				seasonDir := filepath.Join(showDir, "Season 01")
				if err := os.Mkdir(seasonDir, 0755); err != nil {
					return err
				}
				episodeFile := filepath.Join(seasonDir, "Breaking Bad - S01E01 - Pilot.mkv")
				return os.WriteFile(episodeFile, []byte("fake video"), 0644)
			},
			expectedErrors: 0,
			expectedWarns:  2, // Missing tvshow.nfo, season.nfo
		},
		{
			name: "no season directories",
			setupFunc: func(dir string) error {
				showDir := filepath.Join(dir, "Empty Show")
				return os.Mkdir(showDir, 0755)
			},
			expectedErrors: 1,
			expectedWarns:  1, // Missing tvshow.nfo
		},
		{
			name: "invalid season name",
			setupFunc: func(dir string) error {
				showDir := filepath.Join(dir, "Test Show")
				if err := os.Mkdir(showDir, 0755); err != nil {
					return err
				}
				invalidSeason := filepath.Join(showDir, "Season1")
				return os.Mkdir(invalidSeason, 0755)
			},
			expectedErrors: 1, // No valid season directories found
			expectedWarns:  2, // Invalid season dir + missing tvshow.nfo
		},
		{
			name: "wrong episode filename",
			setupFunc: func(dir string) error {
				showDir := filepath.Join(dir, "Show Name")
				if err := os.Mkdir(showDir, 0755); err != nil {
					return err
				}
				seasonDir := filepath.Join(showDir, "Season 01")
				if err := os.Mkdir(seasonDir, 0755); err != nil {
					return err
				}
				wrongFile := filepath.Join(seasonDir, "wrong_episode_name.mkv")
				return os.WriteFile(wrongFile, []byte("fake video"), 0644)
			},
			expectedErrors: 0,
			expectedWarns:  3, // Wrong filename + missing NFOs
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			if err := tt.setupFunc(tmpDir); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			entries, err := os.ReadDir(tmpDir)
			if err != nil {
				t.Fatalf("Failed to read temp dir: %v", err)
			}

			if len(entries) == 0 {
				t.Fatal("No subdirectories created")
			}

			showPath := filepath.Join(tmpDir, entries[0].Name())

			rules := &TVRules{}
			violations := rules.VerifyTVShow(showPath)

			errorCount := 0
			warnCount := 0
			for _, v := range violations {
				if v.Severity == SeverityError {
					errorCount++
				} else {
					warnCount++
				}
			}

			if errorCount != tt.expectedErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectedErrors, errorCount)
				for _, v := range violations {
					if v.Severity == SeverityError {
						t.Logf("  Error: %s - %s", v.Path, v.Message)
					}
				}
			}

			if warnCount < tt.expectedWarns {
				t.Errorf("Expected at least %d warnings, got %d", tt.expectedWarns, warnCount)
			}
		})
	}
}

// TestMusicRules_VerifyMusic tests music directory verification
func TestMusicRules_VerifyMusic(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(string) error
		expectedErrors int
		expectedWarns  int
	}{
		{
			name: "valid artist structure",
			setupFunc: func(dir string) error {
				artistDir := filepath.Join(dir, "Pink Floyd")
				if err := os.Mkdir(artistDir, 0755); err != nil {
					return err
				}
				albumDir := filepath.Join(artistDir, "Dark Side of the Moon (1973)")
				return os.Mkdir(albumDir, 0755)
			},
			expectedErrors: 0,
			expectedWarns:  0,
		},
		{
			name: "invalid album name",
			setupFunc: func(dir string) error {
				artistDir := filepath.Join(dir, "Artist Name")
				if err := os.Mkdir(artistDir, 0755); err != nil {
					return err
				}
				albumDir := filepath.Join(artistDir, "InvalidAlbumName")
				return os.Mkdir(albumDir, 0755)
			},
			expectedErrors: 0,
			expectedWarns:  2, // Invalid album name + no valid albums found
		},
		{
			name: "no album directories",
			setupFunc: func(dir string) error {
				artistDir := filepath.Join(dir, "Empty Artist")
				return os.Mkdir(artistDir, 0755)
			},
			expectedErrors: 0,
			expectedWarns:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			if err := tt.setupFunc(tmpDir); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			entries, err := os.ReadDir(tmpDir)
			if err != nil {
				t.Fatalf("Failed to read temp dir: %v", err)
			}

			if len(entries) == 0 {
				t.Fatal("No subdirectories created")
			}

			artistPath := filepath.Join(tmpDir, entries[0].Name())

			rules := &MusicRules{}
			violations := rules.VerifyMusic(artistPath)

			errorCount := 0
			warnCount := 0
			for _, v := range violations {
				if v.Severity == SeverityError {
					errorCount++
				} else {
					warnCount++
				}
			}

			if errorCount != tt.expectedErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectedErrors, errorCount)
			}

			if warnCount != tt.expectedWarns {
				t.Errorf("Expected %d warnings, got %d", tt.expectedWarns, warnCount)
			}
		})
	}
}

// TestVerifier_VerifyPath tests the main verifier
func TestVerifier_VerifyPath(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func(string) error
		mediaType    types.MediaType
		expectError  bool
		expectIssues bool
	}{
		{
			name: "verify movie directory",
			setupFunc: func(dir string) error {
				movieDir := filepath.Join(dir, "Test Movie (2023)")
				if err := os.Mkdir(movieDir, 0755); err != nil {
					return err
				}
				videoFile := filepath.Join(movieDir, "Test Movie (2023).mkv")
				return os.WriteFile(videoFile, []byte("fake video"), 0644)
			},
			mediaType:    types.MediaTypeMovie,
			expectError:  false,
			expectIssues: true, // Will have warning about missing NFO
		},
		{
			name: "verify TV show directory",
			setupFunc: func(dir string) error {
				showDir := filepath.Join(dir, "Test Show")
				if err := os.Mkdir(showDir, 0755); err != nil {
					return err
				}
				seasonDir := filepath.Join(showDir, "Season 01")
				if err := os.Mkdir(seasonDir, 0755); err != nil {
					return err
				}
				episodeFile := filepath.Join(seasonDir, "Test Show - S01E01.mkv")
				return os.WriteFile(episodeFile, []byte("fake video"), 0644)
			},
			mediaType:    types.MediaTypeTV,
			expectError:  false,
			expectIssues: true,
		},
		{
			name: "nonexistent path",
			setupFunc: func(dir string) error {
				return nil // Don't create anything
			},
			mediaType:    types.MediaTypeMovie,
			expectError:  true,
			expectIssues: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			if err := tt.setupFunc(tmpDir); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			// Get the path to verify
			var verifyPath string
			if tt.name == "nonexistent path" {
				verifyPath = filepath.Join(tmpDir, "nonexistent")
			} else {
				entries, err := os.ReadDir(tmpDir)
				if err != nil {
					t.Fatalf("Failed to read temp dir: %v", err)
				}
				if len(entries) > 0 {
					verifyPath = filepath.Join(tmpDir, entries[0].Name())
				} else {
					verifyPath = tmpDir
				}
			}

			verifier := NewVerifier()
			result, err := verifier.VerifyPath(verifyPath, tt.mediaType)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.expectIssues && !result.HasIssues() {
				t.Error("Expected issues, got none")
			}

			if !tt.expectIssues && result.HasIssues() {
				t.Errorf("Expected no issues, got %d violations", len(result.Violations))
				for _, v := range result.Violations {
					t.Logf("  %s: %s - %s", v.Severity, v.Path, v.Message)
				}
			}
		})
	}
}

// TestVerifier_InferMediaType tests media type inference
func TestVerifier_InferMediaType(t *testing.T) {
	tests := []struct {
		name         string
		dirName      string
		setupFunc    func(string) error
		expectedType types.MediaType
	}{
		{
			name:    "infer movie from year pattern",
			dirName: "The Matrix (1999)",
			setupFunc: func(dir string) error {
				videoFile := filepath.Join(dir, "The Matrix (1999).mkv")
				return os.WriteFile(videoFile, []byte("fake video"), 0644)
			},
			expectedType: types.MediaTypeMovie,
		},
		{
			name:    "infer TV from season folders",
			dirName: "Breaking Bad",
			setupFunc: func(dir string) error {
				seasonDir := filepath.Join(dir, "Season 01")
				return os.Mkdir(seasonDir, 0755)
			},
			expectedType: types.MediaTypeTV,
		},
		{
			name:    "infer music from audio files",
			dirName: "Pink Floyd",
			setupFunc: func(dir string) error {
				audioFile := filepath.Join(dir, "track.mp3")
				return os.WriteFile(audioFile, []byte("fake audio"), 0644)
			},
			expectedType: types.MediaTypeMusic,
		},
		{
			name:    "infer book from book files",
			dirName: "Stephen King",
			setupFunc: func(dir string) error {
				bookFile := filepath.Join(dir, "book.epub")
				return os.WriteFile(bookFile, []byte("fake book"), 0644)
			},
			expectedType: types.MediaTypeBook,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testDir := filepath.Join(tmpDir, tt.dirName)

			if err := os.Mkdir(testDir, 0755); err != nil {
				t.Fatalf("Failed to create test dir: %v", err)
			}

			if err := tt.setupFunc(testDir); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			verifier := NewVerifier()
			inferredType := verifier.inferMediaType(testDir, tt.dirName)

			if inferredType != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, inferredType)
			}
		})
	}
}
