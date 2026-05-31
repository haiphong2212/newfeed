@echo off
REM =============================================================================
REM Auth-Service Directory and File Creation Script
REM =============================================================================
REM This script creates the complete auth-service microservice structure

setlocal enabledelayedexpansion

set "BASE_PATH=d:\newfeed\services\auth-service"

echo.
echo ======= Creating Auth-Service Structure =======
echo Creating directories...
echo.

REM Create all directories
mkdir "!BASE_PATH!" 2>nul
mkdir "!BASE_PATH!\internal" 2>nul
mkdir "!BASE_PATH!\internal\domain" 2>nul
mkdir "!BASE_PATH!\internal\repository" 2>nul
mkdir "!BASE_PATH!\internal\infrastructure" 2>nul
mkdir "!BASE_PATH!\internal\usecase" 2>nul
mkdir "!BASE_PATH!\internal\config" 2>nul
mkdir "!BASE_PATH!\internal\delivery" 2>nul
mkdir "!BASE_PATH!\internal\delivery\grpc" 2>nul
mkdir "!BASE_PATH!\internal\delivery\http" 2>nul
mkdir "!BASE_PATH!\proto" 2>nul

echo ✓ Directories created
echo.
echo ====================================================
echo AUTH-SERVICE SETUP INSTRUCTIONS
echo ====================================================
echo.
echo Base Path: !BASE_PATH!
echo.
echo Next steps:
echo 1. Navigate to the auth-service directory:
echo    cd !BASE_PATH!
echo.
echo 2. Create Go files using the provided templates
echo.
echo 3. Run: go mod download
echo.
echo Directory structure created:
echo  - internal/domain/         (Domain models)
echo  - internal/repository/     (Data access layer)
echo  - internal/infrastructure/ (External services)
echo  - internal/usecase/        (Business logic)
echo  - internal/config/         (Configuration)
echo  - internal/delivery/grpc/  (gRPC server)
echo  - internal/delivery/http/  (HTTP handlers)
echo  - proto/                   (Protocol buffers)
echo.
