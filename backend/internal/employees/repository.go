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
	BeginTx(ctx context.Context) (pgx.Tx, error)
	GenerateEmployeeCode(ctx context.Context, tenantID uuid.UUID) (string, error)
	GenerateEmployeeCodeTx(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID) (string, error)
	Create(ctx context.Context, emp *Employee) error
	CreateTx(ctx context.Context, tx pgx.Tx, emp *Employee) error
	GetByID(ctx context.Context, tenantID, employeeID uuid.UUID) (*Employee, error)
	GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*Employee, error)
	GetByUserID(ctx context.Context, tenantID, userID uuid.UUID) (*Employee, error)
	List(ctx context.Context, tenantID uuid.UUID, filter EmployeeFilter) ([]Employee, int, error)
	Update(ctx context.Context, emp *Employee) error
	UpdateStatus(ctx context.Context, tenantID, employeeID uuid.UUID, req UpdateEmployeeStatusRequest) error
	Deactivate(ctx context.Context, tenantID, employeeID uuid.UUID) error
	ListAvailableDrivers(ctx context.Context, tenantID uuid.UUID, branchID *uuid.UUID) ([]Employee, error)
	GetAccountStatus(ctx context.Context, tenantID, employeeID uuid.UUID) (*EmployeeAccountStatusResponse, error)
	GetActiveVehicleByDriverID(ctx context.Context, tenantID, driverID uuid.UUID) (*AssignedVehicleInfo, error)
}

type pgRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

type dbExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func (r *pgRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

// GenerateEmployeeCode safely generates a sequential, collision-free code per tenant (e.g. EMP-0001)
func (r *pgRepository) GenerateEmployeeCode(ctx context.Context, tenantID uuid.UUID) (string, error) {
	return generateEmployeeCode(ctx, r.pool, tenantID)
}

// GenerateEmployeeCodeTx safely generates a code within an active database transaction
func (r *pgRepository) GenerateEmployeeCodeTx(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID) (string, error) {
	return generateEmployeeCode(ctx, tx, tenantID)
}

func generateEmployeeCode(ctx context.Context, exec dbExecutor, tenantID uuid.UUID) (string, error) {
	query := `
		INSERT INTO tenant_employee_sequences (tenant_id, last_number, updated_at)
		VALUES ($1, 1, NOW())
		ON CONFLICT (tenant_id)
		DO UPDATE SET last_number = tenant_employee_sequences.last_number + 1, updated_at = NOW()
		RETURNING last_number;
	`
	var nextNum int
	if err := exec.QueryRow(ctx, query, tenantID).Scan(&nextNum); err != nil {
		return "", fmt.Errorf("failed to generate employee sequence: %w", err)
	}

	return fmt.Sprintf("EMP-%04d", nextNum), nil
}

func (r *pgRepository) Create(ctx context.Context, emp *Employee) error {
	return insertEmployee(ctx, r.pool, emp)
}

func (r *pgRepository) CreateTx(ctx context.Context, tx pgx.Tx, emp *Employee) error {
	return insertEmployee(ctx, tx, emp)
}

func insertEmployee(ctx context.Context, exec dbExecutor, emp *Employee) error {
	query := `
		INSERT INTO employees (
			tenant_id, user_id, branch_id, employee_code, first_name, last_name,
			email, phone, designation, employment_type, operational_role,
			license_number, status, availability_status, verification_status,
			joining_date, is_active
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11,
			$12, $13, $14, $15,
			COALESCE($16, CURRENT_DATE), $17
		)
		RETURNING id, joining_date, created_at, updated_at
	`

	if emp.AvailabilityStatus == "" {
		emp.AvailabilityStatus = AvailabilityStatusAvailable
	}
	if emp.VerificationStatus == "" {
		emp.VerificationStatus = VerificationStatusVerified
	}

	err := exec.QueryRow(
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
		emp.AvailabilityStatus,
		emp.VerificationStatus,
		emp.JoiningDate,
		emp.IsActive,
	).Scan(
		&emp.ID,
		&emp.JoiningDate,
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
		       e.license_number, e.status, e.availability_status, e.verification_status,
		       e.joining_date, e.is_active, e.created_at, e.updated_at, e.deleted_at,
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
		&emp.AvailabilityStatus,
		&emp.VerificationStatus,
		&emp.JoiningDate,
		&emp.IsActive,
		&emp.CreatedAt,
		&emp.UpdatedAt,
		&emp.DeletedAt,
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
		       e.license_number, e.status, e.availability_status, e.verification_status,
		       e.joining_date, e.is_active, e.created_at, e.updated_at, e.deleted_at,
		       b.name AS branch_name, b.branch_code AS branch_code
		FROM employees e
		LEFT JOIN branches b ON e.branch_id = b.id
		WHERE e.employee_code = $1 AND e.tenant_id = $2 AND e.deleted_at IS NULL
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
		&emp.AvailabilityStatus,
		&emp.VerificationStatus,
		&emp.JoiningDate,
		&emp.IsActive,
		&emp.CreatedAt,
		&emp.UpdatedAt,
		&emp.DeletedAt,
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
		       e.license_number, e.status, e.availability_status, e.verification_status,
		       e.joining_date, e.is_active, e.created_at, e.updated_at, e.deleted_at,
		       b.name AS branch_name, b.branch_code AS branch_code
		FROM employees e
		LEFT JOIN branches b ON e.branch_id = b.id
		WHERE e.user_id = $1 AND e.tenant_id = $2 AND e.deleted_at IS NULL
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
		&emp.AvailabilityStatus,
		&emp.VerificationStatus,
		&emp.JoiningDate,
		&emp.IsActive,
		&emp.CreatedAt,
		&emp.UpdatedAt,
		&emp.DeletedAt,
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
	whereClauses := []string{"e.tenant_id = $1", "e.deleted_at IS NULL"}
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

	if filter.AvailabilityStatus != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("e.availability_status = $%d", argIdx))
		args = append(args, filter.AvailabilityStatus)
		argIdx++
	}

	if filter.VerificationStatus != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("e.verification_status = $%d", argIdx))
		args = append(args, filter.VerificationStatus)
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
		       e.license_number, e.status, e.availability_status, e.verification_status,
		       e.joining_date, e.is_active, e.created_at, e.updated_at, e.deleted_at,
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
			&emp.AvailabilityStatus,
			&emp.VerificationStatus,
			&emp.JoiningDate,
			&emp.IsActive,
			&emp.CreatedAt,
			&emp.UpdatedAt,
			&emp.DeletedAt,
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
			availability_status = $11,
			verification_status = $12,
			is_active = $13,
			updated_at = NOW()
		WHERE id = $14 AND tenant_id = $15 AND deleted_at IS NULL
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
		emp.AvailabilityStatus,
		emp.VerificationStatus,
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

func (r *pgRepository) UpdateStatus(ctx context.Context, tenantID, employeeID uuid.UUID, req UpdateEmployeeStatusRequest) error {
	setClauses := []string{"updated_at = NOW()"}
	args := []interface{}{employeeID, tenantID}
	argIdx := 3

	if req.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *req.Status)
		argIdx++
		if *req.Status == StatusTerminated {
			setClauses = append(setClauses, "is_active = FALSE")
		} else {
			setClauses = append(setClauses, "is_active = TRUE")
		}
	}

	if req.AvailabilityStatus != nil {
		setClauses = append(setClauses, fmt.Sprintf("availability_status = $%d", argIdx))
		args = append(args, *req.AvailabilityStatus)
		argIdx++
	}

	if req.VerificationStatus != nil {
		setClauses = append(setClauses, fmt.Sprintf("verification_status = $%d", argIdx))
		args = append(args, *req.VerificationStatus)
		argIdx++
	}

	query := fmt.Sprintf(`
		UPDATE employees SET %s
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, strings.Join(setClauses, ", "))

	cmdTag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update employee status: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrEmployeeNotFound
	}

	return nil
}

func (r *pgRepository) Deactivate(ctx context.Context, tenantID, employeeID uuid.UUID) error {
	query := `
		UPDATE employees
		SET is_active = FALSE, status = 'TERMINATED', availability_status = 'UNAVAILABLE',
		    deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
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

func (r *pgRepository) ListAvailableDrivers(ctx context.Context, tenantID uuid.UUID, branchID *uuid.UUID) ([]Employee, error) {
	whereClauses := []string{
		"e.tenant_id = $1",
		"e.operational_role = 'DRIVER'",
		"e.status = 'ACTIVE'",
		"e.availability_status = 'AVAILABLE'",
		"e.is_active = TRUE",
		"e.deleted_at IS NULL",
	}
	args := []interface{}{tenantID}

	if branchID != nil {
		whereClauses = append(whereClauses, "e.branch_id = $2")
		args = append(args, *branchID)
	}

	query := fmt.Sprintf(`
		SELECT e.id, e.tenant_id, e.user_id, e.branch_id, e.employee_code, e.first_name, e.last_name,
		       e.email, e.phone, e.designation, e.employment_type, e.operational_role,
		       e.license_number, e.status, e.availability_status, e.verification_status,
		       e.joining_date, e.is_active, e.created_at, e.updated_at, e.deleted_at,
		       b.name AS branch_name, b.branch_code AS branch_code
		FROM employees e
		LEFT JOIN branches b ON e.branch_id = b.id
		WHERE %s
		ORDER BY e.first_name ASC, e.last_name ASC
	`, strings.Join(whereClauses, " AND "))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query available drivers: %w", err)
	}
	defer rows.Close()

	drivers := make([]Employee, 0)
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
			&emp.AvailabilityStatus,
			&emp.VerificationStatus,
			&emp.JoiningDate,
			&emp.IsActive,
			&emp.CreatedAt,
			&emp.UpdatedAt,
			&emp.DeletedAt,
			&emp.BranchName,
			&emp.BranchCode,
		); err != nil {
			return nil, fmt.Errorf("failed to scan driver row: %w", err)
		}
		drivers = append(drivers, emp)
	}

	return drivers, nil
}

func (r *pgRepository) GetAccountStatus(ctx context.Context, tenantID, employeeID uuid.UUID) (*EmployeeAccountStatusResponse, error) {
	query := `
		SELECT
			e.id, e.employee_code, e.first_name, e.last_name, e.operational_role, e.status,
			e.user_id, u.email, u.is_active, tm.role, e.branch_id, b.name
		FROM employees e
		LEFT JOIN users u ON e.user_id = u.id
		LEFT JOIN tenant_memberships tm ON e.user_id = tm.user_id AND tm.tenant_id = e.tenant_id
		LEFT JOIN branches b ON e.branch_id = b.id
		WHERE e.tenant_id = $1 AND e.id = $2 AND e.deleted_at IS NULL
	`
	var (
		resp       EmployeeAccountStatusResponse
		firstName  string
		lastName   string
		userID     *uuid.UUID
		userEmail  *string
		userActive *bool
		systemRole *string
		branchID   *uuid.UUID
		branchName *string
	)

	err := r.pool.QueryRow(ctx, query, tenantID, employeeID).Scan(
		&resp.EmployeeID,
		&resp.EmployeeCode,
		&firstName,
		&lastName,
		&resp.OperationalRole,
		&resp.Status,
		&userID,
		&userEmail,
		&userActive,
		&systemRole,
		&branchID,
		&branchName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEmployeeNotFound
		}
		return nil, fmt.Errorf("failed to query employee account status: %w", err)
	}

	resp.FullName = strings.TrimSpace(firstName + " " + lastName)
	resp.HasAccount = userID != nil
	resp.UserID = userID
	resp.UserEmail = userEmail
	resp.UserIsActive = userActive
	resp.SystemRole = systemRole
	resp.BranchID = branchID
	resp.BranchName = branchName

	return &resp, nil
}

func (r *pgRepository) GetActiveVehicleByDriverID(ctx context.Context, tenantID, driverID uuid.UUID) (*AssignedVehicleInfo, error) {
	query := `
		SELECT v.id, v.registration_number, v.vehicle_type, v.make_model, v.status
		FROM vehicle_assignments a
		JOIN vehicles v ON a.vehicle_id = v.id
		WHERE a.tenant_id = $1 AND a.driver_id = $2 AND a.status = 'ACTIVE' AND v.deleted_at IS NULL
		LIMIT 1
	`
	var info AssignedVehicleInfo
	err := r.pool.QueryRow(ctx, query, tenantID, driverID).Scan(
		&info.ID,
		&info.RegistrationNumber,
		&info.VehicleType,
		&info.MakeModel,
		&info.Status,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // No active vehicle assigned to driver
		}
		return nil, fmt.Errorf("failed to get active vehicle for driver: %w", err)
	}
	return &info, nil
}
