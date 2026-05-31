#!/usr/bin/env pwsh
<#
.SYNOPSIS
Newfeed Platform Bootstrap Script
Creates directory structure and initializes all services

.EXAMPLE
.\BOOTSTRAP.ps1
#>

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

Write-Host "🚀 Bootstrapping Newfeed Community News Platform..." -ForegroundColor Cyan
Write-Host ""

# Colors
$Green = "Green"
$Yellow = "Yellow"
$Red = "Red"

# Step 1: Create directory structure
Write-Host "📁 Step 1: Creating directory structure..." -ForegroundColor Cyan

$services = @(
    "api-gateway",
    "auth-service",
    "user-service",
    "news-service",
    "search-service",
    "chat-service",
    "notification-service",
    "media-service",
    "analytics-service"
)

$directories = @(
    "services",
    "proto",
    "shared",
    "shared/logger",
    "shared/errors",
    "shared/config",
    "shared/middleware",
    "scripts"
)

# Create base directories
foreach ($dir in $directories) {
    $fullPath = Join-Path (Get-Location) $dir
    if (!(Test-Path $fullPath)) {
        New-Item -ItemType Directory -Path $fullPath -Force | Out-Null
        Write-Host "✅ Created: $dir" -ForegroundColor $Green
    }
}

# Create service directories with go.mod
foreach ($service in $services) {
    $servicePath = Join-Path (Get-Location) "services\$service"
    if (!(Test-Path $servicePath)) {
        New-Item -ItemType Directory -Path $servicePath -Force | Out-Null
        Write-Host "✅ Created: services/$service" -ForegroundColor $Green
    }

    # Create go.mod
    $goModPath = Join-Path $servicePath "go.mod"
    if (!(Test-Path $goModPath)) {
        $goModContent = @"
module github.com/haiphong2212/newfeed/$service

go 1.21

require (
    google.golang.org/grpc v1.59.0
    google.golang.org/protobuf v1.31.0
)
"@
        Set-Content -Path $goModPath -Value $goModContent
        Write-Host "✅ Created: services/$service/go.mod" -ForegroundColor $Green
    }

    # Create main.go placeholder
    $mainPath = Join-Path $servicePath "main.go"
    if (!(Test-Path $mainPath)) {
        $mainContent = @"
package main

import (
    "log"
)

func main() {
    log.Println("$service starting...")
    // TODO: Service implementation
}
"@
        Set-Content -Path $mainPath -Value $mainContent
        Write-Host "✅ Created: services/$service/main.go" -ForegroundColor $Green
    }

    # Create Dockerfile
    $dockerfilePath = Join-Path $servicePath "Dockerfile"
    if (!(Test-Path $dockerfilePath)) {
        $dockerfileContent = @"
FROM golang:1.21-alpine AS builder
WORKDIR /build
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o $service .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /build/$service .
EXPOSE 8000 50051
HEALTHCHECK --interval=10s --timeout=5s --retries=5 `
    CMD wget --no-verbose --tries=1 --spider http://localhost:8001/health || exit 1
CMD ["./$service"]
"@
        Set-Content -Path $dockerfilePath -Value $dockerfileContent
        Write-Host "✅ Created: services/$service/Dockerfile" -ForegroundColor $Green
    }
}

Write-Host ""
Write-Host "📝 Step 2: Creating SQL initialization script..." -ForegroundColor Cyan

$sqlPath = Join-Path (Get-Location) "scripts\init-db.sql"
$sqlContent = @'
-- Initial Database Schema for Newfeed Community News Platform
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    avatar_url TEXT,
    bio TEXT,
    email_verified BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Articles table
CREATE TABLE IF NOT EXISTS articles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(300) NOT NULL,
    content TEXT NOT NULL,
    excerpt VARCHAR(500),
    status VARCHAR(50) NOT NULL DEFAULT 'DRAFT',
    published_at TIMESTAMP,
    view_count INT DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Categories table
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Tags table
CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Article tags
CREATE TABLE IF NOT EXISTS article_tags (
    article_id UUID NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (article_id, tag_id)
);

-- Comments table
CREATE TABLE IF NOT EXISTS comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    article_id UUID NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Chat rooms
CREATE TABLE IF NOT EXISTS chat_rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    article_id UUID NOT NULL UNIQUE REFERENCES articles(id) ON DELETE CASCADE,
    name VARCHAR(300) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Chat messages
CREATE TABLE IF NOT EXISTS chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL REFERENCES chat_rooms(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Notifications
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Refresh tokens
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_articles_user_id ON articles(user_id);
CREATE INDEX IF NOT EXISTS idx_articles_status ON articles(status);
CREATE INDEX IF NOT EXISTS idx_articles_created_at ON articles(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_comments_article_id ON comments(article_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_room_id ON chat_messages(room_id);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);

-- Initial categories
INSERT INTO categories (name, slug, description) VALUES
    ('Technology', 'technology', 'Latest tech news'),
    ('Business', 'business', 'Business news'),
    ('Science', 'science', 'Science news'),
    ('Entertainment', 'entertainment', 'Entertainment'),
    ('Health', 'health', 'Health news'),
    ('Sports', 'sports', 'Sports news')
ON CONFLICT (name) DO NOTHING;
'@

Set-Content -Path $sqlPath -Value $sqlContent
Write-Host "✅ Created: scripts/init-db.sql" -ForegroundColor $Green

Write-Host ""
Write-Host "✅ Bootstrap Complete!" -ForegroundColor $Green
Write-Host ""
Write-Host "📋 Next Steps:" -ForegroundColor Cyan
Write-Host "  1. Review configuration: Edit .env if needed" -ForegroundColor $Yellow
Write-Host "  2. Start infrastructure: docker-compose up -d" -ForegroundColor $Yellow
Write-Host "  3. Check services: docker-compose ps" -ForegroundColor $Yellow
Write-Host "  4. View logs: docker-compose logs -f" -ForegroundColor $Yellow
Write-Host ""
Write-Host "🔗 Access Points:" -ForegroundColor Cyan
Write-Host "  - RabbitMQ: http://localhost:15672 (guest/guest)" -ForegroundColor $Yellow
Write-Host "  - MinIO: http://localhost:9001 (minioadmin/minioadmin)" -ForegroundColor $Yellow
Write-Host "  - Elasticsearch: http://localhost:9200" -ForegroundColor $Yellow
Write-Host "  - PostgreSQL: localhost:5432" -ForegroundColor $Yellow
Write-Host ""
