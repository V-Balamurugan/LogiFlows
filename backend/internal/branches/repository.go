package branches

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines database operations for the branches domain.
type Repository interface {
	Create(ctx context.Context, branch *Branch) error
	GetByID(ctx context.Context, tenantID, branchID uuid.UUID) (*Branch, error)
	GetByCode(ctx context.Context, tenantID uuid.UUID, branchCode string) (*Branch, error)
	List(ctx context.Context, tenantID uuid.UUID, filter BranchFilter) ([]Branch, int, error)
	Update(ctx context.Context, branch *Branch) error
	Deactivate(ctx context.Context, tenantID, branchID uuid.UUID) error
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository initializes a new PostgreSQL branches repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) Create(ctx context.Context, b *Branch) error {
	query := `
		INSERT INTO branches (
			tenant_id, branch_code, name, address, city, state, postal_code, country,
			latitude, longitude, location, coverage_radius_km, status, is_active
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, ST_SetSRID(ST_MakePoint($10, $9), 4326), $11, $12, $13
		)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		b.TenantID,
		b.BranchCode,
		b.Name,
		b.Address,
		b.City,
		b.State,
		b.PostalCode,
		b.Country,
		b.Latitude,
		b.Longitude,
		b.CoverageRadiusKM,
		b.Status,
		b.IsActive,
	).Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return ErrBranchCodeExists
		}
		return fmt.Errorf("failed to insert branch: %w", err)
	}

	return nil
}

func (r *pgRepository) GetByID(ctx context.Context, tenantID, branchID uuid.UUID) (*Branch, error) {
	query := `
		SELECT 
			id, tenant_id, branch_code, name, address, city, state, postal_code, country,
			latitude, longitude, coverage_radius_km, status, is_active, created_at, updated_at
		FROM branches
		WHERE id = $1 AND tenant_id = $2
	`

	b := &Branch{}
	var state, postalCode *string
	err := r.pool.QueryRow(ctx, query, branchID, tenantID).Scan(
		&b.ID,
		&b.TenantID,
		&b.BranchCode,
		&b.Name,
		&b.Address,
		&b.City,
		&state,
		&postalCode,
		&b.Country,
		&b.Latitude,
		&b.Longitude,
		&b.CoverageRadiusKM,
		&b.Status,
		&b.IsActive,
		&b.CreatedAt,
		&b.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBranchNotFound
		}
		return nil, fmt.Errorf("failed to query branch by id: %w", err)
	}

	if state != nil {
		b.State = *state
	}
	if postalCode != nil {
		b.PostalCode = *postalCode
	}

	return b, nil
}

func (r *pgRepository) GetByCode(ctx context.Context, tenantID uuid.UUID, branchCode string) (*Branch, error) {
	query := `
		SELECT 
			id, tenant_id, branch_code, name, address, city, state, postal_code, country,
			latitude, longitude, coverage_radius_km, status, is_active, created_at, updated_at
		FROM branches
		WHERE tenant_id = $1 AND UPPER(branch_code) = UPPER($2)
	`

	b := &Branch{}
	var state, postalCode *string
	err := r.pool.QueryRow(ctx, query, tenantID, branchCode).Scan(
		&b.ID,
		&b.TenantID,
		&b.BranchCode,
		&b.Name,
		&b.Address,
		&b.City,
		&state,
		&postalCode,
		&b.Country,
		&b.Latitude,
		&b.Longitude,
		&b.CoverageRadiusKM,
		&b.Status,
		&b.IsActive,
		&b.CreatedAt,
		&b.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBranchNotFound
		}
		return nil, fmt.Errorf("failed to query branch by code: %w", err)
	}

	if state != nil {
		b.State = *state
	}
	if postalCode != nil {
		b.PostalCode = *postalCode
	}

	return b, nil
}

func (r *pgRepository) List(ctx context.Context, tenantID uuid.UUID, filter BranchFilter) ([]Branch, int, error) {
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	offset := (filter.Page - 1) * filter.Limit

	whereClauses := []string{"tenant_id = $1"}
	args := []any{tenantID}
	argIdx := 2

	if filter.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}

	if filter.IsActive != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *filter.IsActive)
		argIdx++
	}

	if strings.TrimSpace(filter.Search) != "" {
		searchPattern := "%" + strings.TrimSpace(filter.Search) + "%"
		whereClauses = append(whereClauses, fmt.Sprintf("(branch_code ILIKE $%d OR name ILIKE $%d OR city ILIKE $%d)", argIdx, argIdx, argIdx))
		args = append(args, searchPattern)
		argIdx++
	}

	// Geographic proximity query if coordinates provided
	isSpatial := filter.NearLat != nil && filter.NearLng != nil
	radiusMeters := 50000.0 // default 50km
	if filter.RadiusKM != nil && *filter.RadiusKM > 0 {
		radiusMeters = *filter.RadiusKM * 1000.0
	}

	var spatialSelect string
	var orderBy string

	if isSpatial {
		whereClauses = append(whereClauses, fmt.Sprintf(
			"ST_DWithin(location::geography, ST_SetSRID(ST_MakePoint($%d, $%d), 4326)::geography, $%d)",
			argIdx, argIdx+1, argIdx+2,
		))
		spatialSelect = fmt.Sprintf(
			", (ST_Distance(location::geography, ST_SetSRID(ST_MakePoint($%d, $%d), 4326)::geography) / 1000.0) AS distance_km",
			argIdx, argIdx+1,
		)
		orderBy = "distance_km ASC"
		args = append(args, *filter.NearLng, *filter.NearLat, radiusMeters)
		argIdx += 3
	} else {
		orderBy = "created_at DESC"
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM branches WHERE %s", whereSQL)
	var total int
	// Count query doesn't need distance args if not in where clause
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count branches: %w", err)
	}

	// Select items
	selectQuery := fmt.Sprintf(`
		SELECT 
			id, tenant_id, branch_code, name, address, city, state, postal_code, country,
			latitude, longitude, coverage_radius_km, status, is_active, created_at, updated_at
			%s
		FROM branches
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, spatialSelect, whereSQL, orderBy, argIdx, argIdx+1)

	args = append(args, filter.Limit, offset)

	rows, err := r.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query branches list: %w", err)
	}
	defer rows.Close()

	var branchesList []Branch
	for rows.Next() {
		var b Branch
		var state, postalCode *string
		var dist *float64

		var scanDest []any
		scanDest = append(scanDest,
			&b.ID,
			&b.TenantID,
			&b.BranchCode,
			&b.Name,
			&b.Address,
			&b.City,
			&state,
			&postalCode,
			&b.Country,
			&b.Latitude,
			&b.Longitude,
			&b.CoverageRadiusKM,
			&b.Status,
			&b.IsActive,
			&b.CreatedAt,
			&b.UpdatedAt,
		)
		if isSpatial {
			scanDest = append(scanDest, &dist)
		}

		if err := rows.Scan(scanDest...); err != nil {
			return nil, 0, fmt.Errorf("failed to scan branch row: %w", err)
		}

		if state != nil {
			b.State = *state
		}
		if postalCode != nil {
			b.PostalCode = *postalCode
		}
		if dist != nil {
			b.DistanceKM = dist
		}

		branchesList = append(branchesList, b)
	}

	if branchesList == nil {
		branchesList = []Branch{}
	}

	return branchesList, total, nil
}

func (r *pgRepository) Update(ctx context.Context, b *Branch) error {
	query := `
		UPDATE branches
		SET 
			name = $1,
			address = $2,
			city = $3,
			state = $4,
			postal_code = $5,
			country = $6,
			latitude = $7,
			longitude = $8,
			location = ST_SetSRID(ST_MakePoint($8, $7), 4326),
			coverage_radius_km = $9,
			status = $10,
			updated_at = NOW()
		WHERE id = $11 AND tenant_id = $12
		RETURNING updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		b.Name,
		b.Address,
		b.City,
		b.State,
		b.PostalCode,
		b.Country,
		b.Latitude,
		b.Longitude,
		b.CoverageRadiusKM,
		b.Status,
		b.ID,
		b.TenantID,
	).Scan(&b.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrBranchNotFound
		}
		return fmt.Errorf("failed to update branch: %w", err)
	}

	return nil
}

func (r *pgRepository) Deactivate(ctx context.Context, tenantID, branchID uuid.UUID) error {
	query := `
		UPDATE branches
		SET is_active = FALSE, status = 'INACTIVE', updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
	`

	cmdTag, err := r.pool.Exec(ctx, query, branchID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to deactivate branch: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrBranchNotFound
	}

	return nil
}
