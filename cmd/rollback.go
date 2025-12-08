package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/opd-ai/go-jf-org/internal/safety"
)

// rollbackCmd represents the rollback command
var rollbackCmd = &cobra.Command{
	Use:   "rollback [transaction-id]",
	Short: "Rollback a completed organization operation",
	Long: `Rollback reverses a completed or failed organization operation by moving
files back to their original locations.

The transaction ID can be found in the output of the organize command or by
listing transactions with 'rollback --list'.

Examples:
  # Rollback a specific transaction
  go-jf-org rollback abc123def456

  # List all transactions
  go-jf-org rollback --list

  # Show details of a transaction
  go-jf-org rollback abc123def456 --show`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRollback,
}

var (
	listTransactions bool
	showTransaction  bool
)

func init() {
	rootCmd.AddCommand(rollbackCmd)

	rollbackCmd.Flags().BoolVarP(&listTransactions, "list", "l", false, "List all transactions")
	rollbackCmd.Flags().BoolVarP(&showTransaction, "show", "s", false, "Show transaction details without rolling back")
}

func runRollback(cmd *cobra.Command, args []string) error {
	// Get transaction log directory
	logDir, err := safety.GetDefaultLogDir()
	if err != nil {
		return fmt.Errorf("failed to get transaction log directory: %w", err)
	}

	tm, err := safety.NewTransactionManager(logDir)
	if err != nil {
		return fmt.Errorf("failed to initialize transaction manager: %w", err)
	}

	// List transactions
	if listTransactions {
		return listAllTransactions(tm)
	}

	// Require transaction ID
	if len(args) == 0 {
		return fmt.Errorf("transaction ID required (use --list to see available transactions)")
	}

	txnID := args[0]

	// Show transaction details
	if showTransaction {
		return showTransactionDetails(tm, txnID)
	}

	// Perform rollback
	return performRollback(tm, txnID)
}

func listAllTransactions(tm *safety.TransactionManager) error {
	ids, err := tm.List()
	if err != nil {
		return fmt.Errorf("failed to list transactions: %w", err)
	}

	if len(ids) == 0 {
		fmt.Println("No transactions found")
		return nil
	}

	fmt.Printf("Found %d transaction(s):\n\n", len(ids))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSTATUS\tOPERATIONS\tTIMESTAMP")
	fmt.Fprintln(w, "--\t------\t----------\t---------")

	for _, id := range ids {
		txn, err := tm.Load(id)
		if err != nil {
			log.Warn().Err(err).Str("id", id).Msg("Failed to load transaction")
			continue
		}

		timestamp := txn.Timestamp.Format("2006-01-02 15:04:05")
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", txn.ID, txn.Status, len(txn.Operations), timestamp)
	}

	w.Flush()

	fmt.Println("\nUse 'rollback <id> --show' to see details of a transaction")
	fmt.Println("Use 'rollback <id>' to rollback a transaction")

	return nil
}

func showTransactionDetails(tm *safety.TransactionManager, txnID string) error {
	txn, err := tm.Load(txnID)
	if err != nil {
		return fmt.Errorf("failed to load transaction: %w", err)
	}

	fmt.Printf("Transaction: %s\n", txn.ID)
	fmt.Printf("Status:      %s\n", txn.Status)
	fmt.Printf("Created:     %s\n", txn.Timestamp.Format(time.RFC1123))
	if !txn.Completed.IsZero() {
		fmt.Printf("Completed:   %s\n", txn.Completed.Format(time.RFC1123))
	}
	if txn.Error != "" {
		fmt.Printf("Error:       %s\n", txn.Error)
	}
	fmt.Printf("\nOperations:  %d\n\n", len(txn.Operations))

	if len(txn.Operations) > 0 {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "TYPE\tSTATUS\tSOURCE\tDESTINATION")
		fmt.Fprintln(w, "----\t------\t------\t-----------")

		for _, op := range txn.Operations {
			source := op.Source
			if source == "" {
				source = "-"
			}
			dest := op.Destination
			if dest == "" {
				dest = "-"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", op.Type, op.Status, source, dest)
		}

		w.Flush()
	}

	return nil
}

func performRollback(tm *safety.TransactionManager, txnID string) error {
	// Load and show transaction info
	txn, err := tm.Load(txnID)
	if err != nil {
		return fmt.Errorf("failed to load transaction: %w", err)
	}

	fmt.Printf("Rolling back transaction: %s\n", txnID)
	fmt.Printf("Status:      %s\n", txn.Status)
	fmt.Printf("Operations:  %d\n\n", len(txn.Operations))

	// Check if already rolled back
	if txn.Status == safety.TransactionStatusRolledBack {
		return fmt.Errorf("transaction has already been rolled back")
	}

	// Perform rollback
	log.Info().Str("transaction", txnID).Msg("Starting rollback")

	if err := tm.Rollback(txnID); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	fmt.Println("✓ Rollback completed successfully")

	return nil
}
