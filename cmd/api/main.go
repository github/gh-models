// Package main provides the HTTP API server for BrainSait AI Platform
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/github/gh-models/internal/api"
	"github.com/github/gh-models/internal/billing"
)

func main() {
	// Load configuration
	cfg := api.LoadConfig()

	// Initialize billing service
	billingService := billing.NewService(cfg.StripeSecretKey, cfg.StripeWebhookSecret)

	// Initialize API server
	server := api.NewServer(cfg, billingService)

	// Setup routes
	router := server.SetupRoutes()

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 BrainSait API Server starting on port %s", cfg.Port)
		log.Printf("📍 Environment: %s", cfg.Environment)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("⏳ Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server stopped gracefully")
}
