@echo off
echo ========================================
echo  Starting Backend Server 2 on port 9002
echo ========================================
cd /d "%~dp0"
go run Backends\backend2.go
