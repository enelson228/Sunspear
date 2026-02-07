package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sunspear/api"
	"sunspear/config"
	"sunspear/services"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize configuration
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Initialize database
	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize Docker service
	dockerService, err := services.NewDockerService()
	if err != nil {
		log.Fatalf("Failed to initialize Docker service: %v", err)
	}
	defer dockerService.Close()

	// Initialize monitoring service
	monitorService := services.NewMonitoringService()
	monitorService.Start()
	defer monitorService.Stop()

	// Initialize marketplace service
	marketplaceService := services.NewMarketplaceService(db)
	if err := marketplaceService.LoadApps(); err != nil {
		log.Printf("Warning: Failed to load marketplace apps: %v", err)
	}

	// Initialize compose service
	composeService := services.NewComposeService(db, dockerService)

	// Create router
	router := api.NewRouter(cfg, db, dockerService, monitorService, marketplaceService, composeService)

	// Configure server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting Sunspear backend on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
