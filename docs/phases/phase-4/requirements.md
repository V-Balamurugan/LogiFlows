# Phase 4 Requirements Specification: Customers and Parcel Booking

**Project:** LogiFlows — Intelligent End-to-End Logistics Coordination and Delivery Management System  
**Phase:** Phase 4 — Customers and Parcel Booking  
**Document Version:** 1.0.0  
**Status:** Approved for Implementation  

---

## 1. Objective & Scope

The objective of Phase 4 is to deliver a complete, robust, enterprise-grade vertical slice for **Customer Management and Parcel Booking**.

This enables logistics tenants, operators, and staff to:
1. Maintain structured customer profiles (Shippers, Consignees, Corporate Accounts, Merchants, Individuals) with contact details, address book, customer codes, and credit tiers.
2. Create parcel bookings seamlessly linked to registered customer profiles or walk-in clients.
3. Automatically generate non-colliding, tamper-evident tracking numbers and unique QR code payloads.
4. Calculate dynamic pricing based on parcel weight, volume dimensions, service level, and declared value.
5. Enforce strict multi-tenant isolation, role-based authorization (RBAC), and transactional integrity.

---

## 2. User Stories

### US-P4-001: Customer Profile Management
**As an** authorized staff member (Tenant Admin, Operator, or Intake Clerk),  
**I want to** create and maintain customer profiles with contact details, billing tier, and physical address,  
**So that** sender and receiver information is organized, reusable, and searchable across booking operations.

### US-P4-002: Parcel Booking Creation
**As an** authorized user,  
**I want to** book a parcel by selecting origin/destination hubs, entering/selecting sender & receiver details, dimensions, weight, and service type,  
**So that** the consignment is registered into the logistics pipeline with an initial status of `CREATED` or `BOOKED`.

### US-P4-003: Unique Tracking Number & Barcode Generation
**As an** authorized user,  
**I want** each parcel booking to atomically generate a unique, non-colliding tracking identifier (`PKG-YYYYMMDD-XXXX`) and signed QR code payload,  
**So that** the shipment can be identified, tracked, and scanned across every custody node.

### US-P4-004: Parcel Detail Inspection & Audit Timeline
**As an** authorized user or customer,  
**I want to** retrieve full parcel booking details, including origin, destination, assigned branch, weight, service type, and initial status history,  
**So that** I can review booking accuracy and verify custody.

### US-P4-005: Input Validation & Clear Error Rejection
**As an** authorized user,  
**I want** invalid parcel booking requests (e.g., negative weight, missing phone, identical origin and destination, invalid service type) to be rejected with explicit HTTP 400 Bad Request error messages,  
**So that** data corruption and faulty manifests are prevented at the API boundary.

### US-P4-006: Tenant & Role Authorization
**As an** organization administrator,  
**I want** tenant isolation and role permissions (`TENANT_ADMIN`, `TENANT_OPERATOR`, `EMPLOYEE`) to be strictly enforced on customer and booking resources,  
**So that** unauthorized users from other organizations cannot read, tamper with, or create records under our tenant account.

---

## 3. Acceptance Criteria

### 3.1 Customer Management Acceptance Criteria
- **AC-CUST-01 (Creation):** Creating a customer requires `name`, `phone`, `city`, `state`, and `postal_code`. `customer_code` must be auto-generated (`CUST-YYYYMMDD-XXXX`) if omitted.
- **AC-CUST-02 (Duplicate Prevention):** `customer_code` must be unique per tenant organization. Duplicate submissions return HTTP 409 Conflict.
- **AC-CUST-03 (Customer Types):** Supports `INDIVIDUAL`, `BUSINESS`, `ENTERPRISE`, and `MERCHANT`.
- **AC-CUST-04 (Status Lifecycle):** Supports `ACTIVE`, `INACTIVE`, and `SUSPENDED`.
- **AC-CUST-05 (Search & Filter):** Customers can be searched via text query (name, code, phone, email) with pagination.
- **AC-CUST-06 (Customer Parcels):** An endpoint `GET /tenants/:id/customers/:customer_id/parcels` returns all shipments where the customer is either the sender or receiver.

### 3.2 Parcel Booking Acceptance Criteria
- **AC-BOOK-01 (Mandatory Fields):** Parcel booking requires sender details, receiver details, `origin_branch_id`, `destination_branch_id`, `weight_kg > 0`, and `dimensions_cm`.
- **AC-BOOK-02 (Customer Linking):** Optional `sender_customer_id` and `receiver_customer_id` can be linked. If supplied, sender/receiver details can be auto-populated.
- **AC-BOOK-03 (Branch Invariants):** `origin_branch_id` and `destination_branch_id` must belong to the active tenant, must exist, and cannot be identical.
- **AC-BOOK-04 (Service Types):** Permitted values are `STANDARD`, `EXPRESS`, `OVERNIGHT`, and `SAME_DAY`.
- **AC-BOOK-05 (Automated Pricing):** 
  - Standard Base: ₹50.00
  - Express Base: ₹120.00
  - Overnight Base: ₹200.00
  - Same Day Base: ₹350.00
  - Weight Surcharge: ₹20.00 per kg
  - Declared Value Insurance: 0.5% on values above ₹1,000.00
  - If a price is explicitly supplied in the request, it is honored; otherwise, the calculated price is assigned.
- **AC-BOOK-06 (Tracking Number Generation):** Auto-generates sequential, collision-free tracking number using PostgreSQL atomic sequence.
- **AC-BOOK-07 (Initial Audit Record):** Initial status is `CREATED`. A corresponding initial entry in `parcel_status_history` is recorded within the same database transaction.
- **AC-BOOK-08 (QR Code Payload):** Tamper-evident checksummed JSON containing tracking number, tenant ID, and parcel ID.

---

## 4. Security & Data Integrity Requirements

1. **Multi-Tenant Isolation:** All database queries must include `WHERE tenant_id = $x`. Attempting to access or link cross-tenant branches or customers must return HTTP 403 Forbidden.
2. **Atomic Transactions:** Parcel creation, initial status history recording, and customer sequence updates must execute inside a PostgreSQL database transaction (`tx.Begin` / `tx.Commit`).
3. **Data Sanitization:** Public tracking responses must omit sensitive customer phone numbers, emails, and exact apartment details (Zero-PII leakage).
4. **Parameterized SQL:** All queries must use parameterized placeholders (`$1, $2...`) via `pgx/v5` to prevent SQL injection vulnerabilities.
