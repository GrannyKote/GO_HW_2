package ledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"ledger/internal/cache"

	"github.com/redis/go-redis/v9"
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

var sqlDB *sql.DB

const budgetsCacheKey = "budgets:all"

// Reset clears tables and cache. Intended for tests.
func Reset() error {
	if sqlDB == nil {
		return nil
	}
	if _, err := sqlDB.Exec(`TRUNCATE TABLE expenses, budgets RESTART IDENTITY`); err != nil {
		return fmt.Errorf("truncate tables: %w", err)
	}
	if c := cache.Client(); c != nil {
		ctx := context.Background()
		_ = c.Del(ctx, budgetsCacheKey).Err()
		keys, _ := c.Keys(ctx, "report:summary:*").Result()
		if len(keys) > 0 {
			_ = c.Del(ctx, keys...).Err()
		}
	}
	return nil
}

// SetBudget adds or updates a budget for the given category.
func SetBudget(b Budget) error {
	if err := b.Validate(); err != nil {
		return err
	}
	if sqlDB == nil {
		return errors.New("database not initialized")
	}

	_, err := sqlDB.Exec(`
		INSERT INTO budgets (category, limit_amount)
		VALUES ($1, $2)
		ON CONFLICT (category) DO UPDATE SET limit_amount = EXCLUDED.limit_amount`,
		b.Category, b.Limit)
	if err != nil {
		return fmt.Errorf("upsert budget: %w", err)
	}

	if c := cache.Client(); c != nil {
		_ = c.Del(context.Background(), budgetsCacheKey).Err()
	}
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

// AddTransaction persists a transaction after validation and budget checks.
func AddTransaction(tx Transaction) (Transaction, error) {
	if err := tx.Validate(); err != nil {
		return Transaction{}, err
	}
	if sqlDB == nil {
		return Transaction{}, errors.New("database not initialized")
	}

	var limit sql.NullFloat64
	err := sqlDB.QueryRow(
		`SELECT limit_amount FROM budgets WHERE category = $1`, tx.Category,
	).Scan(&limit)
	if err != nil && err != sql.ErrNoRows {
		return Transaction{}, fmt.Errorf("load budget: %w", err)
	}

	if limit.Valid {
		var spent float64
		if err := sqlDB.QueryRow(
			`SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE category = $1`, tx.Category,
		).Scan(&spent); err != nil {
			return Transaction{}, fmt.Errorf("sum expenses: %w", err)
		}
		if spent+tx.Amount > limit.Float64 {
			return Transaction{}, ErrBudgetExceeded
		}
	}

	err = sqlDB.QueryRow(`
		INSERT INTO expenses (amount, category, description, date)
		VALUES ($1, $2, $3, $4::date)
		RETURNING id`,
		tx.Amount, tx.Category, tx.Description, tx.Date,
	).Scan(&tx.ID)
	if err != nil {
		return Transaction{}, fmt.Errorf("insert expense: %w", err)
	}
	return tx, nil
}

// ListTransactions returns all expenses ordered by date and id descending.
func ListTransactions() ([]Transaction, error) {
	if sqlDB == nil {
		return nil, errors.New("database not initialized")
	}
	rows, err := sqlDB.Query(`
		SELECT id, amount, category, description, date::text
		FROM expenses
		ORDER BY date DESC, id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list expenses: %w", err)
	}
	defer rows.Close()

	var out []Transaction
	for rows.Next() {
		var tx Transaction
		if err := rows.Scan(&tx.ID, &tx.Amount, &tx.Category, &tx.Description, &tx.Date); err != nil {
			return nil, fmt.Errorf("scan expense: %w", err)
		}
		out = append(out, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate expenses: %w", err)
	}
	if out == nil {
		out = []Transaction{}
	}
	return out, nil
}

// ListBudgets returns all budgets ordered by category.
func ListBudgets() ([]Budget, error) {
	if sqlDB == nil {
		return nil, errors.New("database not initialized")
	}

	ctx := context.Background()
	if c := cache.Client(); c != nil {
		val, err := c.Get(ctx, budgetsCacheKey).Result()
		if err == nil {
			var cached []Budget
			if json.Unmarshal([]byte(val), &cached) == nil {
				return cached, nil
			}
		} else if err != redis.Nil {
			return nil, fmt.Errorf("redis get budgets: %w", err)
		}
	}

	rows, err := sqlDB.Query(`
		SELECT category, limit_amount
		FROM budgets
		ORDER BY category`)
	if err != nil {
		return nil, fmt.Errorf("list budgets: %w", err)
	}
	defer rows.Close()

	var out []Budget
	for rows.Next() {
		var b Budget
		if err := rows.Scan(&b.Category, &b.Limit); err != nil {
			return nil, fmt.Errorf("scan budget: %w", err)
		}
		out = append(out, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate budgets: %w", err)
	}
	if out == nil {
		out = []Budget{}
	}

	if c := cache.Client(); c != nil {
		if raw, err := json.Marshal(out); err == nil {
			_ = c.Set(ctx, budgetsCacheKey, raw, cache.BudgetsListTTL).Err()
		}
	}
	return out, nil
}

// String formats a transaction for console output.
func (t Transaction) String() string {
	return fmt.Sprintf("ID=%d Amount=%.2f Category=%q Description=%q Date=%s",
		t.ID, t.Amount, t.Category, t.Description, t.Date)
}
