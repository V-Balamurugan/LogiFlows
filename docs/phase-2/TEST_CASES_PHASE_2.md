# LogiFlows — Phase 2 Detailed Test Cases & Execution Matrix

**Document Reference**: `docs/phase-2/TEST_CASES_PHASE_2.md`  
**Execution Date**: 2026-09-22  
**Status**: 100% EXECUTED & PASSED  

---

## Master Test Execution Summary

| Test Category | Suite Count | Test Count | Passed | Failed | Status |
| :--- | :---: | :---: | :---: | :---: | :---: |
| Company / Tenant Lifecycle | 4 | 6 | 6 | 0 | **PASS** |
| Branch Management & Spatial Queries | 5 | 10 | 10 | 0 | **PASS** |
| Multi-Tenant Boundary & Security | 6 | 12 | 12 | 0 | **PASS** |
| Regression (Phase 0 & Phase 1) | 3 | 17 | 17 | 0 | **PASS** |
| Web Frontend Components & Build | 4 | 10 | 10 | 0 | **PASS** |
| Mobile Dart Unit & Model Tests | 2 | 8 | 8 | 0 | **PASS** |
| **Total Test Cases** | **24** | **63** | **63** | **0** | **100% PASS** |

---

## 1. Company / Tenant Test Cases

### TC-P2-CMP-001: Get Current Company Profile
* **Module**: Companies / Tenants
* **Objective**: Verify authenticated user can retrieve their active company profile via `/companies/current` and `/tenants/current`.
* **Request**: `GET /api/v1/companies/current` with Bearer token.
* **Expected Status**: `200 OK`
* **Expected Envelope**: Contains company `id`, `name`, `slug`, `contact_email`, `status = ACTIVE`.
* **Actual Status**: `200 OK`
* **Evidence**: `TestCompany_GetCurrentAndBranchStatusUpdate` (backend/tests/integration/company_branch_test.go:34).
* **Result**: **PASS**

### TC-P2-CMP-002: Update Company Metadata (Authorized Admin)
* **Module**: Companies / Tenants
* **Objective**: Verify `TENANT_ADMIN` can modify company name and email.
* **Request**: `PATCH /api/v1/companies/:id` with `{"name": "Apex Global Logistics", "contact_email": "ops@apex.com"}`.
* **Expected Status**: `200 OK`
* **Expected Body**: Returns updated name and email.
* **Actual Status**: `200 OK`
* **Evidence**: `TestRegression_Phase1_Identity_And_MultiTenancy/RBAC_-_Tenant_Admin_Updates_Tenant_Metadata`.
* **Result**: **PASS**

### TC-P2-CMP-003: Update Company Metadata Rejected for Non-Admin
* **Module**: Companies / Tenants
* **Objective**: Verify `VIEWER` or unauthorized users cannot update company details.
* **Request**: `PATCH /api/v1/companies/:id` as `VIEWER`.
* **Expected Status**: `403 Forbidden`
* **Actual Status**: `403 Forbidden`
* **Evidence**: `TestRegression_Phase1_Identity_And_MultiTenancy/RBAC_-_Non-Member_Forbidden_from_Tenant_Update`.
* **Result**: **PASS**

### TC-P2-CMP-004: Cross-Tenant Company Access Denied
* **Module**: Companies / Tenants
* **Objective**: User belonging to Company A cannot view or modify Company B.
* **Request**: `GET /api/v1/companies/COMPANY_B_UUID` with Company A user token.
* **Expected Status**: `403 Forbidden`
* **Actual Status**: `403 Forbidden`
* **Evidence**: `TestSecurity_CrossTenantAccess_Forbidden`.
* **Result**: **PASS**

---

## 2. Branch Management Test Cases

### TC-P2-BR-001: Create Branch with PostGIS Coordinates
* **Module**: Branches
* **Objective**: Create a distribution hub with valid WGS 84 latitude and longitude.
* **Request**: `POST /api/v1/companies/:company_id/branches`
  ```json
  {
    "branch_code": "CHN001",
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
* **Expected Status**: `201 Created`
* **Expected Body**: `"branch_code": "CHN001"`, `"latitude": 13.0827`, `"longitude": 80.2707`, `"operating_status": "ACTIVE"`.
* **Actual Status**: `201 Created`
* **Evidence**: `TestBranch_Create_WithPostGIS`.
* **Result**: **PASS**

### TC-P2-BR-002: Duplicate Branch Code Rejected within Tenant
* **Module**: Branches
* **Objective**: Prevent duplicate branch codes in the same company.
* **Request**: `POST /api/v1/companies/:company_id/branches` with existing `branch_code: "CHN001"`.
* **Expected Status**: `409 Conflict`
* **Actual Status**: `409 Conflict`
* **Evidence**: `TestBranch_DuplicateCode_RejectedWithinTenant`.
* **Result**: **PASS**

### TC-P2-BR-003: Identical Branch Code Allowed in Different Tenants
* **Module**: Branches
* **Objective**: Verify branch code uniqueness is scoped per tenant, allowing different companies to use `MAIN-01`.
* **Request**: Create `branch_code: "MAIN-01"` in Tenant A, then in Tenant B.
* **Expected Status**: `201 Created` in both tenants.
* **Actual Status**: `201 Created` in both.
* **Evidence**: `TestBranch_IdenticalCode_AllowedInDifferentTenants`.
* **Result**: **PASS**

### TC-P2-BR-004: Coordinate Bounds Validation (-90 to +90 lat, -180 to +180 lng)
* **Module**: Branches
* **Objective**: Reject out-of-range coordinates.
* **Request**: `POST /api/v1/companies/:company_id/branches` with `latitude: 95.5`.
* **Expected Status**: `400 Bad Request`
* **Actual Status**: `400 Bad Request`
* **Evidence**: `TestBranch_Create_WithPostGIS`.
* **Result**: **PASS**

### TC-P2-BR-005: Spatial Proximity Radius Filtering
* **Module**: Branches
* **Objective**: Filter branches using PostGIS `ST_DWithin` and calculate geodesic distances.
* **Request**: `GET /api/v1/companies/:company_id/branches?near_lat=28.70&near_lng=77.10&radius_km=45`
* **Expected Status**: `200 OK`
* **Expected Body**: Returns matching branches with computed `distance_km` sorted nearest-first.
* **Actual Status**: `200 OK`
* **Evidence**: `TestBranch_List_And_SpatialFilter`.
* **Result**: **PASS**

### TC-P2-BR-006: Update Branch Operating Status
* **Module**: Branches
* **Objective**: Transition branch operating status (`ACTIVE` -> `INACTIVE`).
* **Request**: `PATCH /api/v1/companies/:company_id/branches/:branch_id/status` with `{"operating_status": "INACTIVE"}`.
* **Expected Status**: `200 OK`
* **Expected Body**: `"operating_status": "INACTIVE"`, `"is_active": false`.
* **Actual Status**: `200 OK`
* **Evidence**: `TestCompany_GetCurrentAndBranchStatusUpdate`.
* **Result**: **PASS**

### TC-P2-BR-007: Invalid Operating Status Rejected
* **Module**: Branches
* **Objective**: Reject unapproved operating status values (e.g. `DESTROYED`).
* **Request**: `PATCH /api/v1/companies/:company_id/branches/:branch_id/status` with `{"operating_status": "DESTROYED"}`.
* **Expected Status**: `400 Bad Request`
* **Actual Status**: `400 Bad Request`
* **Evidence**: `TestCompany_GetCurrentAndBranchStatusUpdate`.
* **Result**: **PASS**

### TC-P2-BR-008: Cross-Tenant Branch Status Update Forbidden
* **Module**: Branches
* **Objective**: Ensure a user in Company B cannot alter the operating status of Company A's branch.
* **Request**: `PATCH /api/v1/companies/COMPANY_B/branches/COMPANY_A_BRANCH/status`.
* **Expected Status**: `403 Forbidden`
* **Actual Status**: `403 Forbidden`
* **Evidence**: `TestCompany_GetCurrentAndBranchStatusUpdate`.
* **Result**: **PASS**

---

## 3. Security & Isolation Test Cases

### TC-P2-SEC-001: Cross-Tenant IDOR Attack Rejection
* **Objective**: Verify that changing the UUID in the URL to a foreign tenant returns `403 Forbidden`.
* **Evidence**: `TestSecurity_CrossTenantAccess_Forbidden`.
* **Result**: **PASS**

### TC-P2-SEC-002: Algorithm None JWT Attack Rejection
* **Objective**: Verify that token tampering with `"alg": "none"` is rejected with `401 Unauthorized`.
* **Evidence**: `TestSecurity_JWT_TamperingAndAttackVectors/Algorithm_Confusion_Attack_-_alg_none`.
* **Result**: **PASS**

### TC-P2-SEC-003: Forged Key Signature Rejection
* **Objective**: Tokens signed with an invalid private key return `401 Unauthorized`.
* **Evidence**: `TestSecurity_JWT_TamperingAndAttackVectors/Forged_Signature_-_Signed_with_Wrong_Key`.
* **Result**: **PASS**

### TC-P2-SEC-004: Refresh Token Reuse & Breach Invalidation
* **Objective**: Replaying an already-rotated refresh token revokes the entire token family and forces re-authentication.
* **Evidence**: `TestAuthAPI_TokenRefresh_Rotation_And_BreachDetection`.
* **Result**: **PASS**

### TC-P2-SEC-005: SQL Injection Fuzzing Resistance
* **Objective**: Test injection payloads (`' OR 1=1 --`, `' UNION SELECT ...`) across search, branch names, and filters.
* **Evidence**: `TestSecurity_Registration_InjectionResistance`.
* **Result**: **PASS**

---

## 4. Frontend & Mobile Test Cases

### TC-P2-FE-001: Frontend Auth Token Storage Tests
* **Runner**: Node.js test runner (`npm test -- --run`).
* **Assertions**: 4 tests asserting saving, retrieval, and purge of access/refresh tokens.
* **Result**: **PASS (4/4)**

### TC-P2-FE-002: Form Validation Rules
* **Runner**: Node.js test runner.
* **Assertions**: 3 tests enforcing RFC email standards and password complexity.
* **Result**: **PASS (3/3)**

### TC-P2-FE-003: API Envelope Parsing
* **Runner**: Node.js test runner.
* **Assertions**: 2 tests unpacking success and error envelopes.
* **Result**: **PASS (2/2)**

### TC-P2-FE-004: Production Build Compilation
* **Runner**: `npm run build`.
* **Outcome**: Transformed 1887 modules in 1.16s generating optimized `dist/`.
* **Result**: **PASS (0 errors)**

### TC-P2-MOB-001: Mobile Company Model Serialization
* **Runner**: Dart test suite (`mobile/test/company_models_test.dart`).
* **Assertions**: Validates deserialization of company JSON and date parsing.
* **Result**: **PASS**

### TC-P2-MOB-002: Mobile Branch Model Serialization
* **Runner**: Dart test suite (`mobile/test/resource_models_test.dart`).
* **Assertions**: Validates branch model coordinates and operating status.
* **Result**: **PASS**
