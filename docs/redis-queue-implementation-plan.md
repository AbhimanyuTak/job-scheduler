# Redis Queue Implementation Plan

## Overview

This document outlines the plan to migrate from the current in-memory worker pool to a Redis-based queue system for handling long-running job tasks (up to 90 seconds).

## Current System Analysis

### Current Architecture
- Single container running both API server and job workers
- In-memory worker pool with 500 concurrent workers
- 10-second HTTP timeout (insufficient for 90-second tasks)
- Polling-based job discovery (100ms intervals)
- Vertical scaling only

### Limitations
- Cannot handle tasks longer than 10 seconds
- Single point of failure
- Resource contention between API and workers
- No horizontal scaling capability
- Fixed resource allocation

## Target Architecture

### Components
1. **API Server Container**: Handles HTTP requests, job management
2. **Worker Containers**: Process job tasks (scalable)
3. **Redis**: Message broker and job queue
4. **PostgreSQL**: Job metadata and execution history

### Benefits
- Handle 90-second tasks without timeout issues
- Horizontal scaling of workers
- Resource isolation between API and workers
- Better fault tolerance
- Auto-scaling capabilities

## Implementation Steps

### Phase 1: Infrastructure Setup

#### 1.1 Add Redis to Docker Compose
- Add Redis service to `docker-compose.yml` and `docker-compose.dev.yml`
- Configure Redis persistence and memory settings
- Set up Redis health checks

#### 1.2 Update Dependencies
- Add Redis client library to `go.mod`
- Add Redis connection management

#### 1.3 Environment Configuration
- Add Redis connection parameters to environment variables
- Update configuration loading in main.go

### Phase 2: Core Queue Implementation

#### 2.1 Redis Client Service
- Create `internal/services/redis_client.go`
- Implement connection management
- Add connection pooling and retry logic

#### 2.2 Job Queue Service
- Create `internal/services/job_queue.go`
- Implement job enqueueing logic
- Add job dequeueing for workers
- Implement job status tracking

#### 2.3 Queue Models
- Create `internal/models/queue_job.go`
- Define queue job structure
- Add serialization/deserialization methods

### Phase 3: Worker Implementation

#### 3.1 Worker Service
- Create `internal/services/worker.go`
- Implement job processing logic
- Add worker health monitoring
- Implement graceful shutdown

#### 3.2 Worker Main Application
- Create `cmd/worker/main.go`
- Separate worker entry point
- Add worker-specific configuration

#### 3.3 Worker Docker Configuration
- Create separate Dockerfile for workers
- Update docker-compose files with worker service
- Configure worker scaling

### Phase 4: API Server Updates

#### 4.1 Job Creation Updates
- Modify job creation to enqueue jobs instead of direct execution
- Update job status handling
- Add queue status endpoints

#### 4.2 Remove Old Worker Pool
- Remove in-memory worker pool from scheduler service
- Update background scheduler to only handle scheduling
- Remove direct job execution logic

#### 4.3 HTTP Client Updates
- Increase HTTP timeout to 90+ seconds for workers
- Add retry logic for failed API calls
- Implement proper error handling

### Phase 5: Monitoring and Observability

#### 5.1 Queue Metrics
- Add Redis queue depth monitoring
- Implement worker performance metrics
- Add job processing time tracking

#### 5.2 Health Checks
- Add Redis health check endpoint
- Implement worker health monitoring
- Add queue health status

#### 5.3 Logging
- Add structured logging for queue operations
- Implement worker-specific logging
- Add job execution tracing

### Phase 6: Auto-scaling and Production Readiness

#### 6.1 Auto-scaling Logic
- Implement queue depth-based scaling
- Add worker performance-based scaling
- Configure scaling policies

#### 6.2 Production Configuration
- Optimize Redis settings for production
- Configure worker resource limits
- Set up monitoring and alerting

## Technical Implementation Details

### Redis Queue Design

#### Queue Structure
```
job_queue:ready          # List of ready jobs
job_queue:processing     # Set of jobs being processed
job_queue:failed         # List of failed jobs
job_queue:completed      # List of completed jobs
```

#### Job Message Format
```json
{
  "id": "job_123",
  "job_id": 456,
  "api_url": "https://api.example.com/endpoint",
  "max_retry_count": 3,
  "retry_count": 0,
  "created_at": "2024-01-01T00:00:00Z",
  "scheduled_at": "2024-01-01T00:00:00Z",
  "timeout": 90
}
```

### Worker Scaling Strategy

#### Metrics for Scaling
- Queue depth (number of pending jobs)
- Average processing time
- Worker CPU/memory usage
- Failed job rate

#### Scaling Triggers
- Scale up when queue depth > threshold
- Scale down when queue depth < threshold and workers idle
- Scale up when average processing time increases
- Scale down when workers are underutilized

### Error Handling and Retry Logic

#### Retry Strategy
- Exponential backoff for failed jobs
- Dead letter queue for permanently failed jobs
- Job timeout handling
- Worker failure recovery

#### Failure Scenarios
- Worker container crashes
- Redis connection failures
- API endpoint failures
- Network timeouts

## Migration Strategy

### Rollout Plan
1. **Phase 1**: Deploy Redis infrastructure alongside current system
2. **Phase 2**: Implement queue system in parallel
3. **Phase 3**: Deploy workers alongside current workers
4. **Phase 4**: Gradually migrate jobs to queue system
5. **Phase 5**: Remove old worker pool system
6. **Phase 6**: Optimize and scale

### Rollback Plan
- Keep current system running during migration
- Feature flag to switch between systems
- Database schema compatibility
- Gradual traffic migration

## Configuration Changes

### Environment Variables
```bash
# Redis Configuration
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Worker Configuration
WORKER_POOL_SIZE=10
WORKER_MAX_CONCURRENT_JOBS=5
WORKER_HTTP_TIMEOUT=90s
WORKER_RETRY_COUNT=3

# Queue Configuration
QUEUE_BATCH_SIZE=100
QUEUE_POLL_INTERVAL=1s
QUEUE_MAX_RETRIES=3
```

### Docker Compose Updates
```yaml
services:
  redis:
    image: redis:7-alpine
    container_name: job-scheduler-redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  worker:
    build: 
      context: .
      dockerfile: Dockerfile.worker
    container_name: job-scheduler-worker
    environment:
      # Worker-specific environment variables
    depends_on:
      redis:
        condition: service_healthy
      postgres:
        condition: service_healthy
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: '0.5'
          memory: 512M
```

## Testing Strategy

### Unit Tests
- Queue service tests
- Worker service tests
- Redis client tests
- Job processing tests

### Integration Tests
- End-to-end job processing
- Worker scaling tests
- Failure recovery tests
- Performance tests

### Load Tests
- High-throughput job processing
- Worker scaling under load
- Redis performance under load
- Long-running task handling

## Performance Considerations

### Redis Optimization
- Connection pooling
- Pipeline operations
- Memory optimization
- Persistence configuration

### Worker Optimization
- Resource limits
- Concurrent job limits
- HTTP client optimization
- Memory management

### Monitoring
- Queue depth monitoring
- Worker performance metrics
- Redis performance metrics
- Job success/failure rates

## Security Considerations

### Redis Security
- Authentication and authorization
- Network security
- Data encryption
- Access control

### Worker Security
- Container security
- Network isolation
- Resource limits
- Process isolation

## Maintenance and Operations

### Monitoring
- Queue health monitoring
- Worker health monitoring
- Redis performance monitoring
- Job execution monitoring

### Maintenance
- Redis backup and recovery
- Worker deployment strategies
- Configuration management
- Log management

### Troubleshooting
- Common issues and solutions
- Debugging procedures
- Performance tuning
- Scaling guidelines

## Success Metrics

### Performance Metrics
- Job processing throughput
- Average job processing time
- Worker utilization
- Queue processing latency

### Reliability Metrics
- Job success rate
- Worker failure rate
- Queue availability
- System uptime

### Scalability Metrics
- Auto-scaling effectiveness
- Resource utilization
- Cost efficiency
- Response time under load

## Timeline Estimate

- **Phase 1**: 1-2 days (Infrastructure setup)
- **Phase 2**: 3-4 days (Core queue implementation)
- **Phase 3**: 2-3 days (Worker implementation)
- **Phase 4**: 2-3 days (API server updates)
- **Phase 5**: 1-2 days (Monitoring)
- **Phase 6**: 2-3 days (Auto-scaling and production)

**Total Estimated Time**: 11-17 days

## Risks and Mitigation

### Technical Risks
- Redis performance bottlenecks
- Worker scaling issues
- Data consistency problems
- Network connectivity issues

### Mitigation Strategies
- Thorough testing and load testing
- Gradual rollout with monitoring
- Fallback mechanisms
- Comprehensive error handling

## Conclusion

This Redis queue implementation will provide a robust, scalable solution for handling long-running job tasks while maintaining high availability and performance. The phased approach ensures minimal disruption to the current system while providing a clear migration path.
