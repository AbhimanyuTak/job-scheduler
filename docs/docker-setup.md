# Docker Setup Guide

This guide covers the complete Docker setup for the Job Scheduler application, including both production and development configurations.

## Overview

The Job Scheduler application is fully containerized using Docker Compose with the following architecture:

- **PostgreSQL Database**: Persistent data storage with health monitoring
- **Job Scheduler App**: Go application with REST API and background scheduler
- **Multi-environment Support**: Separate configurations for development and production

## Quick Start

### Production Deployment

```bash
# Clone the repository
git clone <repository-url>
cd job_scheduler

# Start the complete application stack
docker-compose up -d

# Check service status
docker-compose ps

# View application logs
docker-compose logs -f scheduler

# Test the health endpoint
curl http://localhost:8080/health
```

### Development Setup

```bash
# Start development environment with live reloading
docker-compose -f docker-compose.dev.yml up -d

# View development logs
docker-compose -f docker-compose.dev.yml logs -f scheduler

# Stop development environment
docker-compose -f docker-compose.dev.yml down
```

## Docker Compose Configurations

### Production (`docker-compose.yml`)

**Features:**
- Optimized for production performance
- No volume mounts for application code
- Production logging levels (`info`)
- Restart policies for high availability
- Health checks for both services

**Services:**
- `postgres`: PostgreSQL 15 Alpine with persistent storage
- `scheduler`: Job Scheduler application

### Development (`docker-compose.dev.yml`)

**Features:**
- Live code reloading via volume mounts
- Debug logging enabled
- Separate database volume for development
- Hot reloading for faster development cycles

**Services:**
- `postgres`: PostgreSQL 15 Alpine with dev-specific volume
- `scheduler`: Job Scheduler application with development settings

## Environment Variables

### Database Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `postgres` | Database hostname (use `postgres` in Docker) |
| `DB_PORT` | `5432` | Database port |
| `DB_USER` | `postgres` | Database username |
| `DB_PASSWORD` | `password` | Database password |
| `DB_NAME` | `job_scheduler` | Database name |
| `DB_SSLMODE` | `disable` | SSL connection mode |

### Server Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | Application port |
| `SERVER_HOST` | `0.0.0.0` | Bind address |
| `GIN_MODE` | `release` (prod) / `debug` (dev) | Gin framework mode |

### Logging Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `LOG_LEVEL` | `info` (prod) / `debug` (dev) | Log level |
| `LOG_FORMAT` | `json` | Log format |

### Scheduler Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `SCHEDULER_POLL_INTERVAL` | `5s` | Job polling interval |
| `SCHEDULER_BATCH_SIZE` | `100` | Max jobs per batch |
| `SCHEDULER_HTTP_TIMEOUT` | `30s` | HTTP request timeout |

### Environment

| Variable | Default | Description |
|----------|---------|-------------|
| `ENVIRONMENT` | `production` (prod) / `development` (dev) | Environment type |

## Docker Commands Reference

### Basic Operations

```bash
# Build the application image
docker-compose build

# Start services in background
docker-compose up -d

# Start with build (rebuild if needed)
docker-compose up -d --build

# View logs
docker-compose logs -f [service-name]

# Stop services
docker-compose down

# Stop and remove volumes (WARNING: deletes data)
docker-compose down -v
```

### Service Management

```bash
# Check service status
docker-compose ps

# Restart a specific service
docker-compose restart scheduler

# Scale services (if supported)
docker-compose up -d --scale scheduler=2

# Execute commands in running container
docker-compose exec scheduler /bin/sh
docker-compose exec postgres psql -U postgres -d job_scheduler
```

### Development Commands

```bash
# Start development environment
docker-compose -f docker-compose.dev.yml up -d

# View development logs
docker-compose -f docker-compose.dev.yml logs -f

# Stop development environment
docker-compose -f docker-compose.dev.yml down

# Rebuild development environment
docker-compose -f docker-compose.dev.yml up -d --build
```

## Health Checks

### PostgreSQL Health Check

```bash
# Check database connectivity
docker-compose exec postgres pg_isready -U postgres

# Connect to database
docker-compose exec postgres psql -U postgres -d job_scheduler
```

### Application Health Check

```bash
# HTTP health check
curl http://localhost:8080/health

# Expected response
{"status": "healthy"}
```

## Data Persistence

### Production Data

- **Volume**: `postgres_data`
- **Location**: Docker managed volume
- **Backup**: Use `docker volume` commands

```bash
# Backup production data
docker run --rm -v postgres_data:/data -v $(pwd):/backup alpine tar czf /backup/postgres_backup.tar.gz -C /data .

# Restore production data
docker run --rm -v postgres_data:/data -v $(pwd):/backup alpine tar xzf /backup/postgres_backup.tar.gz -C /data
```

### Development Data

- **Volume**: `postgres_dev_data`
- **Purpose**: Separate from production data
- **Reset**: Can be safely deleted for fresh start

## Security Features

### Application Container

- **Base Image**: Alpine Linux (minimal attack surface)
- **User**: Non-root user (`appuser`)
- **Packages**: Only necessary packages installed
- **Permissions**: Proper file ownership and permissions

### Network Security

- **Internal Communication**: Services communicate via Docker network
- **Port Exposure**: Only necessary ports exposed to host
- **Database Access**: Not directly accessible from host in production

## Troubleshooting

### Common Issues

#### Service Won't Start

```bash
# Check logs for errors
docker-compose logs scheduler
docker-compose logs postgres

# Check service status
docker-compose ps

# Restart services
docker-compose restart
```

#### Database Connection Issues

```bash
# Verify database is running
docker-compose exec postgres pg_isready -U postgres

# Check database logs
docker-compose logs postgres

# Test connection from application container
docker-compose exec scheduler wget -qO- http://postgres:5432
```

#### Port Conflicts

```bash
# Check if ports are in use
netstat -tulpn | grep :8080
netstat -tulpn | grep :5432

# Use different ports in docker-compose.yml
ports:
  - "8081:8080"  # Use port 8081 on host
```

#### Volume Issues

```bash
# List volumes
docker volume ls

# Inspect volume
docker volume inspect job_scheduler_postgres_data

# Remove problematic volume (WARNING: deletes data)
docker volume rm job_scheduler_postgres_data
```

### Performance Tuning

#### Database Performance

```yaml
# Add to postgres service in docker-compose.yml
environment:
  POSTGRES_SHARED_BUFFERS: 256MB
  POSTGRES_EFFECTIVE_CACHE_SIZE: 1GB
  POSTGRES_MAINTENANCE_WORK_MEM: 64MB
```

#### Application Performance

```yaml
# Add to scheduler service in docker-compose.yml
deploy:
  resources:
    limits:
      memory: 512M
      cpus: '0.5'
    reservations:
      memory: 256M
      cpus: '0.25'
```

## Monitoring

### Log Monitoring

```bash
# Follow all logs
docker-compose logs -f

# Follow specific service logs
docker-compose logs -f scheduler

# View last 100 lines
docker-compose logs --tail=100 scheduler
```

### Resource Monitoring

```bash
# Check container resource usage
docker stats

# Check specific container
docker stats job-scheduler-app
```

## Backup and Recovery

### Database Backup

```bash
# Create backup
docker-compose exec postgres pg_dump -U postgres job_scheduler > backup.sql

# Restore backup
docker-compose exec -T postgres psql -U postgres job_scheduler < backup.sql
```

### Volume Backup

```bash
# Backup volume
docker run --rm -v job_scheduler_postgres_data:/data -v $(pwd):/backup alpine tar czf /backup/postgres_data_backup.tar.gz -C /data .

# Restore volume
docker run --rm -v job_scheduler_postgres_data:/data -v $(pwd):/backup alpine tar xzf /backup/postgres_data_backup.tar.gz -C /data
```

## Production Deployment

### Environment-Specific Configuration

1. **Create production environment file**:
   ```bash
   cp config.env.example config.env.prod
   # Edit config.env.prod with production values
   ```

2. **Use environment file in docker-compose**:
   ```yaml
   services:
     scheduler:
       env_file:
         - config.env.prod
   ```

### Security Considerations

- Change default database passwords
- Use secrets management for sensitive data
- Enable SSL for database connections
- Use reverse proxy (nginx/traefik) for production
- Implement proper logging and monitoring
- Regular security updates

### Scaling Considerations

- Use external PostgreSQL for production
- Implement load balancing for multiple app instances
- Use container orchestration (Kubernetes) for large deployments
- Implement proper monitoring and alerting
