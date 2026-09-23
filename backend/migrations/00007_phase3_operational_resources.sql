-- +goose Up
-- Phase 3: Operational Resource Management (Employees, Roles, Vehicles, and Sequences)

-- 1. Extend employees table with availability, verification, joining date, and soft-delete timestamp
ALTER TABLE employees 
    ADD COLUMN IF NOT EXISTS availability_status VARCHAR(50) NOT NULL DEFAULT 'AVAILABLE' 
        CHECK (availability_status IN ('AVAILABLE', 'BUSY', 'OFF_DUTY', 'UNAVAILABLE')),
    ADD COLUMN IF NOT EXISTS verification_status VARCHAR(50) NOT NULL DEFAULT 'VERIFIED' 
        CHECK (verification_status IN ('PENDING', 'VERIFIED', 'REJECTED')),
    ADD COLUMN IF NOT EXISTS joining_date DATE DEFAULT CURRENT_DATE,
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_employees_availability_status ON employees (availability_status);
CREATE INDEX IF NOT EXISTS idx_employees_verification_status ON employees (verification_status);
CREATE INDEX IF NOT EXISTS idx_employees_deleted_at ON employees (deleted_at);

-- 2. Extend vehicles table with availability status and soft-delete timestamp
ALTER TABLE vehicles
    ADD COLUMN IF NOT EXISTS availability_status VARCHAR(50) NOT NULL DEFAULT 'AVAILABLE'
        CHECK (availability_status IN ('AVAILABLE', 'ASSIGNED', 'MAINTENANCE', 'UNAVAILABLE')),
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_vehicles_availability_status ON vehicles (availability_status);
CREATE INDEX IF NOT EXISTS idx_vehicles_deleted_at ON vehicles (deleted_at);

-- 3. Tenant Employee Sequence table for atomic, race-condition-free sequential code generation (EMP-0001, EMP-0002)
CREATE TABLE IF NOT EXISTS tenant_employee_sequences (
    tenant_id UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    last_number INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed sequence table from existing employee counts per tenant
INSERT INTO tenant_employee_sequences (tenant_id, last_number, updated_at)
SELECT tenant_id, COUNT(*), NOW()
FROM employees
GROUP BY tenant_id
ON CONFLICT (tenant_id) DO NOTHING;

-- 4. Expand tenant_memberships role check constraint to include EMPLOYEE
ALTER TABLE tenant_memberships DROP CONSTRAINT IF EXISTS tenant_memberships_role_check;
ALTER TABLE tenant_memberships ADD CONSTRAINT tenant_memberships_role_check 
    CHECK (role IN ('PLATFORM_ADMIN', 'TENANT_ADMIN', 'TENANT_OPERATOR', 'VIEWER', 'EMPLOYEE'));

-- 5. Expand operational_role check constraint to include field operational roles
ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_operational_role_check;
ALTER TABLE employees ADD CONSTRAINT employees_operational_role_check 
    CHECK (operational_role IN (
        'DRIVER', 
        'OPERATOR', 
        'DISPATCHER', 
        'SUPERVISOR', 
        'MANAGER', 
        'BRANCH_MANAGER', 
        'WAREHOUSE_OPERATOR', 
        'DELIVERY_EXECUTIVE'
    ));

-- +goose Down
DELETE FROM employees WHERE operational_role NOT IN ('DRIVER', 'OPERATOR', 'DISPATCHER', 'SUPERVISOR', 'MANAGER');
ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_operational_role_check;
ALTER TABLE employees ADD CONSTRAINT employees_operational_role_check 
    CHECK (operational_role IN ('DRIVER', 'OPERATOR', 'DISPATCHER', 'SUPERVISOR', 'MANAGER'));

DELETE FROM tenant_memberships WHERE role NOT IN ('PLATFORM_ADMIN', 'TENANT_ADMIN', 'TENANT_OPERATOR', 'VIEWER');
ALTER TABLE tenant_memberships DROP CONSTRAINT IF EXISTS tenant_memberships_role_check;
ALTER TABLE tenant_memberships ADD CONSTRAINT tenant_memberships_role_check 
    CHECK (role IN ('PLATFORM_ADMIN', 'TENANT_ADMIN', 'TENANT_OPERATOR', 'VIEWER'));

DROP TABLE IF EXISTS tenant_employee_sequences;

ALTER TABLE vehicles
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS availability_status;

ALTER TABLE employees
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS joining_date,
    DROP COLUMN IF EXISTS verification_status,
    DROP COLUMN IF EXISTS availability_status;
