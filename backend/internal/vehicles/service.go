package vehicles

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/logiflows/logiflows/backend/internal/audit"
	"github.com/logiflows/logiflows/backend/internal/branches"
	"github.com/logiflows/logiflows/backend/internal/employees"
)

type VehicleListResponse struct {
	Vehicles []Vehicle `json:"vehicles"`
	Total    int       `json:"total"`
	Limit    int       `json:"limit"`
	Offset   int       `json:"offset"`
}

type Service interface {
	CreateVehicle(ctx context.Context, tenantID, actorID uuid.UUID, req CreateVehicleRequest, ip, userAgent string) (*Vehicle, error)
	GetVehicle(ctx context.Context, tenantID, vehicleID uuid.UUID) (*Vehicle, error)
	ListVehicles(ctx context.Context, tenantID uuid.UUID, filter VehicleFilter) (*VehicleListResponse, error)
	UpdateVehicle(ctx context.Context, tenantID, vehicleID, actorID uuid.UUID, req UpdateVehicleRequest, ip, userAgent string) (*Vehicle, error)
	DeactivateVehicle(ctx context.Context, tenantID, vehicleID, actorID uuid.UUID, ip, userAgent string) error

	AssignDriver(ctx context.Context, tenantID, vehicleID, actorID uuid.UUID, req AssignVehicleRequest, ip, userAgent string) (*VehicleAssignment, error)
	UnassignDriver(ctx context.Context, tenantID, vehicleID, actorID uuid.UUID, ip, userAgent string) error
	ListAssignments(ctx context.Context, tenantID uuid.UUID, vehicleID, driverID *uuid.UUID, limit, offset int) ([]VehicleAssignment, int, error)
}

type vehicleService struct {
	repo          Repository
	branchesRepo  branches.Repository
	employeesRepo employees.Repository
	auditRepo     audit.Repository
}

func NewService(repo Repository, branchesRepo branches.Repository, employeesRepo employees.Repository, auditRepo audit.Repository) Service {
	return &vehicleService{
		repo:          repo,
		branchesRepo:  branchesRepo,
		employeesRepo: employeesRepo,
		auditRepo:     auditRepo,
	}
}

func (s *vehicleService) CreateVehicle(ctx context.Context, tenantID, actorID uuid.UUID, req CreateVehicleRequest, ip, userAgent string) (*Vehicle, error) {
	if err := req.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	// Check registration number uniqueness in tenant
	existing, err := s.repo.GetVehicleByRegNum(ctx, tenantID, req.RegistrationNumber)
	if err == nil && existing != nil {
		return nil, ErrDuplicateRegistrationNumber
	}

	// Verify Branch association if provided
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

	v := &Vehicle{
		TenantID:           tenantID,
		BranchID:           req.BranchID,
		RegistrationNumber: req.RegistrationNumber,
		VehicleType:        req.VehicleType,
		MakeModel:          req.MakeModel,
		Year:               req.Year,
		MaxWeightKG:        req.MaxWeightKG,
		MaxVolumeCBM:       req.MaxVolumeCBM,
		Status:             VehicleStatusAvailable,
		IsActive:           true,
	}

	if err := s.repo.CreateVehicle(ctx, v); err != nil {
		return nil, err
	}

	// Refetch with joined fields
	populated, err := s.repo.GetVehicleByID(ctx, tenantID, v.ID)
	if err == nil && populated != nil {
		v = populated
	}

	if s.auditRepo != nil {
		vIDStr := v.ID.String()
		_ = s.auditRepo.Log(ctx, &audit.AuditLog{
			TenantID:     &tenantID,
			UserID:       &actorID,
			Action:       "vehicle.created",
			ResourceType: "vehicle",
			ResourceID:   &vIDStr,
			IPAddress:    &ip,
			UserAgent:    &userAgent,
			Status:       audit.StatusSuccess,
			Details: map[string]any{
				"registration_number": v.RegistrationNumber,
				"vehicle_type":        v.VehicleType,
				"branch_id":           v.BranchID,
			},
		})
	}

	return v, nil
}

func (s *vehicleService) GetVehicle(ctx context.Context, tenantID, vehicleID uuid.UUID) (*Vehicle, error) {
	return s.repo.GetVehicleByID(ctx, tenantID, vehicleID)
}

func (s *vehicleService) ListVehicles(ctx context.Context, tenantID uuid.UUID, filter VehicleFilter) (*VehicleListResponse, error) {
	list, total, err := s.repo.ListVehicles(ctx, tenantID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list vehicles: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}

	return &VehicleListResponse{
		Vehicles: list,
		Total:    total,
		Limit:    limit,
		Offset:   filter.Offset,
	}, nil
}

func (s *vehicleService) UpdateVehicle(ctx context.Context, tenantID, vehicleID, actorID uuid.UUID, req UpdateVehicleRequest, ip, userAgent string) (*Vehicle, error) {
	if err := req.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetVehicleByID(ctx, tenantID, vehicleID)
	if err != nil {
		return nil, err
	}

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

	if req.VehicleType != nil {
		existing.VehicleType = *req.VehicleType
	}

	if req.MakeModel != nil {
		mm := strings.TrimSpace(*req.MakeModel)
		existing.MakeModel = &mm
	}

	if req.Year != nil {
		existing.Year = req.Year
	}

	if req.MaxWeightKG != nil {
		existing.MaxWeightKG = *req.MaxWeightKG
	}

	if req.MaxVolumeCBM != nil {
		existing.MaxVolumeCBM = *req.MaxVolumeCBM
	}

	if req.Status != nil {
		existing.Status = *req.Status
		if existing.Status == VehicleStatusDecommissioned {
			existing.IsActive = false
		} else {
			existing.IsActive = true
		}
	}

	if err := s.repo.UpdateVehicle(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update vehicle: %w", err)
	}

	updated, err := s.repo.GetVehicleByID(ctx, tenantID, vehicleID)
	if err != nil {
		return existing, nil
	}

	if s.auditRepo != nil {
		vIDStr := vehicleID.String()
		_ = s.auditRepo.Log(ctx, &audit.AuditLog{
			TenantID:     &tenantID,
			UserID:       &actorID,
			Action:       "vehicle.updated",
			ResourceType: "vehicle",
			ResourceID:   &vIDStr,
			IPAddress:    &ip,
			UserAgent:    &userAgent,
			Status:       audit.StatusSuccess,
			Details: map[string]any{
				"status":       updated.Status,
				"vehicle_type": updated.VehicleType,
				"branch_id":    updated.BranchID,
			},
		})
	}

	return updated, nil
}

func (s *vehicleService) DeactivateVehicle(ctx context.Context, tenantID, vehicleID, actorID uuid.UUID, ip, userAgent string) error {
	existing, err := s.repo.GetVehicleByID(ctx, tenantID, vehicleID)
	if err != nil {
		return err
	}

	// Release any active assignment first
	activeAssignment, err := s.repo.GetActiveAssignmentByVehicle(ctx, tenantID, vehicleID)
	if err == nil && activeAssignment != nil {
		_ = s.repo.CompleteAssignment(ctx, tenantID, activeAssignment.ID)
	}

	if err := s.repo.DeactivateVehicle(ctx, tenantID, vehicleID); err != nil {
		return err
	}

	if s.auditRepo != nil {
		vIDStr := vehicleID.String()
		_ = s.auditRepo.Log(ctx, &audit.AuditLog{
			TenantID:     &tenantID,
			UserID:       &actorID,
			Action:       "vehicle.decommissioned",
			ResourceType: "vehicle",
			ResourceID:   &vIDStr,
			IPAddress:    &ip,
			UserAgent:    &userAgent,
			Status:       audit.StatusSuccess,
			Details: map[string]any{
				"registration_number": existing.RegistrationNumber,
			},
		})
	}

	return nil
}

func (s *vehicleService) AssignDriver(ctx context.Context, tenantID, vehicleID, actorID uuid.UUID, req AssignVehicleRequest, ip, userAgent string) (*VehicleAssignment, error) {
	// 1. Verify vehicle exists and is active
	vehicle, err := s.repo.GetVehicleByID(ctx, tenantID, vehicleID)
	if err != nil {
		return nil, err
	}
	if !vehicle.IsActive || vehicle.Status == VehicleStatusDecommissioned {
		return nil, errors.New("cannot assign an inactive or decommissioned vehicle")
	}

	// 2. Check if vehicle is already assigned
	existingVehicleAssign, err := s.repo.GetActiveAssignmentByVehicle(ctx, tenantID, vehicleID)
	if err == nil && existingVehicleAssign != nil {
		return nil, ErrVehicleAlreadyAssigned
	}

	// 3. Verify driver exists and is active in this tenant
	if s.employeesRepo != nil {
		driver, err := s.employeesRepo.GetByID(ctx, tenantID, req.DriverID)
		if err != nil || driver == nil {
			return nil, ErrDriverNotFound
		}
		if !driver.IsActive {
			return nil, ErrDriverNotFound
		}
	}

	// 4. Check if driver is already assigned to another vehicle
	existingDriverAssign, err := s.repo.GetActiveAssignmentByDriver(ctx, tenantID, req.DriverID)
	if err == nil && existingDriverAssign != nil {
		return nil, ErrDriverAlreadyAssigned
	}

	assignment := &VehicleAssignment{
		TenantID:   tenantID,
		VehicleID:  vehicleID,
		DriverID:   req.DriverID,
		AssignedBy: &actorID,
		Status:     AssignmentStatusActive,
		Notes:      req.Notes,
	}

	if err := s.repo.CreateAssignment(ctx, assignment); err != nil {
		return nil, err
	}

	// 5. Update vehicle status to ASSIGNED
	assignedStatus := VehicleStatusAssigned
	_ = s.repo.UpdateVehicle(ctx, &Vehicle{
		ID:                 vehicle.ID,
		TenantID:           tenantID,
		BranchID:           vehicle.BranchID,
		RegistrationNumber: vehicle.RegistrationNumber,
		VehicleType:        vehicle.VehicleType,
		MakeModel:          vehicle.MakeModel,
		Year:               vehicle.Year,
		MaxWeightKG:        vehicle.MaxWeightKG,
		MaxVolumeCBM:       vehicle.MaxVolumeCBM,
		Status:             assignedStatus,
		IsActive:           true,
	})

	if s.auditRepo != nil {
		assignIDStr := assignment.ID.String()
		_ = s.auditRepo.Log(ctx, &audit.AuditLog{
			TenantID:     &tenantID,
			UserID:       &actorID,
			Action:       "vehicle.assigned",
			ResourceType: "vehicle_assignment",
			ResourceID:   &assignIDStr,
			IPAddress:    &ip,
			UserAgent:    &userAgent,
			Status:       audit.StatusSuccess,
			Details: map[string]any{
				"vehicle_id": vehicleID,
				"driver_id":  req.DriverID,
			},
		})
	}

	return assignment, nil
}

func (s *vehicleService) UnassignDriver(ctx context.Context, tenantID, vehicleID, actorID uuid.UUID, ip, userAgent string) error {
	activeAssignment, err := s.repo.GetActiveAssignmentByVehicle(ctx, tenantID, vehicleID)
	if err != nil {
		return err
	}

	if err := s.repo.CompleteAssignment(ctx, tenantID, activeAssignment.ID); err != nil {
		return err
	}

	// Update vehicle status back to AVAILABLE
	vehicle, err := s.repo.GetVehicleByID(ctx, tenantID, vehicleID)
	if err == nil && vehicle != nil {
		availableStatus := VehicleStatusAvailable
		_ = s.repo.UpdateVehicle(ctx, &Vehicle{
			ID:                 vehicle.ID,
			TenantID:           tenantID,
			BranchID:           vehicle.BranchID,
			RegistrationNumber: vehicle.RegistrationNumber,
			VehicleType:        vehicle.VehicleType,
			MakeModel:          vehicle.MakeModel,
			Year:               vehicle.Year,
			MaxWeightKG:        vehicle.MaxWeightKG,
			MaxVolumeCBM:       vehicle.MaxVolumeCBM,
			Status:             availableStatus,
			IsActive:           true,
		})
	}

	if s.auditRepo != nil {
		assignIDStr := activeAssignment.ID.String()
		_ = s.auditRepo.Log(ctx, &audit.AuditLog{
			TenantID:     &tenantID,
			UserID:       &actorID,
			Action:       "vehicle.unassigned",
			ResourceType: "vehicle_assignment",
			ResourceID:   &assignIDStr,
			IPAddress:    &ip,
			UserAgent:    &userAgent,
			Status:       audit.StatusSuccess,
			Details: map[string]any{
				"vehicle_id": vehicleID,
				"driver_id":  activeAssignment.DriverID,
			},
		})
	}

	return nil
}

func (s *vehicleService) ListAssignments(ctx context.Context, tenantID uuid.UUID, vehicleID, driverID *uuid.UUID, limit, offset int) ([]VehicleAssignment, int, error) {
	return s.repo.ListAssignments(ctx, tenantID, vehicleID, driverID, limit, offset)
}
