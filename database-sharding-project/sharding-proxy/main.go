package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"sharding-proxy/hashring"

	_ "github.com/lib/pq"
)

// Config
var (
	PORT           = getEnv("PORT", "8080")
	DB_USER        = getEnv("DB_USER", "postgres")
	DB_PASSWORD    = getEnv("DB_PASSWORD", "FERODO2001")
	DB_NAME        = getEnv("DB_NAME", "postgres")
	SHARDS         = getEnv("SHARDS", "postgres-shard-0,postgres-shard-1,postgres-shard-2")
	VIRTUAL_NODES  = 50 // Virtual nodes for consistent hashing
)

// Global State
var (
	ring    *hashring.HashRing
	dbPools = make(map[string]*sql.DB)
	poolMu  sync.RWMutex
)

// Data Model
type Entry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Response struct {
	Message string `json:"message"`
	Shard   string `json:"shard"`
	Data    any    `json:"data,omitempty"`
}

func init() {
	// Initialize Hash Ring
	ring = hashring.New(VIRTUAL_NODES)
	
	// Add initial shards
	shards := strings.Split(SHARDS, ",")
	ring.Add(shards...)
	
	log.Printf("Initialized Hash Ring with shards: %v", shards)
}

func main() {
	// Initialize DB Connections
	initDBConnections()

	http.HandleFunc("/write", writeHandler)
	http.HandleFunc("/read", readHandler)
	http.HandleFunc("/stats", statsHandler)
	http.HandleFunc("/health", healthHandler)

	log.Printf("Sharding Proxy running on port %s", PORT)
	log.Fatal(http.ListenAndServe(":"+PORT, nil))
}

// Connect to all known shards
func initDBConnections() {
	poolMu.Lock()
	defer poolMu.Unlock()

	nodes := ring.GetNodes()
	for _, node := range nodes {
		if _, exists := dbPools[node]; !exists {
			// Construct connection string
			// In K8s, the service DNS is usually: pod-name.service-name
			// We assume the node name IS the DNS name for simplicity in this demo
			connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
				node, DB_USER, DB_PASSWORD, DB_NAME)
			
			db, err := sql.Open("postgres", connStr)
			if err != nil {
				log.Printf("Error creating connection for %s: %v", node, err)
				continue
			}

			// Test connection
			// Note: We don't fail fatal here because shards might be starting up
			go func(n string, d *sql.DB) {
				for i := 0; i < 5; i++ {
					if err := d.Ping(); err == nil {
						log.Printf("✅ Connected to shard: %s", n)
						
						// Create table if not exists
						_, _ = d.Exec(`CREATE TABLE IF NOT EXISTS data (
							key TEXT PRIMARY KEY,
							value TEXT,
							created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
						)`)
						return
					}
					time.Sleep(2 * time.Second)
				}
				log.Printf("⚠️  Could not connect to shard %s after retries", n)
			}(node, db)

			dbPools[node] = db
		}
	}
}

func getDB(shard string) (*sql.DB, error) {
	poolMu.RLock()
	defer poolMu.RUnlock()
	
	db, ok := dbPools[shard]
	if !ok {
		return nil, fmt.Errorf("connection not found for shard %s", shard)
	}
	return db, nil
}

// Handlers

func writeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var entry Entry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	if entry.Key == "" {
		http.Error(w, "Key is required", 400)
		return
	}

	// 1. Determine Shard
	shard := ring.Get(entry.Key)
	
	// 2. Get DB Connection
	db, err := getDB(shard)
	if err != nil {
		http.Error(w, fmt.Sprintf("Shard error: %v", err), 500)
		return
	}

	// 3. Write to DB
	_, err = db.Exec("INSERT INTO data (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = $2", 
		entry.Key, entry.Value)
	
	if err != nil {
		log.Printf("Write error to %s: %v", shard, err)
		http.Error(w, "Database error", 500)
		return
	}

	json.NewEncoder(w).Encode(Response{
		Message: "Write successful",
		Shard:   shard,
		Data:    entry,
	})
}

func readHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Key parameter required", 400)
		return
	}

	// 1. Determine Shard
	shard := ring.Get(key)

	// 2. Get DB Connection
	db, err := getDB(shard)
	if err != nil {
		http.Error(w, fmt.Sprintf("Shard error: %v", err), 500)
		return
	}

	// 3. Read from DB
	var value string
	var createdAt time.Time
	err = db.QueryRow("SELECT value, created_at FROM data WHERE key = $1", key).Scan(&value, &createdAt)
	
	if err == sql.ErrNoRows {
		http.Error(w, "Key not found", 404)
		return
	} else if err != nil {
		log.Printf("Read error from %s: %v", shard, err)
		http.Error(w, "Database error", 500)
		return
	}

	json.NewEncoder(w).Encode(Response{
		Message: "Read successful",
		Shard:   shard,
		Data: map[string]interface{}{
			"key":        key,
			"value":      value,
			"created_at": createdAt,
		},
	})
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	// Simple stats showing shard availability
	poolMu.RLock()
	defer poolMu.RUnlock()
	
	stats := make(map[string]string)
	for name, db := range dbPools {
		if err := db.Ping(); err != nil {
			stats[name] = "down"
		} else {
			stats[name] = "up"
		}
	}
	
	json.NewEncoder(w).Encode(map[string]interface{}{
		"shards": stats,
		"total_shards": len(stats),
		"algorithm": "consistent_hashing",
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
