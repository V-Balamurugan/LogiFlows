package users

import (
	"time"

	"github.com/google/uuid"
)

// User represents a system user entity in LogiFlows.
type User struct {
	ID              uuid.UUID  `json:"id"`
	Email           string     `json:"email"`
	PasswordHash    string     `json:"-"` // Explicitly excluded from all JSON serializations
	FullName        string     `json:"full_name"`
	PhoneNumber     *string    `json:"phone_number,omitempty"`
	IsActive        bool       `json:"is_active"`
	IsPlatformAdmin bool       `json:"is_platform_admin"`
	EmailVerified   bool       `json:"email_verified"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
}
