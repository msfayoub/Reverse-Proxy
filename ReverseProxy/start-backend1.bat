@echo off
echo ========================================
echo  Starting Backend Server 1 on port 9001
echo ========================================
cd /d "%~dp0"
go run Backends\backend1.go
