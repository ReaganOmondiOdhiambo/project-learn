#!/bin/bash

# LOAD TESTING SCRIPT
# ===================
# This script generates load to test auto-scaling behavior
# It sends increasing numbers of requests to trigger HPA scaling

set -e

# Configuration
API_URL="${API_URL:-http://localhost:8080}"
DURATION="${DURATION:-300}"  # Test duration in seconds (5 minutes)
INITIAL_RPS="${INITIAL_RPS:-10}"  # Initial requests per second
MAX_RPS="${MAX_RPS:-100}"  # Maximum requests per second

echo "========================================="
echo "Load Testing Configuration"
echo "========================================="
echo "API URL: $API_URL"
echo "Duration: ${DURATION}s"
echo "Initial RPS: $INITIAL_RPS"
echo "Max RPS: $MAX_RPS"
echo "========================================="
echo ""

# Check if API is reachable
echo "Checking API health..."
if ! curl -sf "$API_URL/health" > /dev/null; then
    echo "ERROR: API is not reachable at $API_URL"
    exit 1
fi
echo "✓ API is healthy"
echo ""

# Function to send batch requests
send_requests() {
    local rps=$1
    local duration=$2
    local total_requests=$((rps * duration))
    
    echo "Sending $rps requests/second for ${duration}s (total: $total_requests requests)..."
    
    # Use GNU parallel if available, otherwise use simple loop
    if command -v parallel &> /dev/null; then
        seq 1 $total_requests | parallel -j $rps --delay 1 \
            "curl -s -X POST $API_URL/api/messages \
            -H 'Content-Type: application/json' \
            -d '{\"message\":\"Load test message {}\",\"user_id\":\"load-test\"}' \
            > /dev/null"
    else
        for i in $(seq 1 $total_requests); do
            curl -s -X POST "$API_URL/api/messages" \
                -H "Content-Type: application/json" \
                -d "{\"message\":\"Load test message $i\",\"user_id\":\"load-test\"}" \
                > /dev/null &
            
            # Limit concurrent requests
            if (( i % rps == 0 )); then
                wait
                sleep 1
            fi
        done
        wait
    fi
}

# Gradual load increase
echo "Starting gradual load test..."
echo "This will gradually increase load to trigger auto-scaling"
echo ""

START_TIME=$(date +%s)
CURRENT_TIME=$START_TIME
ELAPSED=0

# Increase load gradually
CURRENT_RPS=$INITIAL_RPS
STEP_DURATION=30  # Increase load every 30 seconds

while [ $ELAPSED -lt $DURATION ]; do
    echo "----------------------------------------"
    echo "Elapsed: ${ELAPSED}s / ${DURATION}s"
    echo "Current load: $CURRENT_RPS RPS"
    echo "----------------------------------------"
    
    send_requests $CURRENT_RPS $STEP_DURATION
    
    # Increase RPS
    CURRENT_RPS=$((CURRENT_RPS + 10))
    if [ $CURRENT_RPS -gt $MAX_RPS ]; then
        CURRENT_RPS=$MAX_RPS
    fi
    
    CURRENT_TIME=$(date +%s)
    ELAPSED=$((CURRENT_TIME - START_TIME))
done

echo ""
echo "========================================="
echo "Load test completed!"
echo "========================================="
echo ""
echo "Check Kubernetes HPA status with:"
echo "  kubectl get hpa"
echo "  kubectl get pods"
echo ""
echo "Check metrics with:"
echo "  kubectl top pods"
echo "  kubectl top nodes"
