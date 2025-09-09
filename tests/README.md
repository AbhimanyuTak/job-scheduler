# Job Scheduler Test Suite

This directory contains comprehensive Go-based end-to-end tests for the Job Scheduler system.

## Test Structure

### Test Files

- **`e2e_test.go`** - End-to-end integration tests covering the full job lifecycle
- **`redis_test.go`** - Redis-specific tests for queue operations and data integrity
- **`config_test.go`** - Test configuration and environment handling

### Test Types

#### 1. Unit Tests
- Fast tests that don't require external dependencies
- Test individual components in isolation
- Run with: `make -f Makefile.test test-unit`

#### 2. Integration Tests
- Tests that require running services (API, Redis, Database)
- Test component interactions
- Run with: `make -f Makefile.test test-integration`

#### 3. End-to-End Tests
- Full system tests covering complete workflows
- Test job creation, scheduling, execution, and completion
- Run with: `make -f Makefile.test test-e2e`

## Running Tests

### Quick Start

```bash
# Run all unit tests
./run-tests.sh unit

# Run integration tests (requires running services)
./run-tests.sh integration

# Run end-to-end tests
./run-tests.sh e2e

# Run quick system check
./run-tests.sh quick
```

### Using Make

```bash
# Unit tests
make -f Makefile.test test-unit

# Integration tests
make -f Makefile.test test-integration

# End-to-end tests
make -f Makefile.test test-e2e

# Redis tests
make -f Makefile.test test-redis

# All tests
make -f Makefile.test test-all

# With coverage
make -f Makefile.test test-coverage

# With race detection
make -f Makefile.test test-race

# Benchmark tests
make -f Makefile.test test-benchmark
```

### Using Go Directly

```bash
# Run all tests
go test ./tests/...

# Run specific test
go test -run TestJobCreation ./tests/...

# Run with integration tests enabled
INTEGRATION_TESTS=true go test ./tests/...

# Run with coverage
go test -cover ./tests/...

# Run with race detection
go test -race ./tests/...
```

## Test Configuration

### Environment Variables

- **`INTEGRATION_TESTS`** - Set to `true` to enable integration tests
- **`TEST_BASE_URL`** - Override API base URL (default: `http://localhost:8080`)
- **`TEST_API_KEY`** - Override API key (default: `test-api-key`)
- **`TEST_REDIS_HOST`** - Override Redis host (default: `localhost`)
- **`TEST_REDIS_PORT`** - Override Redis port (default: `6379`)
- **`TEST_TIMEOUT`** - Override test timeout in seconds (default: `30`)
- **`TEST_MAX_RETRIES`** - Override max retries (default: `3`)
- **`TEST_WAIT_INTERVAL`** - Override wait interval in seconds (default: `2`)

### Example Configuration

```bash
export INTEGRATION_TESTS=true
export TEST_BASE_URL=http://localhost:8080
export TEST_API_KEY=my-test-key
export TEST_TIMEOUT=60
go test ./tests/...
```

## Test Coverage

### What's Tested

#### Job Management
- ✅ Job creation with various schedules
- ✅ Job validation and error handling
- ✅ Job types (AT_LEAST_ONCE, AT_MOST_ONCE)
- ✅ Recurring vs one-time jobs

#### Scheduling
- ✅ Schedule population and validation
- ✅ Next execution time calculation
- ✅ CRON expression parsing

#### Redis Queue System
- ✅ Job enqueueing and dequeuing
- ✅ Queue statistics and monitoring
- ✅ Data integrity and consistency
- ✅ Performance and throughput

#### Worker Processing
- ✅ Job execution and completion
- ✅ Retry mechanisms
- ✅ Error handling and recovery
- ✅ Worker scaling and load distribution

#### System Integration
- ✅ Health checks and monitoring
- ✅ API endpoint functionality
- ✅ Database operations
- ✅ Error handling and validation

### Test Scenarios

#### Basic Functionality
- Health check validation
- Job creation and retrieval
- Schedule population
- Queue statistics

#### Job Execution
- Job processing by workers
- Execution history tracking
- Status updates and monitoring
- Completion handling

#### Error Handling
- Invalid job creation
- Non-existent job access
- Network timeouts
- Service unavailability

#### Performance
- High throughput job creation
- Queue processing under load
- Redis performance
- Memory usage monitoring

#### Integration
- End-to-end workflows
- Service interactions
- Data consistency
- System reliability

## Prerequisites

### Required Services

Before running integration tests, ensure these services are running:

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

### Debug Mode

Run tests with verbose output:

```bash
# Verbose test output
go test -v ./tests/...

# Run specific test with verbose output
go test -v -run TestJobCreation ./tests/...
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

## Performance Considerations

### Test Execution Time

- **Unit tests**: < 1 second
- **Integration tests**: 1-5 minutes
- **End-to-end tests**: 5-10 minutes
- **Benchmark tests**: Variable

### Resource Usage

- **Memory**: Tests use minimal memory
- **CPU**: Tests are CPU-efficient
- **Network**: Tests make HTTP requests to external services
- **Storage**: Tests don't create persistent data

### Optimization Tips

- Run unit tests frequently during development
- Run integration tests before commits
- Run end-to-end tests before releases
- Use parallel test execution when possible
- Monitor test execution time and optimize slow tests
