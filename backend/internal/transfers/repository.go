package transfers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	NextTransferNumber(ctx context.Context, tenantID uuid.UUID) (string, error)
	CreateTransfer(ctx context.Context, t *BranchTransfer, parcelIDs []uuid.UUID) error
	GetTransferByID(ctx context.Context, tenantID, transferID uuid.UUID) (*BranchTransfer, []BranchTransferParcel, error)
	ListTransfers(ctx context.Context, tenantID uuid.UUID, filter BranchTransferFilter) ([]BranchTransfer, int, error)
	DispatchTransfer(ctx context.Context, tenantID, transferID uuid.UUID, notes *string) error
	ReceiveTransfer(ctx context.Context, tenantID, transferID uuid.UUID, parcelIDs []uuid.UUID, notes *string) error
}

type pgRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) NextTransferNumber(ctx context.Context, tenantID uuid.UUID) (string, error) {
	dateStr := time.Now().UTC().Format("20060102")
	query := `
		INSERT INTO tenant_parcel_sequences (tenant_id, last_number, updated_at)
		VALUES ($1, 1, NOW())
		ON CONFLICT (tenant_id)
		DO UPDATE SET last_number = tenant_parcel_sequences.last_number + 1, updated_at = NOW()
		RETURNING last_number;
	`
	var seqNum int
	if err := r.pool.QueryRow(ctx, query, tenantID).Scan(&seqNum); err != nil {
		return "", fmt.Errorf("failed to generate transfer number sequence: %w", err)
	}

	return fmt.Sprintf("TRF-%s-%04d", dateStr, seqNum), nil
}

func (r *pgRepository) CreateTransfer(ctx context.Context, t *BranchTransfer, parcelIDs []uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if t.Status == "" {
		t.Status = StatusPending
	}

	if t.CreatedBy != nil && *t.CreatedBy == uuid.Nil {
		t.CreatedBy = nil
	}

	insertTransferQuery := `
		INSERT INTO branch_transfers (
			tenant_id, transfer_number, source_branch_id, destination_branch_id,
			driver_id, vehicle_id, status, notes, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at;
	`

	err = tx.QueryRow(
		ctx,
		insertTransferQuery,
		t.TenantID,
		t.TransferNumber,
		t.SourceBranchID,
		t.DestinationBranchID,
		t.DriverID,
		t.VehicleID,
		t.Status,
		t.Notes,
		t.CreatedBy,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert branch transfer: %w", err)
	}

	// Insert transfer parcels
	insertParcelQuery := `
		INSERT INTO branch_transfer_parcels (transfer_id, parcel_id)
		VALUES ($1, $2);
	`
	for _, pID := range parcelIDs {
		if _, err := tx.Exec(ctx, insertParcelQuery, t.ID, pID); err != nil {
			return fmt.Errorf("failed to link parcel %s to transfer: %w", pID, err)
		}
	}

	return tx.Commit(ctx)
}

func (r *pgRepository) GetTransferByID(ctx context.Context, tenantID, transferID uuid.UUID) (*BranchTransfer, []BranchTransferParcel, error) {
	query := `
		SELECT 
			bt.id, bt.tenant_id, bt.transfer_number, bt.source_branch_id, bt.destination_branch_id,
			bt.driver_id, bt.vehicle_id, bt.status, bt.dispatched_at, bt.received_at, bt.notes,
			bt.created_by, bt.created_at, bt.updated_at,
			sb.name AS source_branch_name, sb.branch_code AS source_branch_code,
			db.name AS destination_branch_name, db.branch_code AS destination_branch_code,
			CONCAT(e.first_name, ' ', e.last_name) AS driver_name, e.employee_code AS driver_code,
			v.registration_number AS vehicle_registration_number,
			(SELECT COUNT(*) FROM branch_transfer_parcels WHERE transfer_id = bt.id) AS parcels_count
		FROM branch_transfers bt
		JOIN branches sb ON bt.source_branch_id = sb.id
		JOIN branches db ON bt.destination_branch_id = db.id
		LEFT JOIN employees e ON bt.driver_id = e.id
		LEFT JOIN vehicles v ON bt.vehicle_id = v.id
		WHERE bt.id = $1 AND bt.tenant_id = $2;
	`

	bt := &BranchTransfer{}
	err := r.pool.QueryRow(ctx, query, transferID, tenantID).Scan(
		&bt.ID, &bt.TenantID, &bt.TransferNumber, &bt.SourceBranchID, &bt.DestinationBranchID,
		&bt.DriverID, &bt.VehicleID, &bt.Status, &bt.DispatchedAt, &bt.ReceivedAt, &bt.Notes,
		&bt.CreatedBy, &bt.CreatedAt, &bt.UpdatedAt,
		&bt.SourceBranchName, &bt.SourceBranchCode,
		&bt.DestinationBranchName, &bt.DestinationBranchCode,
		&bt.DriverName, &bt.DriverCode,
		&bt.VehicleRegNum,
		&bt.ParcelsCount,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, ErrTransferNotFound
		}
		return nil, nil, fmt.Errorf("failed to get branch transfer: %w", err)
	}

	parcelsQuery := `
		SELECT 
			tp.transfer_id, tp.parcel_id, tp.received, tp.received_at,
			p.tracking_number, p.weight_kg, p.status, p.receiver_name
		FROM branch_transfer_parcels tp
		JOIN parcels p ON tp.parcel_id = p.id
		WHERE tp.transfer_id = $1;
	`

	rows, err := r.pool.Query(ctx, parcelsQuery, transferID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query transfer parcels: %w", err)
	}
	defer rows.Close()

	parcels := make([]BranchTransferParcel, 0)
	for rows.Next() {
		var p BranchTransferParcel
		err := rows.Scan(
			&p.TransferID, &p.ParcelID, &p.Received, &p.ReceivedAt,
			&p.TrackingNumber, &p.WeightKG, &p.Status, &p.ReceiverName,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan transfer parcel: %w", err)
		}
		parcels = append(parcels, p)
	}

	return bt, parcels, nil
}

func (r *pgRepository) ListTransfers(ctx context.Context, tenantID uuid.UUID, filter BranchTransferFilter) ([]BranchTransfer, int, error) {
	baseQuery := `
		FROM branch_transfers bt
		JOIN branches sb ON bt.source_branch_id = sb.id
		JOIN branches db ON bt.destination_branch_id = db.id
		LEFT JOIN employees e ON bt.driver_id = e.id
		LEFT JOIN vehicles v ON bt.vehicle_id = v.id
		WHERE bt.tenant_id = $1
	`

	args := []any{tenantID}
	argIdx := 2

	if filter.Status != nil && *filter.Status != "" {
		baseQuery += fmt.Sprintf(" AND bt.status = $%d", argIdx)
		args = append(args, *filter.Status)
		argIdx++
	}
	if filter.SourceBranchID != nil {
		baseQuery += fmt.Sprintf(" AND bt.source_branch_id = $%d", argIdx)
		args = append(args, *filter.SourceBranchID)
		argIdx++
	}
	if filter.DestinationBranchID != nil {
		baseQuery += fmt.Sprintf(" AND bt.destination_branch_id = $%d", argIdx)
		args = append(args, *filter.DestinationBranchID)
		argIdx++
	}

	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count transfers: %w", err)
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
			bt.id, bt.tenant_id, bt.transfer_number, bt.source_branch_id, bt.destination_branch_id,
			bt.driver_id, bt.vehicle_id, bt.status, bt.dispatched_at, bt.received_at, bt.notes,
			bt.created_by, bt.created_at, bt.updated_at,
			sb.name AS source_branch_name, sb.branch_code AS source_branch_code,
			db.name AS destination_branch_name, db.branch_code AS destination_branch_code,
			CONCAT(e.first_name, ' ', e.last_name) AS driver_name, e.employee_code AS driver_code,
			v.registration_number AS vehicle_registration_number,
			(SELECT COUNT(*) FROM branch_transfer_parcels WHERE transfer_id = bt.id) AS parcels_count
	` + baseQuery + fmt.Sprintf(" ORDER BY bt.created_at DESC LIMIT $%d OFFSET $%d;", argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list transfers: %w", err)
	}
	defer rows.Close()

	transfers := make([]BranchTransfer, 0)
	for rows.Next() {
		var bt BranchTransfer
		err := rows.Scan(
			&bt.ID, &bt.TenantID, &bt.TransferNumber, &bt.SourceBranchID, &bt.DestinationBranchID,
			&bt.DriverID, &bt.VehicleID, &bt.Status, &bt.DispatchedAt, &bt.ReceivedAt, &bt.Notes,
			&bt.CreatedBy, &bt.CreatedAt, &bt.UpdatedAt,
			&bt.SourceBranchName, &bt.SourceBranchCode,
			&bt.DestinationBranchName, &bt.DestinationBranchCode,
			&bt.DriverName, &bt.DriverCode,
			&bt.VehicleRegNum,
			&bt.ParcelsCount,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan transfer row: %w", err)
		}
		transfers = append(transfers, bt)
	}

	return transfers, total, nil
}

func (r *pgRepository) DispatchTransfer(ctx context.Context, tenantID, transferID uuid.UUID, notes *string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Lock transfer
	var currentStatus string
	var sourceBranchID uuid.UUID
	var destBranchID uuid.UUID

	lockQuery := `
		SELECT status, source_branch_id, destination_branch_id 
		FROM branch_transfers 
		WHERE id = $1 AND tenant_id = $2 
		FOR UPDATE;
	`
	err = tx.QueryRow(ctx, lockQuery, transferID, tenantID).Scan(&currentStatus, &sourceBranchID, &destBranchID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrTransferNotFound
		}
		return fmt.Errorf("failed to lock transfer: %w", err)
	}

	if currentStatus != StatusPending {
		return ErrInvalidStatusAction
	}

	// Update transfer to IN_TRANSIT
	updateTransferQuery := `
		UPDATE branch_transfers SET
			status = $1,
			dispatched_at = NOW(),
			notes = COALESCE($2, notes),
			updated_at = NOW()
		WHERE id = $3;
	`
	if _, err := tx.Exec(ctx, updateTransferQuery, StatusInTransit, notes, transferID); err != nil {
		return fmt.Errorf("failed to update transfer status: %w", err)
	}

	// Update all linked parcels to IN_TRANSIT
	updateParcelsQuery := `
		UPDATE parcels SET
			status = 'IN_TRANSIT',
			current_branch_id = NULL,
			updated_at = NOW()
		WHERE id IN (SELECT parcel_id FROM branch_transfer_parcels WHERE transfer_id = $1);
	`
	if _, err := tx.Exec(ctx, updateParcelsQuery, transferID); err != nil {
		return fmt.Errorf("failed to update parcels to IN_TRANSIT: %w", err)
	}

	// Record status history for all parcels
	historyQuery := `
		INSERT INTO parcel_status_history (
			tenant_id, parcel_id, from_status, to_status, branch_id, actor_role, notes, created_at
		)
		SELECT $1, parcel_id, 'RECEIVED_AT_ORIGIN_BRANCH', 'IN_TRANSIT', $2, 'DISPATCHER', 'Dispatched on inter-branch linehaul transfer', NOW()
		FROM branch_transfer_parcels
		WHERE transfer_id = $3;
	`
	if _, err := tx.Exec(ctx, historyQuery, tenantID, sourceBranchID, transferID); err != nil {
		return fmt.Errorf("failed to insert parcels status history: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *pgRepository) ReceiveTransfer(ctx context.Context, tenantID, transferID uuid.UUID, parcelIDs []uuid.UUID, notes *string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var currentStatus string
	var destBranchID uuid.UUID

	lockQuery := `
		SELECT status, destination_branch_id 
		FROM branch_transfers 
		WHERE id = $1 AND tenant_id = $2 
		FOR UPDATE;
	`
	err = tx.QueryRow(ctx, lockQuery, transferID, tenantID).Scan(&currentStatus, &destBranchID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrTransferNotFound
		}
		return fmt.Errorf("failed to lock transfer: %w", err)
	}

	if currentStatus == StatusReceived {
		return ErrTransferAlreadyReceived
	}

	// If no specific parcel IDs provided, receive all parcels on manifest
	if len(parcelIDs) == 0 {
		updateAllParcelsQuery := `
			UPDATE branch_transfer_parcels SET
				received = TRUE,
				received_at = NOW()
			WHERE transfer_id = $1;
		`
		if _, err := tx.Exec(ctx, updateAllParcelsQuery, transferID); err != nil {
			return fmt.Errorf("failed to mark all parcels received: %w", err)
		}
	} else {
		updateParcelsQuery := `
			UPDATE branch_transfer_parcels SET
				received = TRUE,
				received_at = NOW()
			WHERE transfer_id = $1 AND parcel_id = ANY($2);
		`
		if _, err := tx.Exec(ctx, updateParcelsQuery, transferID, parcelIDs); err != nil {
			return fmt.Errorf("failed to mark selected parcels received: %w", err)
		}
	}

	// Update parcels' status to RECEIVED_AT_TRANSFER_BRANCH and set current_branch_id = destBranchID
	updateParcelsStatusQuery := `
		UPDATE parcels SET
			status = 'RECEIVED_AT_TRANSFER_BRANCH',
			current_branch_id = $1,
			updated_at = NOW()
		WHERE id IN (
			SELECT parcel_id FROM branch_transfer_parcels 
			WHERE transfer_id = $2 AND received = TRUE
		);
	`
	if _, err := tx.Exec(ctx, updateParcelsStatusQuery, destBranchID, transferID); err != nil {
		return fmt.Errorf("failed to update parcels status: %w", err)
	}

	// Record history for received parcels
	historyQuery := `
		INSERT INTO parcel_status_history (
			tenant_id, parcel_id, from_status, to_status, branch_id, actor_role, notes, created_at
		)
		SELECT $1, parcel_id, 'IN_TRANSIT', 'RECEIVED_AT_TRANSFER_BRANCH', $2, 'OPERATOR', 'Arrived and scanned into destination hub', NOW()
		FROM branch_transfer_parcels
		WHERE transfer_id = $3 AND received = TRUE;
	`
	if _, err := tx.Exec(ctx, historyQuery, tenantID, destBranchID, transferID); err != nil {
		return fmt.Errorf("failed to insert status history for received parcels: %w", err)
	}

	// Check if all parcels are received
	var pendingCount int
	checkPendingQuery := `SELECT COUNT(*) FROM branch_transfer_parcels WHERE transfer_id = $1 AND received = FALSE;`
	if err := tx.QueryRow(ctx, checkPendingQuery, transferID).Scan(&pendingCount); err != nil {
		return fmt.Errorf("failed to check pending parcels count: %w", err)
	}

	finalStatus := StatusReceived
	if pendingCount > 0 {
		finalStatus = StatusPartiallyReceived
	}

	updateTransferQuery := `
		UPDATE branch_transfers SET
			status = $1,
			received_at = NOW(),
			notes = COALESCE($2, notes),
			updated_at = NOW()
		WHERE id = $3;
	`
	if _, err := tx.Exec(ctx, updateTransferQuery, finalStatus, notes, transferID); err != nil {
		return fmt.Errorf("failed to update transfer status to %s: %w", finalStatus, err)
	}

	return tx.Commit(ctx)
}
