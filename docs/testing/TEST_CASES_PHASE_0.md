# LogiFlows — Phase 0 Test Case Catalog (Foundation)

**Document Reference**: `docs/testing/TEST_CASES_PHASE_0.md`  
**Phase**: Phase 0 — Complete Project Foundation  
**Status**: 100% EXECUTED & PASSED  

---

### TC-P0-CFG-001: Load Default Configuration
* **Phase**: Phase 0 | **Module**: Config | **Feature**: Environment Configuration
* **Test Type**: Unit | **Priority**: P1 | **Severity**: Major | **Automated**: Automated
* **Preconditions**: No environment variables explicitly set.
* **Test Data**: Defaults: `APP_ENV=development`, `APP_PORT=8080`, `DB_PORT=5432`.
* **Test Steps**: Call `config.Load()`.
* **Expected Result**: Configuration loads default values without error.
* **Actual Result**: Defaults loaded successfully; `App.Port == 8080`.
* **Status**: **PASS** | **Execution Date**: 2026-09-20 | **Regression Required**: Yes
* **Related Code**: `backend/internal/config/config.go`
* **Evidence**: `ok github.com/logiflows/logiflows/backend/internal/config 0.00s`

---

### TC-P0-CFG-002: Reject Invalid Environment Name
* **Phase**: Phase 0 | **Module**: Config | **Feature**: Environment Validation
* **Test Type**: Unit | **Priority**: P1 | **Severity**: Moderate | **Automated**: Automated
* **Preconditions**: None.
* **Test Data**: `APP_ENV=invalid_environment`.
* **Test Steps**: Invoke `config.Load()` with invalid environment string.
* **Expected Result**: Returns descriptive error: `invalid APP_ENV`.
* **Actual Result**: Validation error returned as expected.
* **Status**: **PASS** | **Execution Date**: 2026-09-20 | **Regression Required**: Yes
* **Related Code**: `backend/internal/config/config_test.go:TestConfig_Validate_InvalidEnv`
* **Evidence**: Logged error matches regex `invalid APP_ENV`.

---

### TC-P0-CFG-003: Reject Invalid Port Number (>65535)
* **Phase**: Phase 0 | **Module**: Config | **Feature**: Port Validation
* **Test Type**: Unit | **Priority**: P1 | **Severity**: Moderate | **Automated**: Automated
* **Preconditions**: None.
* **Test Data**: `APP_PORT=70000`.
* **Test Steps**: Invoke `config.Load()` with port > 65535.
* **Expected Result**: Validation fails with `APP_PORT must be between 1 and 65535`.
* **Actual Result**: Validation rejected with expected error.
* **Status**: **PASS** | **Execution Date**: 2026-09-20 | **Regression Required**: Yes
* **Related Code**: `backend/internal/config/config_test.go:TestConfig_Validate_InvalidPort`
* **Evidence**: Unit test PASS.

---

### TC-P0-CFG-004: Reject Database Max Idle Connections Exceeding Max Open
* **Phase**: Phase 0 | **Module**: Config | **Feature**: DB Pool Validation
* **Test Type**: Unit | **Priority**: P2 | **Severity**: Moderate | **Automated**: Automated
* **Preconditions**: None.
* **Test Data**: `DB_MAX_OPEN_CONNS=10`, `DB_MAX_IDLE_CONNS=20`.
* **Test Steps**: Invoke `config.Load()`.
* **Expected Result**: Rejected with error stating idle connections cannot exceed open connections.
* **Actual Result**: Rejected cleanly.
* **Status**: **PASS** | **Execution Date**: 2026-09-20 | **Regression Required**: Yes
* **Related Code**: `backend/internal/config/config_test.go:TestConfig_Validate_DBIdleConnsExceedsMaxOpen`
* **Evidence**: Unit test PASS.

---

### TC-P0-CFG-005: Enforce Secure JWT Secret Length in Production
* **Phase**: Phase 0 | **Module**: Config | **Feature**: Production Security Rules
* **Test Type**: Unit | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: `APP_ENV=production`.
* **Test Data**: `JWT_SECRET=short` (< 32 bytes).
* **Test Steps**: Invoke `config.Load()` in production environment.
* **Expected Result**: Startup aborted with error `JWT_SECRET must be at least 32 characters in production`.
* **Actual Result**: Startup aborted safely.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/internal/config/config.go`
* **Evidence**: Unit test PASS.

---

### TC-P0-LOG-001: Structured Logger Contextual Request ID Propagation
* **Phase**: Phase 0 | **Module**: Logger | **Feature**: Observability
* **Test Type**: Unit | **Priority**: P2 | **Severity**: Moderate | **Automated**: Automated
* **Preconditions**: Request context contains `request_id` key.
* **Test Data**: `request_id=trace-uuid-12345`.
* **Test Steps**: Invoke `logger.InfoContext(ctx, "Test Message")` with JSON format.
* **Expected Result**: Output JSON string contains `"request_id":"trace-uuid-12345"`.
* **Actual Result**: Field verified in parsed JSON log.
* **Status**: **PASS** | **Execution Date**: 2026-09-20 | **Regression Required**: Yes
* **Related Code**: `backend/internal/logger/logger_test.go:TestLogger_JSON_WithRequestID`
* **Evidence**: Output: `{"level":"INFO","msg":"Test Message","request_id":"trace-uuid-12345"}`.

---

### TC-P0-MID-001: Request ID Middleware UUID Generation
* **Phase**: Phase 0 | **Module**: Middleware | **Feature**: Tracing
* **Test Type**: Unit | **Priority**: P1 | **Severity**: Moderate | **Automated**: Automated
* **Preconditions**: Incoming HTTP request lacks `X-Request-ID` header.
* **Test Data**: `GET /api/v1/health`.
* **Test Steps**: Pass request through `middleware.RequestID()`.
* **Expected Result**: Response header contains newly generated valid UUIDv4.
* **Actual Result**: Valid UUIDv4 returned in `X-Request-Id`.
* **Status**: **PASS** | **Execution Date**: 2026-09-20 | **Regression Required**: Yes
* **Related Code**: `backend/internal/middleware/request_id_test.go`
* **Evidence**: UUID regex matched `^[0-9a-fA-F-]{36}$`.

---

### TC-P0-MID-002: Request ID Middleware Preserves Client Header
* **Phase**: Phase 0 | **Module**: Middleware | **Feature**: Tracing
* **Test Type**: Unit | **Priority**: P1 | **Severity**: Moderate | **Automated**: Automated
* **Preconditions**: Incoming request includes `X-Request-ID: client-custom-id-999`.
* **Test Data**: `X-Request-ID: client-custom-id-999`.
* **Test Steps**: Pass request through middleware pipeline.
* **Expected Result**: Response header retains `X-Request-Id: client-custom-id-999`.
* **Actual Result**: Custom ID preserved.
* **Status**: **PASS** | **Execution Date**: 2026-09-20 | **Regression Required**: Yes
* **Related Code**: `backend/internal/middleware/request_id_test.go`
* **Evidence**: Assertion verified header match.

---

### TC-P0-MID-003: Recovery Middleware Graceful Panic Handling
* **Phase**: Phase 0 | **Module**: Middleware | **Feature**: Resilience
* **Test Type**: Unit | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: Downstream handler raises an unhandled panic.
* **Test Data**: `panic("unexpected database segmentation crash")`.
* **Test Steps**: Invoke handler wrapped in `middleware.Recovery()`.
* **Expected Result**: Server catches panic, logs stack trace, returns HTTP 500 JSON envelope.
* **Actual Result**: HTTP 500 returned with error code `INTERNAL_SERVER_ERROR`.
* **Status**: **PASS** | **Execution Date**: 2026-09-20 | **Regression Required**: Yes
* **Related Code**: `backend/internal/middleware/recovery_test.go`
* **Evidence**: Response envelope verified.

---

### TC-P0-HLT-001: Health Liveness Probe Endpoint
* **Phase**: Phase 0 | **Module**: Health | **Feature**: Liveness
* **Test Type**: API | **Priority**: P1 | **Severity**: Major | **Automated**: Automated
* **Preconditions**: Server is running.
* **Test Data**: `GET /api/v1/health`.
* **Test Steps**: Dispatch GET request to `/api/v1/health`.
* **Expected Result**: HTTP 200 OK, `status: "ok"`, service: `"logiflows-api"`.
* **Actual Result**: HTTP 200 OK, JSON payload matches contract.
* **Status**: **PASS** | **Execution Date**: 2026-09-20 | **Regression Required**: Yes
* **Related Code**: `backend/internal/health/handler_test.go`
* **Evidence**: Latency < 2ms, HTTP 200.

---

### TC-P0-HLT-002: Health Readiness Probe (All Healthy)
* **Phase**: Phase 0 | **Module**: Health | **Feature**: Readiness
* **Test Type**: API | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: PostgreSQL and Redis containers are healthy and reachable.
* **Test Data**: `GET /api/v1/readiness`.
* **Test Steps**: Dispatch GET request to `/api/v1/readiness`.
* **Expected Result**: HTTP 200 OK, checks show database: `"UP"`, redis: `"UP"`.
* **Actual Result**: HTTP 200 OK with latencies reported for both services.
* **Status**: **PASS** | **Execution Date**: 2026-09-20 | **Regression Required**: Yes
* **Related Code**: `backend/internal/health/handler_test.go`
* **Evidence**: HTTP 200, checks reported healthy.

---

### TC-P0-HLT-003: Health Readiness Probe (Degraded Dependency)
* **Phase**: Phase 0 | **Module**: Health | **Feature**: Readiness
* **Test Type**: API | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: Database pinger returns error.
* **Test Data**: Injected failed DB checker.
* **Test Steps**: Dispatch GET request to `/api/v1/readiness`.
* **Expected Result**: HTTP 503 Service Unavailable, `status: "degraded"`.
* **Actual Result**: HTTP 503 returned.
* **Status**: **PASS** | **Execution Date**: 2026-09-20 | **Regression Required**: Yes
* **Related Code**: `backend/internal/health/handler_test.go`
* **Evidence**: HTTP 503 verified.

---

### TC-P0-DB-001: PostgreSQL Connection Pool & PostGIS Extension
* **Phase**: Phase 0 | **Module**: Database | **Feature**: Spatial DB Connectivity
* **Test Type**: Integration | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: `logiflows-postgres` container running.
* **Test Data**: Query `SELECT PostGIS_Full_Version();`.
* **Test Steps**: Connect via `pgxpool`, execute query, assert version string.
* **Expected Result**: Query returns PostGIS version string (3.4+).
* **Actual Result**: `PostGIS Full Version: POSTGIS="3.4.3 e365945"`. Ping: 1.96ms.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/tests/integration/postgres_integration_test.go`
* **Evidence**: Test output line 54.

---

### TC-P0-RDS-001: Redis Client Connectivity & Cache Operations
* **Phase**: Phase 0 | **Module**: Redis | **Feature**: Cache Operations
* **Test Type**: Integration | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: `logiflows-redis` container running.
* **Test Data**: Key: `test:ping:probe`.
* **Test Steps**: Connect, execute `PING`, `SET key`, `GET key`, `DEL key`.
* **Expected Result**: All operations succeed without error; ping latency < 5ms.
* **Actual Result**: Redis Ping Latency: 530.1µs. Operations succeeded.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/tests/integration/redis_integration_test.go`
* **Evidence**: Test output line 57.

---

### TC-P0-MIG-001: Goose Database Migrations Run Up
* **Phase**: Phase 0 | **Module**: Migration | **Feature**: Schema Automation
* **Test Type**: Integration | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: Database reachable.
* **Test Data**: Migrations directory `backend/migrations`.
* **Test Steps**: Execute `migrations.RunUp()`.
* **Expected Result**: All migrations up to latest version apply idempotently.
* **Actual Result**: Migrated to version 3 successfully.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/tests/integration/migration_test.go`
* **Evidence**: Output: `goose: successfully migrated database to version: 3`.

---

### TC-P0-MIG-002: Database Migrations Rollback and Reapply
* **Phase**: Phase 0 | **Module**: Migration | **Feature**: Schema Rollback
* **Test Type**: Integration | **Priority**: P1 | **Severity**: Major | **Automated**: Automated
* **Preconditions**: Database at version 3.
* **Test Steps**: Execute `migrations.RunDown()`, assert version 2, execute `migrations.RunUp()`, assert version 3.
* **Expected Result**: Migration rolls back safely and re-applies cleanly.
* **Actual Result**: Successfully rolled back and re-applied in 0.09s.
* **Status**: **PASS** | **Execution Date**: 2026-09-21 | **Regression Required**: Yes
* **Related Code**: `backend/tests/integration/migration_test.go`
* **Evidence**: Test output line 50.

---

### TC-P0-SRV-001: Unified 404 Not Found Envelope
* **Phase**: Phase 0 | **Module**: Server | **Feature**: Error Handling
* **Test Type**: API | **Priority**: P2 | **Severity**: Moderate | **Automated**: Automated
* **Preconditions**: API router active.
* **Test Data**: `GET /api/v1/non-existent-route`.
* **Test Steps**: Dispatch request to unmapped route.
* **Expected Result**: HTTP 404, JSON body with `error.code: "NOT_FOUND"`.
* **Actual Result**: HTTP 404 with standardized error JSON.
* **Status**: **PASS** | **Execution Date**: 2026-09-20 | **Regression Required**: Yes
* **Related Code**: `backend/internal/server/router.go`
* **Evidence**: Response: `{"error":{"code":"NOT_FOUND","message":"The requested resource was not found"}}`.

---

### TC-P0-SRV-002: Unified 405 Method Not Allowed Envelope
* **Phase**: Phase 0 | **Module**: Server | **Feature**: Error Handling
* **Test Type**: API | **Priority**: P2 | **Severity**: Moderate | **Automated**: Automated
* **Preconditions**: API router active.
* **Test Data**: `DELETE /api/v1/health`.
* **Test Steps**: Dispatch unsupported HTTP method on existing route.
* **Expected Result**: HTTP 405, JSON body with `error.code: "METHOD_NOT_ALLOWED"`.
* **Actual Result**: HTTP 405 returned with standardized error JSON.
* **Status**: **PASS** | **Execution Date**: 2026-09-20 | **Regression Required**: Yes
* **Related Code**: `backend/internal/server/router.go`
* **Evidence**: HTTP 405 verified.

---

### TC-P0-MOB-001: Mobile API Configuration & Observability Endpoint Constants
* **Phase**: Phase 0 | **Module**: Mobile Core | **Feature**: API Configuration
* **Test Type**: Unit | **Priority**: P1 | **Severity**: Major | **Automated**: Automated
* **Preconditions**: None.
* **Test Data**: Base URL `http://10.0.2.2:8080/api/v1`, fallback `http://localhost:8080/api/v1`, AI URL `http://10.0.2.2:8000/api/v1`.
* **Test Steps**: Inspect `ApiConfig.baseUrl`, `ApiConfig.healthEndpoint`, `ApiConfig.readinessEndpoint`, and fallback URLs.
* **Expected Result**: Endpoints match standard LogiFlows microservice routing contracts.
* **Actual Result**: All constants point to valid host and port specifications.
* **Status**: **PASS** | **Execution Date**: 2026-09-22 | **Regression Required**: Yes
* **Related Code**: `mobile/lib/core/api_config.dart`, `mobile/test/api_config_test.dart`
* **Evidence**: `ApiConfig Unit Tests: TC-P0-MOB-001 passed`.

---

### TC-P0-MOB-002: Token Storage Initialization & Safe Defaults
* **Phase**: Phase 0 | **Module**: Mobile Storage | **Feature**: Secure Token Storage
* **Test Type**: Unit | **Priority**: P1 | **Severity**: Major | **Automated**: Automated
* **Preconditions**: Fresh instance of `InMemorySecureTokenStorage`.
* **Test Steps**: Call `hasValidToken()`, `getAccessToken()`, and `getRefreshToken()`.
* **Expected Result**: `hasValidToken()` returns `false`; token getters return `null`.
* **Actual Result**: Storage initializes in clean, unauthenticated state.
* **Status**: **PASS** | **Execution Date**: 2026-09-22 | **Regression Required**: Yes
* **Related Code**: `mobile/lib/core/token_storage.dart`, `mobile/test/token_storage_test.dart`
* **Evidence**: Unit test passed.

---

### TC-P0-MOB-003: Token Storage Multi-Token Persistence
* **Phase**: Phase 0 | **Module**: Mobile Storage | **Feature**: Secure Token Storage
* **Test Type**: Unit | **Priority**: P1 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: Storage initialized.
* **Test Data**: `accessToken: "access-12345"`, `refreshToken: "refresh-67890"`.
* **Test Steps**: Call `saveTokens()`, then query tokens and validity.
* **Expected Result**: Both tokens persisted; `hasValidToken()` returns `true`.
* **Actual Result**: Tokens match input parameters.
* **Status**: **PASS** | **Execution Date**: 2026-09-22 | **Regression Required**: Yes
* **Related Code**: `mobile/lib/core/token_storage.dart`, `mobile/test/token_storage_test.dart`
* **Evidence**: Unit test passed.

---

### TC-P0-MOB-004: Token Storage Clear on Logout
* **Phase**: Phase 0 | **Module**: Mobile Storage | **Feature**: Session Revocation
* **Test Type**: Unit | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: Storage holds active access and refresh tokens.
* **Test Steps**: Call `clearTokens()`, then query `hasValidToken()`, `getAccessToken()`, `getRefreshToken()`.
* **Expected Result**: Tokens wiped (`null`), `hasValidToken()` evaluates to `false`.
* **Actual Result**: Local storage state completely sanitized.
* **Status**: **PASS** | **Execution Date**: 2026-09-22 | **Regression Required**: Yes
* **Related Code**: `mobile/lib/core/token_storage.dart`, `mobile/test/token_storage_test.dart`
* **Evidence**: Unit test passed.

---

### TC-P0-MOB-005: Driver Dashboard Foundation Rendering
* **Phase**: Phase 0 | **Module**: Mobile UI | **Feature**: Driver Field Dashboard
* **Test Type**: Widget | **Priority**: P1 | **Severity**: Major | **Automated**: Automated
* **Preconditions**: Authenticated driver session context.
* **Test Steps**: Pump `DriverDashboardScreen` into widget tree.
* **Expected Result**: Renders AppBar (`LogiFlows Driver Pro`), system connectivity card, probe button, and custody parcel assignments.
* **Actual Result**: Header, parcel tracking numbers (`TRK-2026-9041`, `TRK-2026-9042`), and statuses rendered correctly.
* **Status**: **PASS** | **Execution Date**: 2026-09-22 | **Regression Required**: Yes
* **Related Code**: `mobile/lib/main.dart`, `mobile/test/driver_dashboard_test.dart`, `mobile/test/widget_test.dart`
* **Evidence**: Widget test passed.

---

### TC-P0-MOB-006: System Connectivity Probe & Status Transition
* **Phase**: Phase 0 | **Module**: Mobile UI | **Feature**: Network Connectivity Probing
* **Test Type**: Widget | **Priority**: P2 | **Severity**: Moderate | **Automated**: Automated
* **Preconditions**: `DriverDashboardScreen` rendered on screen.
* **Test Steps**: Tap "Probe LogiFlows Network" button, verify status chip transition to "Probing Network...", pump timer forward 650ms, verify transition to "Backend Operational".
* **Expected Result**: Status dynamically transitions from Amber "Probing Network..." to Green "Backend Operational".
* **Actual Result**: State transitions verified across timer ticks.
* **Status**: **PASS** | **Execution Date**: 2026-09-22 | **Regression Required**: Yes
* **Related Code**: `mobile/lib/main.dart`, `mobile/test/driver_dashboard_test.dart`
* **Evidence**: Widget test passed.

---

### TC-P0-MOB-007: Mobile App Bootstrapping & Unauthenticated Splash Resolution
* **Phase**: Phase 0 | **Module**: Mobile App | **Feature**: Lifecycle & Bootstrapping
* **Test Type**: Widget | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: Fresh launch with no stored access token.
* **Test Steps**: Pump `LogiFlowsApp` widget, observe initial splash indicator, settle asynchronous auth check.
* **Expected Result**: Shows `CircularProgressIndicator` during token check, then smoothly resolves to `LoginScreen`.
* **Actual Result**: Resolves unauthenticated user directly to login screen without flicker or exception.
* **Status**: **PASS** | **Execution Date**: 2026-09-22 | **Regression Required**: Yes
* **Related Code**: `mobile/lib/main.dart`, `mobile/test/widget_test.dart`
* **Evidence**: Widget test passed.

---

### TC-P0-MOB-008: Platform-Aware Host Resolution (Web, Android Emulator, Desktop)
* **Phase**: Phase 0 | **Module**: Mobile Core | **Feature**: Network Configuration
* **Test Type**: Unit | **Priority**: P0 | **Severity**: Critical | **Automated**: Automated
* **Preconditions**: Varied target runtime environments (`kIsWeb`, `TargetPlatform.android`, `TargetPlatform.windows`, `TargetPlatform.iOS`).
* **Test Steps**: Query `ApiConfig.defaultHost` and `ApiConfig.baseUrl` under Android emulation and Windows/Web desktop, then test `ApiConfig.setCustomHost()`.
* **Expected Result**: Android resolves to `10.0.2.2`; Web and Desktop resolve to `localhost`; custom host override takes precedence when set.
* **Actual Result**: `defaultHost` dynamically adapts to runtime environment; fixes `ClientFailed to fetch` on web/desktop.
* **Status**: **PASS** | **Execution Date**: 2026-09-22 | **Regression Required**: Yes
* **Related Code**: `mobile/lib/core/api_config.dart`, `mobile/test/api_config_test.dart`
* **Evidence**: `test/api_config_test.dart: TC-P0-MOB-008 passed`.


