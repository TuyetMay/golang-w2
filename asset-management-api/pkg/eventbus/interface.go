
package eventbus

import "context"

// EventBus defines the interface for publishing and consuming events
type EventBus interface {
	// Publish sends an event to the specified topic
	Publish(ctx context.Context, topic string, event interface{}) error
	
	// Subscribe starts consuming events from the specified topic
	Subscribe(ctx context.Context, topic string, handler EventHandler) error
	
	// Close closes the event bus connections
	Close() error
}

// EventHandler defines the function signature for event handlers
type EventHandler func(ctx context.Context, event []byte) error

// Event represents a generic event structure
type Event struct {
	EventType string      `json:"eventType"`
	Timestamp string      `json:"timestamp"`
	Data      interface{} `json:"data"`
}