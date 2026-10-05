package ledger

import (
	"errors"
	"fmt"
)

// Transaction represents a financial transaction.
type Transaction struct {
	ID          int
	Amount      float64
	Category    string
	Description string
	Date        string
}

var transactions []Transaction

// AddTransaction appends a transaction to in-memory storage.
// Returns an error if Amount is zero.
func AddTransaction(tx Transaction) error {
	if tx.Amount == 0 {
		return errors.New("transaction amount must not be zero")
	}
	tx.ID = len(transactions) + 1
	transactions = append(transactions, tx)
	return nil
}

// ListTransactions returns a copy of all stored transactions.
func ListTransactions() []Transaction {
	out := make([]Transaction, len(transactions))
	copy(out, transactions)
	return out
}

// String formats a transaction for console output.
func (t Transaction) String() string {
	return fmt.Sprintf("ID=%d Amount=%.2f Category=%q Description=%q Date=%s",
		t.ID, t.Amount, t.Category, t.Description, t.Date)
}
