/**
 * API GATEWAY - Go-based API Gateway
 * ===================================
 * 
 * This service demonstrates:
 * - API Gateway pattern (single entry point for microservices)
 * - Request routing and proxying
 * - Load balancing across multiple instances
 * - CORS handling
 * - Metrics collection
 * 
 * Written in Go for high performance and low resource usage.
 */

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

// Configuration from environment variables
var (
	PRODUCER_URL = getEnv("PRODUCER_URL", "http://producer:5000")
	CONSUMER_URL = getEnv("CONSUMER_URL", "http://consumer:3000")
	PORT         = getEnv("PORT", "8080")
)

// Helper function to get environment variable with default
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Metrics structure to track API usage
type Metrics struct {
	TotalRequests   int64 `json:"total_requests"`
	ProducerCalls   int64 `json:"producer_calls"`
	ConsumerCalls   int64 `json:"consumer_calls"`
	HealthChecks    int64 `json:"health_checks"`
	LastRequestTime string `json:"last_request_time"`
}

var metrics = &Metrics{}

/**
 * CORS Middleware
 * Adds Cross-Origin Resource Sharing headers
 * This allows browsers to make requests from different domains
 */
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next(w, r)
	}
}

/**
 * Logging Middleware
 * Logs all incoming requests with timing information
 */
func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Call the next handler
		next(w, r)
		
		// Log request details
		duration := time.Since(start)
		log.Printf("%s %s - %v", r.Method, r.URL.Path, duration)
		
		// Update metrics
		metrics.TotalRequests++
		metrics.LastRequestTime = time.Now().Format(time.RFC3339)
	}
}

/**
 * Create a reverse proxy to forward requests to backend services
 * This is the core of the API Gateway pattern
 */
func createReverseProxy(targetURL string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}
	
	proxy := httputil.NewSingleHostReverseProxy(url)
	
	// Customize the proxy to add error handling
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Service unavailable",
			"details": err.Error(),
		})
	}
	
	return proxy, nil
}

/**
 * Health check endpoint
 * Checks health of gateway and downstream services
 */
func healthHandler(w http.ResponseWriter, r *http.Request) {
	metrics.HealthChecks++
	
	// Check producer health
	producerHealthy := checkServiceHealth(PRODUCER_URL + "/health")
	
	// Check consumer health
	consumerHealthy := checkServiceHealth(CONSUMER_URL + "/health")
	
	status := "healthy"
	statusCode := http.StatusOK
	
	if !producerHealthy || !consumerHealthy {
		status = "degraded"
		statusCode = http.StatusServiceUnavailable
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": status,
		"service": "api-gateway",
		"timestamp": time.Now().Format(time.RFC3339),
		"services": map[string]bool{
			"producer": producerHealthy,
			"consumer": consumerHealthy,
		},
	})
}

/**
 * Check if a service is healthy by calling its health endpoint
 */
func checkServiceHealth(healthURL string) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(healthURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

/**
 * Metrics endpoint
 * Returns API usage statistics
 */
func metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

/**
 * Root endpoint with API documentation
 */
func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service": "API Gateway",
		"version": "1.0.0",
		"endpoints": map[string]string{
			"POST /api/messages": "Send message to Kafka (proxied to producer)",
			"POST /api/messages/batch": "Send batch messages (proxied to producer)",
			"GET /api/messages": "Get consumed messages (proxied to consumer)",
			"GET /api/stats": "Get consumer stats (proxied to consumer)",
			"GET /health": "Health check",
			"GET /metrics": "API metrics",
		},
	})
}

func main() {
	// Create reverse proxies for each service
	producerProxy, err := createReverseProxy(PRODUCER_URL)
	if err != nil {
		log.Fatalf("Failed to create producer proxy: %v", err)
	}
	
	consumerProxy, err := createReverseProxy(CONSUMER_URL)
	if err != nil {
		log.Fatalf("Failed to create consumer proxy: %v", err)
	}
	
	// Set up routes
	http.HandleFunc("/", corsMiddleware(loggingMiddleware(rootHandler)))
	http.HandleFunc("/health", corsMiddleware(loggingMiddleware(healthHandler)))
	http.HandleFunc("/metrics", corsMiddleware(loggingMiddleware(metricsHandler)))
	
	// Producer routes - forward to producer service
	http.HandleFunc("/api/messages", corsMiddleware(loggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		metrics.ProducerCalls++
		producerProxy.ServeHTTP(w, r)
	})))
	
	http.HandleFunc("/api/messages/batch", corsMiddleware(loggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		metrics.ProducerCalls++
		producerProxy.ServeHTTP(w, r)
	})))
	
	// Consumer routes - forward to consumer service
	http.HandleFunc("/api/stats", corsMiddleware(loggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		metrics.ConsumerCalls++
		consumerProxy.ServeHTTP(w, r)
	})))
	
	// Start server
	addr := "0.0.0.0:" + PORT
	log.Printf("API Gateway starting on %s", addr)
	log.Printf("Proxying to Producer: %s", PRODUCER_URL)
	log.Printf("Proxying to Consumer: %s", CONSUMER_URL)
	
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
