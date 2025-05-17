package models

// OrderCreated is emitted by the Order service.
type OrderCreated struct {
	OrderID string   `json:"order_id"`
	UserID  string   `json:"user_id"`
	Items   []string `json:"items"`
	Total   float64  `json:"total"`
}

// InventoryReserved is emitted when stock reservation succeeds.
type InventoryReserved struct {
	OrderID string   `json:"order_id"`
	Items   []string `json:"items"`
}

// InventoryFailed is emitted when any item is out of stock.
type InventoryFailed struct {
	OrderID string   `json:"order_id"`
	Items   []string `json:"items"`
	Reason  string   `json:"reason"`
}
