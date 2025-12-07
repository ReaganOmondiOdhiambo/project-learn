# 🔌 Kafka Networking: Why Two Ports?

## 🎯 The Question

**Why does Kafka use port `29092` in the services but port `9092` is also configured?**

Looking at `docker-compose.yml`:

```yaml
kafka:
  ports:
    - "9092:9092"
  environment:
    KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092,PLAINTEXT_INTERNAL://kafka:29092

producer:
  environment:
    KAFKA_BROKER: kafka:29092    # ← Why 29092 here?

consumer:
  environment:
    KAFKA_BROKER: kafka:29092    # ← Why 29092 here too?
```

---

## 📚 The Answer: Two Different Networks

Kafka is accessible from **two different places**, so it needs **two different addresses**:

### 1. **Internal Network** (Container-to-Container)
- **Port:** `29092`
- **Hostname:** `kafka`
- **Who uses it:** Producer, Consumer, other Docker containers
- **Network:** Docker internal network

### 2. **External Network** (Host-to-Container)
- **Port:** `9092`
- **Hostname:** `localhost`
- **Who uses it:** Your laptop, CLI tools, external applications
- **Network:** Your computer's network

---

## 🏗️ Architecture Diagram

```
┌─────────────────────────────────────────────────────────┐
│  YOUR LAPTOP (Host Machine)                             │
│                                                          │
│  Terminal Commands:                                     │
│  $ kafka-console-consumer \                             │
│      --bootstrap-server localhost:9092 ───┐             │
│                                            │             │
└────────────────────────────────────────────┼─────────────┘
                                             │
                                    Port 9092 (External)
                                             │
                                             ↓
┌─────────────────────────────────────────────────────────┐
│  DOCKER NETWORK (devops-learning-network)               │
│                                                          │
│  ┌─────────────────┐                                    │
│  │  Producer       │                                    │
│  │  Container      │                                    │
│  │                 │                                    │
│  │  KAFKA_BROKER=  │                                    │
│  │  kafka:29092 ───┼──────┐                             │
│  └─────────────────┘      │                             │
│                           │ Port 29092                  │
│  ┌─────────────────┐      │ (Internal)                  │
│  │  Consumer       │      │                             │
│  │  Container      │      │                             │
│  │                 │      │                             │
│  │  KAFKA_BROKER=  │      │                             │
│  │  kafka:29092 ───┼──────┤                             │
│  └─────────────────┘      │                             │
│                           │                             │
│                           ↓                             │
│                  ┌─────────────────┐                    │
│                  │  Kafka Broker   │                    │
│                  │                 │                    │
│                  │  Listens on:    │                    │
│                  │  - 29092 (int)  │←─ Containers       │
│                  │  - 9092 (ext)   │←─ Host machine     │
│                  └─────────────────┘                    │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

---

## 🔍 Detailed Explanation

### Kafka Configuration in `docker-compose.yml`

```yaml
kafka:
  ports:
    - "9092:9092"    # Maps host port 9092 to container port 9092
  
  environment:
    # Define two listeners
    KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: |
      PLAINTEXT:PLAINTEXT,
      PLAINTEXT_INTERNAL:PLAINTEXT
    
    # Advertise two addresses
    KAFKA_ADVERTISED_LISTENERS: |
      PLAINTEXT://localhost:9092,
      PLAINTEXT_INTERNAL://kafka:29092
    #   ↑                           ↑
    #   External address      Internal address
    
    # Brokers talk to each other using internal listener
    KAFKA_INTER_BROKER_LISTENER_NAME: PLAINTEXT_INTERNAL
```

### Breaking It Down:

#### 1. **KAFKA_LISTENER_SECURITY_PROTOCOL_MAP**
Defines two listener names:
- `PLAINTEXT` - for external connections
- `PLAINTEXT_INTERNAL` - for internal connections

#### 2. **KAFKA_ADVERTISED_LISTENERS**
Tells clients how to connect:
- `PLAINTEXT://localhost:9092` - "If you're on the host machine, connect here"
- `PLAINTEXT_INTERNAL://kafka:29092` - "If you're a container, connect here"

#### 3. **KAFKA_INTER_BROKER_LISTENER_NAME**
Kafka itself uses the internal listener for communication

---

## 🎯 When to Use Which Port

### Use `kafka:29092` (Internal Port)

**In docker-compose.yml service configurations:**

```yaml
producer:
  environment:
    KAFKA_BROKER: kafka:29092    # ✅ Correct

consumer:
  environment:
    KAFKA_BROKER: kafka:29092    # ✅ Correct
```

**Why?**
- Producer and Consumer are **containers**
- They're on the **same Docker network** as Kafka
- They can use the service name `kafka` as hostname
- Port `29092` is the internal listener

---

### Use `localhost:9092` (External Port)

**From your laptop terminal:**

```bash
# List topics
kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic messages \
  --from-beginning

# Create topic
kafka-topics \
  --create \
  --bootstrap-server localhost:9092 \
  --topic test

# Test with kafkacat
kafkacat -b localhost:9092 -L
```

**Why?**
- Your laptop is **outside** the Docker network
- Can't resolve hostname `kafka`
- Must use `localhost` with mapped port `9092`

---

## 🧪 Testing Both Connections

### Test Internal Connection (29092)

```bash
# From inside producer container
docker exec -it producer python3 << EOF
from kafka import KafkaProducer
producer = KafkaProducer(bootstrap_servers=['kafka:29092'])
print("✅ Connected to kafka:29092")
EOF
```

### Test External Connection (9092)

```bash
# From your laptop
docker exec -it kafka kafka-broker-api-versions \
  --bootstrap-server localhost:9092

# Should show Kafka version info
```

---

## ❌ Common Mistakes

### Mistake 1: Using localhost in containers

```yaml
# ❌ WRONG
producer:
  environment:
    KAFKA_BROKER: localhost:9092    # Won't work!
```

**Why it fails:**
- Inside a container, `localhost` refers to **that container**
- Not the Kafka container
- Connection will fail

**Fix:**
```yaml
# ✅ CORRECT
producer:
  environment:
    KAFKA_BROKER: kafka:29092
```

---

### Mistake 2: Using kafka:29092 from host

```bash
# ❌ WRONG (from your laptop)
kafka-console-consumer --bootstrap-server kafka:29092
```

**Why it fails:**
- Your laptop doesn't know what `kafka` is
- It's not on the Docker network
- DNS resolution fails

**Fix:**
```bash
# ✅ CORRECT
kafka-console-consumer --bootstrap-server localhost:9092
```

---

## 🔄 The Complete Flow

### Sending a Message (Internal)

```
1. Client sends HTTP to API Gateway
   curl http://localhost:8080/api/messages

2. API Gateway forwards to Producer
   http://producer:5000/api/messages

3. Producer connects to Kafka
   kafka:29092 (internal network)

4. Kafka stores message

5. Consumer reads from Kafka
   kafka:29092 (internal network)
```

**All container-to-container uses `kafka:29092`**

---

### Debugging from Laptop (External)

```
1. You run kafka-console-consumer
   kafka-console-consumer --bootstrap-server localhost:9092

2. Docker maps localhost:9092 → kafka:9092

3. Kafka receives connection on external listener

4. You see messages!
```

**Host-to-container uses `localhost:9092`**

---

## 📊 Port Mapping Explained

```yaml
ports:
  - "9092:9092"
#    ↑     ↑
#    │     └─ Container port (inside Docker)
#    └─────── Host port (on your laptop)
```

**What this means:**
- Traffic to `localhost:9092` on your laptop
- Gets forwarded to port `9092` inside the Kafka container
- But containers use port `29092` for internal communication

---

## 🌐 Network Isolation

### Docker Network View

```
Docker Network: devops-learning-network
├── kafka (kafka:29092)
├── producer (producer:5000)
├── consumer (consumer:3000)
└── api-gateway (api-gateway:8080)

All can reach each other by service name!
```

### Host Network View

```
Your Laptop
├── localhost:8080 → api-gateway:8080
├── localhost:5000 → producer:5000
├── localhost:3000 → consumer:3000
└── localhost:9092 → kafka:9092

Must use localhost with mapped ports!
```

---

## 🎓 Key Takeaways

1. **Two networks = Two addresses**
   - Internal: `kafka:29092` (containers)
   - External: `localhost:9092` (host)

2. **Containers use service names**
   - `kafka`, `producer`, `consumer`
   - Internal ports: `29092`, `5000`, `3000`

3. **Host uses localhost**
   - `localhost` with mapped ports
   - External port: `9092`

4. **Port mapping is one-way**
   - Host → Container ✅
   - Container → Host ❌ (not needed)

5. **Kafka is smart**
   - Advertises different addresses to different clients
   - Internal clients get `kafka:29092`
   - External clients get `localhost:9092`

---

## 📝 Quick Reference

| Connecting From | Use This | Example |
|----------------|----------|---------|
| Producer container | `kafka:29092` | `KAFKA_BROKER=kafka:29092` |
| Consumer container | `kafka:29092` | `KAFKA_BROKER=kafka:29092` |
| Your laptop terminal | `localhost:9092` | `--bootstrap-server localhost:9092` |
| Kubernetes pod | `kafka-service:9092` | Different setup entirely |

---

## 🔗 Related Files

- [docker-compose.yml](file:///home/reagan/Desktop/phone/devops-learning-project/docker-compose.yml) - See Kafka configuration (lines 30-70)
- [producer/app.py](file:///home/reagan/Desktop/phone/devops-learning-project/producer/app.py) - Uses `kafka:29092`
- [consumer/app.js](file:///home/reagan/Desktop/phone/devops-learning-project/consumer/app.js) - Uses `kafka:29092`

---

**Now you understand why Kafka has two ports and when to use each one!** 🎉
