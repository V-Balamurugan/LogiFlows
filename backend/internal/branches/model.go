package branches

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	StatusActive    = "ACTIVE"
	StatusInactive  = "INACTIVE"
	StatusSuspended = "SUSPENDED"

	DefaultCoverageRadiusKM = 15.0
)

var (
	ErrBranchNotFound        = errors.New("branch not found")
	ErrBranchCodeExists      = errors.New("a branch with this code already exists for this tenant")
	ErrInvalidCoordinates    = errors.New("invalid latitude or longitude coordinates")
	ErrInvalidBranchStatus   = errors.New("invalid branch status. Supported: ACTIVE, INACTIVE, SUSPENDED")
	ErrInvalidCoverageRadius = errors.New("coverage radius must be greater than zero")
)

// Branch represents a delivery hub, fulfillment depot, or sorting center.
type Branch struct {
	ID               uuid.UUID `json:"id"`
	TenantID         uuid.UUID `json:"tenant_id"`
	BranchCode       string    `json:"branch_code"`
	Name             string    `json:"name"`
	Address          string    `json:"address"`
	City             string    `json:"city"`
	State            string    `json:"state,omitempty"`
	PostalCode       string    `json:"postal_code,omitempty"`
	Country          string    `json:"country"`
	Latitude         float64   `json:"latitude"`
	Longitude        float64   `json:"longitude"`
	CoverageRadiusKM float64   `json:"coverage_radius_km"`
	Status           string    `json:"status"`
	IsActive         bool      `json:"is_active"`
	DistanceKM       *float64  `json:"distance_km,omitempty"` // populated during spatial queries
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// CreateBranchRequest encapsulates input required to register a new branch.
type CreateBranchRequest struct {
	BranchCode       string  `json:"branch_code" binding:"required,min=2,max=50"`
	Name             string  `json:"name" binding:"required,min=2,max=255"`
	Address          string  `json:"address" binding:"required"`
	City             string  `json:"city" binding:"required"`
	State            string  `json:"state"`
	PostalCode       string  `json:"postal_code"`
	Country          string  `json:"country"`
	Latitude         float64 `json:"latitude" binding:"required"`
	Longitude        float64 `json:"longitude" binding:"required"`
	CoverageRadiusKM float64 `json:"coverage_radius_km"`
}

// Validate validates the CreateBranchRequest values.
func (r *CreateBranchRequest) Validate() error {
	r.BranchCode = strings.ToUpper(strings.TrimSpace(r.BranchCode))
	r.Name = strings.TrimSpace(r.Name)
	r.Address = strings.TrimSpace(r.Address)
	r.City = strings.TrimSpace(r.City)
	r.State = strings.TrimSpace(r.State)
	r.PostalCode = strings.TrimSpace(r.PostalCode)
	if strings.TrimSpace(r.Country) == "" {
		r.Country = "India"
	} else {
		r.Country = strings.TrimSpace(r.Country)
	}

	if r.Latitude < -90.0 || r.Latitude > 90.0 || r.Longitude < -180.0 || r.Longitude > 180.0 {
		return ErrInvalidCoordinates
	}

	if r.CoverageRadiusKM <= 0 {
		r.CoverageRadiusKM = DefaultCoverageRadiusKM
	}

	return nil
}

// UpdateBranchRequest encapsulates fields that can be modified on an existing branch.
type UpdateBranchRequest struct {
	Name             *string  `json:"name"`
	Address          *string  `json:"address"`
	City             *string  `json:"city"`
	State            *string  `json:"state"`
	PostalCode       *string  `json:"postal_code"`
	Country          *string  `json:"country"`
	Latitude         *float64 `json:"latitude"`
	Longitude        *float64 `json:"longitude"`
	CoverageRadiusKM *float64 `json:"coverage_radius_km"`
	Status           *string  `json:"status"`
}

// UpdateBranchStatusRequest encapsulates payload for setting branch operating status.
type UpdateBranchStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// BranchFilter specifies query parameters for branch listing and spatial search.
type BranchFilter struct {
	Page     int      `form:"page"`
	Limit    int      `form:"limit"`
	Search   string   `form:"search"`
	Status   string   `form:"status"`
	IsActive *bool    `form:"is_active"`
	NearLat  *float64 `form:"near_lat"`
	NearLng  *float64 `form:"near_lng"`
	RadiusKM *float64 `form:"radius_km"`
}

// BranchListResponse encapsulates the paginated response for branch listing.
type BranchListResponse struct {
	Branches []Branch `json:"branches"`
	Total    int      `json:"total"`
	Page     int      `json:"page"`
	Limit    int      `json:"limit"`
}
