@echo off
echo ======================================
echo Initializing Database with Sample Data
echo ======================================
echo.

REM Check if publisher exists
if not exist "bin\publisher.exe" (
    echo Building publisher...
    go build -o bin\publisher.exe cmd\publisher\main.go
    if %errorlevel% neq 0 (
        echo ERROR: Failed to build publisher
        exit /b 1
    )
)

REM Check if service is running
curl -s http://localhost:8080/api/stats >nul 2>&1
if %errorlevel% equ 0 (
    echo [OK] Service is running
) else (
    echo ERROR: Service is not running!
    echo Please start the service first: bin\service.exe
    pause
    exit /b 1
)

REM Check if NATS is running
docker ps | findstr orders_nats >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: NATS Streaming is not running!
    echo Please start Docker containers: docker-compose up -d
    pause
    exit /b 1
)

echo [OK] NATS Streaming is running
echo.

REM Run publisher
echo Publishing 40 sample orders to database...
bin\publisher.exe

echo.
echo ======================================
echo Database initialized successfully!
echo ======================================
echo.
echo You can now:
echo   1. Restart the service to load data from cache
echo   2. Open http://localhost:8080 to view orders
echo.

pause
