package parcels

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
	CreateParcel(ctx context.Context, p *Parcel) error
	GetParcelByID(ctx context.Context, tenantID, parcelID uuid.UUID) (*Parcel, error)
	GetParcelByTrackingNumber(ctx context.Context, trackingNumber string) (*Parcel, error)
	ListParcels(ctx context.Context, tenantID uuid.UUID, filter ParcelFilter) ([]Parcel, int, error)
	UpdateParcel(ctx context.Context, p *Parcel) error
	UpdateParcelStatus(ctx context.Context, tenantID, parcelID uuid.UUID, toStatus string, branchID *uuid.UUID, actorID *uuid.UUID, actorRole string, notes string) error
	GetParcelTimeline(ctx context.Context, tenantID, parcelID uuid.UUID) ([]ParcelStatusHistory, error)
	GetPublicTimeline(ctx context.Context, trackingNumber string) ([]ParcelStatusHistory, error)
	RecordCustodyEvent(ctx context.Context, event *ParcelCustodyEvent) error
	GetCustodyEvents(ctx context.Context, tenantID, parcelID uuid.UUID) ([]ParcelCustodyEvent, error)
	NextTrackingNumber(ctx context.Context, tenantID uuid.UUID) (string, error)
}

type pgRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) NextTrackingNumber(ctx context.Context, tenantID uuid.UUID) (string, error) {
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
		return "", fmt.Errorf("failed to generate tracking number sequence: %w", err)
	}

	tenantShort := strings.ToUpper(strings.ReplaceAll(tenantID.String(), "-", "")[:4])
	return fmt.Sprintf("PKG-%s-%s-%04d", dateStr, tenantShort, seqNum), nil
}

func (r *pgRepository) CreateParcel(ctx context.Context, p *Parcel) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if p.Status == "" {
		p.Status = StatusCreated
	}
	if p.CurrentBranchID == nil {
		p.CurrentBranchID = &p.OriginBranchID
	}

	insertQuery := `
		INSERT INTO parcels (
			tenant_id, tracking_number, sender_name, sender_phone, sender_email, sender_address,
			receiver_name, receiver_phone, receiver_email, receiver_address,
			origin_branch_id, destination_branch_id, current_branch_id,
			weight_kg, dimensions_cm, service_type, declared_value,
			status, special_instructions, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10,
			$11, $12, $13,
			$14, $15, $16, $17,
			$18, $19, $20
		)
		RETURNING id, created_at, updated_at;
	`

	if p.CreatedBy != nil && *p.CreatedBy == uuid.Nil {
		p.CreatedBy = nil
	}

	err = tx.QueryRow(
		ctx,
		insertQuery,
		p.TenantID,
		p.TrackingNumber,
		p.SenderName,
		p.SenderPhone,
		p.SenderEmail,
		p.SenderAddress,
		p.ReceiverName,
		p.ReceiverPhone,
		p.ReceiverEmail,
		p.ReceiverAddress,
		p.OriginBranchID,
		p.DestinationBranchID,
		p.CurrentBranchID,
		p.WeightKG,
		p.DimensionsCM,
		p.ServiceType,
		p.DeclaredValue,
		p.Status,
		p.SpecialInstructions,
		p.CreatedBy,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "uq_tenant_parcel_tracking") {
				return ErrDuplicateTrackingNumber
			}
		}
		return fmt.Errorf("failed to insert parcel: %w", err)
	}

	// Generate and update QR code payload
	qrPayload := GenerateQRPayload(p.TrackingNumber, p.TenantID, p.ID)
	p.QRCodePayload = &qrPayload

	updateQRQuery := `UPDATE parcels SET qr_code_payload = $1 WHERE id = $2;`
	if _, err := tx.Exec(ctx, updateQRQuery, qrPayload, p.ID); err != nil {
		return fmt.Errorf("failed to update parcel qr payload: %w", err)
	}

	// Record initial status history event
	historyQuery := `
		INSERT INTO parcel_status_history (
			tenant_id, parcel_id, from_status, to_status, branch_id, actor_id, actor_role, notes, created_at
		) VALUES ($1, $2, NULL, $3, $4, $5, 'OPERATOR', 'Initial parcel intake recorded', NOW());
	`
	if _, err := tx.Exec(ctx, historyQuery, p.TenantID, p.ID, p.Status, p.OriginBranchID, p.CreatedBy); err != nil {
		return fmt.Errorf("failed to record initial status history: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *pgRepository) GetParcelByID(ctx context.Context, tenantID, parcelID uuid.UUID) (*Parcel, error) {
	query := `
		SELECT 
			p.id, p.tenant_id, p.tracking_number, p.sender_name, p.sender_phone, p.sender_email, p.sender_address,
			p.receiver_name, p.receiver_phone, p.receiver_email, p.receiver_address,
			p.origin_branch_id, p.destination_branch_id, p.current_branch_id,
			p.weight_kg, p.dimensions_cm, p.service_type, p.declared_value,
			p.status, p.special_instructions, p.qr_code_payload, p.created_by,
			p.created_at, p.updated_at, p.deleted_at,
			ob.name AS origin_branch_name, ob.branch_code AS origin_branch_code,
			db.name AS destination_branch_name, db.branch_code AS destination_branch_code,
			cb.name AS current_branch_name, cb.branch_code AS current_branch_code
		FROM parcels p
		LEFT JOIN branches ob ON p.origin_branch_id = ob.id
		LEFT JOIN branches db ON p.destination_branch_id = db.id
		LEFT JOIN branches cb ON p.current_branch_id = cb.id
		WHERE p.id = $1 AND p.tenant_id = $2 AND p.deleted_at IS NULL;
	`

	p := &Parcel{}
	err := r.pool.QueryRow(ctx, query, parcelID, tenantID).Scan(
		&p.ID, &p.TenantID, &p.TrackingNumber, &p.SenderName, &p.SenderPhone, &p.SenderEmail, &p.SenderAddress,
		&p.ReceiverName, &p.ReceiverPhone, &p.ReceiverEmail, &p.ReceiverAddress,
		&p.OriginBranchID, &p.DestinationBranchID, &p.CurrentBranchID,
		&p.WeightKG, &p.DimensionsCM, &p.ServiceType, &p.DeclaredValue,
		&p.Status, &p.SpecialInstructions, &p.QRCodePayload, &p.CreatedBy,
		&p.CreatedAt, &p.UpdatedAt, &p.DeletedAt,
		&p.OriginBranchName, &p.OriginBranchCode,
		&p.DestinationBranchName, &p.DestinationBranchCode,
		&p.CurrentBranchName, &p.CurrentBranchCode,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrParcelNotFound
		}
		return nil, fmt.Errorf("failed to get parcel by id: %w", err)
	}

	return p, nil
}

func (r *pgRepository) GetParcelByTrackingNumber(ctx context.Context, trackingNumber string) (*Parcel, error) {
	query := `
		SELECT 
			p.id, p.tenant_id, p.tracking_number, p.sender_name, p.sender_phone, p.sender_email, p.sender_address,
			p.receiver_name, p.receiver_phone, p.receiver_email, p.receiver_address,
			p.origin_branch_id, p.destination_branch_id, p.current_branch_id,
			p.weight_kg, p.dimensions_cm, p.service_type, p.declared_value,
			p.status, p.special_instructions, p.qr_code_payload, p.created_by,
			p.created_at, p.updated_at, p.deleted_at,
			ob.name AS origin_branch_name, ob.branch_code AS origin_branch_code,
			db.name AS destination_branch_name, db.branch_code AS destination_branch_code,
			cb.name AS current_branch_name, cb.branch_code AS current_branch_code
		FROM parcels p
		LEFT JOIN branches ob ON p.origin_branch_id = ob.id
		LEFT JOIN branches db ON p.destination_branch_id = db.id
		LEFT JOIN branches cb ON p.current_branch_id = cb.id
		WHERE UPPER(p.tracking_number) = UPPER($1) AND p.deleted_at IS NULL;
	`

	p := &Parcel{}
	err := r.pool.QueryRow(ctx, query, trackingNumber).Scan(
		&p.ID, &p.TenantID, &p.TrackingNumber, &p.SenderName, &p.SenderPhone, &p.SenderEmail, &p.SenderAddress,
		&p.ReceiverName, &p.ReceiverPhone, &p.ReceiverEmail, &p.ReceiverAddress,
		&p.OriginBranchID, &p.DestinationBranchID, &p.CurrentBranchID,
		&p.WeightKG, &p.DimensionsCM, &p.ServiceType, &p.DeclaredValue,
		&p.Status, &p.SpecialInstructions, &p.QRCodePayload, &p.CreatedBy,
		&p.CreatedAt, &p.UpdatedAt, &p.DeletedAt,
		&p.OriginBranchName, &p.OriginBranchCode,
		&p.DestinationBranchName, &p.DestinationBranchCode,
		&p.CurrentBranchName, &p.CurrentBranchCode,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrParcelNotFound
		}
		return nil, fmt.Errorf("failed to get parcel by tracking number: %w", err)
	}

	return p, nil
}

func (r *pgRepository) ListParcels(ctx context.Context, tenantID uuid.UUID, filter ParcelFilter) ([]Parcel, int, error) {
	baseQuery := `
		FROM parcels p
		LEFT JOIN branches ob ON p.origin_branch_id = ob.id
		LEFT JOIN branches db ON p.destination_branch_id = db.id
		LEFT JOIN branches cb ON p.current_branch_id = cb.id
		WHERE p.tenant_id = $1 AND p.deleted_at IS NULL
	`

	args := []any{tenantID}
	argIdx := 2

	if filter.Status != nil && *filter.Status != "" {
		baseQuery += fmt.Sprintf(" AND p.status = $%d", argIdx)
		args = append(args, *filter.Status)
		argIdx++
	}
	if filter.OriginBranchID != nil {
		baseQuery += fmt.Sprintf(" AND p.origin_branch_id = $%d", argIdx)
		args = append(args, *filter.OriginBranchID)
		argIdx++
	}
	if filter.DestBranchID != nil {
		baseQuery += fmt.Sprintf(" AND p.destination_branch_id = $%d", argIdx)
		args = append(args, *filter.DestBranchID)
		argIdx++
	}
	if filter.CurrentBranchID != nil {
		baseQuery += fmt.Sprintf(" AND p.current_branch_id = $%d", argIdx)
		args = append(args, *filter.CurrentBranchID)
		argIdx++
	}
	if filter.Search != nil && strings.TrimSpace(*filter.Search) != "" {
		term := "%" + strings.TrimSpace(*filter.Search) + "%"
		baseQuery += fmt.Sprintf(" AND (p.tracking_number ILIKE $%d OR p.sender_name ILIKE $%d OR p.receiver_name ILIKE $%d)", argIdx, argIdx, argIdx)
		args = append(args, term)
		argIdx++
	}
	if filter.DateFrom != nil {
		baseQuery += fmt.Sprintf(" AND p.created_at >= $%d", argIdx)
		args = append(args, *filter.DateFrom)
		argIdx++
	}
	if filter.DateTo != nil {
		baseQuery += fmt.Sprintf(" AND p.created_at <= $%d", argIdx)
		args = append(args, *filter.DateTo)
		argIdx++
	}

	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count parcels: %w", err)
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
			p.id, p.tenant_id, p.tracking_number, p.sender_name, p.sender_phone, p.sender_email, p.sender_address,
			p.receiver_name, p.receiver_phone, p.receiver_email, p.receiver_address,
			p.origin_branch_id, p.destination_branch_id, p.current_branch_id,
			p.weight_kg, p.dimensions_cm, p.service_type, p.declared_value,
			p.status, p.special_instructions, p.qr_code_payload, p.created_by,
			p.created_at, p.updated_at, p.deleted_at,
			ob.name AS origin_branch_name, ob.branch_code AS origin_branch_code,
			db.name AS destination_branch_name, db.branch_code AS destination_branch_code,
			cb.name AS current_branch_name, cb.branch_code AS current_branch_code
	` + baseQuery + fmt.Sprintf(" ORDER BY p.created_at DESC LIMIT $%d OFFSET $%d;", argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list parcels: %w", err)
	}
	defer rows.Close()

	parcels := make([]Parcel, 0)
	for rows.Next() {
		var p Parcel
		err := rows.Scan(
			&p.ID, &p.TenantID, &p.TrackingNumber, &p.SenderName, &p.SenderPhone, &p.SenderEmail, &p.SenderAddress,
			&p.ReceiverName, &p.ReceiverPhone, &p.ReceiverEmail, &p.ReceiverAddress,
			&p.OriginBranchID, &p.DestinationBranchID, &p.CurrentBranchID,
			&p.WeightKG, &p.DimensionsCM, &p.ServiceType, &p.DeclaredValue,
			&p.Status, &p.SpecialInstructions, &p.QRCodePayload, &p.CreatedBy,
			&p.CreatedAt, &p.UpdatedAt, &p.DeletedAt,
			&p.OriginBranchName, &p.OriginBranchCode,
			&p.DestinationBranchName, &p.DestinationBranchCode,
			&p.CurrentBranchName, &p.CurrentBranchCode,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan parcel row: %w", err)
		}
		parcels = append(parcels, p)
	}

	return parcels, total, nil
}

func (r *pgRepository) UpdateParcel(ctx context.Context, p *Parcel) error {
	query := `
		UPDATE parcels SET
			sender_name = COALESCE(NULLIF($1, ''), sender_name),
			sender_phone = COALESCE(NULLIF($2, ''), sender_phone),
			sender_email = $3,
			sender_address = COALESCE(NULLIF($4, ''), sender_address),
			receiver_name = COALESCE(NULLIF($5, ''), receiver_name),
			receiver_phone = COALESCE(NULLIF($6, ''), receiver_phone),
			receiver_email = $7,
			receiver_address = COALESCE(NULLIF($8, ''), receiver_address),
			weight_kg = CASE WHEN $9 > 0 THEN $9 ELSE weight_kg END,
			dimensions_cm = COALESCE(NULLIF($10, ''), dimensions_cm),
			service_type = COALESCE(NULLIF($11, ''), service_type),
			declared_value = CASE WHEN $12 >= 0 THEN $12 ELSE declared_value END,
			special_instructions = $13,
			updated_at = NOW()
		WHERE id = $14 AND tenant_id = $15 AND deleted_at IS NULL;
	`

	cmd, err := r.pool.Exec(
		ctx,
		query,
		p.SenderName,
		p.SenderPhone,
		p.SenderEmail,
		p.SenderAddress,
		p.ReceiverName,
		p.ReceiverPhone,
		p.ReceiverEmail,
		p.ReceiverAddress,
		p.WeightKG,
		p.DimensionsCM,
		p.ServiceType,
		p.DeclaredValue,
		p.SpecialInstructions,
		p.ID,
		p.TenantID,
	)

	if err != nil {
		return fmt.Errorf("failed to update parcel: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrParcelNotFound
	}

	return nil
}

func (r *pgRepository) UpdateParcelStatus(ctx context.Context, tenantID, parcelID uuid.UUID, toStatus string, branchID *uuid.UUID, actorID *uuid.UUID, actorRole string, notes string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var currentStatus string
	var currentBranchID *uuid.UUID

	lockQuery := `
		SELECT status, current_branch_id 
		FROM parcels 
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
		FOR UPDATE;
	`
	err = tx.QueryRow(ctx, lockQuery, parcelID, tenantID).Scan(&currentStatus, &currentBranchID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrParcelNotFound
		}
		return fmt.Errorf("failed to lock parcel for status update: %w", err)
	}

	if !IsValidTransition(currentStatus, toStatus) {
		return fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidStateTransition, currentStatus, toStatus)
	}

	targetBranch := currentBranchID
	if branchID != nil {
		targetBranch = branchID
	}

	updateQuery := `
		UPDATE parcels SET
			status = $1,
			current_branch_id = $2,
			updated_at = NOW()
		WHERE id = $3 AND tenant_id = $4;
	`
	if _, err := tx.Exec(ctx, updateQuery, toStatus, targetBranch, parcelID, tenantID); err != nil {
		return fmt.Errorf("failed to update parcel status: %w", err)
	}

	if actorID != nil && *actorID == uuid.Nil {
		actorID = nil
	}

	historyQuery := `
		INSERT INTO parcel_status_history (
			tenant_id, parcel_id, from_status, to_status, branch_id, actor_id, actor_role, notes, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW());
	`
	if _, err := tx.Exec(ctx, historyQuery, tenantID, parcelID, currentStatus, toStatus, targetBranch, actorID, actorRole, notes); err != nil {
		return fmt.Errorf("failed to insert parcel status history: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *pgRepository) GetParcelTimeline(ctx context.Context, tenantID, parcelID uuid.UUID) ([]ParcelStatusHistory, error) {
	query := `
		SELECT 
			h.id, h.tenant_id, h.parcel_id, h.from_status, h.to_status,
			h.branch_id, b.name AS branch_name,
			h.actor_id, u.full_name AS actor_name, h.actor_role,
			h.notes, h.created_at
		FROM parcel_status_history h
		LEFT JOIN branches b ON h.branch_id = b.id
		LEFT JOIN users u ON h.actor_id = u.id
		WHERE h.parcel_id = $1 AND h.tenant_id = $2
		ORDER BY h.created_at ASC;
	`

	rows, err := r.pool.Query(ctx, query, parcelID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query parcel timeline: %w", err)
	}
	defer rows.Close()

	timeline := make([]ParcelStatusHistory, 0)
	for rows.Next() {
		var h ParcelStatusHistory
		err := rows.Scan(
			&h.ID, &h.TenantID, &h.ParcelID, &h.FromStatus, &h.ToStatus,
			&h.BranchID, &h.BranchName,
			&h.ActorID, &h.ActorName, &h.ActorRole,
			&h.Notes, &h.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan timeline row: %w", err)
		}
		timeline = append(timeline, h)
	}

	return timeline, nil
}

func (r *pgRepository) GetPublicTimeline(ctx context.Context, trackingNumber string) ([]ParcelStatusHistory, error) {
	query := `
		SELECT 
			h.id, h.tenant_id, h.parcel_id, h.from_status, h.to_status,
			h.branch_id, b.name AS branch_name,
			NULL AS actor_id, NULL AS actor_name, NULL AS actor_role,
			h.notes, h.created_at
		FROM parcel_status_history h
		JOIN parcels p ON h.parcel_id = p.id
		LEFT JOIN branches b ON h.branch_id = b.id
		WHERE UPPER(p.tracking_number) = UPPER($1) AND p.deleted_at IS NULL
		ORDER BY h.created_at ASC;
	`

	rows, err := r.pool.Query(ctx, query, trackingNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to query public timeline: %w", err)
	}
	defer rows.Close()

	timeline := make([]ParcelStatusHistory, 0)
	for rows.Next() {
		var h ParcelStatusHistory
		err := rows.Scan(
			&h.ID, &h.TenantID, &h.ParcelID, &h.FromStatus, &h.ToStatus,
			&h.BranchID, &h.BranchName,
			&h.ActorID, &h.ActorName, &h.ActorRole,
			&h.Notes, &h.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan public timeline row: %w", err)
		}
		timeline = append(timeline, h)
	}

	return timeline, nil
}

func (r *pgRepository) RecordCustodyEvent(ctx context.Context, event *ParcelCustodyEvent) error {
	if event.EmployeeID != nil && *event.EmployeeID == uuid.Nil {
		event.EmployeeID = nil
	}

	query := `
		INSERT INTO parcel_custody_events (
			tenant_id, parcel_id, employee_id, from_branch_id, to_branch_id,
			event_type, signature_note, verification_code, created_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, NOW()
		) RETURNING id, created_at;
	`

	return r.pool.QueryRow(
		ctx,
		query,
		event.TenantID,
		event.ParcelID,
		event.EmployeeID,
		event.FromBranchID,
		event.ToBranchID,
		event.EventType,
		event.SignatureNote,
		event.VerificationCode,
	).Scan(&event.ID, &event.CreatedAt)
}

func (r *pgRepository) GetCustodyEvents(ctx context.Context, tenantID, parcelID uuid.UUID) ([]ParcelCustodyEvent, error) {
	query := `
		SELECT 
			c.id, c.tenant_id, c.parcel_id, c.employee_id,
			CONCAT(e.first_name, ' ', e.last_name) AS employee_name,
			c.from_branch_id, fb.name AS from_branch_name,
			c.to_branch_id, tb.name AS to_branch_name,
			c.event_type, c.signature_note, c.verification_code, c.created_at
		FROM parcel_custody_events c
		LEFT JOIN employees e ON c.employee_id = e.id
		LEFT JOIN branches fb ON c.from_branch_id = fb.id
		LEFT JOIN branches tb ON c.to_branch_id = tb.id
		WHERE c.parcel_id = $1 AND c.tenant_id = $2
		ORDER BY c.created_at ASC;
	`

	rows, err := r.pool.Query(ctx, query, parcelID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query custody events: %w", err)
	}
	defer rows.Close()

	events := make([]ParcelCustodyEvent, 0)
	for rows.Next() {
		var ev ParcelCustodyEvent
		err := rows.Scan(
			&ev.ID, &ev.TenantID, &ev.ParcelID, &ev.EmployeeID,
			&ev.EmployeeName,
			&ev.FromBranchID, &ev.FromBranchName,
			&ev.ToBranchID, &ev.ToBranchName,
			&ev.EventType, &ev.SignatureNote, &ev.VerificationCode, &ev.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan custody event: %w", err)
		}
		events = append(events, ev)
	}

	return events, nil
}
