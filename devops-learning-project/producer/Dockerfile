# DOCKERFILE EXPLAINED
# ====================
# A Dockerfile is a recipe for building a Docker image.
# Each instruction creates a new layer in the image.

# FROM - Base image to build upon
# We use Python 3.11 slim version (smaller size, fewer packages)
FROM python:3.11-slim

# LABEL - Metadata about the image (good practice for documentation)
LABEL maintainer="devops-learning"
LABEL description="Kafka Producer Service"

# WORKDIR - Sets the working directory inside the container
# All subsequent commands run from this directory
WORKDIR /app

# COPY - Copy files from host to container
# Copy requirements first (Docker layer caching optimization)
# If requirements don't change, Docker reuses this layer
COPY requirements.txt .

# RUN - Execute commands during image build
# Install Python dependencies
# --no-cache-dir reduces image size by not caching pip downloads
RUN pip install --no-cache-dir -r requirements.txt

# COPY - Copy application code
# This is done after installing dependencies for better caching
COPY app.py .

# EXPOSE - Documents which port the container listens on
# This is informational - doesn't actually publish the port
EXPOSE 5000

# ENV - Set environment variables
# These can be overridden when running the container
ENV PYTHONUNBUFFERED=1

# HEALTHCHECK - Tells Docker how to test if container is healthy
# Docker will run this command periodically
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD python -c "import requests; requests.get('http://localhost:5000/health')" || exit 1

# CMD - Default command to run when container starts
# This starts the Flask application
CMD ["python", "app.py"]
