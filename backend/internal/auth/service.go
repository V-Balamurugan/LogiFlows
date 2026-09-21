package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/logiflows/logiflows/backend/internal/audit"
	"github.com/logiflows/logiflows/backend/internal/memberships"
	"github.com/logiflows/logiflows/backend/internal/tenants"
	"github.com/logiflows/logiflows/backend/internal/users"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountDeactivated = errors.New("user account is deactivated")
	ErrEmailAlreadyExists = errors.New("a user with this email already exists")
)

// Service defines the business logic contract for authentication & user identity.
type Service interface {
	Register(ctx context.Context, req RegisterRequest, ip, userAgent string) (*AuthResponse, error)
	Login(ctx context.Context, req LoginRequest, ip, userAgent string) (*AuthResponse, error)
	GetCurrentUser(ctx context.Context, userID uuid.UUID) (*CurrentUserResponse, error)
}

type authService struct {
	pool           *pgxpool.Pool
	userRepo       users.Repository
	tenantRepo     tenants.Repository
	membershipRepo memberships.Repository
	auditRepo      audit.Repository
	tokenService   *TokenService
}

// NewService creates a new authentication service.
func NewService(
	pool *pgxpool.Pool,
	userRepo users.Repository,
	tenantRepo tenants.Repository,
	membershipRepo memberships.Repository,
	auditRepo audit.Repository,
	tokenService *TokenService,
) Service {
	return &authService{
		pool:           pool,
		userRepo:       userRepo,
		tenantRepo:     tenantRepo,
		membershipRepo: membershipRepo,
		auditRepo:      auditRepo,
		tokenService:   tokenService,
	}
}

// Register registers a new user, creates their initial tenant company, and assigns them as TENANT_ADMIN.
func (s *authService) Register(ctx context.Context, req RegisterRequest, ip, userAgent string) (*AuthResponse, error) {
	// 1. Validate and hash password
	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 2. Prepare User model
	normalizedEmail := strings.ToLower(strings.TrimSpace(req.Email))
	var phone *string
	if strings.TrimSpace(req.PhoneNumber) != "" {
		p := strings.TrimSpace(req.PhoneNumber)
		phone = &p
	}

	user := &users.User{
		Email:           normalizedEmail,
		PasswordHash:    passwordHash,
		FullName:        strings.TrimSpace(req.FullName),
		PhoneNumber:     phone,
		IsActive:        true,
		IsPlatformAdmin: false,
	}

	// 3. Prepare Tenant model
	baseSlug := slugify(req.CompanyName)
	tenant := &tenants.Tenant{
		Name:         strings.TrimSpace(req.CompanyName),
		Slug:         baseSlug,
		Status:       tenants.StatusActive,
		ContactEmail: normalizedEmail,
	}

	// 4. Execute atomic transaction (User + Tenant + Membership)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := s.userRepo.CreateTx(ctx, tx, user); err != nil {
		if errors.Is(err, users.ErrUserAlreadyExists) {
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := s.tenantRepo.CreateTx(ctx, tx, tenant); err != nil {
		if errors.Is(err, tenants.ErrTenantAlreadyExists) {
			// Append random suffix to resolve collision
			tenant.Slug = fmt.Sprintf("%s-%s", baseSlug, uuid.New().String()[:6])
			if err := s.tenantRepo.CreateTx(ctx, tx, tenant); err != nil {
				return nil, fmt.Errorf("failed to create tenant with unique slug: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to create tenant: %w", err)
		}
	}

	membership := &memberships.TenantMembership{
		TenantID: tenant.ID,
		UserID:   user.ID,
		Role:     memberships.RoleTenantAdmin,
		Status:   memberships.StatusActive,
	}
	if err := s.membershipRepo.CreateTx(ctx, tx, membership); err != nil {
		return nil, fmt.Errorf("failed to create tenant membership: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit registration: %w", err)
	}

	// 5. Emit Audit Log
	_ = s.auditRepo.Log(ctx, &audit.AuditLog{
		TenantID:     &tenant.ID,
		UserID:       &user.ID,
		Action:       "USER_REGISTERED",
		ResourceType: "USER",
		ResourceID:   &user.Email,
		IPAddress:    &ip,
		UserAgent:    &userAgent,
		Status:       audit.StatusSuccess,
		Details: map[string]any{
			"company_name": tenant.Name,
			"tenant_slug":  tenant.Slug,
			"role":         membership.Role,
		},
	})

	// 6. Generate signed JWT token
	token, expiresAt, err := s.tokenService.GenerateAccessToken(user.ID, user.Email, user.IsPlatformAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return &AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User: UserSummary{
			ID:              user.ID,
			Email:           user.Email,
			FullName:        user.FullName,
			PhoneNumber:     user.PhoneNumber,
			IsActive:        user.IsActive,
			IsPlatformAdmin: user.IsPlatformAdmin,
		},
		Tenants: []TenantSummary{
			{
				ID:   tenant.ID,
				Name: tenant.Name,
				Slug: tenant.Slug,
				Role: membership.Role,
			},
		},
	}, nil
}

// Login verifies credentials and returns a signed JWT token and user profile.
func (s *authService) Login(ctx context.Context, req LoginRequest, ip, userAgent string) (*AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Constant-time check mitigation
		_ = ComparePassword("$2a$12$e8xL4R5u8r0p8tX1J5vVyeZp8i.3A5jF8QkUo.3n2pZ3O3.5K6s.", req.Password)
		s.logLoginAudit(ctx, nil, email, ip, userAgent, false, "USER_NOT_FOUND")
		return nil, ErrInvalidCredentials
	}

	if !ComparePassword(user.PasswordHash, req.Password) {
		s.logLoginAudit(ctx, &user.ID, email, ip, userAgent, false, "PASSWORD_MISMATCH")
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		s.logLoginAudit(ctx, &user.ID, email, ip, userAgent, false, "ACCOUNT_INACTIVE")
		return nil, ErrAccountDeactivated
	}

	// Update last login
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	// Fetch memberships & tenants
	membershipsList, err := s.membershipRepo.ListByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user memberships: %w", err)
	}

	var tenantSummaries []TenantSummary
	for _, m := range membershipsList {
		t, err := s.tenantRepo.GetByID(ctx, m.TenantID)
		if err == nil && t.Status == tenants.StatusActive {
			tenantSummaries = append(tenantSummaries, TenantSummary{
				ID:   t.ID,
				Name: t.Name,
				Slug: t.Slug,
				Role: m.Role,
			})
		}
	}

	// Generate access token
	token, expiresAt, err := s.tokenService.GenerateAccessToken(user.ID, user.Email, user.IsPlatformAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	s.logLoginAudit(ctx, &user.ID, email, ip, userAgent, true, "OK")

	return &AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User: UserSummary{
			ID:              user.ID,
			Email:           user.Email,
			FullName:        user.FullName,
			PhoneNumber:     user.PhoneNumber,
			IsActive:        user.IsActive,
			IsPlatformAdmin: user.IsPlatformAdmin,
		},
		Tenants: tenantSummaries,
	}, nil
}

// GetCurrentUser returns the authenticated user profile and their active tenant memberships.
func (s *authService) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*CurrentUserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	membershipsList, err := s.membershipRepo.ListByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user memberships: %w", err)
	}

	var tenantSummaries []TenantSummary
	for _, m := range membershipsList {
		t, err := s.tenantRepo.GetByID(ctx, m.TenantID)
		if err == nil && t.Status == tenants.StatusActive {
			tenantSummaries = append(tenantSummaries, TenantSummary{
				ID:   t.ID,
				Name: t.Name,
				Slug: t.Slug,
				Role: m.Role,
			})
		}
	}

	return &CurrentUserResponse{
		User: UserSummary{
			ID:              user.ID,
			Email:           user.Email,
			FullName:        user.FullName,
			PhoneNumber:     user.PhoneNumber,
			IsActive:        user.IsActive,
			IsPlatformAdmin: user.IsPlatformAdmin,
		},
		Tenants: tenantSummaries,
	}, nil
}

func (s *authService) logLoginAudit(ctx context.Context, userID *uuid.UUID, email, ip, userAgent string, success bool, reason string) {
	status := audit.StatusSuccess
	if !success {
		status = audit.StatusFailure
	}
	_ = s.auditRepo.Log(ctx, &audit.AuditLog{
		UserID:       userID,
		Action:       "USER_LOGIN",
		ResourceType: "USER",
		ResourceID:   &email,
		IPAddress:    &ip,
		UserAgent:    &userAgent,
		Status:       status,
		Details: map[string]any{
			"reason": reason,
		},
	})
}

func slugify(s string) string {
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
		return fmt.Sprintf("company-%s", uuid.New().String()[:8])
	}
	return res
}
