package main

import (
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ndquang191/Anochat/api/internal/handler"
	"github.com/ndquang191/Anochat/api/internal/middleware"
	"github.com/ndquang191/Anochat/api/internal/service"
	"github.com/ndquang191/Anochat/api/pkg/config"
	"github.com/ndquang191/Anochat/api/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.Load()
	slog.Info("Configuration loaded successfully")

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

	// Setup Gin
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Setup OAuth configuration
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.OAuth.GoogleClientID,
		ClientSecret: cfg.OAuth.GoogleClientSecret,
		RedirectURL:  cfg.OAuth.RedirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}

	// Initialize services
	db := database.GetDB()
	userService := service.NewUserService(db)
	roomService := service.NewRoomService(db)
	authService := service.NewAuthService(userService, oauthConfig, cfg.OAuth.JWTSecret)
	queueService := service.NewQueueService(db, roomService, userService, cfg)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, oauthConfig)
	userHandler := handler.NewUserHandler(authService, userService)
	queueHandler := handler.NewQueueHandler(queueService)

	// Initialize middleware
	authMiddleware := middleware.AuthMiddleware(authService)

	// Setup routes
	setupRoutes(router, authHandler, userHandler, queueHandler, authMiddleware)

	// Health check endpoint
	router.GET("/healthz", func(c *gin.Context) {
		// Check database health
		if err := database.HealthCheck(); err != nil {
			c.JSON(503, gin.H{
				"status":  "error",
				"message": "Database connection failed",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status":   "ok",
			"message":  "Anonymous Chat API is running",
			"database": "connected",
		})
	})

	// Graceful shutdown
	go func() {
		slog.Info("Starting server", "port", cfg.Port)
		if err := router.Run(":" + cfg.Port); err != nil {
			slog.Error("Failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")

	// Stop queue service gracefully
	queueService.Stop()
	slog.Info("Queue service stopped")
}

// setupRoutes configures all API routes
func setupRoutes(router *gin.Engine, authHandler *handler.AuthHandler, userHandler *handler.UserHandler, queueHandler *handler.QueueHandler, authMiddleware gin.HandlerFunc) {
	// Auth routes (no middleware required)
	router.GET("/auth/google", authHandler.GoogleLogin)
	router.GET("/auth/callback", authHandler.GoogleCallback)
	router.POST("/auth/logout", authHandler.Logout) // New logout endpoint

	// Protected routes (require JWT)
	protected := router.Group("/")
	protected.Use(authMiddleware)
	{
		// User state endpoint (main endpoint)
		protected.GET("/user/state", userHandler.GetUserState)
		// Profile update endpoint
		protected.PUT("/profile", userHandler.UpdateProfile)

		// Queue endpoints
		protected.POST("/queue/join", queueHandler.JoinQueue)
		protected.POST("/queue/leave", queueHandler.LeaveQueue)
		protected.GET("/queue/status", queueHandler.GetQueueStatus)
		protected.GET("/queue/stats", queueHandler.GetQueueStats)
		protected.GET("/queue/match-stats", queueHandler.GetMatchStats)
	}
}
