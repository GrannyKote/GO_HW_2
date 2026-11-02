package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"ledger/ledger"
)

func main() {
	fmt.Println("Ledger service started")

	f, err := os.Open("budgets.json")
	if err != nil {
		log.Fatalf("open budgets.json: %v", err)
	}
	defer f.Close()

	if err := ledger.LoadBudgets(bufio.NewReader(f)); err != nil {
		log.Fatalf("load budgets: %v", err)
	}
	fmt.Println("Budgets loaded from budgets.json")

	if err := ledger.SetBudget(ledger.Budget{Category: "transport", Limit: 2500}); err != nil {
		log.Fatalf("set budget: %v", err)
	}
	fmt.Println("SetBudget: transport limit updated to 2500")

	demoValidate := func(label string, v ledger.Validatable) {
		if err := ledger.CheckValid(v); err != nil {
			fmt.Printf("CheckValid %s: invalid — %v\n", label, err)
			return
		}
		fmt.Printf("CheckValid %s: valid\n", label)
	}
	demoValidate("good transaction", ledger.Transaction{
		Amount: 100, Category: "food", Description: "demo", Date: "2026-10-01",
	})
	demoValidate("bad transaction", ledger.Transaction{
		Amount: -5, Category: "food", Description: "demo", Date: "2026-10-01",
	})
	demoValidate("good budget", ledger.Budget{Category: "health", Limit: 1000})
	demoValidate("bad budget", ledger.Budget{Category: "", Limit: 500})

	tryAdd := func(label string, tx ledger.Transaction) {
		if err := ledger.AddTransaction(tx); err != nil {
			fmt.Printf("%s: rejected — %v\n", label, err)
			return
		}
		fmt.Printf("%s: ok\n", label)
	}

	tryAdd("Food within budget", ledger.Transaction{
		Amount: 1500.50, Category: "food", Description: "Groceries", Date: "2026-10-01",
	})
	tryAdd("Transport within budget", ledger.Transaction{
		Amount: 320, Category: "transport", Description: "Taxi", Date: "2026-10-03",
	})
	tryAdd("Entertainment within budget", ledger.Transaction{
		Amount: 89.99, Category: "entertainment", Description: "Cinema", Date: "2026-10-05",
	})
	tryAdd("Food over budget", ledger.Transaction{
		Amount: 4000, Category: "food", Description: "Would exceed food limit", Date: "2026-10-06",
	})
	tryAdd("Uncategorized (no budget)", ledger.Transaction{
		Amount: 500, Category: "gifts", Description: "No budget set", Date: "2026-10-07",
	})

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
