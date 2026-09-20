package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/logiflows/logiflows/backend/internal/config"
)

func TestConfig_Load_Default(t *testing.T) {
	// Clear any overrides for a clean test
	os.Clearenv()
	_ = os.Setenv("APP_ENV", "test")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected config to load with defaults, got error: %v", err)
	}

	if cfg.App.Name != "logiflows-api" {
		t.Errorf("expected App.Name to be 'logiflows-api', got %s", cfg.App.Name)
	}
	if cfg.App.Port != 8080 {
		t.Errorf("expected App.Port to be 8080, got %d", cfg.App.Port)
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("expected Database.Port to be 5432, got %d", cfg.Database.Port)
	}
	if cfg.Redis.Port != 6379 {
		t.Errorf("expected Redis.Port to be 6379, got %d", cfg.Redis.Port)
	}
}

func TestConfig_Validate_InvalidEnv(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Env:      "invalid_env",
			Port:     8080,
			LogLevel: "info",
		},
		Database: config.DatabaseConfig{
			Host:         "localhost",
			Port:         5432,
			User:         "postgres",
			Name:         "logiflows_dev",
			MaxOpenConns: 10,
			MaxIdleConns: 5,
		},
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatalf("expected error for invalid APP_ENV, got nil")
	}
}

func TestConfig_Validate_InvalidPort(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Env:      "development",
			Port:     99999, // Out of range
			LogLevel: "info",
		},
		Database: config.DatabaseConfig{
			Host:         "localhost",
			Port:         5432,
			User:         "postgres",
			Name:         "logiflows_dev",
			MaxOpenConns: 10,
			MaxIdleConns: 5,
		},
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatalf("expected error for port 99999, got nil")
	}
}

func TestConfig_Validate_DBIdleConnsExceedsMaxOpen(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Env:      "development",
			Port:     8080,
			LogLevel: "info",
		},
		Database: config.DatabaseConfig{
			Host:         "localhost",
			Port:         5432,
			User:         "postgres",
			Name:         "logiflows_dev",
			MaxOpenConns: 5,
			MaxIdleConns: 10, // Invalid: Idle > MaxOpen
		},
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatalf("expected error when MaxIdleConns > MaxOpenConns, got nil")
	}
}

func TestConfig_DSN_And_RedisAddr(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:            "127.0.0.1",
			Port:            5432,
			User:            "test_user",
			Password:        "test_pass",
			Name:            "test_db",
			SSLMode:         "disable",
			MaxOpenConns:    20,
			MaxIdleConns:    5,
			ConnMaxLifetime: 10 * time.Minute,
		},
		Redis: config.RedisConfig{
			Host: "127.0.0.1",
			Port: 6379,
		},
	}

	expectedDSN := "postgres://test_user:test_pass@127.0.0.1:5432/test_db?sslmode=disable"
	if cfg.Database.DSN() != expectedDSN {
		t.Errorf("expected DSN %s, got %s", expectedDSN, cfg.Database.DSN())
	}

	expectedRedisAddr := "127.0.0.1:6379"
	if cfg.Redis.Addr() != expectedRedisAddr {
		t.Errorf("expected Redis Addr %s, got %s", expectedRedisAddr, cfg.Redis.Addr())
	}
}
