package integration_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/logiflows/logiflows/backend/internal/config"
	"github.com/logiflows/logiflows/backend/internal/logger"
	"github.com/logiflows/logiflows/backend/migrations"
)

// TestMain initializes the test environment and ensures that all database
// migrations are applied before any integration tests are executed.
func TestMain(m *testing.M) {
	flag.Parse()

	_ = os.Setenv("APP_ENV", "test")
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: TestMain failed to load config: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log := logger.Init("test", "error")

	if !testing.Short() {
		if err := migrations.RunUp(ctx, cfg.Database.DSN(), log); err != nil {
			fmt.Fprintf(os.Stderr, "FATAL: TestMain failed to apply database migrations: %v\n", err)
			os.Exit(1)
		}
	}

	code := m.Run()
	os.Exit(code)
}
