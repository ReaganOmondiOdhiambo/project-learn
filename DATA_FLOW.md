# 🔄 Data Flow Documentation

This document explains **exactly** how data flows through the entire microservices application, from the moment a client sends a request to when the message is consumed and stored.

---

## 📊 High-Level Flow Overview

```
Client Request
    ↓
API Gateway (Go)
    ↓
Producer Service (Python)
    ↓
Kafka Broker
    ↓
Consumer Service (Node.js)
    ↓
WebSocket Clients (Real-time updates)
```

---

## 🔍 Detailed Step-by-Step Flow

### Step 1: Client Sends HTTP Request

**What happens:**
```bash
curl -X POST http://localhost:8080/api/messages \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello World", "user_id": "user123"}'
```

**Data at this point:**
```json
{
  "message": "Hello World",
  "user_id": "user123"
}
```

**Where it goes:** → API Gateway (Port 8080)

---

### Step 2: API Gateway Receives Request

**File:** `api-gateway/main.go`

**What happens:**
1. Gateway receives POST request at `/api/messages`
2. CORS middleware adds cross-origin headers
3. Logging middleware logs the request
4. Metrics counter increments (`ProducerCalls++`)
5. Reverse proxy forwards request to Producer service

**Code location:**
```go
// Line ~180 in main.go
http.HandleFunc("/api/messages", corsMiddleware(loggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
    metrics.ProducerCalls++
    producerProxy.ServeHTTP(w, r)  // Forward to producer
})))
```

**Data transformation:** None (passes through unchanged)

**Where it goes:** → Producer Service (Port 5000)

---

### Step 3: Producer Service Processes Request

**File:** `producer/app.py`

**What happens:**

1. **Flask receives request** at `/api/messages` endpoint
   ```python
   @app.route('/api/messages', methods=['POST'])
   def send_message():
   ```

2. **Validates JSON data**
   ```python
   data = request.get_json()
   if not data or 'message' not in data:
       return jsonify({'error': 'Message is required'}), 400
   ```

3. **Enriches message with metadata**
   ```python
   message = {
       'message': data['message'],           # Original message
       'user_id': data.get('user_id', 'anonymous'),
       'timestamp': datetime.utcnow().isoformat(),  # When sent
       'service': 'producer'                  # Which service created it
   }
   ```

**Data at this point:**
```json
{
  "message": "Hello World",
  "user_id": "user123",
  "timestamp": "2025-12-03T00:27:55.123456",
  "service": "producer"
}
```

4. **Connects to Kafka**
   ```python
   kafka_producer = get_kafka_producer()  # Connects to kafka:9092
   ```

5. **Publishes to Kafka topic**
   ```python
   future = kafka_producer.send(KAFKA_TOPIC, value=message)
   record_metadata = future.get(timeout=10)  # Wait for confirmation
   ```

6. **Returns success response**
   ```python
   return jsonify({
       'status': 'success',
       'message': 'Message published to Kafka',
       'topic': KAFKA_TOPIC,
       'partition': record_metadata.partition,
       'offset': record_metadata.offset
   }), 201
   ```

**Where it goes:** → Kafka Broker (Port 9092)

---

### Step 4: Kafka Stores Message

**What happens:**

1. **Kafka receives message** on topic `messages`
2. **Determines partition** (based on key or round-robin)
3. **Appends to partition log** (persistent storage)
4. **Assigns offset** (unique ID within partition)
5. **Replicates** (if replication factor > 1)
6. **Acknowledges** back to producer

**Kafka internal structure:**
```
Topic: messages
├── Partition 0
│   ├── Offset 0: {message: "Hello", ...}
│   ├── Offset 1: {message: "World", ...}
│   └── Offset 2: {message: "Hello World", ...}  ← Our message
├── Partition 1
│   └── ...
└── Partition 2
    └── ...
```

**Data persistence:** Message is now stored on disk in Kafka's log files

**Where it goes:** → Consumer Service (pulls from Kafka)

---

### Step 5: Consumer Service Reads Message

**File:** `consumer/app.js`

**What happens:**

1. **Consumer polls Kafka** (continuous loop)
   ```javascript
   await consumer.run({
       eachMessage: async ({ topic, partition, message }) => {
   ```

2. **Receives message** from Kafka
   - Topic: `messages`
   - Partition: `0` (example)
   - Offset: `2` (example)

3. **Parses JSON message**
   ```javascript
   const value = JSON.parse(message.value.toString());
   ```

**Data at this point:**
```json
{
  "message": "Hello World",
  "user_id": "user123",
  "timestamp": "2025-12-03T00:27:55.123456",
  "service": "producer"
}
```

4. **Enriches with consumption metadata**
   ```javascript
   const processedMessage = {
       ...value,                              // Original data
       consumed_at: new Date().toISOString(), // When consumed
       partition: partition,                   // Which partition
       offset: message.offset,                 // Message offset
       consumer_service: 'nodejs-consumer'     // Which consumer
   };
   ```

**Data now:**
```json
{
  "message": "Hello World",
  "user_id": "user123",
  "timestamp": "2025-12-03T00:27:55.123456",
  "service": "producer",
  "consumed_at": "2025-12-03T00:27:56.789012",
  "partition": 0,
  "offset": 2,
  "consumer_service": "nodejs-consumer"
}
```

5. **Stores in memory**
   ```javascript
   processedMessages.push(processedMessage);
   ```

6. **Broadcasts to WebSocket clients**
   ```javascript
   broadcastMessage(processedMessage);
   ```

7. **Kafka commits offset** (marks message as processed)

**Where it goes:** → WebSocket clients + In-memory storage

---

### Step 6: WebSocket Broadcast (Optional)

**File:** `consumer/app.js`

**What happens:**

1. **Iterates through connected clients**
   ```javascript
   wss.clients.forEach((client) => {
       if (client.readyState === WebSocket.OPEN) {
           client.send(payload);
       }
   });
   ```

2. **Sends message to each client**
   ```javascript
   const payload = JSON.stringify({
       type: 'new_message',
       message: processedMessage
   });
   ```

**Data sent to browser:**
```json
{
  "type": "new_message",
  "message": {
    "message": "Hello World",
    "user_id": "user123",
    "timestamp": "2025-12-03T00:27:55.123456",
    "service": "producer",
    "consumed_at": "2025-12-03T00:27:56.789012",
    "partition": 0,
    "offset": 2,
    "consumer_service": "nodejs-consumer"
  }
}
```

**Where it goes:** → Browser/WebSocket clients (real-time update)

---

## 🔄 Complete Data Transformation Journey

### Initial Request
```json
{
  "message": "Hello World",
  "user_id": "user123"
}
```

### After Producer Enrichment
```json
{
  "message": "Hello World",
  "user_id": "user123",
  "timestamp": "2025-12-03T00:27:55.123456",
  "service": "producer"
}
```

### After Consumer Enrichment
```json
{
  "message": "Hello World",
  "user_id": "user123",
  "timestamp": "2025-12-03T00:27:55.123456",
  "service": "producer",
  "consumed_at": "2025-12-03T00:27:56.789012",
  "partition": 0,
  "offset": 2,
  "consumer_service": "nodejs-consumer"
}
```

---

## 🌊 Alternative Flows

### Batch Messages Flow

**Request:**
```bash
curl -X POST http://localhost:8080/api/messages/batch \
  -H "Content-Type: application/json" \
  -d '{"messages": ["msg1", "msg2", "msg3"], "user_id": "user123"}'
```

**Flow:**
1. API Gateway → Producer `/api/messages/batch`
2. Producer loops through array, sends each to Kafka
3. Kafka stores 3 separate messages
4. Consumer processes each message independently
5. 3 WebSocket broadcasts sent

### Query Flow (Getting Stats)

**Request:**
```bash
curl http://localhost:8080/api/stats
```

**Flow:**
1. Client → API Gateway `/api/stats`
2. API Gateway → Consumer `/api/stats`
3. Consumer returns in-memory statistics
4. API Gateway → Client (JSON response)

**No Kafka involved** - direct HTTP request/response

---

## 📡 Network Communication

### Ports Used

| Service | Port | Protocol | Purpose |
|---------|------|----------|---------|
| API Gateway | 8080 | HTTP | External API |
| Producer | 5000 | HTTP | Internal API |
| Consumer | 3000 | HTTP/WS | Internal API + WebSocket |
| Kafka | 9092 | TCP | Message broker |
| Zookeeper | 2181 | TCP | Kafka coordination |

### Service-to-Service Communication

```
Client (external)
    ↓ HTTP
API Gateway (api-gateway:8080)
    ↓ HTTP
Producer (producer:5000)
    ↓ Kafka Protocol (TCP)
Kafka (kafka:9092)
    ↓ Kafka Protocol (TCP)
Consumer (consumer:3000)
    ↓ WebSocket
Browser Clients
```

---

## 🔐 Data Serialization

### HTTP Layer
- **Format:** JSON
- **Encoding:** UTF-8
- **Content-Type:** `application/json`

### Kafka Layer
- **Format:** JSON (serialized to bytes)
- **Encoding:** UTF-8
- **Serializer:** `json.dumps().encode('utf-8')` (Python)
- **Deserializer:** `JSON.parse(buffer.toString())` (Node.js)

---

## ⏱️ Timing & Latency

**Typical flow timing:**

1. **Client → API Gateway:** ~1-5ms (local network)
2. **API Gateway → Producer:** ~1-5ms (container network)
3. **Producer → Kafka:** ~5-20ms (write + ack)
4. **Kafka → Consumer:** ~1-10ms (poll interval)
5. **Consumer → WebSocket:** ~1-5ms (broadcast)

**Total end-to-end:** ~10-45ms

---

## 🔄 Message Lifecycle

```
1. CREATED     → Client creates message
2. VALIDATED   → Producer validates format
3. ENRICHED    → Producer adds metadata
4. PUBLISHED   → Sent to Kafka
5. PERSISTED   → Kafka writes to disk
6. CONSUMED    → Consumer reads from Kafka
7. PROCESSED   → Consumer enriches data
8. STORED      → Saved in memory
9. BROADCASTED → Sent to WebSocket clients
10. DISPLAYED  → Shown in browser
```

---

## 🎯 Key Takeaways

1. **API Gateway** acts as a single entry point and router
2. **Producer** enriches data before publishing
3. **Kafka** provides durable, ordered message storage
4. **Consumer** processes messages and broadcasts updates
5. **Data is enriched** at each stage with metadata
6. **WebSocket** enables real-time updates to browsers
7. **Each service** has a specific responsibility (separation of concerns)

---

## 🧪 Tracing a Message

To trace a message through the system:

1. **Send a message** with a unique identifier
2. **Check producer logs:** `docker-compose logs producer`
3. **Check Kafka:** `docker exec -it kafka kafka-console-consumer --bootstrap-server localhost:9092 --topic messages --from-beginning`
4. **Check consumer logs:** `docker-compose logs consumer`
5. **Check API Gateway metrics:** `curl http://localhost:8080/metrics`

---

## 📚 Related Files

- **API Gateway routing:** [api-gateway/main.go](file:///home/reagan/Desktop/phone/devops-learning-project/api-gateway/main.go) (lines 150-200)
- **Producer publishing:** [producer/app.py](file:///home/reagan/Desktop/phone/devops-learning-project/producer/app.py) (lines 60-100)
- **Consumer processing:** [consumer/app.js](file:///home/reagan/Desktop/phone/devops-learning-project/consumer/app.js) (lines 80-130)
- **Docker networking:** [docker-compose.yml](file:///home/reagan/Desktop/phone/devops-learning-project/docker-compose.yml) (networks section)

---

**Now you understand exactly how data flows through the entire system!** 🎉
