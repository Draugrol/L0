#!/bin/bash

echo "======================================"
echo "Vegeta Stress Test - Order Service"
echo "======================================"
echo ""

# Check if vegeta is installed
if ! command -v vegeta &> /dev/null; then
    echo "ERROR: vegeta is not installed"
    echo "Install: go install github.com/tsenart/vegeta@latest"
    exit 1
fi

# Create targets file
cat > targets.txt << EOF
GET http://localhost:8080/api/orders
GET http://localhost:8080/api/stats
GET http://localhost:8080/api/orders/b563feb7b2b84b6test1
EOF

# Test 1: Low rate
echo "Test 1: Low rate (100 req/s for 10s)"
vegeta attack -targets=targets.txt -rate=100 -duration=10s | vegeta report -type=text
echo ""

# Test 2: Medium rate
echo "Test 2: Medium rate (500 req/s for 30s)"
vegeta attack -targets=targets.txt -rate=500 -duration=30s | vegeta report -type=text
echo ""

# Test 3: High rate
echo "Test 3: High rate (1000 req/s for 30s)"
vegeta attack -targets=targets.txt -rate=1000 -duration=30s | vegeta report -type=text
echo ""

# Test 4: Very high rate
echo "Test 4: Very high rate (2000 req/s for 30s)"
vegeta attack -targets=targets.txt -rate=2000 -duration=30s | vegeta report -type=text
echo ""

# Test 5: Generate detailed report with plots
echo "Test 5: Generating detailed report (1000 req/s for 60s)"
vegeta attack -targets=targets.txt -rate=1000 -duration=60s > results.bin
vegeta report -type=text results.bin
echo ""
echo "Generating HTML report..."
vegeta plot results.bin > plot.html
vegeta report -type=json results.bin > report.json
echo "HTML report saved to: plot.html"
echo "JSON report saved to: report.json"
echo ""

# Cleanup
rm targets.txt

echo "======================================"
echo "Vegeta Stress Test Completed"
echo "======================================"
