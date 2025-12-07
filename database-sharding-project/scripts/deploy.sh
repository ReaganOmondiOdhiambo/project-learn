#!/bin/bash

# Ensure we are in the project root
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

echo "📂 Working directory: $(pwd)"

# Build the proxy image locally so Minikube/K8s can use it
echo "🔨 Building Sharding Proxy Docker Image..."
# Use minikube image build to avoid permission issues with docker-env
minikube image build -t sharding-proxy:latest ./sharding-proxy

echo "🚀 Deploying to Kubernetes..."
kubectl apply -f k8s/postgres-shards.yaml
kubectl apply -f k8s/proxy-deployment.yaml

echo "⏳ Waiting for pods to be ready..."
# Wait for postgres first as proxy depends on it
kubectl wait --for=condition=ready pod -l app=postgres-shard --timeout=120s
kubectl wait --for=condition=ready pod -l app=sharding-proxy --timeout=120s

echo "✅ Deployment Complete!"
echo "Proxy is available at: http://localhost:80"
