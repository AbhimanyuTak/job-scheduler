package tests

import (
	"os"
	"strconv"
	"testing"
	"time"
)

// TestConfig holds configuration for tests
type TestConfig struct {
	BaseURL      string
	APIKey       string
	RedisHost    string
	RedisPort    string
	TestTimeout  time.Duration
	MaxRetries   int
	WaitInterval time.Duration
}

// GetTestConfig returns test configuration with defaults and environment overrides
func GetTestConfig() *TestConfig {
	config := &TestConfig{
		BaseURL:      "http://localhost:8080",
		APIKey:       "test-api-key",
		RedisHost:    "localhost",
		RedisPort:    "6379",
		TestTimeout:  30 * time.Second,
		MaxRetries:   3,
		WaitInterval: 2 * time.Second,
	}

	// Override with environment variables if set
	if baseURL := os.Getenv("TEST_BASE_URL"); baseURL != "" {
		config.BaseURL = baseURL
	}

	if apiKey := os.Getenv("TEST_API_KEY"); apiKey != "" {
		config.APIKey = apiKey
	}

	if redisHost := os.Getenv("TEST_REDIS_HOST"); redisHost != "" {
		config.RedisHost = redisHost
	}

	if redisPort := os.Getenv("TEST_REDIS_PORT"); redisPort != "" {
		config.RedisPort = redisPort
	}

	if timeoutStr := os.Getenv("TEST_TIMEOUT"); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil {
			config.TestTimeout = time.Duration(timeout) * time.Second
		}
	}

	if retriesStr := os.Getenv("TEST_MAX_RETRIES"); retriesStr != "" {
		if retries, err := strconv.Atoi(retriesStr); err == nil {
			config.MaxRetries = retries
		}
	}

	if intervalStr := os.Getenv("TEST_WAIT_INTERVAL"); intervalStr != "" {
		if interval, err := strconv.Atoi(intervalStr); err == nil {
			config.WaitInterval = time.Duration(interval) * time.Second
		}
	}

	return config
}

// IsIntegrationTest returns true if running integration tests
func IsIntegrationTest() bool {
	return os.Getenv("INTEGRATION_TESTS") == "true"
}

// SkipIfNotIntegration skips the test if not running integration tests
func SkipIfNotIntegration(t testing.T) {
	if !IsIntegrationTest() {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=true to run.")
	}
}
