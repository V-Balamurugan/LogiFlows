package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logiflows/logiflows/backend/internal/contextutil"
	"github.com/logiflows/logiflows/backend/internal/response"
)

// Handler exposes HTTP handlers for authentication and user identity endpoints.
type Handler struct {
	service Service
}

// NewHandler initializes a new authentication handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Register handles POST /api/v1/auth/register.
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid registration payload", err.Error())
		return
	}

	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	res, err := h.service.Register(c.Request.Context(), req, ip, userAgent)
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			response.Conflict(c, "An account with this email address already exists")
			return
		}
		if errors.Is(err, ErrPasswordTooShort) ||
			errors.Is(err, ErrPasswordNoUpper) ||
			errors.Is(err, ErrPasswordNoLower) ||
			errors.Is(err, ErrPasswordNoNumber) ||
			errors.Is(err, ErrPasswordNoSymbol) {
			response.BadRequest(c, err.Error(), nil)
			return
		}
		response.InternalServerError(c, "Registration failed due to a server error")
		return
	}

	response.Success(c, http.StatusCreated, res)
}

// Login handles POST /api/v1/auth/login.
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid login credentials format", err.Error())
		return
	}

	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	res, err := h.service.Login(c.Request.Context(), req, ip, userAgent)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.Unauthorized(c, "Invalid email or password")
			return
		}
		if errors.Is(err, ErrAccountDeactivated) {
			response.Forbidden(c, "ACCOUNT_DEACTIVATED", "Your account has been deactivated. Please contact support.")
			return
		}
		response.InternalServerError(c, "Authentication failed due to an unexpected error")
		return
	}

	response.Success(c, http.StatusOK, res)
}

// Me handles GET /api/v1/auth/me.
func (h *Handler) Me(c *gin.Context) {
	userID, ok := contextutil.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Authentication required")
		return
	}

	res, err := h.service.GetCurrentUser(c.Request.Context(), userID)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve current user identity")
		return
	}

	response.Success(c, http.StatusOK, res)
}

// Logout handles POST /api/v1/auth/logout.
func (h *Handler) Logout(c *gin.Context) {
	response.Success(c, http.StatusOK, gin.H{
		"message": "Successfully logged out. Please discard client credentials.",
	})
}
