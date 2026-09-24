package transfers

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Allowed Branch Transfer Statuses
const (
	StatusPending           = "PENDING"
	StatusInTransit         = "IN_TRANSIT"
	StatusReceived          = "RECEIVED"
	StatusCancelled         = "CANCELLED"
	StatusPartiallyReceived = "PARTIALLY_RECEIVED"
)

var (
	ErrTransferNotFound        = errors.New("branch transfer manifest not found")
	ErrInvalidBranches         = errors.New("source_branch_id and destination_branch_id are required")
	ErrSameBranchTransfer      = errors.New("source and destination branches must be different")
	ErrNoParcelsSpecified      = errors.New("at least one parcel_id is required to create a transfer manifest")
	ErrInvalidStatusAction     = errors.New("invalid action for current transfer status")
	ErrTransferAlreadyReceived = errors.New("transfer manifest is already received")
)

// BranchTransfer represents an inter-branch linehaul manifest.
type BranchTransfer struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	TenantID            uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	TransferNumber      string     `json:"transfer_number" db:"transfer_number"`
	SourceBranchID      uuid.UUID  `json:"source_branch_id" db:"source_branch_id"`
	DestinationBranchID uuid.UUID  `json:"destination_branch_id" db:"destination_branch_id"`
	DriverID            *uuid.UUID `json:"driver_id,omitempty" db:"driver_id"`
	VehicleID           *uuid.UUID `json:"vehicle_id,omitempty" db:"vehicle_id"`
	Status              string     `json:"status" db:"status"`
	DispatchedAt        *time.Time `json:"dispatched_at,omitempty" db:"dispatched_at"`
	ReceivedAt          *time.Time `json:"received_at,omitempty" db:"received_at"`
	Notes               *string    `json:"notes,omitempty" db:"notes"`
	CreatedBy           *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`

	// Joined details for UI
	SourceBranchName      *string `json:"source_branch_name,omitempty" db:"source_branch_name"`
	SourceBranchCode      *string `json:"source_branch_code,omitempty" db:"source_branch_code"`
	DestinationBranchName *string `json:"destination_branch_name,omitempty" db:"destination_branch_name"`
	DestinationBranchCode *string `json:"destination_branch_code,omitempty" db:"destination_branch_code"`
	DriverName            *string `json:"driver_name,omitempty" db:"driver_name"`
	DriverCode            *string `json:"driver_code,omitempty" db:"driver_code"`
	VehicleRegNum         *string `json:"vehicle_registration_number,omitempty" db:"vehicle_registration_number"`
	ParcelsCount          int     `json:"parcels_count" db:"parcels_count"`
}

// BranchTransferParcel represents a parcel within a transfer manifest.
type BranchTransferParcel struct {
	TransferID     uuid.UUID  `json:"transfer_id" db:"transfer_id"`
	ParcelID       uuid.UUID  `json:"parcel_id" db:"parcel_id"`
	Received       bool       `json:"received" db:"received"`
	ReceivedAt     *time.Time `json:"received_at,omitempty" db:"received_at"`
	TrackingNumber string     `json:"tracking_number" db:"tracking_number"`
	WeightKG       float64    `json:"weight_kg" db:"weight_kg"`
	Status         string     `json:"status" db:"status"`
	ReceiverName   string     `json:"receiver_name" db:"receiver_name"`
}

// CreateBranchTransferRequest payload.
type CreateBranchTransferRequest struct {
	SourceBranchID      string   `json:"source_branch_id"`
	DestinationBranchID string   `json:"destination_branch_id"`
	DriverID            *string  `json:"driver_id,omitempty"`
	VehicleID           *string  `json:"vehicle_id,omitempty"`
	ParcelIDs           []string `json:"parcel_ids"`
	Notes               *string  `json:"notes,omitempty"`
}

// DispatchTransferRequest payload.
type DispatchTransferRequest struct {
	Notes *string `json:"notes,omitempty"`
}

// ReceiveTransferRequest payload.
type ReceiveTransferRequest struct {
	ParcelIDs []string `json:"parcel_ids,omitempty"`
	Notes     *string  `json:"notes,omitempty"`
}

// BranchTransferFilter options.
type BranchTransferFilter struct {
	Status              *string
	SourceBranchID      *uuid.UUID
	DestinationBranchID *uuid.UUID
	Limit               int
	Offset              int
}

// BranchTransferListResponse wrapper.
type BranchTransferListResponse struct {
	Transfers []BranchTransfer `json:"transfers"`
	Total     int              `json:"total"`
	Limit     int              `json:"limit"`
	Offset    int              `json:"offset"`
}

// ValidateCreateRequest verifies required fields for transfer creation.
func ValidateCreateRequest(req CreateBranchTransferRequest) error {
	if strings.TrimSpace(req.SourceBranchID) == "" || strings.TrimSpace(req.DestinationBranchID) == "" {
		return ErrInvalidBranches
	}
	if strings.EqualFold(strings.TrimSpace(req.SourceBranchID), strings.TrimSpace(req.DestinationBranchID)) {
		return ErrSameBranchTransfer
	}
	if len(req.ParcelIDs) == 0 {
		return ErrNoParcelsSpecified
	}
	return nil
}
