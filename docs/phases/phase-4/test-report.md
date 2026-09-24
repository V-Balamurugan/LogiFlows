# Phase 4 Test Report — Customers and Parcel Booking

## 1. Test Execution Summary
- **Execution Date:** 2026-09-24
- **Backend Test Status:** 100% PASS
  - Unit Tests: PASS (`backend/internal/customers`, `backend/internal/parcels`, etc.)
  - Integration Tests: PASS (`backend/tests/integration/...` 78.8s)
  - Regression Tests: PASS (`backend/tests/regression/...` 13.2s)
- **Frontend Test Status:** 100% PASS (31 passed across 6 test suites)
- **Lint / Code Quality:** PASS (0 errors, `gofmt -l .` clean, `oxlint` 0 errors)
- **Build Status:** PASS (Go `bin/api.exe` built successfully; Vite client bundled successfully)

---

## 2. Test Cases and Results

### 2.1 Backend Unit & Integration Tests
| Test ID | Module | Test Description | Expected Result | Actual Result | Status |
|---|---|---|---|---|---|
| TC-P4-CUST-UNIT-001 | `customers` | Validate individual customer creation payload | Required fields present, valid status | Success, validated | PASS |
| TC-P4-CUST-UNIT-002 | `customers` | Validate enterprise customer with tax ID & company | Required fields present, valid tax format | Success, validated | PASS |
| TC-P4-CUST-UNIT-003 | `customers` | Validate invalid phone number (< 7 chars) | Returns `ErrInvalidCustomerInput` | Error returned | PASS |
| TC-P4-CUST-UNIT-004 | `customers` | Validate empty billing address | Returns `ErrInvalidCustomerInput` | Error returned | PASS |
| TC-P4-CUST-UNIT-005 | `customers` | Validate invalid customer classification | Returns `ErrInvalidCustomerInput` | Error returned | PASS |
| TC-P4-INT-001 | `integration` | Migration 00009 apply up and verify tables | `customers` and `tenant_customer_sequences` exist, columns present in `parcels` | Schema matches version 9 | PASS |
| TC-P4-INT-002 | `integration` | Migration 00009 rollback down and reapply | Migration 9 rolled back and tables dropped; reapply brings them back | Rollback & reapply pass | PASS |
| TC-P4-INT-003 | `integration` | Create customer via API (`POST /customers`) | 201 Created with atomic code `CUST-YYYYMMDD-XXXX` | 201 Created, code assigned | PASS |
| TC-P4-INT-004 | `integration` | Reject customer with missing name or phone | 400 Bad Request with descriptive message | 400 Bad Request | PASS |
| TC-P4-INT-005 | `integration` | Text search across customers | Matching customers returned | Filtered records returned | PASS |
| TC-P4-INT-006 | `integration` | Update customer profile (`PATCH /customers/:id`) | 200 OK with updated attributes | 200 OK, attributes updated | PASS |
| TC-P4-INT-007 | `integration` | Parcel booking with linked customer IDs | 201 Created, auto-fills address, computes price | 201 Created, price calculated | PASS |
| TC-P4-INT-008 | `integration` | List customer parcels (`GET /customers/:id/parcels`) | 200 OK with matching parcels | 200 OK, parcels returned | PASS |
| TC-P4-INT-009 | `integration` | Cross-tenant customer isolation attempt | 404 Not Found (tenant isolation enforced) | 404 Not Found | PASS |
| TC-P4-INT-010 | `integration` | Full Phase 0–3 regression suite | Auth, Branches, Employees, Vehicles, Lifecycle | All suites pass | PASS |

### 2.2 Frontend Unit Tests
| Test ID | Suite | Test Description | Expected Result | Actual Result | Status |
|---|---|---|---|---|---|
| TC-FE-CUST-001 | `phase4_customers_booking` | Validate customer individual payload | Validates required fields | Asserts pass | PASS |
| TC-FE-CUST-002 | `phase4_customers_booking` | Validate enterprise customer with tax ID | Validates enterprise details | Asserts pass | PASS |
| TC-FE-CUST-003 | `phase4_customers_booking` | Dynamic pricing formula calculation | Computes Standard, Express, Overnight, Same-Day + Insurance | Exact match to Go formula | PASS |
| TC-FE-CUST-004 | `phase4_customers_booking` | Parcel booking customer linkage | Asserts sender/receiver customer IDs | Asserts pass | PASS |
| TC-FE-CUST-005 | `phase4_customers_booking` | Customer classification & status invariants | Enforces all 4 types and 3 statuses | Asserts pass | PASS |
| TC-FE-PARCEL-001 | `phase4_parcels` | Create parcel payload validation | Origin, destination, weight required | Asserts pass | PASS |
| TC-FE-PARCEL-002 | `phase4_parcels` | Parcel status state machine transitions | Legal enum values enforced | Asserts pass | PASS |
| TC-FE-PARCEL-008 | `phase4_parcels` | QR code payload validation | Tracking, tenant, parcel UUID | Asserts pass | PASS |
| TC-FE-PARCEL-009 | `phase4_parcels` | Shipping label includes Destination & ID | Destination, recipient, and UUID | Asserts pass | PASS |

---

## 3. Negative Scenarios Verified
1. **Invalid Phone Number:** Rejected with `ErrInvalidCustomerInput` ("phone must be at least 7 characters").
2. **Missing Billing Address:** Rejected with `ErrInvalidCustomerInput` ("billing address is required").
3. **Invalid Customer Type:** Rejection of non-whitelisted strings.
4. **Cross-Tenant Customer Access:** Accessing a customer under Tenant B while authenticated under Tenant A returns `404 Not Found`.
5. **Cross-Tenant Customer Parcel Linking:** Linking a customer belonging to Tenant B when booking under Tenant A triggers validation error or ignores foreign profile.
6. **Partial Parcel Writes:** Database transactions ensure parcels and status histories are atomically committed.
7. **Migration Rollback Integrity:** `TestMigrations_RollbackAndReapply` confirms goose down safely drops sequence tables and re-up recreates them without data corruption.

---

## 4. Test Commands Executed
```bash
# Go Unit & Integration Tests
cd backend
gofmt -l .
go test -v -run TestCustomer_CRUD_And_ParcelBooking_Integration ./tests/integration/...
go test -v -run TestMigrations_RollbackAndReapply ./tests/integration/...
go test -v ./tests/integration/...
go build -v -o bin/api.exe ./cmd/api

# Frontend Tests & Build
cd ../frontend
npm test
npm run lint
npm run build
```
