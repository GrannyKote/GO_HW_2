package ledger

import (
	"context"
	"fmt"
	"log"

	"ledger/internal/cache"
	"ledger/internal/db"
)

// Init opens PostgreSQL and Redis. Call once at process startup (Ledger or Gateway).
func Init(ctx context.Context) error {
	if sqlDB != nil && cache.Client() != nil {
		return nil
	}
	dsn := db.DSNFromEnv()
	conn, err := db.Open(dsn)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	sqlDB = conn
	log.Println("Ledger: connected to PostgreSQL")

	if _, err := cache.Init(ctx); err != nil {
		_ = sqlDB.Close()
		sqlDB = nil
		return fmt.Errorf("redis: %w", err)
	}
	log.Println("Ledger: connected to Redis")
	return nil
}

// Close releases database and Redis connections.
func Close() {
	if sqlDB != nil {
		_ = sqlDB.Close()
		sqlDB = nil
	}
	cache.Close()
}
