// order/handler/order_handler.go

package handler

import (
	"e-commerce/common/models"
	"e-commerce/order/producer"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// OrderHandler holds service dependencies.
type OrderHandler struct {
	Producer *producer.KafkaProducer
	Logger   *zap.Logger
}

// NewOrderHandler creates a handler with a Kafka producer and a zap logger.
func NewOrderHandler(p *producer.KafkaProducer, log *zap.Logger) *OrderHandler {
	return &OrderHandler{
		Producer: p,
		Logger:   log,
	}
}

// CreateOrder handles POST /orders.
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req models.OrderCreated

	// 1) Bind and validate JSON payload.
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Warn("Invalid JSON payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON payload"})
		return
	}
	if req.UserID == "" || len(req.Items) == 0 || req.Total <= 0 {
		h.Logger.Warn("Validation failed on payload", zap.Any("payload", req))
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id, items and total are required"})
		return
	}

	// 2) Publish to Kafka.
	if err := h.Producer.Publish(req); err != nil {
		h.Logger.Error("Failed to publish event", zap.Error(err), zap.String("orderID", req.OrderID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish event"})
		return
	}

	// 3) Success.
	h.Logger.Info("Order received", zap.String("orderID", req.OrderID), zap.String("userID", req.UserID))
	c.JSON(http.StatusAccepted, gin.H{"status": "order received", "order_id": req.OrderID})
}
