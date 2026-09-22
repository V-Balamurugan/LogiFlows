# LogiFlows — Continuous Regression Testing Plan

**Document Reference**: `docs/testing/REGRESSION_TEST_PLAN.md`  
**Version**: 1.0.0  
**Status**: APPROVED  

---

## 1. Regression Testing Objective

LogiFlows follows an Agile Scrum vertical-slice architecture. **Whenever a new phase, feature, or database migration is introduced, all previously completed features must continue to function without degradation or regression.**

### Specific Invariants:
1. When **Phase 2 (Organization & Resources)** is developed:
   - Phase 0 foundation (Health, Readiness, Redis, Postgres, Logging) must remain 100% operational.
   - Phase 1 identity, authentication, JWT tokens, tenant isolation, and RBAC must remain 100% operational.
2. When **Phase 3 (Shipments & Tracking)** is developed:
   - Phase 0, Phase 1, and Phase 2 features must remain 100% operational.
3. When any **bug fix** is applied:
   - A regression test case must be added to the regression suite.
   - The full regression suite must pass before closing the issue.

---

## 2. Regression Suite Structure (`backend/tests/regression/`)

To prevent regression testing from depending on ad-hoc commands, an automated regression package is maintained:
```text
backend/tests/regression/
└── phase1_regression_test.go
```

### 2.1 Scope of Automated Regression Suite
* **Foundation Probes**: Liveness (200), Readiness (200), Database & Redis connectivity.
* **Authentication Contract**: Registration, Login, Token generation, Password hashing security.
* **Session Lifecycle**: Refresh token rotation, breach detection, logout revocation.
* **Tenant Security**: Multi-tenant isolation boundary, role permission matrix, cross-tenant 403 denial.
* **Observability**: Request ID propagation in headers, panic recovery.

---

## 3. Impact Analysis Protocol Before New Phase Merges

Before integrating any future Phase $N$:
1. **Impact Analysis Creation**: Create `docs/testing/IMPACT_ANALYSIS_PHASE_<N>.md` identifying:
   * Modified database tables and new migrations.
   * Modified API routes and middleware.
   * Changes to shared domain models.
   * Changes to authentication/authorization logic.
2. **Backward Compatibility Matrix**:
   * Verify existing API endpoints retain contract structures (`status`, `data`, `error`).
   * Verify existing JWT tokens remain valid and parsable.
   * Verify existing database records remain compatible with new schema constraints.
3. **Database Migration Safety**:
   * Verify new migrations run cleanly on fresh databases (`RunUp`).
   * Verify new migrations run safely on databases pre-populated with Phase 1 data.
   * Verify down migrations (`RunDown`) execute without data corruption.
4. **Execution Protocol**:
   ```bash
   # 1. Run unit tests with race detection
   cd backend && go test -v -race ./internal/...

   # 2. Run integration tests
   cd backend && go test -v ./tests/integration/...

   # 3. Run regression suite
   cd backend && go test -v ./tests/regression/...

   # 4. Run AI service tests
   cd ../ai-service && .venv/Scripts/python -m unittest discover tests

   # 5. Verify frontend build
   cd ../frontend && npm run lint && npm run build
   ```

---

## 4. Release Gate & Sign-Off Criteria

A vertical slice or release is approved for promotion only when:
* **Regression Pass Rate**: **100% PASS** (0 failures across all previously completed phases).
* **Security Defect Count**: **0 open P0 or P1 defects**.
* **Database Migration Idempotency**: Verified.
* **No Unverified Claims**: All test assertions documented with live execution timestamps and evidence.
