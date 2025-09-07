# AT_MOST_ONCE Job Semantics

## What Does AT_MOST_ONCE Mean?

**AT_MOST_ONCE** means the job will be executed **at most once** - it will never be executed more than once, even if it fails. This is the opposite of AT_LEAST_ONCE, which guarantees the job will be executed at least once (with retries).

## Key Characteristics

### ✅ **What AT_MOST_ONCE Guarantees:**
- **No Duplicate Execution**: Job will never run twice
- **No Retries**: If it fails, it's marked as failed and never retried
- **Idempotent Operations**: Safe for operations that shouldn't be repeated

### ❌ **What AT_MOST_ONCE Does NOT Guarantee:**
- **Execution Success**: Job might fail and never succeed
- **Execution**: Job might not execute at all (if system is down)

## How AT_MOST_ONCE Works in Our System

### 1. **Job Creation**
```json
{
  "schedule": "0 0 12 * * *",
  "api": "https://example.com/send-notification",
  "type": "AT_MOST_ONCE",
  "isRecurring": false,
  "description": "Send one-time notification"
}
```

### 2. **Execution Flow**
```
Job Ready → Execute Once → Success/Failure → No Retries
```

### 3. **Retry Logic**
```go
// AT_MOST_ONCE jobs should not be retried on failure
if job.Type == models.AT_MOST_ONCE {
    return false  // Never retry
}
```

## Examples

### Example 1: Successful AT_MOST_ONCE Job
```
Timeline:
T+0s: Job scheduled for 12:00 PM
T+12:00:00: Job executes → SUCCESS ✅
Result: Job completed, schedule deleted (if non-recurring)
```

### Example 2: Failed AT_MOST_ONCE Job
```
Timeline:
T+0s: Job scheduled for 12:00 PM  
T+12:00:00: Job executes → FAILED ❌
T+12:00:01: No retry attempted
Result: Job marked as failed, schedule deleted (if non-recurring)
```

### Example 3: Recurring AT_MOST_ONCE Job
```json
{
  "schedule": "0 0 12 * * MON-FRI",
  "api": "https://example.com/daily-report",
  "type": "AT_MOST_ONCE",
  "isRecurring": true,
  "description": "Daily report (at most once per day)"
}
```

**Behavior:**
- **Monday 12:00 PM**: Execute → SUCCESS → Reschedule for Tuesday
- **Tuesday 12:00 PM**: Execute → FAILED → No retry, reschedule for Wednesday
- **Wednesday 12:00 PM**: Execute → SUCCESS → Reschedule for Thursday
- And so on...

## Use Cases for AT_MOST_ONCE

### 1. **Notifications**
```json
{
  "schedule": "0 0 9 * * *",
  "api": "https://email-service.com/send-welcome",
  "type": "AT_MOST_ONCE",
  "description": "Send welcome email to new user"
}
```
**Why AT_MOST_ONCE**: Don't want to spam users with duplicate emails

### 2. **Data Processing**
```json
{
  "schedule": "0 0 2 * * *",
  "api": "https://analytics.com/process-daily-data",
  "type": "AT_MOST_ONCE",
  "description": "Process daily analytics data"
}
```
**Why AT_MOST_ONCE**: Don't want to double-process the same data

### 3. **Payment Processing**
```json
{
  "schedule": "0 0 1 1 * *",
  "api": "https://billing.com/charge-monthly",
  "type": "AT_MOST_ONCE",
  "description": "Monthly subscription charge"
}
```
**Why AT_MOST_ONCE**: Critical to avoid double-charging customers

### 4. **Resource Cleanup**
```json
{
  "schedule": "0 0 3 * * *",
  "api": "https://storage.com/cleanup-temp-files",
  "type": "AT_MOST_ONCE",
  "description": "Clean up temporary files"
}
```
**Why AT_MOST_ONCE**: Cleanup operations should be idempotent

## Comparison: AT_MOST_ONCE vs AT_LEAST_ONCE

| Aspect | AT_MOST_ONCE | AT_LEAST_ONCE |
|--------|--------------|---------------|
| **Retries on Failure** | ❌ No | ✅ Yes |
| **Duplicate Execution** | ❌ Never | ⚠️ Possible |
| **Guarantee** | At most once | At least once |
| **Use Case** | Notifications, Payments | Data sync, Monitoring |
| **Failure Handling** | Mark as failed | Retry until success |

## Implementation Details

### Database Behavior
```sql
-- AT_MOST_ONCE job execution
INSERT INTO job_executions (job_id, status, retry_count, execution_time) 
VALUES (123, 'FAILED', 0, '2024-01-01 12:00:00');

-- No retry execution created
-- Schedule deleted (if non-recurring) or rescheduled (if recurring)
```

### Logging
```
Job 123 (AT_MOST_ONCE) failed, no retry attempted
Non-recurring job 123 failed, schedule deleted
```

### API Response
```json
{
  "id": 123,
  "status": "FAILED",
  "retryCount": 0,
  "executionTime": "2024-01-01T12:00:00Z",
  "error": "API call failed"
}
```

## Best Practices

### 1. **Choose the Right Job Type**
- **AT_MOST_ONCE**: For operations that shouldn't be repeated
- **AT_LEAST_ONCE**: For operations that must eventually succeed

### 2. **Design Idempotent APIs**
```javascript
// Good: Idempotent API
POST /api/notifications
{
  "userId": 123,
  "type": "welcome",
  "idempotencyKey": "welcome-123-2024-01-01"
}

// Bad: Non-idempotent API
POST /api/notifications
{
  "userId": 123,
  "type": "welcome"
  // No idempotency key - could send duplicate emails
}
```

### 3. **Monitor Failure Rates**
- AT_MOST_ONCE jobs that fail frequently indicate API issues
- Consider switching to AT_LEAST_ONCE if reliability is more important than avoiding duplicates

### 4. **Use Appropriate Scheduling**
- **Non-recurring**: For one-time operations
- **Recurring**: For periodic operations that should happen at most once per period

## Error Scenarios

### Scenario 1: System Down During Execution
```
Timeline:
T+0s: Job scheduled for 12:00 PM
T+12:00:00: System down, job not executed
T+12:00:01: System back up
Result: Job missed, no retry (AT_MOST_ONCE semantics)
```

### Scenario 2: API Temporarily Unavailable
```
Timeline:
T+0s: Job scheduled for 12:00 PM
T+12:00:00: Job executes → API returns 503 Service Unavailable
T+12:00:01: No retry attempted
Result: Job marked as failed, no retry
```

### Scenario 3: Network Timeout
```
Timeline:
T+0s: Job scheduled for 12:00 PM
T+12:00:00: Job executes → Network timeout after 30s
T+12:00:30: No retry attempted
Result: Job marked as failed, no retry
```

## Summary

AT_MOST_ONCE jobs provide a guarantee that operations will never be executed more than once, making them ideal for:

- **Notifications** (avoid spam)
- **Payments** (avoid double-charging)
- **Data processing** (avoid duplicate work)
- **Resource management** (avoid conflicts)

The trade-off is that if a job fails, it will never be retried, so you need to ensure your APIs are reliable or accept that some jobs might fail permanently.
