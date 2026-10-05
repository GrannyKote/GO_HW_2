package main

import (
	"fmt"
	"log"

	"ledger/ledger"
)

func main() {
	fmt.Println("Ledger service started")

	samples := []ledger.Transaction{
		{Amount: 1500.50, Category: "food", Description: "Groceries", Date: "2026-10-01"},
		{Amount: 320, Category: "transport", Description: "Taxi to airport", Date: "2026-10-03"},
		{Amount: 89.99, Category: "entertainment", Description: "Cinema tickets", Date: "2026-10-05"},
	}

	for _, tx := range samples {
		if err := ledger.AddTransaction(tx); err != nil {
			log.Fatalf("add transaction: %v", err)
		}
	}

	if err := ledger.AddTransaction(ledger.Transaction{
		Amount: 0, Category: "test", Description: "invalid", Date: "2026-10-05",
	}); err != nil {
		fmt.Println("Expected error for zero amount:", err)
	}

	fmt.Println("Transactions:")
	for _, tx := range ledger.ListTransactions() {
		fmt.Println(tx)
	}
}
