package ledger

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// Transaction represents a financial transaction.
type Transaction struct {
	ID          int
	Amount      float64
	Category    string
	Description string
	Date        string
}

// Budget holds a spending limit for a category.
type Budget struct {
	Category string  `json:"Category"`
	Limit    float64 `json:"Limit"`
}

var (
	transactions []Transaction
	budgets      = map[string]Budget{
		"food": {Category: "food", Limit: 5000},
		"transport": {Category: "transport", Limit: 3000},
	}
)

// SetBudget adds or updates a budget for the given category.
func SetBudget(b Budget) error {
	if err := b.Validate(); err != nil {
		return err
	}
	budgets[b.Category] = b
	return nil
}

// LoadBudgets reads a JSON array of budgets and stores them via SetBudget.
func LoadBudgets(r io.Reader) error {
	if r == nil {
		return errors.New("budget reader is nil")
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read budgets: %w", err)
	}
	var items []Budget
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("parse budgets JSON: %w", err)
	}
	for _, b := range items {
		if err := SetBudget(b); err != nil {
			return fmt.Errorf("budget %q: %w", b.Category, err)
		}
	}
	return nil
}

func sumByCategory(category string) float64 {
	var total float64
	for _, tx := range transactions {
		if tx.Category == category {
			total += tx.Amount
		}
	}
	return total
}

// AddTransaction appends a transaction to in-memory storage.
// Returns an error if validation fails or the transaction would exceed the category budget.
func AddTransaction(tx Transaction) error {
	if err := tx.Validate(); err != nil {
		return err
	}
	if b, ok := budgets[tx.Category]; ok {
		if sumByCategory(tx.Category)+tx.Amount > b.Limit {
			return errors.New("budget exceeded")
		}
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
