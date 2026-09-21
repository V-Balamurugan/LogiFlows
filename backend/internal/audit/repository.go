package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines persistence operations for audit logs.
type Repository interface {
	Log(ctx context.Context, entry *AuditLog) error
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository initializes a PostgreSQL audit log repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) Log(ctx context.Context, entry *AuditLog) error {
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now().UTC()
	}

	detailsJSON, err := json.Marshal(entry.Details)
	if err != nil {
		detailsJSON = []byte("{}")
	}

	query := `
		INSERT INTO audit_logs (id, tenant_id, user_id, action, resource_type, resource_id, ip_address, user_agent, status, details, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = r.pool.Exec(ctx, query,
		entry.ID,
		entry.TenantID,
		entry.UserID,
		entry.Action,
		entry.ResourceType,
		entry.ResourceID,
		entry.IPAddress,
		entry.UserAgent,
		entry.Status,
		detailsJSON,
		entry.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert audit log entry: %w", err)
	}

	return nil
}
