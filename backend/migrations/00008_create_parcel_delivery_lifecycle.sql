-- +goose Up
-- Phase 4: Parcel Custody, Inter-Branch Linehaul, Delivery Lifecycle & Proof of Delivery

-- 1. Tenant Parcel Sequence table for atomic, race-condition-free sequential tracking number generation (PKG-YYYYMMDD-XXXX)
CREATE TABLE IF NOT EXISTS tenant_parcel_sequences (
    tenant_id UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    last_number INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Parcels table
CREATE TABLE IF NOT EXISTS parcels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    tracking_number VARCHAR(64) NOT NULL,
    sender_name VARCHAR(150) NOT NULL,
    sender_phone VARCHAR(50) NOT NULL,
    sender_email VARCHAR(150),
    sender_address TEXT NOT NULL,
    receiver_name VARCHAR(150) NOT NULL,
    receiver_phone VARCHAR(50) NOT NULL,
    receiver_email VARCHAR(150),
    receiver_address TEXT NOT NULL,
    origin_branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    destination_branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    current_branch_id UUID REFERENCES branches(id) ON DELETE SET NULL,
    weight_kg NUMERIC(10, 2) NOT NULL CHECK (weight_kg > 0),
    dimensions_cm VARCHAR(50) NOT NULL,
    service_type VARCHAR(50) NOT NULL CHECK (service_type IN ('STANDARD', 'EXPRESS', 'OVERNIGHT', 'SAME_DAY')),
    declared_value NUMERIC(12, 2) NOT NULL DEFAULT 0.00 CHECK (declared_value >= 0),
    status VARCHAR(50) NOT NULL DEFAULT 'CREATED' CHECK (status IN (
        'CREATED', 'BOOKED', 'READY_FOR_PICKUP', 'PICKED_UP',
        'RECEIVED_AT_ORIGIN_BRANCH', 'IN_TRANSIT', 'RECEIVED_AT_TRANSFER_BRANCH',
        'OUT_FOR_DELIVERY', 'DELIVERY_ATTEMPTED', 'DELIVERED',
        'DELIVERY_FAILED', 'RETURN_INITIATED', 'RETURNED', 'CANCELLED', 'ON_HOLD'
    )),
    special_instructions TEXT,
    qr_code_payload TEXT,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uq_tenant_parcel_tracking UNIQUE (tenant_id, tracking_number)
);

CREATE INDEX IF NOT EXISTS idx_parcels_tenant_id ON parcels(tenant_id);
CREATE INDEX IF NOT EXISTS idx_parcels_tracking_number ON parcels(tracking_number);
CREATE INDEX IF NOT EXISTS idx_parcels_status ON parcels(status);
CREATE INDEX IF NOT EXISTS idx_parcels_origin_branch ON parcels(origin_branch_id);
CREATE INDEX IF NOT EXISTS idx_parcels_destination_branch ON parcels(destination_branch_id);
CREATE INDEX IF NOT EXISTS idx_parcels_current_branch ON parcels(current_branch_id);
CREATE INDEX IF NOT EXISTS idx_parcels_created_at ON parcels(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_parcels_deleted_at ON parcels(deleted_at);

-- 3. Parcel Status History table (immutable transition audit log)
CREATE TABLE IF NOT EXISTS parcel_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    parcel_id UUID NOT NULL REFERENCES parcels(id) ON DELETE CASCADE,
    from_status VARCHAR(50),
    to_status VARCHAR(50) NOT NULL,
    branch_id UUID REFERENCES branches(id) ON DELETE SET NULL,
    actor_id UUID REFERENCES users(id) ON DELETE SET NULL,
    actor_role VARCHAR(50),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_parcel_status_history_parcel_id ON parcel_status_history(parcel_id);
CREATE INDEX IF NOT EXISTS idx_parcel_status_history_tenant_id ON parcel_status_history(tenant_id);
CREATE INDEX IF NOT EXISTS idx_parcel_status_history_created_at ON parcel_status_history(created_at DESC);

-- 4. Parcel Custody Events table (chain of custody)
CREATE TABLE IF NOT EXISTS parcel_custody_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    parcel_id UUID NOT NULL REFERENCES parcels(id) ON DELETE CASCADE,
    employee_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    from_branch_id UUID REFERENCES branches(id) ON DELETE SET NULL,
    to_branch_id UUID REFERENCES branches(id) ON DELETE SET NULL,
    event_type VARCHAR(50) NOT NULL CHECK (event_type IN ('INTAKE', 'DISPATCH', 'RECEIVE', 'HANDOVER_DELIVERY', 'RETURN_INTAKE')),
    signature_note TEXT,
    verification_code VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_parcel_custody_events_parcel_id ON parcel_custody_events(parcel_id);
CREATE INDEX IF NOT EXISTS idx_parcel_custody_events_tenant_id ON parcel_custody_events(tenant_id);

-- 5. Branch Transfers table (inter-branch manifests)
CREATE TABLE IF NOT EXISTS branch_transfers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    transfer_number VARCHAR(64) NOT NULL,
    source_branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    destination_branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    driver_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    vehicle_id UUID REFERENCES vehicles(id) ON DELETE SET NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'IN_TRANSIT', 'RECEIVED', 'CANCELLED', 'PARTIALLY_RECEIVED')),
    dispatched_at TIMESTAMPTZ,
    received_at TIMESTAMPTZ,
    notes TEXT,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_tenant_transfer_number UNIQUE (tenant_id, transfer_number),
    CONSTRAINT chk_different_branches CHECK (source_branch_id <> destination_branch_id)
);

CREATE INDEX IF NOT EXISTS idx_branch_transfers_tenant_id ON branch_transfers(tenant_id);
CREATE INDEX IF NOT EXISTS idx_branch_transfers_source ON branch_transfers(source_branch_id);
CREATE INDEX IF NOT EXISTS idx_branch_transfers_destination ON branch_transfers(destination_branch_id);
CREATE INDEX IF NOT EXISTS idx_branch_transfers_status ON branch_transfers(status);

-- 6. Branch Transfer Parcels junction table
CREATE TABLE IF NOT EXISTS branch_transfer_parcels (
    transfer_id UUID NOT NULL REFERENCES branch_transfers(id) ON DELETE CASCADE,
    parcel_id UUID NOT NULL REFERENCES parcels(id) ON DELETE CASCADE,
    received BOOLEAN NOT NULL DEFAULT FALSE,
    received_at TIMESTAMPTZ,
    PRIMARY KEY (transfer_id, parcel_id)
);

CREATE INDEX IF NOT EXISTS idx_transfer_parcels_parcel ON branch_transfer_parcels(parcel_id);

-- 7. Delivery Tasks table (last-mile driver assignments)
CREATE TABLE IF NOT EXISTS delivery_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    parcel_id UUID NOT NULL REFERENCES parcels(id) ON DELETE CASCADE,
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    assigned_driver_id UUID NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    vehicle_id UUID REFERENCES vehicles(id) ON DELETE SET NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ASSIGNED' CHECK (status IN ('ASSIGNED', 'IN_PROGRESS', 'COMPLETED', 'FAILED', 'CANCELLED', 'RESCHEDULED')),
    priority VARCHAR(20) NOT NULL DEFAULT 'STANDARD' CHECK (priority IN ('LOW', 'STANDARD', 'HIGH', 'URGENT')),
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    failed_at TIMESTAMPTZ,
    failure_reason TEXT,
    rescheduled_for TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Crucial: Partial unique index preventing race condition double-booking of a parcel across active delivery tasks
CREATE UNIQUE INDEX IF NOT EXISTS idx_active_parcel_delivery_task 
    ON delivery_tasks (tenant_id, parcel_id) 
    WHERE status IN ('ASSIGNED', 'IN_PROGRESS');

CREATE INDEX IF NOT EXISTS idx_delivery_tasks_tenant_id ON delivery_tasks(tenant_id);
CREATE INDEX IF NOT EXISTS idx_delivery_tasks_driver ON delivery_tasks(assigned_driver_id);
CREATE INDEX IF NOT EXISTS idx_delivery_tasks_branch ON delivery_tasks(branch_id);
CREATE INDEX IF NOT EXISTS idx_delivery_tasks_status ON delivery_tasks(status);
CREATE INDEX IF NOT EXISTS idx_delivery_tasks_priority ON delivery_tasks(priority);

-- 8. Delivery Attempts table
CREATE TABLE IF NOT EXISTS delivery_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    delivery_task_id UUID NOT NULL REFERENCES delivery_tasks(id) ON DELETE CASCADE,
    attempt_number INT NOT NULL,
    attempted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    outcome VARCHAR(50) NOT NULL CHECK (outcome IN (
        'SUCCESSFUL', 'CUSTOMER_UNAVAILABLE', 'INCORRECT_ADDRESS',
        'REJECTED_BY_CUSTOMER', 'SECURITY_ACCESS_DENIED', 'OTHER'
    )),
    notes TEXT,
    latitude NUMERIC(10, 7),
    longitude NUMERIC(10, 7)
);

CREATE INDEX IF NOT EXISTS idx_delivery_attempts_task_id ON delivery_attempts(delivery_task_id);

-- 9. Delivery Proofs table
CREATE TABLE IF NOT EXISTS delivery_proofs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    delivery_task_id UUID NOT NULL REFERENCES delivery_tasks(id) ON DELETE CASCADE,
    parcel_id UUID NOT NULL REFERENCES parcels(id) ON DELETE CASCADE,
    proof_type VARCHAR(50) NOT NULL CHECK (proof_type IN ('RECIPIENT_SIGNATURE', 'OTP', 'PHOTO', 'CONTACTLESS_DROP')),
    recipient_name VARCHAR(150) NOT NULL,
    recipient_relationship VARCHAR(50) NOT NULL DEFAULT 'SELF',
    otp_code VARCHAR(10),
    signature_data TEXT,
    photo_url TEXT,
    notes TEXT,
    latitude NUMERIC(10, 7),
    longitude NUMERIC(10, 7),
    verified_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_delivery_proof_task UNIQUE (delivery_task_id)
);

CREATE INDEX IF NOT EXISTS idx_delivery_proof_parcel ON delivery_proofs(parcel_id);

-- +goose Down
DROP TABLE IF EXISTS delivery_proofs;
DROP TABLE IF EXISTS delivery_attempts;
DROP INDEX IF EXISTS idx_active_parcel_delivery_task;
DROP TABLE IF EXISTS delivery_tasks;
DROP TABLE IF EXISTS branch_transfer_parcels;
DROP TABLE IF EXISTS branch_transfers;
DROP TABLE IF EXISTS parcel_custody_events;
DROP TABLE IF EXISTS parcel_status_history;
DROP TABLE IF EXISTS parcels;
DROP TABLE IF EXISTS tenant_parcel_sequences;
