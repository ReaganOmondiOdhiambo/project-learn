package hashring

import (
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

// HashRing implements consistent hashing
type HashRing struct {
	replicas int               // Number of virtual nodes per physical node
	keys     []int             // Sorted hash keys
	hashMap  map[int]string    // Map hash key to physical node
	lock     sync.RWMutex
}

// New creates a new HashRing
func New(replicas int) *HashRing {
	return &HashRing{
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
}

// Add adds keys (physical nodes) to the ring
func (h *HashRing) Add(nodes ...string) {
	h.lock.Lock()
	defer h.lock.Unlock()

	for _, node := range nodes {
		// Create virtual nodes
		for i := 0; i < h.replicas; i++ {
			// Create a unique key for the virtual node: "shard-0:1", "shard-0:2", etc.
			virtualKey := node + ":" + strconv.Itoa(i)
			hash := int(crc32.ChecksumIEEE([]byte(virtualKey)))
			
			h.keys = append(h.keys, hash)
			h.hashMap[hash] = node
		}
	}
	
	// Sort keys for binary search
	sort.Ints(h.keys)
}

// Get returns the closest node in the ring for the given key
func (h *HashRing) Get(key string) string {
	h.lock.RLock()
	defer h.lock.RUnlock()

	if len(h.keys) == 0 {
		return ""
	}

	hash := int(crc32.ChecksumIEEE([]byte(key)))

	// Binary search for appropriate replica
	idx := sort.Search(len(h.keys), func(i int) bool {
		return h.keys[i] >= hash
	})

	// If we've gone past the end, wrap around to the first key
	if idx == len(h.keys) {
		idx = 0
	}

	return h.hashMap[h.keys[idx]]
}

// GetNodes returns all physical nodes in the ring
func (h *HashRing) GetNodes() []string {
	h.lock.RLock()
	defer h.lock.RUnlock()
	
	uniqueNodes := make(map[string]bool)
	for _, node := range h.hashMap {
		uniqueNodes[node] = true
	}
	
	var nodes []string
	for node := range uniqueNodes {
		nodes = append(nodes, node)
	}
	return nodes
}
