.PHONY: help build run test docker-up docker-down migrate clean stress-test stress-quick stress-medium stress-high stress-extreme publisher

help:
	@echo "Available commands:"
	@echo "  make docker-up      - Start PostgreSQL and NATS Streaming"
	@echo "  make docker-down    - Stop all containers"
	@echo "  make migrate        - Run database migrations"
	@echo "  make build          - Build the service"
	@echo "  make run            - Run the service"
	@echo "  make test           - Run tests"
	@echo "  make publisher      - Run test publisher"
	@echo "  make stress-test    - Run all stress tests (quick, medium, high)"
	@echo "  make stress-quick   - Run quick stress test (100 req/s)"
	@echo "  make stress-medium  - Run medium stress test (500 req/s)"
	@echo "  make stress-high    - Run high load test (1000 req/s)"
	@echo "  make stress-extreme - Run extreme load test (2000 req/s)"
	@echo "  make clean          - Clean build artifacts"

docker-up:
	docker-compose up -d
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 5

docker-down:
	docker-compose down

migrate: docker-up
	@echo "Running migrations..."
	@sleep 2
	PGPASSWORD=orderpass psql -h localhost -U orderuser -d ordersdb -f migrations/001_init_schema.sql
	@echo "Migrations completed"

build:
	go build -o bin/service cmd/service/main.go
	go build -o bin/publisher cmd/publisher/main.go

run: build
	./bin/service

test:
	go test -v -cover ./...

test-race:
	CGO_ENABLED=1 go test -v -race -cover ./...

publisher: build
	./bin/publisher

stress-test:
	@echo "======================================"
	@echo "Running all stress tests..."
	@echo "======================================"
	@echo ""
	@echo "Test 1: Quick load (100 req/s, 10s)"
	@echo "GET http://localhost:8080/api/orders" | vegeta attack -duration=10s -rate=100 | vegeta report
	@echo ""
	@echo "Test 2: Medium load (500 req/s, 30s)"
	@echo "GET http://localhost:8080/api/orders" | vegeta attack -duration=30s -rate=500 | vegeta report
	@echo ""
	@echo "Test 3: High load (1000 req/s, 30s)"
	@echo "GET http://localhost:8080/api/orders" | vegeta attack -duration=30s -rate=1000 | vegeta report
	@echo ""
	@echo "======================================"
	@echo "All stress tests completed!"
	@echo "======================================"

stress-quick:
	@echo "Running quick stress test (100 req/s, 10s)..."
	@echo "GET http://localhost:8080/api/orders" | vegeta attack -duration=10s -rate=100 | vegeta report

stress-medium:
	@echo "Running medium stress test (500 req/s, 30s)..."
	@echo "GET http://localhost:8080/api/orders" | vegeta attack -duration=30s -rate=500 | vegeta report

stress-high:
	@echo "Running high load stress test (1000 req/s, 30s)..."
	@echo "GET http://localhost:8080/api/orders" | vegeta attack -duration=30s -rate=1000 | vegeta report

stress-extreme:
	@echo "Running extreme load stress test (2000 req/s, 60s)..."
	@echo "GET http://localhost:8080/api/orders" | vegeta attack -duration=60s -rate=2000 | vegeta report

clean:
	rm -rf bin/
	go clean
