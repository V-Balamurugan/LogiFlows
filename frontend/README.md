# LogiFlows Web Operations Center

> High-performance logistics operations dashboard for live system monitoring, fleet observability, and predictive intelligence.

---

## Technology Stack
- **Framework**: React 18 / Vite
- **Language**: TypeScript
- **Styling**: Vanilla CSS Design System with Glassmorphism & Cyber-Logistics Dark Theme
- **Port**: `5173`

---

## Features
- **Live Infrastructure Probing**: Real-time polling of Go Backend (`:8080`), PostgreSQL + PostGIS, and Redis cache.
- **AI Microservice Integration**: Probing Python FastAPI service (`:8000`) and live delay risk prediction inference.
- **Network Telemetry Log**: Terminal-style telemetry log recording live probe requests and latencies.

---

## Running Locally

```bash
# 1. Install dependencies
npm install

# 2. Start Vite development server
npm run dev

# 3. Build production bundle
npm run build
```
The application will launch on **http://localhost:5173**.
