# Newfeed Platform - Setup Instructions

## Quick Start

### Step 1: Create Directory Structure
Run this in PowerShell from the project root:

```powershell
$services = @("api-gateway", "auth-service", "user-service", "news-service", "search-service", "chat-service", "notification-service", "media-service", "analytics-service")
$dirs = @("proto", "shared/logger", "shared/errors", "shared/config", "shared/middleware", "scripts")

foreach ($service in $services) {
    New-Item -ItemType Directory -Path "services/$service" -Force | Out-Null
}

foreach ($dir in $dirs) {
    New-Item -ItemType Directory -Path $dir -Force | Out-Null
}

Write-Host "✅ Directory structure created"
```

### Step 2: Create go.mod Files for Each Service

Run in PowerShell:

```powershell
$services = @("api-gateway", "auth-service", "user-service", "news-service", "search-service", "chat-service", "notification-service", "media-service", "analytics-service")

foreach ($service in $services) {
    $goModContent = @"
module github.com/haiphong2212/newfeed/$service

go 1.21

require (
    google.golang.org/grpc v1.59.0
    google.golang.org/protobuf v1.31.0
)
"@
    Set-Content -Path "services/$service/go.mod" -Value $goModContent
}

Write-Host "✅ go.mod files created for all services"
```

### Step 3: Create Database SQL File

Create `scripts/init-db.sql` with content from the section below.

### Step 4: Verify Files

```powershell
# Check Docker Compose file
Test-Path docker-compose.yml

# Check .env files
Test-Path .env
Test-Path .env.example

# Check services
Get-ChildItem -Path services -Directory
```

### Step 5: Start Infrastructure

```bash
docker-compose up -d
```

Check status:
```bash
docker-compose ps
```

Access services:
- API Gateway: http://localhost:8000
- RabbitMQ Management: http://localhost:15672 (guest:guest)
- MinIO Console: http://localhost:9001 (minioadmin:minioadmin)
- Elasticsearch: http://localhost:9200
- PostgreSQL: localhost:5432

## Files Already Created

✅ `docker-compose.yml` - Complete infrastructure setup
✅ `.env` - Environment configuration
✅ `.env.example` - Environment template

## Files to Create

### 1. Create `scripts/init-db.sql`

[Copy the SQL schema from the database initialization section]

### 2. Create Placeholder Dockerfile for Each Service

Each service needs a `Dockerfile`. Here's a template:

**services/api-gateway/Dockerfile**
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /build
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o api-gateway .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /build/api-gateway .
EXPOSE 8000 8001
HEALTHCHECK --interval=10s --timeout=5s --retries=5 CMD wget --no-verbose --tries=1 --spider http://localhost:8001/health || exit 1
CMD ["./api-gateway"]
```

Apply this pattern to all other services, changing the port numbers appropriately.

### 3. Create Service Entry Points

Create `main.go` for each service:

**services/api-gateway/main.go**
```go
package main

import (
	"log"
)

func main() {
	log.Println("API Gateway starting...")
	// TODO: Implement
}
```

## Database Schema

The database will be initialized automatically via `scripts/init-db.sql` when PostgreSQL starts.

Tables created:
- users
- articles
- categories
- article_categories
- tags
- article_tags
- comments
- user_follows
- topic_follows
- bookmarks
- reactions
- chat_rooms
- chat_messages
- notifications
- media_metadata
- refresh_tokens
- analytics_events

## Microservices Communication

- **gRPC**: Internal service-to-service communication (ports 50051-50058)
- **REST**: External client API (port 8000 - API Gateway)
- **WebSocket**: Real-time chat (via Chat Service)
- **RabbitMQ**: Async event messaging (port 5672)

## Next Phase

Once infrastructure is running:
1. Build Auth Service (base for all others)
2. Build User Service
3. Build API Gateway
4. Continue with remaining services
