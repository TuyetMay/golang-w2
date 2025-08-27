package kafka

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers        []string
	ProducerConfig ProducerConfig
	ConsumerConfig ConsumerConfig
}

// ProducerConfig holds Kafka producer configuration
type ProducerConfig struct {
	RetryMax         int
	RequiredAcks     int
	FlushTimeout     time.Duration
	FlushFrequency   time.Duration
	FlushMessages    int
	CompressionType  string
	IdempotentWrites bool
}

// ConsumerConfig holds Kafka consumer configuration
type ConsumerConfig struct {
	GroupID        string
	SessionTimeout time.Duration
	HeartbeatInterval time.Duration
	RebalanceTimeout  time.Duration
	AutoCommit       bool
	AutoCommitInterval time.Duration
}

// LoadKafkaConfig loads Kafka configuration from environment variables
func LoadKafkaConfig() *KafkaConfig {
	return &KafkaConfig{
		Brokers: getBrokers(),
		ProducerConfig: ProducerConfig{
			RetryMax:         getIntEnv("KAFKA_PRODUCER_RETRY_MAX", 3),
			RequiredAcks:     getIntEnv("KAFKA_PRODUCER_REQUIRED_ACKS", 1),
			FlushTimeout:     getDurationEnv("KAFKA_PRODUCER_FLUSH_TIMEOUT", 5*time.Second),
			FlushFrequency:   getDurationEnv("KAFKA_PRODUCER_FLUSH_FREQUENCY", 100*time.Millisecond),
			FlushMessages:    getIntEnv("KAFKA_PRODUCER_FLUSH_MESSAGES", 100),
			CompressionType:  getEnv("KAFKA_PRODUCER_COMPRESSION", "snappy"),
			IdempotentWrites: getBoolEnv("KAFKA_PRODUCER_IDEMPOTENT", true),
		},
		ConsumerConfig: ConsumerConfig{
			GroupID:            getEnv("KAFKA_CONSUMER_GROUP_ID", "asset-management-api"),
			SessionTimeout:     getDurationEnv("KAFKA_CONSUMER_SESSION_TIMEOUT", 30*time.Second),
			HeartbeatInterval:  getDurationEnv("KAFKA_CONSUMER_HEARTBEAT_INTERVAL", 3*time.Second),
			RebalanceTimeout:   getDurationEnv("KAFKA_CONSUMER_REBALANCE_TIMEOUT", 60*time.Second),
			AutoCommit:         getBoolEnv("KAFKA_CONSUMER_AUTO_COMMIT", true),
			AutoCommitInterval: getDurationEnv("KAFKA_CONSUMER_AUTO_COMMIT_INTERVAL", 1*time.Second),
		},
	}
}

// getBrokers returns Kafka broker addresses from environment
func getBrokers() []string {
	brokers := getEnv("KAFKA_BROKERS", "localhost:9092")
	return strings.Split(brokers, ",")
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

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
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