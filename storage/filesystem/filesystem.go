package filesystem

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type FileSystemStorage struct {
	baseDir string
	verbose bool
}

type indexEntry struct {
	Version  int        `json:"v"`
	OutputID []byte     `json:"o"`
	Size     int64      `json:"n"`
	Time     *time.Time `json:"t"`
}

func NewFileSystemStorage(baseDir string, verbose bool) (*FileSystemStorage, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}
	return &FileSystemStorage{baseDir: baseDir, verbose: verbose}, nil
}

func (fss *FileSystemStorage) Put(actionID, outputID []byte, data []byte, size int64) (string, error) {
	actionIDHex := fmt.Sprintf("%x", actionID)
	//outputIDHex := fmt.Sprintf("%x", outputID) //Might not need

	actionFile := filepath.Join(fss.baseDir, fmt.Sprintf("a-%s", actionIDHex))
	diskPath := filepath.Join(fss.baseDir, fmt.Sprintf("o-%s", actionIDHex))
	absPath, _ := filepath.Abs(diskPath) //Always return absolute path

	// Write metadata
	now := time.Now()
	ie, err := json.Marshal(indexEntry{
		Version:  1,
		OutputID: outputID,
		Size:     size,
		Time:     &now,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal index entry: %w", err)
	}

	if err = os.WriteFile(actionFile, ie, 0644); err != nil {
		return "", fmt.Errorf("failed to write metafile: %w", err)
	}

	// Write the data
	if size > 0 {
		if err := os.WriteFile(diskPath, data, 0644); err != nil {
			return "", fmt.Errorf("failed to write cache file: %w", err)
		}
	} else {
		//bodysize == 0, touch the file
		zf, err := os.OpenFile(diskPath, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return "", fmt.Errorf("failed to create empty file: %w", err)
		}
		zf.Close()
	}

	return absPath, nil
}

func (fss *FileSystemStorage) Get(actionID []byte) (outputID []byte, size int64, modTime time.Time, diskPath string, found bool, err error) {
	actionIDHex := fmt.Sprintf("%x", actionID)
	actionFile := filepath.Join(fss.baseDir, fmt.Sprintf("a-%s", actionIDHex))

	// Read metadata
	af, err := os.ReadFile(actionFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, 0, time.Time{}, "", false, nil // Not found
		}
		return nil, 0, time.Time{}, "", false, fmt.Errorf("failed to read metafile: %w", err)
	}

	var ie indexEntry
	if err := json.Unmarshal(af, &ie); err != nil {
		return nil, 0, time.Time{}, "", false, fmt.Errorf("failed to unmarshal index entry: %w", err)
	}

	objectFile := filepath.Join(fss.baseDir, fmt.Sprintf("o-%s", actionIDHex))
	info, err := os.Stat(objectFile)
	if os.IsNotExist(err) {
		return nil, 0, time.Time{}, "", false, nil // Not found
	}
	if err != nil {
		return nil, 0, time.Time{}, "", false, err
	}

	diskPath, _ = filepath.Abs(objectFile)
	return ie.OutputID, info.Size(), info.ModTime(), diskPath, true, nil
}
