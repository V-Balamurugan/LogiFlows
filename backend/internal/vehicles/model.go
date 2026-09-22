package vehicles

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Allowed Vehicle Types
const (
	VehicleTypeElectricVan  = "ELECTRIC_VAN"
	VehicleTypeVan          = "VAN"
	VehicleTypeMotorcycle   = "MOTORCYCLE"
	VehicleTypeTruck        = "TRUCK"
	VehicleTypeThreeWheeler = "THREE_WHEELER"
)

// Allowed Vehicle Statuses
const (
	VehicleStatusAvailable      = "AVAILABLE"
	VehicleStatusAssigned       = "ASSIGNED"
	VehicleStatusInTransit      = "IN_TRANSIT"
	VehicleStatusMaintenance    = "MAINTENANCE"
	VehicleStatusDecommissioned = "DECOMMISSIONED"
)

// Allowed Vehicle Assignment Statuses
const (
	AssignmentStatusActive    = "ACTIVE"
	AssignmentStatusCompleted = "COMPLETED"
	AssignmentStatusCancelled = "CANCELLED"
)

var (
	ErrInvalidRegistrationNumber   = errors.New("registration_number is required and must be between 2 and 50 characters")
	ErrInvalidVehicleType          = errors.New("invalid vehicle_type; must be ELECTRIC_VAN, VAN, MOTORCYCLE, TRUCK, or THREE_WHEELER")
	ErrInvalidCapacity             = errors.New("max_weight_kg and max_volume_cbm must be strictly greater than 0")
	ErrDuplicateRegistrationNumber = errors.New("vehicle with this registration number already exists in this tenant organization")
	ErrVehicleNotFound             = errors.New("vehicle not found")
	ErrVehicleAlreadyAssigned      = errors.New("vehicle is already actively assigned to another driver")
	ErrDriverAlreadyAssigned       = errors.New("driver is already actively assigned to another vehicle")
	ErrDriverNotFound              = errors.New("driver not found or is inactive")
	ErrBranchNotFound              = errors.New("assigned branch not found or is inactive")
	ErrBranchCrossTenant           = errors.New("cannot associate vehicle with branch belonging to another organization")
	ErrAssignmentNotFound          = errors.New("active vehicle assignment not found")
)

type Vehicle struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	TenantID           uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	BranchID           *uuid.UUID `json:"branch_id,omitempty" db:"branch_id"`
	RegistrationNumber string     `json:"registration_number" db:"registration_number"`
	VehicleType        string     `json:"vehicle_type" db:"vehicle_type"`
	MakeModel          *string    `json:"make_model,omitempty" db:"make_model"`
	Year               *int       `json:"year,omitempty" db:"year"`
	MaxWeightKG        float64    `json:"max_weight_kg" db:"max_weight_kg"`
	MaxVolumeCBM       float64    `json:"max_volume_cbm" db:"max_volume_cbm"`
	Status             string     `json:"status" db:"status"`
	IsActive           bool       `json:"is_active" db:"is_active"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`

	// Joined details for rich client telemetry & UI
	BranchName        *string    `json:"branch_name,omitempty" db:"branch_name"`
	BranchCode        *string    `json:"branch_code,omitempty" db:"branch_code"`
	CurrentDriverID   *uuid.UUID `json:"current_driver_id,omitempty" db:"current_driver_id"`
	CurrentDriverName *string    `json:"current_driver_name,omitempty" db:"current_driver_name"`
}

type VehicleAssignment struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	TenantID     uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	VehicleID    uuid.UUID  `json:"vehicle_id" db:"vehicle_id"`
	DriverID     uuid.UUID  `json:"driver_id" db:"driver_id"`
	AssignedBy   *uuid.UUID `json:"assigned_by,omitempty" db:"assigned_by"`
	AssignedAt   time.Time  `json:"assigned_at" db:"assigned_at"`
	UnassignedAt *time.Time `json:"unassigned_at,omitempty" db:"unassigned_at"`
	Status       string     `json:"status" db:"status"`
	Notes        *string    `json:"notes,omitempty" db:"notes"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`

	// Joined metadata
	RegistrationNumber *string `json:"registration_number,omitempty" db:"registration_number"`
	DriverName         *string `json:"driver_name,omitempty" db:"driver_name"`
}

type CreateVehicleRequest struct {
	RegistrationNumber string     `json:"registration_number" binding:"required"`
	VehicleType        string     `json:"vehicle_type" binding:"required"`
	MakeModel          *string    `json:"make_model,omitempty"`
	Year               *int       `json:"year,omitempty"`
	MaxWeightKG        float64    `json:"max_weight_kg"`
	MaxVolumeCBM       float64    `json:"max_volume_cbm"`
	BranchID           *uuid.UUID `json:"branch_id,omitempty"`
}

func (r *CreateVehicleRequest) ValidateAndSanitize() error {
	r.RegistrationNumber = strings.ToUpper(strings.TrimSpace(r.RegistrationNumber))
	if len(r.RegistrationNumber) < 2 || len(r.RegistrationNumber) > 50 {
		return ErrInvalidRegistrationNumber
	}

	r.VehicleType = strings.ToUpper(strings.TrimSpace(r.VehicleType))
	switch r.VehicleType {
	case VehicleTypeElectricVan, VehicleTypeVan, VehicleTypeMotorcycle, VehicleTypeTruck, VehicleTypeThreeWheeler:
	default:
		return ErrInvalidVehicleType
	}

	if r.MaxWeightKG <= 0 {
		r.MaxWeightKG = 500.0
	}
	if r.MaxVolumeCBM <= 0 {
		r.MaxVolumeCBM = 3.0
	}

	return nil
}

type UpdateVehicleRequest struct {
	VehicleType  *string    `json:"vehicle_type,omitempty"`
	MakeModel    *string    `json:"make_model,omitempty"`
	Year         *int       `json:"year,omitempty"`
	MaxWeightKG  *float64   `json:"max_weight_kg,omitempty"`
	MaxVolumeCBM *float64   `json:"max_volume_cbm,omitempty"`
	BranchID     *uuid.UUID `json:"branch_id,omitempty"`
	Status       *string    `json:"status,omitempty"`
}

func (r *UpdateVehicleRequest) ValidateAndSanitize() error {
	if r.VehicleType != nil {
		vt := strings.ToUpper(strings.TrimSpace(*r.VehicleType))
		switch vt {
		case VehicleTypeElectricVan, VehicleTypeVan, VehicleTypeMotorcycle, VehicleTypeTruck, VehicleTypeThreeWheeler:
			*r.VehicleType = vt
		default:
			return ErrInvalidVehicleType
		}
	}

	if r.Status != nil {
		st := strings.ToUpper(strings.TrimSpace(*r.Status))
		switch st {
		case VehicleStatusAvailable, VehicleStatusAssigned, VehicleStatusInTransit, VehicleStatusMaintenance, VehicleStatusDecommissioned:
			*r.Status = st
		default:
			return errors.New("invalid status; must be AVAILABLE, ASSIGNED, IN_TRANSIT, MAINTENANCE, or DECOMMISSIONED")
		}
	}

	if r.MaxWeightKG != nil && *r.MaxWeightKG <= 0 {
		return ErrInvalidCapacity
	}
	if r.MaxVolumeCBM != nil && *r.MaxVolumeCBM <= 0 {
		return ErrInvalidCapacity
	}

	return nil
}

type AssignVehicleRequest struct {
	DriverID uuid.UUID `json:"driver_id" binding:"required"`
	Notes    *string   `json:"notes,omitempty"`
}

type VehicleFilter struct {
	Search      string
	VehicleType string
	Status      string
	BranchID    *uuid.UUID
	Limit       int
	Offset      int
}
