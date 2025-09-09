# Job Scheduler API Endpoints

## Base URL
- Development: `http://localhost:8080`
- Production: `https://api.jobscheduler.com`

## Authentication
- API Key: `X-API-Key` header
- JWT Token: `Authorization: Bearer <token>` header

## Endpoints

### Health Check
```
GET /health
```
**Response:**
```json
{
  "status": "healthy"
}
```

### Job Management

#### Create Job
```
POST /api/v1/jobs
```
**Request Body:**
```json
{
  "schedule": "31 10-15 1 * * MON-FRI",
  "api": "https://api.example.com/webhook",
  "type": "AT_LEAST_ONCE",
  "isRecurring": true,
  "description": "Send daily report",
  "maxRetryCount": 3
}
```
**Response:**
```json
{
  "jobId": 1,
  "message": "Job created successfully"
}
```

#### List Jobs
```
GET /api/v1/jobs?limit=100&offset=0
```
**Response:**
```json
{
  "jobs": [
    {
      "id": 1,
      "schedule": "31 10-15 1 * * MON-FRI",
      "api": "https://api.example.com/webhook",
      "type": "AT_LEAST_ONCE",
      "isRecurring": true,
      "isActive": true,
      "description": "Send daily report",
      "maxRetryCount": 3,
      "createdAt": "2025-01-06T10:30:00Z",
      "updatedAt": "2025-01-06T10:30:00Z"
    }
  ],
  "total": 1,
  "limit": 100,
  "offset": 0
}
```

#### Get Job
```
GET /api/v1/jobs/{jobId}
```
**Response:**
```json
{
  "id": 1,
  "schedule": "31 10-15 1 * * MON-FRI",
  "api": "https://api.example.com/webhook",
  "type": "AT_LEAST_ONCE",
  "isRecurring": true,
  "isActive": true,
  "description": "Send daily report",
  "maxRetryCount": 3,
  "createdAt": "2025-01-06T10:30:00Z",
  "updatedAt": "2025-01-06T10:30:00Z"
}
```

#### Get Job Schedule
```
GET /api/v1/jobs/{jobId}/schedule
```
**Response:**
```json
{
  "id": 1,
  "jobId": 1,
  "nextExecutionTime": "2025-01-06T11:00:00Z",
  "createdAt": "2025-01-06T10:30:00Z"
}
```

#### Get Job Execution History
```
GET /api/v1/jobs/{jobId}/history?limit=10&status=SUCCESS
```
**Response:**
```json
{
  "executions": [
    {
      "id": 1,
      "jobId": 1,
      "status": "SUCCESS",
      "executionTime": "2025-01-06T10:30:00Z",
      "executionDuration": 1500,
      "retryCount": 0,
      "createdAt": "2025-01-06T10:30:00Z",
      "updatedAt": "2025-01-06T10:30:00Z"
    }
  ],
  "total": 1,
  "limit": 10
}
```


## Data Types

### Job Types
- `AT_LEAST_ONCE`: Job will be retried on failure until it succeeds or max retries reached
- `AT_MOST_ONCE`: Job will be executed at most once, no retries on failure

### Execution Status
- `SCHEDULED`: Job is scheduled for execution
- `RUNNING`: Job is currently being executed
- `SUCCESS`: Job executed successfully
- `FAILED`: Job execution failed

### CRON Format
Extended CRON format with seconds support:
```
<second> <minute> <hour> <day> <month> <day-of-week>
```

**Examples:**
- `"0 * * * * *"` - Every minute
- `"30 0 * * * *"` - Every hour at 30 seconds
- `"0 0 9 * * MON-FRI"` - Every weekday at 9:00 AM
- `"31 10-15 1 * * MON-FRI"` - Every 31st second of every minute between 01:10-01:15 AM on weekdays

## Error Responses

All error responses follow this format:
```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": {
    "field": "additional error details"
  }
}
```

**Common Error Codes:**
- `JOB_NOT_FOUND`: Job with specified ID not found
- `INVALID_SCHEDULE`: Invalid CRON expression
- `INVALID_API_URL`: Invalid API endpoint URL
- `EXECUTION_NOT_FOUND`: Execution with specified ID not found
- `VALIDATION_ERROR`: Request validation failed
