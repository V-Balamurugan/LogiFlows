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

	expectedTables := []string{
		"users", "tenants", "tenant_memberships", "audit_logs", "refresh_tokens",
		"branches", "employees", "vehicles", "vehicle_assignments", "tenant_employee_sequences",
		"tenant_parcel_sequences", "parcels", "parcel_status_history", "parcel_custody_events",
		"branch_transfers", "branch_transfer_parcels", "delivery_tasks", "delivery_attempts", "delivery_proofs",
		"tenant_customer_sequences", "customers",
	}
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

	// 4. Verify Phase 3 employee availability_status column exists
	var empAvailColExists bool
	empColQuery := `SELECT EXISTS (
		SELECT FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = 'employees' AND column_name = 'availability_status'
	)`
	if err := conn.QueryRow(ctx, empColQuery).Scan(&empAvailColExists); err != nil {
		t.Fatalf("failed to query column existence for employees.availability_status: %v", err)
	}
	if !empAvailColExists {
		t.Errorf("expected column availability_status to exist on employees table")
	}

	// 5. Verify Phase 4 parcels tracking_number and qr_code_payload column exists
	var parcelTrackingColExists bool
	parcelColQuery := `SELECT EXISTS (
		SELECT FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = 'parcels' AND column_name = 'tracking_number'
	)`
	if err := conn.QueryRow(ctx, parcelColQuery).Scan(&parcelTrackingColExists); err != nil {
		t.Fatalf("failed to query column existence for parcels.tracking_number: %v", err)
	}
	if !parcelTrackingColExists {
		t.Errorf("expected column tracking_number to exist on parcels table")
	}

	// 6. Verify Phase 4 customer management columns exist
	var parcelCustomerColExists bool
	parcelCustColQuery := `SELECT EXISTS (
		SELECT FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = 'parcels' AND column_name = 'sender_customer_id'
	)`
	if err := conn.QueryRow(ctx, parcelCustColQuery).Scan(&parcelCustomerColExists); err != nil {
		t.Fatalf("failed to query column existence for parcels.sender_customer_id: %v", err)
	}
	if !parcelCustomerColExists {
		t.Errorf("expected column sender_customer_id to exist on parcels table")
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

	// 1. Ensure migrations are up first
	if err := migrations.RunUp(ctx, cfg.Database.DSN(), log); err != nil {
		t.Fatalf("failed to ensure migrations are up: %v", err)
	}

	// 2. Rollback latest migration (00009_create_customer_management)
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

	// Verify tenant_customer_sequences table is dropped after rollback of migration 00009
	var seqTableExists bool
	query := `SELECT EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'tenant_customer_sequences'
	)`
	if err := conn.QueryRow(ctx, query).Scan(&seqTableExists); err != nil {
		t.Fatalf("failed to check tenant_customer_sequences existence: %v", err)
	}
	if seqTableExists {
		t.Errorf("expected tenant_customer_sequences table to be dropped after rollback")
	}

	// 3. Re-apply migration 00009
	if err := migrations.RunUp(ctx, cfg.Database.DSN(), log); err != nil {
		t.Fatalf("failed to re-apply migrations: %v", err)
	}

	// Verify tenant_customer_sequences table is recreated
	if err := conn.QueryRow(ctx, query).Scan(&seqTableExists); err != nil {
		t.Fatalf("failed to check tenant_customer_sequences re-creation: %v", err)
	}
	if !seqTableExists {
		t.Errorf("expected tenant_customer_sequences table to exist after re-applying migration")
	}
}
