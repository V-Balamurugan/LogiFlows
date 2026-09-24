<#
.SYNOPSIS
    LogiFlows - Run Entire Project Automation Script
.DESCRIPTION
    Launches all LogiFlows services:
    1. Infrastructure: PostgreSQL (PostGIS) & Redis via Docker Compose
    2. Backend API: Go 1.27 REST Server (Port 8080)
    3. AI Microservice: FastAPI Predictive Intelligence (Port 8000)
    4. Frontend Console: React 19 + Vite Web Dashboard (Port 5173)
    5. Mobile Application: Flutter Web Client (Port 8085)
.PARAMETER Mode
    "Windows" (default) - Opens each service in a dedicated named PowerShell window.
    "Background" - Starts all services as background jobs logging to .logs/ directory.
.PARAMETER SkipMobile
    Switch to skip launching the Flutter mobile application.
.PARAMETER OpenBrowser
    Switch to automatically launch the default browser to the web console.
#>

param(
    [ValidateSet("Windows", "Background")]
    [string]$Mode = "Windows",

    [switch]$SkipMobile = $false,
    [switch]$OpenBrowser = $true
)

$ErrorActionPreference = "Stop"
$RootDir = $PSScriptRoot
if (-not $RootDir) { $RootDir = (Get-Location).Path }

function Write-Step {
    param([string]$Message)
    Write-Host "`n[LogiFlows] " -ForegroundColor Cyan -NoNewline
    Write-Host $Message -ForegroundColor White
}

function Write-Success {
    param([string]$Message)
    Write-Host "  [OK] " -ForegroundColor Green -NoNewline
    Write-Host $Message -ForegroundColor Gray
}

function Write-WarnMsg {
    param([string]$Message)
    Write-Host "  [WARN] " -ForegroundColor Yellow -NoNewline
    Write-Host $Message -ForegroundColor Gray
}

Clear-Host
Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "          LogiFlows - Intelligent Logistics Platform           " -ForegroundColor Cyan
Write-Host "               All-in-One Service Orchestrator                  " -ForegroundColor White
Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "Root Directory : $RootDir" -ForegroundColor Gray
Write-Host "Launch Mode    : $Mode" -ForegroundColor Gray

# ---------------------------------------------------------
# 1. Start Infrastructure (Docker Compose: Postgres + Redis)
# ---------------------------------------------------------
Write-Step "1/5 Checking Infrastructure (PostgreSQL & Redis)..."
if (Get-Command docker -ErrorAction SilentlyContinue) {
    docker compose up -d postgres redis
    if ($LASTEXITCODE -eq 0) {
        Write-Success "PostgreSQL (PostGIS 3.4) and Redis 7 containers initiated."
    } else {
        Write-WarnMsg "Docker compose returned non-zero exit code. Please ensure Docker Desktop is running."
    }
} else {
    Write-WarnMsg "Docker CLI not detected in PATH. Ensure PostgreSQL (:5432) and Redis (:6379) are running locally."
}

# ---------------------------------------------------------
# 2. Check and Launch Go Backend (:8080)
# ---------------------------------------------------------
Write-Step "2/5 Launching Go Backend API (:8080)..."
$BackendDir = Join-Path $RootDir "backend"

# Ensure port 8080 is freed if a zombie process is holding it
try {
    $existingPort = Get-NetTCPConnection -LocalPort 8080 -ErrorAction SilentlyContinue
    if ($existingPort) {
        $existingPort | Select-Object -ExpandProperty OwningProcess -Unique | ForEach-Object {
            if ($_ -gt 0) {
                Write-WarnMsg "Port 8080 is currently occupied by PID $_. Terminating stale process..."
                Stop-Process -Id $_ -Force -ErrorAction SilentlyContinue
                Start-Sleep -Milliseconds 500
            }
        }
    }
} catch {
    # Ignore errors during port cleanup
}

if (-not (Test-Path (Join-Path $BackendDir ".env"))) {
    if (Test-Path (Join-Path $RootDir ".env")) {
        Copy-Item (Join-Path $RootDir ".env") (Join-Path $BackendDir ".env") -Force
        Write-Success "Copied root .env to backend/.env"
    }
}

if ($Mode -eq "Windows") {
    $BackendCmd = "cd '$BackendDir'; go run ./cmd/api/main.go"
    Start-Process powershell -ArgumentList "-NoExit", "-Command", "`$Host.UI.RawUI.WindowTitle = 'LogiFlows - Backend API (:8080)'; $BackendCmd"
    Write-Success "Backend API launched in separate window (Port 8080)."
} else {
    $LogsDir = Join-Path $RootDir ".logs"
    if (-not (Test-Path $LogsDir)) { New-Item -ItemType Directory -Path $LogsDir | Out-Null }
    Start-Process powershell -ArgumentList "-Command", "cd '$BackendDir'; go run ./cmd/api/main.go *>&1 > '$LogsDir\backend.log'" -WindowStyle Hidden
    Write-Success "Backend API running in background (Logs: .logs/backend.log)."
}

# ---------------------------------------------------------
# 3. Check and Launch AI Service (:8000)
# ---------------------------------------------------------
Write-Step "3/5 Launching AI Predictive Intelligence Service (:8000)..."
$AiDir = Join-Path $RootDir "ai-service"
$AiVenvPython = Join-Path $AiDir ".venv\Scripts\python.exe"

$PythonExe = "python"
if (Test-Path $AiVenvPython) {
    $PythonExe = $AiVenvPython
    Write-Success "Using dedicated Python virtual environment: $AiVenvPython"
} else {
    Write-WarnMsg "Virtual environment not detected in ai-service/.venv, falling back to system 'python'."
}

if ($Mode -eq "Windows") {
    $AiCmd = "cd '$AiDir'; & '$PythonExe' -m uvicorn app.main:app --host 0.0.0.0 --port 8000 --reload"
    Start-Process powershell -ArgumentList "-NoExit", "-Command", "`$Host.UI.RawUI.WindowTitle = 'LogiFlows - AI Intelligence (:8000)'; $AiCmd"
    Write-Success "AI Service launched in separate window (Port 8000)."
} else {
    $LogsDir = Join-Path $RootDir ".logs"
    Start-Process powershell -ArgumentList "-Command", "cd '$AiDir'; & '$PythonExe' -m uvicorn app.main:app --host 0.0.0.0 --port 8000 *>&1 > '$LogsDir\ai-service.log'" -WindowStyle Hidden
    Write-Success "AI Service running in background (Logs: .logs/ai-service.log)."
}

# ---------------------------------------------------------
# 4. Check and Launch Frontend Web Console (:5173)
# ---------------------------------------------------------
Write-Step "4/5 Launching Frontend Web Dashboard (:5173)..."
$FrontendDir = Join-Path $RootDir "frontend"

if ($Mode -eq "Windows") {
    $FrontendCmd = "cd '$FrontendDir'; npm run dev -- --host 0.0.0.0 --port 5173"
    Start-Process powershell -ArgumentList "-NoExit", "-Command", "`$Host.UI.RawUI.WindowTitle = 'LogiFlows - Web Console (:5173)'; $FrontendCmd"
    Write-Success "Web Frontend launched in separate window (Port 5173)."
} else {
    $LogsDir = Join-Path $RootDir ".logs"
    Start-Process powershell -ArgumentList "-Command", "cd '$FrontendDir'; npm run dev -- --host 0.0.0.0 --port 5173 *>&1 > '$LogsDir\frontend.log'" -WindowStyle Hidden
    Write-Success "Web Frontend running in background (Logs: .logs/frontend.log)."
}

# ---------------------------------------------------------
# 5. Check and Launch Flutter Mobile Application (:8085)
# ---------------------------------------------------------
if (-not $SkipMobile) {
    Write-Step "5/5 Launching Flutter Mobile Client (:8085)..."
    $MobileDir = Join-Path $RootDir "mobile"
    
    # Locate Flutter executable
    $FlutterExe = $null
    $FlutterCmdObj = Get-Command flutter -ErrorAction SilentlyContinue
    if ($FlutterCmdObj) {
        $FlutterExe = $FlutterCmdObj.Source
    } elseif (Test-Path "$env:USERPROFILE\flutter\bin\flutter.bat") {
        $FlutterExe = "$env:USERPROFILE\flutter\bin\flutter.bat"
    } elseif (Test-Path "C:\flutter\bin\flutter.bat") {
        $FlutterExe = "C:\flutter\bin\flutter.bat"
    }

    if ($FlutterExe) {
        Write-Success "Detected Flutter SDK at: $FlutterExe"
        # Release stale lockfile if present
        $LockFile = Join-Path (Split-Path (Split-Path $FlutterExe)) "cache\lockfile"
        if (Test-Path $LockFile) {
            Remove-Item $LockFile -Force -ErrorAction SilentlyContinue
        }

        if ($Mode -eq "Windows") {
            $MobileCmd = "cd '$MobileDir'; & '$FlutterExe' run -d web-server --web-port 8085 --web-hostname 0.0.0.0"
            Start-Process powershell -ArgumentList "-NoExit", "-Command", "`$Host.UI.RawUI.WindowTitle = 'LogiFlows - Mobile Client (:8085)'; $MobileCmd"
            Write-Success "Flutter Mobile Web Server launched in separate window (Port 8085)."
        } else {
            $LogsDir = Join-Path $RootDir ".logs"
            Start-Process powershell -ArgumentList "-Command", "cd '$MobileDir'; & '$FlutterExe' run -d web-server --web-port 8085 --web-hostname 0.0.0.0 *>&1 > '$LogsDir\mobile.log'" -WindowStyle Hidden
            Write-Success "Flutter Mobile Web Server running in background (Logs: .logs/mobile.log)."
        }
    } else {
        Write-WarnMsg "Flutter SDK not found in PATH or standard user directories. Skipped mobile startup."
    }
} else {
    Write-Step "5/5 Skipping mobile startup (-SkipMobile specified)."
}

# ---------------------------------------------------------
# Waiting & Probing Health Endpoints
# ---------------------------------------------------------
Write-Host "`nWaiting for service initializations..." -ForegroundColor Yellow
Start-Sleep -Seconds 4

Write-Host "`n================================================================" -ForegroundColor Cyan
Write-Host "                LogiFlows Running Endpoints                     " -ForegroundColor Green
Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "  * Web Dashboard  : http://localhost:5173" -ForegroundColor Cyan
Write-Host "  * Mobile App     : http://localhost:8085" -ForegroundColor Cyan
Write-Host "  * Backend API    : http://localhost:8080" -ForegroundColor Cyan
Write-Host "  * Backend Swagger: http://localhost:8080/swagger/index.html" -ForegroundColor Cyan
Write-Host "  * AI Service     : http://localhost:8000" -ForegroundColor Cyan
Write-Host "  * AI Docs        : http://localhost:8000/docs" -ForegroundColor Cyan
Write-Host "  * PostgreSQL     : localhost:5432 (Database: logiflows_dev)" -ForegroundColor Gray
Write-Host "  * Redis Cache    : localhost:6379" -ForegroundColor Gray
Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "  To stop all services anytime, run: .\stop-all.ps1" -ForegroundColor Yellow
Write-Host "================================================================`n" -ForegroundColor Cyan

if ($OpenBrowser) {
    Start-Sleep -Seconds 1
    Start-Process "http://localhost:5173"
}
