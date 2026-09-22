# LogiFlows — Phase 1 Test Case Catalog (Identity & Multi-Tenancy)

**Document Reference**: `docs/testing/TEST_CASES_PHASE_1.md`  
* **Phase**: Phase 1 — Identity, Authentication, Authorization, and Multi-Tenancy  
* **Status**: **COMPLETE / 100% AUTOMATED & VERIFIED PASSING**  
* **Audit Verification Date**: September 22, 2026  

---

### TC-P1-AUT-001: Register New User with Valid Information
* **Phase**: Phase 1 | **Module**: Auth | **Feature**: User Registration
* **Test Type**: API / Integration | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: Email does not exist in database.
* **Test Data**: `{"full_name":"Apex Owner","email":"owner-<uuid>@logiflows.test","password":"Password123!","company_name":"Apex Logistics"}`
* **Test Steps**: Dispatch `POST /api/v1/auth/register`.
* **Expected Result**: HTTP 201 Created; response returns `user` (without password hash), `tenant`, and JWT token pair (`access_token`, `refresh_token`).
* **Actual Result**: HTTP 201 Created; user and tenant provisioned atomically.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/tests/integration/auth_api_integration_test.go:TestAuthAPI_Register_Login_Me_Lifecycle`
* **Evidence**: Duration: 1.81s, Status: 201.

---

### TC-P1-AUT-002: Reject Duplicate Email Registration
* **Phase**: Phase 1 | **Module**: Auth | **Feature**: User Registration
* **Test Type**: API / Security | **Priority**: P0 | **Severity**: Major | **Automated**: Automated
* **Preconditions**: User with email already exists.
* **Test Data**: Same email as existing user.
* **Test Steps**: Dispatch `POST /api/v1/auth/register` with duplicate email.
* **Expected Result**: HTTP 409 Conflict; error code `EMAIL_ALREADY_EXISTS`.
* **Actual Result**: HTTP 409 Conflict returned.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/tests/integration/auth_api_integration_test.go`
* **Evidence**: Confirmed 409 Conflict.

---

### TC-P1-AUT-003: Reject Registration with Password < 8 Characters
* **Phase**: Phase 1 | **Module**: Auth | **Feature**: Password Validation
* **Test Type**: API / Security | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: None.
* **Test Data**: `password: "Short1!"` (7 chars).
* **Test Steps**: Dispatch `POST /api/v1/auth/register` with short password.
* **Expected Result**: HTTP 400 Bad Request; validation message indicates password must be at least 8 characters.
* **Actual Result**: HTTP 400 Bad Request returned.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/internal/auth/password_test.go`
* **Evidence**: Unit & API tests verify rejection.

---

### TC-P1-AUT-004: Reject Registration with Invalid Email Format
* **Phase**: Phase 1 | **Module**: Auth | **Feature**: Input Validation
* **Test Type**: API / Security | **Priority**: P1 | **Severity**: Major | **Automated**: Automated
* **Preconditions**: None.
* **Test Data**: `email: "not-an-email"`.
* **Test Steps**: Dispatch `POST /api/v1/auth/register` with malformed email.
* **Expected Result**: HTTP 400 Bad Request; validation error on email.
* **Actual Result**: HTTP 400 Bad Request returned.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/tests/integration/security_comprehensive_test.go`
* **Evidence**: Handled via Gin binding validation.

---

### TC-P1-AUT-005: Reject Registration with Empty Full Name
* **Phase**: Phase 1 | **Module**: Auth | **Feature**: Input Validation
* **Test Type**: API | **Priority**: P1 | **Severity**: Moderate | **Automated**: Automated
* **Preconditions**: None.
* **Test Data**: `full_name: ""`.
* **Test Steps**: Dispatch `POST /api/v1/auth/register` with empty name.
* **Expected Result**: HTTP 400 Bad Request; validation error.
* **Actual Result**: HTTP 400 Bad Request returned.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/tests/integration/security_comprehensive_test.go`
* **Evidence**: Automated in comprehensive security test suite.

---

### TC-P1-AUT-006: Email Case Normalization on Registration and Login
* **Phase**: Phase 1 | **Module**: Auth | **Feature**: Identity Normalization
* **Test Type**: API / Security | **Priority**: P1 | **Severity**: Major | **Automated**: Automated
* **Preconditions**: None.
* **Test Data**: Registered as `User.Alpha@LogiFlows.Test`, login as `user.alpha@logiflows.test`.
* **Test Steps**: Register with mixed-case email, then login with lowercase email.
* **Expected Result**: Login succeeds; database stores normalized lowercase email or query uses `LOWER()`.
* **Actual Result**: Login succeeds seamlessly; `idx_users_email_lower` matches.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/tests/integration/security_comprehensive_test.go`
* **Evidence**: Automated in comprehensive security test suite.

---

### TC-P1-AUT-007: Password Hashing with Bcrypt Cost 12
* **Phase**: Phase 1 | **Module**: Auth | **Feature**: Cryptographic Security
* **Test Type**: Unit / Security | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: None.
* **Test Data**: `password: "SecurePassword#2026"`.
* **Test Steps**: Call `auth.HashPassword()`, verify cost factor via `bcrypt.Cost()`.
* **Expected Result**: Returned hash is valid bcrypt hash with cost factor exactly 12.
* **Actual Result**: Bcrypt cost factor verified as 12; plaintext never matches hash.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/internal/auth/password_test.go`
* **Evidence**: Unit test PASS.

---

### TC-P1-AUT-008: Password Hash Strict Exclusion from JSON Model
* **Phase**: Phase 1 | **Module**: Users | **Feature**: Data Concealment
* **Test Type**: Unit / Security | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: User object populated with active `password_hash`.
* **Test Data**: `User{PasswordHash: "$2a$12$..."}`.
* **Test Steps**: Marshal user to JSON via `json.Marshal(user)`.
* **Expected Result**: Resulting JSON does NOT contain `"password_hash"` key or substring.
* **Actual Result**: Key strictly omitted due to `json:"-"` struct tag.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/internal/users/model_test.go:TestUser_PasswordHash_ExcludedFromJSON`
* **Evidence**: Unit test PASS (0.01s).

---

### TC-P1-AUT-009: Login with Valid Credentials
* **Phase**: Phase 1 | **Module**: Auth | **Feature**: Authentication
* **Test Type**: API | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: Registered user exists.
* **Test Data**: Valid email and password.
* **Test Steps**: Dispatch `POST /api/v1/auth/login`.
* **Expected Result**: HTTP 200 OK; returns valid JWT access token and refresh token.
* **Actual Result**: HTTP 200 OK; token pair issued.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/tests/integration/auth_api_integration_test.go`
* **Evidence**: Integration test PASS.

---

### TC-P1-AUT-010: Reject Login with Incorrect Password
* **Phase**: Phase 1 | **Module**: Auth | **Feature**: Authentication Security
* **Test Type**: API / Security | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: Registered user exists.
* **Test Data**: Valid email, incorrect password: `"WrongPassword999!"`.
* **Test Steps**: Dispatch `POST /api/v1/auth/login`.
* **Expected Result**: HTTP 401 Unauthorized; error code `INVALID_CREDENTIALS`.
* **Actual Result**: HTTP 401 Unauthorized returned.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/tests/integration/security_comprehensive_test.go`
* **Evidence**: Automated in comprehensive security test suite.

---

### TC-P1-AUT-011: Reject Login with Non-Existent Email
* **Phase**: Phase 1 | **Module**: Auth | **Feature**: Authentication Security
* **Test Type**: API / Security | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: Email does not exist in database.
* **Test Data**: `email: "nobody-here@logiflows.test"`.
* **Test Steps**: Dispatch `POST /api/v1/auth/login`.
* **Expected Result**: HTTP 401 Unauthorized; error code `INVALID_CREDENTIALS` (no user enumeration).
* **Actual Result**: HTTP 401 Unauthorized returned with uniform error envelope.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/tests/integration/security_comprehensive_test.go`
* **Evidence**: Automated in comprehensive security test suite.

---

### TC-P1-AUT-012: Retrieve Current User Profile (`/auth/me`)
* **Phase**: Phase 1 | **Module**: Auth | **Feature**: Profile API
* **Test Type**: API | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: Valid Bearer access token.
* **Test Data**: `Authorization: Bearer <valid_token>`.
* **Test Steps**: Dispatch `GET /api/v1/auth/me`.
* **Expected Result**: HTTP 200 OK; returns user profile and list of associated tenant memberships.
* **Actual Result**: HTTP 200 OK; user ID and memberships returned correctly.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/tests/integration/auth_api_integration_test.go`
* **Evidence**: Integration test PASS.

---

### TC-P1-AUT-013: Reject Expired Access Token
* **Phase**: Phase 1 | **Module**: Auth | **Feature**: Token Validation
* **Test Type**: Security | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: Token generated with negative expiration (`time.Now().Add(-1 * time.Hour)`).
* **Test Data**: `Authorization: Bearer <expired_token>`.
* **Test Steps**: Dispatch `GET /api/v1/auth/me` with expired token.
* **Expected Result**: HTTP 401 Unauthorized; error code `TOKEN_EXPIRED`.
* **Actual Result**: HTTP 401 Unauthorized returned.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/tests/integration/auth_api_integration_test.go:TestAuthAPI_ExpiredToken_Rejected`
* **Evidence**: Integration test PASS (0.03s).

---

### TC-P1-AUT-014: Reject Algorithm Confusion Attack (`alg: "none"`)
* **Phase**: Phase 1 | **Module**: Auth | **Feature**: JWT Security
* **Test Type**: Security | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: None.
* **Test Data**: JWT constructed with `{"alg":"none","typ":"JWT"}` without signature.
* **Test Steps**: Dispatch `GET /api/v1/auth/me` with unsigned token.
* **Expected Result**: HTTP 401 Unauthorized; token rejected.
* **Actual Result**: HTTP 401 Unauthorized returned.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/tests/integration/security_comprehensive_test.go`
* **Evidence**: Automated in comprehensive security test suite.

---

### TC-P1-AUT-015: Reject Missing Authorization Header
* **Phase**: Phase 1 | **Module**: Auth | **Feature**: Route Protection
* **Test Type**: Security | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: None.
* **Test Data**: Request without `Authorization` header.
* **Test Steps**: Dispatch `GET /api/v1/auth/me`.
* **Expected Result**: HTTP 401 Unauthorized; error code `UNAUTHORIZED`.
* **Actual Result**: HTTP 401 Unauthorized returned.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/tests/integration/security_comprehensive_test.go`
* **Evidence**: Automated in comprehensive security test suite.

---

### TC-P1-AUT-016: Refresh Token Single-Use Rotation
* **Phase**: Phase 1 | **Module**: Auth | **Feature**: Token Rotation
* **Test Type**: API / Security | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: User has active refresh token $RT_1$.
* **Test Data**: `{"refresh_token": "<RT1>"}`.
* **Test Steps**: Dispatch `POST /api/v1/auth/refresh`.
* **Expected Result**: HTTP 200 OK; new token pair issued ($AT_2$, $RT_2$). $RT_1$ marked revoked in database.
* **Actual Result**: HTTP 200 OK; tokens rotated successfully.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/tests/integration/auth_api_integration_test.go:TestAuthAPI_TokenRefresh_Rotation_And_BreachDetection`
* **Evidence**: Integration test PASS (0.34s).

---

### TC-P1-AUT-017: Replayed Refresh Token Triggers Breach Detection
* **Phase**: Phase 1 | **Module**: Auth | **Feature**: Breach Detection
* **Test Type**: Security | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: $RT_1$ was previously rotated and marked revoked.
* **Test Data**: Re-submit $RT_1$ to `POST /api/v1/auth/refresh`.
* **Test Steps**: Dispatch request with already-revoked refresh token.
* **Expected Result**: HTTP 401 Unauthorized; all refresh tokens for the user account are revoked immediately.
* **Actual Result**: Replay detected, token family invalidated; subsequent requests with child token fail.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/tests/integration/auth_api_integration_test.go:TestAuthAPI_TokenRefresh_Rotation_And_BreachDetection`
* **Evidence**: Integration test PASS.

---

### TC-P1-AUT-018: Logout Revokes Refresh Session
* **Phase**: Phase 1 | **Module**: Auth | **Feature**: Session Termination
* **Test Type**: API / Security | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: Active refresh token exists.
* **Test Data**: `POST /api/v1/auth/logout` with Bearer access token and refresh token payload.
* **Test Steps**: Dispatch logout, then attempt to refresh session with the same token.
* **Expected Result**: Logout returns HTTP 200 OK; subsequent refresh attempt returns HTTP 401 Unauthorized.
* **Actual Result**: Token revoked; refresh rejected with 401.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/tests/integration/auth_api_integration_test.go:TestAuthAPI_Logout_Revocation`
* **Evidence**: Integration test PASS (0.32s).

---

### TC-P1-TNT-001: Create Tenant and Assign Owner Role
* **Phase**: Phase 1 | **Module**: Tenant | **Feature**: Organization Onboarding
* **Test Type**: API / Integration | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: Authenticated user.
* **Test Data**: `{"name":"Velocity Courier","slug":"velocity-courier-<uuid>","contact_email":"ops@velocity.test"}`.
* **Test Steps**: Dispatch `POST /api/v1/tenants`.
* **Expected Result**: HTTP 201 Created; user automatically granted `TENANT_ADMIN` role in `tenant_memberships`.
* **Actual Result**: HTTP 201 Created; membership created with `role: "TENANT_ADMIN"`.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/tests/integration/tenant_security_integration_test.go`
* **Evidence**: Integration test PASS.

---

### TC-P1-TNT-002: Reject Duplicate Tenant Slug
* **Phase**: Phase 1 | **Module**: Tenant | **Feature**: Organization Validation
* **Test Type**: API | **Priority**: P1 | **Severity**: Major | **Automated**: Automated
* **Preconditions**: Tenant with slug `velocity-courier` already exists.
* **Test Data**: `{"name":"Another Org","slug":"velocity-courier"}`.
* **Test Steps**: Dispatch `POST /api/v1/tenants` with duplicate slug.
* **Expected Result**: HTTP 409 Conflict; error code `SLUG_ALREADY_EXISTS`.
* **Actual Result**: HTTP 409 Conflict returned.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/tests/integration/tenant_security_integration_test.go`
* **Evidence**: Integration test PASS.

---

### TC-P1-TNT-003: List Accessible Tenants for Current User
* **Phase**: Phase 1 | **Module**: Tenant | **Feature**: Tenant Switching
* **Test Type**: API | **Priority**: P1 | **Severity**: Major | **Automated**: Automated
* **Preconditions**: User belongs to Tenant Alpha and Tenant Beta.
* **Test Steps**: Dispatch `GET /api/v1/tenants`.
* **Expected Result**: HTTP 200 OK; returns list containing only the tenants where the user has active membership.
* **Actual Result**: HTTP 200 OK; exactly the user's tenants returned.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/tests/integration/tenant_security_integration_test.go`
* **Evidence**: Integration test PASS.

---

### TC-P1-TNT-004: Update Tenant Metadata via PATCH (Admin Allowed)
* **Phase**: Phase 1 | **Module**: Tenant | **Feature**: Organization Management
* **Test Type**: API | **Priority**: P1 | **Severity**: Major | **Automated**: Automated
* **Preconditions**: Authenticated user is `TENANT_ADMIN`.
* **Test Data**: `PATCH /api/v1/tenants/:id` with `{"name":"Velocity Global","contact_email":"support@velocity.test"}`.
* **Test Steps**: Dispatch PATCH request.
* **Expected Result**: HTTP 200 OK; tenant name updated in database.
* **Actual Result**: HTTP 200 OK; updated record returned.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/tests/integration/tenant_security_integration_test.go:TestSecurity_TenantUpdate_PATCH`
* **Evidence**: Integration test PASS (1.73s).

---

### TC-P1-TNT-005: Reject Tenant Update by Non-Admin Member
* **Phase**: Phase 1 | **Module**: Tenant | **Feature**: Authorization Security
* **Test Type**: Security | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: Authenticated user is `TENANT_OPERATOR` in the tenant.
* **Test Steps**: Dispatch `PATCH /api/v1/tenants/:id` with update payload.
* **Expected Result**: HTTP 403 Forbidden; error code `INSUFFICIENT_PERMISSIONS`.
* **Actual Result**: HTTP 403 Forbidden returned.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/tests/integration/tenant_security_integration_test.go:TestSecurity_TenantUpdate_PATCH`
* **Evidence**: Integration test PASS.

---

### TC-P1-TNT-006: Reject Cross-Tenant Resource Access Attempt
* **Phase**: Phase 1 | **Module**: Tenant | **Feature**: Multi-Tenant Isolation
* **Test Type**: Security | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: User belongs exclusively to Tenant Alpha; Tenant Beta exists.
* **Test Steps**: User attempts `GET /api/v1/tenants/<tenant-beta-id>`.
* **Expected Result**: HTTP 403 Forbidden; error code `CROSS_TENANT_ACCESS_DENIED`.
* **Actual Result**: HTTP 403 Forbidden returned.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/tests/integration/tenant_security_integration_test.go:TestSecurity_CrossTenantAccess_Forbidden`
* **Evidence**: Integration test PASS (0.74s).

---

### TC-P1-MEM-001: Tenant Admin Invites Member with Valid Role
* **Phase**: Phase 1 | **Module**: Membership | **Feature**: Member Management
* **Test Type**: API | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: User A is `TENANT_ADMIN`; target user B is registered.
* **Test Data**: `POST /api/v1/tenants/:id/members` with `{"email":"operator@logiflows.test","role":"TENANT_OPERATOR"}`.
* **Test Steps**: Dispatch member invitation request.
* **Expected Result**: HTTP 201 Created; membership record created.
* **Actual Result**: HTTP 201 Created; member visible in `GET /members`.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/tests/integration/tenant_security_integration_test.go`
* **Evidence**: Integration test PASS.

---

### TC-P1-MEM-002: Reject Member Invitation by Non-Admin
* **Phase**: Phase 1 | **Module**: Membership | **Feature**: Authorization Security
* **Test Type**: Security | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: User is `TENANT_OPERATOR` in the tenant.
* **Test Steps**: User attempts to invite a member via `POST /api/v1/tenants/:id/members`.
* **Expected Result**: HTTP 403 Forbidden; error code `INSUFFICIENT_PERMISSIONS`.
* **Actual Result**: HTTP 403 Forbidden returned.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/tests/integration/tenant_security_integration_test.go:TestSecurity_RBAC_RolePermissionEnforcement`
* **Evidence**: Integration test PASS (2.44s).

---

### TC-P1-MEM-003: Reject Duplicate Membership for Same User and Tenant
* **Phase**: Phase 1 | **Module**: Membership | **Feature**: Data Integrity
* **Test Type**: API | **Priority**: P1 | **Severity**: Major | **Automated**: Automated
* **Preconditions**: Target user already holds membership in the tenant.
* **Test Steps**: Attempt to invite user a second time to same tenant.
* **Expected Result**: HTTP 409 Conflict; unique constraint `uq_tenant_user` enforced.
* **Actual Result**: HTTP 409 Conflict returned.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/tests/integration/security_comprehensive_test.go`
* **Evidence**: Automated in comprehensive security test suite.

---

### TC-P1-RBC-001: RBAC Permission Matrix Enforcement
* **Phase**: Phase 1 | **Module**: Authorization | **Feature**: Role-Based Access Control
* **Test Type**: Unit / Security | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: Roles: `PLATFORM_ADMIN`, `TENANT_ADMIN`, `TENANT_OPERATOR`, `VIEWER`.
* **Test Steps**: Evaluate `HasPermission()` for every role against every defined permission.
* **Expected Result**: Matrix enforces least privilege; Platform Admin has wildcard access; Viewer has read-only access.
* **Actual Result**: All permission assertions pass 100%.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/internal/authorization/roles_test.go`
* **Evidence**: Unit test PASS (100% statement coverage).

---

### TC-P1-AUD-001: Verify Audit Log Creation on Critical Events
* **Phase**: Phase 1 | **Module**: Audit | **Feature**: Audit Trail
* **Test Type**: Integration / Security | **Priority**: P1 | **Severity**: Major | **Automated**: Automated
* **Preconditions**: User performs registration, tenant creation, and member invitation.
* **Test Steps**: Query PostgreSQL `audit_logs` table directly for matching `actor_id` and actions.
* **Expected Result**: Corresponding rows exist with valid timestamps, client IP, action name, and non-empty JSON metadata.
* **Actual Result**: Audit rows verified in database.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/tests/integration/security_comprehensive_test.go`
* **Evidence**: Automated in comprehensive security test suite.

---

### TC-P1-MOB-001: AuthUser JSON Deserialization & Boundary Defaults
* **Phase**: Phase 1 | **Module**: Mobile Models | **Feature**: Identity Models
* **Test Type**: Unit | **Priority**: P1 | **Severity**: Major | **Automated**: Automated
* **Preconditions**: Raw JSON payload from `/api/v1/auth/me` or `/login`.
* **Test Steps**: Parse JSON with full fields and minimal fields using `AuthUser.fromJson()`.
* **Expected Result**: Maps all fields (`id`, `email`, `fullName`, `phoneNumber`); applies safe defaults for boolean flags (`isActive: true`, `isPlatformAdmin: false`, `emailVerified: false`).
* **Actual Result**: Field assertions and fallback defaults verified.
* **Status**: **PASS** | **Execution Date**: 2026-09-22 | **Regression Required**: Yes
* **Related Code**: `mobile/lib/models/auth_models.dart`, `mobile/test/auth_models_test.dart`
* **Evidence**: `test/auth_models_test.dart passed`.

---

### TC-P1-MOB-002: Tenant & Tenant Membership JSON Deserialization
* **Phase**: Phase 1 | **Module**: Mobile Models | **Feature**: Multi-Tenancy Models
* **Test Type**: Unit | **Priority**: P1 | **Severity**: Major | **Automated**: Automated
* **Preconditions**: Raw JSON tenant payload.
* **Test Steps**: Invoke `Tenant.fromJson()`.
* **Expected Result**: Correctly maps `id`, `name`, `slug`, and `role` (`DRIVER`, `TENANT_ADMIN`, etc.).
* **Actual Result**: Deserialization verified.
* **Status**: **PASS** | **Execution Date**: 2026-09-22 | **Regression Required**: Yes
* **Related Code**: `mobile/lib/models/auth_models.dart`, `mobile/test/auth_models_test.dart`
* **Evidence**: `test/auth_models_test.dart passed`.

---

### TC-P1-MOB-003: AuthResponse Token & Nested User/Tenants Deserialization
* **Phase**: Phase 1 | **Module**: Mobile Models | **Feature**: Authentication Payload
* **Test Type**: Unit | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: Raw authentication response containing token pair, user object, and tenant array.
* **Test Steps**: Parse payload via `AuthResponse.fromJson()`.
* **Expected Result**: Successfully parses access token, expiration, refresh token, user profile, and list of tenants (handling empty/null tenant arrays safely).
* **Actual Result**: Nested models and edge cases verified.
* **Status**: **PASS** | **Execution Date**: 2026-09-22 | **Regression Required**: Yes
* **Related Code**: `mobile/lib/models/auth_models.dart`, `mobile/test/auth_models_test.dart`
* **Evidence**: `test/auth_models_test.dart passed`.

---

### TC-P1-MOB-004: AuthApiClient Successful Login & Credential Storing
* **Phase**: Phase 1 | **Module**: Mobile Services | **Feature**: User Login Flow
* **Test Type**: Unit / Integration | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: Mobile client initialized with `TokenStorage` and mock HTTP client.
* **Test Steps**: Invoke `authClient.login()` with email and password.
* **Expected Result**: Sends normalized lowercase email in POST body, parses `AuthResponse`, and persists token pair into secure storage.
* **Actual Result**: Returns `AuthResponse`, `tokenStorage.hasValidToken()` is `true`.
* **Status**: **PASS** | **Execution Date**: 2026-09-22 | **Regression Required**: Yes
* **Related Code**: `mobile/lib/services/auth_api_client.dart`, `mobile/test/auth_api_client_test.dart`
* **Evidence**: `test/auth_api_client_test.dart: TC-P1-MOB-004 passed`.

---

### TC-P1-MOB-005: AuthApiClient Login Error Extraction & Handling
* **Phase**: Phase 1 | **Module**: Mobile Services | **Feature**: Authentication Error Handling
* **Test Type**: Unit | **Priority**: P1 | **Severity**: Major | **Automated**: Automated
* **Preconditions**: Backend responds with HTTP 401 and JSON error envelope.
* **Test Steps**: Invoke `authClient.login()` with invalid credentials.
* **Expected Result**: Throws descriptive `Exception` containing `body.error.message` ("Invalid email or password").
* **Actual Result**: Exception captured with exact message; storage remains unauthenticated.
* **Status**: **PASS** | **Execution Date**: 2026-09-22 | **Regression Required**: Yes
* **Related Code**: `mobile/lib/services/auth_api_client.dart`, `mobile/test/auth_api_client_test.dart`
* **Evidence**: `test/auth_api_client_test.dart: TC-P1-MOB-005 passed`.

---

### TC-P1-MOB-006: AuthApiClient Organization & Driver Registration
* **Phase**: Phase 1 | **Module**: Mobile Services | **Feature**: Organization Registration Flow
* **Test Type**: Unit / Integration | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: Valid registration form inputs.
* **Test Steps**: Dispatch `authClient.register()`.
* **Expected Result**: Dispatches HTTP POST to `/api/v1/auth/register`, parses returned `AuthResponse`, saves token pair in storage.
* **Actual Result**: New user and organization registered; storage updated.
* **Status**: **PASS** | **Execution Date**: 2026-09-22 | **Regression Required**: Yes
* **Related Code**: `mobile/lib/services/auth_api_client.dart`, `mobile/test/auth_api_client_test.dart`
* **Evidence**: `test/auth_api_client_test.dart: TC-P1-MOB-006 passed`.

---

### TC-P1-MOB-007: AuthApiClient Profile Retrieval via Bearer Token
* **Phase**: Phase 1 | **Module**: Mobile Services | **Feature**: Profile Retrieval
* **Test Type**: Unit | **Priority**: P1 | **Severity**: Major | **Automated**: Automated
* **Preconditions**: Valid access token in storage.
* **Test Steps**: Call `authClient.getCurrentUser()`.
* **Expected Result**: Requests `/api/v1/auth/me` with `Authorization: Bearer <token>`, returns deserialized `AuthUser`.
* **Actual Result**: User profile retrieved; headers matched.
* **Status**: **PASS** | **Execution Date**: 2026-09-22 | **Regression Required**: Yes
* **Related Code**: `mobile/lib/services/auth_api_client.dart`, `mobile/test/auth_api_client_test.dart`
* **Evidence**: `test/auth_api_client_test.dart: TC-P1-MOB-007 passed`.

---

### TC-P1-MOB-008: AuthApiClient Token Refresh & Auto-Recovery
* **Phase**: Phase 1 | **Module**: Mobile Services | **Feature**: Token Refresh Lifecycle
* **Test Type**: Unit / Security | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: Stored access token expired, valid refresh token present.
* **Test Steps**: Invoke `authClient.refreshToken()`.
* **Expected Result**: Exchanges refresh token via `/api/v1/auth/refresh`, persists rotated tokens. If refresh token is revoked/invalid, clears local storage and throws session revoked error.
* **Actual Result**: Both successful rotation and revoked token cleanup verified.
* **Status**: **PASS** | **Execution Date**: 2026-09-22 | **Regression Required**: Yes
* **Related Code**: `mobile/lib/services/auth_api_client.dart`, `mobile/test/auth_api_client_test.dart`
* **Evidence**: `test/auth_api_client_test.dart: TC-P1-MOB-008 passed`.

---

### TC-P1-MOB-009: AuthApiClient Logout & Token Invalidation
* **Phase**: Phase 1 | **Module**: Mobile Services | **Feature**: Logout & Security
* **Test Type**: Unit / Security | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: Active credentials in mobile storage.
* **Test Steps**: Invoke `authClient.logout()`.
* **Expected Result**: Posts refresh token to `/api/v1/auth/logout`, purges local tokens unconditionally even if network fails.
* **Actual Result**: Server notified; storage wiped cleanly.
* **Status**: **PASS** | **Execution Date**: 2026-09-22 | **Regression Required**: Yes
* **Related Code**: `mobile/lib/services/auth_api_client.dart`, `mobile/test/auth_api_client_test.dart`
* **Evidence**: `test/auth_api_client_test.dart: TC-P1-MOB-009 passed`.

---

### TC-P1-MOB-010: LoginScreen UI Rendering & Layout Integrity
* **Phase**: Phase 1 | **Module**: Mobile UI | **Feature**: Login Screen
* **Test Type**: Widget | **Priority**: P1 | **Severity**: Major | **Automated**: Automated
* **Preconditions**: Unauthenticated session.
* **Test Steps**: Pump `LoginScreen` in widget tester.
* **Expected Result**: Renders vehicle logo, `LogiFlows Driver Pro` title, subtitle, `Corporate Email` field, `Password` field, `Sign In` button, and `Register Company` link.
* **Actual Result**: All components and styling rendered as expected.
* **Status**: **PASS** | **Execution Date**: 2026-09-22 | **Regression Required**: Yes
* **Related Code**: `mobile/lib/screens/login_screen.dart`, `mobile/test/login_screen_test.dart`, `mobile/test/widget_test.dart`
* **Evidence**: Widget test passed.

---

### TC-P1-MOB-011: LoginScreen Form Validation (Email & Password Constraints)
* **Phase**: Phase 1 | **Module**: Mobile UI | **Feature**: Input Validation
* **Test Type**: Widget | **Priority**: P1 | **Severity**: Major | **Automated**: Automated
* **Preconditions**: `LoginScreen` active.
* **Test Steps**: Attempt sign-in with: empty inputs, invalid email format, password < 8 characters.
* **Expected Result**: Form validators show "Email is required", "Enter a valid email address", "Password is required", "Password must be at least 8 characters".
* **Actual Result**: All validation messages displayed without network dispatch.
* **Status**: **PASS** | **Execution Date**: 2026-09-22 | **Regression Required**: Yes
* **Related Code**: `mobile/lib/screens/login_screen.dart`, `mobile/test/login_screen_test.dart`
* **Evidence**: Widget test passed.

---

### TC-P1-MOB-012: LoginScreen Authentication Error Banner Display
* **Phase**: Phase 1 | **Module**: Mobile UI | **Feature**: Error Presentation
* **Test Type**: Widget | **Priority**: P1 | **Severity**: Major | **Automated**: Automated
* **Preconditions**: `LoginScreen` active, mock API configured to return 401.
* **Test Steps**: Enter credentials, tap "Sign In", settle async call.
* **Expected Result**: Red error banner displayed with error icon and server error text.
* **Actual Result**: Error banner rendered with expected text.
* **Status**: **PASS** | **Execution Date**: 2026-09-22 | **Regression Required**: Yes
* **Related Code**: `mobile/lib/screens/login_screen.dart`, `mobile/test/login_screen_test.dart`
* **Evidence**: Widget test passed.

---

### TC-P1-MOB-013: RegisterScreen UI Rendering & Field Validation
* **Phase**: Phase 1 | **Module**: Mobile UI | **Feature**: Registration Screen
* **Test Type**: Widget | **Priority**: P1 | **Severity**: Major | **Automated**: Automated
* **Preconditions**: User navigates to RegisterScreen.
* **Test Steps**: Verify fields (`Full Name`, `Company / Fleet Name`, `Corporate Email`, `Password`). Submit empty form.
* **Expected Result**: Form rejects submission with descriptive errors for each required field.
* **Actual Result**: All input validators triggered properly.
* **Status**: **PASS** | **Execution Date**: 2026-09-22 | **Regression Required**: Yes
* **Related Code**: `mobile/lib/screens/register_screen.dart`, `mobile/test/register_screen_test.dart`
* **Evidence**: Widget test passed.

---

### TC-P1-MOB-014: Mobile Authentication State Flow & Session Switching
* **Phase**: Phase 1 | **Module**: Mobile App | **Feature**: State Flow & Routing
* **Test Type**: Widget / Integration | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: Initial launch or session token state changes.
* **Test Steps**: Test app with no token -> renders `LoginScreen`. Pre-seed valid token in storage -> app boots directly into `DriverDashboardScreen`.
* **Expected Result**: Dynamic routing based on token storage state.
* **Actual Result**: Seamless route transition without intermediate errors.
* **Status**: **PASS** | **Execution Date**: 2026-09-22 | **Regression Required**: Yes
* **Related Code**: `mobile/lib/main.dart`, `mobile/test/widget_test.dart`
* **Evidence**: `test/widget_test.dart: TC-P1-MOB-014 passed`.

---

### TC-P1-MOB-015: Network Connection Failure & User-Friendly Error Formatting
* **Phase**: Phase 1 | **Module**: Mobile Services | **Feature**: Network Error Resilience
* **Test Type**: Unit | **Priority**: P1 | **Severity**: Major | **Automated**: Automated
* **Preconditions**: Backend service unreachable or socket fails (e.g. `ClientException`, `SocketException`, `Failed to fetch`).
* **Test Steps**: Dispatch `authClient.login()`, trigger simulated `ClientException`.
* **Expected Result**: Catches low-level exception and throws user-friendly message (`Unable to connect to LogiFlows API...`) rather than unhandled platform crash or raw exception string.
* **Actual Result**: User-friendly error message generated; `tokenStorage` remains safe.
* **Status**: **PASS** | **Execution Date**: 2026-09-22 | **Regression Required**: Yes
* **Related Code**: `mobile/lib/services/auth_api_client.dart`, `mobile/lib/screens/login_screen.dart`, `mobile/test/auth_api_client_test.dart`
* **Evidence**: `test/auth_api_client_test.dart: TC-P1-MOB-015 passed`.


