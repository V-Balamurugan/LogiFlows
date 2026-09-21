package integration_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/logiflows/logiflows/backend/internal/config"
	"github.com/logiflows/logiflows/backend/internal/logger"
	"github.com/logiflows/logiflows/backend/migrations"
)

func TestMigrations_RunUp_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database migration integration test in short mode")
	}

	_ = os.Setenv("APP_ENV", "test")
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load configuration: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	log := logger.Init("test", "debug")

	// 1. Run migrations up
	if err := migrations.RunUp(ctx, cfg.Database.DSN(), log); err != nil {
		t.Fatalf("failed to run migrations up: %v", err)
	}

	// 2. Connect via pgx to verify tables exist
	conn, err := pgx.Connect(ctx, cfg.Database.DSN())
	if err != nil {
		t.Fatalf("failed to connect to postgresql: %v", err)
	}
	defer conn.Close(ctx)

	expectedTables := []string{"users", "tenants", "tenant_memberships", "audit_logs"}
	for _, table := range expectedTables {
		var exists bool
		query := `SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' AND table_name = $1
		)`
		if err := conn.QueryRow(ctx, query, table).Scan(&exists); err != nil {
			t.Fatalf("failed to query table existence for %s: %v", table, err)
		}
		if !exists {
			t.Errorf("expected table %s to exist after migration", table)
		}
	}
}
