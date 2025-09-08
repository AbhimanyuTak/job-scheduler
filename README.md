# Job Scheduler

A distributed job scheduler built with Go that supports CRON-based scheduling with extended second-level precision and Redis-based queue processing for long-running tasks.

## Features

- **Extended CRON Support**: Parse CRON expressions with seconds (e.g., "31 10-15 1 * * MON-FRI")
- **Execution Types**: Support for AT_LEAST_ONCE and AT_MOST_ONCE execution guarantees
- **Job Management**: Create, delete, and track job execution history
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

### Option 1: Docker (Recommended)

#### Production Setup

1. **Start the complete application stack**:
   ```bash
   # Build and start all services (PostgreSQL + Redis + API Server + Workers)
   docker-compose up -d
   
   # View logs
   docker-compose logs -f scheduler
   docker-compose logs -f worker
   ```

2. **Test the application**:
   ```bash
   # Health check
   curl http://localhost:8080/health
   
   # Check queue statistics
   curl http://localhost:8080/queue/stats
   
   # Create a test job
   curl -X POST http://localhost:8080/api/v1/jobs \
     -H "Content-Type: application/json" \
     -d '{
       "name": "Test Job",
       "description": "A test job",
       "cronExpression": "0 */5 * * * *",
       "httpMethod": "POST",
       "httpUrl": "https://httpbin.org/post",
       "executionType": "AT_LEAST_ONCE"
     }'
   ```

3. **Scale workers**:
   ```bash
   # Scale workers to 5 replicas
   docker-compose up -d --scale worker=5
   
   # Check running services
   docker-compose ps
   ```

4. **Stop the application**:
   ```bash
   docker-compose down
   ```

#### Development Setup

1. **Start development environment**:
   ```bash
   # Start with live reloading and debug mode
   docker-compose -f docker-compose.dev.yml up -d
   
   # View logs
   docker-compose -f docker-compose.dev.yml logs -f scheduler
   docker-compose -f docker-compose.dev.yml logs -f worker
   ```

2. **Development features**:
   - Live code reloading (volume mount)
   - Debug logging enabled
   - Hot reloading for faster development
   - 2 worker replicas for testing

3. **Stop development environment**:
   ```bash
   docker-compose -f docker-compose.dev.yml down
   ```

#### Docker Commands Reference

```bash
# Build the application images
docker-compose build

# Start services in background
docker-compose up -d

# View logs
docker-compose logs -f [service-name]

# Scale workers
docker-compose up -d --scale worker=5

# Stop services
docker-compose down

# Remove volumes (WARNING: deletes database and Redis data)
docker-compose down -v

# Rebuild and restart
docker-compose up -d --build

# Check service status
docker-compose ps

# Execute commands in running containers
docker-compose exec scheduler /bin/sh
docker-compose exec worker /bin/sh
docker-compose exec redis redis-cli
```

### Option 2: Manual Setup

1. **Setup PostgreSQL database** (see [Database Setup Guide](docs/database-setup.md))

2. **Setup Redis server**:
   ```bash
   # Install Redis (Ubuntu/Debian)
   sudo apt-get install redis-server
   
   # Start Redis
   sudo systemctl start redis-server
   ```

3. **Build the applications**:
   ```bash
   # Build API server
   go build ./cmd/scheduler
   
   # Build worker
   go build ./cmd/worker
   ```

4. **Run the services**:
   ```bash
   # Terminal 1: Run API server
   go run ./cmd/scheduler
   
   # Terminal 2: Run worker
   go run ./cmd/worker
   ```

5. **Test endpoints**:
   ```bash
   # Health check
   curl http://localhost:8080/health
   
   # Queue statistics
   curl http://localhost:8080/queue/stats
   ```

## Docker Configuration

### Architecture

The application uses a distributed microservices architecture:

- **PostgreSQL Database**: Persistent data storage with health checks
- **Redis**: Message broker and job queue with persistence
- **API Server**: Go application with REST API and job scheduling
- **Worker Containers**: Scalable job processors (3 replicas in production, 2 in dev)
- **Health Monitoring**: Built-in health checks for all services

### Environment Variables

The Docker setup supports all configuration options from `config.env.example`:

#### Database Configuration
- `DB_HOST`: Database host (default: `postgres` in Docker)
- `DB_PORT`: Database port (default: `5432`)
- `DB_USER`: Database user (default: `postgres`)
- `DB_PASSWORD`: Database password (default: `password`)
- `DB_NAME`: Database name (default: `job_scheduler`)
- `DB_SSLMODE`: SSL mode (default: `disable`)

#### Redis Configuration
- `REDIS_HOST`: Redis host (default: `redis` in Docker)
- `REDIS_PORT`: Redis port (default: `6379`)
- `REDIS_PASSWORD`: Redis password (default: empty)
- `REDIS_DB`: Redis database number (default: `0`)

#### Server Configuration
- `SERVER_PORT`: Application port (default: `8080`)
- `SERVER_HOST`: Bind address (default: `0.0.0.0`)
- `GIN_MODE`: Gin framework mode (`debug` for dev, `release` for prod)

#### Worker Configuration
- `WORKER_POOL_SIZE`: Number of concurrent workers per container (default: `10`)
- `WORKER_HTTP_TIMEOUT`: HTTP request timeout in seconds (default: `90`)

#### Logging Configuration
- `LOG_LEVEL`: Log level (`debug`, `info`, `warn`, `error`)
- `LOG_FORMAT`: Log format (`json` or `text`)

#### Scheduler Configuration
- `SCHEDULER_POLL_INTERVAL`: How often to check for jobs (default: `5s`)
- `SCHEDULER_BATCH_SIZE`: Max jobs to process per batch (default: `100`)
- `SCHEDULER_HTTP_TIMEOUT`: HTTP request timeout (default: `30s`)

#### Environment
- `ENVIRONMENT`: Environment type (`development` or `production`)

### Docker Compose Files

- **`docker-compose.yml`**: Production configuration
  - Optimized for performance
  - No volume mounts for code
  - Production logging levels
  - Restart policies
  - 3 worker replicas with resource limits

- **`docker-compose.dev.yml`**: Development configuration
  - Live code reloading
  - Debug logging
  - Volume mounts for development
  - Separate database and Redis volumes
  - 2 worker replicas for testing

### Health Checks

All services include health checks:

- **PostgreSQL**: Uses `pg_isready` to verify database connectivity
- **Redis**: Uses `redis-cli ping` to verify Redis connectivity
- **API Server**: HTTP health check at `/health` endpoint
- **Workers**: Process health check using `pgrep`

### Data Persistence

- **Production**: 
  - `postgres_data` volume for database persistence
  - `redis_data` volume for Redis persistence
- **Development**: 
  - `postgres_dev_data` volume (separate from production)
  - `redis_dev_data` volume (separate from production)

### Security Features

- Non-root user in all application containers
- Minimal Alpine Linux base images
- No unnecessary packages or tools
- Proper file permissions
- Resource limits for worker containers
- Network isolation between services

## Configuration

- **[Configuration Guide](docs/configuration.md)** - Environment variables and setup
- **[Database Setup Guide](docs/database-setup.md)** - PostgreSQL setup instructions
- **[Docker Setup Guide](docs/docker-setup.md)** - Complete Docker deployment guide

## API Documentation

- **[OpenAPI Specification](docs/api-spec.yaml)** - Complete API specification
- **[API Endpoints](docs/api-endpoints.md)** - Quick reference for all endpoints
- **[Postman Collection](docs/postman-collection.json)** - Import into Postman for testing

### Key Endpoints

- `GET /health` - Health check endpoint
- `GET /queue/stats` - Queue statistics and monitoring
- `POST /api/v1/jobs` - Create a new job
- `GET /api/v1/jobs` - List all jobs
- `GET /api/v1/jobs/{id}` - Get job details
- `PATCH /api/v1/jobs/{id}` - Update job
- `DELETE /api/v1/jobs/{id}` - Soft delete job (can be reactivated by updating isActive)
- `GET /api/v1/jobs/{id}/schedule` - Get job schedule
- `GET /api/v1/jobs/{id}/history` - Get job execution history

## Redis Queue System

### Architecture Overview

The job scheduler uses a Redis-based queue system for distributed job processing:

1. **API Server**: Receives job requests and enqueues them in Redis
2. **Redis Queue**: Stores jobs in different queues (ready, processing, completed, failed)
3. **Worker Containers**: Process jobs from the queue with configurable concurrency
4. **PostgreSQL**: Stores job metadata and execution history

### Queue Types

- **Ready Queue**: Jobs ready for immediate processing
- **Processing Queue**: Jobs currently being executed by workers
- **Completed Queue**: Successfully completed jobs (last 1000)
- **Failed Queue**: Permanently failed jobs
- **Retry Queue**: Jobs scheduled for retry with exponential backoff

### Scaling

#### Horizontal Scaling
```bash
# Scale workers based on demand
docker-compose up -d --scale worker=10

# Monitor queue depth
curl http://localhost:8080/queue/stats
```

#### Auto-Scaling (Future)
The system is designed to support auto-scaling based on:
- Queue depth (number of pending jobs)
- Average processing time
- Worker CPU/memory usage
- Failed job rate

### Long-Running Tasks

The system supports tasks up to 90 seconds duration:
- Workers have configurable HTTP timeout (default: 90s)
- Jobs persist in Redis during processing
- Failed jobs are automatically retried with exponential backoff
- Dead letter queue for permanently failed jobs

### Monitoring

#### Queue Statistics
```bash
# Get real-time queue statistics
curl http://localhost:8080/queue/stats

# Response example:
{
  "queue_stats": {
    "ready": 5,
    "processing": 3,
    "completed": 150,
    "retrying": 2
  }
}
```

#### Health Monitoring
- Redis connectivity monitoring
- Worker process health checks
- Queue depth alerts (configurable)
- Job success/failure rates

## Requirements

- Go 1.23+
- Redis 7+
- PostgreSQL 15+
- Docker & Docker Compose (for containerized deployment)
- Dependencies managed via go.mod
