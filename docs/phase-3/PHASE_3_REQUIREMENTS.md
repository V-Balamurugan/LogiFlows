# LogiFlows — Phase 3 Requirements Specification

**Document Reference**: `docs/phase-3/PHASE_3_REQUIREMENTS.md`  
**Phase**: Phase 3 — Employees, Roles, and Vehicles  
**Version**: 1.0  
**Status**: APPROVED  

---

## 1. Scope & Objective

Phase 3 implements the end-to-end operational resource management module for LogiFlows, including:
1. Complete employee profile lifecycle, operational roles, automated code generation, and verification.
2. Complete vehicle fleet registration, types, capacities, operating status, and availability state tracking.
3. Driver eligibility validation, driver-vehicle assignment lifecycle, concurrency-safe double-booking prevention, and audit tracking.
4. Strict multi-tenant isolation and branch-level scoping.
5. React web console and Flutter mobile interfaces.

---

## 2. Employee User Stories

### US-E01: Create Employee Profile
* **As a**: Tenant Administrator (`TENANT_ADMIN`)
* **I want to**: Register an employee profile with operational role and branch association.
* **Preconditions**: Authenticated user with `TENANT_ADMIN` role in target tenant; branch exists and is active.
* **Request**: `POST /api/v1/tenants/{tenant_id}/employees`
* **Response**: `HTTP 201 Created` with created employee profile.
* **Validation**: First name, last name, designation required; valid employment type and operational role; branch belongs to tenant.
* **Authorization**: `RequireRole(TENANT_ADMIN)`.
* **Positive Tests**: Create employee with complete details; create driver linked to branch.
* **Negative Tests**: Missing first name (`400 Bad Request`); invalid operational role (`400 Bad Request`); branch from other tenant (`400 Bad Request`).

### US-E02: Unique Employee Code Generation
* **As a**: Tenant Administrator
* **I want to**: Have the system automatically generate a unique, sequential employee code (e.g. `EMP-0001`) if not provided.
* **Preconditions**: Tenant exists.
* **Request**: `POST /api/v1/tenants/{tenant_id}/employees` with empty or omitted `employee_code`.
* **Response**: Employee created with generated sequential `employee_code`.
* **Validation**: Code must be unique within tenant; concurrent creations must not produce duplicate key errors.
* **Positive Tests**: Create consecutive employees without code -> `EMP-0001`, `EMP-0002`.
* **Negative Tests**: Explicitly providing an existing employee code returns `409 Conflict`.

### US-E03: Associate Employee with Branch
* **As a**: Tenant Administrator
* **I want to**: Associate an employee with a distribution branch.
* **Preconditions**: Branch exists, belongs to tenant, and is active.
* **Request**: `POST /api/v1/tenants/{tenant_id}/employees` with `branch_id`.
* **Response**: Employee profile with joined branch code and name.
* **Validation**: Cross-tenant branch assignment rejected (`400 Bad Request`).
* **Positive Tests**: Associate employee with active branch.
* **Negative Tests**: Associate with inactive branch or branch belonging to Tenant B (`400 Bad Request`).

### US-E04: View Authorized Employees
* **As an**: Operations User (`TENANT_ADMIN`, `TENANT_OPERATOR`, `VIEWER`)
* **I want to**: List and view details of employees within my tenant organization.
* **Preconditions**: User belongs to tenant.
* **Request**: `GET /api/v1/tenants/{tenant_id}/employees` & `GET /api/v1/tenants/{tenant_id}/employees/{employee_id}`
* **Response**: `HTTP 200 OK` with paginated list or detailed employee profile.
* **Positive Tests**: Retrieve list with pagination; retrieve employee by ID.
* **Negative Tests**: Query non-existent ID (`404 Not Found`); query ID from another tenant (`403 Forbidden` / `404 Not Found`).

### US-E05: Update Employee Information
* **As an**: Tenant Administrator
* **I want to**: Update employee contact details, designation, role, and branch assignment.
* **Preconditions**: Employee exists in tenant.
* **Request**: `PUT /api/v1/tenants/{tenant_id}/employees/{employee_id}` or `PATCH`
* **Response**: `HTTP 200 OK` with updated employee entity.
* **Positive Tests**: Update phone number and designation; reassign to different valid branch.
* **Negative Tests**: Reassign to invalid or foreign branch (`400 Bad Request`).

### US-E06: Activate, Deactivate, or Update Employee Status
* **As an**: Tenant Administrator
* **I want to**: Transition employee status (`ACTIVE`, `ON_LEAVE`, `SUSPENDED`, `TERMINATED`) and availability status (`AVAILABLE`, `BUSY`, `OFF_DUTY`, `UNAVAILABLE`).
* **Preconditions**: Employee exists in tenant.
* **Request**: `PATCH /api/v1/tenants/{tenant_id}/employees/{employee_id}/status`
* **Response**: `HTTP 200 OK` with updated status.
* **Positive Tests**: Set status to `ON_LEAVE`; set availability to `OFF_DUTY`.
* **Negative Tests**: Invalid status value returns `400 Bad Request`.

### US-E07: Filter Employees by Branch, Role, and Status
* **As an**: Operations User
* **I want to**: Filter the employee directory by operational role (`DRIVER`, `OPERATOR`, `DISPATCHER`), branch, status, or search term.
* **Preconditions**: User belongs to tenant.
* **Request**: `GET /api/v1/tenants/{tenant_id}/employees?operational_role=DRIVER&status=ACTIVE&branch_id=...`
* **Response**: Filtered list matching criteria.
* **Positive Tests**: Filter by `operational_role=DRIVER` returns only drivers.

### US-E08: Identify Available Drivers
* **As a**: Dispatcher or Operations Manager
* **I want to**: Quickly view drivers who are currently `ACTIVE` and `AVAILABLE` for dispatch assignments.
* **Preconditions**: User has operator or admin role.
* **Request**: `GET /api/v1/tenants/{tenant_id}/employees?operational_role=DRIVER&status=ACTIVE&availability_status=AVAILABLE`
* **Response**: List of available drivers.
* **Positive Tests**: Only drivers not currently assigned to active vehicles are returned.

### US-E09: Cross-Tenant Isolation for Employees
* **As an**: Authenticated User
* **I must not**: Be able to read, create, update, or deactivate employees belonging to another tenant.
* **Request**: Any employee endpoint with foreign tenant ID.
* **Response**: `HTTP 403 Forbidden`.
* **Positive Tests**: Tenant A user accessing Tenant A data -> `200 OK`.
* **Negative Tests**: Tenant A user accessing Tenant B data -> `403 Forbidden`.

---

## 3. Vehicle User Stories

### US-V01: Register Vehicle
* **As a**: Tenant Administrator or Operator
* **I want to**: Register a delivery vehicle with registration number, type, and capacity.
* **Preconditions**: User has required role; registration number unique within tenant.
* **Request**: `POST /api/v1/tenants/{tenant_id}/vehicles`
* **Response**: `HTTP 201 Created` with vehicle record.
* **Validation**: Registration number, vehicle type, max weight (>0), max volume (>0).
* **Positive Tests**: Register `ELECTRIC_VAN` with 1000kg / 8.5cbm.
* **Negative Tests**: Negative weight or volume (`400 Bad Request`); duplicate registration number within tenant (`409 Conflict`).

### US-V02: Record Vehicle Type and Capacity
* **As an**: Authorized Operator
* **I want to**: Specify standard vehicle categories (`ELECTRIC_VAN`, `VAN`, `MOTORCYCLE`, `TRUCK`, `THREE_WHEELER`) and physical payload limits.
* **Validation**: `vehicle_type` in allowed enum; `max_weight_kg > 0`; `max_volume_cbm > 0`.
* **Positive Tests**: Valid payload limits stored accurately.
* **Negative Tests**: Invalid vehicle type returns `400 Bad Request`.

### US-V03: Associate Vehicle with Branch
* **As an**: Authorized Operator
* **I want to**: Assign a vehicle to a home branch / distribution hub.
* **Preconditions**: Branch exists in same tenant.
* **Validation**: Cross-tenant branch rejected.
* **Positive Tests**: Vehicle successfully linked to branch.
* **Negative Tests**: Foreign branch returns `400 Bad Request`.

### US-V04: View Available Vehicles
* **As an**: Operations User
* **I want to**: List vehicles filtered by status (`AVAILABLE`), type, and branch.
* **Request**: `GET /api/v1/tenants/{tenant_id}/vehicles?status=AVAILABLE`
* **Response**: Paginated list of unassigned, operational vehicles.
* **Positive Tests**: Filter returns only available fleet units.

### US-V05: Update Vehicle Information
* **As an**: Tenant Administrator
* **I want to**: Update vehicle make, model, year, payload capacity, and branch.
* **Request**: `PUT /api/v1/tenants/{tenant_id}/vehicles/{vehicle_id}` or `PATCH`
* **Response**: `HTTP 200 OK` with updated vehicle.

### US-V06: Change Vehicle Status
* **As an**: Authorized Operator
* **I want to**: Update vehicle status (`AVAILABLE`, `ASSIGNED`, `IN_TRANSIT`, `MAINTENANCE`, `DECOMMISSIONED`).
* **Request**: `PATCH /api/v1/tenants/{tenant_id}/vehicles/{vehicle_id}/status`
* **Response**: `HTTP 200 OK` with updated status.

### US-V07: Assign Eligible Driver to Vehicle
* **As an**: Operations User (`TENANT_ADMIN`, `TENANT_OPERATOR`)
* **I want to**: Assign an eligible driver to a vehicle for delivery duty.
* **Validation Rules**:
  1. Employee exists in same tenant.
  2. Employee has `operational_role == 'DRIVER'`.
  3. Employee is `status == 'ACTIVE'`.
  4. Employee is `availability_status == 'AVAILABLE'`.
  5. Vehicle is `is_active == true` and not decommissioned or under maintenance.
* **Side Effects**: Vehicle status becomes `ASSIGNED`; employee availability becomes `BUSY`.
* **Positive Tests**: Assign active, available driver to available van -> `201 Created`.
* **Negative Tests**: Assign non-driver employee returns `400 Bad Request` or `422 Unprocessable Entity`. Assign driver already on active assignment returns `409 Conflict`.

### US-V08: Double-Booking Prevention & Concurrency Protection
* **As a**: System Security & Reliability Architect
* **The system must**: Guarantee that neither a vehicle nor a driver can have more than one simultaneous `ACTIVE` assignment.
* **Enforcement**: Partial unique indexes in PostgreSQL (`uq_active_vehicle_assignment`, `uq_active_driver_assignment`) combined with serializable application-level checking.
* **Response on Conflict**: `HTTP 409 Conflict`.
* **Positive Tests**: Unassign driver restores vehicle to `AVAILABLE` and driver to `AVAILABLE`.
* **Negative Tests**: Concurrent assignment requests result in exactly one winner and one `409 Conflict`.

### US-V09: Cross-Tenant Isolation for Vehicles and Assignments
* **As an**: Authenticated User
* **I must not**: Access, update, decommission, or assign vehicles belonging to another tenant.
* **Response**: `HTTP 403 Forbidden`.

---

## 4. Employee Account & Identity User Stories

### US-E10: Transactional Employee Account Creation
* **As a**: Tenant Administrator (`TENANT_ADMIN`)
* **I want to**: Provision a full system login account atomically when onboarding a new employee.
* **Preconditions**: Authenticated user with `TENANT_ADMIN` role; valid email not already in use; strong password provided (min 8 chars).
* **Request**: `POST /api/v1/tenants/{tenant_id}/employees/with-account`
* **Workflow**:
  1. Validate email and password strength.
  2. Verify tenant and branch ownership.
  3. Hash password using bcrypt (cost 12).
  4. In a single atomic database transaction:
     a. Insert user into `users` table.
     b. Insert membership into `tenant_memberships` table with system role `EMPLOYEE`.
     c. Insert profile into `employees` table with `user_id` linked and operational role assigned.
  5. Roll back entire transaction if any step or constraint fails.
* **Response**: `HTTP 201 Created` with employee details and `account_provisioned: true` (no passwords or hashes returned).
* **Positive Tests**: Successful creation creates user, membership, and employee profile.
* **Negative Tests**: Duplicate email rolls back employee creation (`409 Conflict`); missing branch rolls back user creation (`400 Bad Request`).

### US-E11: Inspect Employee Account Status
* **As a**: Tenant Administrator or Operator
* **I want to**: View whether an employee profile has an active linked login account, their system role, and membership status.
* **Request**: `GET /api/v1/tenants/{tenant_id}/employees/{employee_id}/account-status`
* **Response**: `HTTP 200 OK` with `has_account`, `user_id`, `system_role`, `membership_status`, and `is_verified`.

### US-E12: Employee Self-Profile Retrieval
* **As an**: Authenticated Employee (`EMPLOYEE` system role)
* **I want to**: View my own personal employee profile, assigned branch details, and currently assigned vehicle.
* **Request**: `GET /api/v1/tenants/{tenant_id}/employees/me`
* **Response**: `HTTP 200 OK` with employee record, branch record, and active assigned vehicle specs.
* **Authorization**: Matches `user_id` from JWT context. Non-employee users receive appropriate message or profile view.

---

## 5. Branch Operational Assets User Stories

### US-B01: View Branch Employees
* **As an**: Operations User or Branch Manager
* **I want to**: View all employees stationed at a specific branch.
* **Request**: `GET /api/v1/tenants/{tenant_id}/branches/{branch_id}/employees`
* **Response**: `HTTP 200 OK` with list of employees stationed at the branch.

### US-B02: View Branch Vehicles
* **As an**: Operations User or Branch Manager
* **I want to**: View all commercial fleet vehicles stationed at a specific branch.
* **Request**: `GET /api/v1/tenants/{tenant_id}/branches/{branch_id}/vehicles`
* **Response**: `HTTP 200 OK` with list of vehicles allocated to the branch.

---

## 6. Role-Based Dashboards & Routing User Stories

### US-D01: System Authorization vs. Operational Role Separation
* **System Roles**: `PLATFORM_ADMIN`, `TENANT_ADMIN`, `TENANT_OPERATOR`, `VIEWER`, `EMPLOYEE`.
* **Operational Roles**: `DRIVER`, `OPERATOR`, `DISPATCHER`, `SUPERVISOR`, `MANAGER`, `BRANCH_MANAGER`, `WAREHOUSE_OPERATOR`, `DELIVERY_EXECUTIVE`.
* **Routing**:
  - `PLATFORM_ADMIN` / `TENANT_ADMIN`: Full access to tenant settings, branch management, employee onboarding with account provisioning, fleet management, and assignments.
  - `TENANT_OPERATOR`: Fleet operations, branch assets inspection, driver assignment, vehicle status management.
  - `VIEWER`: Read-only views across authorized tenant resources.
  - `EMPLOYEE`: Dedicated Self-Profile dashboard (`/employees/me`), displaying assigned branch, operational badge, availability toggle, and active vehicle assignment.

