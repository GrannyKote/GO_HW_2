package cache

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var client *redis.Client

// Init connects to Redis using environment variables.
func Init(ctx context.Context) (*redis.Client, error) {
	addr := envOr("REDIS_ADDR", "localhost:6379")
	dbNum, err := strconv.Atoi(envOr("REDIS_DB", "0"))
	if err != nil {
		return nil, fmt.Errorf("REDIS_DB: %w", err)
	}

	c := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       dbNum,
	})
	if err := c.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	client = c
	return c, nil
}

// Client returns the initialized Redis client.
func Client() *redis.Client {
	return client
}

// Close shuts down the Redis client.
func Close() {
	if client != nil {
		_ = client.Close()
		client = nil
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

const (
	ReportSummaryTTL = 30 * time.Second
	BudgetsListTTL   = 20 * time.Second
)
