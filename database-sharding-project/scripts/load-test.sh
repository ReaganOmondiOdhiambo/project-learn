#!/bin/bash

# LOAD TEST SCRIPT
# ================
# Generates random keys and writes them to the proxy to demonstrate distribution.

PROXY_URL="http://localhost:8081"
COUNT=100

echo "🚀 Starting Load Test ($COUNT keys)..."

for i in $(seq 1 $COUNT); do
    KEY="user_$RANDOM"
    VALUE="data_for_$KEY"
    
    # Write
    curl -s -X POST "$PROXY_URL/write" \
        -d "{\"key\":\"$KEY\",\"value\":\"$VALUE\"}" > /dev/null
    
    if [ $((i % 10)) -eq 0 ]; then
        echo -n "."
    fi
done

echo ""
echo "✅ Load Test Complete!"
echo "Check distribution: curl $PROXY_URL/stats"
