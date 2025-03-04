package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/bigwhite/go-cache-prog/cache"
	"github.com/bigwhite/go-cache-prog/protocol"
	"github.com/bigwhite/go-cache-prog/storage/filesystem"
)

func main() {
	log.SetOutput(os.Stderr)

	verbose := false
	for _, arg := range os.Args {
		if arg == "--verbose" {
			verbose = true
			break
		}
	}

	cacheDir := os.Getenv("GOCACHEPROG_DIR")
	if cacheDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get user home directory: %v", err)
		}
		cacheDir = filepath.Join(homeDir, ".gocacheprog")
	}

	if verbose {
		log.Printf("Using cache directory: %s", cacheDir)
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Fatalf("Failed to create cache directory: %v", err)
	}

	// Initialize storage and cache.
	store, err := filesystem.NewFileSystemStorage(cacheDir, verbose)
	if err != nil {
		log.Fatalf("Failed to initialize filesystem storage: %v", err)
	}
	cacheInstance := cache.NewCache(store)

	// Send initial capabilities response.
	initResp := protocol.Response{
		KnownCommands: []protocol.Cmd{protocol.CmdPut, protocol.CmdGet, protocol.CmdClose},
	}
	if err := json.NewEncoder(os.Stdout).Encode(initResp); err != nil {
		log.Fatalf("Failed to send initial response: %v", err)
	}

	reader := bufio.NewReader(os.Stdin)
	requestHandler := protocol.NewRequestHandler(reader, os.Stdout, cacheInstance, verbose)

	if err := requestHandler.HandleRequests(); err != nil {
		log.Fatalf("Error handling requests: %v", err)
	}

	if verbose {
		gets, getMiss := requestHandler.Stats()
		log.Printf("Gets: %d, GetMiss: %d\n", gets, getMiss)
	}
}
