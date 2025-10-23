package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"order-service/internal/models"

	"github.com/nats-io/stan.go"
)

type OrderHandler interface {
	SaveOrder(ctx context.Context, order *models.Order) error
}

type CacheHandler interface {
	Set(orderUID string, order *models.Order)
}

type Subscriber struct {
	sc           stan.Conn
	subscription stan.Subscription
	repo         OrderHandler
	cache        CacheHandler
}

func NewSubscriber(natsURL, clusterID, clientID string, repo OrderHandler, cache CacheHandler) (*Subscriber, error) {
	sc, err := stan.Connect(clusterID, clientID,
		stan.NatsURL(natsURL),
		stan.SetConnectionLostHandler(func(_ stan.Conn, err error) {
			log.Printf("Connection lost: %v", err)
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS Streaming: %w", err)
	}

	return &Subscriber{
		sc:    sc,
		repo:  repo,
		cache: cache,
	}, nil
}

func (s *Subscriber) Subscribe(subject string) error {
	sub, err := s.sc.Subscribe(subject, s.messageHandler,
		stan.SetManualAckMode(),
		stan.DurableName("order-service-durable"),
		stan.AckWait(30*time.Second),
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to subject %s: %w", subject, err)
	}

	s.subscription = sub
	log.Printf("Successfully subscribed to subject: %s", subject)
	return nil
}

func (s *Subscriber) messageHandler(msg *stan.Msg) {
	log.Printf("Received message: %s", string(msg.Data))

	var order models.Order
	if err := json.Unmarshal(msg.Data, &order); err != nil {
		log.Printf("Failed to unmarshal order: %v", err)
		return
	}

	// Set current timestamp if not provided
	if order.DateCreated.IsZero() {
		order.DateCreated = time.Now()
	}

	// Save to database
	ctx := context.Background()
	if err := s.repo.SaveOrder(ctx, &order); err != nil {
		log.Printf("Failed to save order to DB: %v", err)
		return
	}

	// Save to cache
	s.cache.Set(order.OrderUID, &order)

	log.Printf("Order %s processed successfully", order.OrderUID)

	// Acknowledge the message
	if err := msg.Ack(); err != nil {
		log.Printf("Failed to ack message: %v", err)
	}
}

func (s *Subscriber) Close() error {
	if s.subscription != nil {
		if err := s.subscription.Unsubscribe(); err != nil {
			return err
		}
	}
	return s.sc.Close()
}
