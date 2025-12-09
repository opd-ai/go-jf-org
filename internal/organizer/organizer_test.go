package organizer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/opd-ai/go-jf-org/internal/artwork"
	"github.com/opd-ai/go-jf-org/pkg/types"
)

func TestPlanOrganization(t *testing.T) {
	// Create temp directory with test files
	tmpDir := t.TempDir()
	
	// Create test files
	movieFile := filepath.Join(tmpDir, "The.Matrix.1999.1080p.mkv")
	tvFile := filepath.Join(tmpDir, "Breaking.Bad.S01E01.mkv")
	unknownFile := filepath.Join(tmpDir, "unknown.txt")
	
	createTestFile(t, movieFile)
	createTestFile(t, tvFile)
	createTestFile(t, unknownFile)
	
	files := []string{movieFile, tvFile, unknownFile}
	destRoot := filepath.Join(tmpDir, "organized")

	o := NewOrganizer(true)

	tests := []struct {
		name       string
		filter     types.MediaType
		wantCount  int
		checkTypes func([]Plan) bool
	}{
		{
			name:      "no filter",
			filter:    types.MediaTypeUnknown,
			wantCount: 2, // movie and TV, unknown filtered out
			checkTypes: func(plans []Plan) bool {
				hasMovie := false
				hasTV := false
				for _, p := range plans {
					if p.MediaType == types.MediaTypeMovie {
						hasMovie = true
					}
					if p.MediaType == types.MediaTypeTV {
						hasTV = true
					}
				}
				return hasMovie && hasTV
			},
		},
		{
			name:      "movie filter",
			filter:    types.MediaTypeMovie,
			wantCount: 1,
			checkTypes: func(plans []Plan) bool {
				return len(plans) == 1 && plans[0].MediaType == types.MediaTypeMovie
			},
		},
		{
			name:      "tv filter",
			filter:    types.MediaTypeTV,
			wantCount: 1,
			checkTypes: func(plans []Plan) bool {
				return len(plans) == 1 && plans[0].MediaType == types.MediaTypeTV
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plans, err := o.PlanOrganization(files, destRoot, tt.filter)
			if err != nil {
				t.Fatalf("PlanOrganization() error = %v", err)
			}

			if len(plans) != tt.wantCount {
				t.Errorf("PlanOrganization() got %d plans, want %d", len(plans), tt.wantCount)
			}

			if !tt.checkTypes(plans) {
				t.Errorf("PlanOrganization() incorrect media types in plans")
			}
		})
	}
}

func TestPlanOrganization_ConflictDetection(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create test file
	sourceFile := filepath.Join(tmpDir, "The.Matrix.1999.1080p.mkv")
	createTestFile(t, sourceFile)
	
	destRoot := filepath.Join(tmpDir, "organized")
	
	// Create conflicting destination file
	destPath := filepath.Join(destRoot, "The Matrix (1999)", "The Matrix (1999).mkv")
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		t.Fatal(err)
	}
	createTestFile(t, destPath)

	o := NewOrganizer(true)
	plans, err := o.PlanOrganization([]string{sourceFile}, destRoot, types.MediaTypeUnknown)
	if err != nil {
		t.Fatalf("PlanOrganization() error = %v", err)
	}

	if len(plans) != 1 {
		t.Fatalf("Expected 1 plan, got %d", len(plans))
	}

	if !plans[0].Conflict {
		t.Errorf("Expected conflict to be detected")
	}

	if plans[0].ConflictReason == "" {
		t.Errorf("Expected conflict reason to be set")
	}
}

func TestExecute_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	
	sourceFile := filepath.Join(tmpDir, "The.Matrix.1999.1080p.mkv")
	createTestFile(t, sourceFile)
	
	destRoot := filepath.Join(tmpDir, "organized")
	destPath := filepath.Join(destRoot, "The Matrix (1999)", "The Matrix (1999).mkv")

	plan := Plan{
		SourcePath:      sourceFile,
		DestinationPath: destPath,
		MediaType:       types.MediaTypeMovie,
		Operation:       types.OperationMove,
	}

	o := NewOrganizer(true) // dry run mode
	ops, err := o.Execute([]Plan{plan}, "skip")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(ops) != 1 {
		t.Fatalf("Expected 1 operation, got %d", len(ops))
	}

	// In dry run, file should not be moved
	if _, err := os.Stat(sourceFile); os.IsNotExist(err) {
		t.Errorf("Source file was moved in dry-run mode")
	}

	if _, err := os.Stat(destPath); err == nil {
		t.Errorf("Destination file was created in dry-run mode")
	}

	if ops[0].Status != types.OperationStatusCompleted {
		t.Errorf("Expected operation status to be completed, got %s", ops[0].Status)
	}
}

func TestExecute_RealMove(t *testing.T) {
	tmpDir := t.TempDir()
	
	sourceFile := filepath.Join(tmpDir, "The.Matrix.1999.1080p.mkv")
	createTestFile(t, sourceFile)
	
	destRoot := filepath.Join(tmpDir, "organized")
	destPath := filepath.Join(destRoot, "The Matrix (1999)", "The Matrix (1999).mkv")

	plan := Plan{
		SourcePath:      sourceFile,
		DestinationPath: destPath,
		MediaType:       types.MediaTypeMovie,
		Operation:       types.OperationMove,
	}

	o := NewOrganizer(false) // real mode
	ops, err := o.Execute([]Plan{plan}, "skip")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(ops) != 1 {
		t.Fatalf("Expected 1 operation, got %d", len(ops))
	}

	// File should be moved
	if _, err := os.Stat(sourceFile); !os.IsNotExist(err) {
		t.Errorf("Source file still exists after move")
	}

	if _, err := os.Stat(destPath); err != nil {
		t.Errorf("Destination file was not created: %v", err)
	}

	if ops[0].Status != types.OperationStatusCompleted {
		t.Errorf("Expected operation status to be completed, got %s", ops[0].Status)
	}
}

func TestExecute_ConflictSkip(t *testing.T) {
	tmpDir := t.TempDir()
	
	sourceFile := filepath.Join(tmpDir, "The.Matrix.1999.1080p.mkv")
	createTestFile(t, sourceFile)
	
	destRoot := filepath.Join(tmpDir, "organized")
	destPath := filepath.Join(destRoot, "The Matrix (1999)", "The Matrix (1999).mkv")

	plan := Plan{
		SourcePath:      sourceFile,
		DestinationPath: destPath,
		MediaType:       types.MediaTypeMovie,
		Operation:       types.OperationMove,
		Conflict:        true,
		ConflictReason:  "file exists",
	}

	o := NewOrganizer(false)
	ops, err := o.Execute([]Plan{plan}, "skip")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Should have no operations due to skip
	if len(ops) != 0 {
		t.Errorf("Expected 0 operations with skip strategy, got %d", len(ops))
	}

	// Source file should still exist
	if _, err := os.Stat(sourceFile); err != nil {
		t.Errorf("Source file was modified despite skip strategy")
	}
}

func TestExecute_ConflictRename(t *testing.T) {
	tmpDir := t.TempDir()
	
	sourceFile := filepath.Join(tmpDir, "The.Matrix.1999.1080p.mkv")
	createTestFile(t, sourceFile)
	
	destRoot := filepath.Join(tmpDir, "organized")
	destPath := filepath.Join(destRoot, "The Matrix (1999)", "The Matrix (1999).mkv")
	
	// Create conflicting file
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		t.Fatal(err)
	}
	createTestFile(t, destPath)

	plan := Plan{
		SourcePath:      sourceFile,
		DestinationPath: destPath,
		MediaType:       types.MediaTypeMovie,
		Operation:       types.OperationMove,
		Conflict:        true,
		ConflictReason:  "file exists",
	}

	o := NewOrganizer(false)
	ops, err := o.Execute([]Plan{plan}, "rename")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(ops) != 1 {
		t.Fatalf("Expected 1 operation, got %d", len(ops))
	}

	// Check that file was renamed
	if ops[0].Destination == destPath {
		t.Errorf("Expected destination to be renamed, but it wasn't")
	}

	// New destination should exist
	if _, err := os.Stat(ops[0].Destination); err != nil {
		t.Errorf("Renamed destination file was not created: %v", err)
	}

	// Original conflict file should still exist
	if _, err := os.Stat(destPath); err != nil {
		t.Errorf("Original file was overwritten despite rename strategy")
	}
}

func TestFindAvailableName(t *testing.T) {
	tmpDir := t.TempDir()
	
	basePath := filepath.Join(tmpDir, "movie.mkv")
	
	// First call should return -1 suffix
	result1, err := findAvailableName(basePath)
	if err != nil {
		t.Fatalf("findAvailableName() error = %v", err)
	}
	expected1 := filepath.Join(tmpDir, "movie-1.mkv")
	if result1 != expected1 {
		t.Errorf("findAvailableName() = %q, want %q", result1, expected1)
	}
	
	// Create -1 file
	createTestFile(t, result1)
	
	// Second call should return -2 suffix
	result2, err := findAvailableName(basePath)
	if err != nil {
		t.Fatalf("findAvailableName() error = %v", err)
	}
	expected2 := filepath.Join(tmpDir, "movie-2.mkv")
	if result2 != expected2 {
		t.Errorf("findAvailableName() = %q, want %q", result2, expected2)
	}
}

func TestValidatePlan(t *testing.T) {
	tmpDir := t.TempDir()
	
	existingFile := filepath.Join(tmpDir, "existing.mkv")
	createTestFile(t, existingFile)
	
	nonExistentFile := filepath.Join(tmpDir, "nonexistent.mkv")

	o := NewOrganizer(false)

	tests := []struct {
		name      string
		plans     []Plan
		wantError bool
	}{
		{
			name: "valid plan",
			plans: []Plan{
				{
					SourcePath:      existingFile,
					DestinationPath: filepath.Join(tmpDir, "dest", "file.mkv"),
				},
			},
			wantError: false,
		},
		{
			name: "nonexistent source",
			plans: []Plan{
				{
					SourcePath:      nonExistentFile,
					DestinationPath: filepath.Join(tmpDir, "dest", "file.mkv"),
				},
			},
			wantError: true,
		},
		{
			name: "source is directory",
			plans: []Plan{
				{
					SourcePath:      tmpDir,
					DestinationPath: filepath.Join(tmpDir, "dest", "file.mkv"),
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := o.ValidatePlan(tt.plans)
			hasError := len(errs) > 0
			
			if hasError != tt.wantError {
				t.Errorf("ValidatePlan() errors = %v, wantError %v", errs, tt.wantError)
			}
		})
	}
}

// Helper function to create test files
func createTestFile(t *testing.T, path string) {
	t.Helper()
	
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	
	// Write some content so file has size
	if _, err := f.WriteString("test content"); err != nil {
		t.Fatal(err)
	}
}

// TestPlanOrganization_NilMetadataHandling validates the defensive nil check (BUG-NIL-001)
// Note: Current parsers always return valid metadata objects, so this test documents
// the defensive programming practice rather than testing an actual code path.
// The nil check protects against future parser changes that might return (nil, error).
func TestPlanOrganization_NilMetadataHandling(t *testing.T) {
	// This test validates that the organizer has defensive nil checks in place.
	// Current implementation: parsers never return nil metadata, but the code
	// now includes a defensive check after error handling to guard against
	// future modifications that might change this behavior.
	
	tmpDir := t.TempDir()
	movieFile := filepath.Join(tmpDir, "The.Matrix.1999.mkv")
	createTestFile(t, movieFile)
	
	destRoot := filepath.Join(tmpDir, "organized")
	o := NewOrganizer(false)
	
	// This should work normally - metadata will be valid
	plans, err := o.PlanOrganization([]string{movieFile}, destRoot, types.MediaTypeUnknown)
	if err != nil {
		t.Fatalf("PlanOrganization() error = %v", err)
	}
	
	// Should successfully create a plan
	if len(plans) != 1 {
		t.Errorf("expected 1 plan, got %d", len(plans))
	}
	
	// The defensive nil check is in place at organizer.go:99-103
	// If parsers are modified to return (nil, nil), the code will handle it gracefully
}

func TestSetDownloadArtwork(t *testing.T) {
	tests := []struct {
		name     string
		download bool
		size     artwork.ImageSize
	}{
		{
			name:     "enable with small size",
			download: true,
			size:     artwork.SizeSmall,
		},
		{
			name:     "enable with medium size",
			download: true,
			size:     artwork.SizeMedium,
		},
		{
			name:     "enable with large size",
			download: true,
			size:     artwork.SizeLarge,
		},
		{
			name:     "enable with original size",
			download: true,
			size:     artwork.SizeOriginal,
		},
		{
			name:     "disable artwork",
			download: false,
			size:     artwork.SizeMedium,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOrganizer(false)
			
			// Call SetDownloadArtwork
			o.SetDownloadArtwork(tt.download, tt.size)
			
			// Verify the fields are set correctly
			if o.downloadArtwork != tt.download {
				t.Errorf("downloadArtwork = %v, want %v", o.downloadArtwork, tt.download)
			}
			
			if o.artworkSize != tt.size {
				t.Errorf("artworkSize = %v, want %v", o.artworkSize, tt.size)
			}
		})
	}
}

func TestDownloadArtworkForPlan_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	o := NewOrganizer(true) // Dry run mode
	o.SetDownloadArtwork(true, artwork.SizeMedium)
	
	tests := []struct {
		name      string
		mediaType types.MediaType
		metadata  *types.Metadata
		wantOps   int
	}{
		{
			name:      "movie with poster and backdrop",
			mediaType: types.MediaTypeMovie,
			metadata: &types.Metadata{
				Title: "Test Movie",
				Year:  2020,
				MovieMetadata: &types.MovieMetadata{
					PosterURL:   "/poster.jpg",
					BackdropURL: "/backdrop.jpg",
				},
			},
			wantOps: 2, // poster + backdrop
		},
		{
			name:      "movie with only poster",
			mediaType: types.MediaTypeMovie,
			metadata: &types.Metadata{
				Title: "Test Movie",
				Year:  2020,
				MovieMetadata: &types.MovieMetadata{
					PosterURL: "/poster.jpg",
				},
			},
			wantOps: 1, // poster only
		},
		{
			name:      "TV show with poster",
			mediaType: types.MediaTypeTV,
			metadata: &types.Metadata{
				Title: "Test Show",
				Year:  2020,
				TVMetadata: &types.TVMetadata{
					ShowTitle: "Test Show",
					Season:    1,
					Episode:   1,
					PosterURL: "/poster.jpg",
				},
			},
			wantOps: 1, // show poster
		},
		{
			name:      "music with cover",
			mediaType: types.MediaTypeMusic,
			metadata: &types.Metadata{
				Title: "Test Album",
				Year:  2020,
				MusicMetadata: &types.MusicMetadata{
					Artist:         "Test Artist",
					Album:          "Test Album",
					MusicBrainzRID: "test-release-id",
				},
			},
			wantOps: 1, // album cover
		},
		{
			name:      "book with cover",
			mediaType: types.MediaTypeBook,
			metadata: &types.Metadata{
				Title: "Test Book",
				Year:  2020,
				BookMetadata: &types.BookMetadata{
					Author: "Test Author",
					ISBN:   "1234567890",
				},
			},
			wantOps: 1, // book cover
		},
		{
			name:      "no metadata",
			mediaType: types.MediaTypeMovie,
			metadata:  nil,
			wantOps:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			destPath := filepath.Join(tmpDir, "test.mkv")
			if tt.mediaType == types.MediaTypeTV {
				destPath = filepath.Join(tmpDir, "Show", "Season 01", "test.mkv")
			}
			
			plan := Plan{
				SourcePath:      filepath.Join(tmpDir, "source.mkv"),
				DestinationPath: destPath,
				MediaType:       tt.mediaType,
				Metadata:        tt.metadata,
				Operation:       types.OperationMove,
			}

			ops, err := o.downloadArtworkForPlan(nil, plan)
			if err != nil {
				t.Fatalf("downloadArtworkForPlan() error = %v", err)
			}

			if len(ops) != tt.wantOps {
				t.Errorf("downloadArtworkForPlan() got %d operations, want %d", len(ops), tt.wantOps)
			}

			// All operations should be completed in dry-run mode
			for i, op := range ops {
				if op.Status != types.OperationStatusCompleted {
					t.Errorf("operation %d status = %v, want %v", i, op.Status, types.OperationStatusCompleted)
				}
				if op.Type != types.OperationCreateFile {
					t.Errorf("operation %d type = %v, want %v", i, op.Type, types.OperationCreateFile)
				}
			}
		})
	}
}

func TestDownloadArtworkForPlan_Disabled(t *testing.T) {
	tmpDir := t.TempDir()
	o := NewOrganizer(false)
	// Don't enable artwork downloads
	
	plan := Plan{
		SourcePath:      filepath.Join(tmpDir, "source.mkv"),
		DestinationPath: filepath.Join(tmpDir, "dest.mkv"),
		MediaType:       types.MediaTypeMovie,
		Metadata: &types.Metadata{
			Title: "Test Movie",
			Year:  2020,
			MovieMetadata: &types.MovieMetadata{
				PosterURL: "/poster.jpg",
			},
		},
		Operation: types.OperationMove,
	}

	ops, err := o.downloadArtworkForPlan(nil, plan)
	if err != nil {
		t.Fatalf("downloadArtworkForPlan() error = %v", err)
	}

	// Should return no operations when artwork download is disabled
	if len(ops) != 0 {
		t.Errorf("downloadArtworkForPlan() got %d operations, want 0", len(ops))
	}
}

