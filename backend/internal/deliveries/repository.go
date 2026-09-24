package deliveries

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

type Repository interface {
	CreateDeliveryTask(ctx context.Context, t *DeliveryTask) error
	GetDeliveryTaskByID(ctx context.Context, tenantID, taskID uuid.UUID) (*DeliveryTask, error)
	ListDeliveryTasks(ctx context.Context, tenantID uuid.UUID, filter DeliveryTaskFilter) ([]DeliveryTask, int, error)
	UpdateDeliveryTaskStatus(ctx context.Context, tenantID, taskID uuid.UUID, status string, notes *string, failureReason *string) error
	RecordDeliveryAttempt(ctx context.Context, attempt *DeliveryAttempt) error
	GetDeliveryAttempts(ctx context.Context, taskID uuid.UUID) ([]DeliveryAttempt, error)
	SubmitDeliveryProof(ctx context.Context, tenantID uuid.UUID, proof *DeliveryProof) error
}

type pgRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) CreateDeliveryTask(ctx context.Context, t *DeliveryTask) error {
	query := `
		INSERT INTO delivery_tasks (
			tenant_id, parcel_id, branch_id, assigned_driver_id, vehicle_id,
			status, priority, assigned_at, notes
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, NOW(), $8
		)
		RETURNING id, assigned_at, created_at, updated_at;
	`

	if t.Status == "" {
		t.Status = StatusAssigned
	}
	if t.Priority == "" {
		t.Priority = PriorityStandard
	}

	err := r.pool.QueryRow(
		ctx,
		query,
		t.TenantID,
		t.ParcelID,
		t.BranchID,
		t.AssignedDriverID,
		t.VehicleID,
		t.Status,
		t.Priority,
		t.Notes,
	).Scan(&t.ID, &t.AssignedAt, &t.CreatedAt, &t.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "idx_active_parcel_delivery_task") {
				return ErrParcelAlreadyAssigned
			}
		}
		return fmt.Errorf("failed to insert delivery task: %w", err)
	}

	return nil
}

func (r *pgRepository) GetDeliveryTaskByID(ctx context.Context, tenantID, taskID uuid.UUID) (*DeliveryTask, error) {
	query := `
		SELECT 
			dt.id, dt.tenant_id, dt.parcel_id, dt.branch_id, dt.assigned_driver_id, dt.vehicle_id,
			dt.status, dt.priority, dt.assigned_at, dt.started_at, dt.completed_at, dt.failed_at,
			dt.failure_reason, dt.rescheduled_for, dt.notes, dt.created_at, dt.updated_at,
			p.tracking_number, p.sender_name, p.receiver_name, p.receiver_phone, p.receiver_address,
			p.status AS parcel_status, p.weight_kg,
			CONCAT(e.first_name, ' ', e.last_name) AS driver_name, e.employee_code AS driver_code,
			v.registration_number AS vehicle_registration_number,
			b.name AS branch_name,
			(SELECT COUNT(*) FROM delivery_attempts WHERE delivery_task_id = dt.id) AS attempts_count
		FROM delivery_tasks dt
		JOIN parcels p ON dt.parcel_id = p.id
		JOIN employees e ON dt.assigned_driver_id = e.id
		JOIN branches b ON dt.branch_id = b.id
		LEFT JOIN vehicles v ON dt.vehicle_id = v.id
		WHERE dt.id = $1 AND dt.tenant_id = $2;
	`

	dt := &DeliveryTask{}
	err := r.pool.QueryRow(ctx, query, taskID, tenantID).Scan(
		&dt.ID, &dt.TenantID, &dt.ParcelID, &dt.BranchID, &dt.AssignedDriverID, &dt.VehicleID,
		&dt.Status, &dt.Priority, &dt.AssignedAt, &dt.StartedAt, &dt.CompletedAt, &dt.FailedAt,
		&dt.FailureReason, &dt.RescheduledFor, &dt.Notes, &dt.CreatedAt, &dt.UpdatedAt,
		&dt.TrackingNumber, &dt.SenderName, &dt.ReceiverName, &dt.ReceiverPhone, &dt.ReceiverAddress,
		&dt.ParcelStatus, &dt.WeightKG,
		&dt.DriverName, &dt.DriverCode,
		&dt.VehicleRegNum,
		&dt.BranchName,
		&dt.AttemptsCount,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDeliveryTaskNotFound
		}
		return nil, fmt.Errorf("failed to get delivery task by id: %w", err)
	}

	return dt, nil
}

func (r *pgRepository) ListDeliveryTasks(ctx context.Context, tenantID uuid.UUID, filter DeliveryTaskFilter) ([]DeliveryTask, int, error) {
	baseQuery := `
		FROM delivery_tasks dt
		JOIN parcels p ON dt.parcel_id = p.id
		JOIN employees e ON dt.assigned_driver_id = e.id
		JOIN branches b ON dt.branch_id = b.id
		LEFT JOIN vehicles v ON dt.vehicle_id = v.id
		WHERE dt.tenant_id = $1
	`

	args := []any{tenantID}
	argIdx := 2

	if filter.Status != nil && *filter.Status != "" {
		baseQuery += fmt.Sprintf(" AND dt.status = $%d", argIdx)
		args = append(args, *filter.Status)
		argIdx++
	}
	if filter.AssignedDriverID != nil {
		baseQuery += fmt.Sprintf(" AND dt.assigned_driver_id = $%d", argIdx)
		args = append(args, *filter.AssignedDriverID)
		argIdx++
	}
	if filter.BranchID != nil {
		baseQuery += fmt.Sprintf(" AND dt.branch_id = $%d", argIdx)
		args = append(args, *filter.BranchID)
		argIdx++
	}
	if filter.ParcelID != nil {
		baseQuery += fmt.Sprintf(" AND dt.parcel_id = $%d", argIdx)
		args = append(args, *filter.ParcelID)
		argIdx++
	}
	if filter.Priority != nil && *filter.Priority != "" {
		baseQuery += fmt.Sprintf(" AND dt.priority = $%d", argIdx)
		args = append(args, *filter.Priority)
		argIdx++
	}

	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count delivery tasks: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	selectQuery := `
		SELECT 
			dt.id, dt.tenant_id, dt.parcel_id, dt.branch_id, dt.assigned_driver_id, dt.vehicle_id,
			dt.status, dt.priority, dt.assigned_at, dt.started_at, dt.completed_at, dt.failed_at,
			dt.failure_reason, dt.rescheduled_for, dt.notes, dt.created_at, dt.updated_at,
			p.tracking_number, p.sender_name, p.receiver_name, p.receiver_phone, p.receiver_address,
			p.status AS parcel_status, p.weight_kg,
			CONCAT(e.first_name, ' ', e.last_name) AS driver_name, e.employee_code AS driver_code,
			v.registration_number AS vehicle_registration_number,
			b.name AS branch_name,
			(SELECT COUNT(*) FROM delivery_attempts WHERE delivery_task_id = dt.id) AS attempts_count
	` + baseQuery + fmt.Sprintf(" ORDER BY dt.created_at DESC LIMIT $%d OFFSET $%d;", argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list delivery tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]DeliveryTask, 0)
	for rows.Next() {
		var dt DeliveryTask
		err := rows.Scan(
			&dt.ID, &dt.TenantID, &dt.ParcelID, &dt.BranchID, &dt.AssignedDriverID, &dt.VehicleID,
			&dt.Status, &dt.Priority, &dt.AssignedAt, &dt.StartedAt, &dt.CompletedAt, &dt.FailedAt,
			&dt.FailureReason, &dt.RescheduledFor, &dt.Notes, &dt.CreatedAt, &dt.UpdatedAt,
			&dt.TrackingNumber, &dt.SenderName, &dt.ReceiverName, &dt.ReceiverPhone, &dt.ReceiverAddress,
			&dt.ParcelStatus, &dt.WeightKG,
			&dt.DriverName, &dt.DriverCode,
			&dt.VehicleRegNum,
			&dt.BranchName,
			&dt.AttemptsCount,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan delivery task row: %w", err)
		}
		tasks = append(tasks, dt)
	}

	return tasks, total, nil
}

func (r *pgRepository) UpdateDeliveryTaskStatus(ctx context.Context, tenantID, taskID uuid.UUID, status string, notes *string, failureReason *string) error {
	var startedAt *time.Time
	var completedAt *time.Time
	var failedAt *time.Time

	now := time.Now().UTC()
	if status == StatusInProgress {
		startedAt = &now
	} else if status == StatusCompleted {
		completedAt = &now
	} else if status == StatusFailed {
		failedAt = &now
	}

	query := `
		UPDATE delivery_tasks SET
			status = $1,
			notes = COALESCE($2, notes),
			failure_reason = COALESCE($3, failure_reason),
			started_at = COALESCE($4, started_at),
			completed_at = COALESCE($5, completed_at),
			failed_at = COALESCE($6, failed_at),
			updated_at = NOW()
		WHERE id = $7 AND tenant_id = $8;
	`

	cmd, err := r.pool.Exec(ctx, query, status, notes, failureReason, startedAt, completedAt, failedAt, taskID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update delivery task status: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrDeliveryTaskNotFound
	}

	return nil
}

func (r *pgRepository) RecordDeliveryAttempt(ctx context.Context, attempt *DeliveryAttempt) error {
	var nextAttemptNum int
	countQuery := `SELECT COALESCE(MAX(attempt_number), 0) + 1 FROM delivery_attempts WHERE delivery_task_id = $1;`
	if err := r.pool.QueryRow(ctx, countQuery, attempt.DeliveryTaskID).Scan(&nextAttemptNum); err != nil {
		return fmt.Errorf("failed to calculate next attempt number: %w", err)
	}
	attempt.AttemptNumber = nextAttemptNum

	insertQuery := `
		INSERT INTO delivery_attempts (
			delivery_task_id, attempt_number, attempted_at, outcome, notes, latitude, longitude
		) VALUES ($1, $2, NOW(), $3, $4, $5, $6)
		RETURNING id, attempted_at;
	`

	return r.pool.QueryRow(
		ctx,
		insertQuery,
		attempt.DeliveryTaskID,
		attempt.AttemptNumber,
		attempt.Outcome,
		attempt.Notes,
		attempt.Latitude,
		attempt.Longitude,
	).Scan(&attempt.ID, &attempt.AttemptedAt)
}

func (r *pgRepository) GetDeliveryAttempts(ctx context.Context, taskID uuid.UUID) ([]DeliveryAttempt, error) {
	query := `
		SELECT id, delivery_task_id, attempt_number, attempted_at, outcome, notes, latitude, longitude
		FROM delivery_attempts
		WHERE delivery_task_id = $1
		ORDER BY attempt_number ASC;
	`

	rows, err := r.pool.Query(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query delivery attempts: %w", err)
	}
	defer rows.Close()

	attempts := make([]DeliveryAttempt, 0)
	for rows.Next() {
		var a DeliveryAttempt
		err := rows.Scan(
			&a.ID, &a.DeliveryTaskID, &a.AttemptNumber, &a.AttemptedAt,
			&a.Outcome, &a.Notes, &a.Latitude, &a.Longitude,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attempt row: %w", err)
		}
		attempts = append(attempts, a)
	}

	return attempts, nil
}

func (r *pgRepository) SubmitDeliveryProof(ctx context.Context, tenantID uuid.UUID, proof *DeliveryProof) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Lock and verify delivery task
	var taskStatus string
	var parcelID uuid.UUID
	var branchID uuid.UUID
	var driverID uuid.UUID

	lockQuery := `
		SELECT status, parcel_id, branch_id, assigned_driver_id
		FROM delivery_tasks
		WHERE id = $1 AND tenant_id = $2
		FOR UPDATE;
	`
	err = tx.QueryRow(ctx, lockQuery, proof.DeliveryTaskID, tenantID).Scan(&taskStatus, &parcelID, &branchID, &driverID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrDeliveryTaskNotFound
		}
		return fmt.Errorf("failed to lock delivery task: %w", err)
	}

	if taskStatus == StatusCompleted {
		return ErrDeliveryAlreadyCompleted
	}

	proof.ParcelID = parcelID

	// 2. Insert proof of delivery
	proofQuery := `
		INSERT INTO delivery_proofs (
			delivery_task_id, parcel_id, proof_type, recipient_name, recipient_relationship,
			otp_code, signature_data, photo_url, notes, latitude, longitude, verified_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10, $11, NOW()
		) RETURNING id, verified_at;
	`
	err = tx.QueryRow(
		ctx,
		proofQuery,
		proof.DeliveryTaskID,
		proof.ParcelID,
		proof.ProofType,
		proof.RecipientName,
		proof.RecipientRelationship,
		proof.OTPCode,
		proof.SignatureData,
		proof.PhotoURL,
		proof.Notes,
		proof.Latitude,
		proof.Longitude,
	).Scan(&proof.ID, &proof.VerifiedAt)

	if err != nil {
		return fmt.Errorf("failed to insert delivery proof: %w", err)
	}

	// 3. Mark delivery task as completed
	updateTaskQuery := `
		UPDATE delivery_tasks SET
			status = $1,
			completed_at = NOW(),
			updated_at = NOW()
		WHERE id = $2;
	`
	if _, err := tx.Exec(ctx, updateTaskQuery, StatusCompleted, proof.DeliveryTaskID); err != nil {
		return fmt.Errorf("failed to complete delivery task: %w", err)
	}

	// 4. Mark parcel as DELIVERED
	updateParcelQuery := `
		UPDATE parcels SET
			status = 'DELIVERED',
			updated_at = NOW()
		WHERE id = $1;
	`
	if _, err := tx.Exec(ctx, updateParcelQuery, parcelID); err != nil {
		return fmt.Errorf("failed to mark parcel delivered: %w", err)
	}

	// 5. Append to parcel status history
	historyQuery := `
		INSERT INTO parcel_status_history (
			tenant_id, parcel_id, from_status, to_status, branch_id, actor_role, notes, created_at
		) VALUES ($1, $2, 'OUT_FOR_DELIVERY', 'DELIVERED', $3, 'DRIVER', 'Package delivered with verified proof', NOW());
	`
	if _, err := tx.Exec(ctx, historyQuery, tenantID, parcelID, branchID); err != nil {
		return fmt.Errorf("failed to insert parcel history: %w", err)
	}

	return tx.Commit(ctx)
}
