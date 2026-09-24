package deliveries

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Allowed Delivery Task Statuses
const (
	StatusAssigned    = "ASSIGNED"
	StatusInProgress  = "IN_PROGRESS"
	StatusCompleted   = "COMPLETED"
	StatusFailed      = "FAILED"
	StatusCancelled   = "CANCELLED"
	StatusRescheduled = "RESCHEDULED"
)

// Allowed Delivery Task Priorities
const (
	PriorityLow      = "LOW"
	PriorityStandard = "STANDARD"
	PriorityHigh     = "HIGH"
	PriorityUrgent   = "URGENT"
)

// Allowed Delivery Attempt Outcomes
const (
	OutcomeSuccessful           = "SUCCESSFUL"
	OutcomeCustomerUnavailable  = "CUSTOMER_UNAVAILABLE"
	OutcomeIncorrectAddress     = "INCORRECT_ADDRESS"
	OutcomeRejectedByCustomer   = "REJECTED_BY_CUSTOMER"
	OutcomeSecurityAccessDenied = "SECURITY_ACCESS_DENIED"
	OutcomeOther                = "OTHER"
)

// Allowed Delivery Proof Types
const (
	ProofTypeRecipientSignature = "RECIPIENT_SIGNATURE"
	ProofTypeOTP                = "OTP"
	ProofTypePhoto              = "PHOTO"
	ProofTypeContactlessDrop    = "CONTACTLESS_DROP"
)

var (
	ErrDeliveryTaskNotFound         = errors.New("delivery task not found")
	ErrParcelNotEligibleForDelivery = errors.New("parcel is not in an eligible status for last-mile delivery assignment")
	ErrDriverNotEligible            = errors.New("assigned employee must have DRIVER or DELIVERY_EXECUTIVE operational role and be currently active")
	ErrParcelAlreadyAssigned        = errors.New("parcel already has an active delivery task assigned")
	ErrDeliveryAlreadyCompleted     = errors.New("delivery task is already completed or cancelled")
	ErrInvalidAttemptOutcome        = errors.New("invalid delivery attempt outcome")
	ErrInvalidProofType             = errors.New("invalid delivery proof type; must be RECIPIENT_SIGNATURE, OTP, PHOTO, or CONTACTLESS_DROP")
	ErrMissingRecipientName         = errors.New("recipient_name is required for proof of delivery")
	ErrInvalidStatusTransition      = errors.New("invalid delivery task status transition")
)

// DeliveryTask represents a last-mile courier delivery assignment.
type DeliveryTask struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	TenantID         uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	ParcelID         uuid.UUID  `json:"parcel_id" db:"parcel_id"`
	BranchID         uuid.UUID  `json:"branch_id" db:"branch_id"`
	AssignedDriverID uuid.UUID  `json:"assigned_driver_id" db:"assigned_driver_id"`
	VehicleID        *uuid.UUID `json:"vehicle_id,omitempty" db:"vehicle_id"`
	Status           string     `json:"status" db:"status"`
	Priority         string     `json:"priority" db:"priority"`
	AssignedAt       time.Time  `json:"assigned_at" db:"assigned_at"`
	StartedAt        *time.Time `json:"started_at,omitempty" db:"started_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	FailedAt         *time.Time `json:"failed_at,omitempty" db:"failed_at"`
	FailureReason    *string    `json:"failure_reason,omitempty" db:"failure_reason"`
	RescheduledFor   *time.Time `json:"rescheduled_for,omitempty" db:"rescheduled_for"`
	Notes            *string    `json:"notes,omitempty" db:"notes"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`

	// Joined details for rich client UI
	TrackingNumber  *string  `json:"tracking_number,omitempty" db:"tracking_number"`
	SenderName      *string  `json:"sender_name,omitempty" db:"sender_name"`
	ReceiverName    *string  `json:"receiver_name,omitempty" db:"receiver_name"`
	ReceiverPhone   *string  `json:"receiver_phone,omitempty" db:"receiver_phone"`
	ReceiverAddress *string  `json:"receiver_address,omitempty" db:"receiver_address"`
	ParcelStatus    *string  `json:"parcel_status,omitempty" db:"parcel_status"`
	WeightKG        *float64 `json:"weight_kg,omitempty" db:"weight_kg"`
	DriverName      *string  `json:"driver_name,omitempty" db:"driver_name"`
	DriverCode      *string  `json:"driver_code,omitempty" db:"driver_code"`
	VehicleRegNum   *string  `json:"vehicle_registration_number,omitempty" db:"vehicle_registration_number"`
	BranchName      *string  `json:"branch_name,omitempty" db:"branch_name"`
	AttemptsCount   int      `json:"attempts_count" db:"attempts_count"`
}

// DeliveryAttempt represents an attempted delivery drop.
type DeliveryAttempt struct {
	ID             uuid.UUID `json:"id" db:"id"`
	DeliveryTaskID uuid.UUID `json:"delivery_task_id" db:"delivery_task_id"`
	AttemptNumber  int       `json:"attempt_number" db:"attempt_number"`
	AttemptedAt    time.Time `json:"attempted_at" db:"attempted_at"`
	Outcome        string    `json:"outcome" db:"outcome"`
	Notes          *string   `json:"notes,omitempty" db:"notes"`
	Latitude       *float64  `json:"latitude,omitempty" db:"latitude"`
	Longitude      *float64  `json:"longitude,omitempty" db:"longitude"`
}

// DeliveryProof represents the immutable record validating completion.
type DeliveryProof struct {
	ID                    uuid.UUID `json:"id" db:"id"`
	DeliveryTaskID        uuid.UUID `json:"delivery_task_id" db:"delivery_task_id"`
	ParcelID              uuid.UUID `json:"parcel_id" db:"parcel_id"`
	ProofType             string    `json:"proof_type" db:"proof_type"`
	RecipientName         string    `json:"recipient_name" db:"recipient_name"`
	RecipientRelationship string    `json:"recipient_relationship" db:"recipient_relationship"`
	OTPCode               *string   `json:"otp_code,omitempty" db:"otp_code"`
	SignatureData         *string   `json:"signature_data,omitempty" db:"signature_data"`
	PhotoURL              *string   `json:"photo_url,omitempty" db:"photo_url"`
	Notes                 *string   `json:"notes,omitempty" db:"notes"`
	Latitude              *float64  `json:"latitude,omitempty" db:"latitude"`
	Longitude             *float64  `json:"longitude,omitempty" db:"longitude"`
	VerifiedAt            time.Time `json:"verified_at" db:"verified_at"`
}

// CreateDeliveryTaskRequest payload.
type CreateDeliveryTaskRequest struct {
	ParcelID         string  `json:"parcel_id"`
	AssignedDriverID string  `json:"assigned_driver_id"`
	VehicleID        *string `json:"vehicle_id,omitempty"`
	Priority         *string `json:"priority,omitempty"`
	Notes            *string `json:"notes,omitempty"`
}

// UpdateDeliveryTaskStatusRequest payload.
type UpdateDeliveryTaskStatusRequest struct {
	Status        string  `json:"status"`
	Notes         *string `json:"notes,omitempty"`
	FailureReason *string `json:"failure_reason,omitempty"`
}

// RecordDeliveryAttemptRequest payload.
type RecordDeliveryAttemptRequest struct {
	Outcome   string   `json:"outcome"`
	Notes     *string  `json:"notes,omitempty"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
}

// SubmitDeliveryProofRequest payload.
type SubmitDeliveryProofRequest struct {
	ProofType             string   `json:"proof_type"`
	RecipientName         string   `json:"recipient_name"`
	RecipientRelationship *string  `json:"recipient_relationship,omitempty"`
	OTPCode               *string  `json:"otp_code,omitempty"`
	SignatureData         *string  `json:"signature_data,omitempty"`
	PhotoURL              *string  `json:"photo_url,omitempty"`
	Notes                 *string  `json:"notes,omitempty"`
	Latitude              *float64 `json:"latitude,omitempty"`
	Longitude             *float64 `json:"longitude,omitempty"`
}

// DeliveryTaskFilter options.
type DeliveryTaskFilter struct {
	Status           *string
	AssignedDriverID *uuid.UUID
	BranchID         *uuid.UUID
	ParcelID         *uuid.UUID
	Priority         *string
	Limit            int
	Offset           int
}

// DeliveryTaskListResponse wrapper.
type DeliveryTaskListResponse struct {
	Tasks  []DeliveryTask `json:"tasks"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// ValidateCreateRequest verifies required fields for delivery task creation.
func ValidateCreateRequest(req CreateDeliveryTaskRequest) error {
	if strings.TrimSpace(req.ParcelID) == "" {
		return errors.New("parcel_id is required")
	}
	if strings.TrimSpace(req.AssignedDriverID) == "" {
		return errors.New("assigned_driver_id is required")
	}
	if req.Priority != nil && strings.TrimSpace(*req.Priority) != "" {
		p := strings.ToUpper(strings.TrimSpace(*req.Priority))
		if p != PriorityLow && p != PriorityStandard && p != PriorityHigh && p != PriorityUrgent {
			return errors.New("priority must be LOW, STANDARD, HIGH, or URGENT")
		}
	}
	return nil
}

// ValidateProofRequest verifies required fields for POD submission.
func ValidateProofRequest(req SubmitDeliveryProofRequest) error {
	pt := strings.ToUpper(strings.TrimSpace(req.ProofType))
	if pt != ProofTypeRecipientSignature && pt != ProofTypeOTP && pt != ProofTypePhoto && pt != ProofTypeContactlessDrop {
		return ErrInvalidProofType
	}
	if strings.TrimSpace(req.RecipientName) == "" {
		return ErrMissingRecipientName
	}
	return nil
}
