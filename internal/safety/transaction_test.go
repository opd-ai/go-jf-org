package safety

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/opd-ai/go-jf-org/pkg/types"
)

func TestNewTransactionManager(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "txn")

	tm, err := NewTransactionManager(logDir)
	if err != nil {
		t.Fatalf("NewTransactionManager failed: %v", err)
	}

	if tm.logDir != logDir {
		t.Errorf("Expected logDir %s, got %s", logDir, tm.logDir)
	}

	// Verify directory was created
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		t.Error("Transaction log directory was not created")
	}
}

func TestBegin(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "txn")
	tm, _ := NewTransactionManager(logDir)

	txn, err := tm.Begin()
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	if txn.ID == "" {
		t.Error("Transaction ID is empty")
	}

	if txn.Status != TransactionStatusPending {
		t.Errorf("Expected status %s, got %s", TransactionStatusPending, txn.Status)
	}

	if txn.Timestamp.IsZero() {
		t.Error("Transaction timestamp is zero")
	}

	if len(txn.Operations) != 0 {
		t.Errorf("Expected 0 operations, got %d", len(txn.Operations))
	}

	// Verify transaction was saved
	logPath := filepath.Join(logDir, txn.ID+".json")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Transaction log file was not created")
	}
}

func TestAddOperation(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "txn")
	tm, _ := NewTransactionManager(logDir)

	txn, _ := tm.Begin()

	op := types.Operation{
		Type:        types.OperationMove,
		Source:      "/source/file.mkv",
		Destination: "/dest/Movie (2023)/Movie (2023).mkv",
		Status:      types.OperationStatusPending,
	}

	err := tm.AddOperation(txn, op)
	if err != nil {
		t.Fatalf("AddOperation failed: %v", err)
	}

	if len(txn.Operations) != 1 {
		t.Errorf("Expected 1 operation, got %d", len(txn.Operations))
	}

	if txn.Operations[0].Source != op.Source {
		t.Errorf("Expected source %s, got %s", op.Source, txn.Operations[0].Source)
	}

	// Load and verify operation was saved
	loaded, _ := tm.Load(txn.ID)
	if len(loaded.Operations) != 1 {
		t.Error("Operation was not persisted")
	}
}

func TestComplete(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "txn")
	tm, _ := NewTransactionManager(logDir)

	txn, _ := tm.Begin()

	err := tm.Complete(txn)
	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}

	if txn.Status != TransactionStatusCompleted {
		t.Errorf("Expected status %s, got %s", TransactionStatusCompleted, txn.Status)
	}

	if txn.Completed.IsZero() {
		t.Error("Completed timestamp is zero")
	}

	// Load and verify status was saved
	loaded, _ := tm.Load(txn.ID)
	if loaded.Status != TransactionStatusCompleted {
		t.Error("Status was not persisted")
	}
}

func TestFail(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "txn")
	tm, _ := NewTransactionManager(logDir)

	txn, _ := tm.Begin()

	testErr := os.ErrPermission
	err := tm.Fail(txn, testErr)
	if err != nil {
		t.Fatalf("Fail failed: %v", err)
	}

	if txn.Status != TransactionStatusFailed {
		t.Errorf("Expected status %s, got %s", TransactionStatusFailed, txn.Status)
	}

	if txn.Error == "" {
		t.Error("Error message is empty")
	}

	if txn.Completed.IsZero() {
		t.Error("Completed timestamp is zero")
	}

	// Load and verify error was saved
	loaded, _ := tm.Load(txn.ID)
	if loaded.Status != TransactionStatusFailed {
		t.Error("Failed status was not persisted")
	}
	if loaded.Error == "" {
		t.Error("Error message was not persisted")
	}
}

func TestMarkRolledBack(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "txn")
	tm, _ := NewTransactionManager(logDir)

	txn, _ := tm.Begin()
	tm.Complete(txn)

	err := tm.MarkRolledBack(txn)
	if err != nil {
		t.Fatalf("MarkRolledBack failed: %v", err)
	}

	if txn.Status != TransactionStatusRolledBack {
		t.Errorf("Expected status %s, got %s", TransactionStatusRolledBack, txn.Status)
	}

	// Load and verify status was saved
	loaded, _ := tm.Load(txn.ID)
	if loaded.Status != TransactionStatusRolledBack {
		t.Error("Rolled back status was not persisted")
	}
}

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "txn")
	tm, _ := NewTransactionManager(logDir)

	// Create a transaction
	txn, _ := tm.Begin()
	op := types.Operation{
		Type:        types.OperationMove,
		Source:      "/test/source.mkv",
		Destination: "/test/dest.mkv",
		Status:      types.OperationStatusCompleted,
	}
	tm.AddOperation(txn, op)
	tm.Complete(txn)

	// Load it back
	loaded, err := tm.Load(txn.ID)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.ID != txn.ID {
		t.Errorf("Expected ID %s, got %s", txn.ID, loaded.ID)
	}

	if loaded.Status != TransactionStatusCompleted {
		t.Errorf("Expected status %s, got %s", TransactionStatusCompleted, loaded.Status)
	}

	if len(loaded.Operations) != 1 {
		t.Errorf("Expected 1 operation, got %d", len(loaded.Operations))
	}

	// Test loading non-existent transaction
	_, err = tm.Load("nonexistent")
	if err == nil {
		t.Error("Expected error when loading non-existent transaction")
	}
}

func TestList(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "txn")
	tm, _ := NewTransactionManager(logDir)

	// Create multiple transactions
	txn1, _ := tm.Begin()
	txn2, _ := tm.Begin()
	txn3, _ := tm.Begin()

	ids, err := tm.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(ids) != 3 {
		t.Errorf("Expected 3 transactions, got %d", len(ids))
	}

	// Verify all IDs are present
	idMap := make(map[string]bool)
	for _, id := range ids {
		idMap[id] = true
	}

	if !idMap[txn1.ID] {
		t.Error("Transaction 1 ID not in list")
	}
	if !idMap[txn2.ID] {
		t.Error("Transaction 2 ID not in list")
	}
	if !idMap[txn3.ID] {
		t.Error("Transaction 3 ID not in list")
	}
}

func TestTransactionPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "txn")

	// Create and save transaction
	tm1, _ := NewTransactionManager(logDir)
	txn, _ := tm1.Begin()

	op := types.Operation{
		Type:        types.OperationMove,
		Source:      "/test/file.mkv",
		Destination: "/test/organized/file.mkv",
		Status:      types.OperationStatusCompleted,
	}
	tm1.AddOperation(txn, op)
	tm1.Complete(txn)

	// Create new manager and load transaction
	tm2, _ := NewTransactionManager(logDir)
	loaded, err := tm2.Load(txn.ID)
	if err != nil {
		t.Fatalf("Failed to load transaction with new manager: %v", err)
	}

	if loaded.ID != txn.ID {
		t.Error("Transaction ID mismatch after reload")
	}

	if loaded.Status != TransactionStatusCompleted {
		t.Error("Transaction status not persisted")
	}

	if len(loaded.Operations) != 1 {
		t.Error("Operations not persisted")
	}
}

func TestGetDefaultLogDir(t *testing.T) {
	dir, err := GetDefaultLogDir()
	if err != nil {
		t.Fatalf("GetDefaultLogDir failed: %v", err)
	}

	if dir == "" {
		t.Error("Default log dir is empty")
	}

	if !filepath.IsAbs(dir) {
		t.Error("Default log dir is not absolute")
	}

	// Should contain .go-jf-org/txn
	if !filepath.IsAbs(dir) {
		t.Error("Default log dir should be absolute path")
	}

	// Verify it ends with the expected path
	if !strings.HasSuffix(dir, filepath.Join(".go-jf-org", "txn")) {
		t.Logf("Default log dir: %s", dir)
		t.Error("Default log dir does not end with .go-jf-org/txn")
	}
}

func TestConcurrentTransactions(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "txn")
	tm, _ := NewTransactionManager(logDir)

	// Create multiple transactions concurrently
	type result struct {
		err error
	}
	results := make(chan result, 5)

	for i := 0; i < 5; i++ {
		go func() {
			var res result

			txn, err := tm.Begin()
			if err != nil {
				res.err = fmt.Errorf("failed to begin transaction: %w", err)
				results <- res
				return
			}

			op := types.Operation{
				Type:        types.OperationMove,
				Source:      "/test/source.mkv",
				Destination: "/test/dest.mkv",
				Status:      types.OperationStatusPending,
			}

			if err := tm.AddOperation(txn, op); err != nil {
				res.err = fmt.Errorf("failed to add operation: %w", err)
				results <- res
				return
			}

			if err := tm.Complete(txn); err != nil {
				res.err = fmt.Errorf("failed to complete transaction: %w", err)
				results <- res
				return
			}

			results <- res
		}()
	}

	// Wait for all goroutines and check for errors
	for i := 0; i < 5; i++ {
		res := <-results
		if res.err != nil {
			t.Error(res.err)
		}
	}

	// Verify all transactions were created
	ids, _ := tm.List()
	if len(ids) != 5 {
		t.Errorf("Expected 5 transactions, got %d", len(ids))
	}
}

func TestTransactionTimestamps(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "txn")
	tm, _ := NewTransactionManager(logDir)

	before := time.Now()
	txn, _ := tm.Begin()

	// Small delay to ensure timestamps differ
	time.Sleep(10 * time.Millisecond)

	tm.Complete(txn)
	after := time.Now()

	if txn.Timestamp.Before(before) || txn.Timestamp.After(after) {
		t.Error("Transaction timestamp out of expected range")
	}

	if txn.Completed.Before(txn.Timestamp) {
		t.Error("Completed timestamp is before transaction timestamp")
	}

	if txn.Completed.After(after) {
		t.Error("Completed timestamp is after expected time")
	}
}
