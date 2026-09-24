package parcels

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

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

// Create handles POST /api/v1/tenants/:tenant_id/parcels
func (h *Handler) Create(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.Forbidden(c, "MISSING_TENANT_CONTEXT", "Tenant context is required")
		return
	}

	actorID, _ := contextutil.GetUserID(c)

	var req CreateParcelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload format", err.Error())
		return
	}

	parcel, err := h.service.CreateParcel(c.Request.Context(), tenantID, actorID, req)
	if err != nil {
		if errors.Is(err, ErrInvalidSenderDetails) ||
			errors.Is(err, ErrInvalidReceiverDetails) ||
			errors.Is(err, ErrInvalidWeight) ||
			errors.Is(err, ErrInvalidDimensions) ||
			errors.Is(err, ErrInvalidServiceType) ||
			errors.Is(err, ErrInvalidBranch) ||
			errors.Is(err, ErrSameOriginAndDestination) {
			response.BadRequest(c, err.Error(), nil)
			return
		}
		if errors.Is(err, ErrBranchCrossTenant) {
			response.Forbidden(c, "CROSS_TENANT_RESOURCE", err.Error())
			return
		}
		if errors.Is(err, ErrDuplicateTrackingNumber) {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to create parcel: "+err.Error())
		return
	}

	response.Success(c, http.StatusCreated, parcel)
}

// Get handles GET /api/v1/tenants/:tenant_id/parcels/:parcel_id
func (h *Handler) Get(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.Forbidden(c, "MISSING_TENANT_CONTEXT", "Tenant context is required")
		return
	}

	parcelID, err := uuid.Parse(c.Param("parcel_id"))
	if err != nil {
		response.BadRequest(c, "Invalid parcel_id parameter format", nil)
		return
	}

	parcel, err := h.service.GetParcel(c.Request.Context(), tenantID, parcelID)
	if err != nil {
		if errors.Is(err, ErrParcelNotFound) {
			response.NotFound(c, "Parcel not found in this organization")
			return
		}
		response.InternalServerError(c, "Failed to retrieve parcel: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, parcel)
}

// List handles GET /api/v1/tenants/:tenant_id/parcels
func (h *Handler) List(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.Forbidden(c, "MISSING_TENANT_CONTEXT", "Tenant context is required")
		return
	}

	filter := ParcelFilter{
		Limit:  20,
		Offset: 0,
	}

	if st := strings.TrimSpace(c.Query("status")); st != "" {
		upper := strings.ToUpper(st)
		filter.Status = &upper
	}
	if ob := strings.TrimSpace(c.Query("origin_branch_id")); ob != "" {
		if id, err := uuid.Parse(ob); err == nil {
			filter.OriginBranchID = &id
		}
	}
	if db := strings.TrimSpace(c.Query("destination_branch_id")); db != "" {
		if id, err := uuid.Parse(db); err == nil {
			filter.DestBranchID = &id
		}
	}
	if cb := strings.TrimSpace(c.Query("current_branch_id")); cb != "" {
		if id, err := uuid.Parse(cb); err == nil {
			filter.CurrentBranchID = &id
		}
	}
	if s := strings.TrimSpace(c.Query("search")); s != "" {
		filter.Search = &s
	}
	if df := strings.TrimSpace(c.Query("date_from")); df != "" {
		if t, err := time.Parse(time.RFC3339, df); err == nil {
			filter.DateFrom = &t
		}
	}
	if dt := strings.TrimSpace(c.Query("date_to")); dt != "" {
		if t, err := time.Parse(time.RFC3339, dt); err == nil {
			filter.DateTo = &t
		}
	}
	if cust := strings.TrimSpace(c.Query("customer_id")); cust != "" {
		if id, err := uuid.Parse(cust); err == nil {
			filter.CustomerID = &id
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

	res, err := h.service.ListParcels(c.Request.Context(), tenantID, filter)
	if err != nil {
		response.InternalServerError(c, "Failed to list parcels: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, res)
}

// GetByCustomer handles GET /api/v1/tenants/:tenant_id/customers/:customer_id/parcels
func (h *Handler) GetByCustomer(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.Forbidden(c, "MISSING_TENANT_CONTEXT", "Tenant context is required")
		return
	}

	customerID, err := uuid.Parse(c.Param("customer_id"))
	if err != nil {
		response.BadRequest(c, "Invalid customer_id parameter format", nil)
		return
	}

	filter := ParcelFilter{
		CustomerID: &customerID,
		Limit:      20,
		Offset:     0,
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

	res, err := h.service.ListParcels(c.Request.Context(), tenantID, filter)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve customer parcels: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, res)
}

// Update handles PATCH /api/v1/tenants/:tenant_id/parcels/:parcel_id
func (h *Handler) Update(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.Forbidden(c, "MISSING_TENANT_CONTEXT", "Tenant context is required")
		return
	}

	parcelID, err := uuid.Parse(c.Param("parcel_id"))
	if err != nil {
		response.BadRequest(c, "Invalid parcel_id parameter format", nil)
		return
	}

	var req UpdateParcelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload format", err.Error())
		return
	}

	parcel, err := h.service.UpdateParcel(c.Request.Context(), tenantID, parcelID, req)
	if err != nil {
		if errors.Is(err, ErrParcelNotFound) {
			response.NotFound(c, "Parcel not found in this organization")
			return
		}
		response.InternalServerError(c, "Failed to update parcel: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, parcel)
}

// UpdateStatus handles PATCH /api/v1/tenants/:tenant_id/parcels/:parcel_id/status
func (h *Handler) UpdateStatus(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.Forbidden(c, "MISSING_TENANT_CONTEXT", "Tenant context is required")
		return
	}

	parcelID, err := uuid.Parse(c.Param("parcel_id"))
	if err != nil {
		response.BadRequest(c, "Invalid parcel_id parameter format", nil)
		return
	}

	actorID, _ := contextutil.GetUserID(c)
	actorRole := contextutil.GetTenantRole(c)

	var req UpdateParcelStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload format", err.Error())
		return
	}

	parcel, err := h.service.UpdateParcelStatus(c.Request.Context(), tenantID, parcelID, actorID, actorRole, req)
	if err != nil {
		if errors.Is(err, ErrParcelNotFound) {
			response.NotFound(c, "Parcel not found in this organization")
			return
		}
		if errors.Is(err, ErrInvalidStateTransition) || errors.Is(err, ErrInvalidStatus) {
			response.BadRequest(c, err.Error(), nil)
			return
		}
		if errors.Is(err, ErrBranchCrossTenant) {
			response.Forbidden(c, "CROSS_TENANT_RESOURCE", err.Error())
			return
		}
		response.InternalServerError(c, "Failed to update parcel status: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, parcel)
}

// GetTimeline handles GET /api/v1/tenants/:tenant_id/parcels/:parcel_id/timeline
func (h *Handler) GetTimeline(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.Forbidden(c, "MISSING_TENANT_CONTEXT", "Tenant context is required")
		return
	}

	parcelID, err := uuid.Parse(c.Param("parcel_id"))
	if err != nil {
		response.BadRequest(c, "Invalid parcel_id parameter format", nil)
		return
	}

	timeline, err := h.service.GetParcelTimeline(c.Request.Context(), tenantID, parcelID)
	if err != nil {
		if errors.Is(err, ErrParcelNotFound) {
			response.NotFound(c, "Parcel not found in this organization")
			return
		}
		response.InternalServerError(c, "Failed to retrieve parcel timeline: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, timeline)
}

// GetQRCode handles GET /api/v1/tenants/:tenant_id/parcels/:parcel_id/qr
func (h *Handler) GetQRCode(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.Forbidden(c, "MISSING_TENANT_CONTEXT", "Tenant context is required")
		return
	}

	parcelID, err := uuid.Parse(c.Param("parcel_id"))
	if err != nil {
		response.BadRequest(c, "Invalid parcel_id parameter format", nil)
		return
	}

	qr, err := h.service.GetParcelQRCode(c.Request.Context(), tenantID, parcelID)
	if err != nil {
		if errors.Is(err, ErrParcelNotFound) {
			response.NotFound(c, "Parcel not found in this organization")
			return
		}
		response.InternalServerError(c, "Failed to retrieve QR payload: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"parcel_id":  parcelID,
		"qr_payload": qr,
	})
}

// Scan handles POST /api/v1/tenants/:tenant_id/parcels/scan
func (h *Handler) Scan(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.Forbidden(c, "MISSING_TENANT_CONTEXT", "Tenant context is required")
		return
	}

	actorID, _ := contextutil.GetUserID(c)
	actorRole := contextutil.GetTenantRole(c)

	var req ScanParcelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid scan payload format", err.Error())
		return
	}

	result, err := h.service.ScanParcel(c.Request.Context(), tenantID, actorID, actorRole, req)
	if err != nil {
		if errors.Is(err, ErrInvalidQRPayload) {
			response.BadRequest(c, "Tampered or invalid QR payload", nil)
			return
		}
		if errors.Is(err, ErrParcelNotFound) {
			response.NotFound(c, "Parcel not found or belongs to another organization")
			return
		}
		response.InternalServerError(c, "Failed to process scan: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, result)
}

// PublicTrack handles GET /api/v1/tracking/:tracking_number (Unauthenticated Public Endpoint)
func (h *Handler) PublicTrack(c *gin.Context) {
	trackingNumber := c.Param("tracking_number")
	if strings.TrimSpace(trackingNumber) == "" {
		response.BadRequest(c, "tracking_number path parameter is required", nil)
		return
	}

	info, err := h.service.GetPublicTracking(c.Request.Context(), trackingNumber)
	if err != nil {
		if errors.Is(err, ErrParcelNotFound) {
			response.NotFound(c, "Shipment tracking number not found")
			return
		}
		response.InternalServerError(c, "Failed to retrieve shipment tracking: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, info)
}
