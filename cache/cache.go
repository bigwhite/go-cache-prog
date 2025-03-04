package cache

import (
	"fmt"
	"sync"
	"time"

	"github.com/bigwhite/go-cache-prog/storage"
)

type CacheEntry struct {
	OutputID []byte
	Size     int64
	Time     time.Time
	DiskPath string
}

type Cache struct {
	entries map[string]CacheEntry
	mu      sync.RWMutex
	store   storage.Storage
}

func NewCache(store storage.Storage) *Cache {
	return &Cache{
		entries: make(map[string]CacheEntry),
		store:   store,
	}
}

func (c *Cache) Put(actionID, outputID []byte, data []byte, size int64) (string, error) {
	diskPath, err := c.store.Put(actionID, outputID, data, size)
	if err != nil {
		return "", err
	}

	entry := CacheEntry{
		OutputID: outputID,
		Size:     size,
		Time:     time.Now(), // Use current time; storage backend might update it
		DiskPath: diskPath,
	}

	actionIDHex := fmt.Sprintf("%x", actionID)
	c.mu.Lock()
	c.entries[actionIDHex] = entry
	c.mu.Unlock()

	return diskPath, nil
}

func (c *Cache) Get(actionID []byte) (*CacheEntry, bool, error) {
	actionIDHex := fmt.Sprintf("%x", actionID)

	c.mu.RLock()
	entry, exists := c.entries[actionIDHex]
	c.mu.RUnlock()

	if exists {
		return &entry, true, nil
	}

	// If not in the in-memory cache, try the storage backend.
	outputID, size, modTime, diskPath, found, err := c.store.Get(actionID)
	if err != nil {
		return nil, false, err
	}
	if !found {
		//Not found neither in memory cache, nor in file system
		c.mu.Lock()
		delete(c.entries, actionIDHex)
		c.mu.Unlock()

		return nil, false, nil
	}

	// Populate the in-memory cache.
	entry = CacheEntry{
		OutputID: outputID,
		Size:     size,
		Time:     modTime,
		DiskPath: diskPath,
	}
	c.mu.Lock()
	c.entries[actionIDHex] = entry
	c.mu.Unlock()

	return &entry, true, nil
}

// Add Remove and Clear methods if needed for cache eviction.
