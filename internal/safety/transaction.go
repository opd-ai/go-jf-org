package safety

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/opd-ai/go-jf-org/pkg/types"
)

// Transaction represents a set of file operations that can be rolled back
type Transaction struct {
	ID         string               `json:"id"`
	Timestamp  time.Time            `json:"timestamp"`
	Operations []types.Operation    `json:"operations"`
	Status     TransactionStatus    `json:"status"`
	Completed  time.Time            `json:"completed,omitempty"`
	Error      string               `json:"error,omitempty"`
}

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	// TransactionStatusPending represents a pending transaction
	TransactionStatusPending TransactionStatus = "pending"
	// TransactionStatusCompleted represents a completed transaction
	TransactionStatusCompleted TransactionStatus = "completed"
	// TransactionStatusFailed represents a failed transaction
	TransactionStatusFailed TransactionStatus = "failed"
	// TransactionStatusRolledBack represents a rolled back transaction
	TransactionStatusRolledBack TransactionStatus = "rolled_back"
)

// TransactionManager handles transaction logging and retrieval
type TransactionManager struct {
	logDir string
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(logDir string) (*TransactionManager, error) {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create transaction log directory: %w", err)
	}

	return &TransactionManager{
		logDir: logDir,
	}, nil
}

// generateID generates a random transaction ID
func generateID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random fails
		return fmt.Sprintf("txn-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// Begin starts a new transaction
func (tm *TransactionManager) Begin() (*Transaction, error) {
	txn := &Transaction{
		ID:         generateID(),
		Timestamp:  time.Now(),
		Operations: make([]types.Operation, 0),
		Status:     TransactionStatusPending,
	}

	// Write initial transaction log
	if err := tm.save(txn); err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return txn, nil
}

// AddOperation adds an operation to the transaction
func (tm *TransactionManager) AddOperation(txn *Transaction, op types.Operation) error {
	txn.Operations = append(txn.Operations, op)
	return tm.save(txn)
}

// UpdateOperation updates an existing operation in the transaction by index
func (tm *TransactionManager) UpdateOperation(txn *Transaction, index int, op types.Operation) error {
	if index < 0 || index >= len(txn.Operations) {
		return fmt.Errorf("invalid operation index: %d", index)
	}
	txn.Operations[index] = op
	return tm.save(txn)
}

// Complete marks a transaction as completed
func (tm *TransactionManager) Complete(txn *Transaction) error {
	txn.Status = TransactionStatusCompleted
	txn.Completed = time.Now()
	return tm.save(txn)
}

// Fail marks a transaction as failed
func (tm *TransactionManager) Fail(txn *Transaction, err error) error {
	txn.Status = TransactionStatusFailed
	txn.Completed = time.Now()
	if err != nil {
		txn.Error = err.Error()
	}
	return tm.save(txn)
}

// MarkRolledBack marks a transaction as rolled back
func (tm *TransactionManager) MarkRolledBack(txn *Transaction) error {
	txn.Status = TransactionStatusRolledBack
	return tm.save(txn)
}

// Load loads a transaction by ID
func (tm *TransactionManager) Load(id string) (*Transaction, error) {
	path := tm.getLogPath(id)
	
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("transaction %s not found", id)
		}
		return nil, fmt.Errorf("failed to read transaction log: %w", err)
	}

	var txn Transaction
	if err := json.Unmarshal(data, &txn); err != nil {
		return nil, fmt.Errorf("failed to parse transaction log: %w", err)
	}

	return &txn, nil
}

// List returns all transaction IDs
func (tm *TransactionManager) List() ([]string, error) {
	entries, err := os.ReadDir(tm.logDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read transaction directory: %w", err)
	}

	ids := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) == ".json" {
			ids = append(ids, name[:len(name)-5]) // Remove .json extension
		}
	}

	return ids, nil
}

// save writes the transaction to disk
func (tm *TransactionManager) save(txn *Transaction) error {
	path := tm.getLogPath(txn.ID)

	data, err := json.MarshalIndent(txn, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write transaction log: %w", err)
	}

	return nil
}

// getLogPath returns the file path for a transaction log
func (tm *TransactionManager) getLogPath(id string) string {
	return filepath.Join(tm.logDir, id+".json")
}

// GetDefaultLogDir returns the default transaction log directory
func GetDefaultLogDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".go-jf-org", "txn"), nil
}
