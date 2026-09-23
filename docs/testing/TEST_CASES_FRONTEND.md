# LogiFlows — Frontend Test Specifications

## 1. Overview
This document catalogs the test cases implemented for the **LogiFlows Web Frontend** (`frontend/`).
The frontend test harness executes natively using Node.js 24 (`node:test` and `--experimental-strip-types`), allowing blazing fast, dependency-free execution in continuous integration and local development.

- **Test Suite Location:** `frontend/test/frontend_unit.test.ts`
- **Runner Command:** `npm test` (`node --experimental-strip-types --test test/**/*.test.ts`)
- **Execution Engine:** Node 24 native test runner & TypeScript stripper
- **Complementary Verification:** `npm run lint` (`oxlint`), `npm run build` (`vite build`)

---

## 2. Test Case Catalog

| Test ID | Category | Name | Description | Status |
|---|---|---|---|---|
| **TC-FE-001** | Auth Storage | Token Storage Persistence | Verifies access token correctly persists into browser `localStorage`. | **PASS** |
| **TC-FE-002** | Auth Storage | Missing Token Fallback | Verifies unauthenticated state returns `null` when no token exists. | **PASS** |
| **TC-FE-003** | Auth Storage | Token Revocation / Logout | Verifies `localStorage.removeItem()` completely clears stored credentials. | **PASS** |
| **TC-FE-004** | Form Validation | Valid Email Validation | Verifies standard and subaddressed emails (`user@domain.com`, `user+tag@domain.co.in`) pass validation. | **PASS** |
| **TC-FE-005** | Form Validation | Malformed Email Rejection | Verifies invalid formats (`not-an-email`, `user@`, `@domain.com`, empty) are rejected. | **PASS** |
| **TC-FE-006** | Form Validation | Strong Password Acceptance | Verifies passwords satisfying the 5-rule complexity criteria (>=8 chars, upper, lower, number, special) pass. | **PASS** |
| **TC-FE-007** | Form Validation | Weak Password Rejection | Verifies passwords failing length (<8) or missing symbols/uppercase are rejected with appropriate error flags. | **PASS** |
| **TC-FE-008** | API Client | API Envelope Success Unpacking | Verifies standard API envelope `{ success: true, data: { ... } }` correctly extracts typed payload. | **PASS** |
| **TC-FE-009** | API Client | API Envelope Error Handling | Verifies error response `{ success: false, error: { code, message } }` correctly parses and propagates error details. | **PASS** |

---

## 3. Execution Commands & Results

### Unit Test Execution
```bash
$ npm test
✔ LogiFlows Frontend Auth & Storage Suite > should store and retrieve auth token from localStorage (1.42ms)
✔ LogiFlows Frontend Auth & Storage Suite > should return null if token is not present (0.17ms)
✔ LogiFlows Frontend Auth & Storage Suite > should clear auth token on logout (0.19ms)
✔ LogiFlows Input Validation Suite > should validate email formats correctly (0.28ms)
✔ LogiFlows Input Validation Suite > should reject invalid email formats (0.26ms)
✔ LogiFlows Input Validation Suite > should accept passwords meeting 5-rule complexity (0.34ms)
✔ LogiFlows Input Validation Suite > should reject passwords failing complexity rules (0.28ms)
✔ LogiFlows API Envelope & Data Transformations > should unpack standard success envelope payload (0.28ms)
✔ LogiFlows API Envelope & Data Transformations > should handle API error responses gracefully (0.19ms)
ℹ tests 9 | suites 3 | pass 9 | fail 0 | cancelled 0 | skipped 0 | todo 0
```

### Static Analysis & Production Build
```bash
$ npm run lint
Oxlint: 0 errors, 0 warnings

$ npm run build
vite v6.2.0 building for production...
✓ 1883 modules transformed.
dist/index.html                   0.56 kB │ gzip:  0.34 kB
dist/assets/index-D1o6kI0O.css   33.27 kB │ gzip:  6.48 kB
dist/assets/index-BOo9l957.js   366.19 kB │ gzip: 98.41 kB
✓ built in 1.99s
```
