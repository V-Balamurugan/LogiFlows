# LogiFlows — Phase 2 API Contracts & Specification

**Document Reference**: `docs/phase-2/API_CONTRACTS_PHASE_2.md`  
**Phase**: Phase 2 — Multi-Tenant Companies & Distribution Branches  
**Status**: APPROVED & IMPLEMENTED  
**Base URL**: `http://localhost:8080/api/v1` (or production gateway)  
**Security**: Bearer JWT (`Authorization: Bearer <access_token>`)  

---

## Standard API Envelopes

Every endpoint in the LogiFlows REST API complies with the uniform response envelope standard.

### Success Envelope (HTTP 200, 201)
```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": {}
}
```

### Error Envelope (HTTP 400, 401, 403, 404, 409, 500)
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable description of error",
    "details": null
  },
  "request_id": "c1f729da-6acc-48f3-a739-3e05fa8ee681"
}
```

### Standard HTTP Status Codes
* `200 OK`: Resource retrieved or modified successfully.
* `201 Created`: Resource successfully created and persisted.
* `400 Bad Request`: Input validation failed (e.g., invalid coordinates, blank name).
* `401 Unauthorized`: Missing, invalid, or expired JWT access token.
* `403 Forbidden`: Authenticated user lacks permission or attempts cross-tenant access.
* `404 Not Found`: Resource does not exist within the authorized tenant scope.
* `409 Conflict`: Unique constraint violation (e.g., duplicate branch code in tenant).
* `500 Internal Server Error`: Unexpected unhandled server exception.

---

## 1. Company / Tenant Management Endpoints

The API supports both `/companies` and `/tenants` route prefixes interchangeably.

### 1.1 Get Current Company
* **API NAME**: Get Current Company
* **METHOD**: `GET`
* **PATH**: `/api/v1/companies/current` (or `/api/v1/tenants/current`)
* **PURPOSE**: Retrieves the authenticated user's primary company/organization profile.
* **AUTHENTICATION**: Bearer JWT Access Token
* **REQUIRED ROLE**: Any active member (`TENANT_ADMIN`, `TENANT_OPERATOR`, `TENANT_VIEWER`)
* **REQUEST HEADERS**:
  * `Authorization: Bearer <access_token>`
* **CURL EXAMPLE**:
  ```bash
  curl -X GET http://localhost:8080/api/v1/companies/current \
    -H "Authorization: Bearer $ACCESS_TOKEN"
  ```
* **SUCCESS RESPONSE (HTTP 200 OK)**:
  ```json
  {
    "success": true,
    "message": "Current company retrieved successfully",
    "data": {
      "id": "dc774b3b-951a-45d3-9688-a9b82a6c21fe",
      "name": "Apex Logistics Global",
      "slug": "apex-logistics",
      "contact_email": "ops@apexlogistics.com",
      "status": "ACTIVE",
      "created_at": "2026-09-22T10:00:00Z",
      "updated_at": "2026-09-22T10:00:00Z"
    }
  }
  ```

---

### 1.2 Get Company by ID
* **API NAME**: Get Company Details
* **METHOD**: `GET`
* **PATH**: `/api/v1/companies/:company_id` (or `/api/v1/tenants/:tenant_id`)
* **PURPOSE**: Retrieves specific company metadata.
* **AUTHENTICATION**: Bearer JWT Access Token
* **REQUIRED ROLE**: Active tenant member or `PLATFORM_ADMIN`
* **PATH PARAMETERS**:
  * `company_id` (string, UUID, required): Target company UUID
* **CURL EXAMPLE**:
  ```bash
  curl -X GET http://localhost:8080/api/v1/companies/dc774b3b-951a-45d3-9688-a9b82a6c21fe \
    -H "Authorization: Bearer $ACCESS_TOKEN"
  ```
* **SUCCESS RESPONSE (HTTP 200 OK)**:
  ```json
  {
    "success": true,
    "message": "Company retrieved successfully",
    "data": {
      "id": "dc774b3b-951a-45d3-9688-a9b82a6c21fe",
      "name": "Apex Logistics Global",
      "slug": "apex-logistics",
      "contact_email": "ops@apexlogistics.com",
      "status": "ACTIVE",
      "created_at": "2026-09-22T10:00:00Z",
      "updated_at": "2026-09-22T10:00:00Z"
    }
  }
  ```
* **FORBIDDEN RESPONSE (HTTP 403 Forbidden)**:
  ```json
  {
    "success": false,
    "error": {
      "code": "FORBIDDEN",
      "message": "User does not have access to this tenant"
    },
    "request_id": "88ff5bb4-ddc7-421b-8c69-ac68889a79f5"
  }
  ```

---

### 1.3 Update Company Details
* **API NAME**: Update Company Metadata
* **METHOD**: `PATCH`
* **PATH**: `/api/v1/companies/:company_id` (or `/api/v1/tenants/:tenant_id`)
* **PURPOSE**: Updates company name and primary contact email.
* **AUTHENTICATION**: Bearer JWT Access Token
* **REQUIRED ROLE**: `TENANT_ADMIN`, `PLATFORM_ADMIN`
* **REQUEST BODY**:
  ```json
  {
    "name": "Apex Logistics International",
    "contact_email": "support@apexlogistics.com"
  }
  ```
* **SUCCESS RESPONSE (HTTP 200 OK)**:
  ```json
  {
    "success": true,
    "message": "Company updated successfully",
    "data": {
      "id": "dc774b3b-951a-45d3-9688-a9b82a6c21fe",
      "name": "Apex Logistics International",
      "slug": "apex-logistics",
      "contact_email": "support@apexlogistics.com",
      "status": "ACTIVE",
      "created_at": "2026-09-22T10:00:00Z",
      "updated_at": "2026-09-22T10:15:00Z"
    }
  }
  ```

---

## 2. Branch Management Endpoints

Branches represent distribution centers, sorting facilities, and delivery hubs.

### 2.1 Create Branch
* **API NAME**: Create Distribution Branch
* **METHOD**: `POST`
* **PATH**: `/api/v1/companies/:company_id/branches` (or `/api/v1/tenants/:tenant_id/branches`)
* **PURPOSE**: Registers a physical logistics hub with PostGIS coordinates.
* **AUTHENTICATION**: Bearer JWT Access Token
* **REQUIRED ROLE**: `TENANT_ADMIN`, `TENANT_OPERATOR`, `PLATFORM_ADMIN`
* **REQUEST BODY**:
  ```json
  {
    "branch_code": "CHN001",
    "name": "Chennai Central Hub",
    "address_line1": "12 Mount Road, Anna Salai",
    "address_line2": "Building B",
    "city": "Chennai",
    "state": "Tamil Nadu",
    "postal_code": "600001",
    "country": "India",
    "latitude": 13.0827,
    "longitude": 80.2707,
    "operating_status": "ACTIVE"
  }
  ```
* **SUCCESS RESPONSE (HTTP 201 Created)**:
  ```json
  {
    "success": true,
    "message": "Branch created successfully",
    "data": {
      "id": "31197d7e-e7ca-4b53-b9ab-0d0cb6a0113b",
      "tenant_id": "dc774b3b-951a-45d3-9688-a9b82a6c21fe",
      "branch_code": "CHN001",
      "name": "Chennai Central Hub",
      "address_line1": "12 Mount Road, Anna Salai",
      "address_line2": "Building B",
      "city": "Chennai",
      "state": "Tamil Nadu",
      "postal_code": "600001",
      "country": "India",
      "latitude": 13.0827,
      "longitude": 80.2707,
      "operating_status": "ACTIVE",
      "is_active": true,
      "created_at": "2026-09-22T10:20:00Z",
      "updated_at": "2026-09-22T10:20:00Z"
    }
  }
  ```
* **CONFLICT RESPONSE (HTTP 409 Conflict)**:
  ```json
  {
    "success": false,
    "error": {
      "code": "DUPLICATE_BRANCH_CODE",
      "message": "A branch with this code already exists for this company"
    },
    "request_id": "50181913-ca83-43da-b5fb-e57e558f5adb"
  }
  ```

---

### 2.2 List Branches (with Pagination, Search & Spatial Filter)
* **API NAME**: List Company Branches
* **METHOD**: `GET`
* **PATH**: `/api/v1/companies/:company_id/branches`
* **PURPOSE**: Lists company branches with optional pagination, text search, and PostGIS radius search.
* **AUTHENTICATION**: Bearer JWT Access Token
* **QUERY PARAMETERS**:
  * `page` (integer, optional, default: 1): Page index
  * `limit` (integer, optional, default: 20): Items per page
  * `search` (string, optional): Search by name, branch code, or city
  * `status` (string, optional): Filter by `ACTIVE`, `INACTIVE`, `SUSPENDED`
  * `near_lat` (float, optional): Proximity latitude in decimal degrees
  * `near_lng` (float, optional): Proximity longitude in decimal degrees
  * `radius_km` (float, optional, default: 50.0): Spatial search radius in kilometers
* **CURL EXAMPLE**:
  ```bash
  curl -X GET "http://localhost:8080/api/v1/companies/$COMPANY_ID/branches?near_lat=13.08&near_lng=80.27&radius_km=25" \
    -H "Authorization: Bearer $ACCESS_TOKEN"
  ```
* **SUCCESS RESPONSE (HTTP 200 OK)**:
  ```json
  {
    "success": true,
    "message": "Branches retrieved successfully",
    "data": {
      "branches": [
        {
          "id": "31197d7e-e7ca-4b53-b9ab-0d0cb6a0113b",
          "tenant_id": "dc774b3b-951a-45d3-9688-a9b82a6c21fe",
          "branch_code": "CHN001",
          "name": "Chennai Central Hub",
          "city": "Chennai",
          "state": "Tamil Nadu",
          "latitude": 13.0827,
          "longitude": 80.2707,
          "operating_status": "ACTIVE",
          "is_active": true,
          "distance_km": 0.45
        }
      ],
      "total": 1,
      "page": 1,
      "limit": 20
    }
  }
  ```

---

### 2.3 Get Branch by ID
* **API NAME**: Get Branch Details
* **METHOD**: `GET`
* **PATH**: `/api/v1/companies/:company_id/branches/:branch_id`
* **AUTHENTICATION**: Bearer JWT Access Token
* **SUCCESS RESPONSE (HTTP 200 OK)**:
  ```json
  {
    "success": true,
    "message": "Branch retrieved successfully",
    "data": {
      "id": "31197d7e-e7ca-4b53-b9ab-0d0cb6a0113b",
      "tenant_id": "dc774b3b-951a-45d3-9688-a9b82a6c21fe",
      "branch_code": "CHN001",
      "name": "Chennai Central Hub",
      "address_line1": "12 Mount Road, Anna Salai",
      "address_line2": "Building B",
      "city": "Chennai",
      "state": "Tamil Nadu",
      "postal_code": "600001",
      "country": "India",
      "latitude": 13.0827,
      "longitude": 80.2707,
      "operating_status": "ACTIVE",
      "is_active": true,
      "created_at": "2026-09-22T10:20:00Z",
      "updated_at": "2026-09-22T10:20:00Z"
    }
  }
  ```

---

### 2.4 Update Branch Status
* **API NAME**: Update Branch Operating Status
* **METHOD**: `PATCH`
* **PATH**: `/api/v1/companies/:company_id/branches/:branch_id/status` (or `/api/v1/tenants/:tenant_id/branches/:branch_id/status`)
* **PURPOSE**: Changes the operating status (`ACTIVE`, `INACTIVE`, `SUSPENDED`) of a branch.
* **AUTHENTICATION**: Bearer JWT Access Token
* **REQUIRED ROLE**: `TENANT_ADMIN`, `TENANT_OPERATOR`, `PLATFORM_ADMIN`
* **REQUEST BODY**:
  ```json
  {
    "operating_status": "INACTIVE"
  }
  ```
* **SUCCESS RESPONSE (HTTP 200 OK)**:
  ```json
  {
    "success": true,
    "message": "Branch operating status updated successfully",
    "data": {
      "id": "31197d7e-e7ca-4b53-b9ab-0d0cb6a0113b",
      "tenant_id": "dc774b3b-951a-45d3-9688-a9b82a6c21fe",
      "branch_code": "CHN001",
      "name": "Chennai Central Hub",
      "operating_status": "INACTIVE",
      "is_active": false,
      "updated_at": "2026-09-22T10:35:00Z"
    }
  }
  ```
* **VALIDATION ERROR (HTTP 400 Bad Request)**:
  ```json
  {
    "success": false,
    "error": {
      "code": "INVALID_STATUS",
      "message": "operating_status must be one of: ACTIVE, INACTIVE, SUSPENDED"
    },
    "request_id": "8c751ad9-73a6-4edf-b19f-ea1168d9fc53"
  }
  ```

---

### 2.5 Delete / Deactivate Branch
* **API NAME**: Deactivate Branch
* **METHOD**: `DELETE`
* **PATH**: `/api/v1/companies/:company_id/branches/:branch_id`
* **PURPOSE**: Performs a safe soft-delete of a distribution branch.
* **AUTHENTICATION**: Bearer JWT Access Token
* **REQUIRED ROLE**: `TENANT_ADMIN`, `PLATFORM_ADMIN`
* **SUCCESS RESPONSE (HTTP 200 OK)**:
  ```json
  {
    "success": true,
    "message": "Branch deactivated successfully",
    "data": null
  }
  ```

---

## 3. Swagger / OpenAPI Synchronization

All endpoints are registered and documented in `backend/docs/swagger.json` with OpenAPI specification version 1.0. The interactive Swagger UI is available at `/swagger/index.html`.
