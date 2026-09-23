package employees

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/logiflows/logiflows/backend/internal/audit"
	"github.com/logiflows/logiflows/backend/internal/auth"
	"github.com/logiflows/logiflows/backend/internal/branches"
	"github.com/logiflows/logiflows/backend/internal/memberships"
	"github.com/logiflows/logiflows/backend/internal/users"
)

type EmployeeListResponse struct {
	Employees []Employee `json:"employees"`
	Total     int        `json:"total"`
	Limit     int        `json:"limit"`
	Offset    int        `json:"offset"`
}

type Service interface {
	CreateEmployee(ctx context.Context, tenantID, actorID uuid.UUID, req CreateEmployeeRequest, ip, userAgent string) (*Employee, error)
	CreateEmployeeWithAccount(ctx context.Context, tenantID, actorID uuid.UUID, req CreateEmployeeWithAccountRequest, ip, userAgent string) (*Employee, error)
	GetEmployee(ctx context.Context, tenantID, employeeID uuid.UUID) (*Employee, error)
	ListEmployees(ctx context.Context, tenantID uuid.UUID, filter EmployeeFilter) (*EmployeeListResponse, error)
	UpdateEmployee(ctx context.Context, tenantID, employeeID, actorID uuid.UUID, req UpdateEmployeeRequest, ip, userAgent string) (*Employee, error)
	UpdateEmployeeStatus(ctx context.Context, tenantID, employeeID, actorID uuid.UUID, req UpdateEmployeeStatusRequest, ip, userAgent string) (*Employee, error)
	DeactivateEmployee(ctx context.Context, tenantID, employeeID, actorID uuid.UUID, ip, userAgent string) error
	ListAvailableDrivers(ctx context.Context, tenantID uuid.UUID, branchID *uuid.UUID) ([]Employee, error)
	GetAccountStatus(ctx context.Context, tenantID, employeeID uuid.UUID) (*EmployeeAccountStatusResponse, error)
	GetMyProfile(ctx context.Context, tenantID, userID uuid.UUID) (*EmployeeMeResponse, error)
}

type employeeService struct {
	repo            Repository
	branchesRepo    branches.Repository
	auditRepo       audit.Repository
	usersRepo       users.Repository
	membershipsRepo memberships.Repository
}

func NewService(
	repo Repository,
	branchesRepo branches.Repository,
	auditRepo audit.Repository,
	usersRepo users.Repository,
	membershipsRepo memberships.Repository,
) Service {
	return &employeeService{
		repo:            repo,
		branchesRepo:    branchesRepo,
		auditRepo:       auditRepo,
		usersRepo:       usersRepo,
		membershipsRepo: membershipsRepo,
	}
}

func (s *employeeService) CreateEmployee(ctx context.Context, tenantID, actorID uuid.UUID, req CreateEmployeeRequest, ip, userAgent string) (*Employee, error) {
	if err := req.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	// 1. Employee Code assignment or automatic generation
	if req.EmployeeCode == "" {
		generatedCode, err := s.repo.GenerateEmployeeCode(ctx, tenantID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate employee code: %w", err)
		}
		req.EmployeeCode = generatedCode
	} else {
		existing, err := s.repo.GetByCode(ctx, tenantID, req.EmployeeCode)
		if err == nil && existing != nil {
			return nil, ErrDuplicateEmployeeCode
		}
	}

	// 2. Validate Branch Association & Prevent Cross-Tenant Leakage
	if req.BranchID != nil {
		if s.branchesRepo != nil {
			branch, err := s.branchesRepo.GetByID(ctx, tenantID, *req.BranchID)
			if err != nil || branch == nil {
				return nil, ErrBranchNotFound
			}
			if !branch.IsActive {
				return nil, ErrBranchNotFound
			}
		}
	}

	// 3. Validate User Account Association if provided
	if req.UserID != nil {
		existingUserEmp, err := s.repo.GetByUserID(ctx, tenantID, *req.UserID)
		if err == nil && existingUserEmp != nil {
			return nil, ErrUserAlreadyLinked
		}
	}

	emp := &Employee{
		TenantID:           tenantID,
		UserID:             req.UserID,
		BranchID:           req.BranchID,
		EmployeeCode:       req.EmployeeCode,
		FirstName:          req.FirstName,
		LastName:           req.LastName,
		Email:              req.Email,
		Phone:              req.Phone,
		Designation:        req.Designation,
		EmploymentType:     req.EmploymentType,
		OperationalRole:    req.OperationalRole,
		AvailabilityStatus: req.AvailabilityStatus,
		VerificationStatus: req.VerificationStatus,
		LicenseNumber:      req.LicenseNumber,
		Status:             StatusActive,
		IsActive:           true,
	}

	if err := s.repo.Create(ctx, emp); err != nil {
		return nil, err
	}

	// Retrieve populated employee with joined branch details
	populated, err := s.repo.GetByID(ctx, tenantID, emp.ID)
	if err == nil && populated != nil {
		emp = populated
	}

	// 4. Audit Log Emission
	if s.auditRepo != nil {
		empIDStr := emp.ID.String()
		_ = s.auditRepo.Log(ctx, &audit.AuditLog{
			TenantID:     &tenantID,
			UserID:       &actorID,
			Action:       "employee.created",
			ResourceType: "employee",
			ResourceID:   &empIDStr,
			IPAddress:    &ip,
			UserAgent:    &userAgent,
			Status:       audit.StatusSuccess,
			Details: map[string]any{
				"employee_code":       emp.EmployeeCode,
				"operational_role":    emp.OperationalRole,
				"availability_status": emp.AvailabilityStatus,
				"verification_status": emp.VerificationStatus,
				"designation":         emp.Designation,
				"branch_id":           emp.BranchID,
			},
		})
	}

	return emp, nil
}

func (s *employeeService) GetEmployee(ctx context.Context, tenantID, employeeID uuid.UUID) (*Employee, error) {
	return s.repo.GetByID(ctx, tenantID, employeeID)
}

func (s *employeeService) ListEmployees(ctx context.Context, tenantID uuid.UUID, filter EmployeeFilter) (*EmployeeListResponse, error) {
	emps, total, err := s.repo.List(ctx, tenantID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list employees: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}

	return &EmployeeListResponse{
		Employees: emps,
		Total:     total,
		Limit:     limit,
		Offset:    filter.Offset,
	}, nil
}

func (s *employeeService) UpdateEmployee(ctx context.Context, tenantID, employeeID, actorID uuid.UUID, req UpdateEmployeeRequest, ip, userAgent string) (*Employee, error) {
	if err := req.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetByID(ctx, tenantID, employeeID)
	if err != nil {
		return nil, err
	}

	// Validate branch if changing branch assignment
	if req.BranchID != nil {
		if s.branchesRepo != nil {
			branch, err := s.branchesRepo.GetByID(ctx, tenantID, *req.BranchID)
			if err != nil || branch == nil {
				return nil, ErrBranchNotFound
			}
			if !branch.IsActive {
				return nil, ErrBranchNotFound
			}
		}
		existing.BranchID = req.BranchID
	}

	if req.FirstName != nil {
		fn := strings.TrimSpace(*req.FirstName)
		if fn != "" {
			existing.FirstName = fn
		}
	}

	if req.LastName != nil {
		ln := strings.TrimSpace(*req.LastName)
		if ln != "" {
			existing.LastName = ln
		}
	}

	if req.Email != nil {
		existing.Email = req.Email
	}

	if req.Phone != nil {
		existing.Phone = req.Phone
	}

	if req.Designation != nil {
		des := strings.TrimSpace(*req.Designation)
		if des != "" {
			existing.Designation = des
		}
	}

	if req.EmploymentType != nil {
		existing.EmploymentType = *req.EmploymentType
	}

	if req.OperationalRole != nil {
		existing.OperationalRole = *req.OperationalRole
	}

	if req.LicenseNumber != nil {
		existing.LicenseNumber = req.LicenseNumber
	}

	if req.Status != nil {
		existing.Status = *req.Status
		if existing.Status == StatusTerminated {
			existing.IsActive = false
			existing.AvailabilityStatus = AvailabilityStatusUnavailable
		} else {
			existing.IsActive = true
		}
	}

	if req.AvailabilityStatus != nil {
		existing.AvailabilityStatus = *req.AvailabilityStatus
	}

	if req.VerificationStatus != nil {
		existing.VerificationStatus = *req.VerificationStatus
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update employee: %w", err)
	}

	// Refetch with joined fields
	updated, err := s.repo.GetByID(ctx, tenantID, employeeID)
	if err != nil {
		return existing, nil
	}

	if s.auditRepo != nil {
		empIDStr := employeeID.String()
		_ = s.auditRepo.Log(ctx, &audit.AuditLog{
			TenantID:     &tenantID,
			UserID:       &actorID,
			Action:       "employee.updated",
			ResourceType: "employee",
			ResourceID:   &empIDStr,
			IPAddress:    &ip,
			UserAgent:    &userAgent,
			Status:       audit.StatusSuccess,
			Details: map[string]any{
				"status":              updated.Status,
				"availability_status": updated.AvailabilityStatus,
				"verification_status": updated.VerificationStatus,
				"operational_role":    updated.OperationalRole,
				"branch_id":           updated.BranchID,
			},
		})
	}

	return updated, nil
}

func (s *employeeService) UpdateEmployeeStatus(ctx context.Context, tenantID, employeeID, actorID uuid.UUID, req UpdateEmployeeStatusRequest, ip, userAgent string) (*Employee, error) {
	if err := req.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateStatus(ctx, tenantID, employeeID, req); err != nil {
		return nil, err
	}

	updated, err := s.repo.GetByID(ctx, tenantID, employeeID)
	if err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		empIDStr := employeeID.String()
		_ = s.auditRepo.Log(ctx, &audit.AuditLog{
			TenantID:     &tenantID,
			UserID:       &actorID,
			Action:       "employee.status_updated",
			ResourceType: "employee",
			ResourceID:   &empIDStr,
			IPAddress:    &ip,
			UserAgent:    &userAgent,
			Status:       audit.StatusSuccess,
			Details: map[string]any{
				"status":              updated.Status,
				"availability_status": updated.AvailabilityStatus,
				"verification_status": updated.VerificationStatus,
			},
		})
	}

	return updated, nil
}

func (s *employeeService) DeactivateEmployee(ctx context.Context, tenantID, employeeID, actorID uuid.UUID, ip, userAgent string) error {
	existing, err := s.repo.GetByID(ctx, tenantID, employeeID)
	if err != nil {
		return err
	}

	if err := s.repo.Deactivate(ctx, tenantID, employeeID); err != nil {
		return err
	}

	if s.auditRepo != nil {
		empIDStr := employeeID.String()
		_ = s.auditRepo.Log(ctx, &audit.AuditLog{
			TenantID:     &tenantID,
			UserID:       &actorID,
			Action:       "employee.deactivated",
			ResourceType: "employee",
			ResourceID:   &empIDStr,
			IPAddress:    &ip,
			UserAgent:    &userAgent,
			Status:       audit.StatusSuccess,
			Details: map[string]any{
				"employee_code": existing.EmployeeCode,
			},
		})
	}

	return nil
}

func (s *employeeService) ListAvailableDrivers(ctx context.Context, tenantID uuid.UUID, branchID *uuid.UUID) ([]Employee, error) {
	return s.repo.ListAvailableDrivers(ctx, tenantID, branchID)
}

func (s *employeeService) CreateEmployeeWithAccount(ctx context.Context, tenantID, actorID uuid.UUID, req CreateEmployeeWithAccountRequest, ip, userAgent string) (*Employee, error) {
	if err := req.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	// 1. Password generation / validation
	var plainPassword string
	if req.Password != "" {
		plainPassword = req.Password
		if err := auth.ValidatePassword(plainPassword); err != nil {
			return nil, err
		}
	} else if req.SendInvite {
		rawBytes := make([]byte, 8)
		if _, err := rand.Read(rawBytes); err != nil {
			return nil, fmt.Errorf("failed to generate secure invitation password: %w", err)
		}
		plainPassword = fmt.Sprintf("Logi#%s9A!", base64.RawURLEncoding.EncodeToString(rawBytes))
	} else {
		return nil, ErrAccountPasswordRequired
	}

	passwordHash, err := auth.HashPassword(plainPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 2. Validate Branch Association
	if req.BranchID != nil && s.branchesRepo != nil {
		branch, err := s.branchesRepo.GetByID(ctx, tenantID, *req.BranchID)
		if err != nil || branch == nil || !branch.IsActive {
			return nil, ErrBranchNotFound
		}
	}

	// 3. Check existing user by email
	var (
		targetUserID uuid.UUID
		isNewUser    bool
	)
	if s.usersRepo != nil {
		existingUser, err := s.usersRepo.GetByEmail(ctx, req.Email)
		if err == nil && existingUser != nil {
			targetUserID = existingUser.ID
			if s.membershipsRepo != nil {
				existingMem, _ := s.membershipsRepo.GetByUserAndTenant(ctx, targetUserID, tenantID)
				if existingMem != nil {
					return nil, ErrUserAlreadyLinked
				}
			}
			existingEmp, _ := s.repo.GetByUserID(ctx, tenantID, targetUserID)
			if existingEmp != nil {
				return nil, ErrUserAlreadyLinked
			}
		} else {
			isNewUser = true
		}
	}

	// 4. Begin Database Transaction
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start database transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 5. If new user, create user record in tx
	fullName := strings.TrimSpace(req.FirstName + " " + req.LastName)
	if isNewUser && s.usersRepo != nil {
		newUser := &users.User{
			ID:              uuid.New(),
			Email:           req.Email,
			PasswordHash:    passwordHash,
			FullName:        fullName,
			PhoneNumber:     req.Phone,
			IsActive:        true,
			EmailVerified:   true,
			IsPlatformAdmin: false,
		}
		if err := s.usersRepo.CreateTx(ctx, tx, newUser); err != nil {
			return nil, fmt.Errorf("failed to create user account: %w", err)
		}
		targetUserID = newUser.ID
	}

	// 6. Create Tenant Membership in tx
	if s.membershipsRepo != nil {
		membership := &memberships.TenantMembership{
			ID:       uuid.New(),
			TenantID: tenantID,
			UserID:   targetUserID,
			Role:     req.SystemRole,
			Status:   memberships.StatusActive,
		}
		if err := s.membershipsRepo.CreateTx(ctx, tx, membership); err != nil {
			return nil, fmt.Errorf("failed to create tenant membership: %w", err)
		}
	}

	// 7. Generate Employee Code in tx
	code := req.EmployeeCode
	if code == "" {
		generated, err := s.repo.GenerateEmployeeCodeTx(ctx, tx, tenantID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate employee code: %w", err)
		}
		code = generated
	}

	// 8. Create Employee Profile in tx
	emp := &Employee{
		TenantID:           tenantID,
		UserID:             &targetUserID,
		BranchID:           req.BranchID,
		EmployeeCode:       code,
		FirstName:          req.FirstName,
		LastName:           req.LastName,
		Email:              &req.Email,
		Phone:              req.Phone,
		Designation:        req.Designation,
		EmploymentType:     req.EmploymentType,
		OperationalRole:    req.OperationalRole,
		LicenseNumber:      req.LicenseNumber,
		Status:             StatusActive,
		AvailabilityStatus: req.AvailabilityStatus,
		VerificationStatus: req.VerificationStatus,
		IsActive:           true,
	}
	if err := s.repo.CreateTx(ctx, tx, emp); err != nil {
		return nil, err
	}

	// 9. Commit Transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 10. Audit Logging
	if s.auditRepo != nil {
		empIDStr := emp.ID.String()
		_ = s.auditRepo.Log(ctx, &audit.AuditLog{
			TenantID:     &tenantID,
			UserID:       &actorID,
			Action:       "employee.created_with_account",
			ResourceType: "employee",
			ResourceID:   &empIDStr,
			IPAddress:    &ip,
			UserAgent:    &userAgent,
			Status:       audit.StatusSuccess,
			Details: map[string]any{
				"employee_code":    emp.EmployeeCode,
				"user_id":          targetUserID.String(),
				"operational_role": emp.OperationalRole,
				"system_role":      req.SystemRole,
				"branch_id":        emp.BranchID,
			},
		})
	}

	// Retrieve populated employee with joined branch details
	populated, err := s.repo.GetByID(ctx, tenantID, emp.ID)
	if err == nil && populated != nil {
		emp = populated
	}

	return emp, nil
}

func (s *employeeService) GetAccountStatus(ctx context.Context, tenantID, employeeID uuid.UUID) (*EmployeeAccountStatusResponse, error) {
	return s.repo.GetAccountStatus(ctx, tenantID, employeeID)
}

func (s *employeeService) GetMyProfile(ctx context.Context, tenantID, userID uuid.UUID) (*EmployeeMeResponse, error) {
	emp, err := s.repo.GetByUserID(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}

	var assignedVehicle *AssignedVehicleInfo
	if emp.OperationalRole == OperationalRoleDriver {
		veh, err := s.repo.GetActiveVehicleByDriverID(ctx, tenantID, emp.ID)
		if err == nil && veh != nil {
			assignedVehicle = veh
		}
	}

	systemRole := memberships.RoleEmployee
	if s.membershipsRepo != nil {
		if mem, err := s.membershipsRepo.GetByUserAndTenant(ctx, userID, tenantID); err == nil && mem != nil {
			systemRole = mem.Role
		}
	}

	return &EmployeeMeResponse{
		Employee:        *emp,
		SystemRole:      systemRole,
		AssignedVehicle: assignedVehicle,
	}, nil
}
