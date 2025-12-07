# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod ./
# If you had go.sum, you'd copy it here too. Since we're creating fresh, we'll run tidy.
# COPY go.sum ./

# Download dependencies
# Copy source
COPY . .

# Download dependencies
RUN go mod tidy
RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o product-service .

# Final stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/product-service .

EXPOSE 3000

CMD ["./product-service"]
