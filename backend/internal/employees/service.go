package employees

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/logiflows/logiflows/backend/internal/audit"
	"github.com/logiflows/logiflows/backend/internal/branches"
)

type EmployeeListResponse struct {
	Employees []Employee `json:"employees"`
	Total     int        `json:"total"`
	Limit     int        `json:"limit"`
	Offset    int        `json:"offset"`
}

type Service interface {
	CreateEmployee(ctx context.Context, tenantID, actorID uuid.UUID, req CreateEmployeeRequest, ip, userAgent string) (*Employee, error)
	GetEmployee(ctx context.Context, tenantID, employeeID uuid.UUID) (*Employee, error)
	ListEmployees(ctx context.Context, tenantID uuid.UUID, filter EmployeeFilter) (*EmployeeListResponse, error)
	UpdateEmployee(ctx context.Context, tenantID, employeeID, actorID uuid.UUID, req UpdateEmployeeRequest, ip, userAgent string) (*Employee, error)
	DeactivateEmployee(ctx context.Context, tenantID, employeeID, actorID uuid.UUID, ip, userAgent string) error
}

type employeeService struct {
	repo         Repository
	branchesRepo branches.Repository
	auditRepo    audit.Repository
}

func NewService(repo Repository, branchesRepo branches.Repository, auditRepo audit.Repository) Service {
	return &employeeService{
		repo:         repo,
		branchesRepo: branchesRepo,
		auditRepo:    auditRepo,
	}
}

func (s *employeeService) CreateEmployee(ctx context.Context, tenantID, actorID uuid.UUID, req CreateEmployeeRequest, ip, userAgent string) (*Employee, error) {
	if err := req.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	// 1. Check duplicate employee code within this tenant
	existing, err := s.repo.GetByCode(ctx, tenantID, req.EmployeeCode)
	if err == nil && existing != nil {
		return nil, ErrDuplicateEmployeeCode
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
		TenantID:        tenantID,
		UserID:          req.UserID,
		BranchID:        req.BranchID,
		EmployeeCode:    req.EmployeeCode,
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		Email:           req.Email,
		Phone:           req.Phone,
		Designation:     req.Designation,
		EmploymentType:  req.EmploymentType,
		OperationalRole: req.OperationalRole,
		LicenseNumber:   req.LicenseNumber,
		Status:          StatusActive,
		IsActive:        true,
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
				"employee_code":    emp.EmployeeCode,
				"operational_role": emp.OperationalRole,
				"designation":      emp.Designation,
				"branch_id":        emp.BranchID,
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
		} else {
			existing.IsActive = true
		}
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
				"status":           updated.Status,
				"operational_role": updated.OperationalRole,
				"branch_id":        updated.BranchID,
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
