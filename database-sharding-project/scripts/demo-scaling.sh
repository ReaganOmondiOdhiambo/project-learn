#!/bin/bash

# SCALING DEMO SCRIPT
# ===================

echo "📈 Starting Scaling Demo..."

# 1. Current State
echo "1️⃣  Current Status:"
kubectl get pods -l app=postgres-shard
echo ""

# 2. Scale Up
echo "2️⃣  Scaling Database Layer to 4 Shards..."
kubectl scale statefulset postgres-shard --replicas=4
echo "   Waiting for postgres-shard-3 to be ready..."
kubectl wait --for=condition=ready pod postgres-shard-3 --timeout=120s
echo "✅ Shard 3 is Ready!"
echo ""

# 3. Update Proxy Config
echo "3️⃣  Updating Proxy Configuration..."
# In a real production system, this would be a ConfigMap update or dynamic service discovery.
# For this demo, we'll patch the deployment env var.

NEW_SHARDS="postgres-shard-0.postgres-headless,postgres-shard-1.postgres-headless,postgres-shard-2.postgres-headless,postgres-shard-3.postgres-headless"

kubectl set env deployment/sharding-proxy SHARDS=$NEW_SHARDS

echo "   Rolling out updated proxy..."
kubectl rollout status deployment/sharding-proxy
echo "✅ Proxy Updated!"
echo ""

# 4. Verification
echo "4️⃣  Verification:"
echo "   Now try writing new keys. Some should land on postgres-shard-3."
echo "   Run: curl -X POST http://localhost:80/write -d '{\"key\":\"test_new_shard\",\"value\":\"data\"}'"
