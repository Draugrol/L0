#!/bin/bash

echo "======================================"
echo "WRK Stress Test - Order Service"
echo "======================================"
echo ""

# Check if wrk is installed
if ! command -v wrk &> /dev/null; then
    echo "ERROR: wrk is not installed"
    echo "Install: https://github.com/wg/wrk"
    exit 1
fi

# Test 1: Low load
echo "Test 1: Low load (10 connections, 2 threads, 10 seconds)"
wrk -t2 -c10 -d10s http://localhost:8080/api/orders
echo ""

# Test 2: Medium load
echo "Test 2: Medium load (100 connections, 4 threads, 30 seconds)"
wrk -t4 -c100 -d30s http://localhost:8080/api/orders
echo ""

# Test 3: High load
echo "Test 3: High load (400 connections, 12 threads, 30 seconds)"
wrk -t12 -c400 -d30s http://localhost:8080/api/orders
echo ""

# Test 4: Extreme load
echo "Test 4: Extreme load (1000 connections, 12 threads, 60 seconds)"
wrk -t12 -c1000 -d60s http://localhost:8080/api/orders
echo ""

# Test specific order endpoint
echo "Test 5: Specific order lookup (200 connections, 8 threads, 30 seconds)"
wrk -t8 -c200 -d30s http://localhost:8080/api/orders/b563feb7b2b84b6test1
echo ""

echo "======================================"
echo "WRK Stress Test Completed"
echo "======================================"
