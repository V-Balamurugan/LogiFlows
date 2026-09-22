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

	expectedTables := []string{"users", "tenants", "tenant_memberships", "audit_logs", "refresh_tokens"}
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

	// 3. Verify users.email_verified column exists
	var emailVerifiedColExists bool
	colQuery := `SELECT EXISTS (
		SELECT FROM information_schema.columns 
		WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'email_verified'
	)`
	if err := conn.QueryRow(ctx, colQuery).Scan(&emailVerifiedColExists); err != nil {
		t.Fatalf("failed to query column existence for users.email_verified: %v", err)
	}
	if !emailVerifiedColExists {
		t.Errorf("expected column email_verified to exist on users table")
	}
}

func TestMigrations_RollbackAndReapply(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database rollback integration test in short mode")
	}

	_ = os.Setenv("APP_ENV", "test")
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load configuration: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	log := logger.Init("test", "debug")

	// 1. Rollback migration 00003
	defer func() {
		_ = migrations.RunUp(context.Background(), cfg.Database.DSN(), log)
	}()

	if err := migrations.RunDown(ctx, cfg.Database.DSN(), log); err != nil {
		t.Fatalf("failed to rollback migration: %v", err)
	}

	conn, err := pgx.Connect(ctx, cfg.Database.DSN())
	if err != nil {
		t.Fatalf("failed to connect to postgresql: %v", err)
	}
	defer conn.Close(ctx)

	// Verify refresh_tokens is dropped
	var refreshTokensExists bool
	query := `SELECT EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'refresh_tokens'
	)`
	if err := conn.QueryRow(ctx, query).Scan(&refreshTokensExists); err != nil {
		t.Fatalf("failed to check refresh_tokens existence: %v", err)
	}
	if refreshTokensExists {
		t.Errorf("expected refresh_tokens table to be dropped after rollback")
	}

	// 2. Re-apply migration 00003
	if err := migrations.RunUp(ctx, cfg.Database.DSN(), log); err != nil {
		t.Fatalf("failed to re-apply migrations: %v", err)
	}

	// Verify refresh_tokens is recreated
	if err := conn.QueryRow(ctx, query).Scan(&refreshTokensExists); err != nil {
		t.Fatalf("failed to check refresh_tokens re-creation: %v", err)
	}
	if !refreshTokensExists {
		t.Errorf("expected refresh_tokens table to exist after re-applying migration")
	}
}
