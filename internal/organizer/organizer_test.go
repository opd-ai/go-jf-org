package organizer

import (
	"os"
	"path/filepath"
	"testing"

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
	result1 := findAvailableName(basePath)
	expected1 := filepath.Join(tmpDir, "movie-1.mkv")
	if result1 != expected1 {
		t.Errorf("findAvailableName() = %q, want %q", result1, expected1)
	}
	
	// Create -1 file
	createTestFile(t, result1)
	
	// Second call should return -2 suffix
	result2 := findAvailableName(basePath)
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
