@echo off
echo ========================================
echo  Testing Reverse Proxy
echo ========================================
echo.

echo [1] Testing proxy requests (should alternate between backend1 and backend2):
echo.
curl http://localhost:8080/
echo.
curl http://localhost:8080/
echo.
curl http://localhost:8080/
echo.

echo ========================================
echo [2] Checking system status:
echo.
curl http://localhost:8081/status
echo.
echo.

echo ========================================
echo [3] Adding a new backend:
echo.
curl -X POST http://localhost:8081/Backends -H "Content-Type: application/json" -d "{\"url\": \"http://localhost:9003\"}"
echo.
echo.

echo ========================================
echo [4] Checking status after adding backend:
echo.
curl http://localhost:8081/status
echo.
echo.

echo ========================================
echo [5] Removing a backend:
echo.
curl -X DELETE "http://localhost:8081/Backends?url=http://localhost:9003"
echo.
echo.

echo ========================================
echo [6] Final status check:
echo.
curl http://localhost:8081/status
echo.
echo.

echo ========================================
echo Testing complete!
echo ========================================
pause
