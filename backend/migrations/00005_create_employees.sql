-- +goose Up
-- Phase 2 Module C: Create Employees Table with Tenant and Branch Scoping

CREATE TABLE IF NOT EXISTS employees (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    branch_id UUID REFERENCES branches(id) ON DELETE SET NULL,
    employee_code VARCHAR(50) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(50),
    designation VARCHAR(100) NOT NULL,
    employment_type VARCHAR(50) NOT NULL DEFAULT 'FULL_TIME' CHECK (employment_type IN ('FULL_TIME', 'PART_TIME', 'CONTRACTOR', 'INTERN')),
    operational_role VARCHAR(50) NOT NULL DEFAULT 'DRIVER' CHECK (operational_role IN ('DRIVER', 'OPERATOR', 'DISPATCHER', 'SUPERVISOR', 'MANAGER')),
    license_number VARCHAR(100),
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'ON_LEAVE', 'SUSPENDED', 'TERMINATED')),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_tenant_employee_code UNIQUE (tenant_id, employee_code),
    CONSTRAINT uq_tenant_user_id UNIQUE (tenant_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_employees_tenant_id ON employees (tenant_id);
CREATE INDEX IF NOT EXISTS idx_employees_branch_id ON employees (branch_id);
CREATE INDEX IF NOT EXISTS idx_employees_operational_role ON employees (operational_role);
CREATE INDEX IF NOT EXISTS idx_employees_status ON employees (status);
CREATE INDEX IF NOT EXISTS idx_employees_is_active ON employees (is_active);

-- +goose Down
DROP TABLE IF EXISTS employees;
