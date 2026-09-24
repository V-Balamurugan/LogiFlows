# Phase 4 Definition of Done (DoD) Checklist

## Phase Goal
Complete the Customers and Parcel Booking vertical slice with validated APIs, database integration, frontend workflow, tests, and documentation.

---

## Quality Gate Verification Matrix

### 1. Code Quality & Standards
- [x] Code follows the existing Go backend conventions and patterns.
- [x] Code adheres to TypeScript + React best practices with strong types.
- [x] `gofmt -l .` reports zero unformatted files across `backend/`.
- [x] `oxlint` reports zero errors across `frontend/`.
- [x] No debug code, leftover console statements, or hardcoded credentials remain.

### 2. Database Integration & Schema
- [x] Migration `00009_create_customer_management.sql` successfully applied up.
- [x] Both `customers` and `tenant_customer_sequences` tables created with appropriate constraints.
- [x] `parcels` table updated with `sender_customer_id`, `receiver_customer_id`, and `price`.
- [x] Database migration rollback and re-apply verified by automated integration test `TestMigrations_RollbackAndReapply`.
- [x] Atomic customer sequence generation (`CUST-YYYYMMDD-XXXX`) prevents duplicate customer numbers under concurrency.

### 3. Backend APIs & Authorization
- [x] All customer REST endpoints implemented: `POST`, `GET`, `PATCH`, `DELETE`.
- [x] Route-level and resource-level authorization enforced for all operations.
- [x] Multi-tenant isolation verified (cross-tenant access returns 404/403).
- [x] Dynamic pricing engine calculates transparent rates based on service type, weight, and insurance.
- [x] Automatic address and contact auto-population from verified customer profiles.
- [x] Customer parcel listing endpoint (`GET /customers/:id/parcels`) operational.

### 4. Frontend Integration & UI Workflow
- [x] Customer Management & CRM tab wired into primary navigation.
- [x] Responsive Customer directory (`CustomerList.tsx`) with search, filter, and pagination.
- [x] Customer registration/editing modal (`CustomerModal.tsx`) with client-side validation.
- [x] Parcel booking workflow enhanced with customer auto-fill dropdowns.
- [x] Real-time dynamic tariff & insurance estimation card active in booking modal.
- [x] QR code download includes standalone option with Destination Hub, City, Parcel ID, and Tracking Number.
- [x] Production build passes cleanly (`npm run build`).

### 5. Automated Testing & Verification
- [x] Backend unit tests pass (100%).
- [x] Backend integration tests pass (`go test -v ./tests/integration/...` 100% PASS).
- [x] Frontend unit tests pass (31 / 31 tests PASS).
- [x] Regression testing for previous phases (Phase 0, 1, 2, 3) confirmed without breakage.

### 6. Documentation & Evidence
- [x] Baseline audit recorded in `docs/phases/phase-4/baseline-audit.md`.
- [x] User stories and acceptance criteria in `docs/phases/phase-4/requirements.md`.
- [x] Implementation report in `docs/phases/phase-4/implementation-report.md`.
- [x] Test report in `docs/phases/phase-4/test-report.md`.
- [x] Defect log in `docs/phases/phase-4/defect-log.md`.
- [x] Definition of Done verified in `docs/phases/phase-4/definition-of-done.md`.

---

## Sign-Off
Phase 4 — Customers and Parcel Booking has met all Definition of Done criteria and is ready for release.
