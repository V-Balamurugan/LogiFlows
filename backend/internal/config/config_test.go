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

func TestConfig_Load_DBAliases(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("APP_ENV", "test")
	_ = os.Setenv("DB_HOST", "db-host-alias")
	_ = os.Setenv("DB_PORT", "5433")
	_ = os.Setenv("DB_USER", "alias_user")
	_ = os.Setenv("DB_PASSWORD", "alias_password")
	_ = os.Setenv("DB_NAME", "alias_db")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected config to load with DB_* aliases, got error: %v", err)
	}

	if cfg.Database.Host != "db-host-alias" {
		t.Errorf("expected Database.Host to be 'db-host-alias', got %s", cfg.Database.Host)
	}
	if cfg.Database.Port != 5433 {
		t.Errorf("expected Database.Port to be 5433, got %d", cfg.Database.Port)
	}
	if cfg.Database.User != "alias_user" {
		t.Errorf("expected Database.User to be 'alias_user', got %s", cfg.Database.User)
	}
	if cfg.Database.Password != "alias_password" {
		t.Errorf("expected Database.Password to be 'alias_password', got %s", cfg.Database.Password)
	}
	if cfg.Database.Name != "alias_db" {
		t.Errorf("expected Database.Name to be 'alias_db', got %s", cfg.Database.Name)
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
		JWT: config.JWTConfig{
			Secret:       "super-secret-logiflows-dev-jwt-key-min32chars!",
			AccessExpiry: 24 * time.Hour,
			Issuer:       "logiflows-api",
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

func TestConfig_Validate_JWTSecretTooShort(t *testing.T) {
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
			MaxOpenConns: 10,
			MaxIdleConns: 5,
		},
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
		JWT: config.JWTConfig{
			Secret:       "too-short", // Less than 32 chars
			AccessExpiry: 1 * time.Hour,
			Issuer:       "logiflows-api",
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatalf("expected error when JWT secret is less than 32 characters, got nil")
	}
}

func TestConfig_Validate_JWTInvalidExpiry(t *testing.T) {
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
			MaxOpenConns: 10,
			MaxIdleConns: 5,
		},
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
		JWT: config.JWTConfig{
			Secret:       "super-secret-logiflows-dev-jwt-key-min32chars!",
			AccessExpiry: 0, // Invalid expiry
			Issuer:       "logiflows-api",
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatalf("expected error when JWT AccessExpiry <= 0, got nil")
	}
}
