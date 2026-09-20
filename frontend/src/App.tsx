import React, { useState, useEffect } from 'react';
import { 
  Database, 
  Cpu, 
  Server, 
  RefreshCw, 
  Zap, 
  Boxes, 
  Sparkles,
  Terminal
} from 'lucide-react';

interface BackendReadiness {
  status: string;
  service: string;
  version: string;
  timestamp: string;
  checks?: {
    database?: { status: string; latency_ms?: number; error?: string };
    redis?: { status: string; latency_ms?: number; error?: string };
  };
}

interface AIReadiness {
  status: string;
  service: string;
  version: string;
  models_loaded?: Record<string, string>;
}

interface PredictionResult {
  parcel_id: string;
  delay_risk_score: number;
  risk_level: string;
  estimated_delay_minutes: number;
  recommended_action: string;
  confidence: number;
}

interface TelemetryLog {
  id: string;
  timestamp: string;
  target: string;
  status: number;
  latency: number;
  message: string;
}

export default function App() {
  const [backendData, setBackendData] = useState<BackendReadiness | null>(null);
  const [aiData, setAiData] = useState<AIReadiness | null>(null);
  const [backendLoading, setBackendLoading] = useState(false);
  const [aiLoading, setAiLoading] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [telemetry, setTelemetry] = useState<TelemetryLog[]>([]);

  // Prediction Form State
  const [parcelId, setParcelId] = useState('PKG-2026-9041');
  const [origin, setOrigin] = useState('North Distribution Hub');
  const [destination, setDestination] = useState('Central Metro Station');
  const [distanceKm, setDistanceKm] = useState(185);
  const [weather, setWeather] = useState('clear');
  const [traffic, setTraffic] = useState('normal');
  const [predicting, setPredicting] = useState(false);
  const [predictionResult, setPredictionResult] = useState<PredictionResult | null>(null);

  const fetchHealthProbes = async () => {
    // 1. Probe Go Backend
    setBackendLoading(true);
    const beStart = performance.now();
    try {
      const res = await fetch('http://localhost:8080/api/v1/readiness');
      const beLatency = Math.round(performance.now() - beStart);
      if (res.ok) {
        const data = await res.json();
        setBackendData(data);
        logTelemetry('Go Backend (:8080)', res.status, beLatency, 'Readiness probe passed');
      } else {
        setBackendData(null);
        logTelemetry('Go Backend (:8080)', res.status, beLatency, 'Service reported unready');
      }
    } catch (err: any) {
      logTelemetry('Go Backend (:8080)', 0, 0, err.message || 'Connection refused');
      setBackendData(null);
    } finally {
      setBackendLoading(false);
    }

    // 2. Probe AI Service
    setAiLoading(true);
    const aiStart = performance.now();
    try {
      const res = await fetch('http://localhost:8000/api/v1/readiness');
      const aiLatency = Math.round(performance.now() - aiStart);
      if (res.ok) {
        const data = await res.json();
        setAiData(data);
        logTelemetry('AI Service (:8000)', res.status, aiLatency, 'Models operational');
      } else {
        setAiData(null);
        logTelemetry('AI Service (:8000)', res.status, aiLatency, 'Degraded status');
      }
    } catch (err: any) {
      logTelemetry('AI Service (:8000)', 0, 0, err.message || 'Connection refused');
      setAiData(null);
    } finally {
      setAiLoading(false);
    }
  };

  const logTelemetry = (target: string, status: number, latency: number, message: string) => {
    setTelemetry(prev => [
      {
        id: Math.random().toString(36).substring(2, 9),
        timestamp: new Date().toLocaleTimeString(),
        target,
        status,
        latency,
        message,
      },
      ...prev.slice(0, 15) // Keep last 15 logs
    ]);
  };

  const runPrediction = async (e: React.FormEvent) => {
    e.preventDefault();
    setPredicting(true);
    try {
      const res = await fetch('http://localhost:8000/api/v1/predict/delay-risk', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          parcel_id: parcelId,
          origin_branch: origin,
          destination_branch: destination,
          distance_km: Number(distanceKm),
          weather_condition: weather,
          traffic_density: traffic,
          current_custody_type: 'in_transit'
        })
      });
      if (res.ok) {
        const data = await res.json();
        setPredictionResult(data);
        logTelemetry('AI Inference', res.status, 12, `Predicted score ${data.delay_risk_score} for ${parcelId}`);
      } else {
        alert('Prediction request returned an error.');
      }
    } catch (err: any) {
      alert('Failed to connect to AI Service on port 8000. Is it running?');
    } finally {
      setPredicting(false);
    }
  };

  useEffect(() => {
    fetchHealthProbes();
    if (!autoRefresh) return;
    const interval = setInterval(fetchHealthProbes, 5000);
    return () => clearInterval(interval);
  }, [autoRefresh]);

  const allOperational = backendData?.status === 'ready' && aiData?.status === 'ready';

  return (
    <div style={{ maxWidth: '1440px', margin: '0 auto', padding: '24px 20px' }}>
      
      {/* Top Navigation Bar */}
      <header style={{ 
        display: 'flex', 
        justifyContent: 'space-between', 
        alignItems: 'center', 
        marginBottom: '32px',
        paddingBottom: '20px',
        borderBottom: '1px solid var(--border-subtle)'
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '14px' }}>
          <div style={{ 
            width: '44px', 
            height: '44px', 
            borderRadius: '12px', 
            background: 'linear-gradient(135deg, #06b6d4, #3b82f6)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            boxShadow: '0 0 20px rgba(6, 182, 212, 0.4)'
          }}>
            <Boxes color="#ffffff" size={24} />
          </div>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
              <h1 style={{ fontSize: '24px', fontWeight: '800', letterSpacing: '-0.5px' }}>
                LogiFlows
              </h1>
              <span style={{ 
                fontSize: '11px', 
                fontWeight: '700', 
                padding: '2px 8px', 
                borderRadius: '6px', 
                background: 'rgba(56, 189, 248, 0.15)', 
                color: 'var(--accent-cyan)',
                border: '1px solid rgba(56, 189, 248, 0.3)'
              }}>
                OPERATIONS COMMAND CENTER
              </span>
            </div>
            <p style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
              Intelligent Logistics Operations & Telemetry Command Center
            </p>
          </div>
        </div>

        {/* Global Controls & Status */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
          {/* Status Badge */}
          <div style={{
            display: 'flex',
            alignItems: 'center',
            gap: '10px',
            padding: '8px 16px',
            borderRadius: '30px',
            background: allOperational ? 'rgba(16, 185, 129, 0.1)' : 'rgba(245, 158, 11, 0.1)',
            border: `1px solid ${allOperational ? 'rgba(16, 185, 129, 0.3)' : 'rgba(245, 158, 11, 0.3)'}`
          }}>
            <div style={{
              width: '10px',
              height: '10px',
              borderRadius: '50%',
              backgroundColor: allOperational ? 'var(--accent-emerald)' : 'var(--accent-amber)'
            }} className="beacon-pulse" />
            <span style={{ 
              fontSize: '13px', 
              fontWeight: '600', 
              color: allOperational ? 'var(--accent-emerald)' : 'var(--accent-amber)' 
            }}>
              {allOperational ? 'ALL SERVICES OPERATIONAL' : 'SYSTEM DEGRADED / PROBING'}
            </span>
          </div>

          {/* Auto Refresh Toggle */}
          <button 
            type="button"
            onClick={() => setAutoRefresh(!autoRefresh)}
            style={{
              padding: '8px 12px',
              borderRadius: '10px',
              background: autoRefresh ? 'rgba(16, 185, 129, 0.15)' : 'rgba(255, 255, 255, 0.05)',
              border: `1px solid ${autoRefresh ? 'rgba(16, 185, 129, 0.3)' : 'var(--border-subtle)'}`,
              color: autoRefresh ? 'var(--accent-emerald)' : 'var(--text-muted)',
              fontSize: '12px',
              cursor: 'pointer',
              fontWeight: '600'
            }}
          >
            Auto-Probe {autoRefresh ? 'ON' : 'OFF'}
          </button>

          {/* Refresh Button */}
          <button 
            type="button"
            onClick={fetchHealthProbes}
            disabled={backendLoading || aiLoading}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              padding: '8px 16px',
              borderRadius: '10px',
              background: 'rgba(255, 255, 255, 0.05)',
              border: '1px solid var(--border-subtle)',
              color: 'var(--text-primary)',
              cursor: 'pointer',
              fontSize: '13px',
              fontWeight: '500',
              transition: 'all 0.2s'
            }}
          >
            <RefreshCw size={15} className={(backendLoading || aiLoading) ? 'spin-slow' : ''} />
            Probe Network
          </button>
        </div>
      </header>

      {/* Main Grid Layout: Service Cards */}
      <section style={{ marginBottom: '32px' }}>
        <h2 style={{ fontSize: '16px', fontWeight: '700', color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: '16px' }}>
          Infrastructure & Service Mesh
        </h2>

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '20px' }}>
          
          {/* Card 1: Go Backend Engine */}
          <div className="glass-panel" style={{ padding: '20px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '16px' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                <div style={{ padding: '10px', borderRadius: '10px', background: 'rgba(59, 130, 246, 0.15)', color: 'var(--accent-blue)' }}>
                  <Server size={22} />
                </div>
                <div>
                  <h3 style={{ fontSize: '16px', fontWeight: '700' }}>Core Backend API</h3>
                  <span className="mono-text" style={{ fontSize: '12px', color: 'var(--text-muted)' }}>Go 1.27 • Gin • :8080</span>
                </div>
              </div>
              <span style={{
                padding: '4px 10px',
                borderRadius: '20px',
                fontSize: '12px',
                fontWeight: '700',
                background: backendData ? 'rgba(16, 185, 129, 0.15)' : 'rgba(244, 63, 94, 0.15)',
                color: backendData ? 'var(--accent-emerald)' : 'var(--accent-rose)'
              }}>
                {backendData ? 'ONLINE' : 'UNREACHABLE'}
              </span>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', fontSize: '13px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-secondary)' }}>
                <span>Liveness Endpoint:</span>
                <span className="mono-text" style={{ color: 'var(--accent-cyan)' }}>/api/v1/health</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-secondary)' }}>
                <span>Readiness Status:</span>
                <span className="mono-text" style={{ fontWeight: '600' }}>{backendData?.status?.toUpperCase() || 'DOWN'}</span>
              </div>
            </div>
          </div>

          {/* Card 2: PostgreSQL + PostGIS */}
          <div className="glass-panel" style={{ padding: '20px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '16px' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                <div style={{ padding: '10px', borderRadius: '10px', background: 'rgba(6, 182, 212, 0.15)', color: 'var(--accent-cyan)' }}>
                  <Database size={22} />
                </div>
                <div>
                  <h3 style={{ fontSize: '16px', fontWeight: '700' }}>PostgreSQL + PostGIS</h3>
                  <span className="mono-text" style={{ fontSize: '12px', color: 'var(--text-muted)' }}>PostGIS 3.4 • :5432</span>
                </div>
              </div>
              <span style={{
                padding: '4px 10px',
                borderRadius: '20px',
                fontSize: '12px',
                fontWeight: '700',
                background: backendData?.checks?.database?.status === 'UP' ? 'rgba(16, 185, 129, 0.15)' : 'rgba(244, 63, 94, 0.15)',
                color: backendData?.checks?.database?.status === 'UP' ? 'var(--accent-emerald)' : 'var(--accent-rose)'
              }}>
                {backendData?.checks?.database?.status || 'DOWN'}
              </span>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', fontSize: '13px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-secondary)' }}>
                <span>Query Latency:</span>
                <span className="mono-text" style={{ color: 'var(--accent-emerald)', fontWeight: '700' }}>
                  {backendData?.checks?.database?.latency_ms ? `${backendData.checks.database.latency_ms} ms` : 'N/A'}
                </span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-secondary)' }}>
                <span>Extensions:</span>
                <span className="mono-text">postgis, uuid-ossp</span>
              </div>
            </div>
          </div>

          {/* Card 3: Redis Cache */}
          <div className="glass-panel" style={{ padding: '20px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '16px' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                <div style={{ padding: '10px', borderRadius: '10px', background: 'rgba(244, 63, 94, 0.15)', color: 'var(--accent-rose)' }}>
                  <Zap size={22} />
                </div>
                <div>
                  <h3 style={{ fontSize: '16px', fontWeight: '700' }}>Redis Cache & Pub/Sub</h3>
                  <span className="mono-text" style={{ fontSize: '12px', color: 'var(--text-muted)' }}>Redis 7.0 • :6379</span>
                </div>
              </div>
              <span style={{
                padding: '4px 10px',
                borderRadius: '20px',
                fontSize: '12px',
                fontWeight: '700',
                background: backendData?.checks?.redis?.status === 'UP' ? 'rgba(16, 185, 129, 0.15)' : 'rgba(244, 63, 94, 0.15)',
                color: backendData?.checks?.redis?.status === 'UP' ? 'var(--accent-emerald)' : 'var(--accent-rose)'
              }}>
                {backendData?.checks?.redis?.status || 'DOWN'}
              </span>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', fontSize: '13px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-secondary)' }}>
                <span>Ping Latency:</span>
                <span className="mono-text" style={{ color: 'var(--accent-emerald)', fontWeight: '700' }}>
                  {backendData?.checks?.redis?.latency_ms ? `${backendData.checks.redis.latency_ms} ms` : 'N/A'}
                </span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-secondary)' }}>
                <span>Role:</span>
                <span className="mono-text">In-Memory Pub/Sub</span>
              </div>
            </div>
          </div>

          {/* Card 4: AI Predictive Service */}
          <div className="glass-panel" style={{ padding: '20px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '16px' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                <div style={{ padding: '10px', borderRadius: '10px', background: 'rgba(139, 92, 246, 0.15)', color: 'var(--accent-purple)' }}>
                  <Cpu size={22} />
                </div>
                <div>
                  <h3 style={{ fontSize: '16px', fontWeight: '700' }}>AI Predictive Service</h3>
                  <span className="mono-text" style={{ fontSize: '12px', color: 'var(--text-muted)' }}>Python • FastAPI • :8000</span>
                </div>
              </div>
              <span style={{
                padding: '4px 10px',
                borderRadius: '20px',
                fontSize: '12px',
                fontWeight: '700',
                background: aiData ? 'rgba(16, 185, 129, 0.15)' : 'rgba(244, 63, 94, 0.15)',
                color: aiData ? 'var(--accent-emerald)' : 'var(--accent-rose)'
              }}>
                {aiData ? 'ONLINE' : 'UNREACHABLE'}
              </span>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', fontSize: '13px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-secondary)' }}>
                <span>Inference Engine:</span>
                <span className="mono-text" style={{ color: 'var(--accent-purple)' }}>
                  {aiData?.models_loaded?.delay_predictor || 'Unavailable'}
                </span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-secondary)' }}>
                <span>Endpoint:</span>
                <span className="mono-text">/api/v1/predict/delay-risk</span>
              </div>
            </div>
          </div>

        </div>
      </section>

      {/* Two-Column Workspace: AI Delay Prediction Simulator & Real-Time Telemetry */}
      <div style={{ display: 'grid', gridTemplateColumns: 'minmax(0, 1.4fr) minmax(0, 1fr)', gap: '24px' }}>
        
        {/* Left Column: Live AI Prediction Simulator */}
        <section className="glass-panel" style={{ padding: '24px' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '20px' }}>
            <Sparkles size={20} color="var(--accent-cyan)" />
            <h2 style={{ fontSize: '18px', fontWeight: '700' }}>Live AI Delay-Risk Predictor</h2>
          </div>
          <p style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '20px' }}>
            Simulate parcel transit conditions and trigger the FastAPI predictive intelligence microservice to compute delay risk scores.
          </p>

          <form onSubmit={runPrediction} style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px', marginBottom: '24px' }}>
            <div>
              <label style={{ display: 'block', fontSize: '12px', fontWeight: '600', color: 'var(--text-muted)', marginBottom: '6px' }}>
                PARCEL IDENTIFIER
              </label>
              <input 
                type="text" 
                value={parcelId} 
                onChange={e => setParcelId(e.target.value)} 
                style={{
                  width: '100%',
                  padding: '10px 14px',
                  borderRadius: '8px',
                  background: 'rgba(15, 23, 42, 0.8)',
                  border: '1px solid var(--border-subtle)',
                  color: 'white',
                  fontFamily: 'var(--font-mono)',
                  fontSize: '13px'
                }}
              />
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '12px', fontWeight: '600', color: 'var(--text-muted)', marginBottom: '6px' }}>
                TRANSIT DISTANCE (KM)
              </label>
              <input 
                type="number" 
                value={distanceKm} 
                onChange={e => setDistanceKm(Number(e.target.value))} 
                style={{
                  width: '100%',
                  padding: '10px 14px',
                  borderRadius: '8px',
                  background: 'rgba(15, 23, 42, 0.8)',
                  border: '1px solid var(--border-subtle)',
                  color: 'white',
                  fontFamily: 'var(--font-mono)',
                  fontSize: '13px'
                }}
              />
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '12px', fontWeight: '600', color: 'var(--text-muted)', marginBottom: '6px' }}>
                ORIGIN HUB / BRANCH
              </label>
              <input 
                type="text" 
                value={origin} 
                onChange={e => setOrigin(e.target.value)} 
                style={{
                  width: '100%',
                  padding: '10px 14px',
                  borderRadius: '8px',
                  background: 'rgba(15, 23, 42, 0.8)',
                  border: '1px solid var(--border-subtle)',
                  color: 'white',
                  fontSize: '13px'
                }}
              />
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '12px', fontWeight: '600', color: 'var(--text-muted)', marginBottom: '6px' }}>
                DESTINATION HUB / BRANCH
              </label>
              <input 
                type="text" 
                value={destination} 
                onChange={e => setDestination(e.target.value)} 
                style={{
                  width: '100%',
                  padding: '10px 14px',
                  borderRadius: '8px',
                  background: 'rgba(15, 23, 42, 0.8)',
                  border: '1px solid var(--border-subtle)',
                  color: 'white',
                  fontSize: '13px'
                }}
              />
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '12px', fontWeight: '600', color: 'var(--text-muted)', marginBottom: '6px' }}>
                WEATHER CONDITION
              </label>
              <select 
                value={weather} 
                onChange={e => setWeather(e.target.value)} 
                style={{
                  width: '100%',
                  padding: '10px 14px',
                  borderRadius: '8px',
                  background: 'rgba(15, 23, 42, 0.8)',
                  border: '1px solid var(--border-subtle)',
                  color: 'white',
                  fontSize: '13px'
                }}
              >
                <option value="clear">Clear Skies</option>
                <option value="rain">Heavy Rain</option>
                <option value="fog">Dense Fog</option>
                <option value="storm">Severe Storm</option>
              </select>
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '12px', fontWeight: '600', color: 'var(--text-muted)', marginBottom: '6px' }}>
                TRAFFIC DENSITY
              </label>
              <select 
                value={traffic} 
                onChange={e => setTraffic(e.target.value)} 
                style={{
                  width: '100%',
                  padding: '10px 14px',
                  borderRadius: '8px',
                  background: 'rgba(15, 23, 42, 0.8)',
                  border: '1px solid var(--border-subtle)',
                  color: 'white',
                  fontSize: '13px'
                }}
              >
                <option value="low">Low Traffic</option>
                <option value="normal">Normal Traffic</option>
                <option value="high">High Volume</option>
                <option value="congested">Gridlock / Congested</option>
              </select>
            </div>

            <div style={{ gridColumn: 'span 2' }}>
              <button 
                type="submit" 
                disabled={predicting}
                style={{
                  width: '100%',
                  padding: '12px',
                  borderRadius: '10px',
                  background: 'linear-gradient(135deg, #2563eb, #06b6d4)',
                  color: 'white',
                  border: 'none',
                  fontSize: '14px',
                  fontWeight: '600',
                  cursor: 'pointer',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: '8px',
                  boxShadow: '0 4px 15px rgba(37, 99, 235, 0.3)'
                }}
              >
                {predicting ? <RefreshCw size={16} className="spin-slow" /> : <Zap size={16} />}
                Run Predictive Intelligence Inference
              </button>
            </div>
          </form>

          {/* Prediction Result Display */}
          {predictionResult && (
            <div style={{ 
              padding: '18px', 
              borderRadius: '12px', 
              background: 'rgba(15, 23, 42, 0.9)', 
              border: '1px solid rgba(56, 189, 248, 0.3)'
            }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
                <span className="mono-text" style={{ fontSize: '13px', fontWeight: '700' }}>
                  {predictionResult.parcel_id}
                </span>
                <span style={{
                  padding: '4px 12px',
                  borderRadius: '20px',
                  fontSize: '12px',
                  fontWeight: '800',
                  background: predictionResult.risk_level === 'LOW' ? 'rgba(16, 185, 129, 0.2)' :
                              predictionResult.risk_level === 'MODERATE' ? 'rgba(59, 130, 246, 0.2)' :
                              predictionResult.risk_level === 'HIGH' ? 'rgba(245, 158, 11, 0.2)' : 'rgba(244, 63, 94, 0.2)',
                  color: predictionResult.risk_level === 'LOW' ? 'var(--accent-emerald)' :
                         predictionResult.risk_level === 'MODERATE' ? 'var(--accent-blue)' :
                         predictionResult.risk_level === 'HIGH' ? 'var(--accent-amber)' : 'var(--accent-rose)'
                }}>
                  {predictionResult.risk_level} RISK
                </span>
              </div>

              {/* Score Bar */}
              <div style={{ marginBottom: '14px' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '12px', color: 'var(--text-secondary)', marginBottom: '4px' }}>
                  <span>Delay Probability:</span>
                  <span className="mono-text" style={{ fontWeight: '700', color: 'white' }}>
                    {(predictionResult.delay_risk_score * 100).toFixed(0)}%
                  </span>
                </div>
                <div style={{ width: '100%', height: '8px', borderRadius: '4px', background: 'rgba(255, 255, 255, 0.1)', overflow: 'hidden' }}>
                  <div style={{
                    width: `${predictionResult.delay_risk_score * 100}%`,
                    height: '100%',
                    background: predictionResult.delay_risk_score > 0.6 ? 'var(--accent-rose)' :
                                predictionResult.delay_risk_score > 0.3 ? 'var(--accent-amber)' : 'var(--accent-emerald)',
                    borderRadius: '4px',
                    transition: 'width 0.5s ease'
                  }} />
                </div>
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px', fontSize: '13px' }}>
                <div style={{ padding: '10px', borderRadius: '8px', background: 'rgba(255, 255, 255, 0.03)' }}>
                  <span style={{ display: 'block', fontSize: '11px', color: 'var(--text-muted)' }}>ESTIMATED DELAY</span>
                  <span style={{ fontSize: '15px', fontWeight: '700', color: 'white' }}>
                    +{predictionResult.estimated_delay_minutes} Minutes
                  </span>
                </div>
                <div style={{ padding: '10px', borderRadius: '8px', background: 'rgba(255, 255, 255, 0.03)' }}>
                  <span style={{ display: 'block', fontSize: '11px', color: 'var(--text-muted)' }}>MODEL CONFIDENCE</span>
                  <span style={{ fontSize: '15px', fontWeight: '700', color: 'white' }}>
                    {(predictionResult.confidence * 100).toFixed(0)}%
                  </span>
                </div>
              </div>

              <p style={{ marginTop: '12px', fontSize: '12px', color: 'var(--accent-cyan)' }}>
                <strong>Recommendation:</strong> {predictionResult.recommended_action}
              </p>
            </div>
          )}
        </section>

        {/* Right Column: Live Telemetry Terminal Feed */}
        <section className="glass-panel" style={{ padding: '24px', display: 'flex', flexDirection: 'column' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
              <Terminal size={18} color="var(--accent-cyan)" />
              <h2 style={{ fontSize: '16px', fontWeight: '700' }}>Network Telemetry Log</h2>
            </div>
            <span style={{ fontSize: '11px', color: 'var(--text-muted)', fontFamily: 'var(--font-mono)' }}>
              LIVE PROBE FEED
            </span>
          </div>

          <div style={{
            flex: 1,
            maxHeight: '440px',
            overflowY: 'auto',
            background: 'rgba(10, 14, 23, 0.95)',
            border: '1px solid var(--border-subtle)',
            borderRadius: '10px',
            padding: '12px',
            fontFamily: 'var(--font-mono)',
            fontSize: '12px'
          }}>
            {telemetry.length === 0 ? (
              <div style={{ color: 'var(--text-muted)', padding: '20px', textAlign: 'center' }}>
                Waiting for network telemetry packets...
              </div>
            ) : (
              telemetry.map(item => (
                <div key={item.id} style={{ 
                  display: 'flex', 
                  alignItems: 'center', 
                  justifyContent: 'space-between',
                  padding: '6px 0', 
                  borderBottom: '1px solid rgba(255, 255, 255, 0.04)' 
                }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                    <span style={{ color: 'var(--text-muted)', fontSize: '11px' }}>{item.timestamp}</span>
                    <span style={{ color: item.status === 200 ? 'var(--accent-emerald)' : 'var(--accent-rose)', fontWeight: '700' }}>
                      {item.status || 'ERR'}
                    </span>
                    <span style={{ color: 'white' }}>{item.target}</span>
                  </div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                    {item.latency > 0 && (
                      <span style={{ color: 'var(--accent-cyan)', fontSize: '11px' }}>{item.latency}ms</span>
                    )}
                    <span style={{ color: 'var(--text-secondary)', fontSize: '11px' }}>{item.message}</span>
                  </div>
                </div>
              ))
            )}
          </div>
        </section>

      </div>

    </div>
  );
}
