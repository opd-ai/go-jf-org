package safety

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/opd-ai/go-jf-org/pkg/types"
)

// Validator performs pre-operation validation checks
type Validator struct {
	minFreeSpace uint64 // Minimum free space in bytes (10% buffer)
}

// NewValidator creates a new validator with default settings
func NewValidator() *Validator {
	return &Validator{
		minFreeSpace: 1024 * 1024 * 100, // 100 MB minimum
	}
}

// ValidationError represents a validation failure
type ValidationError struct {
	Operation types.Operation
	Reason    string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s operation: %s", e.Operation.Type, e.Reason)
}

// ValidateOperation performs comprehensive validation on an operation before execution
func (v *Validator) ValidateOperation(op types.Operation) error {
	switch op.Type {
	case types.OperationMove, types.OperationRename:
		return v.validateMoveOperation(op)
	case types.OperationCreateDir:
		return v.validateCreateDirOperation(op)
	case types.OperationCreateFile:
		return v.validateCreateFileOperation(op)
	default:
		return &ValidationError{
			Operation: op,
			Reason:    fmt.Sprintf("unknown operation type: %s", op.Type),
		}
	}
}

// validateMoveOperation validates a move/rename operation
func (v *Validator) validateMoveOperation(op types.Operation) error {
	// Check source exists
	sourceInfo, err := os.Stat(op.Source)
	if err != nil {
		if os.IsNotExist(err) {
			return &ValidationError{
				Operation: op,
				Reason:    fmt.Sprintf("source file does not exist: %s", op.Source),
			}
		}
		return &ValidationError{
			Operation: op,
			Reason:    fmt.Sprintf("cannot access source file: %v", err),
		}
	}

	// Ensure source is a file, not a directory
	if sourceInfo.IsDir() {
		return &ValidationError{
			Operation: op,
			Reason:    fmt.Sprintf("source is a directory, not a file: %s", op.Source),
		}
	}

	// Check source is readable
	file, err := os.Open(op.Source)
	if err != nil {
		return &ValidationError{
			Operation: op,
			Reason:    fmt.Sprintf("source file is not readable: %v", err),
		}
	}
	file.Close()

	// Validate destination path
	if err := v.validatePath(op.Destination); err != nil {
		return &ValidationError{
			Operation: op,
			Reason:    fmt.Sprintf("invalid destination path: %v", err),
		}
	}

	// Check destination directory is writable
	destDir := filepath.Dir(op.Destination)
	if err := v.checkWritable(destDir); err != nil {
		return &ValidationError{
			Operation: op,
			Reason:    fmt.Sprintf("destination directory not writable: %v", err),
		}
	}

	// Check sufficient disk space
	fileSize := sourceInfo.Size()
	if err := v.checkDiskSpace(destDir, uint64(fileSize)); err != nil {
		return &ValidationError{
			Operation: op,
			Reason:    fmt.Sprintf("insufficient disk space: %v", err),
		}
	}

	return nil
}

// validateCreateDirOperation validates a directory creation operation
func (v *Validator) validateCreateDirOperation(op types.Operation) error {
	// Validate path
	if err := v.validatePath(op.Destination); err != nil {
		return &ValidationError{
			Operation: op,
			Reason:    fmt.Sprintf("invalid directory path: %v", err),
		}
	}

	// Check if directory already exists
	if _, err := os.Stat(op.Destination); err == nil {
		return &ValidationError{
			Operation: op,
			Reason:    fmt.Sprintf("directory already exists: %s", op.Destination),
		}
	}

	// Check parent directory is writable
	parentDir := filepath.Dir(op.Destination)
	if err := v.checkWritable(parentDir); err != nil {
		return &ValidationError{
			Operation: op,
			Reason:    fmt.Sprintf("parent directory not writable: %v", err),
		}
	}

	return nil
}

// validateCreateFileOperation validates a file creation operation
func (v *Validator) validateCreateFileOperation(op types.Operation) error {
	// Validate path
	if err := v.validatePath(op.Destination); err != nil {
		return &ValidationError{
			Operation: op,
			Reason:    fmt.Sprintf("invalid file path: %v", err),
		}
	}

	// Check if file already exists
	if _, err := os.Stat(op.Destination); err == nil {
		return &ValidationError{
			Operation: op,
			Reason:    fmt.Sprintf("file already exists: %s", op.Destination),
		}
	}

	// Check directory exists and is writable
	destDir := filepath.Dir(op.Destination)
	if err := v.checkWritable(destDir); err != nil {
		return &ValidationError{
			Operation: op,
			Reason:    fmt.Sprintf("destination directory not writable: %v", err),
		}
	}

	return nil
}

// validatePath checks if a path contains unsafe characters
func (v *Validator) validatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path is empty")
	}

	// Check for unsafe characters (following Jellyfin conventions)
	unsafeChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range unsafeChars {
		if strings.Contains(path, char) {
			return fmt.Errorf("path contains unsafe character '%s'", char)
		}
	}

	// Check for leading/trailing dots or spaces in filename
	filename := filepath.Base(path)
	if strings.HasPrefix(filename, ".") && filename != "." && filename != ".." {
		return fmt.Errorf("filename starts with dot: %s", filename)
	}
	if strings.HasPrefix(filename, " ") || strings.HasSuffix(filename, " ") {
		return fmt.Errorf("filename has leading or trailing spaces: %s", filename)
	}
	if strings.HasSuffix(filename, ".") {
		return fmt.Errorf("filename ends with dot: %s", filename)
	}

	return nil
}

// checkWritable verifies a directory exists and is writable
func (v *Validator) checkWritable(dir string) error {
	// Check if directory exists
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			// Try to create it
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("cannot create directory: %w", err)
			}
			return nil
		}
		return fmt.Errorf("cannot access directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", dir)
	}

	// Try to create a temporary file to verify write permissions
	tmpFile := filepath.Join(dir, ".go-jf-org-write-test")
	f, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("directory is not writable: %w", err)
	}
	f.Close()
	os.Remove(tmpFile)

	return nil
}

// checkDiskSpace verifies sufficient disk space is available
func (v *Validator) checkDiskSpace(path string, requiredBytes uint64) error {
	// Add 10% buffer
	requiredBytes = requiredBytes + (requiredBytes / 10)

	// Ensure minimum free space
	if requiredBytes < v.minFreeSpace {
		requiredBytes = v.minFreeSpace
	}

	// Note: syscall.Statfs_t is Unix-specific
	// On Windows, this check will be skipped
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		// If we can't check disk space (e.g., on Windows), just continue
		// This is a best-effort check, not a strict requirement
		return nil
	}

	// Available blocks * block size
	availableBytes := stat.Bavail * uint64(stat.Bsize)

	if availableBytes < requiredBytes {
		return fmt.Errorf("insufficient disk space: need %d bytes, have %d bytes", requiredBytes, availableBytes)
	}

	return nil
}

// ValidatePlan validates all operations in a plan
func (v *Validator) ValidatePlan(operations []types.Operation) []error {
	errors := make([]error, 0)

	for _, op := range operations {
		if err := v.ValidateOperation(op); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}
