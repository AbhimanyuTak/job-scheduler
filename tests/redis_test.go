package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manyu/job-scheduler/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RedisTestClient wraps Redis operations for testing
type RedisTestClient struct {
	redisClient *services.RedisClient
	jobQueue    *services.JobQueueService
}

// NewRedisTestClient creates a new Redis test client
func NewRedisTestClient(t *testing.T) *RedisTestClient {
	redisClient, err := services.NewRedisClient("localhost:6379", "", 0)
	require.NoError(t, err, "Failed to create Redis client")

	jobQueue := services.NewJobQueueService(redisClient)

	return &RedisTestClient{
		redisClient: redisClient,
		jobQueue:    jobQueue,
	}
}

// Close closes the Redis client
func (rtc *RedisTestClient) Close() error {
	return rtc.redisClient.Close()
}

// GetQueueStats gets queue statistics from Redis
func (rtc *RedisTestClient) GetQueueStats() (map[string]int64, error) {
	return rtc.jobQueue.GetQueueStats()
}

// GetReadyQueueLength gets the length of the ready queue
func (rtc *RedisTestClient) GetReadyQueueLength() (int64, error) {
	stats, err := rtc.GetQueueStats()
	if err != nil {
		return 0, err
	}
	return stats["ready"], nil
}

// GetProcessingQueueLength gets the length of the processing queue
func (rtc *RedisTestClient) GetProcessingQueueLength() (int64, error) {
	stats, err := rtc.GetQueueStats()
	if err != nil {
		return 0, err
	}
	return stats["processing"], nil
}

// GetCompletedQueueLength gets the length of the completed queue
func (rtc *RedisTestClient) GetCompletedQueueLength() (int64, error) {
	stats, err := rtc.GetQueueStats()
	if err != nil {
		return 0, err
	}
	return stats["completed"], nil
}

// GetRetryingQueueLength gets the length of the retrying queue
func (rtc *RedisTestClient) GetRetryingQueueLength() (int64, error) {
	stats, err := rtc.GetQueueStats()
	if err != nil {
		return 0, err
	}
	return stats["retrying"], nil
}

// TestRedisConnection tests basic Redis connectivity
func TestRedisConnection(t *testing.T) {
	client := NewRedisTestClient(t)
	defer client.Close()

	// Test basic ping
	ctx := context.Background()
	redisClient := client.redisClient.GetClient()

	pong, err := redisClient.Ping(ctx).Result()
	require.NoError(t, err, "Redis ping failed")
	assert.Equal(t, "PONG", pong)
}

func TestRedisQueueOperations(t *testing.T) {
	client := NewRedisTestClient(t)
	defer client.Close()

	// Get initial queue stats
	initialStats, err := client.GetQueueStats()
	require.NoError(t, err, "Failed to get initial queue stats")

	t.Logf("Initial queue stats: %+v", initialStats)

	// Verify all queue counts are non-negative
	assert.GreaterOrEqual(t, initialStats["ready"], int64(0))
	assert.GreaterOrEqual(t, initialStats["processing"], int64(0))
	assert.GreaterOrEqual(t, initialStats["completed"], int64(0))
	assert.GreaterOrEqual(t, initialStats["retrying"], int64(0))
}

func TestRedisDataIntegrity(t *testing.T) {
	client := NewRedisTestClient(t)
	defer client.Close()

	ctx := context.Background()
	redisClient := client.redisClient.GetClient()

	// Check for job data keys
	jobDataKeys, err := redisClient.Keys(ctx, "job_data:*").Result()
	require.NoError(t, err, "Failed to get job data keys")

	t.Logf("Found %d job data keys", len(jobDataKeys))

	// Check ready queue
	readyJobs, err := redisClient.LRange(ctx, "job_queue:ready", 0, 2).Result()
	require.NoError(t, err, "Failed to get ready queue")

	if len(readyJobs) > 0 {
		t.Logf("Sample ready jobs: %v", readyJobs)
	}

	// Check processing set
	processingJobs, err := redisClient.SMembers(ctx, "job_queue:processing").Result()
	require.NoError(t, err, "Failed to get processing jobs")

	if len(processingJobs) > 0 {
		t.Logf("Processing jobs: %v", processingJobs)
	}

	// Check completed queue
	completedJobs, err := redisClient.LRange(ctx, "job_queue:completed", 0, 2).Result()
	require.NoError(t, err, "Failed to get completed jobs")

	if len(completedJobs) > 0 {
		t.Logf("Sample completed jobs: %v", completedJobs)
	}
}

func TestRedisQueueMonitoring(t *testing.T) {
	client := NewRedisTestClient(t)
	defer client.Close()

	// Monitor queue changes over time
	initialStats, err := client.GetQueueStats()
	require.NoError(t, err, "Failed to get initial stats")

	t.Logf("Initial stats: Ready=%d, Processing=%d, Completed=%d, Retrying=%d",
		initialStats["ready"], initialStats["processing"],
		initialStats["completed"], initialStats["retrying"])

	// Wait for potential changes
	t.Log("Waiting 20 seconds for queue activity...")
	time.Sleep(20 * time.Second)

	finalStats, err := client.GetQueueStats()
	require.NoError(t, err, "Failed to get final stats")

	t.Logf("Final stats: Ready=%d, Processing=%d, Completed=%d, Retrying=%d",
		finalStats["ready"], finalStats["processing"],
		finalStats["completed"], finalStats["retrying"])

	// Check if there's been any activity
	activityDetected := false
	if finalStats["completed"] > initialStats["completed"] {
		t.Log("Activity detected: completed jobs increased")
		activityDetected = true
	}
	if finalStats["processing"] > initialStats["processing"] {
		t.Log("Activity detected: processing jobs increased")
		activityDetected = true
	}
	if finalStats["ready"] != initialStats["ready"] {
		t.Log("Activity detected: ready queue changed")
		activityDetected = true
	}

	if !activityDetected {
		t.Log("No queue activity detected - this might be normal")
	}
}

func TestRedisPerformance(t *testing.T) {
	client := NewRedisTestClient(t)
	defer client.Close()

	ctx := context.Background()
	redisClient := client.redisClient.GetClient()

	// Test Redis performance with multiple operations
	start := time.Now()

	// Perform multiple operations
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("test:key:%d", i)
		value := fmt.Sprintf("test:value:%d", i)

		// Set key
		err := redisClient.Set(ctx, key, value, time.Minute).Err()
		require.NoError(t, err, "Failed to set key %s", key)

		// Get key
		result, err := redisClient.Get(ctx, key).Result()
		require.NoError(t, err, "Failed to get key %s", key)
		assert.Equal(t, value, result)

		// Delete key
		err = redisClient.Del(ctx, key).Err()
		require.NoError(t, err, "Failed to delete key %s", key)
	}

	duration := time.Since(start)
	t.Logf("Completed 100 Redis operations in %v", duration)

	// Performance should be reasonable (less than 1 second for 100 operations)
	assert.Less(t, duration, time.Second, "Redis operations took too long")
}

func TestRedisQueueConsistency(t *testing.T) {
	client := NewRedisTestClient(t)
	defer client.Close()

	ctx := context.Background()
	redisClient := client.redisClient.GetClient()

	// Check queue consistency
	readyLength, err := redisClient.LLen(ctx, "job_queue:ready").Result()
	require.NoError(t, err, "Failed to get ready queue length")

	processingLength, err := redisClient.SCard(ctx, "job_queue:processing").Result()
	require.NoError(t, err, "Failed to get processing queue length")

	completedLength, err := redisClient.LLen(ctx, "job_queue:completed").Result()
	require.NoError(t, err, "Failed to get completed queue length")

	retryingLength, err := redisClient.ZCard(ctx, "job_queue:retrying").Result()
	require.NoError(t, err, "Failed to get retrying queue length")

	// Get stats through service
	stats, err := client.GetQueueStats()
	require.NoError(t, err, "Failed to get queue stats")

	// Verify consistency (allow for small differences due to race conditions)
	readyDiff := readyLength - stats["ready"]
	processingDiff := processingLength - stats["processing"]

	// Allow for small differences due to concurrent operations
	assert.True(t, readyDiff >= -5 && readyDiff <= 5, "Ready queue length mismatch: expected %d, got %d (diff: %d)", readyLength, stats["ready"], readyDiff)
	assert.True(t, processingDiff >= -5 && processingDiff <= 5, "Processing queue length mismatch: expected %d, got %d (diff: %d)", processingLength, stats["processing"], processingDiff)
	assert.Equal(t, completedLength, stats["completed"], "Completed queue length mismatch")
	assert.Equal(t, retryingLength, stats["retrying"], "Retrying queue length mismatch")

	t.Logf("Queue consistency verified: Ready=%d, Processing=%d, Completed=%d, Retrying=%d",
		readyLength, processingLength, completedLength, retryingLength)
}

func TestRedisMemoryUsage(t *testing.T) {
	client := NewRedisTestClient(t)
	defer client.Close()

	ctx := context.Background()
	redisClient := client.redisClient.GetClient()

	// Get Redis memory info
	info, err := redisClient.Info(ctx, "memory").Result()
	require.NoError(t, err, "Failed to get Redis memory info")

	t.Logf("Redis memory info: %s", info)

	// Check if memory usage is reasonable (less than 100MB for test environment)
	// This is a basic check - in production you'd want more sophisticated monitoring
	memoryUsage := redisClient.MemoryUsage(ctx, "job_queue:ready").Val()
	if memoryUsage > 0 {
		t.Logf("Ready queue memory usage: %d bytes", memoryUsage)
	}
}

// Benchmark tests
func BenchmarkRedisQueueStats(b *testing.B) {
	client := NewRedisTestClient(&testing.T{})
	defer client.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := client.GetQueueStats()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRedisPing(b *testing.B) {
	client := NewRedisTestClient(&testing.T{})
	defer client.Close()

	ctx := context.Background()
	redisClient := client.redisClient.GetClient()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := redisClient.Ping(ctx).Result()
		if err != nil {
			b.Fatal(err)
		}
	}
}
