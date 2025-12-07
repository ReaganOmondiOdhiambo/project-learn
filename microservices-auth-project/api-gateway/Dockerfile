# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod ./

COPY . .

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -o api-gateway .

# Final stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/api-gateway .

EXPOSE 8080

CMD ["./api-gateway"]
