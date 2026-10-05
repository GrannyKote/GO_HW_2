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
func SetBudget(b Budget) {
	if b.Category == "" {
		return
	}
	budgets[b.Category] = b
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
		if b.Category == "" {
			return errors.New("budget entry missing category")
		}
		if b.Limit < 0 {
			return fmt.Errorf("budget for category %q has negative limit", b.Category)
		}
		SetBudget(b)
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
// Returns an error if Amount is zero or if the transaction would exceed the category budget.
func AddTransaction(tx Transaction) error {
	if tx.Amount == 0 {
		return errors.New("transaction amount must not be zero")
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
