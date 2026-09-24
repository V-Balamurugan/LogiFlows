package deliveries

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/logiflows/logiflows/backend/internal/contextutil"
	"github.com/logiflows/logiflows/backend/internal/employees"
	"github.com/logiflows/logiflows/backend/internal/parcels"
	"github.com/logiflows/logiflows/backend/internal/response"
)

type Handler struct {
	service      Service
	employeeRepo employees.Repository
}

func NewHandler(service Service, employeeRepo employees.Repository) *Handler {
	return &Handler{
		service:      service,
		employeeRepo: employeeRepo,
	}
}

// Create handles POST /api/v1/tenants/:tenant_id/deliveries
func (h *Handler) Create(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.Forbidden(c, "MISSING_TENANT_CONTEXT", "Tenant context is required")
		return
	}

	var req CreateDeliveryTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload format", err.Error())
		return
	}

	task, err := h.service.CreateDeliveryTask(c.Request.Context(), tenantID, req)
	if err != nil {
		if errors.Is(err, ErrParcelAlreadyAssigned) {
			response.Conflict(c, "Parcel already has an active delivery task assigned")
			return
		}
		if errors.Is(err, ErrParcelNotEligibleForDelivery) || errors.Is(err, ErrDriverNotEligible) {
			response.BadRequest(c, err.Error(), nil)
			return
		}
		if errors.Is(err, parcels.ErrParcelNotFound) || errors.Is(err, employees.ErrEmployeeNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to create delivery task: "+err.Error())
		return
	}

	response.Success(c, http.StatusCreated, task)
}

// Get handles GET /api/v1/tenants/:tenant_id/deliveries/:delivery_id
func (h *Handler) Get(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.Forbidden(c, "MISSING_TENANT_CONTEXT", "Tenant context is required")
		return
	}

	taskID, err := uuid.Parse(c.Param("delivery_id"))
	if err != nil {
		response.BadRequest(c, "Invalid delivery_id parameter format", nil)
		return
	}

	task, err := h.service.GetDeliveryTask(c.Request.Context(), tenantID, taskID)
	if err != nil {
		if errors.Is(err, ErrDeliveryTaskNotFound) {
			response.NotFound(c, "Delivery task not found")
			return
		}
		response.InternalServerError(c, "Failed to get delivery task: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, task)
}

// List handles GET /api/v1/tenants/:tenant_id/deliveries
func (h *Handler) List(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.Forbidden(c, "MISSING_TENANT_CONTEXT", "Tenant context is required")
		return
	}

	filter := DeliveryTaskFilter{
		Limit:  20,
		Offset: 0,
	}

	if st := strings.TrimSpace(c.Query("status")); st != "" {
		upper := strings.ToUpper(st)
		filter.Status = &upper
	}
	if pr := strings.TrimSpace(c.Query("priority")); pr != "" {
		upper := strings.ToUpper(pr)
		filter.Priority = &upper
	}
	if did := strings.TrimSpace(c.Query("driver_id")); did != "" {
		if id, err := uuid.Parse(did); err == nil {
			filter.AssignedDriverID = &id
		}
	}
	if bid := strings.TrimSpace(c.Query("branch_id")); bid != "" {
		if id, err := uuid.Parse(bid); err == nil {
			filter.BranchID = &id
		}
	}
	if pid := strings.TrimSpace(c.Query("parcel_id")); pid != "" {
		if id, err := uuid.Parse(pid); err == nil {
			filter.ParcelID = &id
		}
	}
	if l := strings.TrimSpace(c.Query("limit")); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			filter.Limit = parsed
		}
	}
	if o := strings.TrimSpace(c.Query("offset")); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			filter.Offset = parsed
		}
	}

	res, err := h.service.ListDeliveryTasks(c.Request.Context(), tenantID, filter)
	if err != nil {
		response.InternalServerError(c, "Failed to list delivery tasks: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, res)
}

// GetMyTasks handles GET /api/v1/tenants/:tenant_id/deliveries/my-tasks
func (h *Handler) GetMyTasks(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.Forbidden(c, "MISSING_TENANT_CONTEXT", "Tenant context is required")
		return
	}

	userID, _ := contextutil.GetUserID(c)

	// Look up employee record by authenticated UserID
	emp, err := h.employeeRepo.GetByUserID(c.Request.Context(), tenantID, userID)
	if err != nil {
		// If caller is an admin testing or no linked employee record exists, check query param driver_id
		if did := strings.TrimSpace(c.Query("driver_id")); did != "" {
			if id, err := uuid.Parse(did); err == nil {
				tasks, err := h.service.GetDriverTasks(c.Request.Context(), tenantID, id)
				if err != nil {
					response.InternalServerError(c, "Failed to get driver tasks: "+err.Error())
					return
				}
				response.Success(c, http.StatusOK, gin.H{"tasks": tasks, "total": len(tasks)})
				return
			}
		}
		response.NotFound(c, "No employee profile linked to this user account")
		return
	}

	tasks, err := h.service.GetDriverTasks(c.Request.Context(), tenantID, emp.ID)
	if err != nil {
		response.InternalServerError(c, "Failed to get driver tasks: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"driver_id":   emp.ID,
		"driver_name": emp.FirstName + " " + emp.LastName,
		"tasks":       tasks,
		"total":       len(tasks),
	})
}

// UpdateStatus handles PATCH /api/v1/tenants/:tenant_id/deliveries/:delivery_id/status
func (h *Handler) UpdateStatus(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.Forbidden(c, "MISSING_TENANT_CONTEXT", "Tenant context is required")
		return
	}

	taskID, err := uuid.Parse(c.Param("delivery_id"))
	if err != nil {
		response.BadRequest(c, "Invalid delivery_id parameter format", nil)
		return
	}

	var req UpdateDeliveryTaskStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload format", err.Error())
		return
	}

	task, err := h.service.UpdateDeliveryTaskStatus(c.Request.Context(), tenantID, taskID, req)
	if err != nil {
		if errors.Is(err, ErrDeliveryTaskNotFound) {
			response.NotFound(c, "Delivery task not found")
			return
		}
		if errors.Is(err, ErrInvalidStatusTransition) {
			response.BadRequest(c, err.Error(), nil)
			return
		}
		response.InternalServerError(c, "Failed to update delivery task: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, task)
}

// RecordAttempt handles POST /api/v1/tenants/:tenant_id/deliveries/:delivery_id/attempt
func (h *Handler) RecordAttempt(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.Forbidden(c, "MISSING_TENANT_CONTEXT", "Tenant context is required")
		return
	}

	taskID, err := uuid.Parse(c.Param("delivery_id"))
	if err != nil {
		response.BadRequest(c, "Invalid delivery_id parameter format", nil)
		return
	}

	var req RecordDeliveryAttemptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload format", err.Error())
		return
	}

	attempt, err := h.service.RecordDeliveryAttempt(c.Request.Context(), tenantID, taskID, req)
	if err != nil {
		if errors.Is(err, ErrDeliveryTaskNotFound) {
			response.NotFound(c, "Delivery task not found")
			return
		}
		if errors.Is(err, ErrInvalidAttemptOutcome) {
			response.BadRequest(c, err.Error(), nil)
			return
		}
		response.InternalServerError(c, "Failed to record attempt: "+err.Error())
		return
	}

	response.Success(c, http.StatusCreated, attempt)
}

// GetAttempts handles GET /api/v1/tenants/:tenant_id/deliveries/:delivery_id/attempts
func (h *Handler) GetAttempts(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.Forbidden(c, "MISSING_TENANT_CONTEXT", "Tenant context is required")
		return
	}

	taskID, err := uuid.Parse(c.Param("delivery_id"))
	if err != nil {
		response.BadRequest(c, "Invalid delivery_id parameter format", nil)
		return
	}

	attempts, err := h.service.GetDeliveryAttempts(c.Request.Context(), tenantID, taskID)
	if err != nil {
		if errors.Is(err, ErrDeliveryTaskNotFound) {
			response.NotFound(c, "Delivery task not found")
			return
		}
		response.InternalServerError(c, "Failed to get delivery attempts: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, attempts)
}

// SubmitProof handles POST /api/v1/tenants/:tenant_id/deliveries/:delivery_id/proof
func (h *Handler) SubmitProof(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.Forbidden(c, "MISSING_TENANT_CONTEXT", "Tenant context is required")
		return
	}

	taskID, err := uuid.Parse(c.Param("delivery_id"))
	if err != nil {
		response.BadRequest(c, "Invalid delivery_id parameter format", nil)
		return
	}

	var req SubmitDeliveryProofRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload format", err.Error())
		return
	}

	if err := h.service.SubmitDeliveryProof(c.Request.Context(), tenantID, taskID, req); err != nil {
		if errors.Is(err, ErrDeliveryTaskNotFound) {
			response.NotFound(c, "Delivery task not found")
			return
		}
		if errors.Is(err, ErrDeliveryAlreadyCompleted) {
			response.Conflict(c, "Delivery task is already completed")
			return
		}
		if errors.Is(err, ErrInvalidProofType) || errors.Is(err, ErrMissingRecipientName) {
			response.BadRequest(c, err.Error(), nil)
			return
		}
		response.InternalServerError(c, "Failed to submit proof: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"message":          "Delivery completed successfully with proof of delivery verified",
		"delivery_task_id": taskID,
	})
}
