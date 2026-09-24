<#
.SYNOPSIS
    LogiFlows - Stop All Services Automation Script
.DESCRIPTION
    Stops all development processes on ports 8080, 8000, 5173, and 8085.
    Optionally stops Docker infrastructure containers.
.PARAMETER StopDocker
    Switch to also run 'docker compose down'.
#>

param(
    [switch]$StopDocker = $false
)

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

Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "               LogiFlows - Stopping All Services                " -ForegroundColor White
Write-Host "================================================================" -ForegroundColor Cyan

$PortsToStop = @(
    @{ Port = 8080; Name = "Backend API" },
    @{ Port = 8000; Name = "AI Predictive Service" },
    @{ Port = 5173; Name = "Web Dashboard" },
    @{ Port = 8085; Name = "Flutter Mobile Web Server" }
)

foreach ($target in $PortsToStop) {
    Write-Step "Checking $($target.Name) on port $($target.Port)..."
    $connections = Get-NetTCPConnection -LocalPort $target.Port -ErrorAction SilentlyContinue
    if ($connections) {
        $pids = $connections | Select-Object -ExpandProperty OwningProcess -Unique
        foreach ($procId in $pids) {
            try {
                $proc = Get-Process -Id $procId -ErrorAction SilentlyContinue
                if ($proc) {
                    Stop-Process -Id $procId -Force -ErrorAction SilentlyContinue
                    Write-Success "Terminated $($proc.ProcessName) (PID: $procId) on port $($target.Port)"
                }
            } catch {
                # Ignore process already stopped
            }
        }
    } else {
        Write-Success "Port $($target.Port) is already free."
    }
}

# Clear any lingering Flutter lockfile
$LockFile = "$env:USERPROFILE\flutter\bin\cache\lockfile"
if (Test-Path $LockFile) {
    Remove-Item $LockFile -Force -ErrorAction SilentlyContinue
}

if ($StopDocker) {
    Write-Step "Stopping Docker infrastructure containers (PostgreSQL, Redis)..."
    if (Get-Command docker -ErrorAction SilentlyContinue) {
        docker compose down
        Write-Success "Docker containers stopped."
    }
} else {
    Write-Host "`nNote: Docker containers (PostgreSQL, Redis) are kept running for quick restart." -ForegroundColor Gray
    Write-Host "To stop Docker containers too, run: .\stop-all.ps1 -StopDocker" -ForegroundColor Gray
}

Write-Host "`n================================================================" -ForegroundColor Cyan
Write-Host "              All LogiFlows Services Stopped                    " -ForegroundColor Green
Write-Host "================================================================`n" -ForegroundColor Cyan
