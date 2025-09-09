# Architecture Overview

## System Components

### Core Services
- **API Server**: REST API for job management
- **Scheduler**: Background service for job scheduling
- **Workers**: Scalable job processors
- **PostgreSQL**: Job metadata and execution history
- **Redis**: Message broker and job queue

### Architecture Diagram
```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Client    │───▶│ API Server  │───▶│ PostgreSQL  │
└─────────────┘    └─────────────┘    └─────────────┘
                           │
                           ▼
                   ┌─────────────┐
                   │    Redis    │
                   │   (Queue)   │
                   └─────────────┘
                           │
                           ▼
                   ┌─────────────┐
                   │   Workers   │
                   │ (Scalable)  │
                   └─────────────┘
```

## Job Lifecycle

### 1. Job Creation
1. Client sends POST request to `/api/v1/jobs`
2. API validates job data and CRON schedule
3. Job stored in PostgreSQL
4. Schedule calculated and stored

### 2. Job Scheduling
1. Scheduler polls for ready jobs every 5 seconds
2. Jobs with `nextExecutionTime <= now` are enqueued
3. Job data serialized and pushed to Redis queue

### 3. Job Execution
1. Workers pull jobs from Redis queue
2. HTTP request made to job's API endpoint
3. Execution result stored in PostgreSQL
4. For recurring jobs, next execution time calculated

### 4. Retry Logic
1. Failed jobs moved to retry queue
2. Exponential backoff applied (1s, 2s, 4s, 8s...)
3. Jobs retried up to `maxRetryCount`
4. Permanently failed jobs moved to failed queue

## Queue System

### Queue Types
- **Ready Queue**: Jobs ready for immediate processing
- **Processing Queue**: Jobs currently being executed
- **Retry Queue**: Failed jobs scheduled for retry
- **Completed Queue**: Successfully completed jobs
- **Failed Queue**: Permanently failed jobs

### Redis Data Structures
- **Lists**: Ready and processing queues
- **Sorted Sets**: Retry queue with timestamps
- **Sets**: Completed and failed job tracking
- **Strings**: Job data serialization

## Scaling

### Horizontal Scaling
```bash
# Scale workers based on demand
docker-compose up -d --scale worker=10

# Monitor queue depth
curl http://localhost:8080/queue/stats
```

### Auto-Scaling (Future)
- Queue depth monitoring
- Worker CPU/memory usage
- Failed job rate tracking
- Dynamic worker scaling

## Data Models

### Job
```go
type Job struct {
    ID            uint      `json:"id"`
    Schedule      string    `json:"schedule"`
    API           string    `json:"api"`
    Type          JobType   `json:"type"`
    IsRecurring   bool      `json:"isRecurring"`
    MaxRetryCount int       `json:"maxRetryCount"`
    IsActive      bool      `json:"isActive"`
    CreatedAt     time.Time `json:"createdAt"`
    UpdatedAt     time.Time `json:"updatedAt"`
}
```

### JobSchedule
```go
type JobSchedule struct {
    ID                uint      `json:"id"`
    JobID             uint      `json:"jobId"`
    NextExecutionTime time.Time `json:"nextExecutionTime"`
    CreatedAt         time.Time `json:"createdAt"`
}
```

### JobExecution
```go
type JobExecution struct {
    ID                uint            `json:"id"`
    JobID             uint            `json:"jobId"`
    Status            ExecutionStatus `json:"status"`
    ExecutionTime     time.Time       `json:"executionTime"`
    ExecutionDuration *time.Duration  `json:"executionDuration"`
    RetryCount        int             `json:"retryCount"`
    ErrorMessage      string          `json:"errorMessage"`
    CreatedAt         time.Time       `json:"createdAt"`
    UpdatedAt         time.Time       `json:"updatedAt"`
}
```

## Performance Characteristics

### Throughput
- **Job Creation**: 10k+ jobs/second
- **Job Processing**: 1k+ jobs/second per worker
- **Queue Operations**: 100k+ operations/second

### Latency
- **API Response**: < 10ms
- **Job Scheduling**: < 5 seconds (polling interval)
- **Job Execution**: Up to 90 seconds

### Resource Usage
- **API Server**: 50MB RAM, 0.1 CPU
- **Worker**: 30MB RAM, 0.05 CPU per worker
- **PostgreSQL**: 100MB RAM, 0.2 CPU
- **Redis**: 50MB RAM, 0.1 CPU

## Security

### Network Security
- Internal service communication via Docker network
- Only necessary ports exposed to host
- Database not directly accessible from host

### Application Security
- Non-root user in containers
- Minimal Alpine Linux base images
- Proper file permissions
- Resource limits for containers

### Data Security
- Environment variable configuration
- Secure password handling
- SSL support for database connections
- API key authentication
