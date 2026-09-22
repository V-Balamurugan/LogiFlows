package tenants

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrTenantNotFound      = errors.New("tenant not found")
	ErrTenantAlreadyExists = errors.New("tenant with this slug already exists")
)

// Repository defines persistence operations for tenants.
type Repository interface {
	Create(ctx context.Context, tenant *Tenant) error
	CreateTx(ctx context.Context, tx pgx.Tx, tenant *Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*Tenant, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]Tenant, error)
	ListAll(ctx context.Context) ([]Tenant, error)
	Update(ctx context.Context, tenant *Tenant) error
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository initializes a PostgreSQL tenant repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) Create(ctx context.Context, tenant *Tenant) error {
	return insertTenant(ctx, r.pool, tenant)
}

func (r *pgRepository) CreateTx(ctx context.Context, tx pgx.Tx, tenant *Tenant) error {
	return insertTenant(ctx, tx, tenant)
}

type dbExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
}

func insertTenant(ctx context.Context, exec dbExecutor, tenant *Tenant) error {
	if tenant.ID == uuid.Nil {
		tenant.ID = uuid.New()
	}
	now := time.Now().UTC()
	tenant.CreatedAt = now
	tenant.UpdatedAt = now
	tenant.Slug = strings.ToLower(strings.TrimSpace(tenant.Slug))
	if tenant.Status == "" {
		tenant.Status = StatusActive
	}

	query := `
		INSERT INTO tenants (id, name, slug, status, contact_email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := exec.Exec(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.Slug,
		tenant.Status,
		tenant.ContactEmail,
		tenant.CreatedAt,
		tenant.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // Unique violation
			return ErrTenantAlreadyExists
		}
		return fmt.Errorf("failed to insert tenant: %w", err)
	}

	return nil
}

func (r *pgRepository) GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error) {
	query := `SELECT id, name, slug, status, contact_email, created_at, updated_at FROM tenants WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)
	return scanTenant(row)
}

func (r *pgRepository) GetBySlug(ctx context.Context, slug string) (*Tenant, error) {
	query := `SELECT id, name, slug, status, contact_email, created_at, updated_at FROM tenants WHERE slug = $1`
	row := r.pool.QueryRow(ctx, query, strings.ToLower(strings.TrimSpace(slug)))
	return scanTenant(row)
}

func (r *pgRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]Tenant, error) {
	query := `
		SELECT t.id, t.name, t.slug, t.status, t.contact_email, t.created_at, t.updated_at
		FROM tenants t
		INNER JOIN tenant_memberships tm ON t.id = tm.tenant_id
		WHERE tm.user_id = $1 AND tm.status = 'ACTIVE' AND t.status = 'ACTIVE'
		ORDER BY t.name ASC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tenants for user: %w", err)
	}
	defer rows.Close()

	var result []Tenant
	for rows.Next() {
		var t Tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Status, &t.ContactEmail, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan tenant row: %w", err)
		}
		result = append(result, t)
	}
	return result, nil
}

func (r *pgRepository) ListAll(ctx context.Context) ([]Tenant, error) {
	query := `SELECT id, name, slug, status, contact_email, created_at, updated_at FROM tenants ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all tenants: %w", err)
	}
	defer rows.Close()

	var result []Tenant
	for rows.Next() {
		var t Tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Status, &t.ContactEmail, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan tenant row: %w", err)
		}
		result = append(result, t)
	}
	return result, nil
}

func (r *pgRepository) Update(ctx context.Context, tenant *Tenant) error {
	tenant.UpdatedAt = time.Now().UTC()
	query := `
		UPDATE tenants
		SET name = $1, contact_email = $2, updated_at = $3
		WHERE id = $4
	`
	tag, err := r.pool.Exec(ctx, query, tenant.Name, tenant.ContactEmail, tenant.UpdatedAt, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrTenantNotFound
	}
	return nil
}

func scanTenant(row pgx.Row) (*Tenant, error) {
	var t Tenant
	err := row.Scan(&t.ID, &t.Name, &t.Slug, &t.Status, &t.ContactEmail, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTenantNotFound
		}
		return nil, fmt.Errorf("failed to scan tenant row: %w", err)
	}
	return &t, nil
}
