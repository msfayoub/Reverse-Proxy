@echo off
echo ========================================
echo  Starting Reverse Proxy
echo  Proxy Server: http://localhost:8080
echo  Admin API: http://localhost:8081
echo ========================================
cd /d "%~dp0"
go run main.go
