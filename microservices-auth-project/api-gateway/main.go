package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var (
	AUTH_SERVICE_URL    string
	USER_SERVICE_URL    string
	PRODUCT_SERVICE_URL string
	ORDER_SERVICE_URL   string
	PORT                string
)

func init() {
	godotenv.Load()

	AUTH_SERVICE_URL = getEnv("AUTH_SERVICE_URL", "http://auth-service:4000")
	USER_SERVICE_URL = getEnv("USER_SERVICE_URL", "http://user-service:5000")
	PRODUCT_SERVICE_URL = getEnv("PRODUCT_SERVICE_URL", "http://product-service:3000")
	ORDER_SERVICE_URL = getEnv("ORDER_SERVICE_URL", "http://order-service:6000")
	PORT = getEnv("PORT", "8080")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Proxy Helper
func newProxy(target string) *httputil.ReverseProxy {
	url, _ := url.Parse(target)
	return httputil.NewSingleHostReverseProxy(url)
}

// CORS Middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Logging Middleware
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Proxies
	authProxy := newProxy(AUTH_SERVICE_URL)
	userProxy := newProxy(USER_SERVICE_URL)
	productProxy := newProxy(PRODUCT_SERVICE_URL)
	orderProxy := newProxy(ORDER_SERVICE_URL)

	// Router
	mux := http.NewServeMux()

	// Health Check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy","service":"api-gateway"}`))
	})

	// Route Handlers
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		switch {
		case strings.HasPrefix(path, "/auth"):
			authProxy.ServeHTTP(w, r)
		case strings.HasPrefix(path, "/users"):
			userProxy.ServeHTTP(w, r)
		case strings.HasPrefix(path, "/products"):
			productProxy.ServeHTTP(w, r)
		case strings.HasPrefix(path, "/orders"):
			orderProxy.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})

	// Apply Middleware
	handler := corsMiddleware(loggingMiddleware(mux))

	log.Printf("API Gateway running on port %s", PORT)
	log.Fatal(http.ListenAndServe(":"+PORT, handler))
}
