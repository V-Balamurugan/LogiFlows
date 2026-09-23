# Phase 3 API Request & Response Contracts

**Project**: LogiFlows — Intelligent End-to-End Logistics Coordination  
**Phase**: 3 — Employees, Roles, and Vehicles  
**Target Branch**: `feature/phase-3-employees-roles-vehicles`  
**Base URL**: `/api/v1`  
**Standard Response Envelope**:
```json
{
  "data": { ... },
  "message": "Human-readable description"
}
```
**Standard Error Envelope**:
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error description",
    "request_id": "00000000-0000-0000-0000-000000000000",
    "details": {}
  }
}
```

---

## 1. Employee Endpoints

### 1.1 Onboard Employee Profile
- **Method**: `POST`
- **URL**: `/api/v1/tenants/{tenant_id}/employees`
- **Purpose**: Creates an employee profile within a tenant and assigns them to an active branch.
- **Authentication**: `Bearer <JWT>`
- **Required System Role**: `PLATFORM_ADMIN`, `TENANT_ADMIN`
- **Required Permission**: `employees:write`
- **Tenant Scope**: Must match token claims or token must be Platform Admin.
- **Branch Scope**: Branch specified in body must belong to the target tenant.
- **Request Headers**:
  - `Authorization: Bearer <token>`
  - `Content-Type: application/json`
- **Path Parameters**:
  - `tenant_id` (UUID): Tenant ID.
- **Request Body**:
```json
{
  "first_name": "Arun",
  "last_name": "Kumar",
  "operational_role": "DRIVER",
  "employment_type": "FULL_TIME",
  "branch_id": "90e2b34a-9b16-4d2c-880e-3fa1a684b001",
  "employee_code": "EMP-0012",
  "designation": "Senior Dispatch Driver",
  "email": "arun.kumar@quicklogistics.com",
  "phone": "+919876543210",
  "license_number": "TN01-2020-0012345",
  "joining_date": "2026-03-01"
}
```
*Note*: `employee_code` is optional; if omitted, the system atomically generates an incrementing code `EMP-XXXX` backed by `tenant_employee_sequences`.

- **Success Response (201 Created)**:
```json
{
  "data": {
    "id": "e43b1234-5678-4a9c-9801-112233445566",
    "tenant_id": "00000000-0000-0000-0000-000000000001",
    "branch_id": "90e2b34a-9b16-4d2c-880e-3fa1a684b001",
    "branch_name": "North Hub",
    "employee_code": "EMP-0012",
    "first_name": "Arun",
    "last_name": "Kumar",
    "designation": "Senior Dispatch Driver",
    "operational_role": "DRIVER",
    "employment_type": "FULL_TIME",
    "status": "ACTIVE",
    "availability_status": "AVAILABLE",
    "verification_status": "VERIFIED",
    "joining_date": "2026-03-01",
    "is_active": true,
    "created_at": "2026-09-23T10:15:30Z"
  },
  "message": "Employee created successfully"
}
```
- **Error Responses**:
  - `400 Bad Request`: Validation failure or branch belongs to another tenant.
  - `401 Unauthorized`: Missing or invalid JWT.
  - `403 Forbidden`: Cross-tenant manipulation attempt or insufficient role.
  - `409 Conflict`: Duplicate employee code within the tenant.

---

### 1.2 List Employees
- **Method**: `GET`
- **URL**: `/api/v1/tenants/{tenant_id}/employees`
- **Purpose**: Lists employees matching optional search/filter criteria.
- **Authentication**: `Bearer <JWT>`
- **Required System Role**: `PLATFORM_ADMIN`, `TENANT_ADMIN`, `TENANT_OPERATOR`, `VIEWER`
- **Query Parameters**:
  - `operational_role` (string, optional): `DRIVER`, `OPERATOR`, `DISPATCHER`, `SUPERVISOR`, `MANAGER`
  - `availability_status` (string, optional): `AVAILABLE`, `BUSY`, `OFF_DUTY`, `UNAVAILABLE`
  - `branch_id` (UUID, optional): Scoped branch filter
  - `status` (string, optional): `ACTIVE`, `INACTIVE`, `SUSPENDED`, `TERMINATED`
  - `search` (string, optional): Search by name or code
  - `limit` (integer, default 20)
  - `offset` (integer, default 0)
- **Success Response (200 OK)**:
```json
{
  "data": {
    "employees": [
      {
        "id": "e43b1234-5678-4a9c-9801-112233445566",
        "tenant_id": "00000000-0000-0000-0000-000000000001",
        "branch_id": "90e2b34a-9b16-4d2c-880e-3fa1a684b001",
        "branch_name": "North Hub",
        "employee_code": "EMP-0012",
        "first_name": "Arun",
        "last_name": "Kumar",
        "designation": "Senior Dispatch Driver",
        "operational_role": "DRIVER",
        "employment_type": "FULL_TIME",
        "status": "ACTIVE",
        "availability_status": "AVAILABLE",
        "verification_status": "VERIFIED",
        "is_active": true,
        "created_at": "2026-09-23T10:15:30Z"
      }
    ],
    "total": 1
  }
}
```

---

### 1.3 List Available Drivers
- **Method**: `GET`
- **URL**: `/api/v1/tenants/{tenant_id}/employees/available-drivers`
- **Purpose**: Returns active, verified employees who hold the `DRIVER` operational role and are currently `AVAILABLE`.
- **Authentication**: `Bearer <JWT>`
- **Query Parameters**:
  - `branch_id` (UUID, optional): Scopes available drivers to a specific branch.
- **Success Response (200 OK)**:
```json
{
  "data": {
    "drivers": [
      {
        "id": "e43b1234-5678-4a9c-9801-112233445566",
        "employee_code": "EMP-0012",
        "first_name": "Arun",
        "last_name": "Kumar",
        "operational_role": "DRIVER",
        "branch_id": "90e2b34a-9b16-4d2c-880e-3fa1a684b001",
        "branch_name": "North Hub",
        "availability_status": "AVAILABLE",
        "verification_status": "VERIFIED",
        "status": "ACTIVE"
      }
    ],
    "total": 1
  }
}
```

---

### 1.4 Get Employee Details
- **Method**: `GET`
- **URL**: `/api/v1/tenants/{tenant_id}/employees/{employee_id}`
- **Success Response (200 OK)**:
```json
{
  "data": {
    "id": "e43b1234-5678-4a9c-9801-112233445566",
    "tenant_id": "00000000-0000-0000-0000-000000000001",
    "branch_id": "90e2b34a-9b16-4d2c-880e-3fa1a684b001",
    "employee_code": "EMP-0012",
    "first_name": "Arun",
    "last_name": "Kumar",
    "operational_role": "DRIVER",
    "status": "ACTIVE",
    "availability_status": "AVAILABLE",
    "verification_status": "VERIFIED",
    "is_active": true
  }
}
```

---

### 1.5 Update Employee Profile
- **Method**: `PUT`
- **URL**: `/api/v1/tenants/{tenant_id}/employees/{employee_id}`
- **Purpose**: Modifies employee personal details, branch, or designation.
- **Request Body**:
```json
{
  "first_name": "Arun",
  "last_name": "Kumar",
  "designation": "Lead Courier Driver",
  "branch_id": "90e2b34a-9b16-4d2c-880e-3fa1a684b001",
  "email": "arun.kumar@quicklogistics.com",
  "phone": "+919876543210",
  "license_number": "TN01-2020-0012345"
}
```
- **Success Response (200 OK)**: Returns updated `Employee` envelope.

---

### 1.6 Update Employee Status & Availability
- **Method**: `PATCH`
- **URL**: `/api/v1/tenants/{tenant_id}/employees/{employee_id}/status`
- **Purpose**: Updates operational availability, employment status, or KYC verification status.
- **Request Body**:
```json
{
  "availability_status": "OFF_DUTY",
  "status": "ACTIVE",
  "verification_status": "VERIFIED"
}
```
- **Success Response (200 OK)**:
```json
{
  "data": {
    "id": "e43b1234-5678-4a9c-9801-112233445566",
    "status": "ACTIVE",
    "availability_status": "OFF_DUTY",
    "verification_status": "VERIFIED"
  },
  "message": "Employee status updated successfully"
}
```

---

### 1.7 Deactivate Employee (Soft Delete)
- **Method**: `DELETE`
- **URL**: `/api/v1/tenants/{tenant_id}/employees/{employee_id}`
- **Purpose**: Soft-deactivates an employee (`status = 'TERMINATED'`, `is_active = false`, `deleted_at = NOW()`).
- **Success Response (200 OK)**:
```json
{
  "data": {},
  "message": "Employee deactivated successfully"
}
```

---

### 1.8 Create Employee Profile With Login Account
- **Method**: `POST`
- **URL**: `/api/v1/tenants/{tenant_id}/employees/with-account`
- **Purpose**: Atomically provisions a user account, tenant membership, and employee profile in a single database transaction.
- **Authentication**: `Bearer <JWT>`
- **Required System Role**: `PLATFORM_ADMIN`, `TENANT_ADMIN`
- **Required Permission**: `employees:write`, `memberships:write`
- **Request Body**:
```json
{
  "first_name": "Priya",
  "last_name": "Nair",
  "email": "priya.nair@quicklogistics.com",
  "phone": "+919876543211",
  "designation": "Central Dispatch Coordinator",
  "operational_role": "DISPATCHER",
  "system_role": "TENANT_OPERATOR",
  "password": "SecurePassword123!",
  "send_invite": false,
  "branch_id": "90e2b34a-9b16-4d2c-880e-3fa1a684b001"
}
```
- **Security Invariants**:
  - Passwords hashed using Bcrypt (cost 12).
  - Password and hash are **never** returned in the response or written to logs.
  - Complete transaction roll back occurs if user creation, tenant membership insertion, or employee record creation fails.
- **Success Response (201 Created)**:
```json
{
  "data": {
    "id": "e43b1234-5678-4a9c-9801-998877665544",
    "tenant_id": "00000000-0000-0000-0000-000000000001",
    "user_id": "u1234567-89ab-cdef-0123-456789abcdef",
    "branch_id": "90e2b34a-9b16-4d2c-880e-3fa1a684b001",
    "branch_name": "North Hub",
    "employee_code": "EMP-0013",
    "first_name": "Priya",
    "last_name": "Nair",
    "email": "priya.nair@quicklogistics.com",
    "designation": "Central Dispatch Coordinator",
    "operational_role": "DISPATCHER",
    "employment_type": "FULL_TIME",
    "status": "ACTIVE",
    "availability_status": "AVAILABLE",
    "verification_status": "VERIFIED",
    "is_active": true,
    "created_at": "2026-09-23T11:45:00Z"
  },
  "message": "Employee and login credentials created successfully"
}
```
- **Error Responses**:
  - `400 Bad Request`: Missing email or password shorter than 8 characters.
  - `409 Conflict`: User email or employee code already exists.

---

### 1.9 Get Current Authenticated Employee Self-Profile
- **Method**: `GET`
- **URL**: `/api/v1/tenants/{tenant_id}/employees/me`
- **Purpose**: Returns the calling employee's profile, operational role, system authorization role, and currently assigned fleet vehicle.
- **Authentication**: `Bearer <JWT>`
- **Required System Role**: Any authenticated tenant member (`EMPLOYEE`, `TENANT_OPERATOR`, `TENANT_ADMIN`, `VIEWER`, `PLATFORM_ADMIN`).
- **Success Response (200 OK)**:
```json
{
  "data": {
    "employee": {
      "id": "e43b1234-5678-4a9c-9801-112233445566",
      "tenant_id": "00000000-0000-0000-0000-000000000001",
      "branch_id": "90e2b34a-9b16-4d2c-880e-3fa1a684b001",
      "branch_name": "North Hub",
      "employee_code": "EMP-0012",
      "first_name": "Arun",
      "last_name": "Kumar",
      "designation": "Senior Dispatch Driver",
      "operational_role": "DRIVER",
      "status": "ACTIVE",
      "availability_status": "AVAILABLE",
      "verification_status": "VERIFIED",
      "is_active": true
    },
    "system_role": "EMPLOYEE",
    "assigned_vehicle": {
      "id": "7fa12345-6789-4bcd-88ef-998877665544",
      "registration_number": "TN01AB1234",
      "vehicle_type": "ELECTRIC_VAN",
      "make_model": "Tata Ace EV",
      "status": "ACTIVE"
    }
  }
}
```

---

### 1.10 Get Employee Account & Credential Status
- **Method**: `GET`
- **URL**: `/api/v1/tenants/{tenant_id}/employees/{employee_id}/account-status`
- **Purpose**: Returns whether an employee has linked login credentials, their active user status, and system role.
- **Authentication**: `Bearer <JWT>`
- **Required System Role**: `PLATFORM_ADMIN`, `TENANT_ADMIN`
- **Success Response (200 OK)**:
```json
{
  "data": {
    "employee_id": "e43b1234-5678-4a9c-9801-998877665544",
    "employee_code": "EMP-0013",
    "full_name": "Priya Nair",
    "operational_role": "DISPATCHER",
    "status": "ACTIVE",
    "has_account": true,
    "user_id": "u1234567-89ab-cdef-0123-456789abcdef",
    "user_email": "priya.nair@quicklogistics.com",
    "user_is_active": true,
    "system_role": "TENANT_OPERATOR",
    "branch_id": "90e2b34a-9b16-4d2c-880e-3fa1a684b001",
    "branch_name": "North Hub"
  }
}
```

---

## 2. Vehicle Endpoints

### 2.1 Register Fleet Vehicle
- **Method**: `POST`
- **URL**: `/api/v1/tenants/{tenant_id}/vehicles`
- **Purpose**: Registers a fleet vehicle under a specific tenant and operational branch.
- **Request Body**:
```json
{
  "registration_number": "TN01AB1234",
  "vehicle_type": "ELECTRIC_VAN",
  "make_model": "Tata Ace EV",
  "year": 2025,
  "max_weight_kg": 1000.0,
  "max_volume_cbm": 6.5,
  "branch_id": "90e2b34a-9b16-4d2c-880e-3fa1a684b001"
}
```
- **Success Response (201 Created)**:
```json
{
  "data": {
    "id": "7fa12345-6789-4bcd-88ef-998877665544",
    "tenant_id": "00000000-0000-0000-0000-000000000001",
    "assigned_branch_id": "90e2b34a-9b16-4d2c-880e-3fa1a684b001",
    "branch_name": "North Hub",
    "registration_number": "TN01AB1234",
    "vehicle_type": "ELECTRIC_VAN",
    "make_model": "Tata Ace EV",
    "year": 2025,
    "max_weight_kg": 1000.0,
    "max_volume_cbm": 6.5,
    "status": "AVAILABLE",
    "availability_status": "AVAILABLE",
    "is_active": true,
    "created_at": "2026-09-23T11:00:00Z"
  },
  "message": "Vehicle registered successfully"
}
```
- **Error Responses**:
  - `400 Bad Request`: Negative payload capacity or foreign branch ID.
  - `409 Conflict`: Duplicate normalized registration number within tenant.

---

### 2.2 List Fleet Vehicles
- **Method**: `GET`
- **URL**: `/api/v1/tenants/{tenant_id}/vehicles`
- **Query Parameters**:
  - `vehicle_type` (string, optional): `ELECTRIC_VAN`, `VAN`, `MOTORCYCLE`, `TRUCK`, `THREE_WHEELER`
  - `status` (string, optional): `AVAILABLE`, `ASSIGNED`, `IN_TRANSIT`, `MAINTENANCE`, `DECOMMISSIONED`
  - `availability_status` (string, optional): `AVAILABLE`, `BUSY`, `MAINTENANCE`, `OUT_OF_SERVICE`
  - `branch_id` (UUID, optional)
  - `search` (string, optional): Search by registration or model
  - `limit` (integer, default 20)
  - `offset` (integer, default 0)
- **Success Response (200 OK)**:
```json
{
  "data": {
    "vehicles": [
      {
        "id": "7fa12345-6789-4bcd-88ef-998877665544",
        "registration_number": "TN01AB1234",
        "vehicle_type": "ELECTRIC_VAN",
        "status": "AVAILABLE",
        "availability_status": "AVAILABLE",
        "current_driver_name": null,
        "current_driver_code": null,
        "is_active": true
      }
    ],
    "total": 1
  }
}
```

---

### 2.3 Update Vehicle Specifications
- **Method**: `PATCH`
- **URL**: `/api/v1/tenants/{tenant_id}/vehicles/{vehicle_id}`
- **Request Body**:
```json
{
  "max_weight_kg": 1200.0,
  "max_volume_cbm": 7.0,
  "make_model": "Tata Ace EV Long Range"
}
```
- **Success Response (200 OK)**: Returns updated `Vehicle` envelope.

---

### 2.4 Update Vehicle Status
- **Method**: `PATCH`
- **URL**: `/api/v1/tenants/{tenant_id}/vehicles/{vehicle_id}/status`
- **Request Body**:
```json
{
  "status": "MAINTENANCE",
  "availability_status": "MAINTENANCE"
}
```
- **Success Response (200 OK)**: Returns updated `Vehicle` envelope.

---

### 2.5 Decommission Vehicle
- **Method**: `DELETE`
- **URL**: `/api/v1/tenants/{tenant_id}/vehicles/{vehicle_id}`
- **Success Response (200 OK)**:
```json
{
  "data": {},
  "message": "Vehicle decommissioned successfully"
}
```

---

## 3. Assignment Endpoints

### 3.1 Assign Driver to Vehicle
- **Method**: `POST`
- **URL**: `/api/v1/tenants/{tenant_id}/vehicles/{vehicle_id}/assign`
- **Purpose**: Atomically creates an active assignment between an available, verified driver and an available vehicle.
- **Request Body**:
```json
{
  "driver_id": "e43b1234-5678-4a9c-9801-112233445566",
  "notes": "Morning route delivery shift"
}
```
- **Business Validations**:
  1. Driver and vehicle must exist in the same tenant.
  2. Employee must have `operational_role = 'DRIVER'`.
  3. Employee must have `status = 'ACTIVE'` and `availability_status = 'AVAILABLE'`.
  4. Vehicle must have `status = 'AVAILABLE'` and `availability_status = 'AVAILABLE'`.
- **State Transitions**:
  - Driver `availability_status` becomes `BUSY`.
  - Vehicle `status` becomes `ASSIGNED` and `availability_status` becomes `BUSY`.
- **Success Response (201 Created)**:
```json
{
  "data": {},
  "message": "Driver assigned successfully"
}
```
- **Error Responses**:
  - `400 Bad Request`: Ineligible employee (e.g. `OPERATOR` instead of `DRIVER`, or unverified/inactive).
  - `409 Conflict`: Vehicle or driver already actively assigned (double-booking prevented).

---

### 3.2 Unassign Driver from Vehicle
- **Method**: `POST`
- **URL**: `/api/v1/tenants/{tenant_id}/vehicles/{vehicle_id}/unassign`
- **Purpose**: Terminates active assignment. Restores both vehicle and driver to `AVAILABLE`.
- **Success Response (200 OK)**:
```json
{
  "data": {},
  "message": "Driver unassigned successfully"
}
```

---

### 3.3 List Vehicle Assignments
- **Method**: `GET`
- **URL**: `/api/v1/tenants/{tenant_id}/assignments`
- **Query Parameters**:
  - `vehicle_id` (UUID, optional)
  - `employee_id` (UUID, optional)
  - `status` (string, optional): `ACTIVE`, `COMPLETED`, `TERMINATED`
  - `limit` (integer, default 20)
  - `offset` (integer, default 0)
- **Success Response (200 OK)**:
```json
{
  "data": {
    "assignments": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "tenant_id": "00000000-0000-0000-0000-000000000001",
        "vehicle_id": "7fa12345-6789-4bcd-88ef-998877665544",
        "employee_id": "e43b1234-5678-4a9c-9801-112233445566",
        "registration_number": "TN01AB1234",
        "vehicle_type": "ELECTRIC_VAN",
        "driver_name": "Arun Kumar",
        "driver_code": "EMP-0012",
        "assigned_at": "2026-09-23T11:30:00Z",
        "unassigned_at": null,
        "status": "ACTIVE",
        "notes": "Morning route delivery shift"
      }
    ],
    "total": 1
  }
}
```

---

### 3.4 Get Vehicle Assignment History
- **Method**: `GET`
- **URL**: `/api/v1/tenants/{tenant_id}/vehicles/{vehicle_id}/assignments`
- **Success Response (200 OK)**: Same schema as 3.3 filtered to specific vehicle.

---

## 4. Branch Asset Endpoints

### 4.1 List Branch Personnel
- **Method**: `GET`
- **URL**: `/api/v1/tenants/{tenant_id}/branches/{branch_id}/employees`
- **Purpose**: Returns all employees assigned to a specific branch sorting hub.
- **Authentication**: `Bearer <JWT>`
- **Required System Role**: `PLATFORM_ADMIN`, `TENANT_ADMIN`, `TENANT_OPERATOR`, `VIEWER`
- **Success Response (200 OK)**:
```json
{
  "data": {
    "branch_id": "90e2b34a-9b16-4d2c-880e-3fa1a684b001",
    "employees": [
      {
        "id": "e43b1234-5678-4a9c-9801-112233445566",
        "employee_code": "EMP-0012",
        "first_name": "Arun",
        "last_name": "Kumar",
        "email": "arun.kumar@quicklogistics.com",
        "phone": "+919876543210",
        "designation": "Senior Dispatch Driver",
        "operational_role": "DRIVER",
        "status": "ACTIVE",
        "availability_status": "AVAILABLE",
        "is_active": true
      }
    ],
    "total": 1
  }
}
```

---

### 4.2 List Branch Stationed Vehicles
- **Method**: `GET`
- **URL**: `/api/v1/tenants/{tenant_id}/branches/{branch_id}/vehicles`
- **Purpose**: Returns all fleet vehicles stationed at a specific branch sorting hub.
- **Authentication**: `Bearer <JWT>`
- **Required System Role**: `PLATFORM_ADMIN`, `TENANT_ADMIN`, `TENANT_OPERATOR`, `VIEWER`
- **Success Response (200 OK)**:
```json
{
  "data": {
    "branch_id": "90e2b34a-9b16-4d2c-880e-3fa1a684b001",
    "vehicles": [
      {
        "id": "7fa12345-6789-4bcd-88ef-998877665544",
        "registration_number": "TN01AB1234",
        "vehicle_type": "ELECTRIC_VAN",
        "make_model": "Tata Ace EV",
        "year": 2025,
        "max_weight_kg": 1000.0,
        "max_volume_cbm": 6.5,
        "status": "ACTIVE",
        "availability_status": "AVAILABLE",
        "is_active": true,
        "current_driver_name": "Arun Kumar"
      }
    ],
    "total": 1
  }
}
```

