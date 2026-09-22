package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config represents the application configuration schema.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
}

// AppConfig holds general application and HTTP server settings.
type AppConfig struct {
	Name            string
	Env             string
	Host            string
	Port            int
	LogLevel        string
	RequestIDHeader string
}

// DatabaseConfig holds PostgreSQL connection and pooling settings.
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// JWTConfig holds JWT signing and expiration settings.
type JWTConfig struct {
	Secret       string
	AccessExpiry time.Duration
	Issuer       string
}

// DSN returns the formatted PostgreSQL connection string.
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User,
		d.Password,
		d.Host,
		d.Port,
		d.Name,
		d.SSLMode,
	)
}

// RedisAddr returns the formatted host:port string for Redis.
func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// Load loads configuration from environment variables, optionally reading from .env if present.
func Load() (*Config, error) {
	// Attempt to load .env file if it exists, ignore if missing (standard 12-factor practice)
	_ = godotenv.Load()

	cfg := &Config{
		App: AppConfig{
			Name:            getEnv("APP_NAME", "logiflows-api"),
			Env:             strings.ToLower(getEnv("APP_ENV", "development")),
			Host:            getEnv("APP_HOST", "0.0.0.0"),
			Port:            getEnvAsInt("APP_PORT", 8080),
			LogLevel:        strings.ToLower(getEnv("LOG_LEVEL", "info")),
			RequestIDHeader: getEnv("REQUEST_ID_HEADER", "X-Request-ID"),
		},
		Database: DatabaseConfig{
			Host:            getEnvAny("localhost", "DATABASE_HOST", "DB_HOST"),
			Port:            getEnvAnyAsInt(5432, "DATABASE_PORT", "DB_PORT"),
			User:            getEnvAny("postgres", "DATABASE_USER", "DB_USER"),
			Password:        getEnvAny("postgres_dev_password", "DATABASE_PASSWORD", "DB_PASSWORD"),
			Name:            getEnvAny("logiflows_dev", "DATABASE_NAME", "DB_NAME"),
			SSLMode:         getEnvAny("disable", "DATABASE_SSLMODE", "DB_SSLMODE"),
			MaxOpenConns:    getEnvAnyAsInt(25, "DATABASE_MAX_OPEN_CONNS", "DB_MAX_OPEN_CONNS"),
			MaxIdleConns:    getEnvAnyAsInt(5, "DATABASE_MAX_IDLE_CONNS", "DB_MAX_IDLE_CONNS"),
			ConnMaxLifetime: getEnvAsDuration("DATABASE_CONN_MAX_LIFETIME", 15*time.Minute),
		},
		Redis: RedisConfig{
			Host:     getEnvAny("localhost", "REDIS_HOST"),
			Port:     getEnvAnyAsInt(6379, "REDIS_PORT"),
			Password: getEnvAny("", "REDIS_PASSWORD"),
			DB:       getEnvAnyAsInt(0, "REDIS_DB"),
		},
		JWT: JWTConfig{
			Secret:       getEnv("JWT_SECRET", "super-secret-logiflows-dev-jwt-key-min32chars!"),
			AccessExpiry: getEnvAsDuration("JWT_ACCESS_EXPIRY", 24*time.Hour),
			Issuer:       getEnv("JWT_ISSUER", "logiflows-api"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// Validate verifies that required fields and ranges are valid.
func (c *Config) Validate() error {
	// Validate App Environment
	validEnvs := map[string]bool{
		"development": true,
		"staging":     true,
		"production":  true,
		"test":        true,
	}
	if !validEnvs[c.App.Env] {
		return fmt.Errorf("invalid APP_ENV '%s': must be one of development, staging, production, test", c.App.Env)
	}

	// Validate App Port
	if c.App.Port < 1 || c.App.Port > 65535 {
		return fmt.Errorf("invalid APP_PORT %d: must be between 1 and 65535", c.App.Port)
	}

	// Validate Log Level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[c.App.LogLevel] {
		return fmt.Errorf("invalid LOG_LEVEL '%s': must be one of debug, info, warn, error", c.App.LogLevel)
	}

	// Validate Database Config
	if strings.TrimSpace(c.Database.Host) == "" {
		return fmt.Errorf("DATABASE_HOST cannot be empty")
	}
	if c.Database.Port < 1 || c.Database.Port > 65535 {
		return fmt.Errorf("invalid DATABASE_PORT %d: must be between 1 and 65535", c.Database.Port)
	}
	if strings.TrimSpace(c.Database.User) == "" {
		return fmt.Errorf("DATABASE_USER cannot be empty")
	}
	if strings.TrimSpace(c.Database.Name) == "" {
		return fmt.Errorf("DATABASE_NAME cannot be empty")
	}
	if c.Database.MaxOpenConns < 1 {
		return fmt.Errorf("DATABASE_MAX_OPEN_CONNS must be at least 1")
	}
	if c.Database.MaxIdleConns < 0 {
		return fmt.Errorf("DATABASE_MAX_IDLE_CONNS cannot be negative")
	}
	if c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		return fmt.Errorf("DATABASE_MAX_IDLE_CONNS cannot exceed DATABASE_MAX_OPEN_CONNS")
	}

	// Validate Redis Config
	if strings.TrimSpace(c.Redis.Host) == "" {
		return fmt.Errorf("REDIS_HOST cannot be empty")
	}
	if c.Redis.Port < 1 || c.Redis.Port > 65535 {
		return fmt.Errorf("invalid REDIS_PORT %d: must be between 1 and 65535", c.Redis.Port)
	}
	if c.Redis.DB < 0 {
		return fmt.Errorf("REDIS_DB cannot be negative")
	}

	// Validate JWT Config
	if len(strings.TrimSpace(c.JWT.Secret)) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long for security")
	}
	if c.JWT.AccessExpiry <= 0 {
		return fmt.Errorf("JWT_ACCESS_EXPIRY must be greater than zero")
	}

	return nil
}

// Helpers
func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && strings.TrimSpace(val) != "" {
		return strings.TrimSpace(val)
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	if valStr, ok := os.LookupEnv(key); ok {
		if val, err := strconv.Atoi(strings.TrimSpace(valStr)); err == nil {
			return val
		}
	}
	return defaultVal
}

func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	if valStr, ok := os.LookupEnv(key); ok {
		if val, err := time.ParseDuration(strings.TrimSpace(valStr)); err == nil {
			return val
		}
	}
	return defaultVal
}

func getEnvAny(defaultVal string, keys ...string) string {
	for _, key := range keys {
		if val, ok := os.LookupEnv(key); ok && strings.TrimSpace(val) != "" {
			return strings.TrimSpace(val)
		}
	}
	return defaultVal
}

func getEnvAnyAsInt(defaultVal int, keys ...string) int {
	for _, key := range keys {
		if valStr, ok := os.LookupEnv(key); ok && strings.TrimSpace(valStr) != "" {
			if val, err := strconv.Atoi(strings.TrimSpace(valStr)); err == nil {
				return val
			}
		}
	}
	return defaultVal
}
