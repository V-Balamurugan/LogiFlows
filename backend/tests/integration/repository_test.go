package integration_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/logiflows/logiflows/backend/internal/audit"
	"github.com/logiflows/logiflows/backend/internal/config"
	"github.com/logiflows/logiflows/backend/internal/database"
	"github.com/logiflows/logiflows/backend/internal/memberships"
	"github.com/logiflows/logiflows/backend/internal/tenants"
	"github.com/logiflows/logiflows/backend/internal/users"
)

func TestRepositories_CRUD_And_Constraints(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping repository integration test in short mode")
	}

	_ = os.Setenv("APP_ENV", "test")
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	db, err := database.New(ctx, &cfg.Database)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	userRepo := users.NewRepository(db.Pool())
	tenantRepo := tenants.NewRepository(db.Pool())
	membershipRepo := memberships.NewRepository(db.Pool())
	auditRepo := audit.NewRepository(db.Pool())

	uniqueSuffix := time.Now().UnixNano()
	testEmail := fmt.Sprintf("testuser_%d@logiflows.com", uniqueSuffix)
	testSlug := fmt.Sprintf("test-company-%d", uniqueSuffix)

	// 1. Create User
	user := &users.User{
		Email:        testEmail,
		PasswordHash: "$2a$12$e8xL4R5u8r0p8tX1J5vVyeZp8i.3A5jF8QkUo.3n2pZ3O3.5K6s.",
		FullName:     "Integration Test User",
		IsActive:     true,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Verify duplicate email is rejected (case-insensitive)
	dupUser := &users.User{
		Email:        fmt.Sprintf("TESTUSER_%d@logiflows.com", uniqueSuffix),
		PasswordHash: "any-hash",
		FullName:     "Duplicate",
		IsActive:     true,
	}
	if err := userRepo.Create(ctx, dupUser); err != users.ErrUserAlreadyExists {
		t.Errorf("expected ErrUserAlreadyExists on duplicate email, got: %v", err)
	}

	// 2. Fetch User
	fetchedUser, err := userRepo.GetByEmail(ctx, testEmail)
	if err != nil {
		t.Fatalf("failed to fetch user by email: %v", err)
	}
	if fetchedUser.ID != user.ID {
		t.Errorf("expected user ID %v, got %v", user.ID, fetchedUser.ID)
	}

	// 3. Create Tenant
	tenant := &tenants.Tenant{
		Name:         "Integration Logistics Ltd",
		Slug:         testSlug,
		Status:       tenants.StatusActive,
		ContactEmail: testEmail,
	}
	if err := tenantRepo.Create(ctx, tenant); err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	// Verify duplicate slug is rejected
	dupTenant := &tenants.Tenant{
		Name:         "Duplicate Slug Co",
		Slug:         testSlug,
		Status:       tenants.StatusActive,
		ContactEmail: "other@example.com",
	}
	if err := tenantRepo.Create(ctx, dupTenant); err != tenants.ErrTenantAlreadyExists {
		t.Errorf("expected ErrTenantAlreadyExists on duplicate slug, got: %v", err)
	}

	// 4. Create Membership
	membership := &memberships.TenantMembership{
		TenantID: tenant.ID,
		UserID:   user.ID,
		Role:     memberships.RoleTenantAdmin,
		Status:   memberships.StatusActive,
	}
	if err := membershipRepo.Create(ctx, membership); err != nil {
		t.Fatalf("failed to create membership: %v", err)
	}

	// Verify duplicate membership is rejected
	if err := membershipRepo.Create(ctx, membership); err != memberships.ErrMembershipAlreadyExists {
		t.Errorf("expected ErrMembershipAlreadyExists on duplicate membership, got: %v", err)
	}

	// 5. Query Memberships
	memberDetails, err := membershipRepo.ListByTenantID(ctx, tenant.ID)
	if err != nil {
		t.Fatalf("failed to list tenant members: %v", err)
	}
	if len(memberDetails) != 1 || memberDetails[0].Email != testEmail {
		t.Errorf("expected 1 member with email %s, got: %+v", testEmail, memberDetails)
	}

	// 6. Log Audit Entry
	auditEntry := &audit.AuditLog{
		TenantID:     &tenant.ID,
		UserID:       &user.ID,
		Action:       "USER_REGISTERED",
		ResourceType: "USER",
		ResourceID:   &user.Email,
		Status:       audit.StatusSuccess,
		Details: map[string]any{
			"role":        memberships.RoleTenantAdmin,
			"tenant_slug": tenant.Slug,
		},
	}
	if err := auditRepo.Log(ctx, auditEntry); err != nil {
		t.Fatalf("failed to write audit log: %v", err)
	}

	// 7. Verify non-existent entities return proper errors
	_, err = userRepo.GetByID(ctx, uuid.New())
	if err != users.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got: %v", err)
	}

	_, err = tenantRepo.GetByID(ctx, uuid.New())
	if err != tenants.ErrTenantNotFound {
		t.Errorf("expected ErrTenantNotFound, got: %v", err)
	}

	_, err = membershipRepo.GetByUserAndTenant(ctx, uuid.New(), uuid.New())
	if err != memberships.ErrMembershipNotFound {
		t.Errorf("expected ErrMembershipNotFound, got: %v", err)
	}
}
