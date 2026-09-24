package transfers

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/logiflows/logiflows/backend/internal/branches"
)

type Service interface {
	CreateTransfer(ctx context.Context, tenantID, actorID uuid.UUID, req CreateBranchTransferRequest) (*BranchTransfer, error)
	GetTransfer(ctx context.Context, tenantID, transferID uuid.UUID) (*BranchTransfer, []BranchTransferParcel, error)
	ListTransfers(ctx context.Context, tenantID uuid.UUID, filter BranchTransferFilter) (*BranchTransferListResponse, error)
	DispatchTransfer(ctx context.Context, tenantID, transferID uuid.UUID, req DispatchTransferRequest) error
	ReceiveTransfer(ctx context.Context, tenantID, transferID uuid.UUID, req ReceiveTransferRequest) error
}

type transferService struct {
	repo       Repository
	branchRepo branches.Repository
}

func NewService(repo Repository, branchRepo branches.Repository) Service {
	return &transferService{
		repo:       repo,
		branchRepo: branchRepo,
	}
}

func (s *transferService) CreateTransfer(ctx context.Context, tenantID, actorID uuid.UUID, req CreateBranchTransferRequest) (*BranchTransfer, error) {
	if err := ValidateCreateRequest(req); err != nil {
		return nil, err
	}

	srcID, err := uuid.Parse(strings.TrimSpace(req.SourceBranchID))
	if err != nil {
		return nil, ErrInvalidBranches
	}
	destID, err := uuid.Parse(strings.TrimSpace(req.DestinationBranchID))
	if err != nil {
		return nil, ErrInvalidBranches
	}

	// Verify source branch belongs to tenant
	srcBranch, err := s.branchRepo.GetByID(ctx, tenantID, srcID)
	if err != nil || srcBranch.TenantID != tenantID {
		return nil, errors.New("source branch does not belong to this organization")
	}

	// Verify destination branch belongs to tenant
	destBranch, err := s.branchRepo.GetByID(ctx, tenantID, destID)
	if err != nil || destBranch.TenantID != tenantID {
		return nil, errors.New("destination branch does not belong to this organization")
	}

	var dID *uuid.UUID
	if req.DriverID != nil && strings.TrimSpace(*req.DriverID) != "" {
		id, err := uuid.Parse(strings.TrimSpace(*req.DriverID))
		if err != nil {
			return nil, errors.New("invalid driver_id format")
		}
		dID = &id
	}

	var vID *uuid.UUID
	if req.VehicleID != nil && strings.TrimSpace(*req.VehicleID) != "" {
		id, err := uuid.Parse(strings.TrimSpace(*req.VehicleID))
		if err != nil {
			return nil, errors.New("invalid vehicle_id format")
		}
		vID = &id
	}

	parcelUUIDs := make([]uuid.UUID, 0, len(req.ParcelIDs))
	for _, pStr := range req.ParcelIDs {
		pID, err := uuid.Parse(strings.TrimSpace(pStr))
		if err != nil {
			return nil, errors.New("invalid parcel_id format in parcel_ids list")
		}
		parcelUUIDs = append(parcelUUIDs, pID)
	}

	transferNumber, err := s.repo.NextTransferNumber(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	var createdBy *uuid.UUID
	if actorID != uuid.Nil {
		createdBy = &actorID
	}

	transfer := &BranchTransfer{
		TenantID:            tenantID,
		TransferNumber:      transferNumber,
		SourceBranchID:      srcID,
		DestinationBranchID: destID,
		DriverID:            dID,
		VehicleID:           vID,
		Status:              StatusPending,
		Notes:               req.Notes,
		CreatedBy:           createdBy,
	}

	if err := s.repo.CreateTransfer(ctx, transfer, parcelUUIDs); err != nil {
		return nil, err
	}

	bt, _, err := s.repo.GetTransferByID(ctx, tenantID, transfer.ID)
	return bt, err
}

func (s *transferService) GetTransfer(ctx context.Context, tenantID, transferID uuid.UUID) (*BranchTransfer, []BranchTransferParcel, error) {
	return s.repo.GetTransferByID(ctx, tenantID, transferID)
}

func (s *transferService) ListTransfers(ctx context.Context, tenantID uuid.UUID, filter BranchTransferFilter) (*BranchTransferListResponse, error) {
	transfers, total, err := s.repo.ListTransfers(ctx, tenantID, filter)
	if err != nil {
		return nil, err
	}

	return &BranchTransferListResponse{
		Transfers: transfers,
		Total:     total,
		Limit:     filter.Limit,
		Offset:    filter.Offset,
	}, nil
}

func (s *transferService) DispatchTransfer(ctx context.Context, tenantID, transferID uuid.UUID, req DispatchTransferRequest) error {
	return s.repo.DispatchTransfer(ctx, tenantID, transferID, req.Notes)
}

func (s *transferService) ReceiveTransfer(ctx context.Context, tenantID, transferID uuid.UUID, req ReceiveTransferRequest) error {
	parcelUUIDs := make([]uuid.UUID, 0, len(req.ParcelIDs))
	for _, pStr := range req.ParcelIDs {
		if id, err := uuid.Parse(strings.TrimSpace(pStr)); err == nil {
			parcelUUIDs = append(parcelUUIDs, id)
		}
	}
	return s.repo.ReceiveTransfer(ctx, tenantID, transferID, parcelUUIDs, req.Notes)
}
