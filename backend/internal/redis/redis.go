package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/logiflows/logiflows/backend/internal/config"
	"github.com/redis/go-redis/v9"
)

// Cache defines the contract for Redis cache operations.
type Cache interface {
	Ping(ctx context.Context) (time.Duration, error)
	Client() *redis.Client
	Close() error
}

// RedisCache wraps the go-redis Client.
type RedisCache struct {
	client *redis.Client
}

// New initializes and validates a new Redis client connection.
func New(ctx context.Context, cfg *config.RedisConfig) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis ping failed on startup: %w", err)
	}

	return &RedisCache{client: client}, nil
}

// Ping checks Redis connectivity and records execution duration.
func (r *RedisCache) Ping(ctx context.Context) (time.Duration, error) {
	start := time.Now()
	err := r.client.Ping(ctx).Err()
	latency := time.Since(start)
	if err != nil {
		return latency, fmt.Errorf("redis ping error: %w", err)
	}
	return latency, nil
}

// Client returns the underlying go-redis Client instance.
func (r *RedisCache) Client() *redis.Client {
	return r.client
}

// Close terminates active connections to Redis.
func (r *RedisCache) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}
