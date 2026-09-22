package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

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

const (
	DefaultRefreshTokenTTL = 7 * 24 * time.Hour
)

// Service defines the business logic contract for authentication & user identity.
type Service interface {
	Register(ctx context.Context, req RegisterRequest, ip, userAgent string) (*AuthResponse, error)
	Login(ctx context.Context, req LoginRequest, ip, userAgent string) (*AuthResponse, error)
	RefreshToken(ctx context.Context, req RefreshRequest, ip, userAgent string) (*AuthResponse, error)
	Logout(ctx context.Context, req LogoutRequest, userID *uuid.UUID, ip, userAgent string) error
	GetCurrentUser(ctx context.Context, userID uuid.UUID) (*CurrentUserResponse, error)
}

type authService struct {
	pool            *pgxpool.Pool
	userRepo        users.Repository
	tenantRepo      tenants.Repository
	membershipRepo  memberships.Repository
	auditRepo       audit.Repository
	tokenRepo       RefreshTokenRepository
	tokenService    *TokenService
	refreshTokenTTL time.Duration
}

// NewService creates a new authentication service.
func NewService(
	pool *pgxpool.Pool,
	userRepo users.Repository,
	tenantRepo tenants.Repository,
	membershipRepo memberships.Repository,
	auditRepo audit.Repository,
	tokenRepo RefreshTokenRepository,
	tokenService *TokenService,
) Service {
	return &authService{
		pool:            pool,
		userRepo:        userRepo,
		tenantRepo:      tenantRepo,
		membershipRepo:  membershipRepo,
		auditRepo:       auditRepo,
		tokenRepo:       tokenRepo,
		tokenService:    tokenService,
		refreshTokenTTL: DefaultRefreshTokenTTL,
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
		EmailVerified:   false,
	}

	// 3. Prepare Tenant model with preemptive collision resolution
	baseSlug := slugify(req.CompanyName)
	finalSlug := baseSlug
	if existing, _ := s.tenantRepo.GetBySlug(ctx, baseSlug); existing != nil {
		finalSlug = fmt.Sprintf("%s-%s", baseSlug, uuid.New().String()[:6])
	}
	tenant := &tenants.Tenant{
		Name:         strings.TrimSpace(req.CompanyName),
		Slug:         finalSlug,
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

	// 6. Generate signed JWT access token
	token, expiresAt, err := s.tokenService.GenerateAccessToken(user.ID, user.Email, user.IsPlatformAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// 7. Generate and persist refresh token
	rawRefreshToken, refreshHash, err := GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	refreshExpiry := time.Now().UTC().Add(s.refreshTokenTTL)
	rtRecord := &RefreshToken{
		UserID:    user.ID,
		TokenHash: refreshHash,
		ExpiresAt: refreshExpiry,
		IPAddress: &ip,
		UserAgent: &userAgent,
	}
	if err := s.tokenRepo.Create(ctx, rtRecord); err != nil {
		return nil, fmt.Errorf("failed to persist refresh token: %w", err)
	}

	return &AuthResponse{
		Token:                 token,
		ExpiresAt:             expiresAt,
		RefreshToken:          rawRefreshToken,
		RefreshTokenExpiresAt: refreshExpiry,
		User: UserSummary{
			ID:              user.ID,
			Email:           user.Email,
			FullName:        user.FullName,
			PhoneNumber:     user.PhoneNumber,
			IsActive:        user.IsActive,
			IsPlatformAdmin: user.IsPlatformAdmin,
			EmailVerified:   user.EmailVerified,
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

// Login verifies credentials and returns signed JWT access token, rotated refresh token, and user profile.
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

	// Generate and persist refresh token
	rawRefreshToken, refreshHash, err := GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	refreshExpiry := time.Now().UTC().Add(s.refreshTokenTTL)
	rtRecord := &RefreshToken{
		UserID:    user.ID,
		TokenHash: refreshHash,
		ExpiresAt: refreshExpiry,
		IPAddress: &ip,
		UserAgent: &userAgent,
	}
	if err := s.tokenRepo.Create(ctx, rtRecord); err != nil {
		return nil, fmt.Errorf("failed to persist refresh token: %w", err)
	}

	s.logLoginAudit(ctx, &user.ID, email, ip, userAgent, true, "OK")

	return &AuthResponse{
		Token:                 token,
		ExpiresAt:             expiresAt,
		RefreshToken:          rawRefreshToken,
		RefreshTokenExpiresAt: refreshExpiry,
		User: UserSummary{
			ID:              user.ID,
			Email:           user.Email,
			FullName:        user.FullName,
			PhoneNumber:     user.PhoneNumber,
			IsActive:        user.IsActive,
			IsPlatformAdmin: user.IsPlatformAdmin,
			EmailVerified:   user.EmailVerified,
		},
		Tenants: tenantSummaries,
	}, nil
}

// RefreshToken validates an opaque refresh token, performs single-use rotation, and returns a new token pair.
func (s *authService) RefreshToken(ctx context.Context, req RefreshRequest, ip, userAgent string) (*AuthResponse, error) {
	if strings.TrimSpace(req.RefreshToken) == "" {
		return nil, ErrInvalidToken
	}

	tokenHash := HashRefreshToken(req.RefreshToken)
	rt, err := s.tokenRepo.GetByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, ErrRefreshTokenNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, fmt.Errorf("failed to look up refresh token: %w", err)
	}

	// Breach detection: If an already-revoked refresh token is re-submitted,
	// immediately revoke ALL active refresh tokens for the compromised account.
	if rt.RevokedAt != nil {
		_ = s.tokenRepo.RevokeAllForUser(ctx, rt.UserID)
		_ = s.auditRepo.Log(ctx, &audit.AuditLog{
			UserID:       &rt.UserID,
			Action:       "REFRESH_TOKEN_REUSE_DETECTED",
			ResourceType: "REFRESH_TOKEN",
			ResourceID:   &rt.TokenHash,
			IPAddress:    &ip,
			UserAgent:    &userAgent,
			Status:       audit.StatusFailure,
			Details: map[string]any{
				"reason": "revoked_token_reuse_attempt",
			},
		})
		return nil, ErrRefreshTokenRevoked
	}

	if time.Now().UTC().After(rt.ExpiresAt) {
		return nil, ErrRefreshTokenExpired
	}

	user, err := s.userRepo.GetByID(ctx, rt.UserID)
	if err != nil || !user.IsActive {
		return nil, ErrAccountDeactivated
	}

	// 1. Single-use rotation: Revoke the currently consumed refresh token
	if err := s.tokenRepo.Revoke(ctx, rt.ID); err != nil {
		return nil, fmt.Errorf("failed to revoke old refresh token: %w", err)
	}

	// 2. Issue new signed access token
	token, expiresAt, err := s.tokenService.GenerateAccessToken(user.ID, user.Email, user.IsPlatformAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// 3. Issue new rotated refresh token
	rawRefreshToken, newRefreshHash, err := GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate new refresh token: %w", err)
	}
	refreshExpiry := time.Now().UTC().Add(s.refreshTokenTTL)
	newRT := &RefreshToken{
		UserID:    user.ID,
		TokenHash: newRefreshHash,
		ExpiresAt: refreshExpiry,
		IPAddress: &ip,
		UserAgent: &userAgent,
	}
	if err := s.tokenRepo.Create(ctx, newRT); err != nil {
		return nil, fmt.Errorf("failed to persist rotated refresh token: %w", err)
	}

	// Fetch tenant memberships
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

	_ = s.auditRepo.Log(ctx, &audit.AuditLog{
		UserID:       &user.ID,
		Action:       "TOKEN_REFRESHED",
		ResourceType: "REFRESH_TOKEN",
		IPAddress:    &ip,
		UserAgent:    &userAgent,
		Status:       audit.StatusSuccess,
	})

	return &AuthResponse{
		Token:                 token,
		ExpiresAt:             expiresAt,
		RefreshToken:          rawRefreshToken,
		RefreshTokenExpiresAt: refreshExpiry,
		User: UserSummary{
			ID:              user.ID,
			Email:           user.Email,
			FullName:        user.FullName,
			PhoneNumber:     user.PhoneNumber,
			IsActive:        user.IsActive,
			IsPlatformAdmin: user.IsPlatformAdmin,
			EmailVerified:   user.EmailVerified,
		},
		Tenants: tenantSummaries,
	}, nil
}

// Logout revokes the specified refresh token (or user's sessions) and records an audit log event.
func (s *authService) Logout(ctx context.Context, req LogoutRequest, userID *uuid.UUID, ip, userAgent string) error {
	if strings.TrimSpace(req.RefreshToken) != "" {
		hash := HashRefreshToken(req.RefreshToken)
		if rt, err := s.tokenRepo.GetByHash(ctx, hash); err == nil {
			_ = s.tokenRepo.Revoke(ctx, rt.ID)
			if userID == nil {
				userID = &rt.UserID
			}
		}
	} else if userID != nil {
		// If no specific refresh token provided, revoke all tokens for this user
		_ = s.tokenRepo.RevokeAllForUser(ctx, *userID)
	}

	_ = s.auditRepo.Log(ctx, &audit.AuditLog{
		UserID:       userID,
		Action:       "USER_LOGOUT",
		ResourceType: "USER",
		IPAddress:    &ip,
		UserAgent:    &userAgent,
		Status:       audit.StatusSuccess,
	})
	return nil
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
			EmailVerified:   user.EmailVerified,
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
