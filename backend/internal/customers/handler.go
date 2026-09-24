package customers

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

// Create handles POST /api/v1/tenants/:tenant_id/customers
func (h *Handler) Create(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.Forbidden(c, "MISSING_TENANT_CONTEXT", "Tenant context is required")
		return
	}

	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload format", err.Error())
		return
	}

	customer, err := h.service.CreateCustomer(c.Request.Context(), tenantID, req)
	if err != nil {
		if errors.Is(err, ErrInvalidCustomerName) ||
			errors.Is(err, ErrInvalidCustomerPhone) ||
			errors.Is(err, ErrInvalidCustomerAddress) ||
			errors.Is(err, ErrInvalidCustomerType) {
			response.BadRequest(c, err.Error(), nil)
			return
		}
		if errors.Is(err, ErrDuplicateCustomerCode) {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to create customer: "+err.Error())
		return
	}

	response.Success(c, http.StatusCreated, customer)
}

// Get handles GET /api/v1/tenants/:tenant_id/customers/:customer_id
func (h *Handler) Get(c *gin.Context) {
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

	customer, err := h.service.GetCustomer(c.Request.Context(), tenantID, customerID)
	if err != nil {
		if errors.Is(err, ErrCustomerNotFound) {
			response.NotFound(c, "Customer not found in this organization")
			return
		}
		response.InternalServerError(c, "Failed to retrieve customer: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, customer)
}

// List handles GET /api/v1/tenants/:tenant_id/customers
func (h *Handler) List(c *gin.Context) {
	tenantID, ok := contextutil.GetTenantID(c)
	if !ok {
		response.Forbidden(c, "MISSING_TENANT_CONTEXT", "Tenant context is required")
		return
	}

	filter := CustomerFilter{}

	if search := strings.TrimSpace(c.Query("search")); search != "" {
		filter.Search = &search
	}
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		filter.Status = &status
	}
	if cType := strings.TrimSpace(c.Query("customer_type")); cType != "" {
		filter.CustomerType = &cType
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			filter.Limit = l
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			filter.Offset = o
		}
	}

	res, err := h.service.ListCustomers(c.Request.Context(), tenantID, filter)
	if err != nil {
		response.InternalServerError(c, "Failed to list customers: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, res)
}

// Update handles PUT/PATCH /api/v1/tenants/:tenant_id/customers/:customer_id
func (h *Handler) Update(c *gin.Context) {
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

	var req UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid update request payload format", err.Error())
		return
	}

	updated, err := h.service.UpdateCustomer(c.Request.Context(), tenantID, customerID, req)
	if err != nil {
		if errors.Is(err, ErrCustomerNotFound) {
			response.NotFound(c, "Customer not found in this organization")
			return
		}
		if errors.Is(err, ErrInvalidCustomerType) || errors.Is(err, ErrInvalidCustomerStatus) {
			response.BadRequest(c, err.Error(), nil)
			return
		}
		response.InternalServerError(c, "Failed to update customer: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, updated)
}

// Delete handles DELETE /api/v1/tenants/:tenant_id/customers/:customer_id
func (h *Handler) Delete(c *gin.Context) {
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

	if err := h.service.DeleteCustomer(c.Request.Context(), tenantID, customerID); err != nil {
		if errors.Is(err, ErrCustomerNotFound) {
			response.NotFound(c, "Customer not found in this organization")
			return
		}
		response.InternalServerError(c, "Failed to delete customer: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{"deleted": true})
}
