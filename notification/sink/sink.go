package sink

import (
	"e-commerce/common/models"
	"fmt"
)

// NotificationSink is the interface for sending notifications.
// You can implement this for console logging, SMTP email, webhooks, etc.
type NotificationSink interface {
	NotifyReserved(models.InventoryReserved) error
	NotifyFailed(models.InventoryFailed) error
}

// ConsoleSink writes notifications to stdout.
type ConsoleSink struct{}

// NewConsoleSink constructs a ConsoleSink.
func NewConsoleSink() *ConsoleSink {
	return &ConsoleSink{}
}

// NotifyReserved logs a successful reservation.
func (s *ConsoleSink) NotifyReserved(evt models.InventoryReserved) error {
	fmt.Printf("✅ Order %s confirmed for items %v\n", evt.OrderID, evt.Items)
	return nil
}

// NotifyFailed logs an out-of-stock notification.
func (s *ConsoleSink) NotifyFailed(evt models.InventoryFailed) error {
	fmt.Printf("❌ Order %s failed for items %v: reason: %s\n", evt.OrderID, evt.Items, evt.Reason)
	return nil
}

// EmailSink is a stub for sending real emails.
type EmailSink struct {
	// You could add SMTP client fields here (host, auth, etc.)
}

// NewEmailSink constructs an EmailSink.
func NewEmailSink() *EmailSink {
	return &EmailSink{}
}

// NotifyReserved would send a confirmation email (stub).
func (s *EmailSink) NotifyReserved(evt models.InventoryReserved) error {
	// TODO: integrate real SMTP client
	fmt.Printf("[EMAIL] To user of order %s: your order is confirmed\n", evt.OrderID)
	return nil
}

// NotifyFailed would send a failure email (stub).
func (s *EmailSink) NotifyFailed(evt models.InventoryFailed) error {
	fmt.Printf("[EMAIL] To user of order %s: your order failed — %s\n", evt.OrderID, evt.Reason)
	return nil
}
