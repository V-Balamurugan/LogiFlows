# LogiFlows — Phase 2 Scope & Requirements Specification

**Document Reference**: `docs/phase-2/PHASE_2_SCOPE.md`  
**Phase**: Phase 2 — Organization, Branch, Employee, and Fleet Management  
**Status**: APPROVED & IMPLEMENTED  

---

## 1. Scope Overview

Phase 2 establishes the multi-tenant organizational structure and physical resource hierarchy of LogiFlows:
1. **Organization/Tenant Management Integration**: Tenant switching, metadata updating, member invitation, and RBAC matrix.
2. **Branch Management**: Physical logistics hubs, PostGIS geospatial coordinates, geographic service radius, and spatial querying.
3. **Employee Management**: Staff, dispatchers, and delivery driver profiles, operational roles, driver license tracking, and branch assignment.
4. **Organization Roles & Branch-Level Access Control**: Server-side RBAC enforcement, least-privilege principle, and cross-tenant/cross-branch boundary validation.
5. **Vehicle & Fleet Management**: Zero-emission electric vans, cargo vans, trucks, motorcycles, payload/volume capacity validation, and branch affiliation.
6. **Resource Assignment & Conflict Prevention**: Concurrency-safe driver-vehicle scheduling backed by database partial unique indexes preventing double-booking.
7. **Web & Mobile Client Integration**: React web dashboard components and Flutter mobile client driver tools.

---

## 2. Feature Specifications

### Feature A: Organization & Tenant Integration
* **Business Purpose**: Enable multi-company logistics tenancy where multiple carrier or client organizations operate isolated environments within a shared infrastructure.
* **Allowed Roles**: `PLATFORM_ADMIN`, `TENANT_ADMIN` (Full Management); `TENANT_OPERATOR`, `VIEWER` (Read Only).
* **Database Entities**: `tenants`, `tenant_memberships`, `audit_logs`.
* **Backend Endpoints**:
  * `GET /api/v1/tenants`: List authorized tenants for current user.
  * `GET /api/v1/tenants/:tenant_id`: Get specific tenant details.
  * `PATCH /api/v1/tenants/:tenant_id`: Update tenant metadata (name, slug).
  * `GET /api/v1/tenants/:tenant_id/members`: List team members and assigned roles.
  * `POST /api/v1/tenants/:tenant_id/members`: Invite or add user to tenant with assigned role.
* **Web Screens**: Top navigation organization dropdown, Team Member Modal (`TeamModal.tsx`).
* **Mobile Screens**: Tenant selection on login (`login_screen.dart`).
* **Validation Rules**:
  * Tenant name: 2-100 characters.
  * Tenant slug: Lowercase alphanumeric with hyphens, 2-50 characters.
  * Role: Must be one of `PLATFORM_ADMIN`, `TENANT_ADMIN`, `TENANT_OPERATOR`, `VIEWER`.
* **Security Requirements**: Multi-tenant isolation verified server-side; client cannot spoof `tenant_id`.
* **Completion Criteria**: Tenant isolation tests pass; audit logs recorded for membership changes.

---

### Feature B: Branch Management
* **Business Purpose**: Model physical logistics hubs, distribution centers, and pickup points with geospatial coordinates and coverage boundaries.
* **Allowed Roles**: `PLATFORM_ADMIN`, `TENANT_ADMIN` (CRUD); `TENANT_OPERATOR` (Read); `VIEWER` (Read).
* **Database Entities**: `branches` (with PostGIS `GEOMETRY(Point, 4326)`).
* **Backend Endpoints**:
  * `POST /api/v1/tenants/:tenant_id/branches`: Create new branch.
  * `GET /api/v1/tenants/:tenant_id/branches`: List branches with spatial filtering and pagination.
  * `GET /api/v1/tenants/:tenant_id/branches/:branch_id`: Get branch by ID.
  * `PATCH /api/v1/tenants/:tenant_id/branches/:branch_id`: Update branch details.
  * `DELETE /api/v1/tenants/:tenant_id/branches/:branch_id`: Soft delete branch (`is_active = false`).
* **Web Screens**: Branches & Hubs tab (`BranchList.tsx`) with create modal, search, and spatial filtering.
* **Mobile Screens**: Branch directory screen (`branch_screen.dart`).
* **Validation Rules**:
  * `branch_code`: Required, alphanumeric with hyphens/underscores, unique within tenant.
  * `name`: Required, 2-100 characters.
  * `latitude`: -90.0 to +90.0.
  * `longitude`: -180.0 to +180.0.
  * `coverage_radius_km`: Strictly positive number (defaults to 15.0 km).
* **Security Requirements**: Branch operations strictly scoped to `tenant_id`; cross-tenant access returns 403.
* **Completion Criteria**: PostGIS spatial queries verified with `ST_DistanceSphere`; duplicate code rejected within tenant.

---

### Feature C: Employee Management
* **Business Purpose**: Manage internal operational personnel (drivers, warehouse operators, dispatchers, managers) and associate them with branches.
* **Allowed Roles**: `PLATFORM_ADMIN`, `TENANT_ADMIN` (CRUD); `TENANT_OPERATOR` (Read); `VIEWER` (Read).
* **Database Entities**: `employees`.
* **Backend Endpoints**:
  * `POST /api/v1/tenants/:tenant_id/employees`: Onboard new employee.
  * `GET /api/v1/tenants/:tenant_id/employees`: List employees with filters (role, status, branch).
  * `GET /api/v1/tenants/:tenant_id/employees/:employee_id`: Get employee details.
  * `PUT /api/v1/tenants/:tenant_id/employees/:employee_id`: Update employee profile.
  * `DELETE /api/v1/tenants/:tenant_id/employees/:employee_id`: Soft deactivate employee.
* **Web Screens**: Employees & Staff tab (`EmployeeList.tsx`) with onboarding modal, role filtering.
* **Mobile Screens**: Driver profile view.
* **Validation Rules**:
  * `employee_code`: Unique within tenant.
  * `first_name`, `last_name`: Required, 1-50 characters.
  * `operational_role`: Must be one of `DRIVER`, `OPERATOR`, `DISPATCHER`, `SUPERVISOR`, `MANAGER`.
  * `employment_type`: Must be one of `FULL_TIME`, `PART_TIME`, `CONTRACTOR`, `INTERN`.
  * `branch_id`: If supplied, must exist and belong to the same tenant.
* **Security Requirements**: Cross-tenant branch association rejected with 400 Bad Request; viewer role rejected with 403 Forbidden.
* **Completion Criteria**: Safe user and branch foreign key linking verified; duplicate employee codes rejected.

---

### Feature D: Vehicle Management & Fleet Resource Assignment
* **Business Purpose**: Track delivery fleet assets (electric vans, trucks, bikes) and manage conflict-free assignment of drivers to vehicles.
* **Allowed Roles**: `PLATFORM_ADMIN`, `TENANT_ADMIN`, `TENANT_OPERATOR` (Manage/Assign); `VIEWER` (Read).
* **Database Entities**: `vehicles`, `vehicle_assignments`.
* **Backend Endpoints**:
  * `POST /api/v1/tenants/:tenant_id/vehicles`: Register vehicle.
  * `GET /api/v1/tenants/:tenant_id/vehicles`: List fleet vehicles with filters.
  * `GET /api/v1/tenants/:tenant_id/vehicles/:vehicle_id`: Get vehicle details.
  * `PATCH /api/v1/tenants/:tenant_id/vehicles/:vehicle_id`: Update vehicle specs/status.
  * `DELETE /api/v1/tenants/:tenant_id/vehicles/:vehicle_id`: Decommission vehicle.
  * `POST /api/v1/tenants/:tenant_id/vehicles/:vehicle_id/assign`: Assign vehicle to driver.
  * `POST /api/v1/tenants/:tenant_id/vehicles/:vehicle_id/unassign`: Unassign vehicle from driver.
* **Web Screens**: Fleet Vehicles tab (`VehicleList.tsx`) with vehicle registration modal and driver assignment modal.
* **Mobile Screens**: Driver vehicle custody screen (`vehicle_screen.dart`).
* **Validation Rules**:
  * `registration_number`: Unique within tenant, uppercase, 2-50 chars.
  * `vehicle_type`: `ELECTRIC_VAN`, `VAN`, `MOTORCYCLE`, `TRUCK`, `THREE_WHEELER`.
  * `max_weight_kg`, `max_volume_cbm`: Strictly positive numbers.
  * Driver assignment: Driver must hold `operational_role = 'DRIVER'` and be in `ACTIVE` status.
* **Security Requirements**: Concurrency-safe double-booking prevention backed by partial unique indexes in PostgreSQL.
* **Completion Criteria**: Assigning already assigned vehicle or driver returns 409 Conflict; unassignment restores availability.
