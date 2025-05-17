package service

import (
	"fmt"
	"sync"
)

// StockService manages available item quantities.
type StockService struct {
	mu    sync.Mutex
	stock map[string]int
}

// NewStockService seeds the initial stock levels.
func NewStockService(initial map[string]int) *StockService {
	return &StockService{stock: initial}
}

// Reserve checks if all items are in stock.
// If yes, decrements quantities and returns (true, nil).
// If not, returns (false, error) and leaves stock unchanged.
func (s *StockService) Reserve(items []string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check availability
	for _, item := range items {
		qty, ok := s.stock[item]
		if !ok {
			return false, fmt.Errorf("item %q not recognized", item)
		}
		if qty < 1 {
			return false, fmt.Errorf("item %q out of stock", item)
		}
	}

	// All available → decrement
	for _, item := range items {
		s.stock[item]--
	}
	return true, nil
}
