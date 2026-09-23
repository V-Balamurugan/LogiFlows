# LogiFlows Phase 3 — Role and Permission Matrix

## 1. System Authorization Roles vs. Operational Roles Separation

LogiFlows enforces a strict structural separation between **System Authorization Roles** (which govern system capability, tenant isolation, and API access) and **Operational Roles** (which govern logistical duty, scheduling, and asset handling).

| Concept | System Authorization Role (`system_role`) | Operational Employee Role (`operational_role`) |
| :--- | :--- | :--- |
| **Storage Layer** | `tenant_memberships.role` | `employees.operational_role` |
| **Token Claim** | Evaluated in JWT context & DB membership | Tracked in workforce database & profile |
| **Primary Scope** | API endpoint access & mutation authorization | Workforce scheduling, fleet pairing, task dispatch |
| **Elevation Risk** | **High**: Grants data creation, updates, and management | **Low**: Purely operational classification |

---

## 2. Permitted System Authorization Roles

1. `PLATFORM_ADMIN`: Superuser possessing global administrative privileges across all tenant environments.
2. `TENANT_ADMIN`: Primary organizational administrator with full autonomy over tenant branches, staff, fleet vehicles, and assignments.
3. `TENANT_OPERATOR`: Operational dispatcher/coordinator authorized to manage day-to-day dispatch runs, assign vehicles, update statuses, and monitor hub assets.
4. `VIEWER`: Read-only tenant member permitted to view branches, employee rosters, and fleet records without write/mutate permissions.
5. `EMPLOYEE`: Standard workforce member credential (e.g. drivers, operators) constrained strictly to viewing self profile (`/employees/me`), assigned delivery branch, and assigned fleet vehicle.

---

## 3. Permitted Operational Roles

1. `DRIVER`: Fleet vehicle driver eligible for vehicle dispatch pairings and route executions.
2. `OPERATOR`: Delivery sorting hub and package intake handler.
3. `DISPATCHER`: Operational coordinator overseeing linehaul and local delivery movements.
4. `SUPERVISOR`: Delivery hub supervisor responsible for shift attendance and driver availability.
5. `MANAGER`: Operational territory manager.
6. `BRANCH_MANAGER`: Manager assigned to lead and audit a specific branch sorting facility.
7. `WAREHOUSE_OPERATOR`: Specialized storage and cross-docking facility personnel.
8. `DELIVERY_EXECUTIVE`: Last-mile delivery runner and customer-facing delivery agent.

---

## 4. Fine-Grained Permission Matrix

| Resource / Action | Permission Key | `PLATFORM_ADMIN` | `TENANT_ADMIN` | `TENANT_OPERATOR` | `VIEWER` | `EMPLOYEE` |
| :--- | :--- | :---: | :---: | :---: | :---: | :---: |
| **Branch: Create** | `branch:create` | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Branch: List & Search** | `branch:read` | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Branch: Update** | `branch:update` | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Branch: Deactivate/Delete**| `branch:delete` | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Branch: View Employees** | `branch:read_staff` | ✅ | ✅ | ✅ | ✅ | ❌ |
| **Branch: View Vehicles** | `branch:read_fleet` | ✅ | ✅ | ✅ | ✅ | ❌ |
| **Employee: Create** | `employee:create` | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Employee: Create w/ Account**| `employee:create_account`| ✅ | ✅ | ❌ | ❌ | ❌ |
| **Employee: List & Search** | `employee:read` | ✅ | ✅ | ✅ | ✅ | ❌ |
| **Employee: View Me/Self** | `employee:read_self` | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Employee: Update Status** | `employee:update_status` | ✅ | ✅ | ✅ | ❌ | ❌ |
| **Employee: Account Status**| `employee:account_status`| ✅ | ✅ | ❌ | ❌ | ❌ |
| **Employee: Deactivate** | `employee:delete` | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Vehicle: Create** | `vehicle:create` | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Vehicle: List & Search** | `vehicle:read` | ✅ | ✅ | ✅ | ✅ | ✅ (Own) |
| **Vehicle: Update Status** | `vehicle:update_status` | ✅ | ✅ | ✅ | ❌ | ❌ |
| **Vehicle: Deactivate** | `vehicle:delete` | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Assignment: Create** | `assignment:create` | ✅ | ✅ | ✅ | ❌ | ❌ |
| **Assignment: Unassign** | `assignment:update` | ✅ | ✅ | ✅ | ❌ | ❌ |
| **Assignment: History** | `assignment:read` | ✅ | ✅ | ✅ | ✅ | ❌ |

---

## 5. Security & Elevation Controls

1. **Self-Escalation Prohibition**:
   Authenticated users cannot modify their own `tenant_memberships.role` or operational role. All account provisioning must be performed by a verified `TENANT_ADMIN` or `PLATFORM_ADMIN`.
2. **Strict Server-Side Enforcement**:
   All permissions are verified through Go middleware (`RequireRole`, `RequireTenantMembership`, `RequireTenantContext`). Client-side UI toggles and badges are strictly UX conveniences.
3. **Double-Booking & Conflict Checks**:
   Vehicle assignments check database concurrency with active row status locks to prevent simultaneous driver assignments.
4. **Tenant Isolation**:
   Every database query references `tenant_id = $1` derived exclusively from the authenticated JWT session or verified tenant parameter. Cross-tenant data leakage yields strict `403 Forbidden` or `404 Not Found` responses.
