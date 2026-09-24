package customers

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Customer status constants.
const (
	StatusActive    = "ACTIVE"
	StatusInactive  = "INACTIVE"
	StatusSuspended = "SUSPENDED"
)

// Customer type constants.
const (
	TypeIndividual = "INDIVIDUAL"
	TypeBusiness   = "BUSINESS"
	TypeEnterprise = "ENTERPRISE"
	TypeMerchant   = "MERCHANT"
)

// Customer contract tier constants.
const (
	TierStandard = "STANDARD"
	TierSilver   = "SILVER"
	TierGold     = "GOLD"
	TierVIP      = "VIP"
)

// Domain errors.
var (
	ErrCustomerNotFound       = errors.New("customer not found in this organization")
	ErrDuplicateCustomerCode  = errors.New("customer with this code already exists in this organization")
	ErrInvalidCustomerName    = errors.New("customer name is required and cannot be empty")
	ErrInvalidCustomerPhone   = errors.New("valid customer phone number is required")
	ErrInvalidCustomerAddress = errors.New("address line 1, city, state, and postal code are required")
	ErrInvalidCustomerType    = errors.New("invalid customer type, must be INDIVIDUAL, BUSINESS, ENTERPRISE, or MERCHANT")
	ErrInvalidCustomerStatus  = errors.New("invalid customer status, must be ACTIVE, INACTIVE, or SUSPENDED")
)

// Customer domain model.
type Customer struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	TenantID     uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	CustomerCode string     `json:"customer_code" db:"customer_code"`
	Name         string     `json:"name" db:"name"`
	CompanyName  *string    `json:"company_name,omitempty" db:"company_name"`
	CustomerType string     `json:"customer_type" db:"customer_type"`
	Email        *string    `json:"email,omitempty" db:"email"`
	Phone        string     `json:"phone" db:"phone"`
	AddressLine1 string     `json:"address_line1" db:"address_line1"`
	AddressLine2 *string    `json:"address_line2,omitempty" db:"address_line2"`
	City         string     `json:"city" db:"city"`
	State        string     `json:"state" db:"state"`
	PostalCode   string     `json:"postal_code" db:"postal_code"`
	Country      string     `json:"country" db:"country"`
	Status       string     `json:"status" db:"status"`
	CreditLimit  float64    `json:"credit_limit" db:"credit_limit"`
	ContractTier string     `json:"contract_tier" db:"contract_tier"`
	Notes        *string    `json:"notes,omitempty" db:"notes"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// CreateCustomerRequest payload.
type CreateCustomerRequest struct {
	CustomerCode string   `json:"customer_code,omitempty"`
	Name         string   `json:"name"`
	CompanyName  *string  `json:"company_name,omitempty"`
	CustomerType string   `json:"customer_type,omitempty"`
	Email        *string  `json:"email,omitempty"`
	Phone        string   `json:"phone"`
	AddressLine1 string   `json:"address_line1"`
	AddressLine2 *string  `json:"address_line2,omitempty"`
	City         string   `json:"city"`
	State        string   `json:"state"`
	PostalCode   string   `json:"postal_code"`
	Country      string   `json:"country,omitempty"`
	CreditLimit  *float64 `json:"credit_limit,omitempty"`
	ContractTier string   `json:"contract_tier,omitempty"`
	Notes        *string  `json:"notes,omitempty"`
}

// UpdateCustomerRequest payload.
type UpdateCustomerRequest struct {
	Name         *string  `json:"name,omitempty"`
	CompanyName  *string  `json:"company_name,omitempty"`
	CustomerType *string  `json:"customer_type,omitempty"`
	Email        *string  `json:"email,omitempty"`
	Phone        *string  `json:"phone,omitempty"`
	AddressLine1 *string  `json:"address_line1,omitempty"`
	AddressLine2 *string  `json:"address_line2,omitempty"`
	City         *string  `json:"city,omitempty"`
	State        *string  `json:"state,omitempty"`
	PostalCode   *string  `json:"postal_code,omitempty"`
	Country      *string  `json:"country,omitempty"`
	Status       *string  `json:"status,omitempty"`
	CreditLimit  *float64 `json:"credit_limit,omitempty"`
	ContractTier *string  `json:"contract_tier,omitempty"`
	Notes        *string  `json:"notes,omitempty"`
}

// CustomerFilter query parameters for listing.
type CustomerFilter struct {
	Search       *string
	Status       *string
	CustomerType *string
	Limit        int
	Offset       int
}

// CustomerListResponse envelope.
type CustomerListResponse struct {
	Customers []Customer `json:"customers"`
	Total     int        `json:"total"`
	Limit     int        `json:"limit"`
	Offset    int        `json:"offset"`
}

// ValidateCreateRequest verifies required fields and invariants for customer creation.
func ValidateCreateRequest(req CreateCustomerRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return ErrInvalidCustomerName
	}
	if strings.TrimSpace(req.Phone) == "" {
		return ErrInvalidCustomerPhone
	}
	if strings.TrimSpace(req.AddressLine1) == "" ||
		strings.TrimSpace(req.City) == "" ||
		strings.TrimSpace(req.State) == "" ||
		strings.TrimSpace(req.PostalCode) == "" {
		return ErrInvalidCustomerAddress
	}

	cType := strings.ToUpper(strings.TrimSpace(req.CustomerType))
	if cType != "" {
		validTypes := map[string]bool{
			TypeIndividual: true,
			TypeBusiness:   true,
			TypeEnterprise: true,
			TypeMerchant:   true,
		}
		if !validTypes[cType] {
			return ErrInvalidCustomerType
		}
	}

	return nil
}
