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

	"github.com/vellalasantosh/wound_iq_api_claude/internal/config"
	"github.com/vellalasantosh/wound_iq_api_claude/internal/db"
	"github.com/vellalasantosh/wound_iq_api_claude/internal/handlers"
	"github.com/vellalasantosh/wound_iq_api_claude/internal/repository"
	"github.com/vellalasantosh/wound_iq_api_claude/internal/router"
	"github.com/vellalasantosh/wound_iq_api_claude/internal/routes"
	"github.com/vellalasantosh/wound_iq_api_claude/internal/service"
	"github.com/vellalasantosh/wound_iq_api_claude/internal/utils"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	log.Println("DEBUG — JWT_SECRET from env:", os.Getenv("JWT_SECRET"))

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// ============ NEW: Set JWT Secret ============
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}
	utils.SetJWTSecret(jwtSecret)
	// =============================================

	// Initialize database connection
	database, err := db.NewPostgresDB(cfg.DBDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	log.Println("Successfully connected to PostgreSQL database")

	// ============ NEW: Initialize Auth Components ============
	authRepo := repository.NewAuthRepository(database.DB)
	authService := service.NewAuthService(authRepo)
	authHandler := handlers.NewAuthHandler(authService)
	// =========================================================

	// Initialize router with database connection
	r := router.SetupRouter(database)

	// ============ NEW: Setup Auth Routes ============
	v1 := r.Group("/api/v1")
	routes.SetupAuthRoutes(v1, authHandler)
	// ================================================

	// Configure server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited successfully")
}
