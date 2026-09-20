import unittest
from fastapi.testclient import TestClient
from app.main import app

class TestAIServiceHealth(unittest.TestCase):
    def setUp(self):
        self.client = TestClient(app)

    def test_liveness_probe(self):
        response = self.client.get("/api/v1/health")
        self.assertEqual(response.status_code, 200)
        data = response.json()
        self.assertEqual(data["status"], "ok")
        self.assertEqual(data["service"], "logiflows-ai-service")

    def test_readiness_probe(self):
        response = self.client.get("/api/v1/readiness")
        self.assertEqual(response.status_code, 200)
        data = response.json()
        self.assertEqual(data["status"], "ready")
        self.assertIn("delay_predictor", data["models_loaded"])

    def test_delay_prediction_endpoint(self):
        payload = {
            "parcel_id": "PKG-2026-9901",
            "origin_branch": "BRANCH-NORTH-01",
            "destination_branch": "BRANCH-SOUTH-04",
            "weather_condition": "rain",
            "traffic_density": "high",
            "distance_km": 145.5,
            "current_custody_type": "in_transit"
        }
        response = self.client.post("/api/v1/predict/delay-risk", json=payload)
        self.assertEqual(response.status_code, 200)
        data = response.json()
        self.assertEqual(data["parcel_id"], "PKG-2026-9901")
        self.assertGreater(data["delay_risk_score"], 0.0)
        self.assertIn(data["risk_level"], ["LOW", "MODERATE", "HIGH", "CRITICAL"])
        self.assertTrue(len(data["recommended_action"]) > 0)

if __name__ == "__main__":
    unittest.main()
