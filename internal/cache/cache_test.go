package cache

import (
	"context"
	"testing"
	"time"

	"order-service/internal/models"
)

type mockRepository struct {
	orders []models.Order
	err    error
}

func (m *mockRepository) GetAllOrders(ctx context.Context) ([]models.Order, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.orders, nil
}

func TestNewOrderCache(t *testing.T) {
	cache := NewOrderCache()
	if cache == nil {
		t.Fatal("NewOrderCache returned nil")
	}
	if cache.Size() != 0 {
		t.Errorf("Expected cache size 0, got %d", cache.Size())
	}
}

func TestCacheSetAndGet(t *testing.T) {
	cache := NewOrderCache()

	order := &models.Order{
		OrderUID:    "test123",
		TrackNumber: "TRACK123",
		Entry:       "WBIL",
		CustomerID:  "customer1",
		DateCreated: time.Now(),
	}

	cache.Set(order.OrderUID, order)

	retrieved, exists := cache.Get(order.OrderUID)
	if !exists {
		t.Error("Expected order to exist in cache")
	}
	if retrieved.OrderUID != order.OrderUID {
		t.Errorf("Expected OrderUID %s, got %s", order.OrderUID, retrieved.OrderUID)
	}
}

func TestCacheGetNonExistent(t *testing.T) {
	cache := NewOrderCache()

	_, exists := cache.Get("nonexistent")
	if exists {
		t.Error("Expected order to not exist in cache")
	}
}

func TestCacheGetAll(t *testing.T) {
	cache := NewOrderCache()

	order1 := &models.Order{OrderUID: "order1", TrackNumber: "TRACK1", DateCreated: time.Now()}
	order2 := &models.Order{OrderUID: "order2", TrackNumber: "TRACK2", DateCreated: time.Now()}

	cache.Set(order1.OrderUID, order1)
	cache.Set(order2.OrderUID, order2)

	all := cache.GetAll()
	if len(all) != 2 {
		t.Errorf("Expected 2 orders, got %d", len(all))
	}
}

func TestCacheSize(t *testing.T) {
	cache := NewOrderCache()

	if cache.Size() != 0 {
		t.Errorf("Expected size 0, got %d", cache.Size())
	}

	cache.Set("order1", &models.Order{OrderUID: "order1", DateCreated: time.Now()})
	if cache.Size() != 1 {
		t.Errorf("Expected size 1, got %d", cache.Size())
	}

	cache.Set("order2", &models.Order{OrderUID: "order2", DateCreated: time.Now()})
	if cache.Size() != 2 {
		t.Errorf("Expected size 2, got %d", cache.Size())
	}
}

func TestRestoreFromDB(t *testing.T) {
	cache := NewOrderCache()

	mockOrders := []models.Order{
		{OrderUID: "order1", TrackNumber: "TRACK1", DateCreated: time.Now()},
		{OrderUID: "order2", TrackNumber: "TRACK2", DateCreated: time.Now()},
		{OrderUID: "order3", TrackNumber: "TRACK3", DateCreated: time.Now()},
	}

	mockRepo := &mockRepository{orders: mockOrders}

	err := cache.RestoreFromDB(context.Background(), mockRepo)
	if err != nil {
		t.Fatalf("RestoreFromDB failed: %v", err)
	}

	if cache.Size() != len(mockOrders) {
		t.Errorf("Expected cache size %d, got %d", len(mockOrders), cache.Size())
	}

	for _, order := range mockOrders {
		retrieved, exists := cache.Get(order.OrderUID)
		if !exists {
			t.Errorf("Order %s not found in cache after restore", order.OrderUID)
		}
		if retrieved.TrackNumber != order.TrackNumber {
			t.Errorf("Expected TrackNumber %s, got %s", order.TrackNumber, retrieved.TrackNumber)
		}
	}
}

func TestCacheConcurrency(t *testing.T) {
	cache := NewOrderCache()

	done := make(chan bool)
	numGoroutines := 100

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			order := &models.Order{
				OrderUID:    string(rune(id)),
				TrackNumber: string(rune(id)),
				DateCreated: time.Now(),
			}
			cache.Set(order.OrderUID, order)
			done <- true
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			cache.Get(string(rune(id)))
			done <- true
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}
