# Phase 4 API Contracts & OpenAPI Specifications

## 1. Overview
All Phase 4 API routes reside under `/api/v1` and follow standard LogiFlows JSON response envelope conventions:
- **Success:** `{ "status": "ok", "service": "logiflows-api", "version": "v1", "timestamp": "...", "data": ... }`
- **Error:** `{ "error": { "code": "...", "message": "...", "request_id": "...", "details": ... } }`

---

## 2. Public Tracking Endpoints

### 2.1 Public Parcel Tracking
- **Method:** `GET`
- **URL:** `/api/v1/tracking/:tracking_number`
- **Authentication:** Public (No Bearer token required)
- **Response (200 OK):**
```json
{
  "status": "ok",
  "data": {
    "tracking_number": "PKG-20260924-0001",
    "status": "OUT_FOR_DELIVERY",
    "service_type": "EXPRESS",
    "origin_city": "Chennai",
    "destination_city": "Bengaluru",
    "weight_kg": 2.50,
    "created_at": "2026-09-24T08:30:00Z",
    "timeline": [
      {
        "status": "CREATED",
        "description": "Shipment order booked",
        "timestamp": "2026-09-24T08:30:00Z"
      },
      {
        "status": "RECEIVED_AT_ORIGIN_BRANCH",
        "description": "Package arrived at origin sorting facility",
        "timestamp": "2026-09-24T10:15:00Z"
      },
      {
        "status": "OUT_FOR_DELIVERY",
        "description": "Package is out for delivery with local courier",
        "timestamp": "2026-09-24T14:00:00Z"
      }
    ]
  }
}
```

---

## 3. Parcel Management Endpoints

### 3.1 Create Parcel
- **Method:** `POST`
- **URL:** `/api/v1/tenants/:tenant_id/parcels`
- **Authorization:** `TENANT_ADMIN`, `TENANT_OPERATOR`
- **Request Body:**
```json
{
  "sender_name": "Acme Electronics",
  "sender_phone": "+91 98765 43210",
  "sender_email": "shipping@acme.com",
  "sender_address": "12 Mount Road, Guindy, Chennai",
  "receiver_name": "John Doe",
  "receiver_phone": "+91 91234 56789",
  "receiver_email": "john.doe@example.com",
  "receiver_address": "45 MG Road, Indiranagar, Bengaluru",
  "origin_branch_id": "8f3e5c9b-6481-4202-a72c-29ec97b5e401",
  "destination_branch_id": "7a2b1c3d-1111-2222-3333-444455556666",
  "weight_kg": 3.75,
  "dimensions_cm": "30x20x15",
  "service_type": "EXPRESS",
  "declared_value": 1500.00,
  "special_instructions": "Handle with care - fragile display"
}
```

### 3.2 List Parcels
- **Method:** `GET`
- **URL:** `/api/v1/tenants/:tenant_id/parcels?status=OUT_FOR_DELIVERY&search=John&limit=20&offset=0`
- **Authorization:** `TENANT_ADMIN`, `TENANT_OPERATOR`, `VIEWER`

### 3.3 Get Parcel Details
- **Method:** `GET`
- **URL:** `/api/v1/tenants/:tenant_id/parcels/:parcel_id`
- **Authorization:** `TENANT_ADMIN`, `TENANT_OPERATOR`, `VIEWER`, `EMPLOYEE`

### 3.4 Update Parcel Status
- **Method:** `PATCH`
- **URL:** `/api/v1/tenants/:tenant_id/parcels/:parcel_id/status`
- **Authorization:** `TENANT_ADMIN`, `TENANT_OPERATOR`, `EMPLOYEE`
- **Request Body:**
```json
{
  "status": "RECEIVED_AT_ORIGIN_BRANCH",
  "notes": "Intake scanned by operator at Chennai Central hub"
}
```

### 3.5 QR Code Payload & Scan Verification
- **Method:** `GET`
- **URL:** `/api/v1/tenants/:tenant_id/parcels/:parcel_id/qr`
- **Method:** `POST`
- **URL:** `/api/v1/tenants/:tenant_id/parcels/scan`
- **Request Body:** `{ "qr_payload": "..." }` or `{ "tracking_number": "PKG-20260924-0001" }`

---

## 4. Delivery Task Endpoints

### 4.1 Create Delivery Task
- **Method:** `POST`
- **URL:** `/api/v1/tenants/:tenant_id/deliveries`
- **Authorization:** `TENANT_ADMIN`, `TENANT_OPERATOR`
- **Request Body:**
```json
{
  "parcel_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
  "assigned_driver_id": "4a123456-7890-abcd-ef01-23456789abcd",
  "vehicle_id": "3b234567-8901-bcde-f012-3456789abcde",
  "priority": "HIGH",
  "notes": "Morning slot delivery"
}
```

### 4.2 Get Driver Assigned Tasks
- **Method:** `GET`
- **URL:** `/api/v1/tenants/:tenant_id/deliveries/my-tasks`
- **Authorization:** `EMPLOYEE`, `TENANT_ADMIN`, `TENANT_OPERATOR`

### 4.3 Record Delivery Attempt
- **Method:** `POST`
- **URL:** `/api/v1/tenants/:tenant_id/deliveries/:delivery_id/attempt`
- **Authorization:** `EMPLOYEE`, `TENANT_ADMIN`, `TENANT_OPERATOR`
- **Request Body:**
```json
{
  "outcome": "CUSTOMER_UNAVAILABLE",
  "notes": "Door locked, recipient did not answer call",
  "latitude": 12.9716,
  "longitude": 77.5946
}
```

### 4.4 Submit Proof of Delivery
- **Method:** `POST`
- **URL:** `/api/v1/tenants/:tenant_id/deliveries/:delivery_id/proof`
- **Authorization:** `EMPLOYEE`, `TENANT_ADMIN`, `TENANT_OPERATOR`
- **Request Body:**
```json
{
  "proof_type": "RECIPIENT_SIGNATURE",
  "recipient_name": "Jane Doe",
  "recipient_relationship": "SPOUSE",
  "signature_data": "data:image/svg+xml;base64,...",
  "notes": "Delivered to recipient at front desk",
  "latitude": 12.9716,
  "longitude": 77.5946
}
```

---

## 5. Branch Transfer Endpoints

### 5.1 Create Branch Transfer Manifest
- **Method:** `POST`
- **URL:** `/api/v1/tenants/:tenant_id/transfers`
- **Request Body:**
```json
{
  "source_branch_id": "8f3e5c9b-6481-4202-a72c-29ec97b5e401",
  "destination_branch_id": "7a2b1c3d-1111-2222-3333-444455556666",
  "driver_id": "4a123456-7890-abcd-ef01-23456789abcd",
  "vehicle_id": "3b234567-8901-bcde-f012-3456789abcde",
  "parcel_ids": ["9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"],
  "notes": "Overnight linehaul Chennai -> Bengaluru"
}
```

### 5.2 Dispatch & Receive Transfer
- **Method:** `POST`
- **URL:** `/api/v1/tenants/:tenant_id/transfers/:transfer_id/dispatch`
- **Method:** `POST`
- **URL:** `/api/v1/tenants/:tenant_id/transfers/:transfer_id/receive`
