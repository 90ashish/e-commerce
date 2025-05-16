package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"e-commerce/common/models"
	"e-commerce/order/handler"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// fakePublisher records calls to Publish and can simulate errors.
type fakePublisher struct {
	called  bool
	payload models.OrderCreated
	err     error
}

func (f *fakePublisher) Publish(o models.OrderCreated) error {
	f.called = true
	f.payload = o
	return f.err
}

func TestCreateOrder_Success(t *testing.T) {
	// Prepare Gin in test mode
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a fake publisher that succeeds
	fp := &fakePublisher{}
	h := handler.NewOrderHandler(fp)
	router.POST("/orders", h.CreateOrder)

	// Build a valid order JSON
	order := models.OrderCreated{OrderID: "1", UserID: "u1", Items: []string{"foo"}, Total: 9.99}
	body, _ := json.Marshal(order)

	// Issue the request
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusAccepted, w.Code)
	assert.True(t, fp.called, "Publish should be called")
	assert.Equal(t, order, fp.payload, "Published payload should match request")
}

func TestCreateOrder_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	fp := &fakePublisher{}
	router.POST("/orders", handler.NewOrderHandler(fp).CreateOrder)

	// Send malformed JSON
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, fp.called, "Publish should not be called on bad JSON")
}

func TestCreateOrder_MissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	fp := &fakePublisher{}
	h := handler.NewOrderHandler(fp)
	router.POST("/orders", h.CreateOrder)

	// Missing items and total
	order := models.OrderCreated{OrderID: "2", UserID: "u2", Items: []string{}, Total: 0}
	body, _ := json.Marshal(order)
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, fp.called, "Publish should not be called with invalid fields")
}

func TestCreateOrder_PublishError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Publisher returns error
	fp := &fakePublisher{err: errors.New("kafka down")}
	h := handler.NewOrderHandler(fp)
	router.POST("/orders", h.CreateOrder)

	order := models.OrderCreated{OrderID: "3", UserID: "u3", Items: []string{"bar"}, Total: 5}
	body, _ := json.Marshal(order)
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.True(t, fp.called, "Publish should be called even if it errors")
}
