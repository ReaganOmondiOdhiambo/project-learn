# 🚀 DevOps Learning Project

A comprehensive hands-on project to learn **Docker**, **Kafka**, **Kubernetes**, **CI/CD**, and **Auto-scaling** through a real microservices application.

## 📚 What You'll Learn

### 1. **Docker Fundamentals**
- Writing Dockerfiles
- Multi-stage builds (optimization)
- Docker Compose orchestration
- Container networking
- Health checks
- Resource limits

### 2. **Kafka Event Streaming**
- Producer/Consumer pattern
- Message brokers
- Event-driven architecture
- Topic management
- Consumer groups

### 3. **Kubernetes (K8s)**
- Deployments
- Services (ClusterIP, LoadBalancer)
- StatefulSets (for Kafka/Zookeeper)
- ConfigMaps and Secrets
- Horizontal Pod Autoscaler (HPA)
- Resource requests and limits

### 4. **CI/CD Pipelines**
- GitHub Actions workflows
- Automated testing
- Docker image building
- Container registry
- Automated deployment
- Security scanning

### 5. **Auto-scaling**
- Metrics-based scaling
- CPU and memory triggers
- Load testing
- Monitoring scaling behavior

---

## 🏗️ Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ▼
┌─────────────────┐
│  API Gateway    │  (Go - Port 8080)
│  LoadBalancer   │
└────┬────────┬───┘
     │        │
     ▼        ▼
┌─────────┐ ┌──────────┐
│Producer │ │ Consumer │
│ Service │ │ Service  │
│(Python) │ │ (Node.js)│
│Port 5000│ │ Port 3000│
└────┬────┘ └────┬─────┘
     │           │
     │    ┌──────▼──────┐
     └───►│    Kafka    │
          │  (Message   │
          │   Broker)   │
          └─────────────┘
                │
          ┌─────▼──────┐
          │ Zookeeper  │
          └────────────┘
```

### Services:

1. **API Gateway (Go)**: Single entry point, routes requests, load balancing
2. **Producer (Python/Flask)**: Receives HTTP requests, publishes to Kafka
3. **Consumer (Node.js)**: Consumes Kafka messages, provides WebSocket updates
4. **Kafka**: Message broker for event streaming
5. **Zookeeper**: Kafka cluster coordination

---

## 🚦 Quick Start

### Prerequisites

- **Docker** and **Docker Compose** installed
- **kubectl** (for Kubernetes)
- **Minikube** or **Kind** (for local Kubernetes cluster)
- **Git**

### Option 1: Docker Compose (Easiest)

```bash
# Clone or navigate to the project
cd devops-learning-project

# Start all services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f

# Test the API
curl http://localhost:8080/health

# Send a message
curl -X POST http://localhost:8080/api/messages \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello DevOps!", "user_id": "test"}'

# Stop all services
docker-compose down
```

### Option 2: Kubernetes (For Auto-scaling)

```bash
# Start Minikube
minikube start --cpus=4 --memory=8192

# Enable metrics server (required for HPA)
minikube addons enable metrics-server

# Build Docker images
docker-compose build

# Load images into Minikube
minikube image load producer:latest
minikube image load consumer:latest
minikube image load api-gateway:latest

# Deploy to Kubernetes
kubectl apply -f k8s/

# Check deployments
kubectl get deployments
kubectl get pods
kubectl get services
kubectl get hpa

# Get API Gateway URL
minikube service api-gateway-service --url

# Test the API (use the URL from above)
curl http://<minikube-ip>:<port>/health
```

---

## 🧪 Testing Auto-Scaling

### Method 1: Simple Load Test

```bash
# Make load test script executable
chmod +x load-tests/load-test.sh

# Run load test (generates increasing load)
./load-tests/load-test.sh

# In another terminal, monitor scaling
watch kubectl get hpa
watch kubectl get pods
```

### Method 2: Advanced Load Test with Locust

```bash
# Install Locust
pip install locust

# Run Locust
cd load-tests
locust -f locustfile.py --host=http://localhost:8080

# Open browser to http://localhost:8089
# Set users: 100, spawn rate: 10
# Watch the real-time metrics!
```

### Method 3: Monitor Scaling Behavior

```bash
# Make monitoring script executable
chmod +x monitoring/monitor-scaling.sh

# Start monitoring (in one terminal)
./monitoring/monitor-scaling.sh

# Start load test (in another terminal)
./load-tests/load-test.sh

# Watch the pods scale up and down!
```

---

## 📊 Observing Auto-Scaling

### What to Watch:

1. **HPA Status**:
   ```bash
   kubectl get hpa
   ```
   Shows current CPU/memory usage and replica count

2. **Pod Scaling**:
   ```bash
   kubectl get pods -w
   ```
   Watch pods being created/terminated in real-time

3. **Resource Usage**:
   ```bash
   kubectl top pods
   kubectl top nodes
   ```
   See actual CPU and memory consumption

4. **HPA Events**:
   ```bash
   kubectl describe hpa producer-hpa
   ```
   See scaling decisions and events

### Expected Behavior:

- **Low load**: 2 replicas (minimum)
- **Medium load (70% CPU)**: Scales up to 4-6 replicas
- **High load**: Scales up to 10 replicas (maximum)
- **Load decreases**: Gradually scales down after 60s

---

## 🔧 API Endpoints

### API Gateway (Port 8080)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/` | API documentation |
| GET | `/health` | Health check (all services) |
| GET | `/metrics` | API usage metrics |
| POST | `/api/messages` | Send single message |
| POST | `/api/messages/batch` | Send batch messages |
| GET | `/api/stats` | Consumer statistics |

### Examples:

```bash
# Send a message
curl -X POST http://localhost:8080/api/messages \
  -H "Content-Type: application/json" \
  -d '{"message": "Test message", "user_id": "user123"}'

# Send batch messages
curl -X POST http://localhost:8080/api/messages/batch \
  -H "Content-Type: application/json" \
  -d '{"messages": ["msg1", "msg2", "msg3"], "user_id": "user123"}'

# Get consumer stats
curl http://localhost:8080/api/stats

# Check health
curl http://localhost:8080/health

# Get metrics
curl http://localhost:8080/metrics
```

---

## 🔄 CI/CD Pipeline

The project includes a GitHub Actions workflow (`.github/workflows/ci-cd.yml`) that:

1. **Tests** all services (Python, Node.js, Go)
2. **Builds** Docker images
3. **Pushes** to Docker Hub
4. **Deploys** to Kubernetes
5. **Scans** for security vulnerabilities

### Setup:

1. Fork this repository
2. Add GitHub Secrets:
   - `DOCKER_USERNAME`: Your Docker Hub username
   - `DOCKER_PASSWORD`: Your Docker Hub password/token
   - `KUBECONFIG`: Your Kubernetes config (for deployment)

3. Push to `main` branch - pipeline runs automatically!

---

## 📖 DevOps Concepts Explained

### Docker

**What**: Containerization platform that packages apps with dependencies

**Why**: 
- Consistent environments (dev = prod)
- Isolation
- Portability
- Efficient resource usage

**Key Files**:
- `Dockerfile`: Recipe to build an image
- `docker-compose.yml`: Multi-container orchestration

### Kafka

**What**: Distributed event streaming platform

**Why**:
- Decouples services
- Handles high throughput
- Durable message storage
- Scalable

**Key Concepts**:
- **Producer**: Publishes messages
- **Consumer**: Reads messages
- **Topic**: Category of messages
- **Partition**: Parallel processing

### Kubernetes

**What**: Container orchestration platform

**Why**:
- Auto-scaling
- Self-healing
- Load balancing
- Rolling updates
- Service discovery

**Key Resources**:
- **Deployment**: Manages pod replicas
- **Service**: Network endpoint
- **HPA**: Auto-scales based on metrics
- **StatefulSet**: For stateful apps (databases)

### CI/CD

**What**: Continuous Integration / Continuous Deployment

**Why**:
- Automate testing
- Fast feedback
- Consistent deployments
- Reduce human error

**Pipeline Stages**:
1. Code commit
2. Automated tests
3. Build artifacts
4. Deploy to staging
5. Deploy to production

---

## 🛠️ Troubleshooting

### Docker Compose Issues

**Problem**: Services won't start
```bash
# Check logs
docker-compose logs

# Rebuild images
docker-compose build --no-cache
docker-compose up -d
```

**Problem**: Kafka connection errors
```bash
# Wait for Kafka to be ready (takes 30-60s)
docker-compose logs kafka

# Restart services
docker-compose restart producer consumer
```

### Kubernetes Issues

**Problem**: Pods not starting
```bash
# Check pod status
kubectl get pods
kubectl describe pod <pod-name>

# Check logs
kubectl logs <pod-name>
```

**Problem**: HPA not scaling
```bash
# Ensure metrics-server is running
kubectl get deployment metrics-server -n kube-system

# Check HPA status
kubectl describe hpa producer-hpa

# Verify resource requests are set
kubectl describe deployment producer
```

**Problem**: Images not found
```bash
# Load images into Minikube
eval $(minikube docker-env)
docker-compose build
```

---

## 📚 Learning Resources

### Docker
- [Official Docker Docs](https://docs.docker.com/)
- [Docker Compose Docs](https://docs.docker.com/compose/)

### Kafka
- [Kafka Quickstart](https://kafka.apache.org/quickstart)
- [Kafka in 100 Seconds](https://www.youtube.com/watch?v=uvb00oaa3k8)

### Kubernetes
- [Kubernetes Basics](https://kubernetes.io/docs/tutorials/kubernetes-basics/)
- [K8s HPA Walkthrough](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale-walkthrough/)

### CI/CD
- [GitHub Actions Docs](https://docs.github.com/en/actions)

---

## 🎯 Next Steps

1. **Experiment**: Modify the code, add features
2. **Monitor**: Add Prometheus + Grafana for metrics
3. **Secure**: Add authentication, HTTPS, secrets management
4. **Scale**: Deploy to cloud (AWS EKS, GCP GKE, Azure AKS)
5. **Optimize**: Add caching, database, message persistence

---

## 📝 Project Structure

```
devops-learning-project/
├── producer/                 # Python Flask producer service
│   ├── app.py
│   ├── requirements.txt
│   └── Dockerfile
├── consumer/                 # Node.js consumer service
│   ├── app.js
│   ├── package.json
│   └── Dockerfile
├── api-gateway/             # Go API gateway
│   ├── main.go
│   ├── go.mod
│   └── Dockerfile
├── k8s/                     # Kubernetes manifests
│   ├── producer-deployment.yaml
│   ├── consumer-deployment.yaml
│   ├── api-gateway-deployment.yaml
│   └── kafka-deployment.yaml
├── .github/workflows/       # CI/CD pipeline
│   └── ci-cd.yml
├── load-tests/              # Load testing scripts
│   ├── load-test.sh
│   └── locustfile.py
├── monitoring/              # Monitoring scripts
│   └── monitor-scaling.sh
├── docker-compose.yml       # Local development
├── .env.example            # Environment variables
├── .gitignore
└── README.md               # This file
```

---

## 🤝 Contributing

Feel free to:
- Add more services
- Improve documentation
- Add tests
- Optimize Dockerfiles
- Enhance monitoring

---

## 📄 License

MIT License - Feel free to use for learning!

---

## 🎓 Key Takeaways

After completing this project, you'll understand:

✅ How to containerize applications with Docker  
✅ How to orchestrate multi-container apps with Docker Compose  
✅ How Kafka enables event-driven architectures  
✅ How Kubernetes manages containerized workloads  
✅ How auto-scaling works based on metrics  
✅ How to build CI/CD pipelines  
✅ How to load test and monitor applications  

**Happy Learning! 🚀**
