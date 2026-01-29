package main

import (
	"fmt"
	"net/http"

	"example/data-access/internal/logger"
	"example/data-access/internal/server"

	"github.com/joho/godotenv"
)

func main() {
	// Initialize logger
	logger.InitLoggerDev()
	defer logger.Sync()

	logger.Log.Info("Starting WebSocket API Server")

	// Load .env file
	if err := godotenv.Load(); err != nil {
		logger.Log.Warnw("No .env file found, using existing environment variables", "error", err)
	}

	// Initialize database
	if err := server.InitDatabase(); err != nil {
		logger.Log.Fatalw("Failed to initialize database", "error", err)
	}
	defer server.CloseDatabase()

	// Set up HTTP routes
	http.HandleFunc("/ws", server.HandleWebSocket)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "WebSocket API Server\nConnect to ws://localhost:8080/ws\n")
	})

	// Start server
	logger.Log.Infow("WebSocket server starting", "port", "8080", "endpoint", "ws://localhost:8080/ws")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.Log.Fatalw("Server error", "error", err)
	}
}
