package branches

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/logiflows/logiflows/backend/internal/contextutil"
	"github.com/logiflows/logiflows/backend/internal/response"
)

// Handler exposes HTTP endpoints for delivery branch management.
type Handler struct {
	service Service
}

// NewHandler initializes a new branches HTTP handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Create handles POST /api/v1/tenants/:tenant_id/branches.
func (h *Handler) Create(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.BadRequest(c, "Target tenant ID is required", nil)
		return
	}

	userID, ok := contextutil.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Authentication required")
		return
	}

	var req CreateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid branch creation payload", err.Error())
		return
	}

	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	branch, err := h.service.CreateBranch(c.Request.Context(), tenantID, userID, req, ip, userAgent)
	if err != nil {
		if errors.Is(err, ErrBranchCodeExists) {
			response.Conflict(c, "A branch with this code already exists for this tenant")
			return
		}
		if errors.Is(err, ErrInvalidCoordinates) || errors.Is(err, ErrInvalidCoverageRadius) {
			response.BadRequest(c, err.Error(), nil)
			return
		}
		response.InternalServerError(c, "Failed to create branch hub")
		return
	}

	response.Success(c, http.StatusCreated, branch)
}

// Get handles GET /api/v1/tenants/:tenant_id/branches/:branch_id.
func (h *Handler) Get(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.BadRequest(c, "Target tenant ID is required", nil)
		return
	}

	branchIDStr := c.Param("branch_id")
	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid branch UUID format", nil)
		return
	}

	branch, err := h.service.GetBranch(c.Request.Context(), tenantID, branchID)
	if err != nil {
		if errors.Is(err, ErrBranchNotFound) {
			response.NotFound(c, "Branch not found")
			return
		}
		response.InternalServerError(c, "Failed to fetch branch details")
		return
	}

	response.Success(c, http.StatusOK, branch)
}

// List handles GET /api/v1/tenants/:tenant_id/branches.
func (h *Handler) List(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.BadRequest(c, "Target tenant ID is required", nil)
		return
	}

	var filter BranchFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	res, err := h.service.ListBranches(c.Request.Context(), tenantID, filter)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve branches list")
		return
	}

	response.Success(c, http.StatusOK, res)
}

// Update handles PATCH /api/v1/tenants/:tenant_id/branches/:branch_id.
func (h *Handler) Update(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.BadRequest(c, "Target tenant ID is required", nil)
		return
	}

	userID, ok := contextutil.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Authentication required")
		return
	}

	branchIDStr := c.Param("branch_id")
	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid branch UUID format", nil)
		return
	}

	var req UpdateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid branch update payload", err.Error())
		return
	}

	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	branch, err := h.service.UpdateBranch(c.Request.Context(), tenantID, branchID, userID, req, ip, userAgent)
	if err != nil {
		if errors.Is(err, ErrBranchNotFound) {
			response.NotFound(c, "Branch not found")
			return
		}
		if errors.Is(err, ErrInvalidCoordinates) || errors.Is(err, ErrInvalidBranchStatus) || errors.Is(err, ErrInvalidCoverageRadius) {
			response.BadRequest(c, err.Error(), nil)
			return
		}
		response.InternalServerError(c, "Failed to update branch")
		return
	}

	response.Success(c, http.StatusOK, branch)
}

// Delete handles DELETE /api/v1/tenants/:tenant_id/branches/:branch_id (soft deactivation).
func (h *Handler) Delete(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.BadRequest(c, "Target tenant ID is required", nil)
		return
	}

	userID, ok := contextutil.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Authentication required")
		return
	}

	branchIDStr := c.Param("branch_id")
	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid branch UUID format", nil)
		return
	}

	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	if err := h.service.DeactivateBranch(c.Request.Context(), tenantID, branchID, userID, ip, userAgent); err != nil {
		if errors.Is(err, ErrBranchNotFound) {
			response.NotFound(c, "Branch not found")
			return
		}
		response.InternalServerError(c, "Failed to deactivate branch")
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"message": "Branch deactivated successfully",
	})
}

// UpdateStatus handles PATCH /api/v1/tenants/:tenant_id/branches/:branch_id/status.
func (h *Handler) UpdateStatus(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.BadRequest(c, "Target tenant ID is required", nil)
		return
	}

	userID, ok := contextutil.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Authentication required")
		return
	}

	branchIDStr := c.Param("branch_id")
	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid branch UUID format", nil)
		return
	}

	var req UpdateBranchStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid branch status payload", err.Error())
		return
	}

	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	updateReq := UpdateBranchRequest{
		Status: &req.Status,
	}

	branch, err := h.service.UpdateBranch(c.Request.Context(), tenantID, branchID, userID, updateReq, ip, userAgent)
	if err != nil {
		if errors.Is(err, ErrBranchNotFound) {
			response.NotFound(c, "Branch not found")
			return
		}
		if errors.Is(err, ErrInvalidBranchStatus) {
			response.BadRequest(c, err.Error(), nil)
			return
		}
		response.InternalServerError(c, "Failed to update branch operating status")
		return
	}

	response.Success(c, http.StatusOK, branch)
}
