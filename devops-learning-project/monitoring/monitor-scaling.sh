#!/bin/bash

# KUBERNETES MONITORING SCRIPT
# =============================
# This script monitors Kubernetes resources during load testing
# to observe auto-scaling in action

set -e

INTERVAL="${INTERVAL:-5}"  # Check every 5 seconds
DURATION="${DURATION:-300}"  # Monitor for 5 minutes

echo "========================================="
echo "Kubernetes Auto-Scaling Monitor"
echo "========================================="
echo "Monitoring interval: ${INTERVAL}s"
echo "Duration: ${DURATION}s"
echo "========================================="
echo ""

START_TIME=$(date +%s)
CURRENT_TIME=$START_TIME
ELAPSED=0

# Create log file
LOG_FILE="scaling-monitor-$(date +%Y%m%d-%H%M%S).log"
echo "Logging to: $LOG_FILE"
echo ""

# Header
printf "%-10s %-15s %-15s %-15s %-10s %-10s\n" \
    "TIME" "PRODUCER_PODS" "CONSUMER_PODS" "GATEWAY_PODS" "CPU_AVG" "MEM_AVG" | tee -a "$LOG_FILE"
echo "--------------------------------------------------------------------------------" | tee -a "$LOG_FILE"

while [ $ELAPSED -lt $DURATION ]; do
    TIMESTAMP=$(date +%H:%M:%S)
    
    # Get pod counts
    PRODUCER_PODS=$(kubectl get pods -l app=producer --no-headers 2>/dev/null | grep -c Running || echo "0")
    CONSUMER_PODS=$(kubectl get pods -l app=consumer --no-headers 2>/dev/null | grep -c Running || echo "0")
    GATEWAY_PODS=$(kubectl get pods -l app=api-gateway --no-headers 2>/dev/null | grep -c Running || echo "0")
    
    # Get average CPU and memory (requires metrics-server)
    CPU_AVG=$(kubectl top pods --no-headers 2>/dev/null | awk '{sum+=$2} END {print sum/NR}' || echo "N/A")
    MEM_AVG=$(kubectl top pods --no-headers 2>/dev/null | awk '{sum+=$3} END {print sum/NR}' || echo "N/A")
    
    # Print status
    printf "%-10s %-15s %-15s %-15s %-10s %-10s\n" \
        "$TIMESTAMP" "$PRODUCER_PODS" "$CONSUMER_PODS" "$GATEWAY_PODS" "$CPU_AVG" "$MEM_AVG" | tee -a "$LOG_FILE"
    
    sleep $INTERVAL
    
    CURRENT_TIME=$(date +%s)
    ELAPSED=$((CURRENT_TIME - START_TIME))
done

echo ""
echo "========================================="
echo "Monitoring completed!"
echo "========================================="
echo ""
echo "Summary:"
kubectl get hpa
echo ""
echo "Final pod status:"
kubectl get pods
echo ""
echo "Log saved to: $LOG_FILE"
