package redis

import (
	"os"
	"strconv"
	"time"
)

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host               string
	Port               string
	Password           string
	Database           int
	PoolSize           int
	MinIdleConns       int
	MaxRetries         int
	RetryDelay         time.Duration
	PoolTimeout        time.Duration
	IdleTimeout        time.Duration
	IdleCheckFrequency time.Duration
	MaxConnAge         time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	DialTimeout        time.Duration
}

// LoadRedisConfig loads Redis configuration from environment variables
func LoadRedisConfig() *RedisConfig {
	return &RedisConfig{
		Host:               getEnv("REDIS_HOST", "localhost"),
		Port:               getEnv("REDIS_PORT", "6379"),
		Password:           getEnv("REDIS_PASSWORD", ""),
		Database:           getIntEnv("REDIS_DATABASE", 0),
		PoolSize:           getIntEnv("REDIS_POOL_SIZE", 10),
		MinIdleConns:       getIntEnv("REDIS_MIN_IDLE_CONNS", 5),
		MaxRetries:         getIntEnv("REDIS_MAX_RETRIES", 3),
		RetryDelay:         getDurationEnv("REDIS_RETRY_DELAY", 100*time.Millisecond),
		PoolTimeout:        getDurationEnv("REDIS_POOL_TIMEOUT", 4*time.Second),
		IdleTimeout:        getDurationEnv("REDIS_IDLE_TIMEOUT", 5*time.Minute),
		IdleCheckFrequency: getDurationEnv("REDIS_IDLE_CHECK_FREQUENCY", 1*time.Minute),
		MaxConnAge:         getDurationEnv("REDIS_MAX_CONN_AGE", 0),
		ReadTimeout:        getDurationEnv("REDIS_READ_TIMEOUT", 3*time.Second),
		WriteTimeout:       getDurationEnv("REDIS_WRITE_TIMEOUT", 3*time.Second),
		DialTimeout:        getDurationEnv("REDIS_DIAL_TIMEOUT", 5*time.Second),
	}
}

// GetRedisAddress returns the full Redis address
func (c *RedisConfig) GetRedisAddress() string {
	return c.Host + ":" + c.Port
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}