package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"order-service/internal/cache"
	httpserver "order-service/internal/http"
	"order-service/internal/nats"
	"order-service/internal/repository"
)

const (
	// Database configuration
	dbHost     = "localhost"
	dbPort     = "5432"
	dbUser     = "orderuser"
	dbPassword = "orderpass"
	dbName     = "ordersdb"

	// NATS Streaming configuration
	natsURL      = "nats://localhost:4222"
	natsCluster  = "test-cluster"
	natsClientID = "order-service"
	natsSubject  = "orders"

	// HTTP server configuration
	httpPort = "8080"
)

func main() {
	log.Println("Starting Order Service...")

	// Connect to PostgreSQL
	log.Println("Connecting to PostgreSQL...")
	db, err := repository.NewPostgresDB(dbHost, dbPort, dbUser, dbPassword, dbName)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Successfully connected to PostgreSQL")

	// Initialize repository
	repo := repository.NewOrderRepository(db)

	// Initialize cache
	orderCache := cache.NewOrderCache()

	// Restore cache from database
	ctx := context.Background()
	if err := orderCache.RestoreFromDB(ctx, repo); err != nil {
		log.Printf("Warning: Failed to restore cache from DB: %v", err)
	}

	// Connect to NATS Streaming with retry
	var subscriber *nats.Subscriber
	for i := 0; i < 10; i++ {
		subscriber, err = nats.NewSubscriber(natsURL, natsCluster, natsClientID, repo, orderCache)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to NATS Streaming (attempt %d/10): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Printf("Warning: Could not connect to NATS Streaming after 10 attempts: %v", err)
		log.Println("Service will start without NATS subscription")
	} else {
		defer subscriber.Close()
		log.Println("Successfully connected to NATS Streaming")

		// Subscribe to orders channel
		if err := subscriber.Subscribe(natsSubject); err != nil {
			log.Fatalf("Failed to subscribe to NATS subject: %v", err)
		}
	}

	// Start HTTP server
	server := httpserver.NewServer(orderCache)

	// Run HTTP server in goroutine
	go func() {
		if err := server.Start(httpPort); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	log.Printf("Service started successfully!")
	log.Printf("HTTP server: http://localhost:%s", httpPort)
	log.Printf("Cache size: %d orders", orderCache.Size())

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down gracefully...")
}
