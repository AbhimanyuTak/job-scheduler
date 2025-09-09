# Setup Guide

## Quick Start with Docker

### 1. Start All Services
```bash
# Clone and start
git clone <repository-url>
cd job_scheduler
docker-compose up -d

# Check status
docker-compose ps

# Test health
curl http://localhost:8080/health
```

### 2. Scale Workers
```bash
# Scale to 5 workers
docker-compose up -d --scale worker=5

# Monitor queue
curl http://localhost:8080/queue/stats
```

## Manual Setup

### 1. Prerequisites
- Go 1.23+
- PostgreSQL 15+
- Redis 7+

### 2. Database Setup
```bash
# Install PostgreSQL
brew install postgresql  # macOS
sudo apt-get install postgresql  # Ubuntu

# Create database
psql -U postgres -c "CREATE DATABASE job_scheduler;"
```

### 3. Redis Setup
```bash
# Install Redis
brew install redis  # macOS
sudo apt-get install redis-server  # Ubuntu

# Start Redis
redis-server
```

### 4. Run Application
```bash
# Build
go build ./cmd/scheduler
go build ./cmd/worker

# Run (separate terminals)
./scheduler
./worker
```

## Configuration

### Environment Variables
```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=job_scheduler

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# Server
SERVER_PORT=8080
GIN_MODE=debug
```

### Configuration File
```bash
# Copy template
cp config.env.example config.env

# Edit values
nano config.env
```

## Docker Commands

### Basic Operations
```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f scheduler

# Stop services
docker-compose down

# Reset database
docker-compose down -v
```

### Development
```bash
# Start with live reload
docker-compose up -d

# Rebuild
docker-compose up -d --build

# Execute in container
docker-compose exec scheduler /bin/sh
```

## Testing

### Unit Tests
```bash
# Fast tests (no external dependencies)
make -f test/Makefile.test test-unit

# With coverage
go test -v -short -run ".*Unit" ./internal/... -cover
```

### Integration Tests
```bash
# Requires running services
make -f test/Makefile.test test-integration
```

## Troubleshooting

### Common Issues

#### Port Conflicts
```bash
# Check ports
netstat -tulpn | grep :8080
netstat -tulpn | grep :5432

# Use different ports in docker-compose.yml
ports:
  - "8081:8080"
```

#### Database Connection
```bash
# Test connection
psql -h localhost -U postgres -d job_scheduler

# Check Docker logs
docker-compose logs postgres
```

#### Redis Connection
```bash
# Test Redis
redis-cli ping

# Check Docker logs
docker-compose logs redis
```
