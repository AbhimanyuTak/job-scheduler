# Job Execution Types & Retry System

## Overview

The job scheduler supports two execution types with different retry behaviors to handle various use cases and reliability requirements.

## Execution Types

### AT_MOST_ONCE
- **Guarantee**: Job executes at most once (never twice)
- **Retries**: No retries on failure
- **Use Case**: Idempotent operations, notifications, payments, one-time tasks

### AT_LEAST_ONCE  
- **Guarantee**: Job executes at least once (with retries)
- **Retries**: Automatic retries with exponential backoff
- **Use Case**: Critical operations, data processing, webhooks, monitoring

## Retry Behavior

### AT_LEAST_ONCE Jobs
- **Retries**: Yes, with exponential backoff
- **Max Retries**: Configurable (default: 3)
- **Backoff**: 1s, 2s, 4s, 8s, 16s, 32s, 64s, 128s, 256s, 512s (max 5min)
- **Behavior**: Retry until success or max retries exceeded

### AT_MOST_ONCE Jobs
- **Retries**: No retries
- **Behavior**: Execute once, mark as failed if unsuccessful
- **Rationale**: "At most once" means never execute twice

## Execution Flow

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

## Examples

### Example 1: AT_LEAST_ONCE with Retries
```json
{
  "schedule": "*/10 * * * * *",
  "api": "https://unreliable-api.com/webhook",
  "type": "AT_LEAST_ONCE",
  "isRecurring": true,
  "maxRetryCount": 3
}
```

**Timeline:**
1. **T+0s**: First attempt → FAILED
2. **T+1s**: Retry 1 → FAILED  
3. **T+3s**: Retry 2 → FAILED
4. **T+7s**: Retry 3 → FAILED
5. **T+15s**: Max retries exceeded → Reschedule for next occurrence

### Example 2: AT_MOST_ONCE (No Retries)
```json
{
  "schedule": "0 0 12 * * *",
  "api": "https://critical-api.com/notification",
  "type": "AT_MOST_ONCE",
  "isRecurring": false,
  "maxRetryCount": 5
}
```

**Timeline:**
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

**Timeline:**
1. **T+0s**: First attempt → FAILED
2. **T+1s**: Retry 1 → FAILED
3. **T+3s**: Retry 2 → SUCCESS ✅
4. **Reschedule** for next occurrence

## Use Cases

### AT_MOST_ONCE Use Cases

#### 1. Notifications
```json
{
  "schedule": "0 0 9 * * *",
  "api": "https://email-service.com/send-welcome",
  "type": "AT_MOST_ONCE",
  "description": "Send welcome email to new user"
}
```
**Why**: Don't want to spam users with duplicate emails

#### 2. Payments
```json
{
  "schedule": "0 0 1 1 * *",
  "api": "https://billing.com/charge-monthly",
  "type": "AT_MOST_ONCE",
  "description": "Monthly subscription charge"
}
```
**Why**: Critical to avoid double-charging customers

#### 3. Data Processing
```json
{
  "schedule": "0 0 2 * * *",
  "api": "https://analytics.com/process-daily-data",
  "type": "AT_MOST_ONCE",
  "description": "Process daily analytics data"
}
```
**Why**: Don't want to double-process the same data

### AT_LEAST_ONCE Use Cases

#### 1. Data Synchronization
```json
{
  "schedule": "*/30 * * * * *",
  "api": "https://sync-service.com/update-inventory",
  "type": "AT_LEAST_ONCE",
  "description": "Sync inventory data"
}
```
**Why**: Data consistency is critical, retries ensure eventual success

#### 2. Monitoring & Alerts
```json
{
  "schedule": "*/5 * * * * *",
  "api": "https://monitoring.com/health-check",
  "type": "AT_LEAST_ONCE",
  "description": "System health monitoring"
}
```
**Why**: Monitoring must be reliable, temporary failures should be retried

## Comparison

| Aspect | AT_MOST_ONCE | AT_LEAST_ONCE |
|--------|--------------|---------------|
| **Retries on Failure** | ❌ No | ✅ Yes |
| **Duplicate Execution** | ❌ Never | ⚠️ Possible |
| **Guarantee** | At most once | At least once |
| **Use Case** | Notifications, Payments | Data sync, Monitoring |
| **Failure Handling** | Mark as failed | Retry until success |

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

### 1. Choose the Right Job Type
- **AT_MOST_ONCE**: For operations that shouldn't be repeated
- **AT_LEAST_ONCE**: For operations that must eventually succeed

### 2. Design Idempotent APIs
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

### 3. Monitor Failure Rates
- AT_MOST_ONCE jobs that fail frequently indicate API issues
- Consider switching to AT_LEAST_ONCE if reliability is more important than avoiding duplicates

### 4. Set Appropriate Max Retries
- Balance between reliability and resource usage
- Monitor retry patterns for API issues

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
T+12:00:01: No retry attempted (AT_MOST_ONCE) or Retry in 1s (AT_LEAST_ONCE)
```

### Scenario 3: Network Timeout
```
Timeline:
T+0s: Job scheduled for 12:00 PM
T+12:00:00: Job executes → Network timeout after 30s
T+12:00:30: No retry attempted (AT_MOST_ONCE) or Retry in 1s (AT_LEAST_ONCE)
```

## Monitoring

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

## Summary

**AT_MOST_ONCE** jobs provide a guarantee that operations will never be executed more than once, making them ideal for:
- **Notifications** (avoid spam)
- **Payments** (avoid double-charging)
- **Data processing** (avoid duplicate work)
- **Resource management** (avoid conflicts)

**AT_LEAST_ONCE** jobs ensure eventual success through retries, making them ideal for:
- **Data synchronization** (ensure consistency)
- **Monitoring** (ensure reliability)
- **Critical operations** (ensure completion)

Choose the right type based on your reliability requirements and tolerance for duplicate execution.
