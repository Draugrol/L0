@echo off
setlocal enabledelayedexpansion

echo ======================================
echo Order Service - Automated Setup
echo ======================================
echo.

REM Step 1: Check prerequisites
echo Step 1: Checking prerequisites...

where go >nul 2>nul
if %errorlevel% neq 0 (
    echo ERROR: Go is not installed
    exit /b 1
)
echo [OK] Go is installed

where docker >nul 2>nul
if %errorlevel% neq 0 (
    echo ERROR: Docker is not installed
    exit /b 1
)
echo [OK] Docker is installed

where docker-compose >nul 2>nul
if %errorlevel% neq 0 (
    echo ERROR: Docker Compose is not installed
    exit /b 1
)
echo [OK] Docker Compose is installed

where psql >nul 2>nul
if %errorlevel% neq 0 (
    echo WARNING: psql is not installed. Migration will be done via docker
    set PSQL_DOCKER=1
) else (
    echo [OK] psql is installed
    set PSQL_DOCKER=0
)

echo.

REM Step 2: Install Go dependencies
echo Step 2: Installing Go dependencies...
go mod download
echo [OK] Dependencies installed
echo.

REM Step 3: Start Docker containers
echo Step 3: Starting Docker containers...
docker-compose down -v 2>nul
docker-compose up -d

echo Waiting for PostgreSQL to be ready...
timeout /t 10 /nobreak >nul

echo [OK] Docker containers started
echo.

REM Step 4: Run migrations
echo Step 4: Running database migrations...
if %PSQL_DOCKER%==1 (
    docker cp migrations/001_init_schema.sql orders_postgres:/tmp/
    docker-compose exec -T postgres psql -U orderuser -d ordersdb -f /tmp/001_init_schema.sql
) else (
    set PGPASSWORD=orderpass
    psql -h localhost -U orderuser -d ordersdb -f migrations/001_init_schema.sql
)
echo [OK] Migrations completed
echo.

REM Step 5: Build the project
echo Step 5: Building the project...
if not exist bin mkdir bin
go build -o bin/service.exe cmd/service/main.go
go build -o bin/publisher.exe cmd/publisher/main.go
echo [OK] Build completed
echo.

REM Step 6: Run tests
echo Step 6: Running tests...
go test ./...
echo.

echo ======================================
echo Setup completed successfully!
echo ======================================
echo.
echo To start the service, run:
echo   bin\service.exe
echo.
echo To publish test data, run (in another terminal):
echo   bin\publisher.exe
echo.
echo Then open your browser at:
echo   http://localhost:8080
echo.

pause
