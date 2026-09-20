package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// EmbedMigrations embeds all SQL migration files into the compiled binary.
//
//go:embed *.sql
var EmbedMigrations embed.FS

func init() {
	goose.SetBaseFS(EmbedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		panic(fmt.Sprintf("failed to set goose postgres dialect: %v", err))
	}
}

// RunUp executes pending SQL migrations against the target database.
func RunUp(ctx context.Context, dsn string, log *slog.Logger) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database for migrations: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed before migrations: %w", err)
	}

	log.Info("Executing database migrations...")
	if err := goose.UpContext(ctx, db, "."); err != nil {
		return fmt.Errorf("migration run failed: %w", err)
	}

	log.Info("Database migrations applied successfully")
	return nil
}

// RunStatus checks and logs migration status.
func RunStatus(ctx context.Context, dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database for migration status: %w", err)
	}
	defer db.Close()

	return goose.StatusContext(ctx, db, ".")
}
