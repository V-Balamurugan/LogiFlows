# LogiFlows API Conventions & Standards

## 1. Base URL & Versioning
All REST endpoints are explicitly versioned in the URI path:
```
/api/v1
```
Major breaking changes will introduce `/api/v2` without mutating `/api/v1`.

---

## 2. Request Headers

| Header | Description | Required | Example |
| :--- | :--- | :--- | :--- |
| `Content-Type` | MIME format of request body | Yes (for POST/PUT/PATCH) | `application/json` |
| `Accept` | Desired response format | Recommended | `application/json` |
| `X-Request-ID` | Unique tracing identifier | Optional (generated if omitted) | `550e8400-e29b-41d4-a716-446655440000` |

If the client sends `X-Request-ID`, LogiFlows preserves it. If omitted, the server generates a cryptographically secure UUIDv4/UUIDv7. The `X-Request-ID` is returned in the response header and embedded in structured logs and error payloads.

---

## 3. Standard Response Format

### 3.1 Success Response Envelope
All successful requests return a uniform JSON envelope:
```json
{
  "status": "ok",
  "service": "logiflows-api",
  "version": "v1",
  "timestamp": "2026-09-20T20:30:00Z",
  "data": { ... }
}
```

### 3.2 Error Response Envelope
All client and server errors return an RFC-compliant structured payload without exposing internal stack traces, database schema details, or credentials:
```json
{
  "error": {
    "code": "BAD_REQUEST",
    "message": "The payload provided is invalid or malformed",
    "request_id": "550e8400-e29b-41d4-a716-446655440000",
    "details": [
      {
        "field": "port",
        "issue": "must be an integer between 1 and 65535"
      }
    ]
  }
}
```

---

## 4. Standard HTTP Status Codes

| Status Code | Meaning | When to Use |
| :--- | :--- | :--- |
| `200 OK` | Success | Standard successful GET, PUT, PATCH request |
| `201 Created` | Resource Created | Successful POST creating a resource |
| `204 No Content`| Success, No Body | Successful DELETE request |
| `400 Bad Request`| Client Error | Malformed JSON or invalid parameter validation |
| `401 Unauthorized`| Missing/Invalid Auth | Token absent or invalid (future phases) |
| `403 Forbidden` | Access Denied | Authenticated user lacks permission |
| `404 Not Found` | Resource Not Found | URI or entity does not exist |
| `409 Conflict` | State Conflict | Duplicate entity or concurrency violation |
| `422 Unprocessable`| Validation Failure | Semantic validation errors on input fields |
| `500 Internal Error`| Server Error | Unexpected panic or uncaught internal error |
| `503 Unavailable`| Service Degraded | Readiness check failure (DB/Redis unreachable) |

---

## 5. Foundation Endpoints (Phase 0)

### 5.1 Liveness Probe
- **Path**: `GET /api/v1/health`
- **Purpose**: Evaluates if the HTTP process is running and able to handle incoming TCP traffic. Used by Kubernetes/Docker liveness checks.
- **Expected Status**: `200 OK`
- **Response**:
```json
{
  "status": "ok",
  "service": "logiflows-api",
  "version": "v1",
  "timestamp": "2026-09-20T20:30:00Z"
}
```

### 5.2 Readiness Probe
- **Path**: `GET /api/v1/readiness`
- **Purpose**: Evaluates whether all essential infrastructure dependencies (PostgreSQL with PostGIS, Redis) are reachable and healthy.
- **Expected Status**:
  - `200 OK` if all dependencies are healthy.
  - `503 Service Unavailable` if one or more critical dependencies fail.
- **Response**:
```json
{
  "status": "ready",
  "service": "logiflows-api",
  "version": "v1",
  "timestamp": "2026-09-20T20:30:00Z",
  "checks": {
    "database": {
      "status": "UP",
      "latency_ms": 1.2
    },
    "redis": {
      "status": "UP",
      "latency_ms": 0.8
    }
  }
}
```
