# LogiFlows Phase 4 Test Cases Specification

## 1. Backend Test Cases

| Case ID | Category | Scenario | Expected Result | Status |
| :--- | :--- | :--- | :--- | :--- |
| **TC-P4-BE-001** | Migration | Apply migration 00008 | 8 tables, indexes, and partial unique constraint created | **VERIFIED** |
| **TC-P4-BE-002** | Parcels | Create parcel with origin and destination | Tracking number `PKG-YYYYMMDD-<tenant>-<seq>` generated, status `CREATED` | **VERIFIED** |
| **TC-P4-BE-003** | FSM | Transition `CREATED` to `RECEIVED_AT_ORIGIN_BRANCH` | Allowed for branch intake, audit record logged in `parcel_status_history` | **VERIFIED** |
| **TC-P4-BE-004** | FSM | Transition `CREATED` directly to `DELIVERED` | Rejected with HTTP 400 and `ErrInvalidTransition` | **VERIFIED** |
| **TC-P4-BE-005** | Delivery | Assign parcel to driver & vehicle | Task created in `ASSIGNED` status, parcel transitions to `OUT_FOR_DELIVERY` | **VERIFIED** |
| **TC-P4-BE-006** | Concurrency | Dispatch the same parcel to second driver | Partial unique index violates constraint, service returns HTTP 409 Conflict | **VERIFIED** |
| **TC-P4-BE-007** | Delivery | Record delivery attempt failure | Attempt logged in `delivery_attempts`, task set to `IN_PROGRESS` | **VERIFIED** |
| **TC-P4-BE-008** | Delivery | Submit Proof of Delivery | Proof saved in `delivery_proofs`, task set to `COMPLETED`, parcel set to `DELIVERED` | **VERIFIED** |
| **TC-P4-BE-009** | Transfer | Create linehaul manifest with parcels | Manifest created in `PENDING` status with item count | **VERIFIED** |
| **TC-P4-BE-010** | Transfer | Dispatch manifest into linehaul transit | Status set to `DISPATCHED`, all parcels transition to `IN_TRANSIT` | **VERIFIED** |
| **TC-P4-BE-011** | Transfer | Receive manifest at destination hub | Status set to `RECEIVED`, all parcels transition to `RECEIVED_AT_TRANSFER_BRANCH` | **VERIFIED** |
| **TC-P4-BE-012** | Tracking | Public customer tracking request | Status and milestone history returned; sender/recipient PII, driver IDs omitted | **VERIFIED** |
| **TC-P4-BE-013** | Barcode | Verify valid QR/barcode code | Parcel details returned and custody scan logged | **VERIFIED** |
| **TC-P4-BE-014** | Barcode | Verify invalid/tampered barcode | HTTP 400 returned, scan rejected | **VERIFIED** |

---

## 2. Frontend Test Cases

| Case ID | Category | Scenario | Expected Result | Status |
| :--- | :--- | :--- | :--- | :--- |
| **TC-P4-FE-001** | Validation | Parcel booking form validation | Rejects 0 or negative weight, missing sender/recipient details | **VERIFIED** |
| **TC-P4-FE-002** | FSM | Status badge mapping | Badges render with correct semantic HSL/RGBA colors | **VERIFIED** |
| **TC-P4-FE-003** | Dispatch | Task assignment payload | Validates parcel ID, driver ID, and priority pill | **VERIFIED** |
| **TC-P4-FE-004** | POD | Proof of delivery form submission | Requires recipient name and proof type selection | **VERIFIED** |
| **TC-P4-FE-005** | Tracking | Customer tracking milestone stepper | Progress bar calculates milestone stage accurately without leaking PII | **VERIFIED** |
| **TC-P4-FE-006** | Transfers | Manifest item selector | Calculates selected parcel count and validates destination branch | **VERIFIED** |

---

## 3. Mobile Test Cases

| Case ID | Category | Scenario | Expected Result | Status |
| :--- | :--- | :--- | :--- | :--- |
| **TC-P4-MOB-001** | Model | `ParcelModel.fromJson` | Parses dates, weights, branch IDs, and address fields accurately | **VERIFIED** |
| **TC-P4-MOB-002** | Model | `DeliveryTaskModel.fromJson` | Parses priority, status, and driver metadata | **VERIFIED** |
| **TC-P4-MOB-003** | API | `getDeliveryTasks` | Makes authenticated HTTP GET and returns task list | **VERIFIED** |
| **TC-P4-MOB-004** | API | `recordDeliveryAttempt` | Sends POST payload with outcome and reason; returns attempt model | **VERIFIED** |
| **TC-P4-MOB-005** | API | `submitDeliveryProof` | Sends POST with proof type and recipient; returns success | **VERIFIED** |
| **TC-P4-MOB-006** | UI | `ParcelDeliveryScreen` tabs | Renders Tasks, Scan & Custody, and Transfers tabs | **VERIFIED** |
| **TC-P4-MOB-007** | UI | Task Card actions | Renders "Attempt" and "Complete (POD)" buttons for active tasks | **VERIFIED** |
| **TC-P4-MOB-008** | UI | Custody Scanner | Renders barcode input, triggers scan verification against backend | **VERIFIED** |
