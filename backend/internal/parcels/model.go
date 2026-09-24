package parcels

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Allowed Service Types
const (
	ServiceTypeStandard  = "STANDARD"
	ServiceTypeExpress   = "EXPRESS"
	ServiceTypeOvernight = "OVERNIGHT"
	ServiceTypeSameDay   = "SAME_DAY"
)

// Allowed Parcel Statuses
const (
	StatusCreated                  = "CREATED"
	StatusBooked                   = "BOOKED"
	StatusReadyForPickup           = "READY_FOR_PICKUP"
	StatusPickedUp                 = "PICKED_UP"
	StatusReceivedAtOriginBranch   = "RECEIVED_AT_ORIGIN_BRANCH"
	StatusInTransit                = "IN_TRANSIT"
	StatusReceivedAtTransferBranch = "RECEIVED_AT_TRANSFER_BRANCH"
	StatusOutForDelivery           = "OUT_FOR_DELIVERY"
	StatusDeliveryAttempted        = "DELIVERY_ATTEMPTED"
	StatusDelivered                = "DELIVERED"
	StatusDeliveryFailed           = "DELIVERY_FAILED"
	StatusReturnInitiated          = "RETURN_INITIATED"
	StatusReturned                 = "RETURNED"
	StatusCancelled                = "CANCELLED"
	StatusOnHold                   = "ON_HOLD"
)

// Allowed Custody Event Types
const (
	CustodyEventIntake           = "INTAKE"
	CustodyEventDispatch         = "DISPATCH"
	CustodyEventReceive          = "RECEIVE"
	CustodyEventHandoverDelivery = "HANDOVER_DELIVERY"
	CustodyEventReturnIntake     = "RETURN_INTAKE"
)

var (
	ErrInvalidSenderDetails     = errors.New("sender_name, sender_phone, and sender_address are required")
	ErrInvalidReceiverDetails   = errors.New("receiver_name, receiver_phone, and receiver_address are required")
	ErrInvalidWeight            = errors.New("weight_kg must be greater than 0")
	ErrInvalidDimensions        = errors.New("dimensions_cm must be in format LxWxH (e.g. 30x20x15)")
	ErrInvalidServiceType       = errors.New("invalid service_type; must be STANDARD, EXPRESS, OVERNIGHT, or SAME_DAY")
	ErrInvalidStatus            = errors.New("invalid status value")
	ErrInvalidBranch            = errors.New("origin_branch_id and destination_branch_id are required")
	ErrSameOriginAndDestination = errors.New("origin_branch_id and destination_branch_id must be different branches")
	ErrParcelNotFound           = errors.New("parcel not found")
	ErrDuplicateTrackingNumber  = errors.New("parcel with this tracking number already exists in this tenant organization")
	ErrInvalidStateTransition   = errors.New("invalid parcel status transition")
	ErrParcelCannotBeCancelled  = errors.New("parcel can only be cancelled in CREATED or BOOKED state")
	ErrBranchCrossTenant        = errors.New("cannot assign parcel to a branch belonging to another organization")
	ErrInvalidQRPayload         = errors.New("invalid or tampered QR code payload")
)

// ValidStateTransitions defines the strict state transition table.
var ValidStateTransitions = map[string][]string{
	StatusCreated: {
		StatusBooked,
		StatusReceivedAtOriginBranch,
		StatusCancelled,
	},
	StatusBooked: {
		StatusReadyForPickup,
		StatusReceivedAtOriginBranch,
		StatusCancelled,
	},
	StatusReadyForPickup: {
		StatusPickedUp,
		StatusCancelled,
	},
	StatusPickedUp: {
		StatusReceivedAtOriginBranch,
	},
	StatusReceivedAtOriginBranch: {
		StatusInTransit,
		StatusOutForDelivery,
		StatusOnHold,
	},
	StatusInTransit: {
		StatusReceivedAtTransferBranch,
		StatusReceivedAtOriginBranch,
		StatusOnHold,
	},
	StatusReceivedAtTransferBranch: {
		StatusOutForDelivery,
		StatusInTransit,
		StatusOnHold,
	},
	StatusOutForDelivery: {
		StatusDelivered,
		StatusDeliveryAttempted,
		StatusDeliveryFailed,
	},
	StatusDeliveryAttempted: {
		StatusOutForDelivery,
		StatusReturnInitiated,
		StatusOnHold,
	},
	StatusOnHold: {
		StatusReceivedAtOriginBranch,
		StatusReceivedAtTransferBranch,
		StatusOutForDelivery,
		StatusReturnInitiated,
		StatusCancelled,
	},
	StatusDeliveryFailed: {
		StatusReturnInitiated,
	},
	StatusReturnInitiated: {
		StatusInTransit,
		StatusReturned,
	},
	// Terminal states: DELIVERED, RETURNED, CANCELLED have no outgoing transitions
	StatusDelivered: {},
	StatusReturned:  {},
	StatusCancelled: {},
}

// IsValidTransition evaluates whether a transition from fromStatus to toStatus is permitted.
func IsValidTransition(fromStatus, toStatus string) bool {
	if fromStatus == toStatus {
		return true // Idempotent same-state updates are allowed
	}
	allowed, exists := ValidStateTransitions[fromStatus]
	if !exists {
		return false
	}
	for _, s := range allowed {
		if s == toStatus {
			return true
		}
	}
	return false
}

// Parcel represents a package in the LogiFlows system.
type Parcel struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	TenantID            uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	TrackingNumber      string     `json:"tracking_number" db:"tracking_number"`
	SenderName          string     `json:"sender_name" db:"sender_name"`
	SenderPhone         string     `json:"sender_phone" db:"sender_phone"`
	SenderEmail         *string    `json:"sender_email,omitempty" db:"sender_email"`
	SenderAddress       string     `json:"sender_address" db:"sender_address"`
	ReceiverName        string     `json:"receiver_name" db:"receiver_name"`
	ReceiverPhone       string     `json:"receiver_phone" db:"receiver_phone"`
	ReceiverEmail       *string    `json:"receiver_email,omitempty" db:"receiver_email"`
	ReceiverAddress     string     `json:"receiver_address" db:"receiver_address"`
	OriginBranchID      uuid.UUID  `json:"origin_branch_id" db:"origin_branch_id"`
	DestinationBranchID uuid.UUID  `json:"destination_branch_id" db:"destination_branch_id"`
	CurrentBranchID     *uuid.UUID `json:"current_branch_id,omitempty" db:"current_branch_id"`
	WeightKG            float64    `json:"weight_kg" db:"weight_kg"`
	DimensionsCM        string     `json:"dimensions_cm" db:"dimensions_cm"`
	ServiceType         string     `json:"service_type" db:"service_type"`
	DeclaredValue       float64    `json:"declared_value" db:"declared_value"`
	Status              string     `json:"status" db:"status"`
	SpecialInstructions *string    `json:"special_instructions,omitempty" db:"special_instructions"`
	QRCodePayload       *string    `json:"qr_code_payload,omitempty" db:"qr_code_payload"`
	CreatedBy           *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt           *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`

	// Joined metadata for UI presentation
	OriginBranchName      *string `json:"origin_branch_name,omitempty" db:"origin_branch_name"`
	OriginBranchCode      *string `json:"origin_branch_code,omitempty" db:"origin_branch_code"`
	DestinationBranchName *string `json:"destination_branch_name,omitempty" db:"destination_branch_name"`
	DestinationBranchCode *string `json:"destination_branch_code,omitempty" db:"destination_branch_code"`
	CurrentBranchName     *string `json:"current_branch_name,omitempty" db:"current_branch_name"`
	CurrentBranchCode     *string `json:"current_branch_code,omitempty" db:"current_branch_code"`
}

// ParcelStatusHistory documents an individual status change.
type ParcelStatusHistory struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	TenantID   uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	ParcelID   uuid.UUID  `json:"parcel_id" db:"parcel_id"`
	FromStatus *string    `json:"from_status,omitempty" db:"from_status"`
	ToStatus   string     `json:"to_status" db:"to_status"`
	BranchID   *uuid.UUID `json:"branch_id,omitempty" db:"branch_id"`
	BranchName *string    `json:"branch_name,omitempty" db:"branch_name"`
	ActorID    *uuid.UUID `json:"actor_id,omitempty" db:"actor_id"`
	ActorName  *string    `json:"actor_name,omitempty" db:"actor_name"`
	ActorRole  *string    `json:"actor_role,omitempty" db:"actor_role"`
	Notes      *string    `json:"notes,omitempty" db:"notes"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// ParcelCustodyEvent tracks physical possession handovers.
type ParcelCustodyEvent struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	TenantID         uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	ParcelID         uuid.UUID  `json:"parcel_id" db:"parcel_id"`
	EmployeeID       *uuid.UUID `json:"employee_id,omitempty" db:"employee_id"`
	EmployeeName     *string    `json:"employee_name,omitempty" db:"employee_name"`
	FromBranchID     *uuid.UUID `json:"from_branch_id,omitempty" db:"from_branch_id"`
	FromBranchName   *string    `json:"from_branch_name,omitempty" db:"from_branch_name"`
	ToBranchID       *uuid.UUID `json:"to_branch_id,omitempty" db:"to_branch_id"`
	ToBranchName     *string    `json:"to_branch_name,omitempty" db:"to_branch_name"`
	EventType        string     `json:"event_type" db:"event_type"`
	SignatureNote    *string    `json:"signature_note,omitempty" db:"signature_note"`
	VerificationCode *string    `json:"verification_code,omitempty" db:"verification_code"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

// CreateParcelRequest payload.
type CreateParcelRequest struct {
	TrackingNumber      string   `json:"tracking_number,omitempty"`
	SenderName          string   `json:"sender_name"`
	SenderPhone         string   `json:"sender_phone"`
	SenderEmail         *string  `json:"sender_email,omitempty"`
	SenderAddress       string   `json:"sender_address"`
	ReceiverName        string   `json:"receiver_name"`
	ReceiverPhone       string   `json:"receiver_phone"`
	ReceiverEmail       *string  `json:"receiver_email,omitempty"`
	ReceiverAddress     string   `json:"receiver_address"`
	OriginBranchID      string   `json:"origin_branch_id"`
	DestinationBranchID string   `json:"destination_branch_id"`
	WeightKG            float64  `json:"weight_kg"`
	DimensionsCM        string   `json:"dimensions_cm"`
	ServiceType         string   `json:"service_type"`
	DeclaredValue       *float64 `json:"declared_value,omitempty"`
	SpecialInstructions *string  `json:"special_instructions,omitempty"`
}

// UpdateParcelRequest payload.
type UpdateParcelRequest struct {
	SenderName          *string  `json:"sender_name,omitempty"`
	SenderPhone         *string  `json:"sender_phone,omitempty"`
	SenderEmail         *string  `json:"sender_email,omitempty"`
	SenderAddress       *string  `json:"sender_address,omitempty"`
	ReceiverName        *string  `json:"receiver_name,omitempty"`
	ReceiverPhone       *string  `json:"receiver_phone,omitempty"`
	ReceiverEmail       *string  `json:"receiver_email,omitempty"`
	ReceiverAddress     *string  `json:"receiver_address,omitempty"`
	WeightKG            *float64 `json:"weight_kg,omitempty"`
	DimensionsCM        *string  `json:"dimensions_cm,omitempty"`
	ServiceType         *string  `json:"service_type,omitempty"`
	DeclaredValue       *float64 `json:"declared_value,omitempty"`
	SpecialInstructions *string  `json:"special_instructions,omitempty"`
}

// UpdateParcelStatusRequest payload.
type UpdateParcelStatusRequest struct {
	Status   string  `json:"status"`
	BranchID *string `json:"branch_id,omitempty"`
	Notes    *string `json:"notes,omitempty"`
}

// ScanParcelRequest payload for barcode/QR scanner.
type ScanParcelRequest struct {
	QRPayload      *string `json:"qr_payload,omitempty"`
	TrackingNumber *string `json:"tracking_number,omitempty"`
	BranchID       *string `json:"branch_id,omitempty"`
	Notes          *string `json:"notes,omitempty"`
}

// ScanParcelResponse returned upon successful scan.
type ScanParcelResponse struct {
	Parcel         *Parcel   `json:"parcel"`
	VerifiedAt     time.Time `json:"verified_at"`
	CurrentBranch  *string   `json:"current_branch,omitempty"`
	NextAction     string    `json:"next_action"`
	CustodyHolder  *string   `json:"custody_holder,omitempty"`
}

// PublicTrackingMilestone public item.
type PublicTrackingMilestone struct {
	Status      string    `json:"status"`
	Description string    `json:"description"`
	Location    string    `json:"location,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// PublicTrackingResponse sanitized response for unauthenticated customer tracking.
type PublicTrackingResponse struct {
	TrackingNumber      string                    `json:"tracking_number"`
	Status              string                    `json:"status"`
	ServiceType         string                    `json:"service_type"`
	OriginCity          string                    `json:"origin_city"`
	DestinationCity     string                    `json:"destination_city"`
	WeightKG            float64                   `json:"weight_kg"`
	CreatedAt           time.Time                 `json:"created_at"`
	EstimatedDelivery   *time.Time                `json:"estimated_delivery,omitempty"`
	Milestones          []PublicTrackingMilestone `json:"milestones"`
}

// ParcelFilter options for listing.
type ParcelFilter struct {
	Status         *string
	OriginBranchID *uuid.UUID
	DestBranchID   *uuid.UUID
	CurrentBranchID *uuid.UUID
	Search         *string
	DateFrom       *time.Time
	DateTo         *time.Time
	Limit          int
	Offset         int
}

// ParcelListResponse wraps list results with pagination metadata.
type ParcelListResponse struct {
	Parcels []Parcel `json:"parcels"`
	Total   int      `json:"total"`
	Limit   int      `json:"limit"`
	Offset  int      `json:"offset"`
}

// ValidateCreateRequest verifies required fields and invariants for parcel creation.
func ValidateCreateRequest(req CreateParcelRequest) error {
	if strings.TrimSpace(req.SenderName) == "" || strings.TrimSpace(req.SenderPhone) == "" || strings.TrimSpace(req.SenderAddress) == "" {
		return ErrInvalidSenderDetails
	}
	if strings.TrimSpace(req.ReceiverName) == "" || strings.TrimSpace(req.ReceiverPhone) == "" || strings.TrimSpace(req.ReceiverAddress) == "" {
		return ErrInvalidReceiverDetails
	}
	if req.WeightKG <= 0 {
		return ErrInvalidWeight
	}
	if strings.TrimSpace(req.DimensionsCM) == "" {
		return ErrInvalidDimensions
	}
	validServiceTypes := map[string]bool{
		ServiceTypeStandard:  true,
		ServiceTypeExpress:   true,
		ServiceTypeOvernight: true,
		ServiceTypeSameDay:   true,
	}
	if !validServiceTypes[strings.ToUpper(strings.TrimSpace(req.ServiceType))] {
		return ErrInvalidServiceType
	}
	if strings.TrimSpace(req.OriginBranchID) == "" || strings.TrimSpace(req.DestinationBranchID) == "" {
		return ErrInvalidBranch
	}
	if strings.EqualFold(strings.TrimSpace(req.OriginBranchID), strings.TrimSpace(req.DestinationBranchID)) {
		return ErrSameOriginAndDestination
	}
	return nil
}

// GenerateQRPayload creates a tamper-evident JSON QR string for a parcel.
func GenerateQRPayload(trackingNumber string, tenantID, parcelID uuid.UUID) string {
	checksumSource := fmt.Sprintf("%s:%s:%s", trackingNumber, tenantID.String(), parcelID.String())
	h := sha256.Sum256([]byte(checksumSource))
	checksum := hex.EncodeToString(h[:8])

	payload := map[string]string{
		"tracking":  trackingNumber,
		"tenant_id": tenantID.String(),
		"parcel_id": parcelID.String(),
		"chk":       checksum,
	}
	b, _ := json.Marshal(payload)
	return string(b)
}

// VerifyQRPayload validates a tamper-evident QR payload.
func VerifyQRPayload(payloadStr string) (trackingNumber string, tenantID, parcelID uuid.UUID, err error) {
	var payload map[string]string
	if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
		return "", uuid.Nil, uuid.Nil, ErrInvalidQRPayload
	}

	tracking := payload["tracking"]
	tIDStr := payload["tenant_id"]
	pIDStr := payload["parcel_id"]
	chk := payload["chk"]

	if tracking == "" || tIDStr == "" || pIDStr == "" || chk == "" {
		return "", uuid.Nil, uuid.Nil, ErrInvalidQRPayload
	}

	tID, err := uuid.Parse(tIDStr)
	if err != nil {
		return "", uuid.Nil, uuid.Nil, ErrInvalidQRPayload
	}
	pID, err := uuid.Parse(pIDStr)
	if err != nil {
		return "", uuid.Nil, uuid.Nil, ErrInvalidQRPayload
	}

	checksumSource := fmt.Sprintf("%s:%s:%s", tracking, tID.String(), pID.String())
	h := sha256.Sum256([]byte(checksumSource))
	expectedChk := hex.EncodeToString(h[:8])

	if chk != expectedChk {
		return "", uuid.Nil, uuid.Nil, ErrInvalidQRPayload
	}

	return tracking, tID, pID, nil
}
