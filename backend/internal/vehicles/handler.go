package vehicles

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/logiflows/logiflows/backend/internal/contextutil"
	"github.com/logiflows/logiflows/backend/internal/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Create handles POST /api/v1/tenants/:tenant_id/vehicles
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

	var req CreateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid vehicle creation payload", err.Error())
		return
	}

	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	v, err := h.service.CreateVehicle(c.Request.Context(), tenantID, userID, req, ip, userAgent)
	if err != nil {
		if errors.Is(err, ErrDuplicateRegistrationNumber) {
			response.Conflict(c, "A vehicle with this registration number already exists in this organization")
			return
		}
		if errors.Is(err, ErrBranchNotFound) {
			response.BadRequest(c, "Assigned branch does not exist or is inactive", nil)
			return
		}
		if errors.Is(err, ErrInvalidRegistrationNumber) ||
			errors.Is(err, ErrInvalidVehicleType) ||
			errors.Is(err, ErrInvalidCapacity) {
			response.BadRequest(c, err.Error(), nil)
			return
		}

		response.InternalServerError(c, "Failed to register fleet vehicle")
		return
	}

	response.Success(c, http.StatusCreated, v)
}

// Get handles GET /api/v1/tenants/:tenant_id/vehicles/:vehicle_id
func (h *Handler) Get(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.BadRequest(c, "Target tenant ID is required", nil)
		return
	}

	vehicleIDStr := c.Param("vehicle_id")
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid vehicle UUID format", nil)
		return
	}

	v, err := h.service.GetVehicle(c.Request.Context(), tenantID, vehicleID)
	if err != nil {
		if errors.Is(err, ErrVehicleNotFound) {
			response.NotFound(c, "Vehicle not found")
			return
		}
		response.InternalServerError(c, "Failed to fetch vehicle details")
		return
	}

	response.Success(c, http.StatusOK, v)
}

// List handles GET /api/v1/tenants/:tenant_id/vehicles
func (h *Handler) List(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.BadRequest(c, "Target tenant ID is required", nil)
		return
	}

	search := c.Query("search")
	vehicleType := c.Query("vehicle_type")
	status := c.Query("status")

	var branchID *uuid.UUID
	if bStr := c.Query("branch_id"); bStr != "" {
		if parsed, err := uuid.Parse(bStr); err == nil {
			branchID = &parsed
		}
	}

	limit := 50
	if lStr := c.Query("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if oStr := c.Query("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	filter := VehicleFilter{
		Search:      search,
		VehicleType: vehicleType,
		Status:      status,
		BranchID:    branchID,
		Limit:       limit,
		Offset:      offset,
	}

	resp, err := h.service.ListVehicles(c.Request.Context(), tenantID, filter)
	if err != nil {
		response.InternalServerError(c, "Failed to list fleet vehicles")
		return
	}

	response.Success(c, http.StatusOK, resp)
}

// Update handles PUT /api/v1/tenants/:tenant_id/vehicles/:vehicle_id
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

	vehicleIDStr := c.Param("vehicle_id")
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid vehicle UUID format", nil)
		return
	}

	var req UpdateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid vehicle update payload", err.Error())
		return
	}

	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	v, err := h.service.UpdateVehicle(c.Request.Context(), tenantID, vehicleID, userID, req, ip, userAgent)
	if err != nil {
		if errors.Is(err, ErrVehicleNotFound) {
			response.NotFound(c, "Vehicle not found")
			return
		}
		if errors.Is(err, ErrBranchNotFound) {
			response.BadRequest(c, "Assigned branch does not exist or is inactive", nil)
			return
		}
		if errors.Is(err, ErrInvalidVehicleType) || errors.Is(err, ErrInvalidCapacity) {
			response.BadRequest(c, err.Error(), nil)
			return
		}

		response.InternalServerError(c, "Failed to update fleet vehicle")
		return
	}

	response.Success(c, http.StatusOK, v)
}

// Deactivate handles DELETE /api/v1/tenants/:tenant_id/vehicles/:vehicle_id
func (h *Handler) Deactivate(c *gin.Context) {
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

	vehicleIDStr := c.Param("vehicle_id")
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid vehicle UUID format", nil)
		return
	}

	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	if err := h.service.DeactivateVehicle(c.Request.Context(), tenantID, vehicleID, userID, ip, userAgent); err != nil {
		if errors.Is(err, ErrVehicleNotFound) {
			response.NotFound(c, "Vehicle not found")
			return
		}
		response.InternalServerError(c, "Failed to decommission fleet vehicle")
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"message": "Vehicle decommissioned successfully",
	})
}

// AssignDriver handles POST /api/v1/tenants/:tenant_id/vehicles/:vehicle_id/assign
func (h *Handler) AssignDriver(c *gin.Context) {
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

	vehicleIDStr := c.Param("vehicle_id")
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid vehicle UUID format", nil)
		return
	}

	var req AssignVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid vehicle assignment payload", err.Error())
		return
	}

	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	assignment, err := h.service.AssignDriver(c.Request.Context(), tenantID, vehicleID, userID, req, ip, userAgent)
	if err != nil {
		if errors.Is(err, ErrVehicleAlreadyAssigned) {
			response.Conflict(c, "Vehicle is currently actively assigned to another driver")
			return
		}
		if errors.Is(err, ErrDriverAlreadyAssigned) {
			response.Conflict(c, "Driver is currently actively assigned to another vehicle")
			return
		}
		if errors.Is(err, ErrDriverNotFound) {
			response.BadRequest(c, "Assigned driver was not found or is inactive", nil)
			return
		}
		if errors.Is(err, ErrVehicleNotFound) {
			response.NotFound(c, "Vehicle not found")
			return
		}

		response.InternalServerError(c, "Failed to assign driver to vehicle")
		return
	}

	response.Success(c, http.StatusCreated, assignment)
}

// UnassignDriver handles POST /api/v1/tenants/:tenant_id/vehicles/:vehicle_id/unassign
func (h *Handler) UnassignDriver(c *gin.Context) {
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

	vehicleIDStr := c.Param("vehicle_id")
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid vehicle UUID format", nil)
		return
	}

	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	if err := h.service.UnassignDriver(c.Request.Context(), tenantID, vehicleID, userID, ip, userAgent); err != nil {
		if errors.Is(err, ErrAssignmentNotFound) {
			response.NotFound(c, "No active assignment found for this vehicle")
			return
		}
		response.InternalServerError(c, "Failed to unassign vehicle")
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"message": "Vehicle driver unassigned successfully",
	})
}
