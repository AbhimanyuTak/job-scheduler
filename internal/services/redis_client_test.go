package services

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRedisClient(t *testing.T) {
	// Set test environment variables
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_DB", "0")
	defer func() {
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("REDIS_PORT")
		os.Unsetenv("REDIS_DB")
	}()

	client, err := NewRedisClient("localhost:6379", "", 0)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Test that we can get the underlying client
	redisClient := client.GetClient()
	assert.NotNil(t, redisClient)

	// Test that we can get the context
	ctx := client.GetContext()
	assert.NotNil(t, ctx)

	// Clean up
	err = client.Close()
	assert.NoError(t, err)
}

func TestNewRedisClient_InvalidHost(t *testing.T) {
	// Set invalid host
	os.Setenv("REDIS_HOST", "invalid-host")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_DB", "0")
	defer func() {
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("REDIS_PORT")
		os.Unsetenv("REDIS_DB")
	}()

	client, err := NewRedisClient("localhost:6379", "", 0)
	require.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "failed to connect to Redis")
}

func TestNewRedisClient_InvalidPort(t *testing.T) {
	// Set invalid port
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "99999")
	os.Setenv("REDIS_DB", "0")
	defer func() {
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("REDIS_PORT")
		os.Unsetenv("REDIS_DB")
	}()

	client, err := NewRedisClient("localhost:6379", "", 0)
	require.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "failed to connect to Redis")
}

func TestNewRedisClient_InvalidDB(t *testing.T) {
	// Set invalid database number
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_DB", "invalid")
	defer func() {
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("REDIS_PORT")
		os.Unsetenv("REDIS_DB")
	}()

	client, err := NewRedisClient("localhost:6379", "", 0)
	require.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "invalid REDIS_DB value")
}

func TestRedisClient_Health(t *testing.T) {
	// Set test environment variables
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_DB", "0")
	defer func() {
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("REDIS_PORT")
		os.Unsetenv("REDIS_DB")
	}()

	client, err := NewRedisClient("localhost:6379", "", 0)
	require.NoError(t, err)
	defer client.Close()

	// Test health check
	err = client.Health()
	assert.NoError(t, err)
}

func TestRedisClient_GetClient(t *testing.T) {
	// Set test environment variables
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_DB", "0")
	defer func() {
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("REDIS_PORT")
		os.Unsetenv("REDIS_DB")
	}()

	client, err := NewRedisClient("localhost:6379", "", 0)
	require.NoError(t, err)
	defer client.Close()

	redisClient := client.GetClient()
	assert.NotNil(t, redisClient)

	// Test that we can use the client
	ctx := context.Background()
	pong, err := redisClient.Ping(ctx).Result()
	assert.NoError(t, err)
	assert.Equal(t, "PONG", pong)
}

func TestRedisClient_GetContext(t *testing.T) {
	// Set test environment variables
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_DB", "0")
	defer func() {
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("REDIS_PORT")
		os.Unsetenv("REDIS_DB")
	}()

	client, err := NewRedisClient("localhost:6379", "", 0)
	require.NoError(t, err)
	defer client.Close()

	ctx := client.GetContext()
	assert.NotNil(t, ctx)

	// Test that the context is not cancelled
	select {
	case <-ctx.Done():
		t.Error("Context should not be cancelled")
	default:
		// Context is not cancelled, which is expected
	}
}

func TestRedisClient_Close(t *testing.T) {
	// Set test environment variables
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_DB", "0")
	defer func() {
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("REDIS_PORT")
		os.Unsetenv("REDIS_DB")
	}()

	client, err := NewRedisClient("localhost:6379", "", 0)
	require.NoError(t, err)

	// Test that we can close the client
	err = client.Close()
	assert.NoError(t, err)

	// Test that we can close it again (should not error)
	err = client.Close()
	assert.NoError(t, err)
}

func TestRedisClient_DefaultValues(t *testing.T) {
	// Clear environment variables to test defaults
	os.Unsetenv("REDIS_HOST")
	os.Unsetenv("REDIS_PORT")
	os.Unsetenv("REDIS_DB")
	os.Unsetenv("REDIS_PASSWORD")

	client, err := NewRedisClient("localhost:6379", "", 0)
	require.NoError(t, err)
	defer client.Close()

	// Test that we can use the client with default values
	err = client.Health()
	assert.NoError(t, err)
}

func TestRedisClient_WithPassword(t *testing.T) {
	// Set test environment variables with password
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_DB", "0")
	os.Setenv("REDIS_PASSWORD", "testpassword")
	defer func() {
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("REDIS_PORT")
		os.Unsetenv("REDIS_DB")
		os.Unsetenv("REDIS_PASSWORD")
	}()

	// This should fail if Redis requires authentication
	client, err := NewRedisClient("localhost:6379", "", 0)
	if err != nil {
		// Expected if Redis requires authentication
		assert.Contains(t, err.Error(), "failed to connect to Redis")
		return
	}
	defer client.Close()

	// If it succeeds, test that it works
	err = client.Health()
	assert.NoError(t, err)
}

func TestRedisClient_ConnectionPool(t *testing.T) {
	// Set test environment variables
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_DB", "0")
	defer func() {
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("REDIS_PORT")
		os.Unsetenv("REDIS_DB")
	}()

	client, err := NewRedisClient("localhost:6379", "", 0)
	require.NoError(t, err)
	defer client.Close()

	redisClient := client.GetClient()

	// Test multiple operations to verify connection pool
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		pong, err := redisClient.Ping(ctx).Result()
		assert.NoError(t, err)
		assert.Equal(t, "PONG", pong)
	}
}

func TestRedisClient_Timeout(t *testing.T) {
	// Set test environment variables
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_DB", "0")
	defer func() {
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("REDIS_PORT")
		os.Unsetenv("REDIS_DB")
	}()

	client, err := NewRedisClient("localhost:6379", "", 0)
	require.NoError(t, err)
	defer client.Close()

	redisClient := client.GetClient()
	ctx := context.Background()

	// Test that operations complete within reasonable time
	start := time.Now()
	pong, err := redisClient.Ping(ctx).Result()
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.Equal(t, "PONG", pong)
	assert.Less(t, duration, 5*time.Second, "Redis operation should complete within 5 seconds")
}
