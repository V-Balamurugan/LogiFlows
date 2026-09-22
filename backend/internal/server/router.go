package server

import (
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/logiflows/logiflows/backend/internal/auth"
	"github.com/logiflows/logiflows/backend/internal/config"
	"github.com/logiflows/logiflows/backend/internal/health"
	"github.com/logiflows/logiflows/backend/internal/memberships"
	"github.com/logiflows/logiflows/backend/internal/middleware"
	"github.com/logiflows/logiflows/backend/internal/response"
	"github.com/logiflows/logiflows/backend/internal/tenants"
)

// RouterParams encapsulates dependencies required to wire all application routes.
type RouterParams struct {
	Cfg              *config.Config
	Log              *slog.Logger
	HealthHandler    *health.Handler
	AuthHandler      *auth.Handler
	TenantHandler    *tenants.Handler
	AuthMiddleware   gin.HandlerFunc
	TenantMiddleware gin.HandlerFunc
}

// SetupRouter configures the Gin engine, global middleware, and API routes.
func SetupRouter(params RouterParams) *gin.Engine {
	if strings.ToLower(params.Cfg.App.Env) == "production" || strings.ToLower(params.Cfg.App.Env) == "staging" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Global Middlewares
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID(params.Cfg.App.RequestIDHeader))
	r.Use(middleware.StructuredLogger(params.Log))
	r.Use(middleware.Recovery(params.Log))

	// Handle 404 Not Found uniformly
	r.NoRoute(func(c *gin.Context) {
		response.NotFound(c, "The requested resource was not found")
	})

	// Handle 405 Method Not Allowed
	r.NoMethod(func(c *gin.Context) {
		response.Error(c, 405, "METHOD_NOT_ALLOWED", "HTTP method is not supported on this endpoint", nil)
	})

	// API v1 Route Group
	v1 := r.Group("/api/v1")
	{
		// Observability (Public)
		if params.HealthHandler != nil {
			v1.GET("/health", params.HealthHandler.Liveness)
			v1.GET("/readiness", params.HealthHandler.Readiness)
		}

		// Authentication (Public)
		if params.AuthHandler != nil {
			authRoutes := v1.Group("/auth")
			{
				authRoutes.POST("/register", params.AuthHandler.Register)
				authRoutes.POST("/login", params.AuthHandler.Login)
				authRoutes.POST("/refresh", params.AuthHandler.Refresh)
			}
		}

		// Authenticated Routes
		if params.AuthMiddleware != nil {
			authed := v1.Group("")
			authed.Use(params.AuthMiddleware)
			{
				if params.AuthHandler != nil {
					authed.GET("/auth/me", params.AuthHandler.Me)
					authed.POST("/auth/logout", params.AuthHandler.Logout)
				}

				// Tenant operations
				if params.TenantHandler != nil {
					authed.POST("/tenants", params.TenantHandler.Create)
					authed.GET("/tenants", params.TenantHandler.List)

					// Tenant-scoped routes with multi-tenant isolation enforcement
					if params.TenantMiddleware != nil {
						tenantScoped := authed.Group("/tenants/:tenant_id")
						tenantScoped.Use(params.TenantMiddleware)
						{
							tenantScoped.GET("", params.TenantHandler.Get)
							tenantScoped.PATCH("", middleware.RequireRole(memberships.RoleTenantAdmin), params.TenantHandler.Update)
							tenantScoped.GET("/members", middleware.RequireRole(memberships.RoleTenantAdmin, memberships.RoleTenantOperator, memberships.RoleViewer), params.TenantHandler.ListMembers)
							tenantScoped.POST("/members", middleware.RequireRole(memberships.RoleTenantAdmin), params.TenantHandler.AddMember)
						}
					}
				}
			}
		}
	}

	return r
}
