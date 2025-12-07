#!/bin/bash

# QUICK DEMO - Load Testing & Monitoring
# =======================================
# This script demonstrates load testing in action
# Run this to see your system handle load!

set -e

echo "🚀 DevOps Load Testing Demo"
echo "============================"
echo ""

# Check if services are running
echo "📊 Checking system health..."
if ! curl -sf http://localhost:8080/health > /dev/null; then
    echo "❌ ERROR: Services are not running!"
    echo "   Start them with: docker compose up -d"
    exit 1
fi
echo "✅ All services are healthy!"
echo ""

# Show initial metrics
echo "📈 Initial Metrics:"
echo "-------------------"
curl -s http://localhost:8080/metrics | jq
echo ""

# Send some test messages
echo "📤 Sending 50 test messages..."
for i in {1..50}; do
    curl -s -X POST http://localhost:8080/api/messages \
        -H "Content-Type: application/json" \
        -d "{\"message\":\"Demo message $i\",\"user_id\":\"demo-user\"}" > /dev/null
    
    # Show progress
    if (( i % 10 == 0 )); then
        echo "   Sent $i messages..."
    fi
done
echo "✅ All messages sent!"
echo ""

# Wait a moment for processing
echo "⏳ Waiting for messages to be processed..."
sleep 3
echo ""

# Show updated metrics
echo "📊 Updated Metrics:"
echo "-------------------"
echo "API Gateway:"
curl -s http://localhost:8080/metrics | jq
echo ""
echo "Consumer Stats:"
curl -s http://localhost:3000/api/stats | jq
echo ""

# Show Docker stats
echo "🐳 Container Resource Usage:"
echo "----------------------------"
docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}" \
    api-gateway producer consumer kafka zookeeper
echo ""

echo "✅ Demo Complete!"
echo ""
echo "📚 Next Steps:"
echo "   1. Run full load test: ./load-tests/load-test.sh"
echo "   2. Monitor in real-time: watch -n 2 'curl -s http://localhost:8080/metrics | jq'"
echo "   3. Check consumer stats: curl http://localhost:3000/api/stats | jq"
echo "   4. View Docker stats: docker stats"
echo ""
echo "🎯 For Kubernetes auto-scaling, see: LOAD_TESTING_GUIDE.md"
