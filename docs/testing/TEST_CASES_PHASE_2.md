# LogiFlows — Phase 2 Test Case Catalog & Execution Report

**Document Reference**: `docs/testing/TEST_CASES_PHASE_2.md`  
**Phase**: Phase 2 — Organization, Branch, Employee, and Fleet Management  
**Status**: **100% COMPLETE & VERIFIED**  
**Execution Date**: 2026-09-22  
**Database Schema**: Goose Migrations 00001 through 00006 Applied  

---

## 1. Summary of Execution Results

| Module | Automated Tests | Passed | Failed | Status |
| :--- | :---: | :---: | :---: | :---: |
| **Module B: Branch Management** | 6 Integration + 4 Unit | 10 | 0 | **PASS** |
| **Module C: Employee Management** | 7 Integration + 5 Unit | 12 | 0 | **PASS** |
| **Module D: Roles & Branch Access** | 4 Integration + 3 Regression | 7 | 0 | **PASS** |
| **Module E & F: Vehicle & Fleet Assignment** | 6 Integration + 3 Unit + 4 Regression | 13 | 0 | **PASS** |
| **Frontend UI Integration** | 9 Unit + TypeScript Build (0 errors) | 10 | 0 | **PASS** |
| **TOTAL** | **52 Tests** | **52** | **0** | **100% PASS** |

---

## 2. Test Case Specifications & Execution Evidence

### TC-P2-BRN-001: Create Delivery Branch with PostGIS Coordinates
* **Phase**: Phase 2 | **Module**: Branches | **Feature**: Branch Provisioning
* **Test Type**: Integration / Spatial | **Priority**: P1 | **Severity**: Major | **Automated**: Yes
* **Preconditions**: Authenticated user with `TENANT_ADMIN` role; valid active tenant.
* **Test Data**:
  ```json
  {
    "branch_code": "DEL-HUB-01",
    "name": "Delhi North Hub",
    "address": "Plot 42, GT Karnal Road",
    "city": "New Delhi",
    "country": "India",
    "latitude": 28.7041,
    "longitude": 77.1025,
    "coverage_radius_km": 25.0
  }
  ```
* **Test Steps**: Dispatch `POST /api/v1/tenants/:tenant_id/branches`.
* **Expected Result**: HTTP 201 Created; branch stored in database with valid PostGIS point geometry (`ST_SetSRID(ST_MakePoint(lng, lat), 4326)`).
* **Actual Result**: HTTP 201 Created; branch record persisted with PostGIS spatial point and accurate lat/lng response.
* **Verification Test**: `TestBranch_Create_WithPostGIS` in `backend/tests/integration/branch_integration_test.go`
* **Status**: **PASS**

---

### TC-P2-BRN-002: Reject Duplicate Branch Code Within Same Tenant
* **Phase**: Phase 2 | **Module**: Branches | **Feature**: Branch Code Uniqueness
* **Test Type**: API / Constraint | **Priority**: P1 | **Severity**: Moderate | **Automated**: Yes
* **Preconditions**: Branch with code `BR-DUP-01` exists in current tenant.
* **Test Steps**: Attempt to create second branch with identical `branch_code` within same tenant.
* **Expected Result**: HTTP 409 Conflict; message indicating branch code already exists.
* **Actual Result**: HTTP 409 Conflict returned; second creation rejected cleanly.
* **Verification Test**: `TestBranch_DuplicateCode_RejectedWithinTenant` in `backend/tests/integration/branch_integration_test.go`
* **Status**: **PASS**

---

### TC-P2-BRN-003: Allow Identical Branch Code in Different Tenants
* **Phase**: Phase 2 | **Module**: Branches | **Feature**: Tenant Isolation & Scoped Uniqueness
* **Test Type**: Multi-Tenant / Constraint | **Priority**: P1 | **Severity**: Major | **Automated**: Yes
* **Preconditions**: Tenant Alpha has branch with code `CENTRAL-01`.
* **Test Steps**: Tenant Beta creates branch with identical code `CENTRAL-01`.
* **Expected Result**: HTTP 201 Created; composite unique index `(tenant_id, branch_code)` allows identical codes across different tenants.
* **Actual Result**: HTTP 201 Created in Tenant Beta.
* **Verification Test**: `TestBranch_IdenticalCode_AllowedInDifferentTenants` in `backend/tests/integration/branch_integration_test.go`
* **Status**: **PASS**

---

### TC-P2-BRN-004: Enforce Multi-Tenant Isolation on Branch Retrieval
* **Phase**: Phase 2 | **Module**: Branches | **Feature**: Multi-Tenant Isolation
* **Test Type**: Security / IDOR | **Priority**: P0 | **Severity**: Critical | **Automated**: Yes
* **Preconditions**: User belongs to Tenant Alpha; Branch belongs to Tenant Beta.
* **Test Steps**: User attempts `GET /api/v1/tenants/<tenant-beta-id>/branches/<branch-id>`.
* **Expected Result**: HTTP 403 Forbidden; error code `CROSS_TENANT_ACCESS_DENIED`.
* **Actual Result**: HTTP 403 Forbidden returned by `TenantContext` middleware.
* **Verification Test**: `TestBranch_CrossTenantAccess_Forbidden` in `backend/tests/integration/branch_integration_test.go`
* **Status**: **PASS**

---

### TC-P2-BRN-005: Branch Spatial Filtering, Pagination, and Soft-Deletion
* **Phase**: Phase 2 | **Module**: Branches | **Feature**: Spatial Queries & Lifecycle
* **Test Type**: Spatial / API | **Priority**: P2 | **Severity**: Moderate | **Automated**: Yes
* **Preconditions**: 3 branches registered at different geographic coordinates.
* **Test Steps**:
  1. `GET /branches?near_lat=28.70&near_lng=77.10&radius_km=45` (Spatial distance search).
  2. `PATCH /branches/:id` (Update metadata).
  3. `DELETE /branches/:id` (Soft-delete).
  4. `GET /branches/:id` (Verify `is_active=false` and `status=INACTIVE`).
* **Expected Result**: Spatial query returns branches sorted by distance with `distance_km`; update succeeds; soft-delete preserves historical data without physical row removal.
* **Actual Result**: Spatial query calculates distance using `ST_DistanceSphere`; soft-delete marks `is_active=false`.
* **Verification Test**: `TestBranch_List_And_SpatialFilter` in `backend/tests/integration/branch_integration_test.go`
* **Status**: **PASS**

---

### TC-P2-EMP-001: Register Employee Profile with Operational Role and Branch
* **Phase**: Phase 2 | **Module**: Employees | **Feature**: Employee Management
* **Test Type**: API / Integration | **Priority**: P1 | **Severity**: Major | **Automated**: Yes
* **Preconditions**: Authenticated `TENANT_ADMIN`; valid branch exists in tenant.
* **Test Data**:
  ```json
  {
    "employee_code": "DRV-101",
    "first_name": "Arjun",
    "last_name": "Patel",
    "email": "arjun.patel@logiflows.test",
    "designation": "Delivery Driver",
    "employment_type": "FULL_TIME",
    "operational_role": "DRIVER",
    "license_number": "DL-04-2022-9999",
    "branch_id": "<branch-uuid>"
  }
  ```
* **Test Steps**: Dispatch `POST /api/v1/tenants/:tenant_id/employees`.
* **Expected Result**: HTTP 201 Created; employee record linked to tenant and branch.
* **Actual Result**: HTTP 201 Created; employee code and driver operational role persisted.
* **Verification Test**: `TestEmployee_Create_WithValidRoleAndBranch` in `backend/tests/integration/employee_integration_test.go`
* **Status**: **PASS**

---

### TC-P2-EMP-002: Reject Employee Assignment to Foreign Tenant's Branch
* **Phase**: Phase 2 | **Module**: Employees | **Feature**: Multi-Tenant Foreign Key Integrity
* **Test Type**: Security / Business Rules | **Priority**: P0 | **Severity**: Critical | **Automated**: Yes
* **Preconditions**: Branch belongs to Tenant Beta; Employee creation initiated in Tenant Alpha.
* **Test Steps**: Dispatch `POST /api/v1/tenants/<tenant-alpha-id>/employees` with Tenant Beta's `branch_id`.
* **Expected Result**: HTTP 400 Bad Request; message indicates branch does not belong to organization.
* **Actual Result**: HTTP 400 Bad Request returned; cross-tenant branch association prevented.
* **Verification Test**: `TestEmployee_CrossTenantBranch_Rejected` in `backend/tests/integration/employee_integration_test.go`
* **Status**: **PASS**

---

### TC-P2-EMP-003: Prevent Duplicate Employee Code Within Same Tenant
* **Phase**: Phase 2 | **Module**: Employees | **Feature**: Unique Constraint
* **Test Type**: Constraint | **Priority**: P1 | **Severity**: Moderate | **Automated**: Yes
* **Preconditions**: Employee with code `EMP-001` exists in tenant.
* **Test Steps**: Create second employee in same tenant with code `EMP-001`.
* **Expected Result**: HTTP 409 Conflict.
* **Actual Result**: HTTP 409 Conflict returned; database unique constraint `uq_tenant_employee_code` prevents duplicates.
* **Verification Test**: `TestEmployee_DuplicateCode_RejectedWithinTenant` in `backend/tests/integration/employee_integration_test.go`
* **Status**: **PASS**

---

### TC-P2-EMP-004: Soft-Deactivation and Lifecycle of Employee
* **Phase**: Phase 2 | **Module**: Employees | **Feature**: Employee Lifecycle
* **Test Type**: API | **Priority**: P2 | **Severity**: Moderate | **Automated**: Yes
* **Preconditions**: Active employee exists.
* **Test Steps**:
  1. `PUT /employees/:id` (Update designation).
  2. `DELETE /employees/:id` (Deactivate).
  3. `GET /employees/:id` (Verify state).
* **Expected Result**: Designation updated; deactivation sets `is_active=false` and `status=TERMINATED`.
* **Actual Result**: HTTP 200 on update and delete; GET verifies `is_active=false`.
* **Verification Test**: `TestEmployee_Update_And_SoftDeactivate` in `backend/tests/integration/employee_integration_test.go`
* **Status**: **PASS**

---

### TC-P2-VEH-001: Register Fleet Vehicle with Payload and Volume Capacities
* **Phase**: Phase 2 | **Module**: Vehicles | **Feature**: Fleet Registration
* **Test Type**: API / Integration | **Priority**: P1 | **Severity**: Major | **Automated**: Yes
* **Preconditions**: Authenticated `TENANT_ADMIN` or `TENANT_OPERATOR`.
* **Test Data**:
  ```json
  {
    "registration_number": "DL-01-EV-2026",
    "vehicle_type": "ELECTRIC_VAN",
    "make_model": "Tata Ace EV",
    "year": 2026,
    "max_weight_kg": 750.0,
    "max_volume_cbm": 4.5,
    "assigned_branch_id": "<branch-uuid>"
  }
  ```
* **Test Steps**: Dispatch `POST /api/v1/tenants/:tenant_id/vehicles`.
* **Expected Result**: HTTP 201 Created; vehicle registered with status `AVAILABLE` and `is_active=true`.
* **Actual Result**: HTTP 201 Created; electric van registered and linked to branch.
* **Verification Test**: `TestVehicle_Create_WithBranch` in `backend/tests/integration/vehicle_integration_test.go`
* **Status**: **PASS**

---

### TC-P2-VEH-002: Reject Duplicate Vehicle Plate Within Same Tenant
* **Phase**: Phase 2 | **Module**: Vehicles | **Feature**: Plate Uniqueness
* **Test Type**: Constraint | **Priority**: P1 | **Severity**: Moderate | **Automated**: Yes
* **Preconditions**: Vehicle `DL-01-EV-2026` exists in current tenant.
* **Test Steps**: Create second vehicle in same tenant with plate `DL-01-EV-2026`.
* **Expected Result**: HTTP 409 Conflict.
* **Actual Result**: HTTP 409 Conflict returned; composite index `(tenant_id, registration_number)` enforces uniqueness.
* **Verification Test**: `TestVehicle_DuplicateRegNum_RejectedWithinTenant` in `backend/tests/integration/vehicle_integration_test.go`
* **Status**: **PASS**

---

### TC-P2-VEH-003: Driver Assignment Lifecycle and Double-Booking Prevention
* **Phase**: Phase 2 | **Module**: Vehicles | **Feature**: Resource Scheduling & Conflict Prevention
* **Test Type**: Business Logic / Concurrency | **Priority**: P0 | **Severity**: Critical | **Automated**: Yes
* **Preconditions**: Vehicle 1, Vehicle 2, Driver 1, and Driver 2 exist in same tenant.
* **Test Steps**:
  1. Assign Driver 1 to Vehicle 1 -> HTTP 201 Created; status becomes `ASSIGNED`.
  2. Attempt to assign Driver 1 to Vehicle 2 -> HTTP 409 Conflict (Driver already assigned).
  3. Attempt to assign Driver 2 to Vehicle 1 -> HTTP 409 Conflict (Vehicle already assigned).
  4. Unassign Driver 1 from Vehicle 1 -> HTTP 200 OK; status becomes `AVAILABLE`.
  5. Assign Driver 1 to Vehicle 2 -> HTTP 201 Created (Succeeds now that Driver 1 is free).
* **Expected Result**: Partial unique indexes in PostgreSQL guarantee race-condition-free double-assignment prevention at the database layer.
* **Actual Result**: HTTP 409 Conflict returned on both conflicting assignment attempts; clean unassignment and reassignment verified.
* **Verification Test**: `TestVehicle_DriverAssignment_Lifecycle_And_ConflictPrevention` in `backend/tests/integration/vehicle_integration_test.go`
* **Status**: **PASS**

---

### TC-P2-ORG-001: Least-Privilege Role Authorization Across Phase 2
* **Phase**: Phase 2 | **Module**: Authorization | **Feature**: Role Permission Boundaries
* **Test Type**: Security / RBAC | **Priority**: P0 | **Severity**: Critical | **Automated**: Yes
* **Preconditions**: User with `VIEWER` role in tenant.
* **Test Steps**:
  1. `VIEWER` attempts `POST /branches` -> HTTP 403 Forbidden.
  2. `VIEWER` attempts `POST /employees` -> HTTP 403 Forbidden.
  3. `VIEWER` attempts `POST /vehicles` -> HTTP 403 Forbidden.
  4. `VIEWER` attempts `POST /vehicles/:id/assign` -> HTTP 403 Forbidden.
* **Expected Result**: `RequireRole(RoleTenantAdmin, RoleTenantOperator)` rejects all mutating actions from viewers.
* **Actual Result**: HTTP 403 Forbidden returned on all mutation attempts.
* **Verification Test**: Verified across `TestBranch_ViewerRole_ForbiddenFromBranchCreation`, `TestEmployee_ViewerRole_ForbiddenFromEmployeeMutation`, and `TestVehicle_ViewerRole_ForbiddenFromMutation`.
* **Status**: **PASS**
