-- +goose Up
-- Phase 2: Create Branches Table for Delivery Hubs & Sorting Centers

CREATE TABLE IF NOT EXISTS branches (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    branch_code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    address TEXT NOT NULL,
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100) NOT NULL DEFAULT 'India',
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    location GEOMETRY(Point, 4326),
    coverage_radius_km DOUBLE PRECISION NOT NULL DEFAULT 15.0,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE', 'SUSPENDED')),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_tenant_branch_code UNIQUE (tenant_id, branch_code)
);

CREATE INDEX IF NOT EXISTS idx_branches_tenant_id ON branches (tenant_id);
CREATE INDEX IF NOT EXISTS idx_branches_status ON branches (status);
CREATE INDEX IF NOT EXISTS idx_branches_is_active ON branches (is_active);
CREATE INDEX IF NOT EXISTS idx_branches_location ON branches USING GIST (location);

-- +goose Down
DROP TABLE IF EXISTS branches;
