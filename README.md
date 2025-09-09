# Job Scheduler

A distributed job scheduler built with Go that supports CRON-based scheduling with extended second-level precision and Redis-based queue processing for long-running tasks.

## Features

- **Extended CRON Support**: Parse CRON expressions with seconds (e.g., "31 10-15 1 * * MON-FRI")
- **Execution Types**: Support for AT_LEAST_ONCE and AT_MOST_ONCE execution guarantees
- **Job Management**: Create and track job execution history
- **High Performance**: Designed to handle 10k+ jobs per second
- **Fault Tolerant**: Built-in retry logic and error handling
- **Redis Queue System**: Distributed job processing with horizontal scaling
- **Long-Running Tasks**: Support for tasks up to 90 seconds duration
- **Auto-Scaling Ready**: Worker containers can be scaled based on queue depth

## Project Structure

```
job_scheduler/
├── cmd/
│   ├── scheduler/          # Main API server entry point
│   └── worker/             # Worker service entry point
├── internal/
│   ├── models/             # Data models (Job, JobExecution, QueueJob)
│   ├── storage/            # PostgreSQL storage layer
│   ├── services/           # Core services (Scheduler, Worker, Redis, Queue)
│   ├── handlers/           # HTTP API handlers
│   ├── database/           # Database connection management
│   └── utils/              # Utility functions (schedule parsing)
├── docs/                   # Documentation
├── scripts/                # Database and deployment scripts
├── docker-compose.yml      # Production Docker setup
├── docker-compose.dev.yml  # Development Docker setup
├── Dockerfile              # API server container
├── Dockerfile.worker       # Worker container
├── go.mod
└── README.md
```

## Quick Start

### Docker (Recommended)

```bash
# Production
docker-compose up -d

# Development (with live reload)
docker-compose -f docker-compose.dev.yml up -d

# Test
curl http://localhost:8080/health
curl http://localhost:8080/queue/stats

# Scale workers
docker-compose up -d --scale worker=5

# Stop
docker-compose down
```

### Manual Setup

1. Setup PostgreSQL and Redis
2. Build: `go build ./cmd/scheduler && go build ./cmd/worker`
3. Run: `go run ./cmd/scheduler` and `go run ./cmd/worker`

## Configuration

### Environment Variables

Key configuration options (see `config.env.example` for full list):

- **Database**: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- **Redis**: `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB`
- **Server**: `SERVER_PORT` (default: 8080), `GIN_MODE`
- **Worker**: `WORKER_POOL_SIZE` (default: 10), `WORKER_HTTP_TIMEOUT` (default: 90s)
- **Scheduler**: `SCHEDULER_POLL_INTERVAL` (default: 5s), `SCHEDULER_BATCH_SIZE` (default: 100)

### Docker Files

- **`docker-compose.yml`**: Production (3 workers, optimized)
- **`docker-compose.dev.yml`**: Development (2 workers, live reload)

### Documentation

- **[Configuration Guide](docs/configuration.md)** - Environment variables
- **[Database Setup](docs/database-setup.md)** - PostgreSQL setup
- **[Docker Setup](docs/docker-setup.md)** - Docker deployment
- **[API Spec](docs/api-spec.yaml)** - OpenAPI specification
- **[API Endpoints](docs/api-endpoints.md)** - Quick reference

### Key Endpoints

- `GET /health` - Health check
- `GET /queue/stats` - Queue statistics
- `POST /api/v1/jobs` - Create job
- `GET /api/v1/jobs` - List jobs
- `GET /api/v1/jobs/{id}` - Get job details

## Testing

### Quick Commands

```bash
# Unit tests (fast, no external dependencies)
make -f Makefile.test test-unit

# Integration tests (requires running services)
make -f Makefile.test test-integration

# With coverage
go test -v -short -run ".*Unit" ./internal/... -timeout=30s -cover
```

### Test Types

- **Unit Tests**: Fast tests with mocks (models, handlers, services)
- **Integration Tests**: Tests with real Redis/PostgreSQL
- **E2E Tests**: Full system testing

The test suite covers job lifecycle, queue operations, error handling, and performance.

## Redis Queue System

Distributed job processing with Redis queues and scalable workers.

### Key Features

- **Queue Types**: Ready, Processing, Completed, Failed, Retry
- **Long-Running Tasks**: Up to 90 seconds duration
- **Auto-Retry**: Exponential backoff for failed jobs
- **Horizontal Scaling**: Scale workers based on demand

### Commands

```bash
# Scale workers
docker-compose up -d --scale worker=10

# Monitor queue
curl http://localhost:8080/queue/stats
```

## Requirements

- Go 1.23+
- Redis 7+
- PostgreSQL 15+
- Docker & Docker Compose (for containerized deployment)
- Dependencies managed via go.mod
