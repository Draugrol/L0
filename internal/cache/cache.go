package cache

import (
	"context"
	"fmt"
	"log"
	"sync"

	"order-service/internal/models"
)

type OrderCache struct {
	mu     sync.RWMutex
	orders map[string]*models.Order
}

type Repository interface {
	GetAllOrders(ctx context.Context) ([]models.Order, error)
}

func NewOrderCache() *OrderCache {
	return &OrderCache{
		orders: make(map[string]*models.Order),
	}
}

func (c *OrderCache) Set(orderUID string, order *models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.orders[orderUID] = order
}

func (c *OrderCache) Get(orderUID string) (*models.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	order, exists := c.orders[orderUID]
	return order, exists
}

func (c *OrderCache) GetAll() []models.Order {
	c.mu.RLock()
	defer c.mu.RUnlock()

	orders := make([]models.Order, 0, len(c.orders))
	for _, order := range c.orders {
		orders = append(orders, *order)
	}
	return orders
}

func (c *OrderCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.orders)
}

func (c *OrderCache) RestoreFromDB(ctx context.Context, repo Repository) error {
	log.Println("Restoring cache from database...")

	orders, err := repo.GetAllOrders(ctx)
	if err != nil {
		return fmt.Errorf("failed to restore cache from DB: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for i := range orders {
		c.orders[orders[i].OrderUID] = &orders[i]
	}

	log.Printf("Cache restored successfully. Total orders in cache: %d\n", len(c.orders))
	return nil
}
