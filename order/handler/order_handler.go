package handler

import (
	"e-commerce/common/models"
	"e-commerce/order/producer"
	"net/http"

	"github.com/gin-gonic/gin"
)

// OrderHandler holds service dependencies
type OrderHandler struct {
	Producer *producer.KafkaProducer
}

// NewOrderHandler creates a handler
func NewOrderHandler(p *producer.KafkaProducer) *OrderHandler {
	return &OrderHandler{Producer: p}
}

// CreateOrder handles POST /orders
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req models.OrderCreated
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON payload"})
		return
	}
	if req.UserID == "" || len(req.Items) == 0 || req.Total <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id, items and total are required"})
		return
	}
	if err := h.Producer.Publish(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish event"})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"status": "order received", "order_id": req.OrderID})
}
