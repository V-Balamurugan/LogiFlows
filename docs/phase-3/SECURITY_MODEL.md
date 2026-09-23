# Phase 3 Security & Authorization Model

**Project**: LogiFlows  
**Phase**: 3 — Employees, Roles, and Vehicles  
**Target Branch**: `feature/phase-3-employees-roles-vehicles`  

---

## 1. Multi-Tenant Isolation Architecture

Multi-tenancy in LogiFlows is enforced strictly at the server level. The client is never trusted for tenant ownership, branch affiliation, or role determination.

### 1.1 Tenant Scoping Principles
1. **Server-Extracted Identity**: The authenticated user's JWT claims provide `user_id`, `tenant_id`, and `role`. The route-level middleware (`RequireTenantMember` or `TenantContext`) validates that the user is an active member of the tenant specified in the path `/api/v1/tenants/{tenant_id}/...`.
2. **Platform Admin Exception**: Only `PLATFORM_ADMIN` users possess cross-tenant query and modification authority. For all other roles, a mismatch between the token's active tenant and the URL tenant returns `403 Forbidden`.
3. **Cross-Tenant Foreign Key Prevention**:
   - When creating or updating an employee, `branch_id` is queried directly with `WHERE id = $1 AND tenant_id = $2`. If the branch belongs to a different tenant, the request is rejected with `400 Bad Request` (`BRANCH_NOT_FOUND` / branch does not belong to tenant).
   - When registering or updating a vehicle, `branch_id` is validated with the exact same tenant-scoping query.
   - When creating a driver-vehicle assignment, the service checks that both the employee and the vehicle share the tenant ID of the current request. Cross-tenant assignments are rejected before any database modification occurs.

---

## 2. Separation of System RBAC vs. Operational Roles

A fundamental security architectural principle in LogiFlows is the strict decoupling of **System Authorization Roles** from **Operational Employee Roles**:

| Category | Roles | Purpose & Scope |
|---|---|---|
| **System RBAC Roles** | `PLATFORM_ADMIN`<br>`TENANT_ADMIN`<br>`TENANT_OPERATOR`<br>`VIEWER` | System-level permissions governing API access, user administration, tenant configuration, and CRUD permissions. |
| **Operational Employee Roles** | `DRIVER`<br>`OPERATOR`<br>`DISPATCHER`<br>`SUPERVISOR`<br>`MANAGER` | Operational job classifications within the logistics network. Governs eligibility for vehicle assignment, custody scans, dispatch authority, etc. |

### 2.1 Permission Matrix

| Operation | Platform Admin | Tenant Admin | Tenant Operator | Viewer |
|---|:---:|:---:|:---:|:---:|
| Create Employee (`POST /employees`) | Allowed | Allowed | Denied (403) | Denied (403) |
| List / View Employees (`GET /employees`) | Allowed | Allowed | Allowed | Allowed |
| Update Employee (`PUT /employees/{id}`) | Allowed | Allowed | Denied (403) | Denied (403) |
| Change Employee Status (`PATCH /employees/{id}/status`) | Allowed | Allowed | Allowed | Denied (403) |
| Deactivate Employee (`DELETE /employees/{id}`) | Allowed | Allowed | Denied (403) | Denied (403) |
| Register Vehicle (`POST /vehicles`) | Allowed | Allowed | Denied (403) | Denied (403) |
| List / View Vehicles (`GET /vehicles`) | Allowed | Allowed | Allowed | Allowed |
| Update Vehicle Specs (`PATCH /vehicles/{id}`) | Allowed | Allowed | Denied (403) | Denied (403) |
| Change Vehicle Status (`PATCH /vehicles/{id}/status`) | Allowed | Allowed | Allowed | Denied (403) |
| Assign Driver to Vehicle (`POST /vehicles/{id}/assign`) | Allowed | Allowed | Allowed | Denied (403) |
| Unassign Driver (`POST /vehicles/{id}/unassign`) | Allowed | Allowed | Allowed | Denied (403) |
| Decommission Vehicle (`DELETE /vehicles/{id}`) | Allowed | Allowed | Denied (403) | Denied (403) |

---

## 3. Concurrency Protection & Double-Booking Prevention

### 3.1 Driver Eligibility Verification
Prior to creating an assignment, the service enforces strict operational prerequisites:
1. Employee's `operational_role` must be exactly `DRIVER`. (Non-driver staff cannot be assigned to commercial vehicles).
2. Employee's `status` must be `ACTIVE`.
3. Employee's `availability_status` must be `AVAILABLE`. (Drivers who are `BUSY`, `OFF_DUTY`, or `UNAVAILABLE` cannot be assigned).
4. Vehicle's `status` must be `AVAILABLE`.
5. Vehicle's `availability_status` must be `AVAILABLE`.

### 3.2 Atomic State Transitions
Assignments are executed within a database transaction:
1. `vehicle_assignments` record inserted with `status = 'ACTIVE'`.
2. Vehicle status updated to `ASSIGNED` and availability updated to `BUSY`.
3. Employee availability updated to `BUSY`.
4. Transaction committed atomically.

### 3.3 Database-Enforced Race Condition Prevention
To prevent concurrent race conditions where two operators attempt to assign the same driver or vehicle simultaneously:
- **Partial Unique Indexes**:
  ```sql
  CREATE UNIQUE INDEX idx_active_vehicle_assignment 
  ON vehicle_assignments (vehicle_id) 
  WHERE status = 'ACTIVE';

  CREATE UNIQUE INDEX idx_active_driver_assignment 
  ON vehicle_assignments (employee_id) 
  WHERE status = 'ACTIVE';
  ```
If two concurrent requests bypass memory checks simultaneously, the second transaction is immediately blocked by the database partial unique index and rolls back, returning `409 Conflict` (`ASSIGNMENT_CONFLICT`).

---

## 4. Atomic Employee Code Generation

Employee code collision during simultaneous staff onboarding is eliminated via tenant-partitioned sequence counters:
```sql
CREATE TABLE IF NOT EXISTS tenant_employee_sequences (
    tenant_id UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    current_val INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```
During onboarding:
```sql
INSERT INTO tenant_employee_sequences (tenant_id, current_val, updated_at)
VALUES ($1, 1, NOW())
ON CONFLICT (tenant_id) DO UPDATE
SET current_val = tenant_employee_sequences.current_val + 1,
    updated_at = NOW()
RETURNING current_val;
```
This PostgreSQL atomic upsert guarantees that concurrent requests within the same tenant receive strictly monotonically increasing integer values (e.g. `EMP-0001`, `EMP-0002`), avoiding duplicate code errors (`23505`).

---

## 5. Security Test Suite Summary

The LogiFlows automated regression and security test suites verify:
1. **SEC-01**: Cross-tenant employee retrieval returns 404 or 403.
2. **SEC-02**: Cross-tenant vehicle creation is blocked by tenant scoping.
3. **SEC-03**: Foreign branch assignment (branch in Tenant B assigned to employee in Tenant A) is rejected with 400 Bad Request.
4. **SEC-04**: Role escalation: `VIEWER` and `TENANT_OPERATOR` attempts to onboard employees or decommission vehicles are rejected with 403 Forbidden.
5. **SEC-05**: Non-driver assignment: Attempting to assign an `OPERATOR` or `DISPATCHER` to a vehicle fails with 400 Bad Request.
6. **SEC-06**: Double-booking conflict: Attempting to assign an already-assigned vehicle returns 409 Conflict.
7. **SEC-07**: Concurrent assignment protection: 10 parallel threads attempting assignment yield exactly 1 success and 9 conflicts.
8. **SEC-08**: Token expiration and invalid signature checks reject requests with 401 Unauthorized.
9. **SEC-09**: Transactional Account Rollback: If user account creation or tenant membership insertion fails during `POST /employees/with-account`, the entire transaction rolls back cleanly without leaving orphaned records.
10. **SEC-10**: Password Security: Passwords hashed with Bcrypt cost 12; hashes and plain-text passwords are never returned or logged.
11. **SEC-11**: Self-Profile Authorization: Users with `EMPLOYEE` role can only access their own `/employees/me` profile and cannot access `/employees` roster or mutate other accounts.
12. **SEC-12**: IDOR Prevention: Manipulating `tenant_id` or `branch_id` in path parameters is rejected by server-side tenancy verification.

---

## 6. Employee Account Creation Transactional Security

The `POST /employees/with-account` workflow executes an all-or-nothing database transaction:
```sql
BEGIN;
-- 1. Create or verify users record
-- 2. Store password hash (Bcrypt cost 12)
-- 3. Insert tenant_memberships record (role = EMPLOYEE / TENANT_OPERATOR)
-- 4. Insert employees profile linking user_id
-- 5. Commit transaction
COMMIT;
```
If duplicate email or validation failure occurs at any stage, the transaction automatically issues a `ROLLBACK`, guaranteeing zero data corruption or unlinked auth records.

---

## 7. The `EMPLOYEE` System Role Security Scope

Users authenticating with the `EMPLOYEE` system authorization role operate in a strictly confined security perimeter:
- Permitted to view `/employees/me` (self profile, assigned branch name, and active vehicle).
- Denied access to `/employees` (workforce roster).
- Denied access to `/tenants/{tenant_id}` administration, invitations, or member role changes.
- Denied access to create, edit, or delete branches and vehicles.

