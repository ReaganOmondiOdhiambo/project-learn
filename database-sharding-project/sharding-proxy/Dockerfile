# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod ./
# Copy source
COPY . .

# Download dependencies
RUN go mod tidy

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o sharding-proxy .

# Final stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/sharding-proxy .

EXPOSE 8080

CMD ["./sharding-proxy"]
