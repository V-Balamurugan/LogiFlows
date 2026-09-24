# Phase 4 Defect Log — Customers and Parcel Booking

## Defect Summary
- **Total Defects Identified:** 3
- **Resolved Defects:** 3
- **Unresolved Defects:** 0
- **Blockers:** 0

---

## Logged Defects and Resolutions

### DEF-P4-001: Unused `github.com/google/uuid` Import in Customer Integration Test
- **Severity:** Medium (Compilation Blocker)
- **Component:** `backend/tests/integration/customer_booking_integration_test.go`
- **Symptom:** `go test` failed to compile with error: `"github.com/google/uuid" imported and not used`.
- **Root Cause:** UUID generation was handled in helper methods and domain model without direct test call to `uuid.New()`.
- **Resolution:** Removed the unused import line from `customer_booking_integration_test.go`.
- **Status:** RESOLVED & VERIFIED.

---

### DEF-P4-002: Migration Test Rollback Assertion Failure on Migration 9
- **Severity:** Medium (CI Test Blocker)
- **Component:** `backend/tests/integration/migration_test.go`
- **Symptom:** `TestMigrations_RollbackAndReapply` failed with `expected tenant_parcel_sequences table to be dropped after rollback`.
- **Root Cause:** The rollback test previously hardcoded checks for migration 8 (`tenant_parcel_sequences`). When migration 9 was introduced as the latest migration, `migrations.RunDown` rolled back 1 step (from version 9 to 8), meaning `tenant_parcel_sequences` correctly remained in place while `tenant_customer_sequences` was dropped.
- **Resolution:** Updated `TestMigrations_RollbackAndReapply` to verify the latest migration (migration 9) by asserting `tenant_customer_sequences` is dropped on rollback and recreated on reapply.
- **Status:** RESOLVED & VERIFIED.

---

### DEF-P4-003: Redundant Catch Clause in Frontend Customer Management
- **Severity:** Low (Linter Quality Warning)
- **Component:** `frontend/src/components/customers/CustomerList.tsx`
- **Symptom:** `oxlint` flagged `eslint(no-useless-catch): Unnecessary catch clause` in `handleModalSubmit`.
- **Root Cause:** A `try { ... } catch (err: any) { throw err; } finally { ... }` block re-threw the error without adding handling logic.
- **Resolution:** Removed the redundant catch clause, leaving the `finally` block for modal loading cleanup.
- **Status:** RESOLVED & VERIFIED.
