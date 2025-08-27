package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"asset-management-api/pkg/eventbus"
	
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
)

// KafkaProducer implements EventBus interface for producing messages
type KafkaProducer struct {
	writers map[string]*kafka.Writer
	config  *KafkaConfig
}

// NewKafkaProducer creates a new Kafka producer
func NewKafkaProducer(config *KafkaConfig) *KafkaProducer {
	return &KafkaProducer{
		writers: make(map[string]*kafka.Writer),
		config:  config,
	}
}

// Publish sends an event to the specified Kafka topic
func (p *KafkaProducer) Publish(ctx context.Context, topic string, event interface{}) error {
	writer, err := p.getWriter(topic)
	if err != nil {
		return fmt.Errorf("failed to get writer for topic %s: %w", topic, err)
	}

	// Serialize event to JSON
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create Kafka message
	message := kafka.Message{
		Topic:     topic,
		Value:     eventBytes,
		Time:      time.Now(),
		Headers: []kafka.Header{
			{Key: "content-type", Value: []byte("application/json")},
		},
	}

	// Add partition key if available (for ordering)
	if keyProvider, ok := event.(EventKeyProvider); ok {
		message.Key = []byte(keyProvider.GetPartitionKey())
	}

	// Write message
	err = writer.WriteMessages(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to write message to topic %s: %w", topic, err)
	}

	log.Printf("Published event to topic %s: %s", topic, string(eventBytes))
	return nil
}

// Subscribe is not implemented for producer (only for consumer)
func (p *KafkaProducer) Subscribe(ctx context.Context, topic string, handler eventbus.EventHandler) error {
	return fmt.Errorf("subscribe not supported by producer")
}

// getWriter returns or creates a writer for the specified topic
func (p *KafkaProducer) getWriter(topic string) (*kafka.Writer, error) {
	if writer, exists := p.writers[topic]; exists {
		return writer, nil
	}

	// Get compression codec
	var compressionCodec compress.Codec
	switch p.config.ProducerConfig.CompressionType {
	case "gzip":
		compressionCodec = compress.Gzip
	case "snappy":
		compressionCodec = compress.Snappy
	case "lz4":
		compressionCodec = compress.Lz4
	default:
		compressionCodec = nil // No compression
	}

	// Configure writer
	writer := &kafka.Writer{
		Addr:         kafka.TCP(p.config.Brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{}, // Balance messages across partitions
		RequiredAcks: kafka.RequiredAcks(p.config.ProducerConfig.RequiredAcks),
		BatchSize:    p.config.ProducerConfig.FlushMessages,
		BatchTimeout: p.config.ProducerConfig.FlushFrequency,
		ReadTimeout:  p.config.ProducerConfig.FlushTimeout,
		WriteTimeout: p.config.ProducerConfig.FlushTimeout,
		Compression:  compressionCodec,
		Logger:       kafka.LoggerFunc(log.Printf),
		ErrorLogger:  kafka.LoggerFunc(log.Printf),
	}

	// Enable idempotent writes for exactly-once semantics
	if p.config.ProducerConfig.IdempotentWrites {
		writer.RequiredAcks = kafka.RequireAll
	}

	p.writers[topic] = writer
	return writer, nil
}

// Close closes all writers
func (p *KafkaProducer) Close() error {
	var lastErr error
	for topic, writer := range p.writers {
		if err := writer.Close(); err != nil {
			log.Printf("Error closing writer for topic %s: %v", topic, err)
			lastErr = err
		}
	}
	return lastErr
}

// EventKeyProvider interface for events that need custom partitioning
type EventKeyProvider interface {
	GetPartitionKey() string
}

// Ensure team events implement EventKeyProvider for proper partitioning
func (e *BaseTeamEvent) GetPartitionKey() string {
	return e.TeamID.String()
}

func (e *BaseAssetEvent) GetPartitionKey() string {
	return e.AssetID.String()
}

// Add the event types import here (normally this would be in a separate file)
type BaseTeamEvent struct {
	TeamID string `json:"teamId"`
}

type BaseAssetEvent struct {
	AssetID string `json:"assetId"`
}