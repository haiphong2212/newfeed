# 🚀 Newfeed - Real-Time Community News Platform

A production-oriented microservices backend system built with Go, featuring real-time discussions, event-driven notifications, full-text search, and scalable distributed patterns.

## 📋 Overview

Newfeed is a comprehensive backend platform that enables users to:
- Read and publish articles with real-time indexing
- Participate in live discussion rooms per article
- Receive real-time notifications
- Search content with full-text capabilities
- Follow users and topics
- Engage with interactive reactions and bookmarks

## 🏗️ Architecture

### Microservices (9 services)
1. **API Gateway** - Authentication, authorization, rate limiting, routing
2. **Auth Service** - JWT, token management, user registration  
3. **User Service** - Profiles, follow relationships, preferences
4. **News Service** - Article CRUD, publishing, categorization
5. **Search Service** - Elasticsearch integration, full-text search
6. **Chat Service** - WebSocket, real-time rooms, presence tracking
7. **Notification Service** - Event-driven alerts, RabbitMQ consumer
8. **Media Service** - File upload, MinIO integration, presigned URLs
9. **Analytics Service** - Trending calculation, metrics tracking

### Communication Patterns
- **REST** (8000) - Client → API Gateway
- **gRPC** (50051-58) - Internal service-to-service
- **WebSocket** (8005) - Real-time chat
- **RabbitMQ** (5672) - Async event messaging

### Storage
- **PostgreSQL** - Relational data
- **Redis** - Caching, trending, presence
- **Elasticsearch** - Full-text search
- **MinIO** - Object storage

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.21+

### Setup
```bash
cp .env.example .env
docker-compose up -d
docker-compose ps
```

### Access Points
- API Gateway: http://localhost:8000
- RabbitMQ: http://localhost:15672
- MinIO: http://localhost:9001
- Elasticsearch: http://localhost:9200
- PostgreSQL: localhost:5432

## 📊 Database

Core tables: users, articles, categories, tags, comments, chat_rooms, notifications, bookmarks, reactions

Full schema: `scripts/init-db.sql`

## 🔄 Event-Driven

RabbitMQ Events:
- ArticlePublished
- UserMentioned  
- CommentCreated
- FollowTopicCreated

Dead Letter Queue for retry & monitoring.

## 🔐 Authentication

- JWT access tokens (15m)
- Refresh tokens (7d)
- RBAC: admin, editor, user, moderator

## 📚 Project Structure

```
services/           # 9 microservices
proto/              # gRPC definitions
shared/             # Common libraries
scripts/            # Utilities
docker-compose.yml  # Infrastructure
```

## 📖 Development

```bash
# Build
cd services/auth-service
go build -o auth-service

# Test
go test ./...

# Generate gRPC
protoc --go_out=. --go-grpc_out=. proto/auth.proto
```

---

**Status**: 🟡 Infrastructure setup complete - Services under development