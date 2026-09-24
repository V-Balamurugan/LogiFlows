@echo off
title LogiFlows - Stop Services
echo ================================================================
echo           LogiFlows - Stopping All Platform Services
echo ================================================================
echo.

powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0stop-all.ps1" %*

echo.
pause
