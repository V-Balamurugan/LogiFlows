package vehicles

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

type Repository interface {
	CreateVehicle(ctx context.Context, v *Vehicle) error
	GetVehicleByID(ctx context.Context, tenantID, vehicleID uuid.UUID) (*Vehicle, error)
	GetVehicleByRegNum(ctx context.Context, tenantID uuid.UUID, regNum string) (*Vehicle, error)
	ListVehicles(ctx context.Context, tenantID uuid.UUID, filter VehicleFilter) ([]Vehicle, int, error)
	UpdateVehicle(ctx context.Context, v *Vehicle) error
	DeactivateVehicle(ctx context.Context, tenantID, vehicleID uuid.UUID) error

	CreateAssignment(ctx context.Context, a *VehicleAssignment) error
	GetActiveAssignmentByVehicle(ctx context.Context, tenantID, vehicleID uuid.UUID) (*VehicleAssignment, error)
	GetActiveAssignmentByDriver(ctx context.Context, tenantID, driverID uuid.UUID) (*VehicleAssignment, error)
	CompleteAssignment(ctx context.Context, tenantID, assignmentID uuid.UUID) error
	ListAssignments(ctx context.Context, tenantID uuid.UUID, vehicleID *uuid.UUID, driverID *uuid.UUID, limit, offset int) ([]VehicleAssignment, int, error)
}

type pgRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) CreateVehicle(ctx context.Context, v *Vehicle) error {
	query := `
		INSERT INTO vehicles (
			tenant_id, branch_id, registration_number, vehicle_type,
			make_model, year, max_weight_kg, max_volume_cbm, status, is_active
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8, $9, $10
		)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		v.TenantID,
		v.BranchID,
		v.RegistrationNumber,
		v.VehicleType,
		v.MakeModel,
		v.Year,
		v.MaxWeightKG,
		v.MaxVolumeCBM,
		v.Status,
		v.IsActive,
	).Scan(&v.ID, &v.CreatedAt, &v.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "uq_tenant_registration_number") {
				return ErrDuplicateRegistrationNumber
			}
		}
		return fmt.Errorf("failed to insert vehicle: %w", err)
	}

	return nil
}

func (r *pgRepository) GetVehicleByID(ctx context.Context, tenantID, vehicleID uuid.UUID) (*Vehicle, error) {
	query := `
		SELECT 
			v.id, v.tenant_id, v.branch_id, v.registration_number, v.vehicle_type,
			v.make_model, v.year, v.max_weight_kg, v.max_volume_cbm, v.status, v.is_active,
			v.created_at, v.updated_at,
			b.name AS branch_name, b.branch_code AS branch_code,
			e.id AS current_driver_id,
			CASE WHEN e.id IS NOT NULL THEN CONCAT(e.first_name, ' ', e.last_name) ELSE NULL END AS current_driver_name
		FROM vehicles v
		LEFT JOIN branches b ON v.branch_id = b.id
		LEFT JOIN vehicle_assignments va ON v.id = va.vehicle_id AND va.status = 'ACTIVE'
		LEFT JOIN employees e ON va.driver_id = e.id
		WHERE v.id = $1 AND v.tenant_id = $2
	`

	v := &Vehicle{}
	err := r.pool.QueryRow(ctx, query, vehicleID, tenantID).Scan(
		&v.ID,
		&v.TenantID,
		&v.BranchID,
		&v.RegistrationNumber,
		&v.VehicleType,
		&v.MakeModel,
		&v.Year,
		&v.MaxWeightKG,
		&v.MaxVolumeCBM,
		&v.Status,
		&v.IsActive,
		&v.CreatedAt,
		&v.UpdatedAt,
		&v.BranchName,
		&v.BranchCode,
		&v.CurrentDriverID,
		&v.CurrentDriverName,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrVehicleNotFound
		}
		return nil, fmt.Errorf("failed to query vehicle by id: %w", err)
	}

	return v, nil
}

func (r *pgRepository) GetVehicleByRegNum(ctx context.Context, tenantID uuid.UUID, regNum string) (*Vehicle, error) {
	query := `
		SELECT 
			v.id, v.tenant_id, v.branch_id, v.registration_number, v.vehicle_type,
			v.make_model, v.year, v.max_weight_kg, v.max_volume_cbm, v.status, v.is_active,
			v.created_at, v.updated_at,
			b.name AS branch_name, b.branch_code AS branch_code,
			e.id AS current_driver_id,
			CASE WHEN e.id IS NOT NULL THEN CONCAT(e.first_name, ' ', e.last_name) ELSE NULL END AS current_driver_name
		FROM vehicles v
		LEFT JOIN branches b ON v.branch_id = b.id
		LEFT JOIN vehicle_assignments va ON v.id = va.vehicle_id AND va.status = 'ACTIVE'
		LEFT JOIN employees e ON va.driver_id = e.id
		WHERE v.registration_number = $1 AND v.tenant_id = $2
	`

	v := &Vehicle{}
	err := r.pool.QueryRow(ctx, query, regNum, tenantID).Scan(
		&v.ID,
		&v.TenantID,
		&v.BranchID,
		&v.RegistrationNumber,
		&v.VehicleType,
		&v.MakeModel,
		&v.Year,
		&v.MaxWeightKG,
		&v.MaxVolumeCBM,
		&v.Status,
		&v.IsActive,
		&v.CreatedAt,
		&v.UpdatedAt,
		&v.BranchName,
		&v.BranchCode,
		&v.CurrentDriverID,
		&v.CurrentDriverName,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrVehicleNotFound
		}
		return nil, fmt.Errorf("failed to query vehicle by reg_num: %w", err)
	}

	return v, nil
}

func (r *pgRepository) ListVehicles(ctx context.Context, tenantID uuid.UUID, filter VehicleFilter) ([]Vehicle, int, error) {
	whereClauses := []string{"v.tenant_id = $1"}
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.Search != "" {
		searchTerm := "%" + strings.ToLower(filter.Search) + "%"
		whereClauses = append(whereClauses, fmt.Sprintf("(LOWER(v.registration_number) LIKE $%d OR LOWER(v.make_model) LIKE $%d)", argIdx, argIdx))
		args = append(args, searchTerm)
		argIdx++
	}

	if filter.VehicleType != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("v.vehicle_type = $%d", argIdx))
		args = append(args, filter.VehicleType)
		argIdx++
	}

	if filter.BranchID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("v.branch_id = $%d", argIdx))
		args = append(args, *filter.BranchID)
		argIdx++
	}

	if filter.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("v.status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM vehicles v WHERE %s", whereSQL)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count vehicles: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT 
			v.id, v.tenant_id, v.branch_id, v.registration_number, v.vehicle_type,
			v.make_model, v.year, v.max_weight_kg, v.max_volume_cbm, v.status, v.is_active,
			v.created_at, v.updated_at,
			b.name AS branch_name, b.branch_code AS branch_code,
			e.id AS current_driver_id,
			CASE WHEN e.id IS NOT NULL THEN CONCAT(e.first_name, ' ', e.last_name) ELSE NULL END AS current_driver_name
		FROM vehicles v
		LEFT JOIN branches b ON v.branch_id = b.id
		LEFT JOIN vehicle_assignments va ON v.id = va.vehicle_id AND va.status = 'ACTIVE'
		LEFT JOIN employees e ON va.driver_id = e.id
		WHERE %s
		ORDER BY v.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query vehicles list: %w", err)
	}
	defer rows.Close()

	list := make([]Vehicle, 0)
	for rows.Next() {
		var v Vehicle
		if err := rows.Scan(
			&v.ID,
			&v.TenantID,
			&v.BranchID,
			&v.RegistrationNumber,
			&v.VehicleType,
			&v.MakeModel,
			&v.Year,
			&v.MaxWeightKG,
			&v.MaxVolumeCBM,
			&v.Status,
			&v.IsActive,
			&v.CreatedAt,
			&v.UpdatedAt,
			&v.BranchName,
			&v.BranchCode,
			&v.CurrentDriverID,
			&v.CurrentDriverName,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan vehicle row: %w", err)
		}
		list = append(list, v)
	}

	return list, total, nil
}

func (r *pgRepository) UpdateVehicle(ctx context.Context, v *Vehicle) error {
	query := `
		UPDATE vehicles SET
			vehicle_type = $1,
			make_model = $2,
			year = $3,
			max_weight_kg = $4,
			max_volume_cbm = $5,
			branch_id = $6,
			status = $7,
			is_active = $8,
			updated_at = NOW()
		WHERE id = $9 AND tenant_id = $10
		RETURNING updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		v.VehicleType,
		v.MakeModel,
		v.Year,
		v.MaxWeightKG,
		v.MaxVolumeCBM,
		v.BranchID,
		v.Status,
		v.IsActive,
		v.ID,
		v.TenantID,
	).Scan(&v.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrVehicleNotFound
		}
		return fmt.Errorf("failed to update vehicle: %w", err)
	}

	return nil
}

func (r *pgRepository) DeactivateVehicle(ctx context.Context, tenantID, vehicleID uuid.UUID) error {
	query := `
		UPDATE vehicles
		SET is_active = FALSE, status = 'DECOMMISSIONED', updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
	`
	cmdTag, err := r.pool.Exec(ctx, query, vehicleID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to deactivate vehicle: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrVehicleNotFound
	}
	return nil
}

func (r *pgRepository) CreateAssignment(ctx context.Context, a *VehicleAssignment) error {
	query := `
		INSERT INTO vehicle_assignments (
			tenant_id, vehicle_id, driver_id, assigned_by, assigned_at, status, notes
		) VALUES (
			$1, $2, $3, $4, NOW(), 'ACTIVE', $5
		)
		RETURNING id, assigned_at, created_at, updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		a.TenantID,
		a.VehicleID,
		a.DriverID,
		a.AssignedBy,
		a.Notes,
	).Scan(&a.ID, &a.AssignedAt, &a.CreatedAt, &a.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "uq_active_vehicle_assignment") {
				return ErrVehicleAlreadyAssigned
			}
			if strings.Contains(pgErr.ConstraintName, "uq_active_driver_assignment") {
				return ErrDriverAlreadyAssigned
			}
		}
		return fmt.Errorf("failed to insert vehicle assignment: %w", err)
	}

	return nil
}

func (r *pgRepository) GetActiveAssignmentByVehicle(ctx context.Context, tenantID, vehicleID uuid.UUID) (*VehicleAssignment, error) {
	query := `
		SELECT 
			va.id, va.tenant_id, va.vehicle_id, va.driver_id, va.assigned_by,
			va.assigned_at, va.unassigned_at, va.status, va.notes, va.created_at, va.updated_at,
			v.registration_number,
			CONCAT(e.first_name, ' ', e.last_name) AS driver_name
		FROM vehicle_assignments va
		JOIN vehicles v ON va.vehicle_id = v.id
		JOIN employees e ON va.driver_id = e.id
		WHERE va.tenant_id = $1 AND va.vehicle_id = $2 AND va.status = 'ACTIVE'
	`

	a := &VehicleAssignment{}
	err := r.pool.QueryRow(ctx, query, tenantID, vehicleID).Scan(
		&a.ID,
		&a.TenantID,
		&a.VehicleID,
		&a.DriverID,
		&a.AssignedBy,
		&a.AssignedAt,
		&a.UnassignedAt,
		&a.Status,
		&a.Notes,
		&a.CreatedAt,
		&a.UpdatedAt,
		&a.RegistrationNumber,
		&a.DriverName,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAssignmentNotFound
		}
		return nil, fmt.Errorf("failed to query active assignment: %w", err)
	}

	return a, nil
}

func (r *pgRepository) GetActiveAssignmentByDriver(ctx context.Context, tenantID, driverID uuid.UUID) (*VehicleAssignment, error) {
	query := `
		SELECT 
			va.id, va.tenant_id, va.vehicle_id, va.driver_id, va.assigned_by,
			va.assigned_at, va.unassigned_at, va.status, va.notes, va.created_at, va.updated_at,
			v.registration_number,
			CONCAT(e.first_name, ' ', e.last_name) AS driver_name
		FROM vehicle_assignments va
		JOIN vehicles v ON va.vehicle_id = v.id
		JOIN employees e ON va.driver_id = e.id
		WHERE va.tenant_id = $1 AND va.driver_id = $2 AND va.status = 'ACTIVE'
	`

	a := &VehicleAssignment{}
	err := r.pool.QueryRow(ctx, query, tenantID, driverID).Scan(
		&a.ID,
		&a.TenantID,
		&a.VehicleID,
		&a.DriverID,
		&a.AssignedBy,
		&a.AssignedAt,
		&a.UnassignedAt,
		&a.Status,
		&a.Notes,
		&a.CreatedAt,
		&a.UpdatedAt,
		&a.RegistrationNumber,
		&a.DriverName,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAssignmentNotFound
		}
		return nil, fmt.Errorf("failed to query active assignment by driver: %w", err)
	}

	return a, nil
}

func (r *pgRepository) CompleteAssignment(ctx context.Context, tenantID, assignmentID uuid.UUID) error {
	query := `
		UPDATE vehicle_assignments
		SET status = 'COMPLETED', unassigned_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2 AND status = 'ACTIVE'
	`
	cmdTag, err := r.pool.Exec(ctx, query, assignmentID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to complete vehicle assignment: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrAssignmentNotFound
	}
	return nil
}

func (r *pgRepository) ListAssignments(ctx context.Context, tenantID uuid.UUID, vehicleID *uuid.UUID, driverID *uuid.UUID, limit, offset int) ([]VehicleAssignment, int, error) {
	whereClauses := []string{"va.tenant_id = $1"}
	args := []interface{}{tenantID}
	argIdx := 2

	if vehicleID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("va.vehicle_id = $%d", argIdx))
		args = append(args, *vehicleID)
		argIdx++
	}

	if driverID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("va.driver_id = $%d", argIdx))
		args = append(args, *driverID)
		argIdx++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM vehicle_assignments va WHERE %s", whereSQL)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count vehicle assignments: %w", err)
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT 
			va.id, va.tenant_id, va.vehicle_id, va.driver_id, va.assigned_by,
			va.assigned_at, va.unassigned_at, va.status, va.notes, va.created_at, va.updated_at,
			v.registration_number,
			CONCAT(e.first_name, ' ', e.last_name) AS driver_name
		FROM vehicle_assignments va
		JOIN vehicles v ON va.vehicle_id = v.id
		JOIN employees e ON va.driver_id = e.id
		WHERE %s
		ORDER BY va.assigned_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query vehicle assignments list: %w", err)
	}
	defer rows.Close()

	list := make([]VehicleAssignment, 0)
	for rows.Next() {
		var a VehicleAssignment
		if err := rows.Scan(
			&a.ID,
			&a.TenantID,
			&a.VehicleID,
			&a.DriverID,
			&a.AssignedBy,
			&a.AssignedAt,
			&a.UnassignedAt,
			&a.Status,
			&a.Notes,
			&a.CreatedAt,
			&a.UpdatedAt,
			&a.RegistrationNumber,
			&a.DriverName,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan vehicle assignment row: %w", err)
		}
		list = append(list, a)
	}

	return list, total, nil
}
