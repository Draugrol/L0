#!/bin/bash

set -e

echo "======================================"
echo "Order Service - Automated Setup"
echo "======================================"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Step 1: Check prerequisites
echo -e "${YELLOW}Step 1: Checking prerequisites...${NC}"

if ! command -v go &> /dev/null; then
    echo -e "${RED}ERROR: Go is not installed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Go is installed$(go version)${NC}"

if ! command -v docker &> /dev/null; then
    echo -e "${RED}ERROR: Docker is not installed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Docker is installed${NC}"

if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}ERROR: Docker Compose is not installed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Docker Compose is installed${NC}"

if ! command -v psql &> /dev/null; then
    echo -e "${YELLOW}WARNING: psql is not installed. Migration will be done via docker${NC}"
    PSQL_DOCKER=1
else
    echo -e "${GREEN}✓ psql is installed${NC}"
    PSQL_DOCKER=0
fi

echo ""

# Step 2: Install Go dependencies
echo -e "${YELLOW}Step 2: Installing Go dependencies...${NC}"
go mod download
echo -e "${GREEN}✓ Dependencies installed${NC}"
echo ""

# Step 3: Start Docker containers
echo -e "${YELLOW}Step 3: Starting Docker containers...${NC}"
docker-compose down -v 2>/dev/null || true
docker-compose up -d

echo "Waiting for PostgreSQL to be ready..."
for i in {1..30}; do
    if docker-compose exec -T postgres pg_isready -U orderuser &> /dev/null; then
        echo -e "${GREEN}✓ PostgreSQL is ready${NC}"
        break
    fi
    if [ $i -eq 30 ]; then
        echo -e "${RED}ERROR: PostgreSQL failed to start${NC}"
        exit 1
    fi
    echo -n "."
    sleep 1
done
echo ""

# Step 4: Run migrations
echo -e "${YELLOW}Step 4: Running database migrations...${NC}"
if [ $PSQL_DOCKER -eq 1 ]; then
    docker cp migrations/001_init_schema.sql orders_postgres:/tmp/
    docker-compose exec -T postgres psql -U orderuser -d ordersdb -f /tmp/001_init_schema.sql
else
    PGPASSWORD=orderpass psql -h localhost -U orderuser -d ordersdb -f migrations/001_init_schema.sql
fi
echo -e "${GREEN}✓ Migrations completed${NC}"
echo ""

# Step 5: Build the project
echo -e "${YELLOW}Step 5: Building the project...${NC}"
go build -o bin/service cmd/service/main.go
go build -o bin/publisher cmd/publisher/main.go
echo -e "${GREEN}✓ Build completed${NC}"
echo ""

# Step 6: Run tests
echo -e "${YELLOW}Step 6: Running tests...${NC}"
go test -v ./... || echo -e "${YELLOW}Some tests failed, but continuing...${NC}"
echo ""

echo "======================================"
echo -e "${GREEN}Setup completed successfully!${NC}"
echo "======================================"
echo ""
echo "To start the service, run:"
echo "  ./bin/service"
echo ""
echo "To publish test data, run (in another terminal):"
echo "  ./bin/publisher"
echo ""
echo "Then open your browser at:"
echo "  http://localhost:8080"
echo ""
