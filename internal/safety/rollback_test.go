package safety

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/opd-ai/go-jf-org/pkg/types"
)

func TestRollbackMove(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "txn")
	tm, _ := NewTransactionManager(logDir)

	// Create source file
	sourceFile := filepath.Join(tmpDir, "source", "movie.mkv")
	destFile := filepath.Join(tmpDir, "dest", "Movie (2023)", "Movie (2023).mkv")

	if err := os.MkdirAll(filepath.Dir(sourceFile), 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	if err := os.WriteFile(sourceFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Create transaction and move file
	txn, _ := tm.Begin()

	// Simulate move operation
	if err := os.MkdirAll(filepath.Dir(destFile), 0755); err != nil {
		t.Fatalf("Failed to create destination directory: %v", err)
	}
	if err := os.Rename(sourceFile, destFile); err != nil {
		t.Fatalf("Failed to move source file to destination: %v", err)
	}

	op := types.Operation{
		Type:        types.OperationMove,
		Source:      sourceFile,
		Destination: destFile,
		Status:      types.OperationStatusCompleted,
	}
	tm.AddOperation(txn, op)
	tm.Complete(txn)

	// Rollback
	err := tm.Rollback(txn.ID)
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	// Verify file was moved back
	if _, err := os.Stat(sourceFile); os.IsNotExist(err) {
		t.Error("Source file was not restored")
	}

	if _, err := os.Stat(destFile); !os.IsNotExist(err) {
		t.Error("Destination file still exists")
	}

	// Verify transaction status
	loaded, _ := tm.Load(txn.ID)
	if loaded.Status != TransactionStatusRolledBack {
		t.Errorf("Expected status %s, got %s", TransactionStatusRolledBack, loaded.Status)
	}
}

func TestRollbackMultipleOperations(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "txn")
	tm, _ := NewTransactionManager(logDir)

	// Create multiple source files
	file1 := filepath.Join(tmpDir, "source", "movie1.mkv")
	file2 := filepath.Join(tmpDir, "source", "movie2.mkv")
	dest1 := filepath.Join(tmpDir, "dest", "Movie1 (2023)", "Movie1 (2023).mkv")
	dest2 := filepath.Join(tmpDir, "dest", "Movie2 (2024)", "Movie2 (2024).mkv")

	if err := os.MkdirAll(filepath.Dir(file1), 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	txn, _ := tm.Begin()

	// Simulate move operations
	if err := os.MkdirAll(filepath.Dir(dest1), 0755); err != nil {
		t.Fatalf("Failed to create dest1 directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(dest2), 0755); err != nil {
		t.Fatalf("Failed to create dest2 directory: %v", err)
	}
	if err := os.Rename(file1, dest1); err != nil {
		t.Fatalf("Failed to move file1: %v", err)
	}
	if err := os.Rename(file2, dest2); err != nil {
		t.Fatalf("Failed to move file2: %v", err)
	}

	op1 := types.Operation{
		Type:        types.OperationMove,
		Source:      file1,
		Destination: dest1,
		Status:      types.OperationStatusCompleted,
	}
	op2 := types.Operation{
		Type:        types.OperationMove,
		Source:      file2,
		Destination: dest2,
		Status:      types.OperationStatusCompleted,
	}

	tm.AddOperation(txn, op1)
	tm.AddOperation(txn, op2)
	tm.Complete(txn)

	// Rollback
	err := tm.Rollback(txn.ID)
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	// Verify both files were moved back
	if _, err := os.Stat(file1); os.IsNotExist(err) {
		t.Error("File 1 was not restored")
	}
	if _, err := os.Stat(file2); os.IsNotExist(err) {
		t.Error("File 2 was not restored")
	}

	if _, err := os.Stat(dest1); !os.IsNotExist(err) {
		t.Error("Destination 1 still exists")
	}
	if _, err := os.Stat(dest2); !os.IsNotExist(err) {
		t.Error("Destination 2 still exists")
	}
}

func TestRollbackCreateFile(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "txn")
	tm, _ := NewTransactionManager(logDir)

	// Create a file
	nfoFile := filepath.Join(tmpDir, "Movie (2023)", "movie.nfo")
	if err := os.MkdirAll(filepath.Dir(nfoFile), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.WriteFile(nfoFile, []byte("<movie></movie>"), 0644); err != nil {
		t.Fatalf("Failed to create NFO file: %v", err)
	}

	txn, _ := tm.Begin()
	op := types.Operation{
		Type:        types.OperationCreateFile,
		Destination: nfoFile,
		Status:      types.OperationStatusCompleted,
	}
	tm.AddOperation(txn, op)
	tm.Complete(txn)

	// Rollback
	err := tm.Rollback(txn.ID)
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	// Verify file was removed
	if _, err := os.Stat(nfoFile); !os.IsNotExist(err) {
		t.Error("NFO file was not removed")
	}
}

func TestRollbackCreateDir(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "txn")
	tm, _ := NewTransactionManager(logDir)

	// Create a directory
	newDir := filepath.Join(tmpDir, "Movie (2023)")
	if err := os.MkdirAll(newDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	txn, _ := tm.Begin()
	op := types.Operation{
		Type:        types.OperationCreateDir,
		Destination: newDir,
		Status:      types.OperationStatusCompleted,
	}
	tm.AddOperation(txn, op)
	tm.Complete(txn)

	// Rollback
	err := tm.Rollback(txn.ID)
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	// Verify directory was removed
	if _, err := os.Stat(newDir); !os.IsNotExist(err) {
		t.Error("Directory was not removed")
	}
}

func TestRollbackCreateDirNotEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "txn")
	tm, _ := NewTransactionManager(logDir)

	// Create a directory with a file
	newDir := filepath.Join(tmpDir, "Movie (2023)")
	if err := os.MkdirAll(newDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(newDir, "movie.mkv"), []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	txn, _ := tm.Begin()
	op := types.Operation{
		Type:        types.OperationCreateDir,
		Destination: newDir,
		Status:      types.OperationStatusCompleted,
	}
	tm.AddOperation(txn, op)
	tm.Complete(txn)

	// Rollback
	err := tm.Rollback(txn.ID)
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	// Verify directory still exists (not removed because not empty)
	if _, err := os.Stat(newDir); os.IsNotExist(err) {
		t.Error("Non-empty directory was removed")
	}
}

func TestRollbackAlreadyRolledBack(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "txn")
	tm, _ := NewTransactionManager(logDir)

	txn, _ := tm.Begin()
	tm.Complete(txn)
	
	// First rollback
	tm.Rollback(txn.ID)

	// Second rollback should fail
	err := tm.Rollback(txn.ID)
	if err == nil {
		t.Error("Expected error when rolling back already rolled back transaction")
	}
}

func TestRollbackPendingTransaction(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "txn")
	tm, _ := NewTransactionManager(logDir)

	txn, _ := tm.Begin()

	// Try to rollback pending transaction
	err := tm.Rollback(txn.ID)
	if err == nil {
		t.Error("Expected error when rolling back pending transaction")
	}
}

func TestRollbackNonExistentTransaction(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "txn")
	tm, _ := NewTransactionManager(logDir)

	err := tm.Rollback("nonexistent")
	if err == nil {
		t.Error("Expected error when rolling back non-existent transaction")
	}
}

func TestRollbackMissingDestination(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "txn")
	tm, _ := NewTransactionManager(logDir)

	sourceFile := filepath.Join(tmpDir, "source.mkv")
	destFile := filepath.Join(tmpDir, "dest.mkv")

	txn, _ := tm.Begin()
	op := types.Operation{
		Type:        types.OperationMove,
		Source:      sourceFile,
		Destination: destFile,
		Status:      types.OperationStatusCompleted,
	}
	tm.AddOperation(txn, op)
	tm.Complete(txn)

	// Rollback should fail because destination doesn't exist
	err := tm.Rollback(txn.ID)
	if err == nil {
		t.Error("Expected error when destination file missing")
	}
}

func TestRollbackSourceOccupied(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "txn")
	tm, _ := NewTransactionManager(logDir)

	sourceFile := filepath.Join(tmpDir, "source.mkv")
	destFile := filepath.Join(tmpDir, "dest.mkv")

	// Create destination file
	if err := os.WriteFile(destFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create destination file: %v", err)
	}

	txn, _ := tm.Begin()
	op := types.Operation{
		Type:        types.OperationMove,
		Source:      sourceFile,
		Destination: destFile,
		Status:      types.OperationStatusCompleted,
	}
	tm.AddOperation(txn, op)
	tm.Complete(txn)

	// Create a file at source location
	if err := os.WriteFile(sourceFile, []byte("new content"), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Rollback should fail because source is occupied
	err := tm.Rollback(txn.ID)
	if err == nil {
		t.Error("Expected error when source location occupied")
	}
}

func TestRollbackSkipsPendingOperations(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "txn")
	tm, _ := NewTransactionManager(logDir)

	// Create a file
	sourceFile := filepath.Join(tmpDir, "source.mkv")
	destFile := filepath.Join(tmpDir, "dest.mkv")
	if err := os.WriteFile(sourceFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	txn, _ := tm.Begin()

	// Add completed operation
	if err := os.Rename(sourceFile, destFile); err != nil {
		t.Fatalf("Failed to move file: %v", err)
	}
	op1 := types.Operation{
		Type:        types.OperationMove,
		Source:      sourceFile,
		Destination: destFile,
		Status:      types.OperationStatusCompleted,
	}
	tm.AddOperation(txn, op1)

	// Add pending operation
	op2 := types.Operation{
		Type:        types.OperationMove,
		Source:      "/fake/source.mkv",
		Destination: "/fake/dest.mkv",
		Status:      types.OperationStatusPending,
	}
	tm.AddOperation(txn, op2)

	tm.Complete(txn)

	// Rollback should only reverse completed operations
	err := tm.Rollback(txn.ID)
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	// Verify file was moved back
	if _, err := os.Stat(sourceFile); os.IsNotExist(err) {
		t.Error("Source file was not restored")
	}
}

func TestTryRemoveEmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "txn")
	tm, _ := NewTransactionManager(logDir)

	// Create nested empty directories
	deepDir := filepath.Join(tmpDir, "level1", "level2", "level3")
	os.MkdirAll(deepDir, 0755)

	// Remove deep directory and parents
	tm.tryRemoveEmptyDir(deepDir)

	// Verify all levels were removed
	if _, err := os.Stat(filepath.Join(tmpDir, "level1")); !os.IsNotExist(err) {
		t.Error("Parent directories were not removed")
	}
}

func TestTryRemoveEmptyDirWithFiles(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "txn")
	tm, _ := NewTransactionManager(logDir)

	// Create nested directories with a file in middle level
	deepDir := filepath.Join(tmpDir, "level1", "level2", "level3")
	os.MkdirAll(deepDir, 0755)
	os.WriteFile(filepath.Join(tmpDir, "level1", "level2", "file.txt"), []byte("content"), 0644)

	// Try to remove deep directory
	tm.tryRemoveEmptyDir(deepDir)

	// level3 should be removed but level2 and level1 should remain
	if _, err := os.Stat(deepDir); !os.IsNotExist(err) {
		t.Error("Empty directory was not removed")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "level1", "level2")); os.IsNotExist(err) {
		t.Error("Directory with files was incorrectly removed")
	}
}
