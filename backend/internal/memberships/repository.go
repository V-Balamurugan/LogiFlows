package memberships

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrMembershipNotFound      = errors.New("tenant membership not found")
	ErrMembershipAlreadyExists = errors.New("user is already a member of this tenant")
)

// Repository defines persistence operations for tenant memberships.
type Repository interface {
	Create(ctx context.Context, membership *TenantMembership) error
	CreateTx(ctx context.Context, tx pgx.Tx, membership *TenantMembership) error
	GetByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) (*TenantMembership, error)
	ListByTenantID(ctx context.Context, tenantID uuid.UUID) ([]MemberDetails, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]TenantMembership, error)
	UpdateRole(ctx context.Context, tenantID, userID uuid.UUID, newRole string) error
	Delete(ctx context.Context, tenantID, userID uuid.UUID) error
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository initializes a PostgreSQL membership repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) Create(ctx context.Context, membership *TenantMembership) error {
	return insertMembership(ctx, r.pool, membership)
}

func (r *pgRepository) CreateTx(ctx context.Context, tx pgx.Tx, membership *TenantMembership) error {
	return insertMembership(ctx, tx, membership)
}

type dbExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
}

func insertMembership(ctx context.Context, exec dbExecutor, m *TenantMembership) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	now := time.Now().UTC()
	m.CreatedAt = now
	m.UpdatedAt = now
	if m.Status == "" {
		m.Status = StatusActive
	}

	query := `
		INSERT INTO tenant_memberships (id, tenant_id, user_id, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := exec.Exec(ctx, query,
		m.ID,
		m.TenantID,
		m.UserID,
		m.Role,
		m.Status,
		m.CreatedAt,
		m.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // Unique violation
			return ErrMembershipAlreadyExists
		}
		return fmt.Errorf("failed to insert membership: %w", err)
	}

	return nil
}

func (r *pgRepository) GetByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) (*TenantMembership, error) {
	query := `
		SELECT id, tenant_id, user_id, role, status, created_at, updated_at
		FROM tenant_memberships
		WHERE user_id = $1 AND tenant_id = $2
	`
	row := r.pool.QueryRow(ctx, query, userID, tenantID)

	var m TenantMembership
	err := row.Scan(&m.ID, &m.TenantID, &m.UserID, &m.Role, &m.Status, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMembershipNotFound
		}
		return nil, fmt.Errorf("failed to scan membership row: %w", err)
	}
	return &m, nil
}

func (r *pgRepository) ListByTenantID(ctx context.Context, tenantID uuid.UUID) ([]MemberDetails, error) {
	query := `
		SELECT tm.id, tm.tenant_id, tm.user_id, u.email, u.full_name, tm.role, tm.status, tm.created_at
		FROM tenant_memberships tm
		INNER JOIN users u ON tm.user_id = u.id
		WHERE tm.tenant_id = $1
		ORDER BY tm.created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query members for tenant: %w", err)
	}
	defer rows.Close()

	var list []MemberDetails
	for rows.Next() {
		var d MemberDetails
		if err := rows.Scan(&d.ID, &d.TenantID, &d.UserID, &d.Email, &d.FullName, &d.Role, &d.Status, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan member detail row: %w", err)
		}
		list = append(list, d)
	}
	return list, nil
}

func (r *pgRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]TenantMembership, error) {
	query := `
		SELECT id, tenant_id, user_id, role, status, created_at, updated_at
		FROM tenant_memberships
		WHERE user_id = $1 AND status = 'ACTIVE'
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query memberships for user: %w", err)
	}
	defer rows.Close()

	var list []TenantMembership
	for rows.Next() {
		var m TenantMembership
		if err := rows.Scan(&m.ID, &m.TenantID, &m.UserID, &m.Role, &m.Status, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan membership row: %w", err)
		}
		list = append(list, m)
	}
	return list, nil
}

func (r *pgRepository) UpdateRole(ctx context.Context, tenantID, userID uuid.UUID, newRole string) error {
	query := `UPDATE tenant_memberships SET role = $1, updated_at = $2 WHERE tenant_id = $3 AND user_id = $4`
	cmd, err := r.pool.Exec(ctx, query, newRole, time.Now().UTC(), tenantID, userID)
	if err != nil {
		return fmt.Errorf("failed to update membership role: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrMembershipNotFound
	}
	return nil
}

func (r *pgRepository) Delete(ctx context.Context, tenantID, userID uuid.UUID) error {
	query := `DELETE FROM tenant_memberships WHERE tenant_id = $1 AND user_id = $2`
	cmd, err := r.pool.Exec(ctx, query, tenantID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete membership: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrMembershipNotFound
	}
	return nil
}
