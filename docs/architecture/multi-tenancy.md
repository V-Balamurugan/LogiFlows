# LogiFlows Architecture: Multi-Tenancy & Data Isolation

## 1. Multi-Tenancy Model
LogiFlows adopts a **Pooled / Shared-Database Logical Multi-Tenancy Architecture** with strict row-level relational isolation:
- All tenant companies share the PostgreSQL cluster and schema.
- Every tenant entity has a foreign key relationship to `tenants(id)`.
- All operational queries and mutation requests are filtered by the validated `tenant_id` extracted from the authenticated caller's context.

---

## 2. Onboarding Workflow & Initial Tenant Creation
When a new customer signs up on LogiFlows:
1. The user inputs their personal credentials (`full_name`, `email`, `password`) along with their business identity (`company_name`).
2. An atomic PostgreSQL transaction executes:
   - Inserts `users` record.
   - Generates a URL-safe unique `slug` (e.g., `quickcargo-logistics`) and inserts `tenants` record with `status = 'ACTIVE'`.
   - Inserts `tenant_memberships` associating user and tenant with `role = 'TENANT_ADMIN'` and `status = 'ACTIVE'`.
   - Records an immutable `audit_logs` record (`USER_REGISTERED`).
3. If any step fails (e.g. duplicate email, database connection dropped), the transaction rolls back completely.

---

## 3. Cross-Tenant Leakage Prevention

```mermaid
graph LR
    subgraph Tenant Alpha
        UserA[User A (Tenant Admin Alpha)]
        DataA[(Alpha Data & Memberships)]
    end

    subgraph Tenant Beta
        UserB[User B (Tenant Admin Beta)]
        DataB[(Beta Data & Memberships)]
    end

    UserA -. Attempts to access .-> DataB
    DataB -. Rejected with 403 .-x UserA
```

1. **Never Trust Client Input**: The server never trusts tenant IDs supplied in request payloads for authorization decisions.
2. **Context-Driven Filtering**: All tenant operations extract `tenant_id` strictly from the URL parameter or header validated by `middleware.TenantContext`.
3. **Automated Integration Verification**: `TestSecurity_CrossTenantAccess_Forbidden` continuously tests that an authenticated admin from Company Alpha is blocked from querying or mutating Company Beta's resources, receiving `403 Forbidden` (`CROSS_TENANT_ACCESS_DENIED`).
4. **Referential Integrity**: All tenant membership associations use `ON DELETE CASCADE` foreign keys to ensure no orphaned access rights persist if a tenant or user is expunged.
