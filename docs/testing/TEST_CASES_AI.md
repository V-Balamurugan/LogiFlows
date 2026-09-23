# LogiFlows — AI Service Test Specifications

## 1. Overview
This document catalogs the test cases implemented for the **LogiFlows AI Delay Prediction Service** (`ai-service/`).
The AI service is built using **FastAPI**, **Pydantic v2**, and **scikit-learn / rule-based heuristics** for ETA delay risk analysis.

- **Test Suite Location:** `ai-service/tests/test_ai_service.py`
- **Runner Command:** `python -m unittest discover tests` (via `.venv\Scripts\python`)
- **Execution Framework:** Python Standard Library `unittest` + FastAPI `TestClient`
- **Total Test Cases:** 12

---

## 2. Test Case Catalog

| Test ID | Category | Name | Description | Status |
|---|---|---|---|---|
| **TC-AI-001** | Infrastructure | Root Service Metadata | Verifies `GET /` returns service name, version (`0.1.0`), and `status: "healthy"`. | **PASS** |
| **TC-AI-002** | Infrastructure | Liveness Probe | Verifies `GET /health` returns HTTP 200 with `{ status: "healthy" }`. | **PASS** |
| **TC-AI-003** | Infrastructure | Readiness Probe | Verifies `GET /ready` returns HTTP 200 confirming model and rule engine readiness. | **PASS** |
| **TC-AI-004** | Inference | Normal Delay Prediction | Verifies `POST /predict/delay` under clear weather and light traffic returns low delay risk (`LOW`) and realistic ETA. | **PASS** |
| **TC-AI-005** | Inference | Severe Weather & Congestion | Verifies severe weather (storm) + heavy congestion triggers higher predicted delay and `HIGH` risk level. | **PASS** |
| **TC-AI-006** | Boundary Validation | Zero Distance Rejection | Verifies `distance_km: 0.0` is rejected with HTTP 422 Unprocessable Entity by Pydantic constraint (`gt=0`). | **PASS** |
| **TC-AI-007** | Boundary Validation | Negative Distance Rejection | Verifies `distance_km: -15.5` is rejected with HTTP 422 Unprocessable Entity. | **PASS** |
| **TC-AI-008** | Schema Validation | Missing Required Fields | Verifies omitting required parameters (`origin`, `destination`) returns 422 with structured field error descriptions. | **PASS** |
| **TC-AI-009** | Schema Validation | Empty Body Rejection | Verifies sending an empty JSON payload `{}` returns HTTP 422 with clear validation breakdown. | **PASS** |
| **TC-AI-010** | Business Logic | Confidence Interval Bounds | Verifies the output confidence score is bounded strictly within `[0.0, 1.0]`. | **PASS** |
| **TC-AI-011** | Business Logic | Risk Category Mapping | Verifies risk categorization outputs strictly valid enumerations: `LOW`, `MEDIUM`, or `HIGH`. | **PASS** |
| **TC-AI-012** | Robustness | Fallback Weather Handling | Verifies unrecognized weather conditions default safely to moderate baseline rather than crashing the inference engine. | **PASS** |

---

## 3. Execution Commands & Results

```bash
$ .venv\Scripts\python -m unittest discover tests -v
test_root_endpoint (test_ai_service.TestAIServiceHealth.test_root_endpoint) ... ok
test_health_liveness (test_ai_service.TestAIServiceHealth.test_health_liveness) ... ok
test_health_readiness (test_ai_service.TestAIServiceHealth.test_health_readiness) ... ok
test_predict_delay_normal_conditions (test_ai_service.TestAIDelayPrediction.test_predict_delay_normal_conditions) ... ok
test_predict_delay_severe_weather_and_traffic (test_ai_service.TestAIDelayPrediction.test_predict_delay_severe_weather_and_traffic) ... ok
test_predict_delay_zero_distance_rejected (test_ai_service.TestAIDelayPrediction.test_predict_delay_zero_distance_rejected) ... ok
test_predict_delay_negative_distance_rejected (test_ai_service.TestAIDelayPrediction.test_predict_delay_negative_distance_rejected) ... ok
test_predict_delay_missing_fields_rejected (test_ai_service.TestAIDelayPrediction.test_predict_delay_missing_fields_rejected) ... ok
test_predict_delay_empty_payload_rejected (test_ai_service.TestAIDelayPrediction.test_predict_delay_empty_payload_rejected) ... ok
test_confidence_score_bounded (test_ai_service.TestAIDelayPrediction.test_confidence_score_bounded) ... ok
test_risk_category_valid_enum (test_ai_service.TestAIDelayPrediction.test_risk_category_valid_enum) ... ok
test_fallback_weather_handling (test_ai_service.TestAIDelayPrediction.test_fallback_weather_handling) ... ok

----------------------------------------------------------------------
Ran 12 tests in 1.621s

OK
```
