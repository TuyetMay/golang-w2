package kafka

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"asset-management-api/pkg/eventbus"
	
	"github.com/segmentio/kafka-go"
)

// KafkaConsumer implements EventBus interface for consuming messages
type KafkaConsumer struct {
	readers    map[string]*kafka.Reader
	config     *KafkaConfig
	handlers   map[string]eventbus.EventHandler
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// NewKafkaConsumer creates a new Kafka consumer
func NewKafkaConsumer(config *KafkaConfig) *KafkaConsumer {
	ctx, cancel := context.WithCancel(context.Background())
	return &KafkaConsumer{
		readers:  make(map[string]*kafka.Reader),
		handlers: make(map[string]eventbus.EventHandler),
		config:   config,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Publish is not implemented for consumer (only for producer)
func (c *KafkaConsumer) Publish(ctx context.Context, topic string, event interface{}) error {
	return fmt.Errorf("publish not supported by consumer")
}

// Subscribe starts consuming messages from the specified topic
func (c *KafkaConsumer) Subscribe(ctx context.Context, topic string, handler eventbus.EventHandler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already subscribed
	if _, exists := c.readers[topic]; exists {
		return fmt.Errorf("already subscribed to topic %s", topic)
	}

	// Create reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        c.config.Brokers,
		Topic:          topic,
		GroupID:        c.config.ConsumerConfig.GroupID,
		MinBytes:       1,
		MaxBytes:       10e6, // 10MB
		CommitInterval: c.config.ConsumerConfig.AutoCommitInterval,
		StartOffset:    kafka.LastOffset, // Start from latest messages
		Logger:         kafka.LoggerFunc(log.Printf),
		ErrorLogger:    kafka.LoggerFunc(log.Printf),
	})

	c.readers[topic] = reader
	c.handlers[topic] = handler

	// Start consuming in a separate goroutine
	c.wg.Add(1)
	go c.consumeMessages(topic, reader, handler)

	log.Printf("Subscribed to Kafka topic: %s", topic)
	return nil
}

// consumeMessages consumes messages from a topic in a separate goroutine
func (c *KafkaConsumer) consumeMessages(topic string, reader *kafka.Reader, handler eventbus.EventHandler) {
	defer c.wg.Done()
	
	for {
		select {
		case <-c.ctx.Done():
			log.Printf("Stopping consumer for topic %s", topic)
			return
		default:
			// Read message with timeout
			ctx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
			message, err := reader.ReadMessage(ctx)
			cancel()

			if err != nil {
				if err == context.DeadlineExceeded {
					continue // Timeout is expected, continue polling
				}
				if err == context.Canceled {
					return // Context was cancelled, stop consuming
				}
				log.Printf("Error reading message from topic %s: %v", topic, err)
				time.Sleep(1 * time.Second) // Wait before retrying
				continue
			}

			// Process message
			if err := c.processMessage(topic, message, handler); err != nil {
				log.Printf("Error processing message from topic %s: %v", topic, err)
				// In production, you might want to send failed messages to a dead letter queue
			}
		}
	}
}

// processMessage processes a single message with retry logic
func (c *KafkaConsumer) processMessage(topic string, message kafka.Message, handler eventbus.EventHandler) error {
	maxRetries := 3
	var err error

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Create context with timeout for handler execution
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		
		// Call the handler
		err = handler(ctx, message.Value)
		cancel()

		if err == nil {
			// Log successful processing
			log.Printf("Successfully processed message from topic %s, partition %d, offset %d", 
				topic, message.Partition, message.Offset)
			return nil
		}

		log.Printf("Attempt %d/%d failed for message from topic %s: %v", 
			attempt+1, maxRetries, topic, err)
		
		if attempt < maxRetries-1 {
			// Exponential backoff
			backoffTime := time.Duration(attempt+1) * time.Second
			time.Sleep(backoffTime)
		}
	}

	// All retries failed
	log.Printf("Failed to process message after %d attempts from topic %s: %v", 
		maxRetries, topic, err)
	
	// In production, you might want to send this to a dead letter queue
	c.logFailedMessage(topic, message, err)
	return err
}

// logFailedMessage logs details about failed message processing
func (c *KafkaConsumer) logFailedMessage(topic string, message kafka.Message, err error) {
	log.Printf("FAILED MESSAGE - Topic: %s, Partition: %d, Offset: %d, Error: %v, Message: %s",
		topic, message.Partition, message.Offset, err, string(message.Value))
}

// Close closes all readers and stops consuming
func (c *KafkaConsumer) Close() error {
	log.Println("Closing Kafka consumer...")
	
	// Cancel context to stop all consumers
	c.cancel()
	
	// Wait for all consumer goroutines to finish
	c.wg.Wait()
	
	c.mu.Lock()
	defer c.mu.Unlock()
	
	var lastErr error
	for topic, reader := range c.readers {
		if err := reader.Close(); err != nil {
			log.Printf("Error closing reader for topic %s: %v", topic, err)
			lastErr = err
		}
	}
	
	log.Println("Kafka consumer closed")
	return lastErr
}

// HealthCheck returns the health status of the consumer
func (c *KafkaConsumer) HealthCheck() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	health := map[string]interface{}{
		"status":           "healthy",
		"subscribed_topics": len(c.readers),
		"topics":           make([]string, 0, len(c.readers)),
	}
	
	for topic := range c.readers {
		health["topics"] = append(health["topics"].([]string), topic)
	}
	
	return health
}