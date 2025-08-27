package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Kafka    KafkaConfig
	Redis    RedisConfig // NEW: Added Redis configuration
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type JWTConfig struct {
	SecretKey      string
	ExpirationTime time.Duration
}

type KafkaConfig struct {
	Enabled               bool
	Brokers               []string
	ProducerRetryMax      int
	ProducerRequiredAcks  int
	ProducerFlushTimeout  time.Duration
	ConsumerGroupID       string
	ConsumerSessionTimeout time.Duration
	AutoCommitInterval    time.Duration
}

// NEW: Redis configuration struct
type RedisConfig struct {
	Enabled            bool
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

func Load() (*Config, error) {
	// Load .env file if exists
	_ = godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8000"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 30*time.Second),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password123"),
			DBName:   getEnv("DB_NAME", "asset_db"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		JWT: JWTConfig{
			SecretKey:      getEnv("JWT_SECRET", "your-super-secret-key-change-in-production"),
			ExpirationTime: getDurationEnv("JWT_EXPIRATION", 24*time.Hour),
		},
		Kafka: KafkaConfig{
			Enabled:               getBoolEnv("KAFKA_ENABLED", true),
			Brokers:               getSliceEnv("KAFKA_BROKERS", []string{"localhost:9092"}),
			ProducerRetryMax:      getIntEnv("KAFKA_PRODUCER_RETRY_MAX", 3),
			ProducerRequiredAcks:  getIntEnv("KAFKA_PRODUCER_REQUIRED_ACKS", 1),
			ProducerFlushTimeout:  getDurationEnv("KAFKA_PRODUCER_FLUSH_TIMEOUT", 5*time.Second),
			ConsumerGroupID:       getEnv("KAFKA_CONSUMER_GROUP_ID", "asset-management-api"),
			ConsumerSessionTimeout: getDurationEnv("KAFKA_CONSUMER_SESSION_TIMEOUT", 30*time.Second),
			AutoCommitInterval:    getDurationEnv("KAFKA_CONSUMER_AUTO_COMMIT_INTERVAL", 1*time.Second),
		},
		// NEW: Redis configuration
		Redis: RedisConfig{
			Enabled:            getBoolEnv("REDIS_ENABLED", true),
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
		},
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
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

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getSliceEnv(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		var result []string
		for _, v := range splitAndTrim(value, ",") {
			if v != "" {
				result = append(result, v)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}

func splitAndTrim(s, sep string) []string {
	parts := make([]string, 0)
	for _, part := range strings.Split(s, sep) {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}