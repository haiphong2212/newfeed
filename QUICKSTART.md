# ⚡ Quick Start - 5 Minutes to Running Infrastructure

## One-Command Setup

### Windows (PowerShell)
```powershell
# Run setup + start infrastructure
.\BOOTSTRAP.ps1
docker-compose up -d
docker-compose ps
```

### Linux/Mac (Bash)
```bash
# Run setup + start infrastructure
chmod +x bootstrap.sh
./bootstrap.sh
docker-compose up -d
docker-compose ps
```

## ✅ What You Get

### Infrastructure Running
- ✅ PostgreSQL (5432)
- ✅ Redis (6379)
- ✅ RabbitMQ (5672)
- ✅ MinIO (9000/9001)
- ✅ Elasticsearch (9200)

### Service Placeholders Created
- ✅ 9 service directories
- ✅ go.mod files
- ✅ main.go stubs
- ✅ Dockerfile templates
- ✅ Database schema

## 🌐 Access Services

| Service | URL | Credentials |
|---------|-----|-------------|
| PostgreSQL | localhost:5432 | postgres/postgres |
| Redis | localhost:6379 | - |
| RabbitMQ | http://localhost:15672 | guest/guest |
| MinIO | http://localhost:9001 | minioadmin/minioadmin |
| Elasticsearch | http://localhost:9200 | - |

## 📖 Documentation

After setup, read in this order:
1. **INITIALIZATION_SUMMARY.md** - What was created
2. **DEVELOPMENT_GUIDE.md** - How to build services
3. **README.md** - Full project overview

## 🛠️ Verify Installation

```bash
# Check Docker containers
docker-compose ps

# Test PostgreSQL
psql -h localhost -U postgres -d newfeed -c "SELECT 1;"

# Test Redis
redis-cli ping

# Test Elasticsearch
curl http://localhost:9200/_cluster/health

# Check RabbitMQ
curl http://localhost:15672/api/vhosts -u guest:guest
```

## 🚨 Common Issues

### Ports Already in Use
```bash
# Find what's using port 5432
netstat -ano | findstr :5432

# Kill process (Windows)
taskkill /PID <PID> /F

# Or change port in .env
DB_PORT=5433
```

### Docker Not Running
```bash
# Restart Docker Desktop
# On Linux: sudo systemctl restart docker
```

### Container Fails to Start
```bash
# Check logs
docker-compose logs postgres
docker-compose logs redis

# Rebuild
docker-compose down -v
docker-compose up --build
```

## 📊 Next Steps

1. **Run bootstrap** → Creates project structure
2. **Start docker-compose** → Infrastructure ready
3. **Read DEVELOPMENT_GUIDE.md** → Learn architecture
4. **Implement Auth Service** → Start building services
5. **Test with docker-compose** → Verify everything works

## 💡 Useful Commands

```bash
# View all logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f auth-service

# Stop everything
docker-compose down

# Complete reset (removes volumes)
docker-compose down -v

# Rebuild a service
docker-compose build api-gateway
docker-compose up -d api-gateway

# Execute command in container
docker-compose exec postgres psql -U postgres -d newfeed

# Check resource usage
docker stats

# Inspect network
docker network inspect newfeed-network
```

## 🎯 You're Ready!

Your infrastructure is set up. Now build the services following the DEVELOPMENT_GUIDE.md

---

**First**: Read `INITIALIZATION_SUMMARY.md` for complete overview
**Then**: Follow `DEVELOPMENT_GUIDE.md` to implement Auth Service
