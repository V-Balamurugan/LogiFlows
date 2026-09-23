package employees

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

// Create handles POST /api/v1/tenants/:tenant_id/employees
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

	var req CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid employee creation payload", err.Error())
		return
	}

	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	emp, err := h.service.CreateEmployee(c.Request.Context(), tenantID, userID, req, ip, userAgent)
	if err != nil {
		if errors.Is(err, ErrDuplicateEmployeeCode) {
			response.Conflict(c, "An employee with this code already exists in this organization")
			return
		}
		if errors.Is(err, ErrUserAlreadyLinked) {
			response.Conflict(c, "This user account is already linked to an employee profile in this organization")
			return
		}
		if errors.Is(err, ErrBranchCrossTenant) {
			response.Forbidden(c, "CROSS_TENANT_FORBIDDEN", "Cannot assign an employee to a branch belonging to another organization")
			return
		}
		if errors.Is(err, ErrBranchNotFound) {
			response.BadRequest(c, "Assigned branch does not exist or is inactive", nil)
			return
		}
		if errors.Is(err, ErrInvalidEmployeeCode) ||
			errors.Is(err, ErrInvalidEmployeeName) ||
			errors.Is(err, ErrInvalidDesignation) ||
			errors.Is(err, ErrInvalidEmploymentType) ||
			errors.Is(err, ErrInvalidOperationalRole) ||
			errors.Is(err, ErrInvalidAvailabilityStatus) ||
			errors.Is(err, ErrInvalidVerificationStatus) {
			response.BadRequest(c, err.Error(), nil)
			return
		}

		response.InternalServerError(c, "Failed to register employee profile")
		return
	}

	response.Success(c, http.StatusCreated, emp)
}

// Get handles GET /api/v1/tenants/:tenant_id/employees/:employee_id
func (h *Handler) Get(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.BadRequest(c, "Target tenant ID is required", nil)
		return
	}

	empIDStr := c.Param("employee_id")
	empID, err := uuid.Parse(empIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid employee UUID format", nil)
		return
	}

	emp, err := h.service.GetEmployee(c.Request.Context(), tenantID, empID)
	if err != nil {
		if errors.Is(err, ErrEmployeeNotFound) {
			response.NotFound(c, "Employee profile not found")
			return
		}
		response.InternalServerError(c, "Failed to fetch employee profile")
		return
	}

	response.Success(c, http.StatusOK, emp)
}

// List handles GET /api/v1/tenants/:tenant_id/employees
func (h *Handler) List(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.BadRequest(c, "Target tenant ID is required", nil)
		return
	}

	search := c.Query("search")
	operationalRole := c.Query("operational_role")
	status := c.Query("status")
	availabilityStatus := c.Query("availability_status")
	verificationStatus := c.Query("verification_status")

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

	filter := EmployeeFilter{
		Search:             search,
		OperationalRole:    operationalRole,
		BranchID:           branchID,
		Status:             status,
		AvailabilityStatus: availabilityStatus,
		VerificationStatus: verificationStatus,
		Limit:              limit,
		Offset:             offset,
	}

	resp, err := h.service.ListEmployees(c.Request.Context(), tenantID, filter)
	if err != nil {
		response.InternalServerError(c, "Failed to list employee profiles")
		return
	}

	response.Success(c, http.StatusOK, resp)
}

// Update handles PUT and PATCH /api/v1/tenants/:tenant_id/employees/:employee_id
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

	empIDStr := c.Param("employee_id")
	empID, err := uuid.Parse(empIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid employee UUID format", nil)
		return
	}

	var req UpdateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid employee update payload", err.Error())
		return
	}

	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	emp, err := h.service.UpdateEmployee(c.Request.Context(), tenantID, empID, userID, req, ip, userAgent)
	if err != nil {
		if errors.Is(err, ErrEmployeeNotFound) {
			response.NotFound(c, "Employee profile not found")
			return
		}
		if errors.Is(err, ErrBranchCrossTenant) {
			response.Forbidden(c, "CROSS_TENANT_FORBIDDEN", "Cannot assign an employee to a branch belonging to another organization")
			return
		}
		if errors.Is(err, ErrBranchNotFound) {
			response.BadRequest(c, "Assigned branch does not exist or is inactive", nil)
			return
		}
		if errors.Is(err, ErrInvalidEmploymentType) ||
			errors.Is(err, ErrInvalidOperationalRole) ||
			errors.Is(err, ErrInvalidStatus) ||
			errors.Is(err, ErrInvalidAvailabilityStatus) ||
			errors.Is(err, ErrInvalidVerificationStatus) {
			response.BadRequest(c, err.Error(), nil)
			return
		}

		response.InternalServerError(c, "Failed to update employee profile")
		return
	}

	response.Success(c, http.StatusOK, emp)
}

// UpdateStatus handles PATCH /api/v1/tenants/:tenant_id/employees/:employee_id/status
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

	empIDStr := c.Param("employee_id")
	empID, err := uuid.Parse(empIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid employee UUID format", nil)
		return
	}

	var req UpdateEmployeeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid employee status update payload", err.Error())
		return
	}

	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	emp, err := h.service.UpdateEmployeeStatus(c.Request.Context(), tenantID, empID, userID, req, ip, userAgent)
	if err != nil {
		if errors.Is(err, ErrEmployeeNotFound) {
			response.NotFound(c, "Employee profile not found")
			return
		}
		if errors.Is(err, ErrInvalidStatus) ||
			errors.Is(err, ErrInvalidAvailabilityStatus) ||
			errors.Is(err, ErrInvalidVerificationStatus) {
			response.BadRequest(c, err.Error(), nil)
			return
		}

		response.InternalServerError(c, "Failed to update employee status")
		return
	}

	response.Success(c, http.StatusOK, emp)
}

// Deactivate handles DELETE /api/v1/tenants/:tenant_id/employees/:employee_id
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

	empIDStr := c.Param("employee_id")
	empID, err := uuid.Parse(empIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid employee UUID format", nil)
		return
	}

	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	if err := h.service.DeactivateEmployee(c.Request.Context(), tenantID, empID, userID, ip, userAgent); err != nil {
		if errors.Is(err, ErrEmployeeNotFound) {
			response.NotFound(c, "Employee profile not found")
			return
		}
		response.InternalServerError(c, "Failed to deactivate employee profile")
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"message": "Employee profile deactivated successfully",
	})
}

// ListAvailableDrivers handles GET /api/v1/tenants/:tenant_id/employees/available-drivers
func (h *Handler) ListAvailableDrivers(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.BadRequest(c, "Target tenant ID is required", nil)
		return
	}

	var branchID *uuid.UUID
	if bStr := c.Query("branch_id"); bStr != "" {
		if parsed, err := uuid.Parse(bStr); err == nil {
			branchID = &parsed
		}
	}

	drivers, err := h.service.ListAvailableDrivers(c.Request.Context(), tenantID, branchID)
	if err != nil {
		response.InternalServerError(c, "Failed to query available drivers")
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"drivers": drivers,
		"total":   len(drivers),
	})
}

// CreateWithAccount handles POST /api/v1/tenants/:tenant_id/employees/with-account
func (h *Handler) CreateWithAccount(c *gin.Context) {
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

	var req CreateEmployeeWithAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid employee and account creation payload", err.Error())
		return
	}

	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	emp, err := h.service.CreateEmployeeWithAccount(c.Request.Context(), tenantID, userID, req, ip, userAgent)
	if err != nil {
		if errors.Is(err, ErrDuplicateEmployeeCode) {
			response.Conflict(c, "An employee with this code already exists in this organization")
			return
		}
		if errors.Is(err, ErrUserAlreadyLinked) {
			response.Conflict(c, "A user account with this email is already a member or linked to an employee in this organization")
			return
		}
		if errors.Is(err, ErrBranchNotFound) {
			response.BadRequest(c, "Assigned branch does not exist or is inactive", nil)
			return
		}
		if errors.Is(err, ErrEmailRequired) || errors.Is(err, ErrInvalidEmail) ||
			errors.Is(err, ErrAccountPasswordRequired) || errors.Is(err, ErrInvalidSystemRole) ||
			errors.Is(err, ErrInvalidEmployeeName) || errors.Is(err, ErrInvalidDesignation) ||
			errors.Is(err, ErrInvalidEmploymentType) || errors.Is(err, ErrInvalidOperationalRole) {
			response.BadRequest(c, err.Error(), nil)
			return
		}

		response.InternalServerError(c, "Failed to create employee profile and login account")
		return
	}

	response.Success(c, http.StatusCreated, emp)
}

// GetAccountStatus handles GET /api/v1/tenants/:tenant_id/employees/:employee_id/account-status
func (h *Handler) GetAccountStatus(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.BadRequest(c, "Target tenant ID is required", nil)
		return
	}

	empIDStr := c.Param("employee_id")
	empID, err := uuid.Parse(empIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid employee UUID format", nil)
		return
	}

	status, err := h.service.GetAccountStatus(c.Request.Context(), tenantID, empID)
	if err != nil {
		if errors.Is(err, ErrEmployeeNotFound) {
			response.NotFound(c, "Employee profile not found")
			return
		}
		response.InternalServerError(c, "Failed to retrieve employee account status")
		return
	}

	response.Success(c, http.StatusOK, status)
}

// GetMyProfile handles GET /api/v1/tenants/:tenant_id/employees/me
func (h *Handler) GetMyProfile(c *gin.Context) {
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

	profile, err := h.service.GetMyProfile(c.Request.Context(), tenantID, userID)
	if err != nil {
		if errors.Is(err, ErrEmployeeNotFound) {
			response.NotFound(c, "No employee profile linked to current user account in this organization")
			return
		}
		response.InternalServerError(c, "Failed to retrieve user profile")
		return
	}

	response.Success(c, http.StatusOK, profile)
}
