# Job Retry System

## Overview

The job scheduler now includes a comprehensive retry system that handles failed job executions based on job type and retry configuration. This ensures reliable job execution while respecting different job semantics.

## Retry Behavior by Job Type

### AT_LEAST_ONCE Jobs
- **Retry on Failure**: ✅ YES
- **Max Retries**: Configurable (default: 3)
- **Retry Delay**: Exponential backoff (1s, 2s, 4s, 8s, 16s, 32s, 64s, 128s, 256s, 512s, max 5min)
- **Behavior**: Will retry until successful or max retries exceeded

### AT_MOST_ONCE Jobs
- **Retry on Failure**: ❌ NO
- **Behavior**: Execute once, if it fails, mark as failed and don't retry
- **Rationale**: AT_MOST_ONCE semantics mean "execute at most once", so retries would violate this guarantee

## Retry Flow

### 1. Job Execution
```
Job Ready → Execute → Success/Failure
```

### 2. Success Handling
```
Success → Reschedule (recurring) or Delete Schedule (non-recurring)
```

### 3. Failure Handling
```
Failure → Check Job Type → Check Retry Count → Retry or Give Up
```

## Retry Logic Details

### Retry Count Tracking
- Each execution attempt creates a new `JobExecution` record
- `RetryCount` field tracks how many times this specific job has been attempted
- Retry count starts at 0 for the first execution

### Exponential Backoff
```go
// Retry delays: 1s, 2s, 4s, 8s, 16s, 32s, 64s, 128s, 256s, 512s, max 5min
delaySeconds := 1 << retryCount
maxDelay := 5 * time.Minute
```

### Retry Decision Matrix

| Job Type | Retry Count < Max | Retry Count >= Max | Action |
|----------|------------------|-------------------|---------|
| AT_LEAST_ONCE | ✅ | ❌ | Retry with backoff |
| AT_MOST_ONCE | ❌ | ❌ | No retry (respect semantics) |

## Examples

### Example 1: AT_LEAST_ONCE Job with Retries
```json
{
  "schedule": "*/10 * * * * *",
  "api": "https://unreliable-api.com/webhook",
  "type": "AT_LEAST_ONCE",
  "isRecurring": true,
  "maxRetryCount": 3
}
```

**Execution Timeline:**
1. **T+0s**: First attempt → FAILED
2. **T+1s**: Retry 1 → FAILED  
3. **T+3s**: Retry 2 → FAILED
4. **T+7s**: Retry 3 → FAILED
5. **T+15s**: Max retries exceeded → Reschedule for next occurrence

### Example 2: AT_MOST_ONCE Job (No Retries)
```json
{
  "schedule": "0 0 12 * * *",
  "api": "https://critical-api.com/notification",
  "type": "AT_MOST_ONCE",
  "isRecurring": false,
  "maxRetryCount": 5
}
```

**Execution Timeline:**
1. **T+0s**: First attempt → FAILED
2. **No retries** → Mark as failed, delete schedule

### Example 3: Successful Retry
```json
{
  "schedule": "*/5 * * * * *",
  "api": "https://sometimes-down.com/api",
  "type": "AT_LEAST_ONCE",
  "isRecurring": true,
  "maxRetryCount": 2
}
```

**Execution Timeline:**
1. **T+0s**: First attempt → FAILED
2. **T+1s**: Retry 1 → FAILED
3. **T+3s**: Retry 2 → SUCCESS ✅
4. **Reschedule** for next occurrence

## Database Schema Impact

### JobExecutions Table
Each retry attempt creates a new record:
```sql
-- First attempt
INSERT INTO job_executions (job_id, status, retry_count, execution_time) 
VALUES (1, 'FAILED', 0, '2024-01-01 12:00:00');

-- First retry
INSERT INTO job_executions (job_id, status, retry_count, execution_time) 
VALUES (1, 'FAILED', 1, '2024-01-01 12:00:01');

-- Second retry (success)
INSERT INTO job_executions (job_id, status, retry_count, execution_time) 
VALUES (1, 'SUCCESS', 2, '2024-01-01 12:00:03');
```

### JobSchedules Table
- **During retries**: `next_execution_time` is updated to retry time
- **After max retries**: For recurring jobs, rescheduled to next occurrence
- **After max retries**: For non-recurring jobs, schedule is deleted

## Monitoring and Logging

### Log Messages
```
Job 123 failed, scheduling retry 1/3 in 1s (at 2024-01-01 12:00:01)
Job 123 failed, scheduling retry 2/3 in 2s (at 2024-01-01 12:00:03)
Job 123 failed after 3 retries, rescheduled for next occurrence: 2024-01-01 13:00:00
```

### Execution History
Use the existing API to monitor retry attempts:
```bash
GET /api/v1/jobs/{id}/history?limit=10
```

## Configuration

### Default Settings
- **Max Retry Count**: 3 (configurable per job)
- **Retry Delay**: Exponential backoff starting at 1 second
- **Max Retry Delay**: 5 minutes
- **HTTP Timeout**: 30 seconds per attempt

### Customization
You can modify retry behavior by adjusting:
- `maxRetryCount` in job creation
- `calculateRetryDelay()` function for different backoff strategies
- `shouldRetryJob()` function for custom retry logic

## Best Practices

1. **Set Appropriate Max Retries**: Balance between reliability and resource usage
2. **Monitor Retry Patterns**: High retry rates may indicate API issues
3. **Use AT_MOST_ONCE for Critical Jobs**: Prevents duplicate processing
4. **Use AT_LEAST_ONCE for Reliable Delivery**: Ensures eventual success
5. **Set Reasonable Timeouts**: Prevent long-running failed requests

## Error Handling

### API Failures
- HTTP errors (4xx, 5xx)
- Network timeouts
- Connection refused

### System Failures
- Database connection issues
- Scheduler service crashes
- Resource exhaustion

### Recovery
- Failed retries are logged with full context
- System continues processing other jobs
- Manual intervention possible through job management APIs
