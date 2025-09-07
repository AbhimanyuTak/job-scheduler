# Job Execution Flow and Scheduling

## Overview

The job scheduler supports both recurring and non-recurring (one-time) jobs. This document explains how jobs are scheduled, executed, and managed throughout their lifecycle.

## CRON Schedule Format

The system uses a 6-field CRON format that includes seconds:

```
second minute hour day month weekday
```

Examples:
- `0 0 12 * * *` - Every day at 12:00:00 PM
- `30 0 9 * * MON-FRI` - Every weekday at 9:00:30 AM
- `0 */15 * * * *` - Every 15 minutes
- `0 0 0 1 * *` - First day of every month at midnight

## Job Types

### Recurring Jobs (`isRecurring: true`)

Recurring jobs are executed repeatedly according to their CRON schedule.

**Lifecycle:**
1. Job is created with a valid CRON schedule
2. Next execution time is calculated from the current time
3. Job schedule is created in the database
4. Background scheduler picks up the job when it's time to execute
5. Job is executed via HTTP API call
6. Next execution time is calculated and the schedule is updated
7. Process repeats indefinitely until job is deleted or deactivated

### Non-Recurring Jobs (`isRecurring: false`)

Non-recurring jobs are executed only once at the specified time.

**Lifecycle:**
1. Job is created with a valid CRON schedule (defines when to execute)
2. Next execution time is calculated from the current time
3. Job schedule is created in the database
4. Background scheduler picks up the job when it's time to execute
5. Job is executed via HTTP API call
6. **Schedule is deleted** after execution (job becomes "completed")
7. Job remains in the database for history tracking but won't be executed again

## Key Differences

| Aspect | Recurring Jobs | Non-Recurring Jobs |
|--------|----------------|-------------------|
| Execution Frequency | Repeated according to schedule | Once only |
| Schedule After Execution | Updated with next execution time | Deleted |
| Database Cleanup | Schedule persists | Schedule removed |
| History | Multiple execution records | Single execution record |
| Use Case | Periodic tasks, monitoring | One-time operations, notifications |

## Background Scheduler

The background scheduler runs continuously and:

1. **Checks every 5 seconds** for jobs ready to execute
2. **Fetches jobs** where `next_execution_time <= now()`
3. **Executes jobs** by making HTTP POST calls to the specified API
4. **Records execution results** in the database
5. **Handles rescheduling** based on job type:
   - Recurring: Calculate and update next execution time
   - Non-recurring: Delete the schedule

## Execution Status Tracking

Each job execution is tracked with the following statuses:

- `SCHEDULED` - Job is queued for execution
- `RUNNING` - Job is currently being executed
- `SUCCESS` - Job completed successfully
- `FAILED` - Job execution failed

## Error Handling and Retries

- Jobs support configurable retry counts (default: 3)
- Failed executions are recorded with error details
- Retry logic can be implemented based on job type (`AT_LEAST_ONCE` vs `AT_MOST_ONCE`)

## API Endpoints

### Create Job
```http
POST /api/v1/jobs
{
  "schedule": "0 0 12 * * *",
  "api": "https://example.com/webhook",
  "type": "AT_LEAST_ONCE",
  "isRecurring": false,
  "description": "One-time notification",
  "maxRetryCount": 3
}
```

### Get Job History
```http
GET /api/v1/jobs/{id}/history?limit=10
```

### Get Job Schedule
```http
GET /api/v1/jobs/{id}/schedule
```

## Database Schema

### Jobs Table
- Stores job definitions and metadata
- `isRecurring` field determines execution behavior
- Jobs are soft-deleted (marked as inactive)

### JobSchedules Table
- Stores execution timing information
- **Deleted for non-recurring jobs** after execution
- **Updated for recurring jobs** with next execution time

### JobExecutions Table
- Records all execution attempts and results
- Maintains history for both recurring and non-recurring jobs

## Best Practices

1. **Use appropriate job types**: Set `isRecurring` correctly based on your needs
2. **Monitor execution history**: Check job execution logs regularly
3. **Handle API failures**: Ensure your target APIs can handle the expected load
4. **Clean up old data**: Consider archiving old execution records
5. **Test schedules**: Validate CRON expressions before creating jobs

## Example Scenarios

### Scenario 1: Daily Report (Recurring)
```json
{
  "schedule": "0 0 9 * * MON-FRI",
  "api": "https://reports.company.com/generate-daily",
  "isRecurring": true,
  "description": "Generate daily business report"
}
```
- Executes every weekday at 9:00 AM
- Schedule persists and updates after each execution
- Creates multiple execution records over time

### Scenario 2: Welcome Email (Non-Recurring)
```json
{
  "schedule": "0 0 10 * * *",
  "api": "https://email.service.com/send-welcome",
  "isRecurring": false,
  "description": "Send welcome email to new user"
}
```
- Executes once at 10:00 AM (next occurrence)
- Schedule is deleted after execution
- Creates single execution record
- Job remains in database for audit purposes
