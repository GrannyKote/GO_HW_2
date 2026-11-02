package ledger_test

import (
	"strings"
	"testing"

	"ledger/ledger"
)

func TestTransactionValidate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		tx      ledger.Transaction
		wantErr bool
	}{
		{
			name: "valid transaction",
			tx: ledger.Transaction{
				Amount: 100, Category: "food", Description: "lunch", Date: "2025-09-10",
			},
			wantErr: false,
		},
		{
			name:    "zero amount",
			tx:      ledger.Transaction{Amount: 0, Category: "food", Date: "2025-09-10"},
			wantErr: true,
		},
		{
			name:    "negative amount",
			tx:      ledger.Transaction{Amount: -10, Category: "food", Date: "2025-09-10"},
			wantErr: true,
		},
		{
			name:    "empty category",
			tx:      ledger.Transaction{Amount: 50, Category: "", Date: "2025-09-10"},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.tx.Validate()
			if tc.wantErr && err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
		})
	}
}

func TestBudgetValidate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		b       ledger.Budget
		wantErr bool
	}{
		{
			name:    "positive limit",
			b:       ledger.Budget{Category: "еда", Limit: 5000},
			wantErr: false,
		},
		{
			name:    "zero limit",
			b:       ledger.Budget{Category: "food", Limit: 0},
			wantErr: true,
		},
		{
			name:    "negative limit",
			b:       ledger.Budget{Category: "food", Limit: -100},
			wantErr: true,
		},
		{
			name:    "empty category",
			b:       ledger.Budget{Category: "", Limit: 1000},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.b.Validate()
			if tc.wantErr && err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
		})
	}
}

func TestAddTransactionBudgetLimits(t *testing.T) {
	ledger.Reset()
	t.Cleanup(ledger.Reset)

	if err := ledger.SetBudget(ledger.Budget{Category: "еда", Limit: 5000}); err != nil {
		t.Fatalf("SetBudget: %v", err)
	}

	before := len(ledger.ListTransactions())

	txOK := ledger.Transaction{
		Amount: 3000, Category: "еда", Description: "groceries", Date: "2025-09-10",
	}
	if err := ledger.AddTransaction(txOK); err != nil {
		t.Fatalf("AddTransaction within limit: %v", err)
	}

	txOver := ledger.Transaction{
		Amount: 2500, Category: "еда", Description: "would exceed", Date: "2025-09-11",
	}
	err := ledger.AddTransaction(txOver)
	if err == nil {
		t.Fatal("expected budget exceeded error, got nil")
	}
	if !strings.Contains(err.Error(), "budget exceeded") {
		t.Fatalf("expected budget exceeded, got %q", err.Error())
	}

	after := ledger.ListTransactions()
	if len(after) != before+1 {
		t.Fatalf("transaction count: want %d, got %d", before+1, len(after))
	}
}
