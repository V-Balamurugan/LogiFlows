# Phase 4 Requirements: Parcel & Delivery Lifecycle

## 1. Objective
Phase 4 implements the end-to-end parcel intake, state-machine tracking, inter-branch custody transfers, last-mile delivery task management, delivery proof validation, and customer-facing shipment transparency for LogiFlows.

---

## 2. Functional Requirements

### 2.1 Parcel Management (FR-PARCEL)
- **FR-PARCEL-001:** System must allow authenticated Tenant Admins and Tenant Operators to register parcels within their tenant organization.
- **FR-PARCEL-002:** System must atomically generate a unique, non-colliding, human-readable parcel tracking code format: `PKG-YYYYMMDD-XXXX` (where XXXX is a monotonically increasing sequence per tenant).
- **FR-PARCEL-003:** System must enforce sender details: full name, phone number, address.
- **FR-PARCEL-004:** System must enforce recipient details: full name, phone number, delivery address, optional email.
- **FR-PARCEL-005:** System must validate package physical properties: weight in kilograms (> 0), dimensions in centimeters (L x W x H format), and declared service level (`STANDARD`, `EXPRESS`, `OVERNIGHT`, `SAME_DAY`).
- **FR-PARCEL-006:** System must validate that origin and destination branches belong to the active tenant organization.
- **FR-PARCEL-007:** System must support parcel cancellation only when parcel is in `CREATED` or `BOOKED` state.
- **FR-PARCEL-008:** System must support paginated parcel listing with filtering by status, origin/destination branch, date range, and text search across tracking number or customer names.

### 2.2 Parcel State Machine & Auditing (FR-STATE)
- **FR-STATE-001:** System must enforce 15 discrete parcel operational statuses:
  - `CREATED`: Initial intake draft.
  - `BOOKED`: Confirmed and scheduled.
  - `READY_FOR_PICKUP`: Awaiting driver pickup.
  - `PICKED_UP`: Picked up by courier.
  - `RECEIVED_AT_ORIGIN_BRANCH`: Checked into origin hub.
  - `IN_TRANSIT`: Moving between distribution hubs or on route.
  - `RECEIVED_AT_TRANSFER_BRANCH`: Checked into intermediate transfer hub.
  - `OUT_FOR_DELIVERY`: Dispatched with last-mile driver.
  - `DELIVERY_ATTEMPTED`: Attempted but uncompleted delivery.
  - `DELIVERED`: Successfully delivered with proof.
  - `DELIVERY_FAILED`: Terminal failed delivery.
  - `RETURN_INITIATED`: Returning to origin branch or sender.
  - `RETURNED`: Safely returned to sender or origin hub.
  - `CANCELLED`: Cancelled prior to transit.
  - `ON_HOLD`: Temporarily held for customs, address clarification, or inspection.
- **FR-STATE-002:** Disallowed state transitions must be strictly rejected with HTTP 400 Bad Request.
- **FR-STATE-003:** Every status transition must atomically record an audit history row (`parcel_status_history`) documenting previous status, new status, actor ID, branch ID, notes, and timestamp.

### 2.3 Delivery Task Dispatch & Assignment (FR-DELIVERY)
- **FR-DELIVERY-001:** System must allow dispatchers to create delivery tasks associating an available parcel with an active driver (`operational_role = 'DRIVER'` or `'DELIVERY_EXECUTIVE'`) and optional vehicle.
- **FR-DELIVERY-002:** System must reject assigning parcels that are not in an eligible state (`RECEIVED_AT_ORIGIN_BRANCH`, `RECEIVED_AT_TRANSFER_BRANCH`, `OUT_FOR_DELIVERY`, or `DELIVERY_ATTEMPTED`).
- **FR-DELIVERY-003:** System must reject assigning inactive, suspended, or off-duty drivers.
- **FR-DELIVERY-004:** System must enforce atomic concurrency isolation preventing duplicate active delivery tasks for the same parcel (database partial unique index returning HTTP 409 Conflict on collision).
- **FR-DELIVERY-005:** Delivery drivers must have access to their assigned delivery tasks and recipient delivery details via employee endpoints (`/deliveries/my-tasks`).
- **FR-DELIVERY-006:** System must allow drivers to record delivery attempts with outcome categorization (`CUSTOMER_UNAVAILABLE`, `INCORRECT_ADDRESS`, `REJECTED_BY_CUSTOMER`, `SECURITY_ACCESS_DENIED`, `OTHER`), notes, and GPS coordinates.

### 2.4 Inter-Branch Transfers & Custody Tracking (FR-TRANSFER)
- **FR-TRANSFER-001:** System must support multi-parcel batch transfers between branches with generated manifest identifier `TRF-YYYYMMDD-XXXX`.
- **FR-TRANSFER-002:** Source and destination branches must be validated as distinct branches within the same tenant organization.
- **FR-TRANSFER-003:** Transfer dispatch must atomically update transfer status to `IN_TRANSIT`, update contained parcels to `IN_TRANSIT`, and record custody handover events.
- **FR-TRANSFER-004:** Destination branch receiving must confirm parcel arrival, update parcels' `current_branch_id` to destination, and prevent duplicate receiving.

### 2.5 QR & Barcode Workflow (FR-QR)
- **FR-QR-001:** System must generate secure QR payloads for parcels containing tracking number, tenant ID, and verification checksum.
- **FR-QR-002:** QR payload must never encode passwords, user tokens, internal database secrets, or sensitive PII.
- **FR-QR-003:** Scanning a parcel barcode/QR must validate tracking authenticity and log a scan event with the scanning employee and branch location.

### 2.6 Proof of Delivery (FR-POD)
- **FR-POD-001:** Transition to `DELIVERED` requires valid proof of delivery submission (`RECIPIENT_SIGNATURE`, `OTP`, `PHOTO`, or `CONTACTLESS_DROP`).
- **FR-POD-002:** POD must record recipient name, relationship, signature confirmation note or OTP, timestamp, and optional GPS coordinates.
- **FR-POD-003:** Duplicate POD submissions must be rejected.

### 2.7 Public & Customer Tracking (FR-TRACK)
- **FR-TRACK-001:** Public tracking endpoint `GET /api/v1/tracking/:tracking_number` must be accessible without bearer authentication.
- **FR-TRACK-002:** Public response must be strictly sanitized: exposing only tracking number, service type, current status, milestone timestamps, and city-level origin/destination information.
- **FR-TRACK-003:** Public response must NEVER expose customer phone numbers, full street addresses, driver full names, or internal database UUIDs.

---

## 3. Non-Functional Requirements
- **NFR-SEC-001 (Multi-Tenant Isolation):** Zero cross-tenant data leakage. All database queries must enforce `tenant_id = $1`.
- **NFR-SEC-002 (IDOR Prevention):** Cross-tenant UUID references must return HTTP 404 or 403.
- **NFR-PERF-001 (Response Time):** 95th percentile API response time < 50ms for parcel operations.
- **NFR-DATA-001 (Integrity):** Foreign keys, check constraints, and partial indexes must enforce business invariants at the database level.
