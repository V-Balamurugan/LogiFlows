package employees

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
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
	OperationalRoleDriver     = "DRIVER"
	OperationalRoleOperator   = "OPERATOR"
	OperationalRoleDispatcher = "DISPATCHER"
	OperationalRoleSupervisor = "SUPERVISOR"
	OperationalRoleManager    = "MANAGER"
)

// Allowed Employee Statuses
const (
	StatusActive     = "ACTIVE"
	StatusOnLeave    = "ON_LEAVE"
	StatusSuspended  = "SUSPENDED"
	StatusTerminated = "TERMINATED"
)

var (
	ErrInvalidEmployeeCode    = errors.New("employee_code is required and must be between 2 and 50 characters")
	ErrInvalidEmployeeName    = errors.New("first_name and last_name are required")
	ErrInvalidDesignation     = errors.New("designation is required")
	ErrInvalidEmploymentType  = errors.New("invalid employment_type; must be FULL_TIME, PART_TIME, CONTRACTOR, or INTERN")
	ErrInvalidOperationalRole = errors.New("invalid operational_role; must be DRIVER, OPERATOR, DISPATCHER, SUPERVISOR, or MANAGER")
	ErrInvalidStatus          = errors.New("invalid status; must be ACTIVE, ON_LEAVE, SUSPENDED, or TERMINATED")
	ErrEmployeeNotFound       = errors.New("employee not found")
	ErrDuplicateEmployeeCode  = errors.New("employee code already exists in this tenant organization")
	ErrUserAlreadyLinked      = errors.New("this user account is already linked to an employee in this organization")
	ErrBranchCrossTenant      = errors.New("cannot assign employee to a branch belonging to another organization")
	ErrBranchNotFound         = errors.New("assigned branch does not exist or is inactive")
)

type Employee struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	TenantID        uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	UserID          *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	BranchID        *uuid.UUID `json:"branch_id,omitempty" db:"branch_id"`
	EmployeeCode    string     `json:"employee_code" db:"employee_code"`
	FirstName       string     `json:"first_name" db:"first_name"`
	LastName        string     `json:"last_name" db:"last_name"`
	Email           *string    `json:"email,omitempty" db:"email"`
	Phone           *string    `json:"phone,omitempty" db:"phone"`
	Designation     string     `json:"designation" db:"designation"`
	EmploymentType  string     `json:"employment_type" db:"employment_type"`
	OperationalRole string     `json:"operational_role" db:"operational_role"`
	LicenseNumber   *string    `json:"license_number,omitempty" db:"license_number"`
	Status          string     `json:"status" db:"status"`
	IsActive        bool       `json:"is_active" db:"is_active"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`

	// Joined metadata for UI display
	BranchName *string `json:"branch_name,omitempty" db:"branch_name"`
	BranchCode *string `json:"branch_code,omitempty" db:"branch_code"`
}

type CreateEmployeeRequest struct {
	EmployeeCode    string     `json:"employee_code" binding:"required"`
	FirstName       string     `json:"first_name" binding:"required"`
	LastName        string     `json:"last_name" binding:"required"`
	Email           *string    `json:"email,omitempty"`
	Phone           *string    `json:"phone,omitempty"`
	Designation     string     `json:"designation" binding:"required"`
	EmploymentType  string     `json:"employment_type"`
	OperationalRole string     `json:"operational_role"`
	LicenseNumber   *string    `json:"license_number,omitempty"`
	UserID          *uuid.UUID `json:"user_id,omitempty"`
	BranchID        *uuid.UUID `json:"branch_id,omitempty"`
}

func (r *CreateEmployeeRequest) ValidateAndSanitize() error {
	r.EmployeeCode = strings.ToUpper(strings.TrimSpace(r.EmployeeCode))
	if len(r.EmployeeCode) < 2 || len(r.EmployeeCode) > 50 {
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
		case OperationalRoleDriver, OperationalRoleOperator, OperationalRoleDispatcher, OperationalRoleSupervisor, OperationalRoleManager:
		default:
			return ErrInvalidOperationalRole
		}
	}

	return nil
}

type UpdateEmployeeRequest struct {
	FirstName       *string    `json:"first_name,omitempty"`
	LastName        *string    `json:"last_name,omitempty"`
	Email           *string    `json:"email,omitempty"`
	Phone           *string    `json:"phone,omitempty"`
	Designation     *string    `json:"designation,omitempty"`
	EmploymentType  *string    `json:"employment_type,omitempty"`
	OperationalRole *string    `json:"operational_role,omitempty"`
	LicenseNumber   *string    `json:"license_number,omitempty"`
	BranchID        *uuid.UUID `json:"branch_id,omitempty"`
	Status          *string    `json:"status,omitempty"`
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
		case OperationalRoleDriver, OperationalRoleOperator, OperationalRoleDispatcher, OperationalRoleSupervisor, OperationalRoleManager:
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

	return nil
}

type EmployeeFilter struct {
	Search          string
	OperationalRole string
	BranchID        *uuid.UUID
	Status          string
	Limit           int
	Offset          int
}
