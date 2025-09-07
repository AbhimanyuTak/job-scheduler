# Job Scheduler

A distributed job scheduler built with Go that supports CRON-based scheduling with extended second-level precision.

## Features

- **Extended CRON Support**: Parse CRON expressions with seconds (e.g., "31 10-15 1 * * MON-FRI")
- **Execution Types**: Support for AT_LEAST_ONCE and AT_MOST_ONCE execution guarantees
- **Job Management**: Create, delete, and track job execution history
- **High Performance**: Designed to handle 10k+ jobs per second
- **Fault Tolerant**: Built-in retry logic and error handling

## Project Structure

```
job_scheduler/
├── cmd/
│   └── scheduler/          # Main application entry point
├── internal/
│   ├── models/             # Data models (Job, JobExecution)
│   ├── storage/            # In-memory storage layer
│   ├── scheduler/          # Core scheduling logic
│   └── api/                # HTTP API handlers
├── go.mod
└── README.md
```

## Quick Start

### Option 1: Docker (Recommended)

1. **Configure the application**:
   ```bash
   # Edit config.env with your database credentials
   nano config.env
   ```

2. **Start the database**:
   ```bash
   ./scripts/docker-setup.sh
   ```

3. **Run the application**:
   ```bash
   go run ./cmd/scheduler
   ```

4. **Test health endpoint**:
   ```bash
   curl http://localhost:8080/health
   ```

### Option 2: Manual Setup

1. **Setup PostgreSQL database** (see [Database Setup Guide](docs/database-setup.md))

2. **Build the application**:
   ```bash
   go build ./cmd/scheduler
   ```

3. **Run the server**:
   ```bash
   go run ./cmd/scheduler
   ```

4. **Test health endpoint**:
   ```bash
   curl http://localhost:8080/health
   ```

## Configuration

- **[Configuration Guide](docs/configuration.md)** - Environment variables and setup
- **[Database Setup Guide](docs/database-setup.md)** - PostgreSQL setup instructions

## API Documentation

- **[OpenAPI Specification](docs/api-spec.yaml)** - Complete API specification
- **[API Endpoints](docs/api-endpoints.md)** - Quick reference for all endpoints
- **[Postman Collection](docs/postman-collection.json)** - Import into Postman for testing

### Key Endpoints

- `GET /health` - Health check endpoint
- `POST /api/v1/jobs` - Create a new job
- `GET /api/v1/jobs` - List all jobs
- `GET /api/v1/jobs/{id}` - Get job details
- `PATCH /api/v1/jobs/{id}` - Update job
- `DELETE /api/v1/jobs/{id}` - Soft delete job (can be reactivated by updating isActive)
- `GET /api/v1/jobs/{id}/schedule` - Get job schedule
- `GET /api/v1/jobs/{id}/history` - Get job execution history

## Development Status

✅ **Hour 1 Complete**: Project setup, basic models, storage layer, and HTTP server foundation

**Next Steps**:
- Hour 2: CRON parser implementation
- Hour 3: Storage layer completion
- Hour 4: HTTP server setup with middleware
- Hour 5: Job creation API
- Hour 6: Job management APIs
- Hour 7: Scheduler engine
- Hour 8: Integration & testing

## Requirements

- Go 1.21+
- Dependencies managed via go.mod
