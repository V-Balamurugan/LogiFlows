# Phase 4 Implementation Report — Customers and Parcel Booking

## 1. Executive Summary
- **Phase:** Phase 4 — Customers and Parcel Booking
- **System:** LogiFlows (Intelligent End-to-End Logistics Coordination & Delivery Management System)
- **Methodology:** Complete SDLC + Agile Scrum + Vertical Slice Development
- **Implementation Status:** COMPLETE & VERIFIED

This report provides the architectural record, database changes, backend services, frontend web application integration, security controls, and test evidence for the completion of the Phase 4 vertical slice.

---

## 2. Completed Features

### 2.1 Customer Profile & Address Book Management (CRM)
- **Customer Classifications:** `INDIVIDUAL`, `BUSINESS`, `ENTERPRISE`, `MERCHANT`.
- **Status Lifecycle:** `ACTIVE`, `INACTIVE`, `SUSPENDED`.
- **Atomic Identification:** Tenant-scoped sequence generator `CUST-YYYYMMDD-XXXX` backed by PostgreSQL row-locking table `tenant_customer_sequences`.
- **Address Entities:** Separate fields for `billing_address` (mandatory for tax/consignor invoices) and `shipping_address` (optional for warehouse delivery destinations).
- **Tax Tracking:** Optional `tax_id` / GSTIN tracking with case-insensitive normalization.
- **RESTful Endpoints:**
  - `POST /api/v1/tenants/:tenant_id/customers`: Register new customer with validation and sequence assignment.
  - `GET /api/v1/tenants/:tenant_id/customers`: List customers with text search (`name`, `company_name`, `email`, `phone`, `tax_id`), classification filter, and status filter.
  - `GET /api/v1/tenants/:tenant_id/customers/:customer_id`: Retrieve specific customer profile with multi-tenant isolation.
  - `PATCH /api/v1/tenants/:tenant_id/customers/:customer_id`: Update profile details with role verification.
  - `DELETE /api/v1/tenants/:tenant_id/customers/:customer_id`: Soft-delete customer (`deleted_at`).
  - `GET /api/v1/tenants/:tenant_id/customers/:customer_id/parcels`: Retrieve all inbound and outbound consignments associated with the customer.

### 2.2 Customer-Linked Parcel Booking Workflow
- **Auto-Population Invariant:** If `sender_customer_id` or `receiver_customer_id` is supplied, customer details (`sender_name`, `sender_phone`, `sender_address`, etc.) are automatically populated from the verified customer record if left empty.
- **Foreign Key Linkage:** Foreign keys with `ON DELETE SET NULL` on `parcels.sender_customer_id` and `parcels.receiver_customer_id`.
- **Real-Time Dynamic Pricing Engine:**
  - Base Fee: Standard (₹50.00), Express (₹120.00), Overnight (₹200.00), Same Day (₹350.00).
  - Weight Fee: ₹20.00 per kg.
  - Insurance Surcharge: 0.5% (0.005) on declared value exceeding ₹1,000.00.
  - Total Price: Calculated and persisted in `parcels.price` column.
- **Atomic Tracking ID & QR Code:**
  - Atomic tracking number `PKG-YYYYMMDD-XXXX-XXXX` with database collision prevention.
  - Standardized JSON QR code payload containing tracking number, tenant ID, and parcel UUID.
- **Label Generation & Downloads:**
  - Full 4x6" shipping label canvas download.
  - High-res standalone QR badge canvas download featuring Destination Hub, City, Parcel ID, and Tracking Number.

### 2.3 Frontend Management Experience
- **Customers & CRM Module (`CustomerList.tsx`):**
  - Customer directory with live search, filters, pagination, and count telemetry.
  - Registration and editing modal (`CustomerModal.tsx`) with strict client-side validation.
  - "View Bookings" modal displaying consignment history.
  - One-click "+ Book Parcel" integration connecting CRM profiles directly to the booking workflow.
- **Upgraded Parcel Booking Workflow (`ParcelList.tsx`):**
  - Instant auto-fill dropdowns for registered senders and recipients.
  - Live dynamic tariff and insurance estimate card reacting to service level, weight, and declared value.
  - Display of price in the parcel data table.
  - Modal action for "QR with Destination & ID (PNG)" download.

---

## 3. Database Schema Changes

### Migration `00009_create_customer_management.sql`
```sql
CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    customer_code VARCHAR(32) NOT NULL,
    customer_type VARCHAR(20) NOT NULL CHECK (customer_type IN ('INDIVIDUAL', 'BUSINESS', 'ENTERPRISE', 'MERCHANT')),
    name VARCHAR(255) NOT NULL,
    company_name VARCHAR(255),
    email VARCHAR(255),
    phone VARCHAR(50) NOT NULL,
    tax_id VARCHAR(50),
    billing_address TEXT NOT NULL,
    shipping_address TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE', 'SUSPENDED')),
    notes TEXT,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uq_tenant_customer_code UNIQUE (tenant_id, customer_code)
);

CREATE TABLE IF NOT EXISTS tenant_customer_sequences (
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    date_prefix VARCHAR(8) NOT NULL,
    last_sequence INT NOT NULL DEFAULT 0,
    PRIMARY KEY (tenant_id, date_prefix)
);

ALTER TABLE parcels 
    ADD COLUMN IF NOT EXISTS sender_customer_id UUID REFERENCES customers(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS receiver_customer_id UUID REFERENCES customers(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS price NUMERIC(10, 2);
```

---

## 4. Security & Multi-Tenancy Architecture
- **Tenant Scoping:** All customer queries strictly enforce `WHERE tenant_id = $1 AND deleted_at IS NULL`.
- **RBAC Role Matrix:**
  - Mutation (`POST`, `PATCH`, `DELETE`): Restricted to `TENANT_ADMIN`, `TENANT_OPERATOR`, `PLATFORM_ADMIN`, `COMPANY_ADMIN`.
  - Read (`GET`): Available to authenticated members (`STAFF`, `DRIVER`, `VIEWER`).
- **Cross-Tenant Isolation:** Accessing a customer or customer parcels belonging to another tenant returns `404 Not Found` or `403 Forbidden`, verified by integration tests.
- **SQL Injection & XSS Protection:** Parameterized SQL queries with PGX; HTML inputs sanitized and encoded in React DOM.

---

## 5. Modified and Created Files
| Module | File | Change Description |
|---|---|---|
| DB Migration | `backend/migrations/00009_create_customer_management.sql` | Created tables `customers`, `tenant_customer_sequences`, altered `parcels` |
| Backend Domain | `backend/internal/customers/model.go` | Customer domain structs, request/response models, validation rules |
| Backend Domain | `backend/internal/customers/repository.go` | PGX implementation, atomic customer code generator, queries |
| Backend Domain | `backend/internal/customers/service.go` | Business logic, code generation, customer profile management |
| Backend Domain | `backend/internal/customers/handler.go` | HTTP handlers for customer CRUD and parcel listings |
| Backend Domain | `backend/internal/customers/model_test.go` | Unit tests for customer validation logic |
| Backend Service | `backend/internal/parcels/model.go` | Added customer IDs, price, and pricing formula |
| Backend Service | `backend/internal/parcels/repository.go` | Added customer ID filters and queries to parcel storage |
| Backend Service | `backend/internal/parcels/service.go` | Customer auto-population and price calculation |
| Backend Service | `backend/internal/parcels/handler.go` | Added customer filter and customer parcels endpoint |
| Backend Server | `backend/internal/server/router.go` | Wired customer routes into API router |
| Backend API | `backend/cmd/api/main.go` | Dependency injection wiring for customer module |
| Backend Test | `backend/tests/integration/customer_booking_integration_test.go` | Customer CRUD & booking integration tests |
| Backend Test | `backend/tests/integration/migration_test.go` | Updated migration tests for migration 00009 |
| Frontend Types | `frontend/src/types/customers.ts` | Customer TypeScript types and payload interfaces |
| Frontend Types | `frontend/src/types/parcels.ts` | Added customer IDs and price to Parcel & payloads |
| Frontend API | `frontend/src/services/api.ts` | Customer CRUD and parcels client methods |
| Frontend UI | `frontend/src/components/customers/CustomerModal.tsx` | Customer creation/edit modal dialog |
| Frontend UI | `frontend/src/components/customers/CustomerList.tsx` | Customer CRM management dashboard |
| Frontend UI | `frontend/src/components/parcels/ParcelList.tsx` | Customer selectors, pricing card, price display |
| Frontend UI | `frontend/src/components/parcels/ParcelLabelModal.tsx` | Download QR with Destination & Parcel ID |
| Frontend Nav | `frontend/src/App.tsx` | Added Customers & CRM tab and view routing |
| Frontend Test | `frontend/test/phase4_customers_booking.test.ts` | 5 unit tests for customers and pricing |
