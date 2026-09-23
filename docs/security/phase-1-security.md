# LogiFlows Security Architecture & Threat Model — Phase 1

## 1. Cryptographic Standards & Algorithms

| Domain | Standard / Algorithm | Parameters / Configuration | Justification |
| :--- | :--- | :--- | :--- |
| **Password Hashing** | `bcrypt` (Blowfish-based adaptive hashing) | Computational cost factor = 12 | Resist offline dictionary & GPU-accelerated brute force attacks while maintaining sub-second login response latency. |
| **JWT Access Tokens** | `HMAC-SHA256` (`HS256`) | Minimum 32-character secret (`JWT_SECRET`) | Stateless, high-throughput verification across containerized backends. |
| **Refresh Tokens** | Cryptographically secure pseudo-random number generator (`crypto/rand`) | 32 bytes (256 bits of entropy) formatted as 64 hex characters | Unpredictable, unforgeable session handles immune to sequence prediction. |
| **Refresh Token Persistence** | `SHA-256` digest hashing | `VARCHAR(64)` indexed hash in PostgreSQL | Prevents exposure of raw session tokens in the event of database dumps or read replicas. |

---

## 2. Threat Model & Security Mitigations

| Threat ID | Threat Description | Attack Vector | LogiFlows Defense Implementation |
| :--- | :--- | :--- | :--- |
| **TH-01** | **User Enumeration / Timing Attacks** | Measuring response latency of `/api/v1/auth/login` to deduce whether an email exists in the database. | If user record does not exist, login handler executes a constant-time `ComparePassword` against a static bcrypt dummy hash before returning `401 Unauthorized`. |
| **TH-02** | **JWT Algorithm Confusion** | Attacker tampers with token header specifying `alg: "none"` or `alg: "RS256"`. | `ValidateAccessToken` explicitly casts and verifies `token.Method.(*jwt.SigningMethodHMAC)`. Any other algorithm causes immediate rejection. |
| **TH-03** | **Credential Leakage in Serialization** | Accidental serialization of `password_hash` in API responses or logs. | Domain model `users.User` tags `PasswordHash` with `json:"-"`. Automated unit test `TestUser_PasswordHash_ExcludedFromJSON` enforces this constraint. |
| **TH-04** | **Cross-Tenant Data Tampering** | User belonging to Tenant Alpha passes `tenant_id` of Tenant Beta in URL or payload to read/mutate private data. | `middleware.TenantContext` queries `tenant_memberships` for `(user_id, target_tenant_id)`. Rejects non-members with `403 Forbidden` (`CROSS_TENANT_ACCESS_DENIED`). Tested in `TestSecurity_CrossTenantAccess_Forbidden`. |
| **TH-05** | **Privilege Escalation** | `TENANT_OPERATOR` attempts to perform `TENANT_ADMIN` actions (e.g. adding team members). | `middleware.RequireRole` verifies active role against required permissions. Non-admins receive `403 Forbidden` (`INSUFFICIENT_PERMISSIONS`). Tested in `TestSecurity_RBAC_RolePermissionEnforcement`. |
| **TH-06** | **Refresh Token Theft & Reuse** | Attacker intercepts a refresh token and uses it to maintain persistent access. | **Single-use rotation + Breach Detection**: Consuming a refresh token immediately revokes it. If an already-revoked token is submitted, the server revokes all sessions for that account and logs `REFRESH_TOKEN_REUSE_DETECTED`. |
| **TH-07** | **SQL Injection** | Malicious characters injected into email or tenant slug queries. | 100% of database interactions utilize parameterized SQL queries (`$1`, `$2`, etc.) via `pgx/v5`. No dynamic string concatenation. |

---

## 3. Security Audit Logging Policy
- **Storage**: Immutable append-only PostgreSQL table `audit_logs`.
- **Captured Events**:
  - `USER_REGISTERED`
  - `USER_LOGIN`
  - `LOGIN_FAILED`
  - `TOKEN_REFRESHED`
  - `REFRESH_TOKEN_REUSE_DETECTED`
  - `USER_LOGOUT`
  - `TENANT_CREATED`
  - `TENANT_UPDATED`
  - `MEMBER_ADDED`
- **Sanitization Rule**: Secrets, plain passwords, JWT signatures, and raw refresh tokens are **strictly prohibited** from the `details` JSONB column.
- **Retention**: Production systems partition `audit_logs` by month with a 1-year retention policy for regulatory compliance.
