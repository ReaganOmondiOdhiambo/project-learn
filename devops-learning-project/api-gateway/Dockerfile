# MULTI-STAGE DOCKERFILE FOR GO
# ==============================
# Multi-stage builds create smaller, more secure images
# Stage 1 builds the app, Stage 2 runs it

# Stage 1: Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /build

# Copy go module files
COPY go.mod ./

# Download dependencies (cached if go.mod doesn't change)
RUN go mod download

# Copy source code
COPY main.go ./

# Build the application
# CGO_ENABLED=0 creates a static binary (no external dependencies)
# -ldflags="-w -s" strips debug info (smaller binary)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o api-gateway .

# Stage 2: Runtime stage
# Use minimal scratch image (only ~2MB!)
FROM alpine:latest

# Add ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy only the binary from builder stage
COPY --from=builder /build/api-gateway .

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the binary
CMD ["./api-gateway"]
