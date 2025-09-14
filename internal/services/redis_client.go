package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps the Redis client with connection management
type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisClient creates a new Redis client with provided configuration
func NewRedisClient(addr, password string, db int) (*RedisClient, error) {
	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
		// Connection pool settings
		PoolSize:     10,
		MinIdleConns: 5,
		// Timeout settings
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		// Retry settings
		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,
	})

	ctx := context.Background()

	// Test connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("Connected to Redis at %s (DB: %d)", addr, db)

	return &RedisClient{
		client: rdb,
		ctx:    ctx,
	}, nil
}

// GetClient returns the underlying Redis client
func (rc *RedisClient) GetClient() *redis.Client {
	return rc.client
}

// GetContext returns the context for Redis operations
func (rc *RedisClient) GetContext() context.Context {
	return rc.ctx
}

// Close closes the Redis connection
func (rc *RedisClient) Close() error {
	return rc.client.Close()
}

// Health checks if Redis is healthy
func (rc *RedisClient) Health() error {
	return rc.client.Ping(rc.ctx).Err()
}
