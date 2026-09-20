# LogiFlows AI Predictive Intelligence Service

> Real-time delivery delay risk prediction, route anomaly assessment, and dynamic ETA estimation microservice.

---

## Technology Stack
- **Framework**: Python 3.12+ / FastAPI
- **Server Engine**: Uvicorn ASGI Server
- **Data Validation**: Pydantic v2
- **Port**: `8000`

---

## Endpoints

| Method | Path | Description |
| :--- | :--- | :--- |
| `GET` | `/health` / `/api/v1/health` | Liveness probe verifying microservice uptime |
| `GET` | `/api/v1/readiness` | Readiness probe confirming model availability |
| `POST`| `/api/v1/predict/delay-risk` | Calculates delivery delay risk score & recommended action |
| `GET` | `/docs` | Interactive OpenAPI / Swagger UI documentation |

---

## Running Locally

```bash
# 1. Activate virtual environment
# Windows:
.\.venv\Scripts\activate
# Linux/macOS:
source .venv/bin/activate

# 2. Run with Uvicorn
uvicorn app.main:app --host 0.0.0.0 --port 8000 --reload
```

---

## Running Tests
```bash
python -m unittest discover -s tests -p "test_*.py"
```
