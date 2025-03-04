package storage

import (
	"time"
)

// Storage interface defines the methods any storage backend must implement.
type Storage interface {
	Put(actionID, outputID []byte, data []byte, size int64) (string, error) // Returns the disk path
	Get(actionID []byte) (outputID []byte, size int64, modTime time.Time, diskPath string, found bool, err error)
	//Remove(actionID []byte) error // Optional: For cache eviction policies
	//Clear() error               // Optional: For clearing the entire cache
}
