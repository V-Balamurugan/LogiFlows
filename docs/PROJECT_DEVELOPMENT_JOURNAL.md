# LogiFlows Master Project Development Journal

## 1. Project Overview & Architecture
LogiFlows is an enterprise multi-tenant logistics coordination and delivery management platform. The system coordinates end-to-end parcel custody, distribution hubs, workforce allocation, and commercial fleet scheduling across independent logistics operators and merchant organizations.

### High-Level Tech Stack
- **Backend**: Go 1.24+ with Gin Web Framework, PostgreSQL driver (`pgx`), Redis client (`go-redis`), JWT auth.
- **Database**: PostgreSQL 16 with PostGIS 3.4 spatial extensions and Redis 7.2 distributed caching.
- **Web Frontend**: React 18 / 19, TypeScript 5.8, Vite 8, Tailwind CSS, Lucide Icons.
- **Mobile Client**: Flutter 3 / Dart 3 with Material 3 design and offline-safe secure token storage.
- **AI Microservice**: Python 3.12, FastAPI, PyTorch / Scikit-learn for parcel route optimization.

---

## 2. Phase Breakdown & Milestones

| Phase | Milestone Name | Key Deliverables | Status |
|-------|----------------|------------------|--------|
| **Phase 0** | Foundation & Infrastructure | Project skeleton, Docker compose, CI/CD, database migrations runner, health probes, structured logging. | **COMPLETE** |
| **Phase 1** | Identity & Multi-Tenancy | Multi-tenant data model, user registration, JWT auth, refresh token rotation, RBAC, React Auth UI. | **COMPLETE** |
| **Phase 2** | Multi-Branch, Workforce & Fleet Management | PostGIS spatial branches, workforce onboarding, fleet management, vehicle-driver scheduling, web & mobile UI. | **COMPLETE** |
| **Phase 3** | Employees, Roles, and Vehicles | Operational workforce management, atomic `EMP-XXXX` code sequence generation, operational roles, KYC verification, fleet vehicle capacity constraints, driver eligibility & assignment foundation, concurrency double-booking prevention, React & Flutter integration. | **COMPLETE** |
| **Phase 4** | Parcel Custody & Real-Time Tracking | Parcel state machine, barcode/QR custody transfers, real-time GPS websocket tracking, proof of delivery. | Planned |
| **Phase 5** | AI Optimization & Dispatch Heuristics | Machine learning delivery ETA, vehicle bin-packing, dynamic route optimization. | Planned |

---

## 3. Phase 0 Milestone Summary
- **Database**: Initialized PostgreSQL with `uuid-ossp` and created migration tracking table.
- **Backend**: Implemented configuration loading via Viper, structured logger with Zerolog, Gin router with CORS, recovery, and request tracing middlewares.
- **Health Checks**: Implemented `/health/live` and `/health/ready` probing database and Redis connectivity.
- **Documentation**: Established `docs/project-journal/phase-0.md`.

---

## 4. Phase 1 Milestone Summary
- **Database Migration**: `00002_create_identity_and_tenancy.sql` established `users`, `tenants`, `tenant_members`, and `audit_logs`.
- **Backend Services**: Implemented bcrypt password hashing, HMAC-SHA256 JWT generation, refresh token rotation with single-use revocation, and tenant isolation middleware.
- **Frontend UI**: Built React + TypeScript authentication screens (Login, Registration, Tenant switcher).
- **Security Audit**: Verified zero cross-tenant leakage (`403 Forbidden` on foreign tenant requests).
- **Documentation**: Created `docs/project-journal/phase-1.md` and detailed test execution reports.

---

## 5. Phase 2 Milestone Summary: Multi-Tenant Companies and Distribution Branches
- **Database Migrations & Spatial Models**:
  - `00001_create_users_and_tenants.sql`: Foundation multi-tenant organizations (`tenants`), user identity, and role memberships.
  - `00004_create_branches.sql`: Distribution branches with native PostGIS spatial points (`GEOMETRY(Point, 4326)`), coverage radii, GIST spatial indexes, and tenant-scoped compound uniqueness `(tenant_id, branch_code)`.
  - Foundation dependencies (`00005_create_employees.sql`, `00006_create_vehicles_and_assignments.sql`).
- **Backend Architecture & Routes**:
  - Router aliasing: Seamless interoperability supporting both `/api/v1/companies` and `/api/v1/tenants` prefixes with `RequireTenantContext` middleware.
  - Organization profile: `GET /api/v1/companies/current` returning current tenant details, metadata, and status.
  - Branch operating status: `PATCH /api/v1/companies/:id/branches/:id/status` validating operational state transitions (`ACTIVE`, `INACTIVE`, `SUSPENDED`).
  - Spatial proximity search: PostGIS `ST_DWithin` and `ST_Distance` calculations with geodesic accuracy.
  - Parameterized SQL queries: 100% SQL injection immunity via `pgx/v5`.
- **Web Frontend (React 19 + TypeScript + Vite)**:
  - `CompanyProfile.tsx`: Organization details, status badge, metadata editing form, compliance checklist, and team modal.
  - `BranchList.tsx`: Distribution hub directory with PostGIS coordinate pills, status toggles, and branch creation modal.
  - `src/services/api.ts`: Centralized typed API client with token injection and error mapping.
  - Production build: 1887 modules compiled with 0 errors. Vitest component tests passed. ESLint passed with 0 errors.
- **Mobile Client (Flutter 3 / Dart)**:
  - `CompanyModel` and `BranchModel`: Strongly-typed Dart domain models with full JSON serialization.
  - `ResourceApiClient`: Configurable base URL, Bearer JWT injection, and spatial parameter builders.
  - `CompanyScreen` and `BranchScreen`: Dark Material 3 mobile screens with pull-to-refresh and error recovery.
  - `main.dart`: Integrated bottom navigation bar with `Company`, `Hubs`, `Fleet`, and `Custody`.
  - Unit tests: `company_models_test.dart` and `resource_models_test.dart`.

---

## 6. Phase 3 Milestone Summary: Employees, Accounts, Roles, and Operations

- **Database Migrations & Models**:
  - `00007_phase3_operational_resources.sql`: Added `availability_status` (`AVAILABLE`, `BUSY`, `OFF_DUTY`, `UNAVAILABLE`) and `verification_status` (`PENDING`, `VERIFIED`, `REJECTED`) to `employees`. Added `availability_status` (`AVAILABLE`, `BUSY`, `MAINTENANCE`, `OUT_OF_SERVICE`) to `vehicles`. Expanded operational role check constraint to include `BRANCH_MANAGER`, `WAREHOUSE_OPERATOR`, `DELIVERY_EXECUTIVE`.
  - `tenant_employee_sequences`: Dedicated per-tenant atomic sequence generator table guaranteeing monotonically increasing non-colliding `EMP-XXXX` employee codes.
  - Partial unique indexes (`idx_active_vehicle_assignment`, `idx_active_driver_assignment`) enforcing that a vehicle and driver can each participate in at most one `ACTIVE` assignment at any given time.
  - Migration rollback & apply verified via `backend/tests/integration/migration_test.go` (100% pass).
- **Backend Architecture & APIs**:
  - **Employee Account Creation**: Transactional atomic account provisioning (`POST /api/v1/tenants/{tenant_id}/employees/with-account`). Coordinates user registration, Bcrypt (cost 12) password hashing, tenant membership creation (`EMPLOYEE` system role), and employee profile creation inside a single atomic database transaction. Automatically rolls back on duplicate email or foreign branch error. Never leaks password hashes or plain-text passwords.
  - **Account Status Inspection**: `GET /api/v1/tenants/{tenant_id}/employees/{employee_id}/account-status` returning whether an employee has a linked login account, system role, and membership status.
  - **Employee Self-Profile**: `GET /api/v1/tenants/{tenant_id}/employees/me` returning the authenticated employee's personal profile, assigned branch, and currently assigned vehicle specifications.
  - **Branch Operational Assets**: `GET /api/v1/tenants/{tenant_id}/branches/{branch_id}/employees` and `GET /api/v1/tenants/{tenant_id}/branches/{branch_id}/vehicles` for viewing stationed workforce and commercial fleet.
  - **Auto-Code Generation**: Atomic unique employee codes (`EMP-XXXX`) when blank in `POST /tenants/{tenant_id}/employees`.
  - **Availability & KYC Status Updates**: `PATCH /tenants/{tenant_id}/employees/{employee_id}/status`.
  - **Available Drivers Endpoint**: `GET /tenants/{tenant_id}/employees/available-drivers`.
  - **Vehicle Fleet Management**: Validated capacity limits, operating status updates via `PATCH /tenants/{tenant_id}/vehicles/{vehicle_id}/status`.
  - **Driver-Vehicle Assignment Foundation**:
    - Enforces `operational_role = 'DRIVER'`, active employment, and mutual availability.
    - Atomically updates driver availability to `BUSY` and vehicle status to `ASSIGNED`.
    - Handles concurrent assignment races with HTTP 409 Conflict.
    - Unassignment endpoint: `POST /tenants/{tenant_id}/vehicles/{vehicle_id}/unassign` atomically restoring both vehicle and driver to `AVAILABLE`.
    - Assignment history endpoints: `/tenants/{tenant_id}/assignments` and `/tenants/{tenant_id}/vehicles/{vehicle_id}/assignments`.
- **Phase 2 Regression Bug Fix**:
  - Resolved employee soft-delete bug in `internal/employees/repository.go` by removing `AND deleted_at IS NULL` from `GetByID` while preserving list filtering.
- **Web Frontend (React + TypeScript + Vite)**:
  - `EmployeeList.tsx`: Role badges, pulsing availability indicators, KYC verification badges, auto-code helper text, available drivers quick filter, status update modal, Account Provisioning toggle/form, and Account Status inspection modal.
  - `EmployeeSelfProfile.tsx`: Self-profile component for `EMPLOYEE` role displaying profile, branch, availability toggle, and assigned vehicle.
  - `BranchList.tsx`: Added "View Staff & Fleet" action button and modal for branch personnel and fleet assets.
  - `VehicleList.tsx`: Fleet overview, zero-emission electric badges, capacity metrics, active driver pill with driver code, real-time driver assignment modal (filtered by branch and available drivers), unassign action, and vehicle status management modal.
  - Role-based routing in `App.tsx` with "My Staff Profile" tab and role-aware navigation.
  - Automated tests: 16/16 unit tests pass across 4 suites (`npm test`). Build: Exit code 0 (`npm run build`). Lint: 0 errors (`npm run lint`).
- **Mobile Client (Flutter / Dart)**:
  - Updated models: `EmployeeModel`, `VehicleModel`, `AssignmentModel`, `EmployeeMeModel`, `EmployeeAccountStatusModel`, `BranchEmployeeSummaryModel`, `BranchVehicleSummaryModel`, `AssignedVehicleModel`.
  - Updated `ResourceApiClient` with Phase 3 endpoints.
  - Implemented `EmployeeScreen` (`employee_screen.dart`) with role filters, available drivers view, interactive availability dialog, and Self-Profile bottom sheet (`_showMyProfileSheet`).
  - Updated `VehicleScreen` (`vehicle_screen.dart`) with availability badges, dynamic driver assignment bottom sheet, and unassignment action.
  - Added Staff destination to bottom navigation in `main.dart`.
  - Documented host Flutter CLI environment limitation honestly.
- **Verification & Testing**:
  - 100% pass rate across all 19 Go backend packages (`go test ./...`), including `employee_account_integration_test.go` (3/3 pass).
  - Concurrency integration tests: 10 parallel goroutines for code auto-generation and 10 parallel assignment attempts (1 success, 9 conflicts) passed cleanly.
  - Regression suite: All 7 suites covering Phase 0, 1, 2, and 3 passed.
  - Security audit: All security scenarios passed (zero cross-tenant leaks, zero IDOR, no exposed passwords or hashes).
- **OpenAPI / Swagger**: Updated `backend/docs/swagger.json` with all Phase 3 endpoints and data models.
- **Documentation**: 14 comprehensive markdown reports generated in `docs/phase-3/`.

---

## 7. Current Repository Status & Branching
- Active Branch: `feature/phase-4-parcel-delivery-lifecycle`
- Base Branch: `feature/phase-3-employees-roles-vehicles`
- Working Tree: Clean and verified across Go backend, PostgreSQL/PostGIS database, React frontend, and Flutter mobile codebase.
- Phase 3 Final Status: **COMPLETE & PRODUCTION READY**.
- Next Target: Phase 4 Parcel Custody & Real-Time Tracking.

---

## 8. Phase 4 Milestone Summary: Parcel Custody & Real-Time Tracking

- **Kickoff & Precheck Audit**:
  - Full audit of Phase 0-3 subsystems passed with 100% success rate:
    - Backend: 21 packages tested (`go test ./...`), 0 failures.
    - Frontend: 16/16 unit tests passed, 0 lint errors, production build verified.
    - Mobile: `flutter analyze` completed with 0 errors, 58/58 unit and widget tests passed.
  - Precheck report published to `docs/phase-4/PHASE_3_PRECHECK_REPORT.md`.
- **System Architecture & Specifications**:
  - `docs/phase-4/PHASE_4_REQUIREMENTS.md`: Detailed functional and non-functional requirements.
  - `docs/phase-4/DATABASE_DESIGN.md`: Schemas for `parcels`, `parcel_status_history`, `parcel_custody_events`, `branch_transfers`, `delivery_tasks`, `delivery_attempts`, `delivery_proofs`, and `tenant_parcel_sequences`.
  - `docs/phase-4/PARCEL_STATUS_STATE_MACHINE.md`: Deterministic 15-state lifecycle model.
  - `docs/phase-4/DELIVERY_WORKFLOW.md`: Last-mile dispatch, concurrency race condition prevention, and inter-branch transfers.
  - `docs/phase-4/API_CONTRACTS_PHASE_4.md`: Complete OpenAPI specifications for parcel, delivery, transfer, and tracking endpoints.

