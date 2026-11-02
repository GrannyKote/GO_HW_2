package ledger

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ledger/internal/cache"

	"github.com/redis/go-redis/v9"
)

// ReportCategoryRow is spending aggregated for one category.
type ReportCategoryRow struct {
	Category string  `json:"category"`
	Total    float64 `json:"total"`
}

// ReportSummary is an expenses report for a date range.
type ReportSummary struct {
	From       string              `json:"from"`
	To         string              `json:"to"`
	Categories []ReportCategoryRow `json:"categories"`
	GrandTotal float64             `json:"grand_total"`
}

// GetReportSummary returns aggregated expenses by category for [from, to] (YYYY-MM-DD).
// Results are cached in Redis for a short TTL.
func GetReportSummary(ctx context.Context, from, to string) (ReportSummary, error) {
	if sqlDB == nil {
		return ReportSummary{}, fmt.Errorf("database not initialized")
	}

	key := fmt.Sprintf("report:summary:%s:%s", from, to)
	if c := cache.Client(); c != nil {
		val, err := c.Get(ctx, key).Result()
		if err == nil {
			var cached ReportSummary
			if json.Unmarshal([]byte(val), &cached) == nil {
				return cached, nil
			}
		} else if err != redis.Nil {
			return ReportSummary{}, fmt.Errorf("redis get: %w", err)
		}
	}

	summary, err := loadReportFromDB(from, to)
	if err != nil {
		return ReportSummary{}, err
	}

	if c := cache.Client(); c != nil {
		raw, err := json.Marshal(summary)
		if err == nil {
			_ = c.Set(ctx, key, raw, cache.ReportSummaryTTL).Err()
		}
	}

	return summary, nil
}

func loadReportFromDB(from, to string) (ReportSummary, error) {
	rows, err := sqlDB.Query(`
		SELECT category, COALESCE(SUM(amount), 0)
		FROM expenses
		WHERE date >= $1::date AND date <= $2::date
		GROUP BY category
		ORDER BY category`,
		from, to)
	if err != nil {
		return ReportSummary{}, fmt.Errorf("query report: %w", err)
	}
	defer rows.Close()

	var summary ReportSummary
	summary.From = from
	summary.To = to

	for rows.Next() {
		var row ReportCategoryRow
		if err := rows.Scan(&row.Category, &row.Total); err != nil {
			return ReportSummary{}, fmt.Errorf("scan report row: %w", err)
		}
		summary.GrandTotal += row.Total
		summary.Categories = append(summary.Categories, row)
	}
	if err := rows.Err(); err != nil {
		return ReportSummary{}, fmt.Errorf("iterate report: %w", err)
	}
	if summary.Categories == nil {
		summary.Categories = []ReportCategoryRow{}
	}
	return summary, nil
}

// ValidateReportRange checks date strings before querying.
func ValidateReportRange(from, to string) error {
	if from == "" || to == "" {
		return fmt.Errorf("from and to query parameters are required (YYYY-MM-DD)")
	}
	if _, err := time.Parse("2006-01-02", from); err != nil {
		return fmt.Errorf("invalid from date: use YYYY-MM-DD")
	}
	if _, err := time.Parse("2006-01-02", to); err != nil {
		return fmt.Errorf("invalid to date: use YYYY-MM-DD")
	}
	return nil
}
