package auth

import (
	"time"

	"github.com/google/uuid"
)

// RegisterRequest holds the payload for registering a new tenant company & user.
type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	FullName    string `json:"full_name" binding:"required"`
	PhoneNumber string `json:"phone_number"`
	CompanyName string `json:"company_name" binding:"required"`
}

// LoginRequest holds the credentials for user authentication.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UserSummary represents a safe user projection for client responses.
type UserSummary struct {
	ID              uuid.UUID `json:"id"`
	Email           string    `json:"email"`
	FullName        string    `json:"full_name"`
	PhoneNumber     *string   `json:"phone_number,omitempty"`
	IsActive        bool      `json:"is_active"`
	IsPlatformAdmin bool      `json:"is_platform_admin"`
}

// TenantSummary represents the tenant context and caller's role in that tenant.
type TenantSummary struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
	Role string    `json:"role"`
}

// AuthResponse is returned on successful registration or login.
type AuthResponse struct {
	Token     string          `json:"token"`
	ExpiresAt time.Time       `json:"expires_at"`
	User      UserSummary     `json:"user"`
	Tenants   []TenantSummary `json:"tenants"`
}

// CurrentUserResponse is returned by GET /api/v1/auth/me.
type CurrentUserResponse struct {
	User    UserSummary     `json:"user"`
	Tenants []TenantSummary `json:"tenants"`
}
