package tenants

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logiflows/logiflows/backend/internal/contextutil"
	"github.com/logiflows/logiflows/backend/internal/response"
)

// Handler exposes HTTP handlers for multi-tenant management endpoints.
type Handler struct {
	service Service
}

// NewHandler initializes a new tenants handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Create handles POST /api/v1/tenants.
func (h *Handler) Create(c *gin.Context) {
	userID, ok := contextutil.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Authentication required")
		return
	}

	var req CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid tenant creation payload", err.Error())
		return
	}

	tenant, err := h.service.CreateTenant(c.Request.Context(), userID, req)
	if err != nil {
		if errors.Is(err, ErrTenantAlreadyExists) {
			response.Conflict(c, "A company with this identifier already exists")
			return
		}
		response.InternalServerError(c, "Failed to create tenant company")
		return
	}

	response.Success(c, http.StatusCreated, tenant)
}

// List handles GET /api/v1/tenants.
func (h *Handler) List(c *gin.Context) {
	userID, ok := contextutil.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Authentication required")
		return
	}

	isPlatformAdmin := contextutil.IsPlatformAdmin(c)
	list, err := h.service.ListUserTenants(c.Request.Context(), userID, isPlatformAdmin)
	if err != nil {
		response.InternalServerError(c, "Failed to list tenant companies")
		return
	}

	if list == nil {
		list = []Tenant{}
	}

	response.Success(c, http.StatusOK, list)
}

// Get handles GET /api/v1/tenants/:tenant_id.
func (h *Handler) Get(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.BadRequest(c, "Valid tenant ID is required", nil)
		return
	}

	tenant, err := h.service.GetTenant(c.Request.Context(), tenantID)
	if err != nil {
		if errors.Is(err, ErrTenantNotFound) {
			response.NotFound(c, "Tenant company not found")
			return
		}
		response.InternalServerError(c, "Failed to fetch tenant details")
		return
	}

	response.Success(c, http.StatusOK, tenant)
}

// ListMembers handles GET /api/v1/tenants/:tenant_id/members.
func (h *Handler) ListMembers(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.BadRequest(c, "Valid tenant ID is required", nil)
		return
	}

	members, err := h.service.ListMembers(c.Request.Context(), tenantID)
	if err != nil {
		response.InternalServerError(c, "Failed to list tenant members")
		return
	}

	response.Success(c, http.StatusOK, members)
}

// AddMember handles POST /api/v1/tenants/:tenant_id/members.
func (h *Handler) AddMember(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.BadRequest(c, "Valid tenant ID is required", nil)
		return
	}

	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid member invitation payload", err.Error())
		return
	}

	member, err := h.service.AddMember(c.Request.Context(), tenantID, req)
	if err != nil {
		if errors.Is(err, ErrInvalidRole) {
			response.BadRequest(c, "Invalid role specified. Supported: TENANT_ADMIN, TENANT_OPERATOR, VIEWER", nil)
			return
		}
		if errors.Is(err, ErrUserNotFound) {
			response.NotFound(c, "No user found with the specified email address")
			return
		}
		if errors.Is(err, ErrAlreadyMember) {
			response.Conflict(c, "User is already an active member of this tenant")
			return
		}
		response.InternalServerError(c, "Failed to add tenant member")
		return
	}

	response.Success(c, http.StatusCreated, member)
}

// Update handles PATCH /api/v1/tenants/:tenant_id.
func (h *Handler) Update(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.BadRequest(c, "Valid tenant ID is required", nil)
		return
	}

	var req UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid tenant update payload", err.Error())
		return
	}

	tenant, err := h.service.UpdateTenant(c.Request.Context(), tenantID, req)
	if err != nil {
		if errors.Is(err, ErrTenantNotFound) {
			response.NotFound(c, "Tenant company not found")
			return
		}
		response.InternalServerError(c, "Failed to update tenant company")
		return
	}

	response.Success(c, http.StatusOK, tenant)
}

// GetCurrent handles GET /api/v1/tenants/current (and /api/v1/companies/current).
func (h *Handler) GetCurrent(c *gin.Context) {
	userID, ok := contextutil.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Authentication required")
		return
	}

	// 1. If an active tenant ID is specified in the request context or header, try that first
	if tenantID, ok := contextutil.GetTenantID(c); ok {
		tenant, err := h.service.GetTenant(c.Request.Context(), tenantID)
		if err == nil && tenant != nil {
			response.Success(c, http.StatusOK, tenant)
			return
		}
	}

	// 2. Otherwise return the user's primary/first tenant membership
	isPlatformAdmin := contextutil.IsPlatformAdmin(c)
	userTenants, err := h.service.ListUserTenants(c.Request.Context(), userID, isPlatformAdmin)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve current company details")
		return
	}

	if len(userTenants) == 0 {
		response.NotFound(c, "No company associated with current user")
		return
	}

	response.Success(c, http.StatusOK, userTenants[0])
}
