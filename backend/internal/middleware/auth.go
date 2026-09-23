package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/logiflows/logiflows/backend/internal/auth"
	"github.com/logiflows/logiflows/backend/internal/contextutil"
	"github.com/logiflows/logiflows/backend/internal/response"
	"github.com/logiflows/logiflows/backend/internal/users"
)

const (
	ContextUser = "userEntity"
)

// Auth authenticates incoming requests via the Bearer JWT token in the Authorization header.
func Auth(tokenService *auth.TokenService, userRepo users.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Missing Authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Unauthorized(c, "Invalid Authorization header format. Expected 'Bearer <token>'")
			return
		}

		tokenStr := strings.TrimSpace(parts[1])
		claims, err := tokenService.ValidateAccessToken(tokenStr)
		if err != nil {
			if err == auth.ErrExpiredToken {
				response.Unauthorized(c, "Authentication token has expired")
				return
			}
			response.Unauthorized(c, "Invalid or corrupted authentication token")
			return
		}

		// Verify that the user exists and is active
		user, err := userRepo.GetByID(c.Request.Context(), claims.UserID)
		if err != nil || !user.IsActive {
			response.Unauthorized(c, "User account is deactivated or no longer exists")
			return
		}

		// Inject verified user metadata into request context via contextutil
		contextutil.SetUserID(c, user.ID)
		contextutil.SetUserEmail(c, user.Email)
		contextutil.SetIsPlatformAdmin(c, user.IsPlatformAdmin)
		c.Set(ContextUser, user)

		c.Next()
	}
}

// GetUserID retrieves the authenticated user's ID from context.
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	return contextutil.GetUserID(c)
}

// IsPlatformAdmin returns true if the authenticated user is a Platform Admin.
func IsPlatformAdmin(c *gin.Context) bool {
	return contextutil.IsPlatformAdmin(c)
}

// GetUser retrieves the cached User domain entity from context.
func GetUser(c *gin.Context) *users.User {
	val, exists := c.Get(ContextUser)
	if !exists {
		return nil
	}
	u, ok := val.(*users.User)
	if !ok {
		return nil
	}
	return u
}
