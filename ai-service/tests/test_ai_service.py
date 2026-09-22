import unittest
from fastapi.testclient import TestClient
from app.main import app

class TestAIServiceComprehensive(unittest.TestCase):
    def setUp(self):
        self.client = TestClient(app)

    def test_root_endpoint(self):
        """Verify root metadata endpoint returns 200 and service identity."""
        response = self.client.get("/")
        self.assertEqual(response.status_code, 200)
        data = response.json()
        self.assertEqual(data["service"], "logiflows-ai-service")
        self.assertEqual(data["status"], "online")

    def test_liveness_probe(self):
        """Verify liveness probe returns status 'ok' on both routes."""
        for path in ["/health", "/api/v1/health"]:
            response = self.client.get(path)
            self.assertEqual(response.status_code, 200)
            data = response.json()
            self.assertEqual(data["status"], "ok")
            self.assertEqual(data["service"], "logiflows-ai-service")

    def test_readiness_probe(self):
        """Verify readiness probe confirms loaded ML models."""
        response = self.client.get("/api/v1/readiness")
        self.assertEqual(response.status_code, 200)
        data = response.json()
        self.assertEqual(data["status"], "ready")
        self.assertIn("delay_predictor", data["models_loaded"])
        self.assertIn("eta_model", data["models_loaded"])

    def test_delay_prediction_standard_payload(self):
        """Verify delay prediction with normal driving parameters."""
        payload = {
            "parcel_id": "PKG-2026-9901",
            "origin_branch": "BRANCH-NORTH-01",
            "destination_branch": "BRANCH-SOUTH-04",
            "weather_condition": "clear",
            "traffic_density": "normal",
            "distance_km": 45.0,
            "current_custody_type": "in_transit"
        }
        response = self.client.post("/api/v1/predict/delay-risk", json=payload)
        self.assertEqual(response.status_code, 200)
        data = response.json()
        self.assertEqual(data["parcel_id"], "PKG-2026-9901")
        self.assertGreaterEqual(data["delay_risk_score"], 0.0)
        self.assertLessEqual(data["delay_risk_score"], 1.0)
        self.assertIn(data["risk_level"], ["LOW", "MODERATE", "HIGH", "CRITICAL"])
        self.assertGreater(len(data["recommended_action"]), 0)

    def test_delay_prediction_storm_and_congestion(self):
        """Verify severe storm and congested conditions increase delay risk."""
        payload = {
            "parcel_id": "PKG-2026-CRIT-99",
            "origin_branch": "BRANCH-EAST-01",
            "destination_branch": "BRANCH-WEST-02",
            "weather_condition": "storm",
            "traffic_density": "congested",
            "distance_km": 180.0,
            "current_custody_type": "in_transit"
        }
        response = self.client.post("/api/v1/predict/delay-risk", json=payload)
        self.assertEqual(response.status_code, 200)
        data = response.json()
        self.assertIn(data["risk_level"], ["HIGH", "CRITICAL"])
        self.assertGreater(data["delay_risk_score"], 0.4)
        self.assertGreater(data["estimated_delay_minutes"], 0)

    def test_delay_prediction_negative_distance_rejected(self):
        """Verify distance <= 0 is rejected with 422 Unprocessable Entity."""
        payload = {
            "parcel_id": "PKG-NEG-DIST",
            "origin_branch": "BRANCH-A",
            "destination_branch": "BRANCH-B",
            "distance_km": -15.5,
            "current_custody_type": "branch"
        }
        response = self.client.post("/api/v1/predict/delay-risk", json=payload)
        self.assertEqual(response.status_code, 422)

    def test_delay_prediction_zero_distance_rejected(self):
        """Verify distance == 0 is rejected with 422 Unprocessable Entity."""
        payload = {
            "parcel_id": "PKG-ZERO-DIST",
            "origin_branch": "BRANCH-A",
            "destination_branch": "BRANCH-B",
            "distance_km": 0.0,
            "current_custody_type": "branch"
        }
        response = self.client.post("/api/v1/predict/delay-risk", json=payload)
        self.assertEqual(response.status_code, 422)

    def test_delay_prediction_missing_required_fields(self):
        """Verify missing mandatory fields return 422 Unprocessable Entity."""
        payload = {
            "distance_km": 25.0
            # missing parcel_id, origin_branch, destination_branch, current_custody_type
        }
        response = self.client.post("/api/v1/predict/delay-risk", json=payload)
        self.assertEqual(response.status_code, 422)

    def test_delay_prediction_empty_body(self):
        """Verify empty POST body returns 422."""
        response = self.client.post("/api/v1/predict/delay-risk", json={})
        self.assertEqual(response.status_code, 422)

if __name__ == "__main__":
    unittest.main()
