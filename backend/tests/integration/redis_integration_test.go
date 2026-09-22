package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/logiflows/logiflows/backend/internal/config"
	"github.com/logiflows/logiflows/backend/internal/redis"
)

func TestRedis_ConnectionAndPing(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Connect to Redis
	cache, err := redis.New(ctx, &cfg.Redis)
	if err != nil {
		t.Fatalf("failed to connect to Redis: %v (is Docker Compose running?)", err)
	}
	defer cache.Close()

	// 2. Test Ping latency
	latency, err := cache.Ping(ctx)
	if err != nil {
		t.Fatalf("expected redis ping to succeed, got error: %v", err)
	}
	if latency < 0 {
		t.Errorf("expected non-negative latency, got %v", latency)
	}
	t.Logf("Redis Ping Latency: %v", latency)

	// 3. Test Set / Get / Del key operations
	testKey := "logiflows:phase0:test_key"
	testVal := "phase-0-active"

	err = cache.Client().Set(ctx, testKey, testVal, 10*time.Second).Err()
	if err != nil {
		t.Fatalf("failed to set test key in redis: %v", err)
	}

	val, err := cache.Client().Get(ctx, testKey).Result()
	if err != nil {
		t.Fatalf("failed to get test key from redis: %v", err)
	}
	if val != testVal {
		t.Errorf("expected '%s', got '%s'", testVal, val)
	}

	// Clean up
	_ = cache.Client().Del(ctx, testKey)
}
