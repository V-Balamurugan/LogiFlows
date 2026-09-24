# LogiFlows Phase 4 Master Test Plan

## 1. Overview
The Phase 4 test suite covers the complete parcel and delivery management lifecycle across the database, backend services, REST API routes, React web UI, and Flutter mobile applications.

## 2. Test Strategy & Layers

### 2.1 Database & Migration Testing (`backend/tests/integration/migration_test.go`)
- Verifies migration `00008_create_parcel_delivery_lifecycle.sql` applies cleanly onto Phase 3 schema.
- Verifies all 8 tables, partial unique index `idx_active_parcel_delivery_task`, and foreign keys are created.
- Verifies database rollback down to previous version and re-upgrade.

### 2.2 Backend Domain & Integration Testing (`backend/tests/integration/parcel_delivery_lifecycle_integration_test.go`)
- **Parcel CRUD**: Creation, auto-generated unique tracking numbers (`PKG-YYYYMMDD-<tenant>-<seq>`), filtering, tenant isolation.
- **FSM Engine**: 15-state validation, rejection of illegal status transitions, status audit history recording.
- **Delivery Dispatch & Concurrency**: Task creation, assignment to driver/vehicle, conflict prevention on concurrent dispatch (HTTP 409).
- **Delivery Attempts & Rescheduling**: Logging attempt failures with standardized reasons.
- **Proof of Delivery (POD)**: Recipient signature/OTP verification, atomic status update to `DELIVERED`.
- **Linehaul Transfers**: Multi-parcel manifests, dispatch into transit, receiving at destination hub with bulk custody update.
- **QR / Barcode Verification**: Authenticating tracking tokens, detecting invalid codes, recording scan custody events.
- **Public Customer Tracking**: Verifying public route access without authentication and asserting zero PII leakage.

### 2.3 React Web Frontend Testing (`frontend/test/phase4_parcels.test.ts`)
- Validation of parcel booking forms and payload constraints.
- Status badge mapping and FSM eligibility rules.
- Delivery task dispatch payload structure and conflict handling.
- Proof of delivery form requirements (recipient name, signature note).
- Customer tracking milestone progression and PII sanitization.
- Production build verification (`npm run build`) and strict lint validation (`npm run lint`).

### 2.4 Flutter Mobile Testing (`mobile/test/phase4_parcel_delivery_test.dart`)
- Deserialization and serialization of `ParcelModel`, `DeliveryTaskModel`, `DeliveryAttemptModel`, `BranchTransferModel`, `PublicTrackingModel`.
- Mock HTTP verification of all Phase 4 `ResourceApiClient` methods.
- Widget tests: `ParcelDeliveryScreen` tabs rendering, delivery task cards, attempt modal, proof of delivery submission, and barcode scanner interface.
- Static analysis (`flutter analyze`) and full regression test execution.

---

## 3. Test Execution Summary

| Test Suite | Tests Executed | Passed | Failed | Execution Time |
| :--- | :--- | :--- | :--- | :--- |
| **Go Backend (All 25 pkgs)** | 100+ tests | 100% | 0 | 64.3s |
| **React Web Tests** | 23 tests (5 suites) | 23 | 0 | 679ms |
| **Flutter Mobile Tests** | 69 tests | 69 | 0 | 49s |
| **Migration Up/Down Test** | Integrated | 100% | 0 | In suite |
| **Total** | **190+ tests** | **100%** | **0** | **Clean** |
