#!/bin/bash

# Script to run the entire stack: Docker + Service + Publisher

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "======================================"
echo "Starting Order Service Stack"
echo "======================================"
echo ""

# Check if containers are running
if ! docker-compose ps | grep -q "Up"; then
    echo -e "${YELLOW}Starting Docker containers...${NC}"
    docker-compose up -d
    echo "Waiting for services to be ready..."
    sleep 5
fi

# Check if service binary exists
if [ ! -f "bin/service" ]; then
    echo -e "${YELLOW}Building service...${NC}"
    go build -o bin/service cmd/service/main.go
fi

# Check if publisher binary exists
if [ ! -f "bin/publisher" ]; then
    echo -e "${YELLOW}Building publisher...${NC}"
    go build -o bin/publisher cmd/publisher/main.go
fi

echo -e "${GREEN}Starting Order Service...${NC}"
./bin/service &
SERVICE_PID=$!

echo "Waiting for service to start..."
sleep 3

echo -e "${GREEN}Publishing test orders...${NC}"
./bin/publisher

echo ""
echo "======================================"
echo -e "${GREEN}Stack is running!${NC}"
echo "======================================"
echo ""
echo "Service PID: $SERVICE_PID"
echo "Web Interface: http://localhost:8080"
echo ""
echo "Press Ctrl+C to stop the service"
echo ""

# Wait for interrupt
trap "echo 'Stopping...'; kill $SERVICE_PID; docker-compose stop; exit 0" INT TERM

wait $SERVICE_PID
