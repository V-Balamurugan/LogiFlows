package integration_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/logiflows/logiflows/backend/internal/config"
	"github.com/logiflows/logiflows/backend/internal/database"
	"github.com/logiflows/logiflows/backend/internal/logger"
	"github.com/logiflows/logiflows/backend/migrations"
)

func TestPostgreSQL_ConnectionAndPostGIS(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Connect to PostgreSQL
	db, err := database.New(ctx, &cfg.Database)
	if err != nil {
		t.Fatalf("failed to connect to PostgreSQL: %v (is Docker Compose running?)", err)
	}
	defer db.Close()

	// 2. Test Ping latency
	latency, err := db.Ping(ctx)
	if err != nil {
		t.Fatalf("expected ping to succeed, got error: %v", err)
	}
	if latency <= 0 {
		t.Errorf("expected positive latency, got %v", latency)
	}
	t.Logf("PostgreSQL Ping Latency: %v", latency)

	// 3. Run Goose Migrations (enables PostGIS and uuid-ossp)
	log := logger.Init("test", "error")
	if err := migrations.RunUp(ctx, cfg.Database.DSN(), log); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// 4. Verify PostGIS Extension
	postgisVersion, err := db.VerifyPostGIS(ctx)
	if err != nil {
		t.Fatalf("failed to verify PostGIS: %v", err)
	}
	if !strings.Contains(strings.ToUpper(postgisVersion), "POSTGIS") {
		t.Errorf("expected PostGIS version string to contain 'POSTGIS', got: %s", postgisVersion)
	}
	t.Logf("PostGIS Full Version: %s", postgisVersion)

	// 5. Verify uuid-ossp extension is installed
	var extName string
	err = db.Pool().QueryRow(ctx, "SELECT extname FROM pg_extension WHERE extname = 'uuid-ossp'").Scan(&extName)
	if err != nil {
		t.Fatalf("expected uuid-ossp extension to be installed: %v", err)
	}
	if extName != "uuid-ossp" {
		t.Errorf("expected 'uuid-ossp', got '%s'", extName)
	}
}
