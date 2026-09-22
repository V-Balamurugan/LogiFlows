package branches

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/logiflows/logiflows/backend/internal/audit"
)

// Service defines business operations for delivery branches.
type Service interface {
	CreateBranch(ctx context.Context, tenantID, userID uuid.UUID, req CreateBranchRequest, ip, userAgent string) (*Branch, error)
	GetBranch(ctx context.Context, tenantID, branchID uuid.UUID) (*Branch, error)
	ListBranches(ctx context.Context, tenantID uuid.UUID, filter BranchFilter) (*BranchListResponse, error)
	UpdateBranch(ctx context.Context, tenantID, branchID, userID uuid.UUID, req UpdateBranchRequest, ip, userAgent string) (*Branch, error)
	DeactivateBranch(ctx context.Context, tenantID, branchID, userID uuid.UUID, ip, userAgent string) error
}

type branchService struct {
	repo      Repository
	auditRepo audit.Repository
}

// NewService initializes a new branch service.
func NewService(repo Repository, auditRepo audit.Repository) Service {
	return &branchService{
		repo:      repo,
		auditRepo: auditRepo,
	}
}

func (s *branchService) CreateBranch(ctx context.Context, tenantID, userID uuid.UUID, req CreateBranchRequest, ip, userAgent string) (*Branch, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Verify code uniqueness within this tenant
	existing, err := s.repo.GetByCode(ctx, tenantID, req.BranchCode)
	if err == nil && existing != nil {
		return nil, ErrBranchCodeExists
	}

	branch := &Branch{
		TenantID:         tenantID,
		BranchCode:       req.BranchCode,
		Name:             req.Name,
		Address:          req.Address,
		City:             req.City,
		State:            req.State,
		PostalCode:       req.PostalCode,
		Country:          req.Country,
		Latitude:         req.Latitude,
		Longitude:        req.Longitude,
		CoverageRadiusKM: req.CoverageRadiusKM,
		Status:           StatusActive,
		IsActive:         true,
	}

	if err := s.repo.Create(ctx, branch); err != nil {
		return nil, fmt.Errorf("failed to persist branch: %w", err)
	}

	// Emit security audit log
	if s.auditRepo != nil {
		branchIDStr := branch.ID.String()
		_ = s.auditRepo.Log(ctx, &audit.AuditLog{
			TenantID:     &tenantID,
			UserID:       &userID,
			Action:       "BRANCH_CREATED",
			ResourceType: "BRANCH",
			ResourceID:   &branchIDStr,
			IPAddress:    &ip,
			UserAgent:    &userAgent,
			Status:       audit.StatusSuccess,
			Details: map[string]any{
				"branch_code":        branch.BranchCode,
				"name":               branch.Name,
				"city":               branch.City,
				"latitude":           branch.Latitude,
				"longitude":          branch.Longitude,
				"coverage_radius_km": branch.CoverageRadiusKM,
			},
		})
	}

	return branch, nil
}

func (s *branchService) GetBranch(ctx context.Context, tenantID, branchID uuid.UUID) (*Branch, error) {
	return s.repo.GetByID(ctx, tenantID, branchID)
}

func (s *branchService) ListBranches(ctx context.Context, tenantID uuid.UUID, filter BranchFilter) (*BranchListResponse, error) {
	items, total, err := s.repo.List(ctx, tenantID, filter)
	if err != nil {
		return nil, err
	}

	page := filter.Page
	if page <= 0 {
		page = 1
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}

	return &BranchListResponse{
		Branches: items,
		Total:    total,
		Page:     page,
		Limit:    limit,
	}, nil
}

func (s *branchService) UpdateBranch(ctx context.Context, tenantID, branchID, userID uuid.UUID, req UpdateBranchRequest, ip, userAgent string) (*Branch, error) {
	branch, err := s.repo.GetByID(ctx, tenantID, branchID)
	if err != nil {
		return nil, err
	}

	// Apply field modifications
	if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
		branch.Name = strings.TrimSpace(*req.Name)
	}
	if req.Address != nil && strings.TrimSpace(*req.Address) != "" {
		branch.Address = strings.TrimSpace(*req.Address)
	}
	if req.City != nil && strings.TrimSpace(*req.City) != "" {
		branch.City = strings.TrimSpace(*req.City)
	}
	if req.State != nil {
		branch.State = strings.TrimSpace(*req.State)
	}
	if req.PostalCode != nil {
		branch.PostalCode = strings.TrimSpace(*req.PostalCode)
	}
	if req.Country != nil && strings.TrimSpace(*req.Country) != "" {
		branch.Country = strings.TrimSpace(*req.Country)
	}
	if req.Latitude != nil {
		if *req.Latitude < -90.0 || *req.Latitude > 90.0 {
			return nil, ErrInvalidCoordinates
		}
		branch.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		if *req.Longitude < -180.0 || *req.Longitude > 180.0 {
			return nil, ErrInvalidCoordinates
		}
		branch.Longitude = *req.Longitude
	}
	if req.CoverageRadiusKM != nil {
		if *req.CoverageRadiusKM <= 0 {
			return nil, ErrInvalidCoverageRadius
		}
		branch.CoverageRadiusKM = *req.CoverageRadiusKM
	}
	if req.Status != nil {
		status := strings.ToUpper(strings.TrimSpace(*req.Status))
		if status != StatusActive && status != StatusInactive && status != StatusSuspended {
			return nil, ErrInvalidBranchStatus
		}
		branch.Status = status
		if status == StatusInactive || status == StatusSuspended {
			branch.IsActive = false
		} else {
			branch.IsActive = true
		}
	}

	if err := s.repo.Update(ctx, branch); err != nil {
		return nil, err
	}

	// Emit security audit log
	if s.auditRepo != nil {
		branchIDStr := branch.ID.String()
		_ = s.auditRepo.Log(ctx, &audit.AuditLog{
			TenantID:     &tenantID,
			UserID:       &userID,
			Action:       "BRANCH_UPDATED",
			ResourceType: "BRANCH",
			ResourceID:   &branchIDStr,
			IPAddress:    &ip,
			UserAgent:    &userAgent,
			Status:       audit.StatusSuccess,
			Details: map[string]any{
				"branch_code": branch.BranchCode,
				"name":        branch.Name,
				"status":      branch.Status,
			},
		})
	}

	return branch, nil
}

func (s *branchService) DeactivateBranch(ctx context.Context, tenantID, branchID, userID uuid.UUID, ip, userAgent string) error {
	branch, err := s.repo.GetByID(ctx, tenantID, branchID)
	if err != nil {
		return err
	}

	if err := s.repo.Deactivate(ctx, tenantID, branchID); err != nil {
		return err
	}

	// Emit security audit log
	if s.auditRepo != nil {
		branchIDStr := branch.ID.String()
		_ = s.auditRepo.Log(ctx, &audit.AuditLog{
			TenantID:     &tenantID,
			UserID:       &userID,
			Action:       "BRANCH_DEACTIVATED",
			ResourceType: "BRANCH",
			ResourceID:   &branchIDStr,
			IPAddress:    &ip,
			UserAgent:    &userAgent,
			Status:       audit.StatusSuccess,
			Details: map[string]any{
				"branch_code": branch.BranchCode,
				"name":        branch.Name,
			},
		})
	}

	return nil
}
