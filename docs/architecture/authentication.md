# LogiFlows Architecture: Authentication & Session Strategy

## 1. Overview
The LogiFlows Authentication subsystem provides secure, stateless identity verification across web dashboards, API integrations, and mobile custody applications. It combines short-lived signed JSON Web Tokens (JWT) with long-lived, opaque, cryptographically hashed refresh tokens.

---

## 2. Token Architecture & Dual-Token Strategy

```mermaid
sequenceDiagram
    autonumber
    actor User as Client (Web / Mobile)
    participant API as LogiFlows API Server
    participant DB as PostgreSQL 16
    participant Audit as Audit Logger

    User->>API: POST /api/v1/auth/login {email, password}
    API->>DB: Query user by LOWER(email)
    DB-->>API: User record (including password_hash)
    API->>API: Verify bcrypt hash (work factor 12)
    API->>API: Generate Access Token (JWT HMAC-SHA256, 24h)
    API->>API: Generate Opaque Refresh Token (256-bit cryptorand)
    API->>DB: Store SHA-256(refresh_token) in refresh_tokens table
    API->>Audit: Log USER_LOGIN (Success)
    API-->>User: 200 OK {token, expires_at, refresh_token, refresh_token_expires_at, user, tenants}

    Note over User,API: Subsequent Authenticated Requests
    User->>API: GET /api/v1/auth/me (Authorization: Bearer <token>)
    API->>API: Verify JWT signature & expiration
    API->>DB: Query user status (is_active)
    API-->>User: 200 OK {user profile, active tenants}

    Note over User,API: Token Refresh (Single-Use Rotation)
    User->>API: POST /api/v1/auth/refresh {refresh_token}
    API->>DB: Query refresh_tokens by SHA-256(refresh_token)
    alt Token Revoked (Breach Attempt)
        API->>DB: Revoke ALL refresh tokens for user
        API->>Audit: Log REFRESH_TOKEN_REUSE_DETECTED
        API-->>User: 401 Unauthorized (Breach detected)
    else Token Valid & Unexpired
        API->>DB: Mark old refresh token revoked_at = NOW()
        API->>API: Generate new Access Token + new Refresh Token
        API->>DB: Store new SHA-256(refresh_token)
        API->>Audit: Log TOKEN_REFRESHED
        API-->>User: 200 OK {new token, new refresh_token}
    end

    Note over User,API: Logout
    User->>API: POST /api/v1/auth/logout {refresh_token} (Bearer Token)
    API->>DB: Mark refresh token revoked_at = NOW()
    API->>Audit: Log USER_LOGOUT
    API-->>User: 200 OK {message: "Successfully logged out"}
```

---

## 3. JWT Access Token Specification
- **Algorithm**: `HMAC-SHA256` (`HS256`).
- **Signature Security**: The verification handler enforces `jwt.SigningMethodHMAC` explicitly to defeat algorithm confusion (`none` algorithm or RSA/HMAC substitution attacks).
- **Secret Constraints**: Enforced at server startup; minimum 32 characters (`JWT_SECRET`).
- **Claims Structure**:
  ```json
  {
    "sub": "b2f6b8b0-8f92-4f0e-bb69-52e6d6bb1b01",
    "email": "balamurugan@quickcargo.com",
    "is_platform_admin": false,
    "iss": "logiflows-api",
    "exp": 1790184000,
    "iat": 1789320000
  }
  ```
- **Lifespan**: Configurable via `JWT_ACCESS_EXPIRY` (default `24h`).

---

## 4. Opaque Refresh Token Lifecycle & Breach Detection
1. **Generation**: 32 cryptographically secure random bytes generated via `crypto/rand`, encoded to a 64-character hexadecimal string.
2. **Persistence**: The raw refresh token is returned to the client and never saved in plain text. Only its SHA-256 digest (`VARCHAR(64)`) is stored in `refresh_tokens`.
3. **Single-Use Rotation**: Every successful call to `/api/v1/auth/refresh` immediately revokes the submitted refresh token (`revoked_at = NOW()`) and issues a fresh token pair.
4. **Breach Detection / Token Reuse Prevention**:
   - If an already-revoked refresh token is presented, the system assumes a token compromise (e.g., token theft by a malicious actor).
   - The server immediately triggers `tokenRepo.RevokeAllForUser(ctx, rt.UserID)`, invalidating every active session for that account.
   - A critical security audit event (`REFRESH_TOKEN_REUSE_DETECTED`) is logged with client IP and User-Agent.
   - The request is terminated with `401 Unauthorized`.

---

## 5. Defense-in-Depth Protections
- **Password Security**: Bcrypt with computational cost factor 12. Password complexity requires uppercase, lowercase, digit, and special character.
- **Timing Attack Mitigation**: When login fails due to a non-existent email, a dummy bcrypt comparison is executed against a constant hash to equalize server response time and prevent user enumeration.
- **Zero Secret Leakage**: The domain entity `users.User` defines `PasswordHash string json:"-"` to guarantee that password hashes can never be serialized in JSON responses.
