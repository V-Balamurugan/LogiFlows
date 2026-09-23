# Phase 3 Detailed Test Cases & Execution Matrix

**Project**: LogiFlows  
**Phase**: 3 — Employees, Roles, and Vehicles  
**Target Branch**: `feature/phase-3-employees-roles-vehicles`  

---

## 1. Employee Management Test Cases (US-E01 — US-E09)

| Test ID | User Story | Description | Preconditions | Inputs / Request | Expected Output | Status |
|---|---|---|---|---|---|:---:|
| **TC-E01** | US-E01 | Create employee with valid data | Tenant Admin JWT | `POST /employees`<br>`first_name: "Kavitha"`, `operational_role: "DRIVER"` | HTTP 201 Created; profile created with status `ACTIVE`. | **PASS** |
| **TC-E02** | US-E02 | Auto-generate unique employee code | Tenant Admin JWT | `POST /employees`<br>`employee_code: ""` (blank) | HTTP 201 Created; `employee_code` formatted as `EMP-XXXX` (e.g. `EMP-0001`). | **PASS** |
| **TC-E03** | US-E03 | Validate branch belongs to current tenant | Tenant Admin JWT | `POST /employees`<br>`branch_id: <foreign_tenant_branch>` | HTTP 400 Bad Request; error code `BRANCH_NOT_FOUND` / tenant mismatch. | **PASS** |
| **TC-E04** | US-E04 | View authorized employee directory | Authenticated user | `GET /employees` | HTTP 200 OK; lists only current tenant employees. | **PASS** |
| **TC-E05** | US-E05 | Update employee profile information | Tenant Admin JWT | `PUT /employees/{id}`<br>`first_name: "Kavitha Updated"` | HTTP 200 OK; updated name returned in response envelope. | **PASS** |
| **TC-E06** | US-E06 | Activate/deactivate employee (soft-delete) | Tenant Admin JWT | `DELETE /employees/{id}` | HTTP 200 OK; `status = 'TERMINATED'`, `is_active = false`. | **PASS** |
| **TC-E07** | US-E07 | Filter employees by branch, role, and status | Operator JWT | `GET /employees?operational_role=DRIVER` | HTTP 200 OK; all returned records have role `DRIVER`. | **PASS** |
| **TC-E08** | US-E08 | Retrieve verified available drivers | Operator JWT | `GET /employees/available-drivers` | HTTP 200 OK; returns only drivers with status `ACTIVE` and availability `AVAILABLE`. | **PASS** |
| **TC-E09** | US-E09 | Prevent cross-tenant employee access | User from Tenant B | `GET /tenants/{tenantA}/employees/{empId}` | HTTP 403 Forbidden or 404 Not Found; zero data leakage. | **PASS** |

---

## 2. Vehicle Fleet Test Cases (US-V01 — US-V09)

| Test ID | User Story | Description | Preconditions | Inputs / Request | Expected Output | Status |
|---|---|---|---|---|---|:---:|
| **TC-V01** | US-V01 | Register fleet vehicle with valid specs | Tenant Admin JWT | `POST /vehicles`<br>`registration_number: "TN01AB1234"`, `type: "VAN"` | HTTP 201 Created; vehicle stored with status `AVAILABLE`. | **PASS** |
| **TC-V02** | US-V02 | Enforce positive weight and volume capacity | Tenant Admin JWT | `POST /vehicles`<br>`max_weight_kg: -50.0` | HTTP 400 Bad Request; validation error on weight. | **PASS** |
| **TC-V03** | US-V03 | Validate branch association matches tenant | Tenant Admin JWT | `POST /vehicles`<br>`branch_id: <foreign_tenant_branch>` | HTTP 400 Bad Request; foreign branch rejected. | **PASS** |
| **TC-V04** | US-V04 | View available fleet vehicles | Operator JWT | `GET /vehicles?status=AVAILABLE` | HTTP 200 OK; returns list of available fleet vehicles. | **PASS** |
| **TC-V05** | US-V05 | Update vehicle specifications | Tenant Admin JWT | `PATCH /vehicles/{id}`<br>`max_weight_kg: 1500.0` | HTTP 200 OK; payload updated in database. | **PASS** |
| **TC-V06** | US-V06 | Change vehicle operating status | Operator JWT | `PATCH /vehicles/{id}/status`<br>`status: "MAINTENANCE"` | HTTP 200 OK; status updated to `MAINTENANCE`. | **PASS** |
| **TC-V07** | US-V07 | Assign eligible driver to vehicle | Operator JWT | `POST /vehicles/{id}/assign`<br>`driver_id: <eligible_driver_id>` | HTTP 201 Created; driver marked `BUSY`, vehicle marked `ASSIGNED`. | **PASS** |
| **TC-V08** | US-V08 | Prevent double-booking / conflicting assignments | Operator JWT | `POST /vehicles/{assigned_id}/assign`<br>`driver_id: <any_driver>` | HTTP 409 Conflict; assignment refused with conflict message. | **PASS** |
| **TC-V09** | US-V09 | Prevent cross-tenant vehicle access | User from Tenant B | `GET /tenants/{tenantA}/vehicles/{vehId}` | HTTP 403 Forbidden or 404 Not Found; zero data leakage. | **PASS** |

---

## 3. Concurrency & Security Test Cases

| Test ID | Category | Description | Execution Parameters | Expected Output | Status |
|---|---|---|---|---|:---:|
| **TC-CONC-01** | Concurrency | 10 concurrent requests to auto-generate employee codes | 10 concurrent goroutines calling `Create` with blank code | All 10 codes unique, strictly sequential, no duplicate key errors (`23505`). | **PASS** |
| **TC-CONC-02** | Concurrency | 10 concurrent requests attempting to assign the same vehicle | 10 concurrent goroutines calling `AssignDriver` | Exactly 1 HTTP 201 Success, exactly 9 HTTP 409 Conflicts. | **PASS** |
| **TC-SEC-01** | RBAC | Viewer attempts to onboard employee | `VIEWER` JWT calling `POST /employees` | HTTP 403 Forbidden. | **PASS** |
| **TC-SEC-02** | RBAC | Operator attempts to decommission vehicle | `TENANT_OPERATOR` calling `DELETE /vehicles/{id}` | HTTP 403 Forbidden. | **PASS** |
| **TC-SEC-03** | Eligibility | Assign non-driver employee (`OPERATOR`) as driver | Valid operator ID passed to `/assign` | HTTP 400 Bad Request (`DRIVER_INELIGIBLE`). | **PASS** |
| **TC-SEC-04** | Eligibility | Assign off-duty driver (`OFF_DUTY`) as driver | Driver with `availability_status: 'OFF_DUTY'` | HTTP 400 Bad Request (`DRIVER_NOT_AVAILABLE`). | **PASS** |
