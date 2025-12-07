/**
 * CONSUMER SERVICE - Message Consumer with Kafka
 * ===============================================
 * 
 * This Node.js service demonstrates:
 * - Kafka consumer implementation
 * - Real-time message processing
 * - WebSocket for live updates
 * - Express.js REST API
 * - Docker containerization
 * 
 * The consumer reads messages from Kafka and can broadcast them via WebSocket.
 */

const express = require('express');
const { Kafka } = require('kafkajs');
const http = require('http');
const WebSocket = require('ws');

// Configuration from environment variables
const KAFKA_BROKER = process.env.KAFKA_BROKER || 'kafka:9092';
const KAFKA_TOPIC = process.env.KAFKA_TOPIC || 'messages';
const KAFKA_GROUP_ID = process.env.KAFKA_GROUP_ID || 'consumer-group-1';
const PORT = process.env.PORT || 3000;

// Initialize Express app
const app = express();
const server = http.createServer(app);

// Initialize WebSocket server for real-time updates
// This allows browsers to receive messages in real-time
const wss = new WebSocket.Server({ server });

// Store processed messages in memory (for demo purposes)
// In production, you'd use a database
const processedMessages = [];
const MAX_STORED_MESSAGES = 100;

// Track connected WebSocket clients
let connectedClients = 0;

/**
 * WebSocket connection handler
 * Clients can connect to receive real-time message updates
 */
wss.on('connection', (ws) => {
    connectedClients++;
    console.log(`New WebSocket client connected. Total clients: ${connectedClients}`);

    // Send recent messages to newly connected client
    ws.send(JSON.stringify({
        type: 'history',
        messages: processedMessages.slice(-10) // Last 10 messages
    }));

    ws.on('close', () => {
        connectedClients--;
        console.log(`Client disconnected. Total clients: ${connectedClients}`);
    });
});

/**
 * Broadcast message to all connected WebSocket clients
 */
function broadcastMessage(message) {
    const payload = JSON.stringify({
        type: 'new_message',
        message: message
    });

    wss.clients.forEach((client) => {
        if (client.readyState === WebSocket.OPEN) {
            client.send(payload);
        }
    });
}

/**
 * Initialize Kafka Consumer
 * KafkaJS is a modern Kafka client for Node.js
 */
const kafka = new Kafka({
    clientId: 'consumer-service',
    brokers: [KAFKA_BROKER],
    // Retry configuration for resilience
    retry: {
        initialRetryTime: 100,
        retries: 8
    }
});

const consumer = kafka.consumer({
    groupId: KAFKA_GROUP_ID,
    // Start reading from the earliest message if no offset exists
    // This ensures we don't miss messages on first startup
    sessionTimeout: 30000,
    heartbeatInterval: 3000
});

/**
 * Start consuming messages from Kafka
 */
async function startConsumer() {
    try {
        // Connect to Kafka
        await consumer.connect();
        console.log('Connected to Kafka broker');

        // Subscribe to the topic
        await consumer.subscribe({
            topic: KAFKA_TOPIC,
            fromBeginning: false // Only read new messages
        });

        console.log(`Subscribed to topic: ${KAFKA_TOPIC}`);

        // Start consuming messages
        await consumer.run({
            // eachMessage is called for every message received
            eachMessage: async ({ topic, partition, message }) => {
                try {
                    // Parse the message value (it's JSON)
                    const value = JSON.parse(message.value.toString());

                    // Add processing metadata
                    const processedMessage = {
                        ...value,
                        consumed_at: new Date().toISOString(),
                        partition: partition,
                        offset: message.offset,
                        consumer_service: 'nodejs-consumer'
                    };

                    // Store the message
                    processedMessages.push(processedMessage);

                    // Keep only the last MAX_STORED_MESSAGES
                    if (processedMessages.length > MAX_STORED_MESSAGES) {
                        processedMessages.shift();
                    }

                    // Log the message
                    console.log('Processed message:', {
                        offset: message.offset,
                        partition: partition,
                        message: value.message
                    });

                    // Broadcast to WebSocket clients
                    broadcastMessage(processedMessage);

                } catch (error) {
                    console.error('Error processing message:', error);
                }
            },
        });

    } catch (error) {
        console.error('Error starting consumer:', error);
        // Retry connection after delay
        setTimeout(startConsumer, 5000);
    }
}

// REST API Endpoints

app.use(express.json());

/**
 * Health check endpoint
 * Used by container orchestration to verify service health
 */
app.get('/health', (req, res) => {
    res.json({
        status: 'healthy',
        service: 'consumer',
        connected_clients: connectedClients,
        messages_processed: processedMessages.length,
        timestamp: new Date().toISOString()
    });
});

/**
 * Get all processed messages
 * Useful for debugging and viewing message history
 */
app.get('/api/messages', (req, res) => {
    const limit = parseInt(req.query.limit) || 50;
    res.json({
        total: processedMessages.length,
        messages: processedMessages.slice(-limit)
    });
});

/**
 * Get statistics about message processing
 */
app.get('/api/stats', (req, res) => {
    res.json({
        total_messages: processedMessages.length,
        connected_websocket_clients: connectedClients,
        kafka_topic: KAFKA_TOPIC,
        kafka_group: KAFKA_GROUP_ID
    });
});

// Graceful shutdown handler
// This ensures we close connections properly when container stops
process.on('SIGTERM', async () => {
    console.log('SIGTERM received, shutting down gracefully...');
    await consumer.disconnect();
    server.close(() => {
        console.log('Server closed');
        process.exit(0);
    });
});

// Start the consumer
startConsumer();

// Start the HTTP server
server.listen(PORT, '0.0.0.0', () => {
    console.log(`Consumer service listening on port ${PORT}`);
    console.log(`WebSocket server ready for connections`);
});
