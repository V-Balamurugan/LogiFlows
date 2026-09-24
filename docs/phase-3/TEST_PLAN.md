# Phase 3 Test Plan — Operational Resource Management

**Project**: LogiFlows  
**Phase**: 3 — Employees, Roles, and Vehicles  
**Target Branch**: `feature/phase-3-employees-roles-vehicles`  

---

## 1. Objectives

Verify the complete, secure, and performant implementation of Phase 3 operational resource management across:
1. Multi-tenant employee profile lifecycle, unique code auto-generation, operational roles, KYC verification, and availability state tracking.
2. Vehicle fleet registration, physical weight/volume constraints, branch allocation, and operational status management.
3. Driver-vehicle assignment rules, eligibility criteria, mutual availability enforcement, atomic transitions, and concurrency race-condition prevention.
4. Multi-tenant data isolation and branch-level scoping.
5. Zero regressions across Phase 0, Phase 1, and Phase 2 features.
6. React web frontend integration and Flutter mobile app models/screens.

---

## 2. Test Scope & Categorization

| Tier | Category | Target Packages / Directories | Execution Tool |
|---|---|---|---|
| **Tier 1** | Backend Unit Tests | `backend/internal/employees`<br>`backend/internal/vehicles`<br>`backend/internal/server` | `go test -v ./internal/...` |
| **Tier 2** | Database Migrations | `backend/tests/integration/migration_test.go` | `go test -v ./tests/integration -run TestMigrations` |
| **Tier 3** | Concurrency & Race Conditions | `backend/tests/integration/phase3_concurrency_test.go` | `go test -v ./tests/integration -run TestPhase3` |
| **Tier 4** | End-to-End Regression | `backend/tests/regression/...` (Phases 0, 1, 2, 3) | `go test -v ./tests/regression/...` |
| **Tier 5** | Frontend Unit & Type Tests | `frontend/test/**/*.test.ts`<br>`frontend/src/` | `npm test`<br>`npm run build`<br>`npm run lint` |
| **Tier 6** | Mobile Dart Unit Tests | `mobile/test/resource_models_test.dart` | Static analysis / `flutter test` |

---

## 3. Test Environment & Fixtures

- **PostgreSQL**: PostgreSQL 16 + PostGIS running via Docker on port 5432 (`logiflows_test` / `postgres`).
- **Redis**: Redis 7 running via Docker on port 6379.
- **Database Schema**: Goose migrations 00001 through 00007 applied.
- **Test Tenants**:
  - Tenant A: `00000000-0000-0000-0000-000000000001`
  - Tenant B: `00000000-0000-0000-0000-000000000002`

---

## 4. Test Scenarios Matrix

### 4.1 Employee Management
- `TC-EMP-01`: Create employee with explicit employee code (e.g. `EMP-0012`).
- `TC-EMP-02`: Create employee with auto-generated code (blank `employee_code`) -> verified sequence output `EMP-0001`, `EMP-0002`.
- `TC-EMP-03`: Concurrent creation of 10 employees within the same tenant generates 10 unique, non-colliding employee codes.
- `TC-EMP-04`: Create employee with branch belonging to another tenant -> Rejected with 400 Bad Request.
- `TC-EMP-05`: Duplicate employee code within the same tenant -> Rejected with 409 Conflict.
- `TC-EMP-06`: Update employee operational availability (`AVAILABLE` -> `OFF_DUTY` -> `AVAILABLE`).
- `TC-EMP-07`: Soft-deactivate employee (`status = 'TERMINATED'`, `is_active = false`). Verify exclusion from active lists while remaining queryable by ID.

### 4.2 Fleet & Vehicle Management
- `TC-VEH-01`: Register vehicle with valid payload (positive `max_weight_kg`, positive `max_volume_cbm`).
- `TC-VEH-02`: Register vehicle with invalid capacity (negative weight or zero volume) -> Rejected with 400 Bad Request.
- `TC-VEH-03`: Duplicate registration number within the same tenant -> Rejected with 409 Conflict.
- `TC-VEH-04`: Update vehicle operational status (`AVAILABLE` -> `MAINTENANCE` -> `AVAILABLE`).
- `TC-VEH-05`: Decommission vehicle (`DELETE`). Verify soft-deactivation.

### 4.3 Driver-Vehicle Assignment & Eligibility
- `TC-ASG-01`: Query `/employees/available-drivers` returns only verified, active drivers with `availability_status = 'AVAILABLE'`.
- `TC-ASG-02`: Assign eligible driver to available vehicle -> HTTP 201 Created. Driver becomes `BUSY`, vehicle becomes `ASSIGNED`.
- `TC-ASG-03`: Assign ineligible employee (`operational_role = 'OPERATOR'`) -> Rejected with 400 Bad Request (`DRIVER_INELIGIBLE`).
- `TC-ASG-04`: Assign unavailable driver (`availability_status = 'BUSY'`) -> Rejected with 400 Bad Request.
- `TC-ASG-05`: Assign vehicle already in `ASSIGNED` state -> Rejected with 409 Conflict (`DOUBLE_BOOKING`).
- `TC-ASG-06`: Concurrent assignment race condition (10 simultaneous goroutines attempting assignment of the same vehicle and driver) -> Exactly 1 success (HTTP 201) and 9 conflicts (HTTP 409).
- `TC-ASG-07`: Unassign vehicle -> Both vehicle and driver restored to `AVAILABLE`. Assignment status marked `TERMINATED`.
- `TC-ASG-08`: Assignment history accurately logs start time and unassign time.

---

## 5. Pass/Fail Criteria

1. **Zero compilation, vet, or lint errors** in Go and TypeScript.
2. **100% pass rate** across all backend unit, integration, and regression suites.
3. **100% pass rate** in frontend unit tests and production bundle build.
4. **No regressions** in Phase 0, Phase 1, or Phase 2 APIs.
5. All security boundaries (cross-tenant, role escalation, double-booking) strictly enforced.
