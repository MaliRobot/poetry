package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"poetry/db"
	"poetry/worker"
	"strconv"
	"syscall"
	"time"
)

func main() {
	// Get configuration from environment variables or use defaults
	bufferSize := getEnvInt("WORKER_BUFFER_SIZE", 10)
	maxWorkers := getEnvInt("WORKER_MAX_WORKERS", 3)
	port := getEnvString("WORKER_PORT", "8081")

	log.Printf("Starting worker service on port %s with buffer size %d and %d workers", port, bufferSize, maxWorkers)

	// Connect to MongoDB
	mongoDBConnection, err := db.NewMongoDBConnection()
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoDBConnection.Disconnect()

	// Create and start worker
	w := worker.NewWorker(mongoDBConnection, bufferSize, maxWorkers)
	w.Start()

	// Set up HTTP server for receiving jobs
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/status", func(rw http.ResponseWriter, r *http.Request) {
		statusHandler(rw, r, w)
	})
	http.HandleFunc("/jobs", func(rw http.ResponseWriter, r *http.Request) {
		jobHandler(rw, r, w)
	})

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Worker HTTP server listening on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-quit
	log.Println("Shutting down worker service...")

	// Stop worker gracefully
	w.Stop()

	log.Println("Worker service stopped")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func statusHandler(w http.ResponseWriter, r *http.Request, worker *worker.Worker) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"queue_size": worker.GetQueueSize(),
		"timestamp":  time.Now().Unix(),
	})
}

func jobHandler(w http.ResponseWriter, r *http.Request, worker *worker.Worker) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var poems []db.Poem
	if err := json.NewDecoder(r.Body).Decode(&poems); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if len(poems) == 0 {
		http.Error(w, "No poems provided", http.StatusBadRequest)
		return
	}

	err := worker.AddJob(poems)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Job accepted",
		"poem_count": len(poems),
		"queue_size": worker.GetQueueSize(),
	})
}

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
