package employees

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
	Create(ctx context.Context, emp *Employee) error
	GetByID(ctx context.Context, tenantID, employeeID uuid.UUID) (*Employee, error)
	GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*Employee, error)
	GetByUserID(ctx context.Context, tenantID, userID uuid.UUID) (*Employee, error)
	List(ctx context.Context, tenantID uuid.UUID, filter EmployeeFilter) ([]Employee, int, error)
	Update(ctx context.Context, emp *Employee) error
	Deactivate(ctx context.Context, tenantID, employeeID uuid.UUID) error
}

type pgRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) Create(ctx context.Context, emp *Employee) error {
	query := `
		INSERT INTO employees (
			tenant_id, user_id, branch_id, employee_code, first_name, last_name,
			email, phone, designation, employment_type, operational_role,
			license_number, status, is_active
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11,
			$12, $13, $14
		)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		emp.TenantID,
		emp.UserID,
		emp.BranchID,
		emp.EmployeeCode,
		emp.FirstName,
		emp.LastName,
		emp.Email,
		emp.Phone,
		emp.Designation,
		emp.EmploymentType,
		emp.OperationalRole,
		emp.LicenseNumber,
		emp.Status,
		emp.IsActive,
	).Scan(
		&emp.ID,
		&emp.CreatedAt,
		&emp.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "uq_tenant_employee_code") {
				return ErrDuplicateEmployeeCode
			}
			if strings.Contains(pgErr.ConstraintName, "uq_tenant_user_id") {
				return ErrUserAlreadyLinked
			}
		}
		return fmt.Errorf("failed to insert employee: %w", err)
	}

	return nil
}

func (r *pgRepository) GetByID(ctx context.Context, tenantID, employeeID uuid.UUID) (*Employee, error) {
	query := `
		SELECT e.id, e.tenant_id, e.user_id, e.branch_id, e.employee_code, e.first_name, e.last_name,
		       e.email, e.phone, e.designation, e.employment_type, e.operational_role,
		       e.license_number, e.status, e.is_active, e.created_at, e.updated_at,
		       b.name AS branch_name, b.branch_code AS branch_code
		FROM employees e
		LEFT JOIN branches b ON e.branch_id = b.id
		WHERE e.id = $1 AND e.tenant_id = $2
	`

	emp := &Employee{}
	err := r.pool.QueryRow(ctx, query, employeeID, tenantID).Scan(
		&emp.ID,
		&emp.TenantID,
		&emp.UserID,
		&emp.BranchID,
		&emp.EmployeeCode,
		&emp.FirstName,
		&emp.LastName,
		&emp.Email,
		&emp.Phone,
		&emp.Designation,
		&emp.EmploymentType,
		&emp.OperationalRole,
		&emp.LicenseNumber,
		&emp.Status,
		&emp.IsActive,
		&emp.CreatedAt,
		&emp.UpdatedAt,
		&emp.BranchName,
		&emp.BranchCode,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEmployeeNotFound
		}
		return nil, fmt.Errorf("failed to query employee by id: %w", err)
	}

	return emp, nil
}

func (r *pgRepository) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*Employee, error) {
	query := `
		SELECT e.id, e.tenant_id, e.user_id, e.branch_id, e.employee_code, e.first_name, e.last_name,
		       e.email, e.phone, e.designation, e.employment_type, e.operational_role,
		       e.license_number, e.status, e.is_active, e.created_at, e.updated_at,
		       b.name AS branch_name, b.branch_code AS branch_code
		FROM employees e
		LEFT JOIN branches b ON e.branch_id = b.id
		WHERE e.employee_code = $1 AND e.tenant_id = $2
	`

	emp := &Employee{}
	err := r.pool.QueryRow(ctx, query, code, tenantID).Scan(
		&emp.ID,
		&emp.TenantID,
		&emp.UserID,
		&emp.BranchID,
		&emp.EmployeeCode,
		&emp.FirstName,
		&emp.LastName,
		&emp.Email,
		&emp.Phone,
		&emp.Designation,
		&emp.EmploymentType,
		&emp.OperationalRole,
		&emp.LicenseNumber,
		&emp.Status,
		&emp.IsActive,
		&emp.CreatedAt,
		&emp.UpdatedAt,
		&emp.BranchName,
		&emp.BranchCode,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEmployeeNotFound
		}
		return nil, fmt.Errorf("failed to query employee by code: %w", err)
	}

	return emp, nil
}

func (r *pgRepository) GetByUserID(ctx context.Context, tenantID, userID uuid.UUID) (*Employee, error) {
	query := `
		SELECT e.id, e.tenant_id, e.user_id, e.branch_id, e.employee_code, e.first_name, e.last_name,
		       e.email, e.phone, e.designation, e.employment_type, e.operational_role,
		       e.license_number, e.status, e.is_active, e.created_at, e.updated_at,
		       b.name AS branch_name, b.branch_code AS branch_code
		FROM employees e
		LEFT JOIN branches b ON e.branch_id = b.id
		WHERE e.user_id = $1 AND e.tenant_id = $2
	`

	emp := &Employee{}
	err := r.pool.QueryRow(ctx, query, userID, tenantID).Scan(
		&emp.ID,
		&emp.TenantID,
		&emp.UserID,
		&emp.BranchID,
		&emp.EmployeeCode,
		&emp.FirstName,
		&emp.LastName,
		&emp.Email,
		&emp.Phone,
		&emp.Designation,
		&emp.EmploymentType,
		&emp.OperationalRole,
		&emp.LicenseNumber,
		&emp.Status,
		&emp.IsActive,
		&emp.CreatedAt,
		&emp.UpdatedAt,
		&emp.BranchName,
		&emp.BranchCode,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEmployeeNotFound
		}
		return nil, fmt.Errorf("failed to query employee by user_id: %w", err)
	}

	return emp, nil
}

func (r *pgRepository) List(ctx context.Context, tenantID uuid.UUID, filter EmployeeFilter) ([]Employee, int, error) {
	whereClauses := []string{"e.tenant_id = $1"}
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.Search != "" {
		searchTerm := "%" + strings.ToLower(filter.Search) + "%"
		whereClauses = append(whereClauses, fmt.Sprintf("(LOWER(e.first_name) LIKE $%d OR LOWER(e.last_name) LIKE $%d OR LOWER(e.employee_code) LIKE $%d OR LOWER(e.designation) LIKE $%d)", argIdx, argIdx, argIdx, argIdx))
		args = append(args, searchTerm)
		argIdx++
	}

	if filter.OperationalRole != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("e.operational_role = $%d", argIdx))
		args = append(args, filter.OperationalRole)
		argIdx++
	}

	if filter.BranchID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("e.branch_id = $%d", argIdx))
		args = append(args, *filter.BranchID)
		argIdx++
	}

	if filter.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("e.status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM employees e WHERE %s", whereSQL)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count employees: %w", err)
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
		SELECT e.id, e.tenant_id, e.user_id, e.branch_id, e.employee_code, e.first_name, e.last_name,
		       e.email, e.phone, e.designation, e.employment_type, e.operational_role,
		       e.license_number, e.status, e.is_active, e.created_at, e.updated_at,
		       b.name AS branch_name, b.branch_code AS branch_code
		FROM employees e
		LEFT JOIN branches b ON e.branch_id = b.id
		WHERE %s
		ORDER BY e.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query employees list: %w", err)
	}
	defer rows.Close()

	employeesList := make([]Employee, 0)
	for rows.Next() {
		var emp Employee
		if err := rows.Scan(
			&emp.ID,
			&emp.TenantID,
			&emp.UserID,
			&emp.BranchID,
			&emp.EmployeeCode,
			&emp.FirstName,
			&emp.LastName,
			&emp.Email,
			&emp.Phone,
			&emp.Designation,
			&emp.EmploymentType,
			&emp.OperationalRole,
			&emp.LicenseNumber,
			&emp.Status,
			&emp.IsActive,
			&emp.CreatedAt,
			&emp.UpdatedAt,
			&emp.BranchName,
			&emp.BranchCode,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan employee row: %w", err)
		}
		employeesList = append(employeesList, emp)
	}

	return employeesList, total, nil
}

func (r *pgRepository) Update(ctx context.Context, emp *Employee) error {
	query := `
		UPDATE employees SET
			first_name = $1,
			last_name = $2,
			email = $3,
			phone = $4,
			designation = $5,
			employment_type = $6,
			operational_role = $7,
			license_number = $8,
			branch_id = $9,
			status = $10,
			is_active = $11,
			updated_at = NOW()
		WHERE id = $12 AND tenant_id = $13
		RETURNING updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		emp.FirstName,
		emp.LastName,
		emp.Email,
		emp.Phone,
		emp.Designation,
		emp.EmploymentType,
		emp.OperationalRole,
		emp.LicenseNumber,
		emp.BranchID,
		emp.Status,
		emp.IsActive,
		emp.ID,
		emp.TenantID,
	).Scan(&emp.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrEmployeeNotFound
		}
		return fmt.Errorf("failed to update employee: %w", err)
	}

	return nil
}

func (r *pgRepository) Deactivate(ctx context.Context, tenantID, employeeID uuid.UUID) error {
	query := `
		UPDATE employees
		SET is_active = FALSE, status = 'TERMINATED', updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
	`
	cmdTag, err := r.pool.Exec(ctx, query, employeeID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to deactivate employee: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrEmployeeNotFound
	}
	return nil
}
