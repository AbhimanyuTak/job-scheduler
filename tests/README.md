# Test Suite Details

## Test Files

- **`e2e_test.go`** - End-to-end integration tests covering the full job lifecycle
- **`redis_test.go`** - Redis-specific tests for queue operations and data integrity
- **`config_test.go`** - Test configuration and environment handling

## Running Tests

### Using Make
```bash
# Unit tests
make -f test/Makefile.test test-unit

# Integration tests
make -f test/Makefile.test test-integration

# End-to-end tests
make -f test/Makefile.test test-e2e

# Redis tests
make -f test/Makefile.test test-redis

# All tests
make -f test/Makefile.test test-all

# With coverage
make -f test/Makefile.test test-coverage
```

### Using Go Directly
```bash
# Run all tests
go test ./tests/...

# Run specific test
go test -run TestJobCreation ./tests/...

# Run with integration tests enabled
INTEGRATION_TESTS=true go test ./tests/...
```

## Test Configuration

### Environment Variables
- **`INTEGRATION_TESTS`** - Set to `true` to enable integration tests
- **`TEST_BASE_URL`** - Override API base URL (default: `http://localhost:8080`)
- **`TEST_API_KEY`** - Override API key (default: `test-api-key`)
- **`TEST_REDIS_HOST`** - Override Redis host (default: `localhost`)
- **`TEST_REDIS_PORT`** - Override Redis port (default: `6379`)
- **`TEST_TIMEOUT`** - Override test timeout in seconds (default: `30`)

## Test Scenarios

### Basic Functionality
- Health check validation
- Job creation and retrieval
- Schedule population
- Queue statistics

### Job Execution
- Job processing by workers
- Execution history tracking
- Status updates and monitoring
- Completion handling

### Error Handling
- Invalid job creation
- Non-existent job access
- Network timeouts
- Service unavailability

### Performance
- High throughput job creation
- Queue processing under load
- Redis performance
- Memory usage monitoring

## Prerequisites

### Required Services
```bash
# Start all services
docker compose up -d

# Verify services are running
docker compose ps

# Check API health
curl http://localhost:8080/health

# Check Redis
docker compose exec redis redis-cli ping
```

### Required Tools
- Go 1.23+
- Docker and Docker Compose
- Make (optional, for convenience commands)
- curl (for health checks)

## Test Data

### Test Jobs
Tests create various types of jobs:
- **Recurring jobs** - Execute every 10-30 seconds
- **One-time jobs** - Execute once at a specific time
- **High retry jobs** - Jobs with multiple retry attempts
- **Different APIs** - Various test endpoints (httpbin.org)

### Test Endpoints
Tests use these external endpoints:
- `https://httpbin.org/status/200` - Simple success response
- `https://httpbin.org/delay/1` - Delayed response (1 second)
- `https://httpbin.org/json` - JSON response
- `https://httpbin.org/status/500` - Error response (for retry testing)

## Troubleshooting

### Common Issues

#### Services Not Running
```bash
# Check service status
docker compose ps

# Start services
docker compose up -d

# Check logs
docker compose logs scheduler
docker compose logs worker
docker compose logs redis
```

#### Test Timeouts
```bash
# Increase timeout
export TEST_TIMEOUT=60
go test ./tests/...
```

#### Redis Connection Issues
```bash
# Check Redis
docker compose exec redis redis-cli ping

# Check Redis logs
docker compose logs redis
```

#### API Connection Issues
```bash
# Check API health
curl http://localhost:8080/health

# Check API logs
docker compose logs scheduler
```

## Contributing

### Adding New Tests

1. **Create test functions** following Go testing conventions
2. **Use descriptive names** that explain what's being tested
3. **Include proper assertions** using testify/assert
4. **Handle errors appropriately** using testify/require
5. **Add documentation** explaining the test purpose

### Test Naming Convention
- **Unit tests**: `TestFunctionName`
- **Integration tests**: `TestIntegrationFeature`
- **End-to-end tests**: `TestEndToEndWorkflow`
- **Benchmark tests**: `BenchmarkOperation`

### Example Test Structure
```go
func TestJobCreation(t *testing.T) {
    // Setup
    client := NewTestClient()
    
    // Test data
    req := CreateJobRequest{
        Schedule: "*/10 * * * * *",
        API:      "https://httpbin.org/status/200",
        Type:     models.AT_LEAST_ONCE,
        // ... other fields
    }
    
    // Execute
    jobID := client.CreateJob(t, req)
    
    // Verify
    job := client.GetJob(t, jobID)
    assert.Equal(t, req.Schedule, job.Schedule)
    assert.True(t, job.IsActive)
}
```