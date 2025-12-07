#!/usr/bin/env python3

"""
LOAD TESTING WITH LOCUST
=========================
This is a more sophisticated load testing tool using Locust.
It provides a web UI to monitor load testing in real-time.

Install: pip install locust
Run: locust -f locustfile.py --host=http://localhost:8080
Then open http://localhost:8089 in your browser
"""

from locust import HttpUser, task, between
import json
import random

class MessageUser(HttpUser):
    """
    Simulates a user sending messages to the API.
    Locust will spawn multiple instances of this user.
    """
    
    # Wait between 1-3 seconds between tasks
    wait_time = between(1, 3)
    
    @task(3)  # Weight: 3 (runs 3x more often than other tasks)
    def send_single_message(self):
        """Send a single message to the producer"""
        payload = {
            "message": f"Test message {random.randint(1, 10000)}",
            "user_id": f"user_{random.randint(1, 100)}"
        }
        
        with self.client.post(
            "/api/messages",
            json=payload,
            catch_response=True
        ) as response:
            if response.status_code == 201:
                response.success()
            else:
                response.failure(f"Got status code {response.status_code}")
    
    @task(1)  # Weight: 1
    def send_batch_messages(self):
        """Send a batch of messages"""
        batch_size = random.randint(5, 20)
        payload = {
            "messages": [f"Batch message {i}" for i in range(batch_size)],
            "user_id": f"batch_user_{random.randint(1, 50)}"
        }
        
        with self.client.post(
            "/api/messages/batch",
            json=payload,
            catch_response=True
        ) as response:
            if response.status_code == 201:
                response.success()
            else:
                response.failure(f"Got status code {response.status_code}")
    
    @task(2)  # Weight: 2
    def get_stats(self):
        """Get consumer statistics"""
        with self.client.get("/api/stats", catch_response=True) as response:
            if response.status_code == 200:
                response.success()
            else:
                response.failure(f"Got status code {response.status_code}")
    
    @task(1)
    def health_check(self):
        """Check API health"""
        with self.client.get("/health", catch_response=True) as response:
            if response.status_code == 200:
                response.success()
            else:
                response.failure(f"Got status code {response.status_code}")

    def on_start(self):
        """Called when a simulated user starts"""
        print(f"User {self.user_id} started")
    
    def on_stop(self):
        """Called when a simulated user stops"""
        print(f"User {self.user_id} stopped")
