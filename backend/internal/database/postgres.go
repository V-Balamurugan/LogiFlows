package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/logiflows/logiflows/backend/internal/config"
)

// DB defines the contract for PostgreSQL operations in LogiFlows.
type DB interface {
	Ping(ctx context.Context) (time.Duration, error)
	VerifyPostGIS(ctx context.Context) (string, error)
	Pool() *pgxpool.Pool
	Close()
}

// PostgresDB wraps pgxpool.Pool.
type PostgresDB struct {
	pool *pgxpool.Pool
}

// New initializes a new PostgreSQL connection pool using pgxpool.
func New(ctx context.Context, cfg *config.DatabaseConfig) (*PostgresDB, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database DSN: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.MaxIdleConns)
	poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime

	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connectCtx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize postgres pool: %w", err)
	}

	// Verify connectivity immediately
	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database ping failed on startup: %w", err)
	}

	return &PostgresDB{pool: pool}, nil
}

// Ping executes a lightweight query and returns latency.
func (p *PostgresDB) Ping(ctx context.Context) (time.Duration, error) {
	start := time.Now()
	var result int
	err := p.pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	latency := time.Since(start)
	if err != nil {
		return latency, fmt.Errorf("database ping error: %w", err)
	}
	return latency, nil
}

// VerifyPostGIS queries the PostGIS version to ensure geospatial extensions are functional.
func (p *PostgresDB) VerifyPostGIS(ctx context.Context) (string, error) {
	var version string
	err := p.pool.QueryRow(ctx, "SELECT PostGIS_Full_Version()").Scan(&version)
	if err != nil {
		return "", fmt.Errorf("PostGIS extension query failed: %w", err)
	}
	return version, nil
}

// Pool returns the underlying pgx connection pool.
func (p *PostgresDB) Pool() *pgxpool.Pool {
	return p.pool
}

// Close gracefully closes all active database pool connections.
func (p *PostgresDB) Close() {
	if p.pool != nil {
		p.pool.Close()
	}
}
