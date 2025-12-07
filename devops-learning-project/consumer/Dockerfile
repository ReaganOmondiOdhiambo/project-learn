# DOCKERFILE FOR NODE.JS CONSUMER SERVICE
# ========================================

# Use official Node.js LTS (Long Term Support) image
# Alpine variant is much smaller (~50MB vs ~900MB)
FROM node:18-alpine

# Add metadata
LABEL maintainer="devops-learning"
LABEL description="Kafka Consumer Service with WebSocket"

# Set working directory
WORKDIR /app

# Copy package files first (for layer caching)
# If package.json doesn't change, npm install layer is reused
COPY package*.json ./

# Install dependencies
# --production flag skips devDependencies (smaller image)
# --silent reduces npm output noise
RUN npm install --production --silent

# Copy application code
COPY app.js .

# Expose HTTP and WebSocket port
EXPOSE 3000

# Set Node environment to production
ENV NODE_ENV=production

# Health check - verifies the service is responding
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD node -e "require('http').get('http://localhost:3000/health', (r) => {process.exit(r.statusCode === 200 ? 0 : 1)})"

# Run the application
CMD ["node", "app.js"]
