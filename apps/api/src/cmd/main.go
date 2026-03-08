package main

import (
	"api/src/routes"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"utils/db"
	"utils/env"
	"utils/logger"
)

func init() {
	logger.Init()
}

func main() {
	// Get port from environment variable, default to 8080
	port := env.GetStringOrDefault("PORT", "8080")

	// Initialize Database
	dbAuth := env.GetStringOrDefault("DB_URI", "")
	if dbAuth == "" {
		logger.Error("DB_URI environment variable is not set")
		os.Exit(1)
	}

	dbConn, err := db.Connect(dbAuth)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := dbConn.Close(); err != nil {
			logger.Error("Failed to close database connection", "error", err)
		}
	}()

	// Router
	r := routes.NewRouter(dbConn)

	// Configure server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("Server exited gracefully")
}
