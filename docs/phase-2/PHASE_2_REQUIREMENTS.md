# LogiFlows — Phase 2 Requirements & Acceptance Criteria

**Document Reference**: `docs/phase-2/PHASE_2_REQUIREMENTS.md`  
**Execution Date**: 2026-09-22  
**Baseline**: LogiFlows SDLC Specification & Phase 2 Official Scope  
**Status**: APPROVED & VERIFIED  

---

## 1. Scope Definition & Objectives

Phase 2 of LogiFlows delivers an enterprise-grade, multi-tenant vertical slice for **Company/Tenant Management** and **Distribution Branch Operations**. The system guarantees strict cryptographic and logical tenant isolation, spatial GIS distribution, role-based authorization, and real-time synchronized status transitions across Go backend, PostgreSQL/PostGIS database, React web console, and Flutter mobile client.

### Explicit Phase 2 Scope Boundaries
* **In Scope**:
  1. Company/Tenant management lifecycle (`GET /companies/current`, `GET /companies/:id`, `PATCH /companies/:id`).
  2. Branch management lifecycle (`POST`, `GET`, `PATCH`, `DELETE` on `/companies/:id/branches`).
  3. Company-to-branch parent ownership and referential integrity.
  4. Cryptographic and server-enforced multi-tenant isolation.
  5. PostGIS geospatial point indexing (`GEOMETRY(Point, 4326)`), coordinate bounding, and radius search.
  6. Branch operational status transitions (`ACTIVE`, `INACTIVE`, `SUSPENDED`).
  7. Web console administration (`CompanyProfile.tsx`, `BranchList.tsx`).
  8. Mobile Flutter field access (`CompanyScreen`, `BranchScreen`, `ResourceApiClient`).
  9. OpenAPI 3.0 / Swagger synchronization.
  10. 100% automated test coverage (unit, integration, regression, and security).
* **Deferred to Later Phases**:
  * Live GPS telemetry and continuous driver location streaming (Phase 4).
  * QR custody scans, barcode parcel tracking, and proof-of-delivery (Phase 5).
  * Machine learning delay prediction integration into production routing (Phase 6).

---

## 2. User Stories & Acceptance Criteria

### User Story 1: Platform & Tenant Lifecycle Management
* **Role**: `PLATFORM_ADMIN`, `TENANT_ADMIN`
* **Preconditions**: User is authenticated with a valid JWT access token and holds `TENANT_ADMIN` membership in the targeted organization.
* **Request**: 
  * `GET /api/v1/companies/current` or `GET /api/v1/tenants/current`
  * `GET /api/v1/companies/{company_id}` or `GET /api/v1/tenants/{tenant_id}`
* **Expected Response**: `200 OK` with JSON envelope containing `id`, `name`, `slug`, `contact_email`, `status`, `created_at`, and `updated_at`.
* **Validation Rules**:
  * `tenant_id` / `company_id` must be a valid RFC 4122 UUID.
* **Authorization Rules**: User must be an active member of the requested company/tenant.
* **Negative Test Cases**:
  * Request without `Authorization` header -> `401 Unauthorized`.
  * Request with expired JWT -> `401 Unauthorized`.
  * User requests a company UUID they do not belong to -> `403 Forbidden`.
* **Completion Evidence**: `TestCompany_GetCurrentAndBranchStatusUpdate`, `TestSecurity_CrossTenantAccess_Forbidden`.

---

### User Story 2: Company Profile & Metadata Update
* **Role**: `TENANT_ADMIN`
* **Preconditions**: Authenticated user with `TENANT_ADMIN` role in targeted tenant.
* **Request**: `PATCH /api/v1/companies/{company_id}` or `PATCH /api/v1/tenants/{tenant_id}`
  ```json
  {
    "name": "Apex Logistics Global",
    "contact_email": "ops@apexlogistics.com"
  }
  ```
* **Expected Response**: `200 OK` with updated company details.
* **Validation Rules**:
  * `name` must not be blank (min 2, max 255 chars).
  * `contact_email` if provided must be a valid RFC 5322 email string.
* **Authorization Rules**: `VIEWER` or `OPERATOR` roles cannot mutate company profile.
* **Negative Test Cases**:
  * Empty company name -> `400 Bad Request`.
  * Attempt by `VIEWER` -> `403 Forbidden`.
  * Cross-tenant update -> `403 Forbidden`.
* **Completion Evidence**: `TestRegression_Phase1_Identity_And_MultiTenancy/RBAC_-_Tenant_Admin_Updates_Tenant_Metadata`.

---

### User Story 3: Distribution Branch Creation
* **Role**: `TENANT_ADMIN`, `TENANT_OPERATOR`
* **Preconditions**: Authenticated user in company with administrative or operational role.
* **Request**: `POST /api/v1/companies/{company_id}/branches`
  ```json
  {
    "branch_code": "CHN-01",
    "name": "Chennai Central Hub",
    "address_line1": "12 Mount Road",
    "city": "Chennai",
    "state": "Tamil Nadu",
    "postal_code": "600001",
    "latitude": 13.0827,
    "longitude": 80.2707,
    "operating_status": "ACTIVE"
  }
  ```
* **Expected Response**: `201 Created` with branch record, PostGIS point location, and active status.
* **Validation Rules**:
  * `branch_code`: Alphanumeric, max 50 chars, unique within tenant.
  * `latitude`: -90.0 to +90.0.
  * `longitude`: -180.0 to +180.0.
  * `name`: min 2, max 255 chars.
* **Authorization Rules**: `VIEWER` role is rejected with `403 Forbidden`.
* **Negative Test Cases**:
  * Duplicate `branch_code` in same tenant -> `409 Conflict`.
  * Same `branch_code` in a *different* tenant -> `201 Created` (multi-tenant code independence).
  * Out-of-bounds latitude (e.g., 95.0) -> `400 Bad Request`.
* **Completion Evidence**: `TestBranch_Create_WithPostGIS`, `TestBranch_DuplicateCode_RejectedWithinTenant`, `TestBranch_IdenticalCode_AllowedInDifferentTenants`.

---

### User Story 4: Branch Querying, Search, Filtering & Spatial Proximity
* **Role**: Any authorized tenant member (`TENANT_ADMIN`, `TENANT_OPERATOR`, `TENANT_VIEWER`).
* **Preconditions**: User belongs to company.
* **Request**: 
  * `GET /api/v1/companies/{company_id}/branches?page=1&limit=20`
  * Spatial query: `GET /api/v1/companies/{company_id}/branches?near_lat=13.08&near_lng=80.27&radius_km=25`
* **Expected Response**: `200 OK` with pagination envelope:
  ```json
  {
    "status": "ok",
    "data": {
      "branches": [...],
      "total": 5,
      "page": 1,
      "limit": 20
    }
  }
  ```
* **Validation Rules**: Positive integers for pagination, valid coordinate floats for spatial queries.
* **Authorization Rules**: Scoped strictly to user's tenant ID.
* **Negative Test Cases**: Negative radius or non-numeric coordinates -> `400 Bad Request`.
* **Completion Evidence**: `TestBranch_List_And_SpatialFilter`.

---

### User Story 5: Branch Metadata & Operating Status Transition
* **Role**: `TENANT_ADMIN`, `TENANT_OPERATOR`
* **Preconditions**: Branch exists and belongs to user's company.
* **Request**: `PATCH /api/v1/companies/{company_id}/branches/{branch_id}/status`
  ```json
  {
    "operating_status": "INACTIVE"
  }
  ```
* **Expected Response**: `200 OK` with updated branch status and audit event logged.
* **Validation Rules**: Status must be one of `ACTIVE`, `INACTIVE`, `SUSPENDED`.
* **Authorization Rules**: Cannot change status of branch belonging to another company.
* **Negative Test Cases**: Invalid status string `DESTROYED` -> `400 Bad Request`.
* **Completion Evidence**: `TestCompany_GetCurrentAndBranchStatusUpdate`.

---

### User Story 6: Detailed Branch Inspection with PostGIS Coordinates
* **Role**: Any authorized tenant member.
* **Preconditions**: Branch belongs to tenant.
* **Request**: `GET /api/v1/companies/{company_id}/branches/{branch_id}`
* **Expected Response**: `200 OK` with full address, coordinates (`latitude`, `longitude`), and operating status.
* **Negative Test Cases**:
  * Requesting another tenant's branch ID -> `403 Forbidden`.
  * Non-existent UUID -> `404 Not Found`.
* **Completion Evidence**: `TestBranch_CrossTenantAccess_Forbidden`.

---

### User Story 7: Strict Tenant Cryptographic & Logical Isolation
* **Role**: Any authenticated user.
* **Preconditions**: User possesses valid JWT for Tenant A.
* **Request**: Manipulate path parameter or body to access Tenant B's data (`/companies/TENANT_B_UUID/...`).
* **Expected Response**: `403 Forbidden` with standardized error envelope.
* **Authorization Rules**: Database queries unconditionally filter on `tenant_id = $1` derived from verified JWT claims.
* **Completion Evidence**: `TestSecurity_CrossTenantAccess_Forbidden`, `TestRegression_Phase1_Identity_And_MultiTenancy`.

---

### User Story 8: Branch Scope & Ownership Integrity
* **Role**: Any authenticated user.
* **Preconditions**: Two active companies exist in database.
* **Request**: User submits a branch ID belonging to Company B inside a request URL targeting Company A.
* **Expected Response**: `404 Not Found` or `403 Forbidden` (no data leakage).
* **Completion Evidence**: `TestBranch_CrossTenantAccess_Forbidden`.

---

### User Story 9: React Web Console Integration
* **Role**: Web portal administrator.
* **Preconditions**: Go backend running on port 8080.
* **Request**: Navigate to Company Profile and Branch Management tabs in React console.
* **Expected Response**:
  * Real-time tenant profile rendering with member counts, active status badge, and update form.
  * Branch list with search, PostGIS coordinate pills, status badges, and creation modal.
  * Toast notifications on success and error mapping.
* **Completion Evidence**: Frontend production build (`dist/`), 9 component unit tests passing.

---

### User Story 10: Flutter Mobile Resource Client
* **Role**: Mobile driver or field manager.
* **Preconditions**: Authenticated mobile session with JWT stored securely.
* **Request**: Access Company and Hubs screens in mobile client.
* **Expected Response**: Dark Material 3 interface showing company metadata and distribution branch list with offline resilience.
* **Completion Evidence**: `CompanyModel`, `BranchModel`, `ResourceApiClient`, `company_models_test.dart`.
