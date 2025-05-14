package models

// OrderCreated represents an order creation event
type OrderCreated struct {
	OrderID string   `json:"order_id"`
	UserID  string   `json:"user_id"`
	Items   []string `json:"items"`
	Total   float64  `json:"total"`
}
