package server

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/logiflows/logiflows/backend/docs"
	"github.com/logiflows/logiflows/backend/internal/auth"
	"github.com/logiflows/logiflows/backend/internal/branches"
	"github.com/logiflows/logiflows/backend/internal/config"
	"github.com/logiflows/logiflows/backend/internal/employees"
	"github.com/logiflows/logiflows/backend/internal/health"
	"github.com/logiflows/logiflows/backend/internal/memberships"
	"github.com/logiflows/logiflows/backend/internal/middleware"
	"github.com/logiflows/logiflows/backend/internal/response"
	"github.com/logiflows/logiflows/backend/internal/tenants"
	"github.com/logiflows/logiflows/backend/internal/vehicles"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// RouterParams encapsulates dependencies required to wire all application routes.
type RouterParams struct {
	Cfg              *config.Config
	Log              *slog.Logger
	HealthHandler    *health.Handler
	AuthHandler      *auth.Handler
	TenantHandler    *tenants.Handler
	BranchHandler    *branches.Handler
	EmployeeHandler  *employees.Handler
	VehicleHandler   *vehicles.Handler
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

	// Swagger & OpenAPI Documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
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

				// Tenant & Company operations
				if params.TenantHandler != nil {
					authed.POST("/tenants", params.TenantHandler.Create)
					authed.GET("/tenants", params.TenantHandler.List)
					authed.GET("/tenants/current", params.TenantHandler.GetCurrent)

					// Company route aliases
					authed.POST("/companies", params.TenantHandler.Create)
					authed.GET("/companies", params.TenantHandler.List)
					authed.GET("/companies/current", params.TenantHandler.GetCurrent)

					// Tenant-scoped routes with multi-tenant isolation enforcement
					if params.TenantMiddleware != nil {
						// Register both /tenants/:tenant_id and /companies/:tenant_id for API flexibility
						for _, prefix := range []string{"/tenants/:tenant_id", "/companies/:tenant_id"} {
							tenantScoped := authed.Group(prefix)
							tenantScoped.Use(params.TenantMiddleware)
							{
								tenantScoped.GET("", params.TenantHandler.Get)
								tenantScoped.PATCH("", middleware.RequireRole(memberships.RoleTenantAdmin), params.TenantHandler.Update)
								tenantScoped.GET("/members", middleware.RequireRole(memberships.RoleTenantAdmin, memberships.RoleTenantOperator, memberships.RoleViewer), params.TenantHandler.ListMembers)
								tenantScoped.POST("/members", middleware.RequireRole(memberships.RoleTenantAdmin), params.TenantHandler.AddMember)

								// Delivery Branch Operations
								if params.BranchHandler != nil {
									branchRoutes := tenantScoped.Group("/branches")
									{
										branchRoutes.POST("", middleware.RequireRole(memberships.RoleTenantAdmin), params.BranchHandler.Create)
										branchRoutes.GET("", middleware.RequireRole(memberships.RoleTenantAdmin, memberships.RoleTenantOperator, memberships.RoleViewer), params.BranchHandler.List)
										branchRoutes.GET("/:branch_id", middleware.RequireRole(memberships.RoleTenantAdmin, memberships.RoleTenantOperator, memberships.RoleViewer), params.BranchHandler.Get)
										branchRoutes.PATCH("/:branch_id", middleware.RequireRole(memberships.RoleTenantAdmin), params.BranchHandler.Update)
										branchRoutes.PATCH("/:branch_id/status", middleware.RequireRole(memberships.RoleTenantAdmin), params.BranchHandler.UpdateStatus)
										branchRoutes.DELETE("/:branch_id", middleware.RequireRole(memberships.RoleTenantAdmin), params.BranchHandler.Delete)
									}
								}
							}
						}

						tenantScoped := authed.Group("/tenants/:tenant_id")
						tenantScoped.Use(params.TenantMiddleware)
						{

							// Employee Operations
							if params.EmployeeHandler != nil {
								empRoutes := tenantScoped.Group("/employees")
								{
									empRoutes.POST("", middleware.RequireRole(memberships.RoleTenantAdmin), params.EmployeeHandler.Create)
									empRoutes.GET("", middleware.RequireRole(memberships.RoleTenantAdmin, memberships.RoleTenantOperator, memberships.RoleViewer), params.EmployeeHandler.List)
									empRoutes.GET("/:employee_id", middleware.RequireRole(memberships.RoleTenantAdmin, memberships.RoleTenantOperator, memberships.RoleViewer), params.EmployeeHandler.Get)
									empRoutes.PUT("/:employee_id", middleware.RequireRole(memberships.RoleTenantAdmin), params.EmployeeHandler.Update)
									empRoutes.PATCH("/:employee_id", middleware.RequireRole(memberships.RoleTenantAdmin), params.EmployeeHandler.Update)
									empRoutes.DELETE("/:employee_id", middleware.RequireRole(memberships.RoleTenantAdmin), params.EmployeeHandler.Deactivate)
								}
							}

							// Vehicle Fleet Operations
							if params.VehicleHandler != nil {
								vehRoutes := tenantScoped.Group("/vehicles")
								{
									vehRoutes.POST("", middleware.RequireRole(memberships.RoleTenantAdmin), params.VehicleHandler.Create)
									vehRoutes.GET("", middleware.RequireRole(memberships.RoleTenantAdmin, memberships.RoleTenantOperator, memberships.RoleViewer), params.VehicleHandler.List)
									vehRoutes.GET("/:vehicle_id", middleware.RequireRole(memberships.RoleTenantAdmin, memberships.RoleTenantOperator, memberships.RoleViewer), params.VehicleHandler.Get)
									vehRoutes.PUT("/:vehicle_id", middleware.RequireRole(memberships.RoleTenantAdmin), params.VehicleHandler.Update)
									vehRoutes.PATCH("/:vehicle_id", middleware.RequireRole(memberships.RoleTenantAdmin), params.VehicleHandler.Update)
									vehRoutes.DELETE("/:vehicle_id", middleware.RequireRole(memberships.RoleTenantAdmin), params.VehicleHandler.Deactivate)
									vehRoutes.POST("/:vehicle_id/assign", middleware.RequireRole(memberships.RoleTenantAdmin, memberships.RoleTenantOperator), params.VehicleHandler.AssignDriver)
									vehRoutes.POST("/:vehicle_id/unassign", middleware.RequireRole(memberships.RoleTenantAdmin, memberships.RoleTenantOperator), params.VehicleHandler.UnassignDriver)
								}
							}
						}
					}
				}
			}
		}
	}

	return r
}
