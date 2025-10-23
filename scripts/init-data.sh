#!/bin/bash

echo "======================================"
echo "Initializing Database with Sample Data"
echo "======================================"
echo ""

# Check if publisher exists
if [ ! -f "bin/publisher" ] && [ ! -f "bin/publisher.exe" ]; then
    echo "Building publisher..."
    go build -o bin/publisher cmd/publisher/main.go || go build -o bin/publisher.exe cmd/publisher/main.go
fi

# Check if service is running
if curl -s http://localhost:8080/api/stats > /dev/null 2>&1; then
    echo "✓ Service is running"
else
    echo "ERROR: Service is not running!"
    echo "Please start the service first: ./bin/service.exe"
    exit 1
fi

# Check if NATS is running
if ! docker ps | grep -q orders_nats; then
    echo "ERROR: NATS Streaming is not running!"
    echo "Please start Docker containers: docker-compose up -d"
    exit 1
fi

echo "✓ NATS Streaming is running"
echo ""

# Run publisher
echo "Publishing 40 sample orders to database..."
if [ -f "bin/publisher.exe" ]; then
    ./bin/publisher.exe
else
    ./bin/publisher
fi

echo ""
echo "======================================"
echo "Database initialized successfully!"
echo "======================================"
echo ""
echo "You can now:"
echo "  1. Restart the service to load data from cache"
echo "  2. Open http://localhost:8080 to view orders"
echo ""
