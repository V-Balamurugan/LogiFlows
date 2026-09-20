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
			Host:            getEnv("DATABASE_HOST", "localhost"),
			Port:            getEnvAsInt("DATABASE_PORT", 5432),
			User:            getEnv("DATABASE_USER", "postgres"),
			Password:        getEnv("DATABASE_PASSWORD", "postgres_dev_password"),
			Name:            getEnv("DATABASE_NAME", "logiflows_dev"),
			SSLMode:         getEnv("DATABASE_SSLMODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DATABASE_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DATABASE_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DATABASE_CONN_MAX_LIFETIME", 15*time.Minute),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
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
