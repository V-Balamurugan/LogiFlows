package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/logiflows/logiflows/backend/internal/config"
	"github.com/logiflows/logiflows/backend/internal/redis"
)

func TestRedis_UnreachableConfig_FailsGracefully(t *testing.T) {
	cfg := &config.RedisConfig{
		Host: "127.0.0.1",
		Port: 9998, // Unreachable port
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cache, err := redis.New(ctx, cfg)
	if err == nil {
		if cache != nil {
			_ = cache.Close()
		}
		t.Fatalf("expected error connecting to invalid redis port, got nil")
	}
}

func TestRedis_LiveConnection_PingAndClose(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live redis test in short mode")
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cache, err := redis.New(ctx, &cfg.Redis)
	if err != nil {
		t.Skipf("skipping live redis test: redis not reachable: %v", err)
		return
	}
	defer func() { _ = cache.Close() }()

	// 1. Verify Ping
	latency, err := cache.Ping(ctx)
	if err != nil {
		t.Fatalf("cache.Ping failed: %v", err)
	}
	if latency < 0 {
		t.Errorf("expected non-negative latency, got %v", latency)
	}

	// 2. Verify Client accessor
	if cache.Client() == nil {
		t.Errorf("expected non-nil redis Client")
	}
}
