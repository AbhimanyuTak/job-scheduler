# Database Schema Design - Version 1.0

## Overview
This document outlines the PostgreSQL database schema for the distributed job scheduler, designed to support high-throughput job scheduling with proper durability and consistency.

## Database Tables

### 1. Job Table
Stores the job configuration and metadata.

```sql
CREATE TABLE jobs (
    id SERIAL PRIMARY KEY,
    schedule VARCHAR(100) NOT NULL,           -- Unix CRON format with seconds
    api TEXT NOT NULL,                        -- HTTP endpoint to POST to
    type VARCHAR(20) NOT NULL,                -- ATLEAST_ONCE, ATMOST_ONCE
    is_recurring BOOLEAN NOT NULL DEFAULT true,
    is_active BOOLEAN NOT NULL DEFAULT true,  -- Soft delete flag
    description TEXT,
    max_retry_count INTEGER NOT NULL DEFAULT 3,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

### 2. JobSchedule Table
Stores the next execution times for jobs (optimized for scheduler queries).

```sql
CREATE TABLE job_schedules (
    id SERIAL PRIMARY KEY,
    job_id INTEGER NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    next_execution_time TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT unique_job_schedule UNIQUE(job_id)
);

-- Critical index for scheduler performance
CREATE INDEX idx_job_schedules_next_execution ON job_schedules(next_execution_time);
CREATE INDEX idx_job_schedules_job_id ON job_schedules(job_id);
```

### 3. JobExecution Table
Stores execution history and status tracking.

```sql
CREATE TABLE job_executions (
    id SERIAL PRIMARY KEY,
    job_id INTEGER NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL,              -- SCHEDULED, RUNNING, FAILED, SUCCESS
    error TEXT,
    execution_time TIMESTAMP WITH TIME ZONE NOT NULL,
    execution_duration INTERVAL,
    retry_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_job_executions_job_id ON job_executions(job_id);
CREATE INDEX idx_job_executions_status ON job_executions(status);
CREATE INDEX idx_job_executions_execution_time ON job_executions(execution_time);
```

## Updated Go Models

### Job Model
```go
type Job struct {
    ID            uint      `json:"id" gorm:"primaryKey"`
    Schedule      string    `json:"schedule" gorm:"size:100;not null"`
    API           string    `json:"api" gorm:"type:text;not null"`
    Type          JobType   `json:"type" gorm:"size:20;not null"`
    IsRecurring   bool      `json:"isRecurring" gorm:"default:true"`
    IsActive      bool      `json:"isActive" gorm:"default:true;index"`
    Description   string    `json:"description" gorm:"type:text"`
    MaxRetryCount int       `json:"maxRetryCount" gorm:"default:3"`
    CreatedAt     time.Time `json:"createdAt"`
    UpdatedAt     time.Time `json:"updatedAt"`
    DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}
```

### JobSchedule Model
```go
type JobSchedule struct {
    ID                uint           `json:"id" gorm:"primaryKey"`
    JobID             uint           `json:"jobId" gorm:"not null;uniqueIndex;index"`
    NextExecutionTime time.Time      `json:"nextExecutionTime" gorm:"not null;index"`
    CreatedAt         time.Time      `json:"createdAt"`
    DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
}
```

### JobExecution Model
```go
type JobExecution struct {
    ID                uint            `json:"id" gorm:"primaryKey"`
    JobID             uint            `json:"jobId" gorm:"not null;index"`
    Status            ExecutionStatus `json:"status" gorm:"size:20;not null;index"`
    Error             string          `json:"error,omitempty" gorm:"type:text"`
    ExecutionTime     time.Time       `json:"executionTime" gorm:"not null;index"`
    ExecutionDuration *time.Duration  `json:"executionDuration,omitempty"`
    RetryCount        int             `json:"retryCount" gorm:"default:0"`
    CreatedAt         time.Time       `json:"createdAt"`
    UpdatedAt         time.Time       `json:"updatedAt"`
    DeletedAt         gorm.DeletedAt  `json:"-" gorm:"index"`
}
```

## Database Operations

### Key Queries for Performance

1. **Get Jobs Ready for Execution** (Critical for scheduler performance):
```sql
SELECT js.job_id, j.schedule, j.api, j.type, j.max_retry_count, js.next_execution_time
FROM job_schedules js
JOIN jobs j ON js.job_id = j.id
WHERE js.next_execution_time <= NOW()
ORDER BY js.next_execution_time ASC
LIMIT 1000;
```

2. **Update Next Execution Time**:
```sql
UPDATE job_schedules 
SET next_execution_time = $1 
WHERE job_id = $2;
```

3. **Get Job Execution History**:
```sql
SELECT * FROM job_executions 
WHERE job_id = $1 
ORDER BY created_at DESC 
LIMIT $2;
```

4. **Create New Job** (returns the generated ID):
```sql
INSERT INTO jobs (schedule, api, type, is_recurring, description, max_retry_count)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;
```

5. **Create Job Schedule**:
```sql
INSERT INTO job_schedules (job_id, next_execution_time)
VALUES ($1, $2)
RETURNING id;
```

6. **Create Job Execution**:
```sql
INSERT INTO job_executions (job_id, status, execution_time, retry_count)
VALUES ($1, $2, $3, $4)
RETURNING id;
```

## Migration Strategy

### Version 1.0 Migration
```sql
-- Create tables with SERIAL primary keys
-- Add only essential indexes for performance
-- Application-level validation for constraints
```

## Performance Considerations

1. **Integer Primary Keys**: Using SERIAL (auto-increment) integers for better performance and smaller index sizes
2. **Essential Indexes Only**: Only indexes critical for scheduler performance
3. **Application-Level Validation**: Constraints handled in Go code for flexibility
4. **Connection Pooling**: Use connection pooling for high-throughput scenarios
5. **Batch Operations**: Implement batch inserts/updates for better performance
6. **Cleanup**: Implement cleanup jobs for old execution records

## Dependencies to Add

```go
require (
    github.com/lib/pq v1.10.9                    // PostgreSQL driver
    github.com/jmoiron/sqlx v1.3.5               // SQL extensions
    github.com/golang-migrate/migrate/v4 v4.16.2 // Database migrations
)
```

## Implementation Plan

### Phase 1: Database Setup (2 hours)
1. Add PostgreSQL dependencies
2. Create migration files
3. Update data models
4. Create database connection utilities

### Phase 2: Storage Implementation (3 hours)
1. Implement PostgreSQL storage layer
2. Replace in-memory storage
3. Add transaction support
4. Implement connection pooling

### Phase 3: Testing & Optimization (1 hour)
1. Database integration tests
2. Performance testing
3. Error handling validation

## Version History

- **v1.0** (2025-01-06): Initial schema design with three tables
  - Job table for configuration
  - JobSchedule table for execution scheduling
  - JobExecution table for history tracking
  - Updated to use SERIAL primary keys for better performance
  - Simplified schema with only essential indexes and constraints
  - Application-level validation for flexibility
