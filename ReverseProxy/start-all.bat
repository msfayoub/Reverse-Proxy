@echo off
echo ========================================
echo  Starting Complete Reverse Proxy System
echo ========================================
echo.
echo This will open 3 windows:
echo   1. Backend Server 1 (port 9001)
echo   2. Backend Server 2 (port 9002)
echo   3. Reverse Proxy (ports 8080 and 8081)
echo.
echo Wait for all servers to start, then run test-proxy.bat
echo.
pause

cd /d "%~dp0"

start "Backend 1" cmd /k start-backend1.bat
timeout /t 2 /nobreak >nul

start "Backend 2" cmd /k start-backend2.bat
timeout /t 2 /nobreak >nul

start "Reverse Proxy" cmd /k start-proxy.bat

echo.
echo ========================================
echo All servers are starting...
echo ========================================
echo.
echo Backend 1: http://localhost:9001
echo Backend 2: http://localhost:9002
echo Proxy:     http://localhost:8080
echo Admin API: http://localhost:8081
echo.
echo Run test-proxy.bat to test the system!
echo.
pause
