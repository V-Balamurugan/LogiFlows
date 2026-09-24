package customers

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
	NextCustomerCode(ctx context.Context, tenantID uuid.UUID) (string, error)
	CreateCustomer(ctx context.Context, c *Customer) error
	GetCustomerByID(ctx context.Context, tenantID, customerID uuid.UUID) (*Customer, error)
	GetCustomerByCode(ctx context.Context, tenantID uuid.UUID, code string) (*Customer, error)
	ListCustomers(ctx context.Context, tenantID uuid.UUID, filter CustomerFilter) ([]Customer, int, error)
	UpdateCustomer(ctx context.Context, c *Customer) error
	DeleteCustomer(ctx context.Context, tenantID, customerID uuid.UUID) error
}

type pgRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) NextCustomerCode(ctx context.Context, tenantID uuid.UUID) (string, error) {
	dateStr := time.Now().UTC().Format("20060102")
	query := `
		INSERT INTO tenant_customer_sequences (tenant_id, last_number, updated_at)
		VALUES ($1, 1, NOW())
		ON CONFLICT (tenant_id)
		DO UPDATE SET last_number = tenant_customer_sequences.last_number + 1, updated_at = NOW()
		RETURNING last_number;
	`

	var seqNum int
	if err := r.pool.QueryRow(ctx, query, tenantID).Scan(&seqNum); err != nil {
		return "", fmt.Errorf("failed to generate customer code sequence: %w", err)
	}

	return fmt.Sprintf("CUST-%s-%04d", dateStr, seqNum), nil
}

func (r *pgRepository) CreateCustomer(ctx context.Context, c *Customer) error {
	query := `
		INSERT INTO customers (
			tenant_id, customer_code, name, company_name, customer_type,
			email, phone, address_line1, address_line2, city, state,
			postal_code, country, status, credit_limit, contract_tier, notes
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10, $11,
			$12, $13, $14, $15, $16, $17
		) RETURNING id, created_at, updated_at;
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		c.TenantID,
		c.CustomerCode,
		c.Name,
		c.CompanyName,
		c.CustomerType,
		c.Email,
		c.Phone,
		c.AddressLine1,
		c.AddressLine2,
		c.City,
		c.State,
		c.PostalCode,
		c.Country,
		c.Status,
		c.CreditLimit,
		c.ContractTier,
		c.Notes,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "uq_tenant_customer_code") {
				return ErrDuplicateCustomerCode
			}
		}
		return fmt.Errorf("failed to insert customer: %w", err)
	}

	return nil
}

func (r *pgRepository) GetCustomerByID(ctx context.Context, tenantID, customerID uuid.UUID) (*Customer, error) {
	query := `
		SELECT 
			id, tenant_id, customer_code, name, company_name, customer_type,
			email, phone, address_line1, address_line2, city, state,
			postal_code, country, status, credit_limit, contract_tier, notes,
			created_at, updated_at, deleted_at
		FROM customers
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL;
	`

	c := &Customer{}
	err := r.pool.QueryRow(ctx, query, customerID, tenantID).Scan(
		&c.ID, &c.TenantID, &c.CustomerCode, &c.Name, &c.CompanyName, &c.CustomerType,
		&c.Email, &c.Phone, &c.AddressLine1, &c.AddressLine2, &c.City, &c.State,
		&c.PostalCode, &c.Country, &c.Status, &c.CreditLimit, &c.ContractTier, &c.Notes,
		&c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCustomerNotFound
		}
		return nil, fmt.Errorf("failed to get customer by id: %w", err)
	}

	return c, nil
}

func (r *pgRepository) GetCustomerByCode(ctx context.Context, tenantID uuid.UUID, code string) (*Customer, error) {
	query := `
		SELECT 
			id, tenant_id, customer_code, name, company_name, customer_type,
			email, phone, address_line1, address_line2, city, state,
			postal_code, country, status, credit_limit, contract_tier, notes,
			created_at, updated_at, deleted_at
		FROM customers
		WHERE customer_code = $1 AND tenant_id = $2 AND deleted_at IS NULL;
	`

	c := &Customer{}
	err := r.pool.QueryRow(ctx, query, code, tenantID).Scan(
		&c.ID, &c.TenantID, &c.CustomerCode, &c.Name, &c.CompanyName, &c.CustomerType,
		&c.Email, &c.Phone, &c.AddressLine1, &c.AddressLine2, &c.City, &c.State,
		&c.PostalCode, &c.Country, &c.Status, &c.CreditLimit, &c.ContractTier, &c.Notes,
		&c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCustomerNotFound
		}
		return nil, fmt.Errorf("failed to get customer by code: %w", err)
	}

	return c, nil
}

func (r *pgRepository) ListCustomers(ctx context.Context, tenantID uuid.UUID, filter CustomerFilter) ([]Customer, int, error) {
	baseQuery := " FROM customers WHERE tenant_id = $1 AND deleted_at IS NULL"
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.Status != nil && strings.TrimSpace(*filter.Status) != "" {
		baseQuery += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, strings.ToUpper(strings.TrimSpace(*filter.Status)))
		argIdx++
	}

	if filter.CustomerType != nil && strings.TrimSpace(*filter.CustomerType) != "" {
		baseQuery += fmt.Sprintf(" AND customer_type = $%d", argIdx)
		args = append(args, strings.ToUpper(strings.TrimSpace(*filter.CustomerType)))
		argIdx++
	}

	if filter.Search != nil && strings.TrimSpace(*filter.Search) != "" {
		term := "%" + strings.TrimSpace(*filter.Search) + "%"
		baseQuery += fmt.Sprintf(" AND (name ILIKE $%d OR customer_code ILIKE $%d OR company_name ILIKE $%d OR phone ILIKE $%d OR email ILIKE $%d)", argIdx, argIdx, argIdx, argIdx, argIdx)
		args = append(args, term)
		argIdx++
	}

	countQuery := "SELECT COUNT(*)" + baseQuery
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count customers: %w", err)
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
			id, tenant_id, customer_code, name, company_name, customer_type,
			email, phone, address_line1, address_line2, city, state,
			postal_code, country, status, credit_limit, contract_tier, notes,
			created_at, updated_at, deleted_at
	` + baseQuery + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d;", argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list customers: %w", err)
	}
	defer rows.Close()

	customersList := make([]Customer, 0)
	for rows.Next() {
		var c Customer
		err := rows.Scan(
			&c.ID, &c.TenantID, &c.CustomerCode, &c.Name, &c.CompanyName, &c.CustomerType,
			&c.Email, &c.Phone, &c.AddressLine1, &c.AddressLine2, &c.City, &c.State,
			&c.PostalCode, &c.Country, &c.Status, &c.CreditLimit, &c.ContractTier, &c.Notes,
			&c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan customer row: %w", err)
		}
		customersList = append(customersList, c)
	}

	return customersList, total, nil
}

func (r *pgRepository) UpdateCustomer(ctx context.Context, c *Customer) error {
	query := `
		UPDATE customers SET
			name = $1,
			company_name = $2,
			customer_type = $3,
			email = $4,
			phone = $5,
			address_line1 = $6,
			address_line2 = $7,
			city = $8,
			state = $9,
			postal_code = $10,
			country = $11,
			status = $12,
			credit_limit = $13,
			contract_tier = $14,
			notes = $15,
			updated_at = NOW()
		WHERE id = $16 AND tenant_id = $17 AND deleted_at IS NULL;
	`

	cmd, err := r.pool.Exec(
		ctx,
		query,
		c.Name,
		c.CompanyName,
		c.CustomerType,
		c.Email,
		c.Phone,
		c.AddressLine1,
		c.AddressLine2,
		c.City,
		c.State,
		c.PostalCode,
		c.Country,
		c.Status,
		c.CreditLimit,
		c.ContractTier,
		c.Notes,
		c.ID,
		c.TenantID,
	)

	if err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrCustomerNotFound
	}

	return nil
}

func (r *pgRepository) DeleteCustomer(ctx context.Context, tenantID, customerID uuid.UUID) error {
	query := `
		UPDATE customers 
		SET deleted_at = NOW(), status = 'INACTIVE' 
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL;
	`

	cmd, err := r.pool.Exec(ctx, query, customerID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrCustomerNotFound
	}

	return nil
}
