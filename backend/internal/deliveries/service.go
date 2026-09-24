package deliveries

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/logiflows/logiflows/backend/internal/employees"
	"github.com/logiflows/logiflows/backend/internal/parcels"
)

type Service interface {
	CreateDeliveryTask(ctx context.Context, tenantID uuid.UUID, req CreateDeliveryTaskRequest) (*DeliveryTask, error)
	GetDeliveryTask(ctx context.Context, tenantID, taskID uuid.UUID) (*DeliveryTask, error)
	ListDeliveryTasks(ctx context.Context, tenantID uuid.UUID, filter DeliveryTaskFilter) (*DeliveryTaskListResponse, error)
	GetDriverTasks(ctx context.Context, tenantID, driverEmployeeID uuid.UUID) ([]DeliveryTask, error)
	UpdateDeliveryTaskStatus(ctx context.Context, tenantID, taskID uuid.UUID, req UpdateDeliveryTaskStatusRequest) (*DeliveryTask, error)
	RecordDeliveryAttempt(ctx context.Context, tenantID, taskID uuid.UUID, req RecordDeliveryAttemptRequest) (*DeliveryAttempt, error)
	GetDeliveryAttempts(ctx context.Context, tenantID, taskID uuid.UUID) ([]DeliveryAttempt, error)
	SubmitDeliveryProof(ctx context.Context, tenantID, taskID uuid.UUID, req SubmitDeliveryProofRequest) error
}

type deliveryService struct {
	repo         Repository
	parcelRepo   parcels.Repository
	employeeRepo employees.Repository
}

func NewService(repo Repository, parcelRepo parcels.Repository, employeeRepo employees.Repository) Service {
	return &deliveryService{
		repo:         repo,
		parcelRepo:   parcelRepo,
		employeeRepo: employeeRepo,
	}
}

func (s *deliveryService) CreateDeliveryTask(ctx context.Context, tenantID uuid.UUID, req CreateDeliveryTaskRequest) (*DeliveryTask, error) {
	if err := ValidateCreateRequest(req); err != nil {
		return nil, err
	}

	pID, err := uuid.Parse(strings.TrimSpace(req.ParcelID))
	if err != nil {
		return nil, errors.New("invalid parcel_id format")
	}
	dID, err := uuid.Parse(strings.TrimSpace(req.AssignedDriverID))
	if err != nil {
		return nil, errors.New("invalid assigned_driver_id format")
	}

	var vID *uuid.UUID
	if req.VehicleID != nil && strings.TrimSpace(*req.VehicleID) != "" {
		parsedV, err := uuid.Parse(strings.TrimSpace(*req.VehicleID))
		if err != nil {
			return nil, errors.New("invalid vehicle_id format")
		}
		vID = &parsedV
	}

	// 1. Verify parcel belongs to tenant and is in dispatchable status
	parcel, err := s.parcelRepo.GetParcelByID(ctx, tenantID, pID)
	if err != nil {
		return nil, parcels.ErrParcelNotFound
	}

	eligibleStatuses := map[string]bool{
		parcels.StatusReceivedAtOriginBranch:   true,
		parcels.StatusReceivedAtTransferBranch: true,
		parcels.StatusOutForDelivery:           true,
		parcels.StatusDeliveryAttempted:        true,
	}
	if !eligibleStatuses[parcel.Status] {
		return nil, ErrParcelNotEligibleForDelivery
	}

	// 2. Verify driver belongs to tenant and has eligible role
	driver, err := s.employeeRepo.GetByID(ctx, tenantID, dID)
	if err != nil {
		return nil, employees.ErrEmployeeNotFound
	}
	if !driver.IsActive || driver.Status != employees.StatusActive {
		return nil, ErrDriverNotEligible
	}
	if driver.OperationalRole != employees.OperationalRoleDriver && driver.OperationalRole != employees.OperationalRoleDeliveryExecutive {
		return nil, ErrDriverNotEligible
	}

	branchID := parcel.OriginBranchID
	if parcel.CurrentBranchID != nil {
		branchID = *parcel.CurrentBranchID
	}

	priority := PriorityStandard
	if req.Priority != nil && strings.TrimSpace(*req.Priority) != "" {
		priority = strings.ToUpper(strings.TrimSpace(*req.Priority))
	}

	task := &DeliveryTask{
		TenantID:         tenantID,
		ParcelID:         pID,
		BranchID:         branchID,
		AssignedDriverID: dID,
		VehicleID:        vID,
		Status:           StatusAssigned,
		Priority:         priority,
		Notes:            req.Notes,
	}

	if err := s.repo.CreateDeliveryTask(ctx, task); err != nil {
		return nil, err
	}

	// Atomically mark parcel as OUT_FOR_DELIVERY if not already
	if parcel.Status != parcels.StatusOutForDelivery {
		notes := "Dispatched for last-mile delivery"
		_ = s.parcelRepo.UpdateParcelStatus(ctx, tenantID, pID, parcels.StatusOutForDelivery, &branchID, nil, "DISPATCHER", notes)
	}

	return s.repo.GetDeliveryTaskByID(ctx, tenantID, task.ID)
}

func (s *deliveryService) GetDeliveryTask(ctx context.Context, tenantID, taskID uuid.UUID) (*DeliveryTask, error) {
	return s.repo.GetDeliveryTaskByID(ctx, tenantID, taskID)
}

func (s *deliveryService) ListDeliveryTasks(ctx context.Context, tenantID uuid.UUID, filter DeliveryTaskFilter) (*DeliveryTaskListResponse, error) {
	tasks, total, err := s.repo.ListDeliveryTasks(ctx, tenantID, filter)
	if err != nil {
		return nil, err
	}

	return &DeliveryTaskListResponse{
		Tasks:  tasks,
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}, nil
}

func (s *deliveryService) GetDriverTasks(ctx context.Context, tenantID, driverEmployeeID uuid.UUID) ([]DeliveryTask, error) {
	filter := DeliveryTaskFilter{
		AssignedDriverID: &driverEmployeeID,
		Limit:            50,
		Offset:           0,
	}
	tasks, _, err := s.repo.ListDeliveryTasks(ctx, tenantID, filter)
	return tasks, err
}

func (s *deliveryService) UpdateDeliveryTaskStatus(ctx context.Context, tenantID, taskID uuid.UUID, req UpdateDeliveryTaskStatusRequest) (*DeliveryTask, error) {
	newStatus := strings.ToUpper(strings.TrimSpace(req.Status))
	if newStatus == "" {
		return nil, errors.New("status is required")
	}

	validStatuses := map[string]bool{
		StatusAssigned:    true,
		StatusInProgress:  true,
		StatusCompleted:   true,
		StatusFailed:      true,
		StatusCancelled:   true,
		StatusRescheduled: true,
	}
	if !validStatuses[newStatus] {
		return nil, ErrInvalidStatusTransition
	}

	if err := s.repo.UpdateDeliveryTaskStatus(ctx, tenantID, taskID, newStatus, req.Notes, req.FailureReason); err != nil {
		return nil, err
	}

	return s.repo.GetDeliveryTaskByID(ctx, tenantID, taskID)
}

func (s *deliveryService) RecordDeliveryAttempt(ctx context.Context, tenantID, taskID uuid.UUID, req RecordDeliveryAttemptRequest) (*DeliveryAttempt, error) {
	task, err := s.repo.GetDeliveryTaskByID(ctx, tenantID, taskID)
	if err != nil {
		return nil, err
	}

	outcome := strings.ToUpper(strings.TrimSpace(req.Outcome))
	validOutcomes := map[string]bool{
		OutcomeSuccessful:           true,
		OutcomeCustomerUnavailable:  true,
		OutcomeIncorrectAddress:     true,
		OutcomeRejectedByCustomer:   true,
		OutcomeSecurityAccessDenied: true,
		OutcomeOther:                true,
	}
	if !validOutcomes[outcome] {
		return nil, ErrInvalidAttemptOutcome
	}

	attempt := &DeliveryAttempt{
		DeliveryTaskID: taskID,
		Outcome:        outcome,
		Notes:          req.Notes,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
	}

	if err := s.repo.RecordDeliveryAttempt(ctx, attempt); err != nil {
		return nil, err
	}

	// Update parcel status to DELIVERY_ATTEMPTED
	notes := "Delivery attempt recorded: " + outcome
	if req.Notes != nil && *req.Notes != "" {
		notes += " - " + *req.Notes
	}
	_ = s.parcelRepo.UpdateParcelStatus(ctx, tenantID, task.ParcelID, parcels.StatusDeliveryAttempted, &task.BranchID, nil, "DRIVER", notes)

	return attempt, nil
}

func (s *deliveryService) GetDeliveryAttempts(ctx context.Context, tenantID, taskID uuid.UUID) ([]DeliveryAttempt, error) {
	if _, err := s.repo.GetDeliveryTaskByID(ctx, tenantID, taskID); err != nil {
		return nil, err
	}
	return s.repo.GetDeliveryAttempts(ctx, taskID)
}

func (s *deliveryService) SubmitDeliveryProof(ctx context.Context, tenantID, taskID uuid.UUID, req SubmitDeliveryProofRequest) error {
	if err := ValidateProofRequest(req); err != nil {
		return err
	}

	proofType := strings.ToUpper(strings.TrimSpace(req.ProofType))
	rel := "SELF"
	if req.RecipientRelationship != nil && strings.TrimSpace(*req.RecipientRelationship) != "" {
		rel = strings.ToUpper(strings.TrimSpace(*req.RecipientRelationship))
	}

	proof := &DeliveryProof{
		DeliveryTaskID:        taskID,
		ProofType:             proofType,
		RecipientName:         strings.TrimSpace(req.RecipientName),
		RecipientRelationship: rel,
		OTPCode:               req.OTPCode,
		SignatureData:         req.SignatureData,
		PhotoURL:              req.PhotoURL,
		Notes:                 req.Notes,
		Latitude:              req.Latitude,
		Longitude:             req.Longitude,
	}

	return s.repo.SubmitDeliveryProof(ctx, tenantID, proof)
}
