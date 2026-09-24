@echo off
title LogiFlows - Service Orchestrator
echo ================================================================
echo           LogiFlows - Starting Entire Platform Stack
echo ================================================================
echo.

powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0run-all.ps1" %*

echo.
pause
