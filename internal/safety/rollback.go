package safety

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"

	"github.com/opd-ai/go-jf-org/pkg/types"
)

// Rollback reverses a completed or failed transaction
func (tm *TransactionManager) Rollback(txnID string) error {
	// Load the transaction
	txn, err := tm.Load(txnID)
	if err != nil {
		return fmt.Errorf("failed to load transaction: %w", err)
	}

	// Check if already rolled back
	if txn.Status == TransactionStatusRolledBack {
		return fmt.Errorf("transaction %s has already been rolled back", txnID)
	}

	// Can only rollback completed or failed transactions
	if txn.Status != TransactionStatusCompleted && txn.Status != TransactionStatusFailed {
		return fmt.Errorf("cannot rollback transaction in status %s", txn.Status)
	}

	log.Info().Str("transaction", txnID).Int("operations", len(txn.Operations)).Msg("Starting rollback")

	// Reverse operations in reverse order
	var rollbackErrors []error
	successCount := 0

	for i := len(txn.Operations) - 1; i >= 0; i-- {
		op := txn.Operations[i]

		// Only rollback completed operations
		if op.Status != types.OperationStatusCompleted {
			log.Debug().
				Str("type", string(op.Type)).
				Str("source", op.Source).
				Str("status", string(op.Status)).
				Msg("Skipping operation - not completed")
			continue
		}

		if err := tm.rollbackOperation(op); err != nil {
			log.Error().
				Err(err).
				Str("type", string(op.Type)).
				Str("source", op.Source).
				Str("destination", op.Destination).
				Msg("Failed to rollback operation")
			rollbackErrors = append(rollbackErrors, err)
		} else {
			successCount++
		}
	}

	// Mark transaction as rolled back
	if err := tm.MarkRolledBack(txn); err != nil {
		log.Error().Err(err).Msg("Failed to mark transaction as rolled back")
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	log.Info().
		Str("transaction", txnID).
		Int("success", successCount).
		Int("failed", len(rollbackErrors)).
		Msg("Rollback complete")

	if len(rollbackErrors) > 0 {
		return fmt.Errorf("rollback completed with %d errors", len(rollbackErrors))
	}

	return nil
}

// rollbackOperation reverses a single operation
func (tm *TransactionManager) rollbackOperation(op types.Operation) error {
	switch op.Type {
	case types.OperationMove:
		return tm.rollbackMove(op)
	case types.OperationRename:
		return tm.rollbackRename(op)
	case types.OperationCreateDir:
		return tm.rollbackCreateDir(op)
	case types.OperationCreateFile:
		return tm.rollbackCreateFile(op)
	default:
		return fmt.Errorf("unknown operation type: %s", op.Type)
	}
}

// rollbackMove reverses a file move operation
func (tm *TransactionManager) rollbackMove(op types.Operation) error {
	// Move file back from destination to source
	log.Debug().
		Str("from", op.Destination).
		Str("to", op.Source).
		Msg("Rolling back move operation")

	// Check if destination still exists
	if _, err := os.Stat(op.Destination); os.IsNotExist(err) {
		return fmt.Errorf("destination file no longer exists: %s", op.Destination)
	}

	// Check if source location is available (not recreated)
	if _, err := os.Stat(op.Source); err == nil {
		return fmt.Errorf("source location already occupied: %s", op.Source)
	}

	// Ensure source directory exists
	sourceDir := filepath.Dir(op.Source)
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		return fmt.Errorf("failed to create source directory: %w", err)
	}

	// Move file back
	if err := os.Rename(op.Destination, op.Source); err != nil {
		return fmt.Errorf("failed to move file back: %w", err)
	}

	log.Info().
		Str("from", op.Destination).
		Str("to", op.Source).
		Msg("File moved back successfully")

	// Try to remove empty destination directory
	destDir := filepath.Dir(op.Destination)
	tm.tryRemoveEmptyDir(destDir)

	return nil
}

// rollbackRename reverses a file rename operation
func (tm *TransactionManager) rollbackRename(op types.Operation) error {
	// Rename is essentially the same as move for rollback purposes
	return tm.rollbackMove(op)
}

// rollbackCreateDir removes a created directory if it's empty
func (tm *TransactionManager) rollbackCreateDir(op types.Operation) error {
	log.Debug().Str("dir", op.Destination).Msg("Rolling back directory creation")

	// Check if directory exists
	info, err := os.Stat(op.Destination)
	if os.IsNotExist(err) {
		// Directory already gone, nothing to do
		log.Debug().Str("dir", op.Destination).Msg("Directory already removed")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to stat directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", op.Destination)
	}

	// Only remove if empty
	entries, err := os.ReadDir(op.Destination)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	if len(entries) > 0 {
		log.Warn().
			Str("dir", op.Destination).
			Int("files", len(entries)).
			Msg("Directory not empty, skipping removal")
		return nil
	}

	// Remove empty directory
	if err := os.Remove(op.Destination); err != nil {
		return fmt.Errorf("failed to remove directory: %w", err)
	}

	log.Info().Str("dir", op.Destination).Msg("Directory removed")
	return nil
}

// rollbackCreateFile removes a created file
func (tm *TransactionManager) rollbackCreateFile(op types.Operation) error {
	log.Debug().Str("file", op.Destination).Msg("Rolling back file creation")

	// Check if file exists
	if _, err := os.Stat(op.Destination); os.IsNotExist(err) {
		// File already gone, nothing to do
		log.Debug().Str("file", op.Destination).Msg("File already removed")
		return nil
	}

	// Remove the file
	if err := os.Remove(op.Destination); err != nil {
		return fmt.Errorf("failed to remove file: %w", err)
	}

	log.Info().Str("file", op.Destination).Msg("File removed")
	return nil
}

// tryRemoveEmptyDir attempts to remove a directory if it's empty, doesn't error if not empty
func (tm *TransactionManager) tryRemoveEmptyDir(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) > 0 {
		return
	}

	if err := os.Remove(dir); err != nil {
		log.Debug().Err(err).Str("dir", dir).Msg("Could not remove directory")
		return
	}

	log.Debug().Str("dir", dir).Msg("Removed empty directory")

	// Recursively try to remove parent if empty
	parent := filepath.Dir(dir)
	if parent != dir && parent != "." && parent != "/" {
		tm.tryRemoveEmptyDir(parent)
	}
}
