# 🛠️ Newfeed Development Guide

Complete guide for building and developing the Newfeed microservices platform.

## Table of Contents
1. [Prerequisites](#prerequisites)
2. [Project Setup](#project-setup)
3. [Architecture Overview](#architecture-overview)
4. [Service Development](#service-development)
5. [Database Schema](#database-schema)
6. [gRPC Implementation](#grpc-implementation)
7. [Testing Strategy](#testing-strategy)
8. [Deployment](#deployment)

## Prerequisites

### Required Software
- Go 1.21+
- Docker & Docker Compose
- Git
- Protocol Buffer Compiler (protoc)
- gRPC tools

### Installation

**Go 1.21+**
```bash
# Check version
go version
```

**Docker**
```bash
# Windows/Mac: Download from https://www.docker.com/products/docker-desktop
# Linux: sudo apt-get install docker.io docker-compose
```

**Protocol Buffers**
```bash
# Install protoc
go install github.com/protocolbuffers/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## Project Setup

### 1. Bootstrap the Project

**Windows (PowerShell)**
```powershell
.\BOOTSTRAP.ps1
```

**Linux/Mac (Bash)**
```bash
chmod +x bootstrap.sh
./bootstrap.sh
```

This creates:
- Service directories
- `go.mod` files for each service
- Placeholder `main.go` and `Dockerfile` for each service
- Database initialization SQL script

### 2. Start Infrastructure

```bash
# Start all containers
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f postgres
docker-compose logs -f rabbitmq
```

### 3. Verify Connectivity

```bash
# PostgreSQL
psql -h localhost -U postgres -d newfeed

# Redis
redis-cli ping

# RabbitMQ
curl http://localhost:15672/api/vhosts (guest:guest)

# MinIO
curl http://localhost:9000/minio/health/live

# Elasticsearch
curl http://localhost:9200/_cluster/health
```

## Architecture Overview

### Service Layers (Per Service)

Each microservice follows this structure:

```
service/
├── main.go                 # Entry point
├── go.mod                  # Dependencies
├── Dockerfile              # Container definition
├── internal/
│   ├── domain/             # Business entities
│   │   ├── user.go
│   │   ├── article.go
│   │   └── errors.go       # Domain errors
│   ├── usecase/            # Business logic
│   │   ├── create_user.go
│   │   ├── publish_article.go
│   │   └── interfaces.go   # Contracts
│   ├── repository/         # Data access
│   │   ├── user_repo.go
│   │   ├── article_repo.go
│   │   └── interfaces.go
│   ├── delivery/           # Transport layer
│   │   ├── grpc/           # gRPC handlers
│   │   │   └── user_service.go
│   │   ├── http/           # REST handlers
│   │   │   └── middleware.go
│   │   └── websocket/      # WebSocket handlers
│   ├── infrastructure/     # External services
│   │   ├── postgres.go
│   │   ├── redis.go
│   │   ├── rabbitmq.go
│   │   └── migrations.go
│   └── config/             # Configuration
│       └── config.go
├── proto/                  # gRPC definitions
│   └── service.proto
└── tests/                  # Test files
    ├── unit_test.go
    └── integration_test.go
```

### Communication Patterns

```
┌─────────────────────────────────────────────────────┐
│                    Clients                           │
└──────────────────┬──────────────────────────────────┘
                   │ REST (Port 8000)
┌──────────────────▼──────────────────────────────────┐
│              API Gateway                             │
│ - JWT Validation                                    │
│ - RBAC                                              │
│ - Rate Limiting                                     │
│ - Request Routing                                   │
└──────────────────┬──────────────────────────────────┘
                   │ gRPC (Ports 50051-58)
    ┌──────────────┬──────────────────────┐
    │              │                      │
    ▼              ▼                      ▼
┌─────────┐  ┌──────────┐          ┌──────────┐
│  Auth   │  │  User    │  ......  │ Analytics│
│Service  │  │  Service │          │ Service  │
└─────────┘  └──────────┘          └──────────┘
    │              │
    └──────────┬───┘
               │ RabbitMQ Events (Async)
        ┌──────▼──────────┐
        │ Event Queue     │
        │ - DLQ           │
        │ - Retry Logic   │
        └─────────────────┘
```

## Service Development

### Step 1: Create Service Directory

```bash
mkdir -p services/my-service/internal/{domain,usecase,repository,delivery,infrastructure}
cd services/my-service
```

### Step 2: Initialize Go Module

```bash
go mod init github.com/haiphong2212/newfeed/my-service
```

### Step 3: Add Core Dependencies

```bash
# gRPC and Protocol Buffers
go get google.golang.org/grpc
go get google.golang.org/protobuf

# Database
go get github.com/lib/pq           # PostgreSQL
go get github.com/redis/go-redis   # Redis

# Logging
go get github.com/sirupsen/logrus

# Configuration
go get github.com/kelseyhightower/envconfig

# Testing
go get github.com/stretchr/testify
```

### Step 4: Create Domain Layer

**internal/domain/user.go**
```go
package domain

import "time"

type User struct {
    ID        string    `db:"id"`
    Email     string    `db:"email"`
    Username  string    `db:"username"`
    Password  string    `db:"password_hash"`
    CreatedAt time.Time `db:"created_at"`
}

type CreateUserInput struct {
    Email    string
    Username string
    Password string
}

// Domain errors
type ErrUserNotFound struct{}
type ErrEmailExists struct{}
type ErrInvalidEmail struct{}
```

### Step 5: Create Repository Interface

**internal/repository/interfaces.go**
```go
package repository

import (
    "context"
    "github.com/haiphong2212/newfeed/my-service/internal/domain"
)

type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    GetByID(ctx context.Context, id string) (*domain.User, error)
    GetByEmail(ctx context.Context, email string) (*domain.User, error)
    Update(ctx context.Context, user *domain.User) error
    Delete(ctx context.Context, id string) error
}
```

### Step 6: Create Use Case

**internal/usecase/create_user.go**
```go
package usecase

import (
    "context"
    "github.com/haiphong2212/newfeed/my-service/internal/domain"
    "github.com/haiphong2212/newfeed/my-service/internal/repository"
)

type CreateUserUseCase struct {
    userRepo repository.UserRepository
}

func NewCreateUserUseCase(userRepo repository.UserRepository) *CreateUserUseCase {
    return &CreateUserUseCase{userRepo: userRepo}
}

func (u *CreateUserUseCase) Execute(ctx context.Context, input *domain.CreateUserInput) (*domain.User, error) {
    // Validate input
    if err := u.validate(input); err != nil {
        return nil, err
    }

    // Hash password
    passwordHash := hashPassword(input.Password)

    // Create user
    user := &domain.User{
        Email:    input.Email,
        Username: input.Username,
        Password: passwordHash,
    }

    // Save to database
    if err := u.userRepo.Create(ctx, user); err != nil {
        return nil, err
    }

    return user, nil
}

func (u *CreateUserUseCase) validate(input *domain.CreateUserInput) error {
    // Validation logic
    return nil
}
```

### Step 7: Create gRPC Handlers

**proto/user.proto**
```proto
syntax = "proto3";

package user;

message User {
    string id = 1;
    string email = 2;
    string username = 3;
    int64 created_at = 4;
}

message CreateUserRequest {
    string email = 1;
    string username = 2;
    string password = 3;
}

message CreateUserResponse {
    User user = 1;
}

service UserService {
    rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
    rpc GetUser(GetUserRequest) returns (GetUserResponse);
}
```

Generate code:
```bash
protoc --go_out=. --go-grpc_out=. proto/user.proto
```

### Step 8: Create Main Entry Point

**main.go**
```go
package main

import (
    "log"
    "net"
    "os"
    "os/signal"
    "syscall"

    "google.golang.org/grpc"
    "github.com/haiphong2212/newfeed/my-service/internal/config"
    "github.com/haiphong2212/newfeed/my-service/internal/delivery/grpc"
)

func main() {
    // Load config
    cfg := config.Load()

    // Setup database connection
    db, err := setupDatabase(cfg)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()

    // Create gRPC server
    grpcServer := grpc.NewServer()

    // Setup listeners
    lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
    if err != nil {
        log.Fatal("Failed to listen:", err)
    }

    // Start server in goroutine
    go func() {
        if err := grpcServer.Serve(lis); err != nil {
            log.Fatal("Server error:", err)
        }
    }()

    log.Println("Service started on port", cfg.GRPCPort)

    // Graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
    <-sigChan

    log.Println("Shutting down gracefully...")
    grpcServer.GracefulStop()
    log.Println("Service stopped")
}

func setupDatabase(cfg *config.Config) (*sql.DB, error) {
    // Database connection logic
    dsn := "postgres://..."
    return sql.Open("postgres", dsn)
}
```

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Articles Table
```sql
CREATE TABLE articles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(300) NOT NULL,
    content TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'DRAFT',
    published_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

See `scripts/init-db.sql` for complete schema.

## gRPC Implementation

### Proto File Best Practices

```proto
syntax = "proto3";

package myservice;

option go_package = "github.com/haiphong2212/newfeed/my-service/proto";

// Clear message structures
message CreateArticleRequest {
    string title = 1;
    string content = 2;
    repeated string tags = 3;
}

message Article {
    string id = 1;
    string user_id = 2;
    string title = 3;
    string content = 4;
    string status = 5;
    int64 published_at = 6;  // Unix timestamp
}

// Define multiple RPC methods
service ArticleService {
    rpc CreateArticle(CreateArticleRequest) returns (Article);
    rpc GetArticle(GetArticleRequest) returns (Article);
    rpc ListArticles(ListArticlesRequest) returns (ListArticlesResponse);
}
```

### Interceptors for Middleware

```go
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, 
        handler grpc.UnaryHandler) (interface{}, error) {
        
        // Pre-processing
        log.Printf("Calling %s", info.FullMethod)

        // Call handler
        resp, err := handler(ctx, req)

        // Post-processing
        if err != nil {
            log.Printf("Error: %v", err)
        }

        return resp, err
    }
}
```

## Testing Strategy

### Unit Tests

**internal/usecase/create_user_test.go**
```go
package usecase

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestCreateUser_Execute_Success(t *testing.T) {
    mockRepo := new(MockUserRepository)
    mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

    useCase := NewCreateUserUseCase(mockRepo)
    result, err := useCase.Execute(context.Background(), &domain.CreateUserInput{
        Email:    "test@example.com",
        Username: "testuser",
        Password: "password123",
    })

    assert.NoError(t, err)
    assert.NotNil(t, result)
    mockRepo.AssertExpectations(t)
}
```

### Integration Tests

```bash
# Start test containers
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
go test -v -tags integration ./...

# Cleanup
docker-compose -f docker-compose.test.yml down
```

### End-to-End Tests

Test complete workflows:
1. User registration
2. Article publishing
3. Chat message sending
4. Notification delivery

## Deployment

### Docker Build

```bash
# Build specific service
docker build -t newfeed-api-gateway services/api-gateway/

# Run container
docker run -p 8000:8000 newfeed-api-gateway

# Push to registry
docker tag newfeed-api-gateway registry.example.com/newfeed-api-gateway:latest
docker push registry.example.com/newfeed-api-gateway:latest
```

### Environment Variables

Create `.env` file based on `.env.example`:

```bash
ENV=production
LOG_LEVEL=info
JWT_SECRET=your-secret-key
DB_PASSWORD=strong-password
MINIO_ROOT_PASSWORD=strong-password
```

### Kubernetes Deployment (Optional)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api-gateway
  template:
    metadata:
      labels:
        app: api-gateway
    spec:
      containers:
      - name: api-gateway
        image: registry.example.com/newfeed-api-gateway:latest
        ports:
        - containerPort: 8000
        env:
        - name: ENV
          value: "production"
        livenessProbe:
          httpGet:
            path: /health
            port: 8001
          initialDelaySeconds: 10
          periodSeconds: 10
```

## Best Practices

### Code Organization
- ✅ Keep domain logic separate from delivery
- ✅ Use dependency injection
- ✅ Interface-based design for testability
- ✅ Clear error handling

### Database
- ✅ Use connection pooling
- ✅ Add proper indexes
- ✅ Use transactions for consistency
- ✅ Implement migrations

### Error Handling
```go
type ServiceError struct {
    Code    string
    Message string
    Details map[string]interface{}
}

// Use custom errors consistently
if err != nil {
    return nil, &ServiceError{
        Code:    "USER_NOT_FOUND",
        Message: "User does not exist",
    }
}
```

### Logging
```go
import "github.com/sirupsen/logrus"

log := logrus.WithFields(logrus.Fields{
    "service": "user-service",
    "request_id": correlationID,
    "user_id": userID,
})
log.Info("User created successfully")
```

## Troubleshooting

### Service won't start
```bash
# Check logs
docker-compose logs user-service

# Verify dependencies
docker-compose ps

# Test connectivity
telnet localhost 5432  # PostgreSQL
```

### Database connection fails
```bash
# Check PostgreSQL is running
psql -h localhost -U postgres -d newfeed

# Check credentials in .env
grep DB_ .env
```

### gRPC connection timeout
```bash
# Verify service is running
docker-compose ps | grep auth-service

# Check port forwarding
netstat -an | grep 50051
```

---

**Next Step**: Start with Auth Service implementation to establish the authentication foundation for all other services.
