package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/logiflows/logiflows/backend/internal/audit"
	"github.com/logiflows/logiflows/backend/internal/auth"
	"github.com/logiflows/logiflows/backend/internal/branches"
	"github.com/logiflows/logiflows/backend/internal/config"
	"github.com/logiflows/logiflows/backend/internal/database"
	"github.com/logiflows/logiflows/backend/internal/deliveries"
	"github.com/logiflows/logiflows/backend/internal/employees"
	"github.com/logiflows/logiflows/backend/internal/health"
	"github.com/logiflows/logiflows/backend/internal/logger"
	"github.com/logiflows/logiflows/backend/internal/memberships"
	"github.com/logiflows/logiflows/backend/internal/middleware"
	"github.com/logiflows/logiflows/backend/internal/parcels"
	"github.com/logiflows/logiflows/backend/internal/redis"
	"github.com/logiflows/logiflows/backend/internal/server"
	"github.com/logiflows/logiflows/backend/internal/tenants"
	"github.com/logiflows/logiflows/backend/internal/transfers"
	"github.com/logiflows/logiflows/backend/internal/users"
	"github.com/logiflows/logiflows/backend/internal/vehicles"
	"github.com/logiflows/logiflows/backend/migrations"
)

func main() {
	// 1. Load and validate configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: Configuration error: %v\n", err)
		os.Exit(1)
	}

	// 2. Initialize structured logging
	log := logger.Init(cfg.App.Env, cfg.App.LogLevel)
	log.Info("Starting LogiFlows API service",
		"service", cfg.App.Name,
		"env", cfg.App.Env,
		"log_level", cfg.App.LogLevel,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 3. Connect to PostgreSQL
	log.Info("Connecting to PostgreSQL database...",
		"host", cfg.Database.Host,
		"port", cfg.Database.Port,
		"db", cfg.Database.Name,
	)
	db, err := database.New(ctx, &cfg.Database)
	if err != nil {
		log.Error("Failed to connect to PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	log.Info("PostgreSQL connection established successfully")

	// 4. Run database migrations
	if err := migrations.RunUp(ctx, cfg.Database.DSN(), log); err != nil {
		log.Error("Database migration failed", "error", err)
		os.Exit(1)
	}

	// 5. Verify PostGIS geospatial extension
	postgisVersion, err := db.VerifyPostGIS(ctx)
	if err != nil {
		log.Warn("PostGIS extension check reported an issue", "error", err)
	} else {
		log.Info("PostGIS extension active", "version", postgisVersion)
	}

	// 6. Connect to Redis
	log.Info("Connecting to Redis cache...",
		"host", cfg.Redis.Host,
		"port", cfg.Redis.Port,
		"db", cfg.Redis.DB,
	)
	cache, err := redis.New(ctx, &cfg.Redis)
	if err != nil {
		log.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := cache.Close(); err != nil {
			log.Warn("Error closing Redis client", "error", err)
		}
	}()
	log.Info("Redis connection established successfully")

	// 7. Setup Repositories, Services, and Handlers
	userRepo := users.NewRepository(db.Pool())
	tenantRepo := tenants.NewRepository(db.Pool())
	membershipRepo := memberships.NewRepository(db.Pool())
	auditRepo := audit.NewRepository(db.Pool())

	tokenRepo := auth.NewRefreshTokenRepository(db.Pool())
	tokenService := auth.NewTokenService(cfg.JWT.Secret, cfg.JWT.AccessExpiry, cfg.JWT.Issuer)
	authService := auth.NewService(db.Pool(), userRepo, tenantRepo, membershipRepo, auditRepo, tokenRepo, tokenService)
	tenantService := tenants.NewService(db.Pool(), tenantRepo, membershipRepo, userRepo, auditRepo)

	healthHandler := health.NewHandler(db, cache)
	authHandler := auth.NewHandler(authService)
	tenantHandler := tenants.NewHandler(tenantService)

	branchRepo := branches.NewRepository(db.Pool())
	branchService := branches.NewService(branchRepo, auditRepo)
	branchHandler := branches.NewHandler(branchService)

	employeeRepo := employees.NewRepository(db.Pool())
	employeeService := employees.NewService(employeeRepo, branchRepo, auditRepo, userRepo, membershipRepo)
	employeeHandler := employees.NewHandler(employeeService)

	vehicleRepo := vehicles.NewRepository(db.Pool())
	vehicleService := vehicles.NewService(vehicleRepo, branchRepo, employeeRepo, auditRepo)
	vehicleHandler := vehicles.NewHandler(vehicleService)

	// Phase 4: Parcel, Delivery, and Transfer Handlers
	parcelRepo := parcels.NewRepository(db.Pool())
	parcelService := parcels.NewService(parcelRepo, branchRepo)
	parcelHandler := parcels.NewHandler(parcelService)

	deliveryRepo := deliveries.NewRepository(db.Pool())
	deliveryService := deliveries.NewService(deliveryRepo, parcelRepo, employeeRepo)
	deliveryHandler := deliveries.NewHandler(deliveryService, employeeRepo)

	transferRepo := transfers.NewRepository(db.Pool())
	transferService := transfers.NewService(transferRepo, branchRepo)
	transferHandler := transfers.NewHandler(transferService)

	authMiddleware := middleware.Auth(tokenService, userRepo)
	tenantMiddleware := middleware.TenantContext(membershipRepo, tenantRepo)

	router := server.SetupRouter(server.RouterParams{
		Cfg:              cfg,
		Log:              log,
		HealthHandler:    healthHandler,
		AuthHandler:      authHandler,
		TenantHandler:    tenantHandler,
		BranchHandler:    branchHandler,
		EmployeeHandler:  employeeHandler,
		VehicleHandler:   vehicleHandler,
		ParcelHandler:    parcelHandler,
		DeliveryHandler:  deliveryHandler,
		TransferHandler:  transferHandler,
		AuthMiddleware:   authMiddleware,
		TenantMiddleware: tenantMiddleware,
	})

	// 8. Initialize HTTP Server
	srv := server.New(cfg, log, router)
	srvErr := srv.Start()

	// 9. Listen for OS interrupt signals for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	select {
	case err := <-srvErr:
		if err != nil {
			log.Error("HTTP server failed to run", "error", err)
			os.Exit(1)
		}
	case sig := <-quit:
		log.Info("Received shutdown signal", "signal", sig.String())
	}

	// 10. Execute Graceful Shutdown
	log.Info("Commencing graceful service shutdown (10s timeout)...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("Server forced to shutdown with error", "error", err)
	} else {
		log.Info("HTTP server stopped gracefully")
	}

	log.Info("LogiFlows API service shutdown complete")
}
