# Database Sharding & Scaling Project

A deep-dive into **Database Sharding** and **Horizontal Scaling** using Kubernetes and Go.

## 🧠 The Core Concept: Consistent Hashing

Most databases scale vertically (bigger machine). To scale horizontally (more machines), you need **Sharding**.
But how do you know which machine holds `user_123`'s data?

We built a **Custom Sharding Proxy** in Go that uses **Consistent Hashing**:
1.  We map both **Shards** (DB Nodes) and **Data Keys** onto a "Ring" (0 to 2^32).
2.  A key is stored on the first Shard found moving clockwise on the Ring.
3.  **Benefit**: When adding a new Shard, only `1/N` keys need to move.

## 🏗 Architecture

-   **Sharding Proxy** (Go): The "Router". Receives requests, hashes the key, and forwards to the correct DB.
-   **PostgreSQL Shards** (K8s StatefulSet): 3 independent DB instances (`postgres-shard-0`, `1`, `2`).
-   **Kubernetes**: Orchestrates the deployment.

## 🚀 Quick Start

### Prerequisites
-   Minikube or any K8s cluster
-   Docker
-   Go 1.21+ (optional, for local dev)

### Deploy
```bash
./scripts/deploy.sh
```

### Usage

**1. Write Data**
```bash
curl -X POST http://localhost:80/write \
  -d '{"key":"user_1","value":"Alice"}'
```
Response: `{"message":"Write successful","shard":"postgres-shard-0.postgres-headless"}`

**2. Read Data**
```bash
curl "http://localhost:80/read?key=user_1"
```

**3. View Distribution Stats**
```bash
curl http://localhost:80/stats
```

## 📈 Scaling Demo

To see sharding in action:

1.  **Generate Load**: Run the load generator (or manually write 100 keys).
2.  **Scale Out**:
    ```bash
    ./scripts/demo-scaling.sh
    ```
    This script will:
    -   Scale the StatefulSet to 4 replicas.
    -   Update the Proxy configuration to include `postgres-shard-3`.
    -   Show how new keys start landing on the new shard.

## 📂 Project Structure

-   `sharding-proxy/`: Go application source code.
    -   `hashring/`: The Consistent Hashing algorithm.
-   `k8s/`: Kubernetes manifests.
-   `scripts/`: Helper scripts.
