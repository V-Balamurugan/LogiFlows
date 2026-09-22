package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenRevoked  = errors.New("refresh token has been revoked")
	ErrRefreshTokenExpired  = errors.New("refresh token has expired")
)

// RefreshToken represents a persisted refresh token record in PostgreSQL.
type RefreshToken struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	IPAddress *string    `json:"ip_address,omitempty"`
	UserAgent *string    `json:"user_agent,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// RefreshTokenRepository defines the interface for refresh token lifecycle persistence.
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) error
	GetByHash(ctx context.Context, hash string) (*RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
}

type pgRefreshTokenRepository struct {
	pool *pgxpool.Pool
}

// NewRefreshTokenRepository creates a new PostgreSQL refresh token repository.
func NewRefreshTokenRepository(pool *pgxpool.Pool) RefreshTokenRepository {
	return &pgRefreshTokenRepository{pool: pool}
}

func (r *pgRefreshTokenRepository) Create(ctx context.Context, token *RefreshToken) error {
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}
	now := time.Now().UTC()
	token.CreatedAt = now

	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, revoked_at, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.RevokedAt,
		token.IPAddress,
		token.UserAgent,
		token.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert refresh token: %w", err)
	}
	return nil
}

func (r *pgRefreshTokenRepository) GetByHash(ctx context.Context, hash string) (*RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, revoked_at, ip_address, user_agent, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`
	var t RefreshToken
	err := r.pool.QueryRow(ctx, query, hash).Scan(
		&t.ID,
		&t.UserID,
		&t.TokenHash,
		&t.ExpiresAt,
		&t.RevokedAt,
		&t.IPAddress,
		&t.UserAgent,
		&t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRefreshTokenNotFound
		}
		return nil, fmt.Errorf("failed to query refresh token by hash: %w", err)
	}
	return &t, nil
}

func (r *pgRefreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked_at = $1 WHERE id = $2 AND revoked_at IS NULL`
	_, err := r.pool.Exec(ctx, query, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}
	return nil
}

func (r *pgRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked_at = $1 WHERE user_id = $2 AND revoked_at IS NULL`
	_, err := r.pool.Exec(ctx, query, time.Now().UTC(), userID)
	if err != nil {
		return fmt.Errorf("failed to revoke all refresh tokens for user: %w", err)
	}
	return nil
}
