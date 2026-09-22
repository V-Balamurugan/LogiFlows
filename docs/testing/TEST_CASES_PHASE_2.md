# LogiFlows — Phase 2 Test Case Catalog (Organization & Resource Management)

**Document Reference**: `docs/testing/TEST_CASES_PHASE_2.md`  
**Phase**: Phase 2 — Organization and Resource Management  
**Status**: **PLANNED / NOT IMPLEMENTED**  

> [!IMPORTANT]
> **Implementation Scope Notice**:
> Phase 2 application modules (Branches, Employees, Vehicles, and Resource Assignment) have not yet been implemented in the codebase (no database migrations, domain models, services, or HTTP routes exist).
> In adherence to the testing principle *"Never generate fake test results or false completion claims"*, all test cases in this catalog are designed in advance as **`PLANNED / NOT IMPLEMENTED`**. They define the acceptance criteria and regression harness for the future Phase 2 vertical slice.

---

### TC-P2-BRN-001: Create Delivery Branch with PostGIS Coordinates
* **Phase**: Phase 2 | **Module**: Branches | **Feature**: Branch Provisioning
* **Test Type**: API / Spatial | **Priority**: P1 | **Severity**: Major | **Automated**: Planned
* **Preconditions**: Authenticated user with `TENANT_ADMIN` role; valid active tenant.
* **Test Data**:
  ```json
  {
    "branch_code": "BRN-DELHI-NORTH",
    "name": "Delhi North Sorting Hub",
    "address": "Plot 12, Industrial Area, GT Karnal Rd, Delhi",
    "latitude": 28.7041,
    "longitude": 77.1025,
    "coverage_radius_km": 25.0
  }
  ```
* **Test Steps**: Dispatch `POST /api/v1/tenants/:tenant_id/branches`.
* **Expected Result**: HTTP 201 Created; branch stored in database with valid PostGIS point geometry (`ST_SetSRID(ST_MakePoint(lng, lat), 4326)`).
* **Actual Result**: Feature not implemented.
* **Status**: **PLANNED / NOT IMPLEMENTED** | **Regression Required**: Yes
* **Related API**: `POST /api/v1/tenants/:tenant_id/branches`

---

### TC-P2-BRN-002: Reject Duplicate Branch Code Within Same Tenant
* **Phase**: Phase 2 | **Module**: Branches | **Feature**: Branch Code Uniqueness
* **Test Type**: API / Constraint | **Priority**: P1 | **Severity**: Moderate | **Automated**: Planned
* **Preconditions**: Branch with code `BRN-DELHI-NORTH` exists in current tenant.
* **Test Steps**: Attempt to create second branch with same `branch_code` within same tenant.
* **Expected Result**: HTTP 409 Conflict; error code `BRANCH_CODE_EXISTS`.
* **Actual Result**: Feature not implemented.
* **Status**: **PLANNED / NOT IMPLEMENTED** | **Regression Required**: Yes
* **Related API**: `POST /api/v1/tenants/:tenant_id/branches`

---

### TC-P2-BRN-003: Enforce Multi-Tenant Isolation on Branch Retrieval
* **Phase**: Phase 2 | **Module**: Branches | **Feature**: Multi-Tenant Isolation
* **Test Type**: Security | **Priority**: P0 | **Severity**: Critical | **Automated**: Planned
* **Preconditions**: User belongs to Tenant Alpha; Branch belongs to Tenant Beta.
* **Test Steps**: User attempts `GET /api/v1/tenants/<tenant-beta-id>/branches/<branch-id>`.
* **Expected Result**: HTTP 403 Forbidden; error code `CROSS_TENANT_ACCESS_DENIED`.
* **Actual Result**: Feature not implemented.
* **Status**: **PLANNED / NOT IMPLEMENTED** | **Regression Required**: Yes
* **Related API**: `GET /api/v1/tenants/:tenant_id/branches/:branch_id`

---

### TC-P2-BRN-004: Branch Pagination, Geographic Filtering, and Soft-Deletion
* **Phase**: Phase 2 | **Module**: Branches | **Feature**: Branch Management
* **Test Type**: API | **Priority**: P2 | **Severity**: Moderate | **Automated**: Planned
* **Preconditions**: 15 branches exist in tenant.
* **Test Steps**:
  1. `GET /branches?page=1&limit=10` (Verify pagination metadata).
  2. `GET /branches?near_lat=28.70&near_lng=77.10&radius_km=10` (Verify spatial query).
  3. `DELETE /branches/:id` (Verify soft-delete flag `is_active = false`).
* **Expected Result**: Correct page chunks returned; spatial distance filtering accurate; soft-delete preserves historical shipments.
* **Actual Result**: Feature not implemented.
* **Status**: **PLANNED / NOT IMPLEMENTED** | **Regression Required**: Yes

---

### TC-P2-EMP-001: Register Employee Profile Linked to User and Branch
* **Phase**: Phase 2 | **Module**: Employees | **Feature**: Employee Onboarding
* **Test Type**: API | **Priority**: P1 | **Severity**: Major | **Automated**: Planned
* **Preconditions**: User has membership in tenant; branch belongs to same tenant.
* **Test Data**:
  ```json
  {
    "user_id": "<user-uuid>",
    "branch_id": "<branch-uuid>",
    "employee_code": "EMP-DEL-042",
    "designation": "Delivery Associate",
    "employment_type": "FULL_TIME"
  }
  ```
* **Test Steps**: Dispatch `POST /api/v1/tenants/:tenant_id/employees`.
* **Expected Result**: HTTP 201 Created; employee record linked with foreign keys to user, tenant, and branch.
* **Actual Result**: Feature not implemented.
* **Status**: **PLANNED / NOT IMPLEMENTED** | **Regression Required**: Yes

---

### TC-P2-EMP-002: Reject Employee Assignment to Foreign Tenant's Branch
* **Phase**: Phase 2 | **Module**: Employees | **Feature**: Multi-Tenant Validation
* **Test Type**: Security | **Priority**: P0 | **Severity**: Critical | **Automated**: Planned
* **Preconditions**: Branch belongs to Tenant Beta; Employee being created in Tenant Alpha.
* **Test Steps**: Attempt to link Tenant Alpha employee to Tenant Beta branch ID.
* **Expected Result**: HTTP 400 Bad Request or 403 Forbidden (`INVALID_BRANCH_ASSOCIATION`).
* **Actual Result**: Feature not implemented.
* **Status**: **PLANNED / NOT IMPLEMENTED** | **Regression Required**: Yes

---

### TC-P2-EMP-003: Update Employee Operational Status and Deactivation
* **Phase**: Phase 2 | **Module**: Employees | **Feature**: Employee Lifecycle
* **Test Type**: API | **Priority**: P2 | **Severity**: Moderate | **Automated**: Planned
* **Preconditions**: Active employee exists.
* **Test Steps**: Update status to `ON_LEAVE`, then `TERMINATED`.
* **Expected Result**: Status changes recorded in audit log; terminated employee cannot receive new parcel custody assignments.
* **Actual Result**: Feature not implemented.
* **Status**: **PLANNED / NOT IMPLEMENTED** | **Regression Required**: Yes

---

### TC-P2-VEH-001: Register Fleet Vehicle with Capacity Specification
* **Phase**: Phase 2 | **Module**: Vehicles | **Feature**: Fleet Management
* **Test Type**: API | **Priority**: P1 | **Severity**: Major | **Automated**: Planned
* **Preconditions**: Authenticated `TENANT_ADMIN` or `TENANT_OPERATOR`.
* **Test Data**:
  ```json
  {
    "registration_number": "DL-01-AB-1234",
    "vehicle_type": "ELECTRIC_VAN",
    "max_weight_kg": 750.0,
    "max_volume_cbm": 4.5,
    "assigned_branch_id": "<branch-uuid>"
  }
  ```
* **Test Steps**: Dispatch `POST /api/v1/tenants/:tenant_id/vehicles`.
* **Expected Result**: HTTP 201 Created; vehicle registered with status `AVAILABLE`.
* **Actual Result**: Feature not implemented.
* **Status**: **PLANNED / NOT IMPLEMENTED** | **Regression Required**: Yes

---

### TC-P2-VEH-002: Reject Duplicate Vehicle Registration Number Within Tenant
* **Phase**: Phase 2 | **Module**: Vehicles | **Feature**: Unique Constraint
* **Test Type**: API | **Priority**: P1 | **Severity**: Moderate | **Automated**: Planned
* **Preconditions**: Vehicle `DL-01-AB-1234` already exists in tenant fleet.
* **Test Steps**: Dispatch `POST /api/v1/tenants/:tenant_id/vehicles` with duplicate plate.
* **Expected Result**: HTTP 409 Conflict (`VEHICLE_ALREADY_EXISTS`).
* **Actual Result**: Feature not implemented.
* **Status**: **PLANNED / NOT IMPLEMENTED** | **Regression Required**: Yes

---

### TC-P2-VEH-003: Assign Vehicle to Driver with Conflict Detection
* **Phase**: Phase 2 | **Module**: Vehicles | **Feature**: Resource Scheduling
* **Test Type**: API / Business Rules | **Priority**: P2 | **Severity**: Major | **Automated**: Planned
* **Preconditions**: Vehicle is currently in state `ASSIGNED` to Driver 1.
* **Test Steps**: Attempt to simultaneously assign vehicle to Driver 2 without unassigning Driver 1.
* **Expected Result**: HTTP 409 Conflict (`VEHICLE_CURRENTLY_ASSIGNED`).
* **Actual Result**: Feature not implemented.
* **Status**: **PLANNED / NOT IMPLEMENTED** | **Regression Required**: Yes

---

### TC-P2-ORG-001: Organization Role Permission Boundary Enforcement
* **Phase**: Phase 2 | **Module**: Organization | **Feature**: Organization-Level RBAC
* **Test Type**: Security | **Priority**: P0 | **Severity**: Critical | **Automated**: Planned
* **Preconditions**: Users with roles `TENANT_ADMIN`, `TENANT_OPERATOR`, and `VIEWER`.
* **Test Steps**:
  1. `VIEWER` attempts to create vehicle -> Rejected (403).
  2. `TENANT_OPERATOR` creates vehicle -> Allowed (201).
  3. `TENANT_OPERATOR` attempts to delete branch -> Rejected (403).
  4. `TENANT_ADMIN` deletes branch -> Allowed (200).
* **Expected Result**: Least privilege principle strictly maintained across resource management endpoints.
* **Actual Result**: Feature not implemented.
* **Status**: **PLANNED / NOT IMPLEMENTED** | **Regression Required**: Yes
