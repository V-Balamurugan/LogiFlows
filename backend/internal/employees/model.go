package employees

import (
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/logiflows/logiflows/backend/internal/memberships"
)

// Allowed Employment Types
const (
	EmploymentTypeFullTime   = "FULL_TIME"
	EmploymentTypePartTime   = "PART_TIME"
	EmploymentTypeContractor = "CONTRACTOR"
	EmploymentTypeIntern     = "INTERN"
)

// Allowed Operational Roles
const (
	OperationalRoleDriver            = "DRIVER"
	OperationalRoleOperator          = "OPERATOR"
	OperationalRoleDispatcher        = "DISPATCHER"
	OperationalRoleSupervisor        = "SUPERVISOR"
	OperationalRoleManager           = "MANAGER"
	OperationalRoleBranchManager     = "BRANCH_MANAGER"
	OperationalRoleWarehouseOperator = "WAREHOUSE_OPERATOR"
	OperationalRoleDeliveryExecutive = "DELIVERY_EXECUTIVE"
)

// Allowed Employee Statuses
const (
	StatusActive     = "ACTIVE"
	StatusOnLeave    = "ON_LEAVE"
	StatusSuspended  = "SUSPENDED"
	StatusTerminated = "TERMINATED"
)

// Allowed Employee Availability Statuses
const (
	AvailabilityStatusAvailable   = "AVAILABLE"
	AvailabilityStatusBusy        = "BUSY"
	AvailabilityStatusOffDuty     = "OFF_DUTY"
	AvailabilityStatusUnavailable = "UNAVAILABLE"
)

// Allowed Employee Verification Statuses
const (
	VerificationStatusPending  = "PENDING"
	VerificationStatusVerified = "VERIFIED"
	VerificationStatusRejected = "REJECTED"
)

var (
	ErrInvalidEmployeeCode       = errors.New("employee_code must be between 2 and 50 characters")
	ErrInvalidEmployeeName       = errors.New("first_name and last_name are required")
	ErrInvalidDesignation        = errors.New("designation is required")
	ErrInvalidEmploymentType     = errors.New("invalid employment_type; must be FULL_TIME, PART_TIME, CONTRACTOR, or INTERN")
	ErrInvalidOperationalRole    = errors.New("invalid operational_role; must be DRIVER, OPERATOR, DISPATCHER, SUPERVISOR, MANAGER, BRANCH_MANAGER, WAREHOUSE_OPERATOR, or DELIVERY_EXECUTIVE")
	ErrInvalidStatus             = errors.New("invalid status; must be ACTIVE, ON_LEAVE, SUSPENDED, or TERMINATED")
	ErrInvalidAvailabilityStatus = errors.New("invalid availability_status; must be AVAILABLE, BUSY, OFF_DUTY, or UNAVAILABLE")
	ErrInvalidVerificationStatus = errors.New("invalid verification_status; must be PENDING, VERIFIED, or REJECTED")
	ErrEmployeeNotFound          = errors.New("employee not found")
	ErrDuplicateEmployeeCode     = errors.New("employee code already exists in this tenant organization")
	ErrUserAlreadyLinked         = errors.New("this user account is already linked to an employee in this organization")
	ErrBranchCrossTenant         = errors.New("cannot assign employee to a branch belonging to another organization")
	ErrBranchNotFound            = errors.New("assigned branch does not exist or is inactive")
	ErrDriverNotEligible         = errors.New("employee does not have an active DRIVER operational role or is not available")
	ErrEmailRequired             = errors.New("email is required for employee account creation")
	ErrInvalidEmail              = errors.New("invalid email address format")
	ErrAccountPasswordRequired   = errors.New("password is required unless invitation flow is enabled")
	ErrInvalidSystemRole         = errors.New("invalid system role for employee account; must be EMPLOYEE, TENANT_OPERATOR, or VIEWER")
)

type Employee struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	TenantID           uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	UserID             *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	BranchID           *uuid.UUID `json:"branch_id,omitempty" db:"branch_id"`
	EmployeeCode       string     `json:"employee_code" db:"employee_code"`
	FirstName          string     `json:"first_name" db:"first_name"`
	LastName           string     `json:"last_name" db:"last_name"`
	Email              *string    `json:"email,omitempty" db:"email"`
	Phone              *string    `json:"phone,omitempty" db:"phone"`
	Designation        string     `json:"designation" db:"designation"`
	EmploymentType     string     `json:"employment_type" db:"employment_type"`
	OperationalRole    string     `json:"operational_role" db:"operational_role"`
	LicenseNumber      *string    `json:"license_number,omitempty" db:"license_number"`
	Status             string     `json:"status" db:"status"`
	AvailabilityStatus string     `json:"availability_status" db:"availability_status"`
	VerificationStatus string     `json:"verification_status" db:"verification_status"`
	JoiningDate        *time.Time `json:"joining_date,omitempty" db:"joining_date"`
	IsActive           bool       `json:"is_active" db:"is_active"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`

	// Joined metadata for UI display
	BranchName *string `json:"branch_name,omitempty" db:"branch_name"`
	BranchCode *string `json:"branch_code,omitempty" db:"branch_code"`
}

type CreateEmployeeRequest struct {
	EmployeeCode       string     `json:"employee_code,omitempty"` // Optional: automatically generated if empty
	FirstName          string     `json:"first_name" binding:"required"`
	LastName           string     `json:"last_name" binding:"required"`
	Email              *string    `json:"email,omitempty"`
	Phone              *string    `json:"phone,omitempty"`
	Designation        string     `json:"designation" binding:"required"`
	EmploymentType     string     `json:"employment_type,omitempty"`
	OperationalRole    string     `json:"operational_role,omitempty"`
	AvailabilityStatus string     `json:"availability_status,omitempty"`
	VerificationStatus string     `json:"verification_status,omitempty"`
	LicenseNumber      *string    `json:"license_number,omitempty"`
	UserID             *uuid.UUID `json:"user_id,omitempty"`
	BranchID           *uuid.UUID `json:"branch_id,omitempty"`
}

func (r *CreateEmployeeRequest) ValidateAndSanitize() error {
	r.EmployeeCode = strings.ToUpper(strings.TrimSpace(r.EmployeeCode))
	if r.EmployeeCode != "" && (len(r.EmployeeCode) < 2 || len(r.EmployeeCode) > 50) {
		return ErrInvalidEmployeeCode
	}

	r.FirstName = strings.TrimSpace(r.FirstName)
	r.LastName = strings.TrimSpace(r.LastName)
	if r.FirstName == "" || r.LastName == "" {
		return ErrInvalidEmployeeName
	}

	r.Designation = strings.TrimSpace(r.Designation)
	if r.Designation == "" {
		return ErrInvalidDesignation
	}

	if r.EmploymentType == "" {
		r.EmploymentType = EmploymentTypeFullTime
	} else {
		r.EmploymentType = strings.ToUpper(strings.TrimSpace(r.EmploymentType))
		switch r.EmploymentType {
		case EmploymentTypeFullTime, EmploymentTypePartTime, EmploymentTypeContractor, EmploymentTypeIntern:
		default:
			return ErrInvalidEmploymentType
		}
	}

	if r.OperationalRole == "" {
		r.OperationalRole = OperationalRoleDriver
	} else {
		r.OperationalRole = strings.ToUpper(strings.TrimSpace(r.OperationalRole))
		switch r.OperationalRole {
		case OperationalRoleDriver, OperationalRoleOperator, OperationalRoleDispatcher,
			OperationalRoleSupervisor, OperationalRoleManager, OperationalRoleBranchManager,
			OperationalRoleWarehouseOperator, OperationalRoleDeliveryExecutive:
		default:
			return ErrInvalidOperationalRole
		}
	}

	if r.AvailabilityStatus == "" {
		r.AvailabilityStatus = AvailabilityStatusAvailable
	} else {
		r.AvailabilityStatus = strings.ToUpper(strings.TrimSpace(r.AvailabilityStatus))
		switch r.AvailabilityStatus {
		case AvailabilityStatusAvailable, AvailabilityStatusBusy, AvailabilityStatusOffDuty, AvailabilityStatusUnavailable:
		default:
			return ErrInvalidAvailabilityStatus
		}
	}

	if r.VerificationStatus == "" {
		r.VerificationStatus = VerificationStatusVerified
	} else {
		r.VerificationStatus = strings.ToUpper(strings.TrimSpace(r.VerificationStatus))
		switch r.VerificationStatus {
		case VerificationStatusPending, VerificationStatusVerified, VerificationStatusRejected:
		default:
			return ErrInvalidVerificationStatus
		}
	}

	return nil
}

type UpdateEmployeeRequest struct {
	FirstName          *string    `json:"first_name,omitempty"`
	LastName           *string    `json:"last_name,omitempty"`
	Email              *string    `json:"email,omitempty"`
	Phone              *string    `json:"phone,omitempty"`
	Designation        *string    `json:"designation,omitempty"`
	EmploymentType     *string    `json:"employment_type,omitempty"`
	OperationalRole    *string    `json:"operational_role,omitempty"`
	AvailabilityStatus *string    `json:"availability_status,omitempty"`
	VerificationStatus *string    `json:"verification_status,omitempty"`
	LicenseNumber      *string    `json:"license_number,omitempty"`
	BranchID           *uuid.UUID `json:"branch_id,omitempty"`
	Status             *string    `json:"status,omitempty"`
}

func (r *UpdateEmployeeRequest) ValidateAndSanitize() error {
	if r.EmploymentType != nil {
		et := strings.ToUpper(strings.TrimSpace(*r.EmploymentType))
		switch et {
		case EmploymentTypeFullTime, EmploymentTypePartTime, EmploymentTypeContractor, EmploymentTypeIntern:
			*r.EmploymentType = et
		default:
			return ErrInvalidEmploymentType
		}
	}

	if r.OperationalRole != nil {
		op := strings.ToUpper(strings.TrimSpace(*r.OperationalRole))
		switch op {
		case OperationalRoleDriver, OperationalRoleOperator, OperationalRoleDispatcher,
			OperationalRoleSupervisor, OperationalRoleManager, OperationalRoleBranchManager,
			OperationalRoleWarehouseOperator, OperationalRoleDeliveryExecutive:
			*r.OperationalRole = op
		default:
			return ErrInvalidOperationalRole
		}
	}

	if r.Status != nil {
		st := strings.ToUpper(strings.TrimSpace(*r.Status))
		switch st {
		case StatusActive, StatusOnLeave, StatusSuspended, StatusTerminated:
			*r.Status = st
		default:
			return ErrInvalidStatus
		}
	}

	if r.AvailabilityStatus != nil {
		av := strings.ToUpper(strings.TrimSpace(*r.AvailabilityStatus))
		switch av {
		case AvailabilityStatusAvailable, AvailabilityStatusBusy, AvailabilityStatusOffDuty, AvailabilityStatusUnavailable:
			*r.AvailabilityStatus = av
		default:
			return ErrInvalidAvailabilityStatus
		}
	}

	if r.VerificationStatus != nil {
		vf := strings.ToUpper(strings.TrimSpace(*r.VerificationStatus))
		switch vf {
		case VerificationStatusPending, VerificationStatusVerified, VerificationStatusRejected:
			*r.VerificationStatus = vf
		default:
			return ErrInvalidVerificationStatus
		}
	}

	return nil
}

type UpdateEmployeeStatusRequest struct {
	Status             *string `json:"status,omitempty"`
	AvailabilityStatus *string `json:"availability_status,omitempty"`
	VerificationStatus *string `json:"verification_status,omitempty"`
}

func (r *UpdateEmployeeStatusRequest) ValidateAndSanitize() error {
	if r.Status != nil {
		st := strings.ToUpper(strings.TrimSpace(*r.Status))
		switch st {
		case StatusActive, StatusOnLeave, StatusSuspended, StatusTerminated:
			*r.Status = st
		default:
			return ErrInvalidStatus
		}
	}

	if r.AvailabilityStatus != nil {
		av := strings.ToUpper(strings.TrimSpace(*r.AvailabilityStatus))
		switch av {
		case AvailabilityStatusAvailable, AvailabilityStatusBusy, AvailabilityStatusOffDuty, AvailabilityStatusUnavailable:
			*r.AvailabilityStatus = av
		default:
			return ErrInvalidAvailabilityStatus
		}
	}

	if r.VerificationStatus != nil {
		vf := strings.ToUpper(strings.TrimSpace(*r.VerificationStatus))
		switch vf {
		case VerificationStatusPending, VerificationStatusVerified, VerificationStatusRejected:
			*r.VerificationStatus = vf
		default:
			return ErrInvalidVerificationStatus
		}
	}

	if r.Status == nil && r.AvailabilityStatus == nil && r.VerificationStatus == nil {
		return errors.New("at least one of status, availability_status, or verification_status must be provided")
	}

	return nil
}

type EmployeeFilter struct {
	Search             string
	OperationalRole    string
	BranchID           *uuid.UUID
	Status             string
	AvailabilityStatus string
	VerificationStatus string
	Limit              int
	Offset             int
}

type CreateEmployeeWithAccountRequest struct {
	EmployeeCode       string     `json:"employee_code,omitempty"`
	FirstName          string     `json:"first_name" binding:"required"`
	LastName           string     `json:"last_name" binding:"required"`
	Email              string     `json:"email" binding:"required"`
	Phone              *string    `json:"phone,omitempty"`
	Designation        string     `json:"designation" binding:"required"`
	EmploymentType     string     `json:"employment_type,omitempty"`
	OperationalRole    string     `json:"operational_role,omitempty"`
	AvailabilityStatus string     `json:"availability_status,omitempty"`
	VerificationStatus string     `json:"verification_status,omitempty"`
	LicenseNumber      *string    `json:"license_number,omitempty"`
	BranchID           *uuid.UUID `json:"branch_id,omitempty"`

	// Account creation details
	Password   string `json:"password,omitempty"`
	SystemRole string `json:"system_role,omitempty"` // Default: "EMPLOYEE", can be "TENANT_OPERATOR" or "VIEWER"
	SendInvite bool   `json:"send_invite,omitempty"`
}

func (r *CreateEmployeeWithAccountRequest) ValidateAndSanitize() error {
	r.Email = strings.ToLower(strings.TrimSpace(r.Email))
	if r.Email == "" {
		return ErrEmailRequired
	}
	if _, err := mail.ParseAddress(r.Email); err != nil {
		return ErrInvalidEmail
	}

	r.FirstName = strings.TrimSpace(r.FirstName)
	r.LastName = strings.TrimSpace(r.LastName)
	if r.FirstName == "" || r.LastName == "" {
		return ErrInvalidEmployeeName
	}

	r.Designation = strings.TrimSpace(r.Designation)
	if r.Designation == "" {
		return ErrInvalidDesignation
	}

	r.EmployeeCode = strings.ToUpper(strings.TrimSpace(r.EmployeeCode))
	if r.EmployeeCode != "" && (len(r.EmployeeCode) < 2 || len(r.EmployeeCode) > 50) {
		return ErrInvalidEmployeeCode
	}

	if r.EmploymentType == "" {
		r.EmploymentType = EmploymentTypeFullTime
	} else {
		r.EmploymentType = strings.ToUpper(strings.TrimSpace(r.EmploymentType))
		switch r.EmploymentType {
		case EmploymentTypeFullTime, EmploymentTypePartTime, EmploymentTypeContractor, EmploymentTypeIntern:
		default:
			return ErrInvalidEmploymentType
		}
	}

	if r.OperationalRole == "" {
		r.OperationalRole = OperationalRoleDriver
	} else {
		r.OperationalRole = strings.ToUpper(strings.TrimSpace(r.OperationalRole))
		switch r.OperationalRole {
		case OperationalRoleDriver, OperationalRoleOperator, OperationalRoleDispatcher,
			OperationalRoleSupervisor, OperationalRoleManager, OperationalRoleBranchManager,
			OperationalRoleWarehouseOperator, OperationalRoleDeliveryExecutive:
		default:
			return ErrInvalidOperationalRole
		}
	}

	if r.AvailabilityStatus == "" {
		r.AvailabilityStatus = AvailabilityStatusAvailable
	} else {
		r.AvailabilityStatus = strings.ToUpper(strings.TrimSpace(r.AvailabilityStatus))
		switch r.AvailabilityStatus {
		case AvailabilityStatusAvailable, AvailabilityStatusBusy, AvailabilityStatusOffDuty, AvailabilityStatusUnavailable:
		default:
			return ErrInvalidAvailabilityStatus
		}
	}

	if r.VerificationStatus == "" {
		r.VerificationStatus = VerificationStatusVerified
	} else {
		r.VerificationStatus = strings.ToUpper(strings.TrimSpace(r.VerificationStatus))
		switch r.VerificationStatus {
		case VerificationStatusPending, VerificationStatusVerified, VerificationStatusRejected:
		default:
			return ErrInvalidVerificationStatus
		}
	}

	if r.SystemRole == "" {
		r.SystemRole = memberships.RoleEmployee
	} else {
		r.SystemRole = strings.ToUpper(strings.TrimSpace(r.SystemRole))
		switch r.SystemRole {
		case memberships.RoleEmployee, memberships.RoleTenantOperator, memberships.RoleViewer:
		default:
			return ErrInvalidSystemRole
		}
	}

	if r.Password == "" && !r.SendInvite {
		return ErrAccountPasswordRequired
	}

	return nil
}

type EmployeeAccountStatusResponse struct {
	EmployeeID      uuid.UUID  `json:"employee_id"`
	EmployeeCode    string     `json:"employee_code"`
	FullName        string     `json:"full_name"`
	OperationalRole string     `json:"operational_role"`
	Status          string     `json:"status"`
	HasAccount      bool       `json:"has_account"`
	UserID          *uuid.UUID `json:"user_id,omitempty"`
	UserEmail       *string    `json:"user_email,omitempty"`
	UserIsActive    *bool      `json:"user_is_active,omitempty"`
	SystemRole      *string    `json:"system_role,omitempty"`
	BranchID        *uuid.UUID `json:"branch_id,omitempty"`
	BranchName      *string    `json:"branch_name,omitempty"`
}

type AssignedVehicleInfo struct {
	ID                 uuid.UUID `json:"id"`
	RegistrationNumber string    `json:"registration_number"`
	VehicleType        string    `json:"vehicle_type"`
	MakeModel          *string   `json:"make_model,omitempty"`
	Status             string    `json:"status"`
}

type EmployeeMeResponse struct {
	Employee        Employee             `json:"employee"`
	SystemRole      string               `json:"system_role"`
	AssignedVehicle *AssignedVehicleInfo `json:"assigned_vehicle,omitempty"`
}
