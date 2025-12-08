package safety

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/opd-ai/go-jf-org/pkg/types"
)

func TestValidateMoveOperation_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	v := NewValidator()

	// Create source file
	sourceFile := filepath.Join(tmpDir, "source.mkv")
	destFile := filepath.Join(tmpDir, "dest", "movie.mkv")
	if err := os.WriteFile(sourceFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(destFile), 0755); err != nil {
		t.Fatalf("Failed to create destination directory: %v", err)
	}

	op := types.Operation{
		Type:        types.OperationMove,
		Source:      sourceFile,
		Destination: destFile,
	}

	err := v.ValidateOperation(op)
	if err != nil {
		t.Errorf("Validation failed for valid operation: %v", err)
	}
}

func TestValidateMoveOperation_NonExistentSource(t *testing.T) {
	tmpDir := t.TempDir()
	v := NewValidator()

	op := types.Operation{
		Type:        types.OperationMove,
		Source:      filepath.Join(tmpDir, "nonexistent.mkv"),
		Destination: filepath.Join(tmpDir, "dest", "movie.mkv"),
	}

	err := v.ValidateOperation(op)
	if err == nil {
		t.Error("Expected validation error for non-existent source")
	} else if _, ok := err.(*ValidationError); !ok {
		t.Errorf("Expected ValidationError, got %T: %v", err, err)
	}
}

func TestValidateMoveOperation_SourceIsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	v := NewValidator()

	sourceDir := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	op := types.Operation{
		Type:        types.OperationMove,
		Source:      sourceDir,
		Destination: filepath.Join(tmpDir, "dest", "movie.mkv"),
	}

	err := v.ValidateOperation(op)
	if err == nil {
		t.Error("Expected validation error for directory source")
	} else if _, ok := err.(*ValidationError); !ok {
		t.Errorf("Expected ValidationError, got %T: %v", err, err)
	}
}

func TestValidateMoveOperation_UnreadableSource(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Permission testing not reliable on Windows")
	}
	tmpDir := t.TempDir()
	v := NewValidator()

	sourceFile := filepath.Join(tmpDir, "source.mkv")
	if err := os.WriteFile(sourceFile, []byte("content"), 0000); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	op := types.Operation{
		Type:        types.OperationMove,
		Source:      sourceFile,
		Destination: filepath.Join(tmpDir, "dest", "movie.mkv"),
	}

	err := v.ValidateOperation(op)
	if err == nil {
		t.Error("Expected validation error for unreadable source")
	} else if _, ok := err.(*ValidationError); !ok {
		t.Errorf("Expected ValidationError, got %T: %v", err, err)
	}

	// Cleanup
	os.Chmod(sourceFile, 0644)
}

func TestValidateCreateDirOperation_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	v := NewValidator()

	newDir := filepath.Join(tmpDir, "newdir")

	op := types.Operation{
		Type:        types.OperationCreateDir,
		Destination: newDir,
	}

	err := v.ValidateOperation(op)
	if err != nil {
		t.Errorf("Validation failed for valid directory creation: %v", err)
	}
}

func TestValidateCreateDirOperation_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	v := NewValidator()

	existingDir := filepath.Join(tmpDir, "existing")
	os.MkdirAll(existingDir, 0755)

	op := types.Operation{
		Type:        types.OperationCreateDir,
		Destination: existingDir,
	}

	err := v.ValidateOperation(op)
	if err == nil {
		t.Error("Expected validation error for existing directory")
	}
}

func TestValidateCreateFileOperation_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	v := NewValidator()

	newFile := filepath.Join(tmpDir, "movie.nfo")

	op := types.Operation{
		Type:        types.OperationCreateFile,
		Destination: newFile,
	}

	err := v.ValidateOperation(op)
	if err != nil {
		t.Errorf("Validation failed for valid file creation: %v", err)
	}
}

func TestValidateCreateFileOperation_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	v := NewValidator()

	existingFile := filepath.Join(tmpDir, "existing.nfo")
	if err := os.WriteFile(existingFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	op := types.Operation{
		Type:        types.OperationCreateFile,
		Destination: existingFile,
	}

	err := v.ValidateOperation(op)
	if err == nil {
		t.Error("Expected validation error for existing file")
	} else if _, ok := err.(*ValidationError); !ok {
		t.Errorf("Expected ValidationError, got %T: %v", err, err)
	}
}

func TestValidatePath(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name      string
		path      string
		shouldErr bool
	}{
		{
			name:      "valid path",
			path:      "/home/user/movies/Movie (2023).mkv",
			shouldErr: false,
		},
		{
			name:      "empty path",
			path:      "",
			shouldErr: true,
		},
		{
			name:      "path with colon",
			path:      "/home/user:movies/file.mkv",
			shouldErr: true,
		},
		{
			name:      "path with asterisk",
			path:      "/home/user/movies/*.mkv",
			shouldErr: true,
		},
		{
			name:      "path with question mark",
			path:      "/home/user/movies/file?.mkv",
			shouldErr: true,
		},
		{
			name:      "path with angle brackets",
			path:      "/home/user/movies/<file>.mkv",
			shouldErr: true,
		},
		{
			name:      "filename with leading dot",
			path:      "/home/user/movies/.hidden.mkv",
			shouldErr: true,
		},
		{
			name:      "filename with trailing dot",
			path:      "/home/user/movies/file.mkv.",
			shouldErr: true,
		},
		{
			name:      "filename with leading space",
			path:      "/home/user/movies/ file.mkv",
			shouldErr: true,
		},
		{
			name:      "filename with trailing space",
			path:      "/home/user/movies/file.mkv ",
			shouldErr: true,
		},
		{
			name:      "valid path with spaces",
			path:      "/home/user/My Movies/Movie Name (2023).mkv",
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.validatePath(tt.path)
			if tt.shouldErr && err == nil {
				t.Errorf("Expected error for path: %s", tt.path)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error for valid path %s: %v", tt.path, err)
			}
		})
	}
}

func TestCheckWritable(t *testing.T) {
	tmpDir := t.TempDir()
	v := NewValidator()

	tests := []struct {
		name      string
		setup     func() string
		shouldErr bool
	}{
		{
			name: "existing writable directory",
			setup: func() string {
				dir := filepath.Join(tmpDir, "writable")
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatalf("Failed to create writable directory: %v", err)
				}
				return dir
			},
			shouldErr: false,
		},
		{
			name: "non-existent directory (should create)",
			setup: func() string {
				return filepath.Join(tmpDir, "newdir")
			},
			shouldErr: false,
		},
		{
			name: "path is a file, not directory",
			setup: func() string {
				file := filepath.Join(tmpDir, "file.txt")
				if err := os.WriteFile(file, []byte("content"), 0644); err != nil {
					t.Fatalf("Failed to create file: %v", err)
				}
				return file
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			err := v.checkWritable(path)
			if tt.shouldErr && err == nil {
				t.Error("Expected error for non-writable directory")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidatePlan(t *testing.T) {
	tmpDir := t.TempDir()
	v := NewValidator()

	// Create some valid source files
	file1 := filepath.Join(tmpDir, "source1.mkv")
	file2 := filepath.Join(tmpDir, "source2.mkv")
	os.WriteFile(file1, []byte("content1"), 0644)
	os.WriteFile(file2, []byte("content2"), 0644)

	operations := []types.Operation{
		{
			Type:        types.OperationMove,
			Source:      file1,
			Destination: filepath.Join(tmpDir, "dest", "movie1.mkv"),
		},
		{
			Type:        types.OperationMove,
			Source:      file2,
			Destination: filepath.Join(tmpDir, "dest", "movie2.mkv"),
		},
		{
			Type:        types.OperationMove,
			Source:      filepath.Join(tmpDir, "nonexistent.mkv"),
			Destination: filepath.Join(tmpDir, "dest", "movie3.mkv"),
		},
	}

	errors := v.ValidatePlan(operations)

	// Should have 1 error (nonexistent source)
	if len(errors) != 1 {
		t.Errorf("Expected 1 validation error, got %d", len(errors))
	}
}

func TestValidationError_Error(t *testing.T) {
	op := types.Operation{
		Type:   types.OperationMove,
		Source: "/source.mkv",
	}

	err := &ValidationError{
		Operation: op,
		Reason:    "test reason",
	}

	expected := "validation failed for move operation: test reason"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestValidateOperation_UnknownType(t *testing.T) {
	v := NewValidator()

	op := types.Operation{
		Type: types.OperationType("unknown"),
	}

	err := v.ValidateOperation(op)
	if err == nil {
		t.Error("Expected error for unknown operation type")
	}
}

func TestCheckDiskSpace(t *testing.T) {
	tmpDir := t.TempDir()
	v := NewValidator()

	// Test with small required size (should pass)
	err := v.checkDiskSpace(tmpDir, 1024) // 1 KB
	if err != nil {
		t.Errorf("Unexpected error for small disk space requirement: %v", err)
	}

	// Test with very large required size (might fail depending on available space)
	// We just verify the function doesn't panic
	_ = v.checkDiskSpace(tmpDir, 1024*1024*1024*1024*100) // 100 TB
}
