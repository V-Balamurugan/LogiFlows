package tenants

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/logiflows/logiflows/backend/internal/audit"
	"github.com/logiflows/logiflows/backend/internal/authorization"
	"github.com/logiflows/logiflows/backend/internal/memberships"
	"github.com/logiflows/logiflows/backend/internal/users"
)

var (
	ErrInvalidRole   = errors.New("invalid membership role")
	ErrUserNotFound  = errors.New("user with this email does not exist")
	ErrAlreadyMember = errors.New("user is already a member of this tenant")
)

// Service defines the operations contract for multi-tenant management.
type Service interface {
	CreateTenant(ctx context.Context, creatorID uuid.UUID, req CreateTenantRequest) (*Tenant, error)
	GetTenant(ctx context.Context, tenantID uuid.UUID) (*Tenant, error)
	ListUserTenants(ctx context.Context, userID uuid.UUID, isPlatformAdmin bool) ([]Tenant, error)
	ListMembers(ctx context.Context, tenantID uuid.UUID) ([]memberships.MemberDetails, error)
	AddMember(ctx context.Context, tenantID uuid.UUID, req AddMemberRequest) (*memberships.MemberDetails, error)
}

type tenantService struct {
	pool           *pgxpool.Pool
	tenantRepo     Repository
	membershipRepo memberships.Repository
	userRepo       users.Repository
	auditRepo      audit.Repository
}

// NewService creates a new tenant management service.
func NewService(
	pool *pgxpool.Pool,
	tenantRepo Repository,
	membershipRepo memberships.Repository,
	userRepo users.Repository,
	auditRepo audit.Repository,
) Service {
	return &tenantService{
		pool:           pool,
		tenantRepo:     tenantRepo,
		membershipRepo: membershipRepo,
		userRepo:       userRepo,
		auditRepo:      auditRepo,
	}
}

func (s *tenantService) CreateTenant(ctx context.Context, creatorID uuid.UUID, req CreateTenantRequest) (*Tenant, error) {
	name := strings.TrimSpace(req.Name)
	contactEmail := strings.ToLower(strings.TrimSpace(req.ContactEmail))
	slug := slugifyTenant(name)

	tenant := &Tenant{
		Name:         name,
		Slug:         slug,
		Status:       StatusActive,
		ContactEmail: contactEmail,
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := s.tenantRepo.CreateTx(ctx, tx, tenant); err != nil {
		if errors.Is(err, ErrTenantAlreadyExists) {
			tenant.Slug = fmt.Sprintf("%s-%s", slug, uuid.New().String()[:6])
			if err := s.tenantRepo.CreateTx(ctx, tx, tenant); err != nil {
				return nil, fmt.Errorf("failed to create tenant with unique slug: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to create tenant: %w", err)
		}
	}

	membership := &memberships.TenantMembership{
		TenantID: tenant.ID,
		UserID:   creatorID,
		Role:     memberships.RoleTenantAdmin,
		Status:   memberships.StatusActive,
	}
	if err := s.membershipRepo.CreateTx(ctx, tx, membership); err != nil {
		return nil, fmt.Errorf("failed to assign creator membership: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit tenant creation: %w", err)
	}

	_ = s.auditRepo.Log(ctx, &audit.AuditLog{
		TenantID:     &tenant.ID,
		UserID:       &creatorID,
		Action:       "TENANT_CREATED",
		ResourceType: "TENANT",
		ResourceID:   &tenant.Slug,
		Status:       audit.StatusSuccess,
		Details: map[string]any{
			"name": tenant.Name,
		},
	})

	return tenant, nil
}

func (s *tenantService) GetTenant(ctx context.Context, tenantID uuid.UUID) (*Tenant, error) {
	return s.tenantRepo.GetByID(ctx, tenantID)
}

func (s *tenantService) ListUserTenants(ctx context.Context, userID uuid.UUID, isPlatformAdmin bool) ([]Tenant, error) {
	if isPlatformAdmin {
		return s.tenantRepo.ListAll(ctx)
	}
	return s.tenantRepo.ListByUserID(ctx, userID)
}

func (s *tenantService) ListMembers(ctx context.Context, tenantID uuid.UUID) ([]memberships.MemberDetails, error) {
	return s.membershipRepo.ListByTenantID(ctx, tenantID)
}

func (s *tenantService) AddMember(ctx context.Context, tenantID uuid.UUID, req AddMemberRequest) (*memberships.MemberDetails, error) {
	role := strings.ToUpper(strings.TrimSpace(req.Role))
	if !authorization.IsValidRole(role) || role == authorization.RolePlatformAdmin {
		return nil, ErrInvalidRole
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	membership := &memberships.TenantMembership{
		TenantID: tenantID,
		UserID:   user.ID,
		Role:     role,
		Status:   memberships.StatusActive,
	}

	if err := s.membershipRepo.Create(ctx, membership); err != nil {
		if errors.Is(err, memberships.ErrMembershipAlreadyExists) {
			return nil, ErrAlreadyMember
		}
		return nil, err
	}

	_ = s.auditRepo.Log(ctx, &audit.AuditLog{
		TenantID:     &tenantID,
		UserID:       &user.ID,
		Action:       "TENANT_MEMBER_ADDED",
		ResourceType: "MEMBERSHIP",
		ResourceID:   &user.Email,
		Status:       audit.StatusSuccess,
		Details: map[string]any{
			"role": role,
		},
	})

	return &memberships.MemberDetails{
		ID:        membership.ID,
		TenantID:  tenantID,
		UserID:    user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		Role:      role,
		Status:    memberships.StatusActive,
		CreatedAt: membership.CreatedAt,
	}, nil
}

func slugifyTenant(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else if r == ' ' || r == '-' || r == '_' {
			b.WriteRune('-')
		}
	}
	res := b.String()
	for strings.Contains(res, "--") {
		res = strings.ReplaceAll(res, "--", "-")
	}
	res = strings.Trim(res, "-")
	if res == "" {
		return fmt.Sprintf("tenant-%s", uuid.New().String()[:8])
	}
	return res
}
