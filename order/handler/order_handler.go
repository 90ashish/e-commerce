package handler

import (
	"net/http"

	"e-commerce/common/models"

	"github.com/gin-gonic/gin"
)

// OrderPublisher abstracts the Publish and Close methods for event publishing.
// This matches the interface in producer/publisher.go.
type OrderPublisher interface {
	Publish(models.OrderCreated) error
}

// OrderHandler holds dependencies for handling HTTP requests.
// It depends on OrderPublisher, allowing for a mock in tests.
type OrderHandler struct {
	Publisher OrderPublisher
}

// NewOrderHandler creates a new OrderHandler with the given publisher.
func NewOrderHandler(p OrderPublisher) *OrderHandler {
	return &OrderHandler{Publisher: p}
}

// CreateOrder handles POST /orders
// 1. Bind JSON to OrderCreated
// 2. Validate fields
// 3. Publish event via Publisher
// 4. Return appropriate HTTP status and JSON
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req models.OrderCreated
	// Bind incoming JSON into struct. On error, respond 400.
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON payload"})
		return
	}
	// Basic business validation
	if req.UserID == "" || len(req.Items) == 0 || req.Total <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id, items and total are required"})
		return
	}
	// Publish the order event. On error, respond 500.
	if err := h.Publisher.Publish(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish event"})
		return
	}
	// Successful acceptance: echo back order_id
	c.JSON(http.StatusAccepted, gin.H{"status": "order received", "order_id": req.OrderID})
}
