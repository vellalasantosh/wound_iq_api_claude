package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/vellalasantosh/wound_iq_api_claude/internal/config"
	"github.com/vellalasantosh/wound_iq_api_claude/internal/db"
	"github.com/vellalasantosh/wound_iq_api_claude/internal/handlers"
	"github.com/vellalasantosh/wound_iq_api_claude/internal/repository"
	"github.com/vellalasantosh/wound_iq_api_claude/internal/router"
	"github.com/vellalasantosh/wound_iq_api_claude/internal/routes"
	"github.com/vellalasantosh/wound_iq_api_claude/internal/service"
	"github.com/vellalasantosh/wound_iq_api_claude/internal/utils"
)

func main() {

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment vars")
	}

	// Load base config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Load JWT secret
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}
	utils.SetJWTSecret(jwtSecret)

	// Initialize database
	database, err := db.NewPostgresDB(cfg.DBDSN)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer database.Close()

	log.Println("Successfully connected to PostgreSQL")

	// Initialize Auth components
	authRepo := repository.NewAuthRepository(database.DB)
	authService := service.NewAuthService(authRepo)
	authHandler := handlers.NewAuthHandler(authService)

	// Initialize Router
	r := router.SetupRouter(database)

	// Attach Auth Routes under unified /api/v1
	v1 := r.Group("/api/v1")
	routes.SetupAuthRoutes(v1, authHandler)

	// Configure HTTP Server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		log.Printf("Server starting on port %s...", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Forced shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
