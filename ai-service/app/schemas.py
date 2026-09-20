from typing import Optional, List, Dict, Any
from pydantic import BaseModel, Field
from datetime import datetime, timezone

class HealthResponse(BaseModel):
    status: str = "ok"
    service: str = "logiflows-ai-service"
    version: str = "v1"
    timestamp: datetime = Field(default_factory=lambda: datetime.now(timezone.utc))

class ReadinessResponse(BaseModel):
    status: str = "ready"
    service: str = "logiflows-ai-service"
    version: str = "v1"
    timestamp: datetime = Field(default_factory=lambda: datetime.now(timezone.utc))
    models_loaded: Dict[str, str] = Field(default_factory=dict)

class DelayPredictionRequest(BaseModel):
    parcel_id: str
    origin_branch: str
    destination_branch: str
    weather_condition: Optional[str] = "clear"  # clear, rain, storm, fog
    traffic_density: Optional[str] = "normal"   # low, normal, high, congested
    distance_km: float = Field(..., gt=0)
    current_custody_type: str = "branch"        # branch, in_transit, hub, driver

class DelayPredictionResponse(BaseModel):
    parcel_id: str
    delay_risk_score: float = Field(..., ge=0.0, le=1.0) # 0.0 (no risk) to 1.0 (imminent delay)
    risk_level: str  # LOW, MODERATE, HIGH, CRITICAL
    estimated_delay_minutes: int
    recommended_action: str
    confidence: float
