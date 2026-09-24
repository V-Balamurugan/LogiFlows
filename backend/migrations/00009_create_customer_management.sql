-- +goose Up
-- Phase 4: Customer Management & Parcel Booking Enhancement

-- 1. Tenant Customer Sequence table for atomic, race-condition-free sequential customer code generation (CUST-YYYYMMDD-XXXX)
CREATE TABLE IF NOT EXISTS tenant_customer_sequences (
    tenant_id UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    last_number INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Customers table (Shippers, Consignees, Corporate Clients, Merchants)
CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    customer_code VARCHAR(64) NOT NULL,
    name VARCHAR(150) NOT NULL,
    company_name VARCHAR(150),
    customer_type VARCHAR(50) NOT NULL DEFAULT 'INDIVIDUAL' CHECK (customer_type IN ('INDIVIDUAL', 'BUSINESS', 'ENTERPRISE', 'MERCHANT')),
    email VARCHAR(150),
    phone VARCHAR(50) NOT NULL,
    address_line1 TEXT NOT NULL,
    address_line2 TEXT,
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100) NOT NULL,
    postal_code VARCHAR(20) NOT NULL,
    country VARCHAR(100) NOT NULL DEFAULT 'India',
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE', 'SUSPENDED')),
    credit_limit NUMERIC(12, 2) NOT NULL DEFAULT 0.00 CHECK (credit_limit >= 0),
    contract_tier VARCHAR(50) NOT NULL DEFAULT 'STANDARD' CHECK (contract_tier IN ('STANDARD', 'SILVER', 'GOLD', 'VIP')),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uq_tenant_customer_code UNIQUE (tenant_id, customer_code)
);

CREATE INDEX IF NOT EXISTS idx_customers_tenant_id ON customers(tenant_id);
CREATE INDEX IF NOT EXISTS idx_customers_customer_code ON customers(customer_code);
CREATE INDEX IF NOT EXISTS idx_customers_phone ON customers(phone);
CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email);
CREATE INDEX IF NOT EXISTS idx_customers_status ON customers(status);
CREATE INDEX IF NOT EXISTS idx_customers_type ON customers(customer_type);
CREATE INDEX IF NOT EXISTS idx_customers_deleted_at ON customers(deleted_at);

-- 3. Enhance Parcels table with customer references and calculated pricing
ALTER TABLE parcels ADD COLUMN IF NOT EXISTS sender_customer_id UUID REFERENCES customers(id) ON DELETE SET NULL;
ALTER TABLE parcels ADD COLUMN IF NOT EXISTS receiver_customer_id UUID REFERENCES customers(id) ON DELETE SET NULL;
ALTER TABLE parcels ADD COLUMN IF NOT EXISTS price NUMERIC(10, 2) NOT NULL DEFAULT 0.00;

CREATE INDEX IF NOT EXISTS idx_parcels_sender_customer ON parcels(sender_customer_id);
CREATE INDEX IF NOT EXISTS idx_parcels_receiver_customer ON parcels(receiver_customer_id);

-- +goose Down
DROP INDEX IF EXISTS idx_parcels_receiver_customer;
DROP INDEX IF EXISTS idx_parcels_sender_customer;
ALTER TABLE parcels DROP COLUMN IF EXISTS price;
ALTER TABLE parcels DROP COLUMN IF EXISTS receiver_customer_id;
ALTER TABLE parcels DROP COLUMN IF EXISTS sender_customer_id;

DROP TABLE IF EXISTS customers;
DROP TABLE IF EXISTS tenant_customer_sequences;
