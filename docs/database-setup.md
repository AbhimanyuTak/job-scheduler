# Database Setup Guide

## Quick Start with Docker (Recommended)

### 1. Development Setup
```bash
# Start PostgreSQL database only
./scripts/docker-setup.sh

# Or manually with docker-compose
docker-compose -f docker-compose.dev.yml up -d postgres
```

### 2. Full Application Stack
```bash
# Run entire application with Docker
./scripts/docker-run.sh

# Or manually with docker-compose
docker-compose up --build -d
```

## Manual PostgreSQL Setup

### 1. Install PostgreSQL
```bash
# macOS with Homebrew
brew install postgresql
brew services start postgresql

# Ubuntu/Debian
sudo apt-get install postgresql postgresql-contrib
sudo systemctl start postgresql

# Docker (standalone)
docker run --name postgres-job-scheduler \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=job_scheduler \
  -p 5432:5432 \
  -d postgres:15
```

### 2. Create Database
```sql
-- Connect to PostgreSQL
psql -U postgres

-- Create database
CREATE DATABASE job_scheduler;

-- Create user (optional)
CREATE USER scheduler_user WITH PASSWORD 'password';
GRANT ALL PRIVILEGES ON DATABASE job_scheduler TO scheduler_user;
```

### 3. Environment Variables
Set the following environment variables or create a `.env` file:

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=password
export DB_NAME=job_scheduler
export DB_SSLMODE=disable
```

### 4. Run the Application
```bash
# The application will auto-migrate the database schema
go run ./cmd/scheduler
```

## Database Schema

The application uses GORM auto-migration to create the following tables:

- `jobs` - Job configuration and metadata (with `is_active` field for soft deletion)
- `job_schedules` - Next execution times for jobs
- `job_executions` - Execution history and status tracking

### Soft Delete Feature

Jobs support soft deletion using the `is_active` field:
- When a job is "deleted", `is_active` is set to `false`
- Only active jobs (`is_active = true`) are returned in queries
- Jobs can be reactivated using the `ActivateJob` method
- This preserves job history and execution records

## Docker Services

### Available Services

1. **PostgreSQL Database** (`postgres`)
   - Port: 5432
   - Database: job_scheduler
   - Username: postgres
   - Password: password

2. **Job Scheduler Application** (`scheduler`)
   - Port: 8080
   - URL: http://localhost:8080
   - Health: http://localhost:8080/health

### Docker Commands

```bash
# Start development database only
docker-compose -f docker-compose.dev.yml up -d postgres

# Start full application stack
docker-compose up --build -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down

# Stop and remove volumes (clean slate)
docker-compose down -v
```

## Testing Database Connection

```bash
# Test connection
psql -h localhost -U postgres -d job_scheduler -c "SELECT version();"

# Check tables
psql -h localhost -U postgres -d job_scheduler -c "\dt"

# Test with Docker
docker-compose exec postgres psql -U postgres -d job_scheduler -c "SELECT version();"
```
