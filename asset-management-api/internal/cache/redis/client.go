package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps the Redis client with additional functionality
type RedisClient struct {
	client *redis.Client
	config *RedisConfig
}

// NewRedisClient creates a new Redis client instance
func NewRedisClient(config *RedisConfig) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:               config.GetRedisAddress(),
		Password:           config.Password,
		DB:                 config.Database,
		PoolSize:           config.PoolSize,
		MinIdleConns:       config.MinIdleConns,
		MaxRetries:         config.MaxRetries,
		RetryDelayFunc:     func(retry int, err error) time.Duration { return config.RetryDelay },
		PoolTimeout:        config.PoolTimeout,
		IdleTimeout:        config.IdleTimeout,
		IdleCheckFrequency: config.IdleCheckFrequency,
		MaxConnAge:         config.MaxConnAge,
		ReadTimeout:        config.ReadTimeout,
		WriteTimeout:       config.WriteTimeout,
		DialTimeout:        config.DialTimeout,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("Successfully connected to Redis at %s", config.GetRedisAddress())
	
	return &RedisClient{
		client: rdb,
		config: config,
	}, nil
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// Health returns the health status of Redis connection
func (r *RedisClient) Health() map[string]interface{} {
	ctx := context.Background()
	
	start := time.Now()
	err := r.client.Ping(ctx).Err()
	latency := time.Since(start)
	
	health := map[string]interface{}{
		"status":     "healthy",
		"latency_ms": latency.Milliseconds(),
		"address":    r.config.GetRedisAddress(),
		"database":   r.config.Database,
	}
	
	if err != nil {
		health["status"] = "unhealthy"
		health["error"] = err.Error()
	}
	
	return health
}

// Generic methods for basic operations
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisClient) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	return count > 0, err
}

// JSON methods
func (r *RedisClient) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return r.Set(ctx, key, jsonBytes, expiration)
}

func (r *RedisClient) GetJSON(ctx context.Context, key string, dest interface{}) error {
	jsonStr, err := r.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(jsonStr), dest)
}

// List methods
func (r *RedisClient) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.LPush(ctx, key, values...).Err()
}

func (r *RedisClient) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.LRange(ctx, key, start, stop).Result()
}

func (r *RedisClient) LRem(ctx context.Context, key string, count int64, value interface{}) error {
	return r.client.LRem(ctx, key, count, value).Err()
}

func (r *RedisClient) LLen(ctx context.Context, key string) (int64, error) {
	return r.client.LLen(ctx, key).Result()
}

// Hash methods
func (r *RedisClient) HSet(ctx context.Context, key string, values ...interface{}) error {
	return r.client.HSet(ctx, key, values...).Err()
}

func (r *RedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

func (r *RedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

func (r *RedisClient) HDel(ctx context.Context, key string, fields ...string) error {
	return r.client.HDel(ctx, key, fields...).Err()
}

func (r *RedisClient) HExists(ctx context.Context, key, field string) (bool, error) {
	return r.client.HExists(ctx, key, field).Result()
}

// Set expiration
func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// Pipeline for batch operations
func (r *RedisClient) Pipeline() redis.Pipeliner {
	return r.client.Pipeline()
}

// Transaction support
func (r *RedisClient) TxPipeline() redis.Pipeliner {
	return r.client.TxPipeline()
}