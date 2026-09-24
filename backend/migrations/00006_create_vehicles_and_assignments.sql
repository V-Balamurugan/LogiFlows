-- +goose Up
-- Phase 2 Module E & F: Create Vehicles and Vehicle Assignments Tables

CREATE TABLE IF NOT EXISTS vehicles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    branch_id UUID REFERENCES branches(id) ON DELETE SET NULL,
    registration_number VARCHAR(50) NOT NULL,
    vehicle_type VARCHAR(50) NOT NULL CHECK (vehicle_type IN ('ELECTRIC_VAN', 'VAN', 'MOTORCYCLE', 'TRUCK', 'THREE_WHEELER')),
    make_model VARCHAR(100),
    year INT,
    max_weight_kg DOUBLE PRECISION NOT NULL DEFAULT 500.0,
    max_volume_cbm DOUBLE PRECISION NOT NULL DEFAULT 3.0,
    status VARCHAR(50) NOT NULL DEFAULT 'AVAILABLE' CHECK (status IN ('AVAILABLE', 'ASSIGNED', 'IN_TRANSIT', 'MAINTENANCE', 'DECOMMISSIONED')),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_tenant_registration_number UNIQUE (tenant_id, registration_number)
);

CREATE INDEX IF NOT EXISTS idx_vehicles_tenant_id ON vehicles (tenant_id);
CREATE INDEX IF NOT EXISTS idx_vehicles_branch_id ON vehicles (branch_id);
CREATE INDEX IF NOT EXISTS idx_vehicles_status ON vehicles (status);
CREATE INDEX IF NOT EXISTS idx_vehicles_is_active ON vehicles (is_active);

CREATE TABLE IF NOT EXISTS vehicle_assignments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    driver_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    assigned_by UUID REFERENCES users(id) ON DELETE SET NULL,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    unassigned_at TIMESTAMPTZ,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'COMPLETED', 'CANCELLED')),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_vehicle_assignments_tenant_id ON vehicle_assignments (tenant_id);
CREATE INDEX IF NOT EXISTS idx_vehicle_assignments_vehicle_id ON vehicle_assignments (vehicle_id);
CREATE INDEX IF NOT EXISTS idx_vehicle_assignments_driver_id ON vehicle_assignments (driver_id);

-- Enforce single active assignment per vehicle within tenant
CREATE UNIQUE INDEX IF NOT EXISTS uq_active_vehicle_assignment 
    ON vehicle_assignments (tenant_id, vehicle_id) 
    WHERE status = 'ACTIVE';

-- Enforce single active assignment per driver within tenant
CREATE UNIQUE INDEX IF NOT EXISTS uq_active_driver_assignment 
    ON vehicle_assignments (tenant_id, driver_id) 
    WHERE status = 'ACTIVE';

-- +goose Down
DROP TABLE IF EXISTS vehicle_assignments;
DROP TABLE IF EXISTS vehicles;
