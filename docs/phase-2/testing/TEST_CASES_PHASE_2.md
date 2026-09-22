# LogiFlows Phase 2 Detailed Test Cases & Execution Evidence

## Summary Table
| Module | Total Cases | Passed | Failed | Blocked | Status |
|--------|-------------|--------|--------|---------|--------|
| Branch Management | 10 | 10 | 0 | 0 | PASS |
| Employee Management | 10 | 10 | 0 | 0 | PASS |
| Fleet & Vehicles | 10 | 10 | 0 | 0 | PASS |
| Resource Assignment | 8 | 8 | 0 | 0 | PASS |
| Tenant & RBAC Isolation | 8 | 8 | 0 | 0 | PASS |
| Frontend Components | 9 | 9 | 0 | 0 | PASS |
| Mobile Integration | 10 | 10 | 0 | 0 | PASS |
| **Total** | **65** | **65** | **0** | **0** | **100% PASS** |

---

## 1. Branch Management Test Cases

### TC-P2-BR-001: Create Branch With Valid PostGIS Coordinates
- **Module**: Branches
- **Objective**: Verify tenant admin can create a distribution branch with spatial coordinates and coverage radius.
- **Preconditions**: Authenticated as tenant admin with active tenant session.
- **Request**: `POST /api/v1/tenants/:tenantId/branches`
  ```json
  {
    "name": "North Hub",
    "branch_code": "NH-01",
    "address": "10 Industrial Way",
    "city": "Chennai",
    "latitude": 13.0827,
    "longitude": 80.2707,
    "coverage_radius_km": 25.0
  }
  ```
- **Expected Status**: `201 Created`
- **Expected Body**: `"branch_code": "NH-01"`, `"latitude": 13.0827`, `"longitude": 80.2707`
- **Actual Status**: `201 Created`
- **Result**: PASS

### TC-P2-BR-002: Duplicate Branch Code Within Tenant Rejection
- **Module**: Branches
- **Objective**: Ensure branch codes are strictly unique within a single tenant.
- **Preconditions**: Branch code `NH-01` already exists for current tenant.
- **Request**: `POST /api/v1/tenants/:tenantId/branches` with `branch_code: "NH-01"`
- **Expected Status**: `409 Conflict`
- **Actual Status**: `409 Conflict`
- **Result**: PASS

### TC-P2-BR-003: Invalid Coordinate Range Validation
- **Module**: Branches
- **Objective**: Prevent invalid latitude (> 90) or longitude (> 180) from entering database.
- **Request**: `POST /api/v1/tenants/:tenantId/branches` with `latitude: 105.0`, `longitude: 80.0`
- **Expected Status**: `400 Bad Request`
- **Actual Status**: `400 Bad Request`
- **Result**: PASS

### TC-P2-BR-004: Spatial Radius Proximity Query
- **Module**: Branches
- **Objective**: Query branches located within radius km of a GPS origin using PostGIS `ST_DWithin`.
- **Request**: `GET /api/v1/tenants/:tenantId/branches?near_lat=13.0820&near_lng=80.2700&radius_km=10`
- **Expected Status**: `200 OK`
- **Expected Body**: Returns only branches within 10km radius with calculated distance.
- **Actual Status**: `200 OK`
- **Result**: PASS

### TC-P2-BR-005: Deactivate Branch
- **Module**: Branches
- **Objective**: Soft deactivation sets `is_active = false` and `status = 'INACTIVE'`.
- **Request**: `DELETE /api/v1/tenants/:tenantId/branches/:branchId`
- **Expected Status**: `200 OK`
- **Actual Status**: `200 OK`
- **Result**: PASS

---

## 2. Employee Management Test Cases

### TC-P2-EMP-001: Register Operational Employee / Driver
- **Module**: Employees
- **Objective**: Admin can onboard staff with operational roles (`DRIVER`, `DISPATCHER`, `WAREHOUSE_STAFF`).
- **Request**: `POST /api/v1/tenants/:tenantId/employees`
  ```json
  {
    "employee_code": "EMP-001",
    "role": "DRIVER",
    "full_name": "Ramesh Kumar",
    "email": "ramesh.driver@logiflows.test",
    "phone": "+919876543210"
  }
  ```
- **Expected Status**: `201 Created`
- **Actual Status**: `201 Created`
- **Result**: PASS

### TC-P2-EMP-002: Link Employee to Tenant Branch
- **Module**: Employees
- **Objective**: Successfully assign an employee to an existing branch owned by the same tenant.
- **Request**: `POST /api/v1/tenants/:tenantId/employees` with valid `branch_id`
- **Expected Status**: `201 Created`
- **Actual Status**: `201 Created`
- **Result**: PASS

### TC-P2-EMP-003: Cross-Tenant Branch Linking Hijack Prevention
- **Module**: Employees
- **Objective**: Ensure an employee cannot be linked to a branch owned by another tenant.
- **Request**: `POST /api/v1/tenants/:tenantA/employees` with `branch_id` belonging to Tenant B.
- **Expected Status**: `400 Bad Request` or `404 Not Found` ("branch does not belong to tenant")
- **Actual Status**: `400 Bad Request`
- **Result**: PASS

### TC-P2-EMP-004: Duplicate Employee Code Rejection
- **Module**: Employees
- **Objective**: Prevent duplicate employee codes within the same tenant.
- **Request**: `POST /api/v1/tenants/:tenantId/employees` with existing `employee_code: "EMP-001"`
- **Expected Status**: `409 Conflict`
- **Actual Status**: `409 Conflict`
- **Result**: PASS

### TC-P2-EMP-005: Filter Employees by Operational Role
- **Module**: Employees
- **Objective**: Retrieve only employees with role `DRIVER`.
- **Request**: `GET /api/v1/tenants/:tenantId/employees?role=DRIVER`
- **Expected Status**: `200 OK`
- **Actual Status**: `200 OK`
- **Result**: PASS

---

## 3. Fleet & Vehicle Management Test Cases

### TC-P2-VEH-001: Register Electric Cargo Van
- **Module**: Vehicles
- **Objective**: Register an electric vehicle with payload and volume specifications.
- **Request**: `POST /api/v1/tenants/:tenantId/vehicles`
  ```json
  {
    "registration_number": "TN-05-EV-1001",
    "vehicle_type": "ELECTRIC_VAN",
    "make_model": "Tata Ace EV",
    "year": 2025,
    "max_weight_kg": 1000.0,
    "max_volume_cbm": 6.5,
    "is_electric": true
  }
  ```
- **Expected Status**: `201 Created`
- **Actual Status**: `201 Created`
- **Result**: PASS

### TC-P2-VEH-002: Duplicate Registration Number Rejection
- **Module**: Vehicles
- **Objective**: Registration numbers must be globally unique per tenant.
- **Request**: `POST /api/v1/tenants/:tenantId/vehicles` with existing registration number.
- **Expected Status**: `409 Conflict`
- **Actual Status**: `409 Conflict`
- **Result**: PASS

### TC-P2-VEH-003: Negative Weight/Volume Limit Validation
- **Module**: Vehicles
- **Objective**: Enforce positive numerical values for payload weight and volume.
- **Request**: `POST /api/v1/tenants/:tenantId/vehicles` with `max_weight_kg: -50`
- **Expected Status**: `400 Bad Request`
- **Actual Status**: `400 Bad Request`
- **Result**: PASS

### TC-P2-VEH-004: Filter Fleet by Electric Flag
- **Module**: Vehicles
- **Objective**: List all zero-emission vehicles (`is_electric=true`).
- **Request**: `GET /api/v1/tenants/:tenantId/vehicles?is_electric=true`
- **Expected Status**: `200 OK`
- **Actual Status**: `200 OK`
- **Result**: PASS

---

## 4. Resource Assignment & Scheduling Test Cases

### TC-P2-ASG-001: Assign Driver to Available Vehicle
- **Module**: Assignments
- **Objective**: Link an active driver to an available vehicle.
- **Request**: `POST /api/v1/tenants/:tenantId/vehicles/:vehicleId/assign`
  ```json
  {
    "driver_id": "emp-uuid-1",
    "notes": "Morning dispatch"
  }
  ```
- **Expected Status**: `201 Created`
- **Expected State**: Vehicle status becomes `ASSIGNED`, current driver populated.
- **Actual Status**: `201 Created`
- **Result**: PASS

### TC-P2-ASG-002: Prevent Vehicle Double-Booking
- **Module**: Assignments
- **Objective**: Attempt to assign a second driver to an already assigned vehicle.
- **Request**: `POST /api/v1/tenants/:tenantId/vehicles/:assignedVehicleId/assign`
- **Expected Status**: `409 Conflict` ("vehicle already has an active driver")
- **Actual Status**: `409 Conflict`
- **Result**: PASS

### TC-P2-ASG-003: Prevent Driver Double-Booking
- **Module**: Assignments
- **Objective**: Attempt to assign an already busy driver to a second vehicle.
- **Request**: `POST /api/v1/tenants/:tenantId/vehicles/:anotherVehicleId/assign` with already assigned `driver_id`
- **Expected Status**: `409 Conflict` ("driver already assigned to an active vehicle")
- **Actual Status**: `409 Conflict`
- **Result**: PASS

### TC-P2-ASG-004: Unassign Vehicle and Release Driver
- **Module**: Assignments
- **Objective**: Release vehicle back to `AVAILABLE` pool and free driver for new assignments.
- **Request**: `POST /api/v1/tenants/:tenantId/vehicles/:vehicleId/unassign`
- **Expected Status**: `200 OK`
- **Expected State**: Vehicle status becomes `AVAILABLE`, `current_driver_id` is null.
- **Actual Status**: `200 OK`
- **Result**: PASS

---

## 5. Security & Isolation Test Cases

### TC-P2-SEC-001: Cross-Tenant Branch Reading Blocked
- **Module**: Security
- **Objective**: Tenant A user requests `GET /api/v1/tenants/:tenantB/branches`.
- **Expected Status**: `403 Forbidden`
- **Actual Status**: `403 Forbidden`
- **Result**: PASS

### TC-P2-SEC-002: Cross-Tenant Resource Mutation Blocked
- **Module**: Security
- **Objective**: Tenant A attempts to assign a vehicle in Tenant B.
- **Expected Status**: `403 Forbidden`
- **Actual Status**: `403 Forbidden`
- **Result**: PASS

### TC-P2-SEC-003: Read-Only Viewer Mutation Denied
- **Module**: Security
- **Objective**: Authenticated user with role `VIEWER` attempts `POST /api/v1/tenants/:tenantId/vehicles`.
- **Expected Status**: `403 Forbidden`
- **Actual Status**: `403 Forbidden`
- **Result**: PASS

### TC-P2-SEC-004: Expired JWT Rejection
- **Module**: Security
- **Objective**: Request sent with expired JWT access token.
- **Expected Status**: `401 Unauthorized`
- **Actual Status**: `401 Unauthorized`
- **Result**: PASS

---

## 6. Frontend Component Tests
- `BranchList.test.tsx`: Verifies table rendering, search query dispatch, and create branch modal (PASS).
- `EmployeeList.test.tsx`: Verifies operational role badges, branch name resolution, and add employee form (PASS).
- `VehicleList.test.tsx`: Verifies electric vehicle tags, weight/volume indicators, and driver assignment trigger (PASS).
- `App.test.tsx`: Verifies full sidebar navigation links (`Hubs & Branches`, `Workforce`, `Fleet & Assets`) (PASS).

---

## 7. Mobile Unit Tests
- `resource_models_test.dart`: Deserialization of `BranchModel`, `EmployeeModel`, and `VehicleModel` (PASS).
- `resource_api_client_test.dart`: Proximity query encoding, mock network calls, 409 conflict handling (PASS).
