# Load Testing & Load Balancing Guide 🚀

This guide will show you how to perform load testing and observe load balancing in action with your DevOps learning project.

## Table of Contents
1. [Quick Start - Docker Compose](#quick-start---docker-compose)
2. [Advanced - Kubernetes with Auto-Scaling](#advanced---kubernetes-with-auto-scaling)
3. [Understanding the Results](#understanding-the-results)

---

## Quick Start - Docker Compose

### Step 1: Verify Services are Running

```bash
docker ps --filter "name=kafka|producer|consumer|api-gateway"
```

You should see all services running and healthy.

### Step 2: Simple Load Test (Manual)

Send a single message to test the system:

```bash
curl -X POST http://localhost:8080/api/messages \
  -H "Content-Type: application/json" \
  -d '{"message":"Hello from load test!","user_id":"test-user"}'
```

### Step 3: Run Automated Load Test

The project includes a load testing script. Run it:

```bash
cd /home/reagan/Desktop/phone/devops-learning-project
chmod +x load-tests/load-test.sh
./load-tests/load-test.sh
```

**What this does:**
- Starts with 10 requests/second
- Gradually increases to 100 requests/second
- Runs for 5 minutes (300 seconds)
- Increases load every 30 seconds

### Step 4: Monitor the System

While the load test is running, open a new terminal and monitor:

**Monitor Docker containers:**
```bash
watch -n 2 'docker stats --no-stream'
```

**Monitor API Gateway metrics:**
```bash
watch -n 2 'curl -s http://localhost:8080/metrics | jq'
```

**Monitor Consumer stats:**
```bash
watch -n 2 'curl -s http://localhost:3000/api/stats | jq'
```

### Step 5: Scale Services Manually (Docker Compose)

To see load balancing in action, scale up the services:

```bash
# Scale producer to 3 instances
docker compose -f docker-compose.yml up -d --scale producer=3

# Scale consumer to 3 instances
docker compose -f docker-compose.yml up -d --scale consumer=3

# Scale api-gateway to 2 instances
docker compose -f docker-compose.yml up -d --scale api-gateway=2
```

**Note:** You'll need to remove the `container_name` from docker-compose.yml for scaling to work, or use the Kubernetes approach below for better load balancing.

---

## Advanced - Kubernetes with Auto-Scaling

For true auto-scaling and load balancing, use Kubernetes:

### Step 1: Set Up Kubernetes

If you don't have Kubernetes running, start Minikube:

```bash
# Start Minikube
minikube start

# Enable metrics server (required for auto-scaling)
minikube addons enable metrics-server
```

### Step 2: Deploy to Kubernetes

```bash
cd /home/reagan/Desktop/phone/devops-learning-project

# Apply all Kubernetes configurations
kubectl apply -f k8s/
```

### Step 3: Verify Deployment

```bash
# Check all pods
kubectl get pods

# Check services
kubectl get services

# Check HPA (Horizontal Pod Autoscaler)
kubectl get hpa
```

### Step 4: Run Load Test Against Kubernetes

First, get the API Gateway URL:

```bash
# If using Minikube
minikube service api-gateway --url
```

Then run the load test:

```bash
# Set the API URL and run the test
API_URL=$(minikube service api-gateway --url) ./load-tests/load-test.sh
```

### Step 5: Monitor Auto-Scaling in Real-Time

Open a **second terminal** and run the monitoring script:

```bash
cd /home/reagan/Desktop/phone/devops-learning-project
chmod +x monitoring/monitor-scaling.sh
./monitoring/monitor-scaling.sh
```

This will show you:
- Number of pods for each service
- CPU and memory usage
- Auto-scaling events in real-time

### Step 6: Watch Pods Scale

In a **third terminal**, watch pods being created/destroyed:

```bash
watch -n 1 'kubectl get pods'
```

You'll see new pods being created as load increases!

---

## Understanding the Results

### What to Look For

#### 1. **Load Balancing**
- Requests are distributed across multiple pod instances
- Each pod handles a portion of the traffic
- No single pod is overwhelmed

#### 2. **Auto-Scaling Triggers**
- When CPU usage exceeds 50%, new pods are created
- When load decreases, pods are terminated
- Scaling happens gradually (not instantly)

#### 3. **Kafka Message Distribution**
- Messages are distributed across Kafka partitions
- Multiple consumers can process messages in parallel
- Consumer group ensures each message is processed once

### Key Metrics to Monitor

| Metric | What it Shows | Where to Find |
|--------|---------------|---------------|
| **Pod Count** | Number of running instances | `kubectl get pods` |
| **CPU Usage** | Processing load per pod | `kubectl top pods` |
| **Memory Usage** | Memory consumption | `kubectl top pods` |
| **Request Count** | Total API requests | `curl localhost:8080/metrics` |
| **Message Throughput** | Messages processed/sec | `curl localhost:3000/api/stats` |

### Expected Behavior

1. **Initial State**: 1 pod per service
2. **Load Increases**: CPU usage rises
3. **Scaling Up**: HPA creates new pods (takes 30-60 seconds)
4. **Load Distributed**: Requests spread across pods
5. **Load Decreases**: After 5 minutes of low load, pods are removed
6. **Back to Normal**: Returns to minimum pod count

---

## Customizing the Load Test

### Change Test Duration

```bash
DURATION=600 ./load-tests/load-test.sh  # Run for 10 minutes
```

### Change Request Rate

```bash
INITIAL_RPS=20 MAX_RPS=200 ./load-tests/load-test.sh
```

### Test Specific Endpoints

Edit `load-tests/load-test.sh` and change the curl command to test different endpoints.

---

## Troubleshooting

### Load Test Fails

**Problem:** `ERROR: API is not reachable`

**Solution:**
```bash
# Check if services are running
docker ps

# Check API Gateway health
curl http://localhost:8080/health
```

### Auto-Scaling Not Working

**Problem:** Pods don't scale up

**Solution:**
```bash
# Check if metrics-server is running
kubectl get deployment metrics-server -n kube-system

# Check HPA status
kubectl describe hpa producer-hpa

# Verify resource requests are set
kubectl describe deployment producer
```

### Pods Stuck in Pending

**Problem:** New pods don't start

**Solution:**
```bash
# Check node resources
kubectl top nodes

# Describe the pending pod
kubectl describe pod <pod-name>
```

---

## Advanced Scenarios

### Stress Test (Maximum Load)

```bash
# Send 500 requests/second for 2 minutes
INITIAL_RPS=500 MAX_RPS=500 DURATION=120 ./load-tests/load-test.sh
```

### Spike Test (Sudden Load)

```bash
# Send 1000 requests immediately
for i in {1..1000}; do
  curl -X POST http://localhost:8080/api/messages \
    -H "Content-Type: application/json" \
    -d "{\"message\":\"Spike test $i\",\"user_id\":\"spike-test\"}" &
done
wait
```

### Endurance Test (Long Duration)

```bash
# Run for 1 hour with moderate load
DURATION=3600 INITIAL_RPS=30 MAX_RPS=50 ./load-tests/load-test.sh
```

---

## Clean Up

### Stop Docker Compose Services

```bash
docker compose -f docker-compose.yml down
```

### Delete Kubernetes Resources

```bash
kubectl delete -f k8s/
```

### Stop Minikube

```bash
minikube stop
```

---

## Next Steps

1. ✅ Run the basic load test
2. ✅ Monitor system metrics
3. ✅ Deploy to Kubernetes
4. ✅ Observe auto-scaling
5. 📊 Analyze the results
6. 🔧 Tune HPA settings for optimal performance
7. 📈 Set up Prometheus & Grafana for advanced monitoring

---

## Additional Resources

- **Kafka Documentation**: Understanding message distribution
- **Kubernetes HPA**: How auto-scaling decisions are made
- **Docker Compose Scaling**: Limitations and best practices
- **Load Testing Tools**: Apache JMeter, k6, Locust for more advanced testing

Happy Load Testing! 🎯
