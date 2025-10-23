package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"order-service/internal/models"
)

type CacheService interface {
	Get(orderUID string) (*models.Order, bool)
	GetAll() []models.Order
	Size() int
}

type Server struct {
	cache CacheService
}

func NewServer(cache CacheService) *Server {
	return &Server{cache: cache}
}

func (s *Server) Start(port string) error {
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/orders", s.handleGetAllOrders)
	mux.HandleFunc("/api/orders/", s.handleGetOrder)
	mux.HandleFunc("/api/stats", s.handleStats)

	// Static files and UI
	mux.HandleFunc("/", s.handleIndex)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Printf("HTTP server starting on port %s", port)
	return http.ListenAndServe(":"+port, s.loggingMiddleware(mux))
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Order Service</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        h1 {
            color: white;
            text-align: center;
            margin-bottom: 30px;
            font-size: 2.5em;
            text-shadow: 2px 2px 4px rgba(0,0,0,0.2);
        }
        .search-box {
            background: white;
            padding: 20px;
            border-radius: 10px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
            margin-bottom: 20px;
            display: flex;
            gap: 10px;
        }
        .search-box input {
            flex: 1;
            padding: 15px;
            border: 2px solid #e0e0e0;
            border-radius: 5px;
            font-size: 16px;
        }
        .search-box button {
            padding: 15px 30px;
            background: #667eea;
            color: white;
            border: none;
            border-radius: 5px;
            font-size: 16px;
            cursor: pointer;
            transition: background 0.3s;
            white-space: nowrap;
        }
        .search-box button:hover {
            background: #5568d3;
        }
        .back-button {
            display: inline-block;
            padding: 10px 20px;
            background: #764ba2;
            color: white;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            margin-bottom: 15px;
            font-size: 14px;
            transition: background 0.3s;
        }
        .back-button:hover {
            background: #6a3f8f;
        }
        .stats {
            background: white;
            padding: 20px;
            border-radius: 10px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
            margin-bottom: 20px;
            text-align: center;
        }
        .stats h2 {
            color: #667eea;
            margin-bottom: 10px;
        }
        .result {
            background: white;
            padding: 20px;
            border-radius: 10px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
            margin-top: 20px;
        }
        .order-card {
            background: #f8f9fa;
            padding: 20px;
            border-radius: 8px;
            margin-bottom: 15px;
            border-left: 4px solid #667eea;
        }
        .order-list {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
            gap: 10px;
            margin-top: 15px;
        }
        .order-list-item {
            background: white;
            padding: 15px;
            border-radius: 5px;
            border: 2px solid #e0e0e0;
            cursor: pointer;
            transition: all 0.2s;
        }
        .order-list-item:hover {
            border-color: #667eea;
            transform: translateY(-2px);
            box-shadow: 0 4px 8px rgba(0,0,0,0.1);
        }
        .order-list-item .uid {
            font-family: monospace;
            color: #667eea;
            font-weight: bold;
            font-size: 0.9em;
            word-break: break-all;
        }
        .order-list-item .info {
            margin-top: 8px;
            font-size: 0.85em;
            color: #666;
        }
        code {
            background: #f0f0f0;
            padding: 2px 6px;
            border-radius: 3px;
            font-family: monospace;
            font-size: 0.9em;
        }
        .order-header {
            font-size: 1.2em;
            font-weight: bold;
            color: #667eea;
            margin-bottom: 10px;
        }
        .order-section {
            margin: 15px 0;
            padding: 10px;
            background: white;
            border-radius: 5px;
        }
        .order-section h3 {
            color: #764ba2;
            margin-bottom: 8px;
            font-size: 1.1em;
        }
        .order-field {
            margin: 5px 0;
            padding: 5px;
        }
        .order-field strong {
            color: #555;
        }
        .item-card {
            background: #e8eaf6;
            padding: 10px;
            margin: 10px 0;
            border-radius: 5px;
            border-left: 3px solid #764ba2;
        }
        .error {
            color: #d32f2f;
            padding: 15px;
            background: #ffebee;
            border-radius: 5px;
            margin: 10px 0;
        }
        pre {
            background: #f5f5f5;
            padding: 15px;
            border-radius: 5px;
            overflow-x: auto;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>📦 Order Service Dashboard</h1>

        <div class="stats">
            <h2>Cache Statistics</h2>
            <p id="stats-info">Loading...</p>
        </div>

        <div class="search-box">
            <input type="text" id="orderUID" placeholder="Enter Order UID (e.g., b563feb7b2b84b6test)" onkeypress="if(event.key==='Enter') searchOrder()">
            <button onclick="searchOrder()">Search Order</button>
        </div>

        <div id="result"></div>
    </div>

    <script>
        async function loadStats() {
            try {
                const response = await fetch('/api/stats');
                const data = await response.json();
                document.getElementById('stats-info').innerHTML =
                    '<strong>Total Orders in Cache: ' + data.total_orders + '</strong>';
            } catch (error) {
                document.getElementById('stats-info').innerHTML =
                    '<span class="error">Failed to load stats</span>';
            }
        }

        async function searchOrder() {
            const orderUID = document.getElementById('orderUID').value.trim();
            if (!orderUID) {
                alert('Please enter an Order UID');
                return;
            }

            try {
                const response = await fetch('/api/orders/' + orderUID);
                if (response.status === 404) {
                    document.getElementById('result').innerHTML =
                        '<div class="error">Order not found</div>';
                    return;
                }

                const order = await response.json();
                displayOrder(order);
            } catch (error) {
                document.getElementById('result').innerHTML =
                    '<div class="error">Error: ' + error.message + '</div>';
            }
        }

        async function loadAllOrders() {
            try {
                const response = await fetch('/api/orders');
                const orders = await response.json();

                if (orders.length === 0) {
                    document.getElementById('result').innerHTML =
                        '<div class="result"><p>No orders found</p></div>';
                    return;
                }

                let html = '<div class="result">';
                html += '<h2>All Orders (' + orders.length + ')</h2>';
                html += '<p style="color: #666; margin-bottom: 15px;">Click on any order to view details</p>';
                html += '<div class="order-list">';

                orders.forEach(order => {
                    html += '<div class="order-list-item" onclick="loadOrderByUID(\'' + order.order_uid + '\')">';
                    html += '<div class="uid">' + order.order_uid + '</div>';
                    html += '<div class="info">';
                    if (order.delivery && order.delivery.name) {
                        html += '👤 ' + order.delivery.name + '<br>';
                    }
                    if (order.delivery && order.delivery.city) {
                        html += '📍 ' + order.delivery.city;
                    }
                    html += '</div>';
                    html += '</div>';
                });

                html += '</div></div>';
                document.getElementById('result').innerHTML = html;
            } catch (error) {
                document.getElementById('result').innerHTML =
                    '<div class="error">Error: ' + error.message + '</div>';
            }
        }

        async function loadOrderByUID(uid) {
            try {
                const response = await fetch('/api/orders/' + uid);
                if (response.status === 404) {
                    document.getElementById('result').innerHTML =
                        '<div class="error">Order not found</div>';
                    return;
                }

                const order = await response.json();
                displayOrder(order);

                // Update search field
                document.getElementById('orderUID').value = uid;
            } catch (error) {
                document.getElementById('result').innerHTML =
                    '<div class="error">Error: ' + error.message + '</div>';
            }
        }

        function displayOrder(order) {
            document.getElementById('result').innerHTML =
                '<div class="result">' +
                '<button class="back-button" onclick="loadAllOrders()">← Назад к списку заказов</button>' +
                '<h2>Order Details</h2>' +
                formatOrderCard(order) + '</div>';
        }

        function formatOrderCard(order) {
            let html = '<div class="order-card">';
            html += '<div class="order-header">Order UID: ' + order.order_uid + '</div>';

            html += '<div class="order-section">';
            html += '<div class="order-field"><strong>Order UID:</strong> <code>' + order.order_uid + '</code></div>';
            html += '<div class="order-field"><strong>Track Number:</strong> ' + order.track_number + '</div>';
            html += '<div class="order-field"><strong>Entry:</strong> ' + order.entry + '</div>';
            html += '<div class="order-field"><strong>Customer ID:</strong> ' + order.customer_id + '</div>';
            html += '<div class="order-field"><strong>Delivery Service:</strong> ' + order.delivery_service + '</div>';
            html += '</div>';

            if (order.delivery) {
                html += '<div class="order-section"><h3>🚚 Delivery Information</h3>';
                html += '<div class="order-field"><strong>Name:</strong> ' + order.delivery.name + '</div>';
                html += '<div class="order-field"><strong>Phone:</strong> ' + order.delivery.phone + '</div>';
                html += '<div class="order-field"><strong>Address:</strong> ' + order.delivery.address + ', ' +
                        order.delivery.city + ', ' + order.delivery.region + '</div>';
                html += '<div class="order-field"><strong>Email:</strong> ' + order.delivery.email + '</div>';
                html += '</div>';
            }

            if (order.payment) {
                html += '<div class="order-section"><h3>💳 Payment Information</h3>';
                html += '<div class="order-field"><strong>Transaction:</strong> ' + order.payment.transaction + '</div>';
                html += '<div class="order-field"><strong>Amount:</strong> ' + order.payment.amount + ' ' + order.payment.currency + '</div>';
                html += '<div class="order-field"><strong>Provider:</strong> ' + order.payment.provider + '</div>';
                html += '<div class="order-field"><strong>Bank:</strong> ' + order.payment.bank + '</div>';
                html += '</div>';
            }

            if (order.items && order.items.length > 0) {
                html += '<div class="order-section"><h3>📦 Items (' + order.items.length + ')</h3>';
                order.items.forEach(item => {
                    html += '<div class="item-card">';
                    html += '<div><strong>' + item.name + '</strong> - ' + item.brand + '</div>';
                    html += '<div>Price: ' + item.price + ' | Sale: ' + item.sale + '% | Total: ' + item.total_price + '</div>';
                    html += '<div>Size: ' + item.size + ' | Status: ' + item.status + '</div>';
                    html += '</div>';
                });
                html += '</div>';
            }

            html += '</div>';
            return html;
        }

        // Load stats on page load
        loadStats();
        // Refresh stats every 5 seconds
        setInterval(loadStats, 5000);
        // Load all orders on page load
        loadAllOrders();
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

func (s *Server) handleGetOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orderUID := r.URL.Path[len("/api/orders/"):]
	if orderUID == "" {
		http.Error(w, "Order UID required", http.StatusBadRequest)
		return
	}

	order, exists := s.cache.Get(orderUID)
	if !exists {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (s *Server) handleGetAllOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orders := s.cache.GetAll()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := map[string]interface{}{
		"total_orders": s.cache.Size(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
