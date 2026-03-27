package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("GET /api/v1/health", handleHealth)

	// Serve static frontend files
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "../../frontend/dist"
	}
	absStatic, _ := filepath.Abs(staticDir)

	// Serve static files, fall back to index.html for SPA routing
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(absStatic, r.URL.Path)

		// Check if the file exists
		if _, err := os.Stat(path); err == nil {
			http.ServeFile(w, r, path)
			return
		}

		// Fall back to index.html for SPA routing
		http.ServeFile(w, r, filepath.Join(absStatic, "index.html"))
	})

	log.Printf("Starting server on :%s", port)
	log.Printf("Serving static files from %s", absStatic)

	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
