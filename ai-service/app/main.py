from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from datetime import datetime

from .schemas import (
    HealthResponse,
    ReadinessResponse,
    DelayPredictionRequest,
    DelayPredictionResponse
)
from .services.predictor import DelayPredictor

app = FastAPI(
    title="LogiFlows AI Predictive Intelligence Service",
    description="Microservice providing real-time delivery delay risk prediction and dynamic ETA intelligence.",
    version="1.0.0"
)

# Enable CORS for Frontend communication
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

predictor = DelayPredictor()

@app.get("/", tags=["Root"])
def root():
    return {
        "service": "logiflows-ai-service",
        "status": "online",
        "docs": "/docs",
        "version": "v1"
    }

@app.get("/health", response_model=HealthResponse, tags=["Health"])
@app.get("/api/v1/health", response_model=HealthResponse, tags=["Health"])
def health_check():
    """Liveness probe: confirms the AI microservice process is running."""
    return HealthResponse()

@app.get("/api/v1/readiness", response_model=ReadinessResponse, tags=["Health"])
def readiness_check():
    """Readiness probe: confirms that predictive models and pipelines are ready for inference."""
    return ReadinessResponse(
        status="ready",
        models_loaded={
            "delay_predictor": predictor.model_version,
            "eta_model": "v0.1.0-initialized"
        }
    )

@app.post("/api/v1/predict/delay-risk", response_model=DelayPredictionResponse, tags=["Predictions"])
def predict_delay_risk(request: DelayPredictionRequest):
    """Calculates the delivery delay risk score for a parcel in transit."""
    try:
        return predictor.predict(request)
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Prediction execution failed: {str(e)}")

if __name__ == "__main__":
    import uvicorn
    uvicorn.run("app.main:app", host="0.0.0.0", port=8000, reload=True)
