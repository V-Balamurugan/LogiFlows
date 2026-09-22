package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/logiflows/logiflows/backend/internal/authorization"
	"github.com/logiflows/logiflows/backend/internal/response"
)

// RequireRole enforces role-based access control against the validated tenant role or platform admin status.
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		isPlatformAdmin := IsPlatformAdmin(c)
		tenantRole := GetTenantRole(c)

		if authorization.HasRole(tenantRole, isPlatformAdmin, allowedRoles...) {
			c.Next()
			return
		}

		response.Forbidden(c, "INSUFFICIENT_PERMISSIONS", "Your role does not grant permission to perform this action")
	}
}
