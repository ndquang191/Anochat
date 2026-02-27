package main

import (
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ndquang191/Anochat/api/internal/handler"
	"github.com/ndquang191/Anochat/api/internal/middleware"
	"github.com/ndquang191/Anochat/api/internal/repository"
	"github.com/ndquang191/Anochat/api/internal/service"
	"github.com/ndquang191/Anochat/api/internal/ws"
	"github.com/ndquang191/Anochat/api/pkg/cache"
	"github.com/ndquang191/Anochat/api/pkg/config"
	"github.com/ndquang191/Anochat/api/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	cfg := config.Load()

	var zapLogger *zap.Logger
	if cfg.IsProduction() {
		zapLogger, _ = zap.NewProduction()
	} else {
		zapLogger, _ = zap.NewDevelopment()
	}
	defer zapLogger.Sync()
	slog.SetDefault(slog.New(zapslog.NewHandler(zapLogger.Core(), nil)))
	slog.Info("Configuration loaded successfully")

	if err := database.InitDatabase(cfg); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer database.CloseDatabase()

	if err := cache.InitRedis(cfg); err != nil {
		slog.Error("Failed to initialize Redis", "error", err)
		os.Exit(1)
	}
	defer cache.CloseRedis()

	if err := database.RunMigrations(); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	router.Use(middleware.CORSMiddleware(cfg.ClientURL))
	router.Use(middleware.RateLimitMiddleware(cache.Client, cfg.Security.RateLimit, cfg.Security.RateLimit*2))
	slog.Info("Rate limiting enabled", "rate", cfg.Security.RateLimit, "burst", cfg.Security.RateLimit*2)

	oauthConfig := &oauth2.Config{
		ClientID:     cfg.OAuth.GoogleClientID,
		ClientSecret: cfg.OAuth.GoogleClientSecret,
		RedirectURL:  cfg.OAuth.RedirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}

	db := database.GetDB()

	userRepo := repository.NewUserRepository(db)
	profileRepo := repository.NewProfileRepository(db)
	roomRepo := repository.NewRoomRepository(db)
	messageRepo := repository.NewMessageRepository(db)

	userService := service.NewUserService(userRepo, profileRepo)
	roomService := service.NewRoomService(roomRepo, messageRepo)
	messageService := service.NewMessageService(messageRepo, roomRepo)
	authService := service.NewAuthService(userService, oauthConfig, cfg.OAuth.JWTSecret, roomRepo, messageRepo)
	queueService := service.NewQueueService(roomService, userService, roomRepo, cfg)

	wsHub := ws.NewHub(queueService, messageService, roomService, cache.Client)
	go wsHub.Run()
	slog.Info("WebSocket hub started")

	queueService.SetMatchNotifier(wsHub)

	authHandler := handler.NewAuthHandler(authService, oauthConfig, cfg)
	userHandler := handler.NewUserHandler(userService, roomService, roomRepo, messageRepo, cfg)
	queueHandler := handler.NewQueueHandler(queueService, cfg)
	wsHandler := handler.NewWebSocketHandler(wsHub, authService, cfg)

	authMiddleware := middleware.AuthMiddleware(authService, cfg)

	setupRoutes(router, authHandler, userHandler, queueHandler, wsHandler, authMiddleware)

	router.GET("/healthz", func(c *gin.Context) {
		if err := database.HealthCheck(); err != nil {
			c.JSON(503, gin.H{
				"status":  "error",
				"message": "Database connection failed",
				"error":   err.Error(),
			})
			return
		}
		redisStatus := "connected"
		if err := cache.HealthCheck(); err != nil {
			redisStatus = "error: " + err.Error()
		}
		c.JSON(200, gin.H{
			"status":   "ok",
			"message":  "Anonymous Chat API is running",
			"database": "connected",
			"redis":    redisStatus,
		})
	})

	go func() {
		slog.Info("Starting server", "port", cfg.Port)
		if err := router.Run(":" + cfg.Port); err != nil {
			slog.Error("Failed to start server", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")

	queueService.Stop()
	slog.Info("Queue service stopped")
}

func setupRoutes(router *gin.Engine, authHandler *handler.AuthHandler, userHandler *handler.UserHandler, queueHandler *handler.QueueHandler, wsHandler *handler.WebSocketHandler, authMiddleware gin.HandlerFunc) {
	router.GET("/auth/google", authHandler.GoogleLogin)
	router.GET("/auth/callback", authHandler.GoogleCallback)
	router.POST("/auth/logout", authHandler.Logout)

	protected := router.Group("/")
	protected.Use(authMiddleware)
	{
		protected.GET("/user/state", userHandler.GetUserState)
		protected.PUT("/profile", userHandler.UpdateProfile)
		protected.POST("/room/leave", userHandler.LeaveCurrentRoom)

		protected.POST("/queue/join", queueHandler.JoinQueue)
		protected.POST("/queue/leave", queueHandler.LeaveQueue)
		protected.GET("/queue/status", queueHandler.GetQueueStatus)
		protected.GET("/queue/stats", queueHandler.GetQueueStats)
		protected.GET("/queue/match-stats", queueHandler.GetMatchStats)

		protected.GET("/ws", wsHandler.HandleWebSocket)
	}
}
