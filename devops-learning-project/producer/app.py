"""
PRODUCER SERVICE - Message Producer with Kafka
================================================

This service demonstrates:
- REST API creation with Flask
- Kafka producer implementation
- Message publishing to Kafka topics
- Docker containerization
- Health checks for container orchestration

The producer accepts HTTP POST requests and publishes messages to Kafka.
"""

from flask import Flask, request, jsonify
from kafka import KafkaProducer
import json
import os
import logging
from datetime import datetime

# Configure logging to see what's happening
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = Flask(__name__)

# Environment variables for configuration (12-factor app principle)
# This allows us to change config without changing code
KAFKA_BROKER = os.getenv('KAFKA_BROKER', 'kafka:9092')
KAFKA_TOPIC = os.getenv('KAFKA_TOPIC', 'messages')

# Initialize Kafka Producer
# This connects to the Kafka broker and prepares to send messages
producer = None

def get_kafka_producer():
    """
    Lazy initialization of Kafka producer.
    This pattern allows the app to start even if Kafka isn't ready yet.
    """
    global producer
    if producer is None:
        try:
            producer = KafkaProducer(
                bootstrap_servers=[KAFKA_BROKER],
                # Serialize messages to JSON format
                value_serializer=lambda v: json.dumps(v).encode('utf-8'),
                # Wait for acknowledgment from Kafka leader
                acks='all',
                # Retry failed sends
                retries=3
            )
            logger.info(f"Connected to Kafka broker at {KAFKA_BROKER}")
        except Exception as e:
            logger.error(f"Failed to connect to Kafka: {e}")
            raise
    return producer

@app.route('/health', methods=['GET'])
def health_check():
    """
    Health check endpoint for container orchestration.
    Kubernetes/Docker uses this to know if the container is healthy.
    """
    return jsonify({
        'status': 'healthy',
        'service': 'producer',
        'timestamp': datetime.utcnow().isoformat()
    }), 200

@app.route('/api/messages', methods=['POST'])
def send_message():
    """
    API endpoint to receive messages and publish them to Kafka.
    
    Expected JSON body:
    {
        "message": "Your message here",
        "user_id": "optional_user_id"
    }
    """
    try:
        # Get JSON data from request
        data = request.get_json()
        
        if not data or 'message' not in data:
            return jsonify({'error': 'Message is required'}), 400
        
        # Prepare message with metadata
        message = {
            'message': data['message'],
            'user_id': data.get('user_id', 'anonymous'),
            'timestamp': datetime.utcnow().isoformat(),
            'service': 'producer'
        }
        
        # Get Kafka producer instance
        kafka_producer = get_kafka_producer()
        
        # Send message to Kafka topic
        # This is asynchronous - message is queued and sent in background
        future = kafka_producer.send(KAFKA_TOPIC, value=message)
        
        # Wait for confirmation (optional, makes it synchronous)
        record_metadata = future.get(timeout=10)
        
        logger.info(f"Message sent to topic {record_metadata.topic} partition {record_metadata.partition}")
        
        return jsonify({
            'status': 'success',
            'message': 'Message published to Kafka',
            'topic': KAFKA_TOPIC,
            'partition': record_metadata.partition,
            'offset': record_metadata.offset
        }), 201
        
    except Exception as e:
        logger.error(f"Error publishing message: {e}")
        return jsonify({'error': str(e)}), 500

@app.route('/api/messages/batch', methods=['POST'])
def send_batch_messages():
    """
    Endpoint to send multiple messages at once.
    Useful for load testing and demonstrating Kafka's throughput.
    
    Expected JSON body:
    {
        "messages": ["msg1", "msg2", "msg3"],
        "user_id": "optional_user_id"
    }
    """
    try:
        data = request.get_json()
        
        if not data or 'messages' not in data or not isinstance(data['messages'], list):
            return jsonify({'error': 'Messages array is required'}), 400
        
        kafka_producer = get_kafka_producer()
        sent_count = 0
        
        # Send each message
        for msg in data['messages']:
            message = {
                'message': msg,
                'user_id': data.get('user_id', 'anonymous'),
                'timestamp': datetime.utcnow().isoformat(),
                'service': 'producer'
            }
            kafka_producer.send(KAFKA_TOPIC, value=message)
            sent_count += 1
        
        # Flush to ensure all messages are sent
        kafka_producer.flush()
        
        logger.info(f"Batch of {sent_count} messages sent to Kafka")
        
        return jsonify({
            'status': 'success',
            'messages_sent': sent_count,
            'topic': KAFKA_TOPIC
        }), 201
        
    except Exception as e:
        logger.error(f"Error in batch send: {e}")
        return jsonify({'error': str(e)}), 500

if __name__ == '__main__':
    # Run Flask app
    # 0.0.0.0 makes it accessible from outside the container
    # Port 5000 is the default Flask port
    app.run(host='0.0.0.0', port=5000, debug=False)
