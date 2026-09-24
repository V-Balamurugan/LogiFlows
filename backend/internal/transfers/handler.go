package transfers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

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

// Create handles POST /api/v1/tenants/:tenant_id/transfers
func (h *Handler) Create(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.Forbidden(c, "MISSING_TENANT_CONTEXT", "Tenant context is required")
		return
	}

	actorID, _ := contextutil.GetUserID(c)

	var req CreateBranchTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload format", err.Error())
		return
	}

	transfer, err := h.service.CreateTransfer(c.Request.Context(), tenantID, actorID, req)
	if err != nil {
		if errors.Is(err, ErrInvalidBranches) || errors.Is(err, ErrSameBranchTransfer) || errors.Is(err, ErrNoParcelsSpecified) {
			response.BadRequest(c, err.Error(), nil)
			return
		}
		response.InternalServerError(c, "Failed to create branch transfer: "+err.Error())
		return
	}

	response.Success(c, http.StatusCreated, transfer)
}

// Get handles GET /api/v1/tenants/:tenant_id/transfers/:transfer_id
func (h *Handler) Get(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.Forbidden(c, "MISSING_TENANT_CONTEXT", "Tenant context is required")
		return
	}

	transferID, err := uuid.Parse(c.Param("transfer_id"))
	if err != nil {
		response.BadRequest(c, "Invalid transfer_id parameter format", nil)
		return
	}

	transfer, parcels, err := h.service.GetTransfer(c.Request.Context(), tenantID, transferID)
	if err != nil {
		if errors.Is(err, ErrTransferNotFound) {
			response.NotFound(c, "Branch transfer not found")
			return
		}
		response.InternalServerError(c, "Failed to get branch transfer: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"transfer": transfer,
		"parcels":  parcels,
	})
}

// List handles GET /api/v1/tenants/:tenant_id/transfers
func (h *Handler) List(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.Forbidden(c, "MISSING_TENANT_CONTEXT", "Tenant context is required")
		return
	}

	filter := BranchTransferFilter{
		Limit:  20,
		Offset: 0,
	}

	if st := strings.TrimSpace(c.Query("status")); st != "" {
		upper := strings.ToUpper(st)
		filter.Status = &upper
	}
	if sb := strings.TrimSpace(c.Query("source_branch_id")); sb != "" {
		if id, err := uuid.Parse(sb); err == nil {
			filter.SourceBranchID = &id
		}
	}
	if db := strings.TrimSpace(c.Query("destination_branch_id")); db != "" {
		if id, err := uuid.Parse(db); err == nil {
			filter.DestinationBranchID = &id
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

	res, err := h.service.ListTransfers(c.Request.Context(), tenantID, filter)
	if err != nil {
		response.InternalServerError(c, "Failed to list branch transfers: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, res)
}

// Dispatch handles POST /api/v1/tenants/:tenant_id/transfers/:transfer_id/dispatch
func (h *Handler) Dispatch(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.Forbidden(c, "MISSING_TENANT_CONTEXT", "Tenant context is required")
		return
	}

	transferID, err := uuid.Parse(c.Param("transfer_id"))
	if err != nil {
		response.BadRequest(c, "Invalid transfer_id parameter format", nil)
		return
	}

	var req DispatchTransferRequest
	_ = c.ShouldBindJSON(&req)

	if err := h.service.DispatchTransfer(c.Request.Context(), tenantID, transferID, req); err != nil {
		if errors.Is(err, ErrTransferNotFound) {
			response.NotFound(c, "Branch transfer not found")
			return
		}
		if errors.Is(err, ErrInvalidStatusAction) {
			response.BadRequest(c, "Transfer manifest can only be dispatched when in PENDING status", nil)
			return
		}
		response.InternalServerError(c, "Failed to dispatch transfer: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"message":     "Transfer manifest dispatched successfully and parcels transitioned to IN_TRANSIT",
		"transfer_id": transferID,
	})
}

// Receive handles POST /api/v1/tenants/:tenant_id/transfers/:transfer_id/receive
func (h *Handler) Receive(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.Forbidden(c, "MISSING_TENANT_CONTEXT", "Tenant context is required")
		return
	}

	transferID, err := uuid.Parse(c.Param("transfer_id"))
	if err != nil {
		response.BadRequest(c, "Invalid transfer_id parameter format", nil)
		return
	}

	var req ReceiveTransferRequest
	_ = c.ShouldBindJSON(&req)

	if err := h.service.ReceiveTransfer(c.Request.Context(), tenantID, transferID, req); err != nil {
		if errors.Is(err, ErrTransferNotFound) {
			response.NotFound(c, "Branch transfer not found")
			return
		}
		if errors.Is(err, ErrTransferAlreadyReceived) {
			response.Conflict(c, "Transfer manifest has already been fully received")
			return
		}
		response.InternalServerError(c, "Failed to receive transfer: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"message":     "Transfer manifest received and parcels scanned into destination branch",
		"transfer_id": transferID,
	})
}
