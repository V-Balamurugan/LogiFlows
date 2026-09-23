package middleware

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/logiflows/logiflows/backend/internal/contextutil"
	"github.com/logiflows/logiflows/backend/internal/memberships"
	"github.com/logiflows/logiflows/backend/internal/response"
	"github.com/logiflows/logiflows/backend/internal/tenants"
)

const (
	ContextTenantMembership = "tenantMembership"
)

// TenantContext extracts the tenant context from the URL param (:tenant_id) or X-Tenant-ID header
// and strictly validates that the authenticated caller has an active membership in that tenant.
func TenantContext(membershipRepo memberships.Repository, tenantRepo tenants.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := contextutil.GetUserID(c)
		if !exists {
			response.Unauthorized(c, "Authentication required to establish tenant context")
			return
		}

		tenantIDStr := c.Param("tenant_id")
		if tenantIDStr == "" {
			tenantIDStr = c.GetHeader("X-Tenant-ID")
		}
		if tenantIDStr == "" {
			tenantIDStr = c.Query("tenant_id")
		}

		if strings.TrimSpace(tenantIDStr) == "" {
			response.BadRequest(c, "Target tenant ID is required via URL parameter or X-Tenant-ID header", nil)
			return
		}

		tenantID, err := uuid.Parse(strings.TrimSpace(tenantIDStr))
		if err != nil {
			response.BadRequest(c, "Invalid tenant UUID format", nil)
			return
		}

		// Verify that the tenant exists and is active
		tenant, err := tenantRepo.GetByID(c.Request.Context(), tenantID)
		if err != nil {
			if errors.Is(err, tenants.ErrTenantNotFound) {
				response.NotFound(c, "Requested tenant company does not exist")
				return
			}
			response.InternalServerError(c, "Failed to verify tenant company status")
			return
		}

		if tenant.Status != tenants.StatusActive {
			response.Forbidden(c, "TENANT_INACTIVE", "This tenant company is suspended or deactivated")
			return
		}

		// If user is a Platform Admin, grant administrative tenant access
		if contextutil.IsPlatformAdmin(c) {
			contextutil.SetTenantID(c, tenantID)
			contextutil.SetTenantRole(c, memberships.RolePlatformAdmin)
			c.Next()
			return
		}

		// Verify membership for regular users
		membership, err := membershipRepo.GetByUserAndTenant(c.Request.Context(), userID, tenantID)
		if err != nil {
			if errors.Is(err, memberships.ErrMembershipNotFound) {
				response.Forbidden(c, "CROSS_TENANT_ACCESS_DENIED", "Access denied: You are not a member of this tenant")
				return
			}
			response.InternalServerError(c, "Failed to verify tenant membership")
			return
		}

		if membership.Status != memberships.StatusActive {
			response.Forbidden(c, "MEMBERSHIP_INACTIVE", "Your membership in this tenant is not active")
			return
		}

		// Inject validated tenant context
		contextutil.SetTenantID(c, tenantID)
		contextutil.SetTenantRole(c, membership.Role)
		c.Set(ContextTenantMembership, membership)

		c.Next()
	}
}

// GetTenantID retrieves the validated tenant ID from context.
func GetTenantID(c *gin.Context) (uuid.UUID, bool) {
	return contextutil.GetTenantID(c)
}

// GetTenantRole retrieves the validated role within the current tenant from context.
func GetTenantRole(c *gin.Context) string {
	return contextutil.GetTenantRole(c)
}
