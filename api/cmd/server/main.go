package main

import (
	"github.com/ndquang191/Anochat/api/pkg/database"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Initialize database
	if err := database.InitDatabase(); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer database.CloseDatabase()

	// Run database migrations
	if err := database.RunMigrations(); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Setup Gin
	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Health check endpoint
	router.GET("/healthz", func(c *gin.Context) {
		// Check database health
		if err := database.HealthCheck(); err != nil {
			c.JSON(503, gin.H{
				"status": "error",
				"message": "Database connection failed",
				"error": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "ok",
			"message": "Anonymous Chat API is running",
			"database": "connected",
		})
	})

	// Graceful shutdown
	go func() {
		slog.Info("Starting server", "port", port)
		if err := router.Run(":" + port); err != nil {
			slog.Error("Failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")
}