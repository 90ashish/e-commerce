package producer

import "e-commerce/common/models"

// OrderPublisher defines methods to publish OrderCreated events
// Abstracting allows mocking in unit tests.
type OrderPublisher interface {
	// Publish sends an OrderCreated event to a broker
	Publish(models.OrderCreated) error
	// Close flushes and releases any resources (e.g. network connections)
	Close() error
}
