package users

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
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
)

// Repository defines the persistence interface for users.
type Repository interface {
	Create(ctx context.Context, user *User) error
	CreateTx(ctx context.Context, tx pgx.Tx, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
	SetEmailVerified(ctx context.Context, id uuid.UUID, verified bool) error
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new PostgreSQL user repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) Create(ctx context.Context, user *User) error {
	return insertUser(ctx, r.pool, user)
}

func (r *pgRepository) CreateTx(ctx context.Context, tx pgx.Tx, user *User) error {
	return insertUser(ctx, tx, user)
}

type dbExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
}

func insertUser(ctx context.Context, exec dbExecutor, user *User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))

	query := `
		INSERT INTO users (id, email, password_hash, full_name, phone_number, is_active, is_platform_admin, email_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := exec.Exec(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.PhoneNumber,
		user.IsActive,
		user.IsPlatformAdmin,
		user.EmailVerified,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // Unique violation
			return ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

func (r *pgRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `
		SELECT id, email, password_hash, full_name, phone_number, is_active, is_platform_admin, email_verified, created_at, updated_at, last_login_at
		FROM users
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)
	return scanUser(row)
}

func (r *pgRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, password_hash, full_name, phone_number, is_active, is_platform_admin, email_verified, created_at, updated_at, last_login_at
		FROM users
		WHERE LOWER(email) = LOWER($1)
	`
	row := r.pool.QueryRow(ctx, query, strings.TrimSpace(email))
	return scanUser(row)
}

func (r *pgRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET last_login_at = $1, updated_at = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to update last login for user: %w", err)
	}
	return nil
}

func (r *pgRepository) SetEmailVerified(ctx context.Context, id uuid.UUID, verified bool) error {
	query := `UPDATE users SET email_verified = $1, updated_at = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, verified, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to update email verification status: %w", err)
	}
	return nil
}

func scanUser(row pgx.Row) (*User, error) {
	var u User
	err := row.Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.FullName,
		&u.PhoneNumber,
		&u.IsActive,
		&u.IsPlatformAdmin,
		&u.EmailVerified,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.LastLoginAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to scan user row: %w", err)
	}
	return &u, nil
}
