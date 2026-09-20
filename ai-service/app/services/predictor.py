from ..schemas import DelayPredictionRequest, DelayPredictionResponse

class DelayPredictor:
    """Predictive intelligence engine for delivery delay estimation."""
    
    def __init__(self):
        self.model_version = "v0.1.0-heuristic-baseline"
        
    def predict(self, req: DelayPredictionRequest) -> DelayPredictionResponse:
        score = 0.1  # baseline low risk
        
        # Traffic weighting
        traffic_multipliers = {
            "low": 0.0,
            "normal": 0.1,
            "high": 0.35,
            "congested": 0.55
        }
        score += traffic_multipliers.get(req.traffic_density.lower(), 0.1)
        
        # Weather weighting
        weather_multipliers = {
            "clear": 0.0,
            "rain": 0.2,
            "storm": 0.45,
            "fog": 0.3
        }
        score += weather_multipliers.get(req.weather_condition.lower(), 0.0)
        
        # Long distance factor
        if req.distance_km > 300:
            score += 0.15
        elif req.distance_km > 100:
            score += 0.08
            
        # Clamp score between 0.05 and 0.98
        score = min(max(score, 0.05), 0.98)
        
        if score >= 0.75:
            risk_level = "CRITICAL"
            delay_mins = int(45 + (score * 60))
            action = "Reroute driver or notify destination hub immediately."
        elif score >= 0.50:
            risk_level = "HIGH"
            delay_mins = int(25 + (score * 40))
            action = "Monitor transit checkpoints; prepare secondary delivery window."
        elif score >= 0.25:
            risk_level = "MODERATE"
            delay_mins = int(10 + (score * 20))
            action = "Standard tracking active; no intervention needed."
        else:
            risk_level = "LOW"
            delay_mins = 0
            action = "On-time delivery expected."
            
        return DelayPredictionResponse(
            parcel_id=req.parcel_id,
            delay_risk_score=round(score, 2),
            risk_level=risk_level,
            estimated_delay_minutes=delay_mins,
            recommended_action=action,
            confidence=0.92
        )
