package contextutil

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	KeyUserID          = "userID"
	KeyUserEmail       = "userEmail"
	KeyIsPlatformAdmin = "isPlatformAdmin"
	KeyTenantID        = "tenantID"
	KeyTenantRole      = "tenantRole"
)

// SetUserID sets the authenticated user ID in the Gin context.
func SetUserID(c *gin.Context, id uuid.UUID) {
	c.Set(KeyUserID, id)
}

// GetUserID retrieves the authenticated user ID from the Gin context.
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(KeyUserID)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}

// SetUserEmail sets the authenticated user email in the Gin context.
func SetUserEmail(c *gin.Context, email string) {
	c.Set(KeyUserEmail, email)
}

// GetUserEmail retrieves the authenticated user email from the Gin context.
func GetUserEmail(c *gin.Context) string {
	val, exists := c.Get(KeyUserEmail)
	if !exists {
		return ""
	}
	email, ok := val.(string)
	if !ok {
		return ""
	}
	return email
}

// SetIsPlatformAdmin sets the platform admin flag in the Gin context.
func SetIsPlatformAdmin(c *gin.Context, isAdmin bool) {
	c.Set(KeyIsPlatformAdmin, isAdmin)
}

// IsPlatformAdmin checks if the caller is flagged as a platform admin.
func IsPlatformAdmin(c *gin.Context) bool {
	val, exists := c.Get(KeyIsPlatformAdmin)
	if !exists {
		return false
	}
	isAdmin, ok := val.(bool)
	return ok && isAdmin
}

// SetTenantID sets the validated tenant ID in the Gin context.
func SetTenantID(c *gin.Context, id uuid.UUID) {
	c.Set(KeyTenantID, id)
}

// GetTenantID retrieves the validated tenant ID from the Gin context.
func GetTenantID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(KeyTenantID)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}

// SetTenantRole sets the validated role within the active tenant.
func SetTenantRole(c *gin.Context, role string) {
	c.Set(KeyTenantRole, role)
}

// GetTenantRole retrieves the validated role within the active tenant.
func GetTenantRole(c *gin.Context) string {
	val, exists := c.Get(KeyTenantRole)
	if !exists {
		return ""
	}
	role, ok := val.(string)
	if !ok {
		return ""
	}
	return role
}
