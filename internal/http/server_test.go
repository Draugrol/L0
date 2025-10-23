package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"order-service/internal/models"
)

type mockCache struct {
	orders map[string]*models.Order
}

func newMockCache() *mockCache {
	return &mockCache{
		orders: make(map[string]*models.Order),
	}
}

func (m *mockCache) Get(orderUID string) (*models.Order, bool) {
	order, exists := m.orders[orderUID]
	return order, exists
}

func (m *mockCache) GetAll() []models.Order {
	orders := make([]models.Order, 0, len(m.orders))
	for _, order := range m.orders {
		orders = append(orders, *order)
	}
	return orders
}

func (m *mockCache) Size() int {
	return len(m.orders)
}

func (m *mockCache) Set(orderUID string, order *models.Order) {
	m.orders[orderUID] = order
}

func TestHandleGetOrder(t *testing.T) {
	cache := newMockCache()
	testOrder := &models.Order{
		OrderUID:    "test123",
		TrackNumber: "TRACK123",
		Entry:       "WBIL",
		CustomerID:  "customer1",
		DateCreated: time.Now(),
	}
	cache.Set(testOrder.OrderUID, testOrder)

	server := NewServer(cache)

	req := httptest.NewRequest(http.MethodGet, "/api/orders/test123", nil)
	w := httptest.NewRecorder()

	server.handleGetOrder(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var order models.Order
	if err := json.NewDecoder(w.Body).Decode(&order); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if order.OrderUID != testOrder.OrderUID {
		t.Errorf("Expected OrderUID %s, got %s", testOrder.OrderUID, order.OrderUID)
	}
}

func TestHandleGetOrderNotFound(t *testing.T) {
	cache := newMockCache()
	server := NewServer(cache)

	req := httptest.NewRequest(http.MethodGet, "/api/orders/nonexistent", nil)
	w := httptest.NewRecorder()

	server.handleGetOrder(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHandleGetAllOrders(t *testing.T) {
	cache := newMockCache()
	cache.Set("order1", &models.Order{OrderUID: "order1", DateCreated: time.Now()})
	cache.Set("order2", &models.Order{OrderUID: "order2", DateCreated: time.Now()})

	server := NewServer(cache)

	req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
	w := httptest.NewRecorder()

	server.handleGetAllOrders(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var orders []models.Order
	if err := json.NewDecoder(w.Body).Decode(&orders); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(orders) != 2 {
		t.Errorf("Expected 2 orders, got %d", len(orders))
	}
}

func TestHandleStats(t *testing.T) {
	cache := newMockCache()
	cache.Set("order1", &models.Order{OrderUID: "order1", DateCreated: time.Now()})
	cache.Set("order2", &models.Order{OrderUID: "order2", DateCreated: time.Now()})
	cache.Set("order3", &models.Order{OrderUID: "order3", DateCreated: time.Now()})

	server := NewServer(cache)

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	w := httptest.NewRecorder()

	server.handleStats(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var stats map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&stats); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	totalOrders := int(stats["total_orders"].(float64))
	if totalOrders != 3 {
		t.Errorf("Expected 3 orders in stats, got %d", totalOrders)
	}
}

func TestHandleIndexPageLoads(t *testing.T) {
	cache := newMockCache()
	server := NewServer(cache)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	server.handleIndex(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Expected Content-Type text/html, got %s", contentType)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	cache := newMockCache()
	server := NewServer(cache)

	req := httptest.NewRequest(http.MethodPost, "/api/orders/test123", nil)
	w := httptest.NewRecorder()

	server.handleGetOrder(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}
