# LogiFlows Delivery Workflow & Branch Custody

## 1. End-to-End Operational Lifecycle

The LogiFlows delivery workflow spans four distinct operational stages:
1. **Intake & Sorting (Hub Operations)**
2. **Inter-Branch Linehaul (Transfer)**
3. **Last-Mile Dispatch (Delivery Task)**
4. **Proof of Delivery & Handover (POD)**

```mermaid
sequenceDiagram
    autonumber
    actor Sender
    actor Dispatcher
    actor Driver
    actor Recipient
    participant API as LogiFlows API
    participant DB as PostgreSQL Database

    Sender->>API: 1. Create Parcel (Sender, Receiver, Dimensions)
    API->>DB: 2. Generate Tracking PKG-..., Store Parcel
    API-->>Sender: 3. Return Parcel Details & QR Payload
    
    Dispatcher->>API: 4. Check-in Parcel at Origin Branch
    API->>DB: 5. Transition to RECEIVED_AT_ORIGIN_BRANCH
    
    Dispatcher->>API: 6. Create Delivery Task (Assign Driver & Vehicle)
    API->>DB: 7. Validate Driver Availability, Insert Task (409 on Conflict)
    
    Driver->>API: 8. Get Assigned Tasks (/deliveries/my-tasks)
    Driver->>API: 9. Start Delivery (Transition to OUT_FOR_DELIVERY)
    
    alt Delivery Successful
        Driver->>Recipient: 10. Handover Package & Capture POD (Signature/OTP)
        Driver->>API: 11. Submit Proof of Delivery
        API->>DB: 12. Validate POD, Transition to DELIVERED, Complete Task
    else Recipient Unavailable
        Driver->>API: 13. Record Delivery Attempt (CUSTOMER_UNAVAILABLE)
        API->>DB: 14. Log Attempt, Transition to DELIVERY_ATTEMPTED
    end
```

---

## 2. Concurrency Conflict Prevention
- **Race Condition Scenario:** Two dispatchers simultaneously assign the same parcel to two different drivers.
- **Resolution Strategy:**
  1. Transactional isolation with check on parcel eligibility.
  2. Database-level partial unique index:
     `CREATE UNIQUE INDEX idx_active_parcel_delivery_task ON delivery_tasks (tenant_id, parcel_id) WHERE status IN ('ASSIGNED', 'IN_PROGRESS');`
  3. When duplicate assignment occurs, PostgreSQL throws error `23505 (unique_violation)`.
  4. Repository maps this directly to `ErrParcelAlreadyAssigned`, returning `HTTP 409 Conflict`.

---

## 3. Branch Transfer & Custody Invariants
- A transfer can only be created between two distinct branches of the same tenant.
- A parcel cannot be associated with multiple active transfers simultaneously.
- When a destination branch operator scans a parcel during receiving:
  - If the parcel is marked on the manifest, its `current_branch_id` is updated to the destination branch.
  - Its status transitions from `IN_TRANSIT` to `RECEIVED_AT_TRANSFER_BRANCH`.
  - Re-scanning an already received parcel returns an idempotent success confirmation, preventing duplicate receipt logging.
