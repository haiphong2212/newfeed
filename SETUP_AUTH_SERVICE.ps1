#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Auth-Service Complete Setup Script
    Creates all directories and files for the auth-service microservice
#>

$ErrorActionPreference = "Stop"
$basePath = "d:\newfeed\services\auth-service"

# Create directory structure
$directories = @(
    $basePath,
    "$basePath\internal",
    "$basePath\internal\domain",
    "$basePath\internal\repository",
    "$basePath\internal\infrastructure",
    "$basePath\internal\usecase",
    "$basePath\internal\config",
    "$basePath\internal\delivery",
    "$basePath\internal\delivery\grpc",
    "$basePath\internal\delivery\http",
    "$basePath\proto"
)

Write-Host "Creating auth-service structure..." -ForegroundColor Cyan

foreach ($dir in $directories) {
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir -Force | Out-Null
        Write-Host "✓ Created: $dir" -ForegroundColor Green
    }
}

# Create go.mod
$goModContent = @"
module github.com/newfeed/services/auth-service

go 1.21

require (
	github.com/golang-jwt/jwt/v5 v5.0.0
	github.com/google/uuid v1.4.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.18.1
	golang.org/x/crypto v0.15.0
	google.golang.org/grpc v1.59.0
	google.golang.org/protobuf v1.31.0
	github.com/lib/pq v1.10.9
	github.com/joho/godotenv v1.5.1
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	golang.org/x/net v0.18.0 // indirect
	golang.org/x/sys v0.14.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20231211222908-948df8a8f126 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231211222908-948df8a8f126 // indirect
)
"@

Set-Content -Path "$basePath\go.mod" -Value $goModContent
Write-Host "✓ Created: go.mod" -ForegroundColor Green

Write-Host ""
Write-Host "Auth-service structure created successfully!" -ForegroundColor Green
Write-Host "Path: $basePath" -ForegroundColor Cyan
