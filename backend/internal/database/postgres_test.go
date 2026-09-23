package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/logiflows/logiflows/backend/internal/config"
	"github.com/logiflows/logiflows/backend/internal/database"
)

func TestPostgres_InvalidConfig_FailsGracefully(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Host:            "127.0.0.1",
		Port:            9999, // Unreachable port
		User:            "invalid_user",
		Password:        "invalid_pass",
		Name:            "invalid_db",
		SSLMode:         "disable",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 5 * time.Minute,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	db, err := database.New(ctx, cfg)
	if err == nil {
		if db != nil {
			db.Close()
		}
		t.Fatalf("expected error for unreachable postgres instance, got nil")
	}
}

func TestPostgres_LiveConnection_PingAndVerify(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live database test in short mode")
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := database.New(ctx, &cfg.Database)
	if err != nil {
		t.Skipf("skipping live database test: postgres not reachable: %v", err)
		return
	}
	defer db.Close()

	// 1. Verify Ping
	latency, err := db.Ping(ctx)
	if err != nil {
		t.Fatalf("db.Ping failed: %v", err)
	}
	if latency < 0 {
		t.Errorf("expected positive latency, got %v", latency)
	}

	// 2. Verify Pool accessor
	if db.Pool() == nil {
		t.Errorf("expected non-nil pgx pool")
	}

	// 3. Verify PostGIS
	version, err := db.VerifyPostGIS(ctx)
	if err != nil {
		t.Logf("VerifyPostGIS returned notice: %v", err)
	} else if version == "" {
		t.Errorf("expected non-empty PostGIS version string")
	}
}
