# LogiFlows — Defect Reporting & Bug Registry

**Document Reference**: `docs/testing/BUG_REPORT.md`  
**Version**: 1.0.0  
**Status**: ACTIVE  

---

## 1. Defect Management Workflow

When an automated or manual test fails:
1. **Record**: Capture the Test Case ID, environment details, and full error log.
2. **Reproduce**: Verify the issue on a clean isolated container run.
3. **Analyze**: Identify the root cause (e.g., validation logic, schema constraint, race condition, or middleware ordering).
4. **Fix**: Apply the targeted architectural or code correction.
5. **Regression Test**: Add a permanent regression test asserting the fix.
6. **Re-execute**: Run the full suite to verify zero regressions.
7. **Document**: Update this registry with resolution status.

---

## 2. Standard Defect Template

```text
Bug ID: BUG-<PHASE>-<SEQ>
Title: <Concise, descriptive title>
Phase: <Phase 0 / Phase 1 / Phase 2 / E2E>
Module: <Auth / Tenants / Database / Middleware / Frontend>
Severity: <Critical / Major / Moderate / Minor>
Priority: <P0 / P1 / P2 / P3>
Environment: Windows 11 / Docker Compose (PostgreSQL 16 + Redis 7)
Preconditions: <Required system state>
Steps to Reproduce:
  1. ...
  2. ...
Expected Result: <What should occur according to specification>
Actual Result: <Observed failure or error message>
Root Cause: <Technical explanation of why failure occurred>
Fix Applied: <Code changes implemented>
Regression Test: <Test function or file verifying resolution>
Verification Result: <PASS / FAIL>
Status: <OPEN / IN_PROGRESS / RESOLVED / CLOSED>
```

---

## 3. Active & Resolved Defect Registry

### BUG-P1-001: Missing UUID Import in Auth Handler
* **Phase**: Phase 1 | **Module**: Auth | **Severity**: Critical | **Priority**: P0
* **Preconditions**: Adding token refresh and logout routes to `internal/auth/handler.go`.
* **Steps to Reproduce**: Run `go test ./internal/auth/...`
* **Expected Result**: Clean compilation.
* **Actual Result**: `undefined: uuid.UUID` compile failure on line 136.
* **Root Cause**: `github.com/google/uuid` package was used in handler signature but omitted from file imports.
* **Fix Applied**: Added `"github.com/google/uuid"` to imports in `backend/internal/auth/handler.go`.
* **Regression Test**: `go test -v ./internal/auth/...`
* **Verification Result**: **PASS**
* **Status**: **RESOLVED**

---

### BUG-P0-002: Windows Sub-Millisecond Redis Ping Latency False Failure
* **Phase**: Phase 0 | **Module**: Redis Integration | **Severity**: Moderate | **Priority**: P2
* **Preconditions**: Running on Windows loopback socket.
* **Steps to Reproduce**: Run `TestRedis_ConnectionAndPing` in `tests/integration/redis_integration_test.go`.
* **Expected Result**: Ping latency checked safely.
* **Actual Result**: Test failed when Windows high-resolution timer rounded loopback response to `0s`, failing `latency <= 0`.
* **Root Cause**: Test assertion checked `latency <= 0` instead of `latency < 0`.
* **Fix Applied**: Updated check to `if latency < 0 { t.Fatalf("invalid latency") }`.
* **Regression Test**: `backend/tests/integration/redis_integration_test.go`
* **Verification Result**: **PASS** (Latency reported: 530.1µs)
* **Status**: **RESOLVED**

---

### BUG-P1-003: Duplicate Route Registration for `/api/v1/auth/logout`
* **Phase**: Phase 1 | **Module**: Server / Router | **Severity**: Major | **Priority**: P1
* **Preconditions**: Server router initialization in `backend/internal/server/router.go`.
* **Steps to Reproduce**: Wire `/api/v1/auth/logout` under both public and authenticated router groups.
* **Expected Result**: Single route definition enforcing authentication.
* **Actual Result**: Duplicate route panic or unauthenticated exposure.
* **Root Cause**: Route was duplicated during manual route refactoring.
* **Fix Applied**: Confined `POST /api/v1/auth/logout` solely to the authenticated router group.
* **Regression Test**: `TestAuthAPI_Logout_Revocation`
* **Verification Result**: **PASS**
* **Status**: **RESOLVED**
