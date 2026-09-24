# LogiFlows API Specification — Phase 1: Authentication & Tenancy

## 1. Overview & Conventions
- **Base URL**: `/api/v1`
- **Envelope Standard**: All responses use consistent JSON envelopes with `status`, `service`, `version`, `timestamp`, and `data` or `error`.
- **Request ID Tracking**: Header `X-Request-ID` is extracted or assigned, propagated in response headers and error envelopes.

---

## 2. Authentication Endpoints

### 2.1 Register Company & Initial User
- **Endpoint**: `POST /api/v1/auth/register`
- **Access**: Public
- **Description**: Atomically registers a new user, provisions their company tenant, and assigns them as `TENANT_ADMIN`.
- **Request Body**:
  ```json
  {
    "email": "owner@quickcargo.com",
    "password": "SecureP@ssw0rd!2026",
    "full_name": "V. Balamurugan",
    "phone_number": "+91-9876543210",
    "company_name": "QuickCargo Logistics Ltd"
  }
  ```
- **Response (201 Created)**:
  ```json
  {
    "status": "success",
    "service": "logiflows-api",
    "version": "1.0.0",
    "timestamp": "2026-09-21T15:00:00Z",
    "data": {
      "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "expires_at": "2026-09-22T15:00:00Z",
      "refresh_token": "a1b2c3d4e5f6...64hexchars",
      "refresh_token_expires_at": "2026-09-28T15:00:00Z",
      "user": {
        "id": "a50c822e-b6a8-48dc-8ce2-c11df241bf11",
        "email": "owner@quickcargo.com",
        "full_name": "V. Balamurugan",
        "phone_number": "+91-9876543210",
        "is_active": true,
        "is_platform_admin": false,
        "email_verified": false
      },
      "tenants": [
        {
          "id": "e40d822e-c7b9-49ed-9df3-d22ef352cf22",
          "name": "QuickCargo Logistics Ltd",
          "slug": "quickcargo-logistics-ltd",
          "role": "TENANT_ADMIN"
        }
      ]
    }
  }
  ```

---

### 2.2 User Login
- **Endpoint**: `POST /api/v1/auth/login`
- **Access**: Public
- **Description**: Validates credentials using constant-time bcrypt comparison, returns signed JWT access token and opaque refresh token.
- **Request Body**:
  ```json
  {
    "email": "owner@quickcargo.com",
    "password": "SecureP@ssw0rd!2026"
  }
  ```
- **Response (200 OK)**: Same payload structure as 2.1.
- **Error Response (401 Unauthorized)**:
  ```json
  {
    "error": {
      "code": "UNAUTHORIZED",
      "message": "Invalid email or password",
      "request_id": "req-98765"
    }
  }
  ```

---

### 2.3 Single-Use Token Refresh
- **Endpoint**: `POST /api/v1/auth/refresh`
- **Access**: Public
- **Description**: Validates opaque refresh token, revokes it, and rotates to a brand-new token pair. If a revoked token is presented, triggers breach detection and invalidates all user tokens.
- **Request Body**:
  ```json
  {
    "refresh_token": "a1b2c3d4e5f6...64hexchars"
  }
  ```
- **Response (200 OK)**: Rotated access token + fresh refresh token.

---

### 2.4 Current User Profile
- **Endpoint**: `GET /api/v1/auth/me`
- **Access**: Bearer Token
- **Headers**: `Authorization: Bearer <access_token>`
- **Response (200 OK)**:
  ```json
  {
    "status": "success",
    "data": {
      "user": {
        "id": "a50c822e-b6a8-48dc-8ce2-c11df241bf11",
        "email": "owner@quickcargo.com",
        "full_name": "V. Balamurugan",
        "is_active": true,
        "is_platform_admin": false,
        "email_verified": false
      },
      "tenants": [
        {
          "id": "e40d822e-c7b9-49ed-9df3-d22ef352cf22",
          "name": "QuickCargo Logistics Ltd",
          "slug": "quickcargo-logistics-ltd",
          "role": "TENANT_ADMIN"
        }
      ]
    }
  }
  ```

---

### 2.5 Logout & Session Invalidation
- **Endpoint**: `POST /api/v1/auth/logout`
- **Access**: Bearer Token
- **Headers**: `Authorization: Bearer <access_token>`
- **Request Body (Optional)**:
  ```json
  {
    "refresh_token": "a1b2c3d4e5f6...64hexchars"
  }
  ```
- **Response (200 OK)**:
  ```json
  {
    "status": "success",
    "data": {
      "message": "Successfully logged out. Refresh token and session revoked."
    }
  }
  ```

---

## 3. Tenancy Endpoints

### 3.1 Create Additional Tenant Company
- **Endpoint**: `POST /api/v1/tenants`
- **Access**: Bearer Token
- **Request Body**:
  ```json
  {
    "name": "Apex Delivery Hub",
    "contact_email": "contact@apexdelivery.com"
  }
  ```
- **Response (201 Created)**: Created tenant entity with caller assigned `TENANT_ADMIN`.

---

### 3.2 List Accessible Tenant Companies
- **Endpoint**: `GET /api/v1/tenants`
- **Access**: Bearer Token
- **Description**: Returns all companies the caller belongs to (or all platform companies if `PLATFORM_ADMIN`).

---

### 3.3 Get Tenant Details
- **Endpoint**: `GET /api/v1/tenants/:tenant_id`
- **Access**: Bearer Token + Tenant Context
- **Description**: Returns tenant details if caller is a member. Emits `403 Forbidden` (`CROSS_TENANT_ACCESS_DENIED`) on cross-tenant attempts.

---

### 3.4 Update Tenant Metadata
- **Endpoint**: `PATCH /api/v1/tenants/:tenant_id`
- **Access**: Bearer Token + Tenant Context (`TENANT_ADMIN` required)
- **Request Body**:
  ```json
  {
    "name": "Apex Regional Delivery Hub",
    "contact_email": "ops@apexdelivery.com"
  }
  ```
- **Response (200 OK)**: Updated tenant entity.

---

### 3.5 List Company Members
- **Endpoint**: `GET /api/v1/tenants/:tenant_id/members`
- **Access**: Bearer Token + Tenant Context (`VIEWER` or above)
- **Response (200 OK)**: List of active member summaries with roles.

---

### 3.6 Invite / Add Member to Company
- **Endpoint**: `POST /api/v1/tenants/:tenant_id/members`
- **Access**: Bearer Token + Tenant Context (`TENANT_ADMIN` required)
- **Request Body**:
  ```json
  {
    "email": "driver01@apexdelivery.com",
    "role": "TENANT_OPERATOR"
  }
  ```
- **Response (201 Created)**: Created `tenant_memberships` record.
