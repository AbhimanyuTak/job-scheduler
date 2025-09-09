# API Reference

## Base URL
- Local: `http://localhost:8080`
- API Path: `/api/v1`

## Authentication
- API Key: `X-API-Key` header
- JWT Token: `Authorization: Bearer <token>` header

## Endpoints

### Health Check
```http
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
```http
POST /api/v1/jobs
```
**Request:**
```json
{
  "schedule": "0 0 9 * * MON-FRI",
  "api": "https://api.example.com/webhook",
  "type": "AT_LEAST_ONCE",
  "isRecurring": true,
  "description": "Daily report",
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
```http
GET /api/v1/jobs?limit=100&offset=0
```

#### Get Job
```http
GET /api/v1/jobs/{id}
```

#### Get Job Schedule
```http
GET /api/v1/jobs/{id}/schedule
```

#### Get Job History
```http
GET /api/v1/jobs/{id}/history?limit=10&status=SUCCESS
```

### Queue Statistics
```http
GET /queue/stats
```
**Response:**
```json
{
  "queue_stats": {
    "ready": 5,
    "processing": 3,
    "completed": 150,
    "retrying": 2
  }
}
```

## Data Types

### Job Types
- `AT_LEAST_ONCE`: Retry on failure until success or max retries
- `AT_MOST_ONCE`: Execute once, no retries

### Execution Status
- `SCHEDULED`: Job scheduled for execution
- `RUNNING`: Currently executing
- `SUCCESS`: Executed successfully
- `FAILED`: Execution failed

### CRON Format
Extended 6-field format: `<second> <minute> <hour> <day> <month> <day-of-week>`

**Examples:**
- `"0 * * * * *"` - Every minute
- `"0 0 9 * * MON-FRI"` - Weekdays at 9:00 AM
- `"30 0 9 * * MON-FRI"` - Weekdays at 9:00:30 AM

## Error Responses
```json
{
  "error": "Error message",
  "code": "ERROR_CODE"
}
```

**Common Error Codes:**
- `JOB_NOT_FOUND`: Job not found
- `INVALID_SCHEDULE`: Invalid CRON expression
- `VALIDATION_ERROR`: Request validation failed
