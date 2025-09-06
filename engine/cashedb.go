package engine

import (
	"strings"
	"sync"
)

// RamStorage is a concurrent in-memory key-value store optimized for large key counts.
type CasheDb struct {
	store sync.Map // map[string][]byte
}

// Set stores a value for a key.
func (r *CasheDb) Set(key string, value []byte) {
	r.store.Store(key, value)
}

// Get retrieves the value for a key.
func (r *CasheDb) Get(key string) ([]byte, bool) {
	val, ok := r.store.Load(key)
	if !ok {
		return nil, false
	}
	return val.([]byte), true
}

// Exists checks if a key exists.
func (r *CasheDb) Exists(key string) bool {
	_, ok := r.store.Load(key)
	return ok
}

// Remove deletes a key.
func (r *CasheDb) Remove(key string) {
	r.store.Delete(key)
}

// Scan returns all keys matching a prefix.
func (r *CasheDb) Scan(prefix string) []string {
	var keys []string
	r.store.Range(func(k, _ any) bool {
		key := k.(string)
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
		return true
	})
	return keys
}

// MGet returns values for a slice of keys.
func (r *CasheDb) MGet(keys []string) [][]byte {
	results := make([][]byte, len(keys))
	for i, k := range keys {
		if v, ok := r.Get(k); ok {
			results[i] = v
		} else {
			results[i] = nil
		}
	}
	return results
}
