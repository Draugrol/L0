package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"order-service/internal/models"

	"github.com/nats-io/stan.go"
)

const (
	natsURL      = "nats://localhost:4222"
	natsCluster  = "test-cluster"
	natsClientID = "order-publisher"
	natsSubject  = "orders"
)

var (
	names = []string{
		"Ivan Petrov", "Anna Smirnova", "Dmitry Ivanov", "Elena Kuznetsova",
		"Sergey Popov", "Maria Sokolova", "Alexander Volkov", "Olga Novikova",
		"Mikhail Fedorov", "Natalia Morozova", "Pavel Orlov", "Tatiana Egorova",
	}

	cities = []string{
		"Moscow", "Saint Petersburg", "Novosibirsk", "Yekaterinburg",
		"Kazan", "Nizhny Novgorod", "Chelyabinsk", "Samara",
		"Omsk", "Rostov-on-Don", "Ufa", "Krasnoyarsk",
	}

	addresses = []string{
		"Lenina St.", "Pushkina St.", "Kirova St.", "Sovetskaya St.",
		"Gagarina St.", "Mira St.", "Komsomolskaya St.", "Pobedy St.",
	}

	products = []struct {
		name  string
		brand string
		price int
	}{
		{"Mascaras", "Vivienne Sabo", 453},
		{"Lipstick", "MAC", 890},
		{"Foundation", "Maybelline", 650},
		{"Perfume", "Chanel", 3500},
		{"Shampoo", "L'Oreal", 320},
		{"Face Cream", "Nivea", 450},
		{"Sunglasses", "Ray-Ban", 5200},
		{"Watch", "Casio", 2800},
		{"Sneakers", "Nike", 4500},
		{"T-Shirt", "Adidas", 1200},
		{"Jeans", "Levi's", 3200},
		{"Backpack", "Puma", 2100},
	}

	currencies = []string{"USD", "EUR", "RUB"}
	banks      = []string{"alpha", "sberbank", "tinkoff", "vtb", "raiffeisen"}
	providers  = []string{"wbpay", "yandexpay", "sberpay", "cloudpayments"}
	deliveries = []string{"meest", "cdek", "boxberry", "pochta", "dhl"}
	entries    = []string{"WBIL", "WBMSK", "WBSPB", "WBNSK"}
)

func main() {
	log.Println("Starting NATS Publisher...")
	rand.Seed(time.Now().UnixNano())

	// Connect to NATS Streaming
	sc, err := stan.Connect(natsCluster, natsClientID, stan.NatsURL(natsURL))
	if err != nil {
		log.Fatalf("Failed to connect to NATS Streaming: %v", err)
	}
	defer sc.Close()

	log.Println("Connected to NATS Streaming")

	// Generate and publish sample orders
	numOrders := 40
	for i := 1; i <= numOrders; i++ {
		order := generateSampleOrder(i)

		data, err := json.Marshal(order)
		if err != nil {
			log.Printf("Failed to marshal order: %v", err)
			continue
		}

		err = sc.Publish(natsSubject, data)
		if err != nil {
			log.Printf("Failed to publish order: %v", err)
			continue
		}

		log.Printf("Published order: %s (Customer: %s)", order.OrderUID, order.Delivery.Name)
		time.Sleep(300 * time.Millisecond)
	}

	log.Printf("All %d orders published successfully", numOrders)
}

func generateSampleOrder(num int) models.Order {
	orderUID := fmt.Sprintf("b563feb7b2b84b6test%d", num)
	trackNumber := fmt.Sprintf("WBILMTESTTRACK%05d", num)

	// Random selections
	name := names[rand.Intn(len(names))]
	city := cities[rand.Intn(len(cities))]
	address := fmt.Sprintf("%s %d", addresses[rand.Intn(len(addresses))], rand.Intn(200)+1)
	currency := currencies[rand.Intn(len(currencies))]
	bank := banks[rand.Intn(len(banks))]
	provider := providers[rand.Intn(len(providers))]
	delivery := deliveries[rand.Intn(len(deliveries))]
	entry := entries[rand.Intn(len(entries))]

	// Generate random items (1-4 items per order)
	numItems := rand.Intn(4) + 1
	items := make([]models.Item, numItems)
	totalAmount := 0
	goodsTotal := 0

	for j := 0; j < numItems; j++ {
		product := products[rand.Intn(len(products))]
		sale := rand.Intn(50) + 10 // 10-60% sale
		totalPrice := product.price * (100 - sale) / 100

		items[j] = models.Item{
			ChrtID:      9934930 + num*100 + j,
			TrackNumber: trackNumber,
			Price:       product.price,
			Rid:         fmt.Sprintf("ab4219087a764ae0btest%d%d", num, j),
			Name:        product.name,
			Sale:        sale,
			Size:        fmt.Sprintf("%d", rand.Intn(10)),
			TotalPrice:  totalPrice,
			NmID:        2389212 + num*10 + j,
			Brand:       product.brand,
			Status:      202,
		}
		goodsTotal += totalPrice
	}

	deliveryCost := rand.Intn(1000) + 500 // 500-1500
	totalAmount = goodsTotal + deliveryCost

	// Generate phone number
	phone := fmt.Sprintf("+7%d%d%d%d%d%d%d%d%d%d",
		rand.Intn(10), rand.Intn(10), rand.Intn(10), rand.Intn(10), rand.Intn(10),
		rand.Intn(10), rand.Intn(10), rand.Intn(10), rand.Intn(10), rand.Intn(10))

	// Generate email
	emailPrefix := fmt.Sprintf("customer%d", num)
	emailDomain := []string{"gmail.com", "yandex.ru", "mail.ru", "outlook.com"}
	email := fmt.Sprintf("%s@%s", emailPrefix, emailDomain[rand.Intn(len(emailDomain))])

	return models.Order{
		OrderUID:    orderUID,
		TrackNumber: trackNumber,
		Entry:       entry,
		Delivery: models.Delivery{
			Name:    name,
			Phone:   phone,
			Zip:     fmt.Sprintf("%d", 100000+rand.Intn(900000)),
			City:    city,
			Address: address,
			Region:  city + " Region",
			Email:   email,
		},
		Payment: models.Payment{
			Transaction:  orderUID,
			RequestID:    "",
			Currency:     currency,
			Provider:     provider,
			Amount:       totalAmount,
			PaymentDt:    time.Now().Add(-time.Duration(rand.Intn(720)) * time.Hour).Unix(), // Random time in last 30 days
			Bank:         bank,
			DeliveryCost: deliveryCost,
			GoodsTotal:   goodsTotal,
			CustomFee:    rand.Intn(300),
		},
		Items:             items,
		Locale:            []string{"en", "ru"}[rand.Intn(2)],
		InternalSignature: "",
		CustomerID:        fmt.Sprintf("customer%d", num),
		DeliveryService:   delivery,
		Shardkey:          fmt.Sprintf("%d", rand.Intn(10)),
		SmID:              num,
		DateCreated:       time.Now().Add(-time.Duration(rand.Intn(720)) * time.Hour), // Random time in last 30 days
		OofShard:          fmt.Sprintf("%d", rand.Intn(5)),
	}
}
