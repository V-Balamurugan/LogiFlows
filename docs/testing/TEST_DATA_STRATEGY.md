# LogiFlows — Test Data Management Strategy

**Document Reference**: `docs/testing/TEST_DATA_STRATEGY.md`  
**Version**: 1.0.0  
**Status**: APPROVED  

---

## 1. Objectives & Principles

The Test Data Strategy ensures that all automated and manual tests executed across LogiFlows remain **isolated, repeatable, deterministic, and safe**.

### Core Tenets
1. **Zero Production Contamination**: Tests must never execute against production databases or services.
2. **Deterministic Uniqueness**: All dynamic fixtures must utilize UUID-based namespaces to eliminate race conditions during concurrent test executions.
3. **No Secrets in Repositories**: Test configurations must utilize explicit test credentials (`postgres`/`postgres`) and dummy JWT secrets without exposing production API keys.
4. **Clean Slate & Teardown**: Test executions must clean up or isolate created state so that subsequent runs do not fail due to duplicate constraint violations.
5. **Privacy Compliance**: No real personally identifiable information (PII), customer data, or live carrier credentials may be used in test fixtures.

---

## 2. Test Fixture Generation Strategy

### 2.1 Dynamic Email & Identifier Generation
All test users are generated with a unique timestamp or UUID suffix:
```go
func GenerateTestEmail(prefix string) string {
    return fmt.Sprintf("%s-%s@logiflows.test", prefix, uuid.New().String()[:8])
}
```
* **Domain Standard**: Always use `.test` or `.example` pseudo-TLDs (e.g., `@logiflows.test`), adhering to RFC 2606.
* **Predictable Prefixes**:
  * `owner-<uuid>@logiflows.test`: Tenant company owners.
  * `operator-<uuid>@logiflows.test`: Tenant branch/fleet operators.
  * `viewer-<uuid>@logiflows.test`: Read-only viewers.
  * `platform-admin-<uuid>@logiflows.test`: Global platform administrators.

### 2.2 Multi-Tenant Organization Fixtures
To guarantee strict tenant isolation testing:
* Every tenant isolation test suite spins up at least two distinct organizations:
  * **Tenant Alpha**: `slug: "alpha-logistics-<uuid>"`
  * **Tenant Beta**: `slug: "beta-express-<uuid>"`
* Resources created within Tenant Alpha are strictly bound to `tenant_alpha_id`.
* Verification queries ensure Tenant Beta users receive `403 Forbidden` (`CROSS_TENANT_ACCESS_DENIED`) when attempting to reference Tenant Alpha IDs.

---

## 3. Database Isolation Models

### 3.1 Migration Validation Database
* Integration tests verify migrations using Goose on the local PostgreSQL container `logiflows_dev`.
* `TestMigrations_RollbackAndReapply` performs:
  1. `RunDown`: Rolls back the latest migration (`00003_add_refresh_tokens_and_user_verification.sql`).
  2. Verifies table schema status.
  3. `RunUp`: Re-applies the migration to restore the complete schema to version 3.

### 3.2 Relational Integrity & Cascades
* Database schema employs foreign keys with `ON DELETE CASCADE` between `tenants` and `tenant_memberships`, and between `users` and `refresh_tokens`.
* Deleting a test tenant or user automatically purges associated child records, preventing orphaned data accumulation.

---

## 4. Security & Sensitive Data Prevention

| Category | Policy | Implementation / Check |
| :--- | :--- | :--- |
| **Passwords** | Never use real passwords in tests | Use standard test strings: `Pass@123456!`, `Secret#2026Test` |
| **Hashes** | Bcrypt cost in tests | Cost 12 is tested; fast mock or cost 4 may be considered for large unit test suites, but full cost 12 is verified in auth integration tests |
| **Tokens** | Mock / Ephemeral JWTs | Signed using test secret `test-jwt-secret-minimum-32-bytes-long!` |
| **Customer PII** | Synthetic dummy data | All names are generated from fictitious dictionaries (e.g., "Alice Tester", "Bob Operator") |
| **Audit Logs** | Immutable verification | Audit records verify metadata without storing passwords or tokens |

---

## 5. Teardown & Maintenance Protocol

1. **Docker Volume Reset**:
   When necessary to purge test artifacts and start with a completely pristine database:
   ```bash
   make dev-reset
   ```
   * Stops containers.
   * Purges PostgreSQL volume.
   * Restarts containers and runs initial extensions.
2. **Go Test Cleanup**:
   Go test functions utilize `t.Cleanup(func() { ... })` where temporary environment variables or files are created.
