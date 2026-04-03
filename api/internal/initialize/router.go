package initialize

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"

	"github.com/ndquang191/Anochat/api/internal/handler"
	"github.com/ndquang191/Anochat/api/internal/middleware"
	"github.com/ndquang191/Anochat/api/internal/repository"
	"github.com/ndquang191/Anochat/api/internal/service"
	"github.com/ndquang191/Anochat/api/internal/ws"
	"github.com/ndquang191/Anochat/api/pkg/cache"
	"github.com/ndquang191/Anochat/api/pkg/config"
	"github.com/ndquang191/Anochat/api/pkg/database"
	"github.com/ndquang191/Anochat/api/pkg/metrics"
)

type Server struct {
	Router       *gin.Engine
	QueueService *service.QueueService
}

func Router(cfg *config.Config, db *gorm.DB, redisClient *redis.Client) *Server {
	// Repositories
	userRepo := repository.NewUserRepository(db)
	profileRepo := repository.NewProfileRepository(db)
	roomRepo := repository.NewRoomRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	bannedWordRepo := repository.NewBannedWordRepository(db)
	reportRepo := repository.NewReportRepository(db)

	if count, err := userRepo.Count(context.Background()); err == nil {
		metrics.TotalUsers.Set(float64(count))
	}

	// OAuth
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.OAuth.GoogleClientID,
		ClientSecret: cfg.OAuth.GoogleClientSecret,
		RedirectURL:  cfg.OAuth.RedirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}

	// Services
	userService := service.NewUserService(userRepo, profileRepo)
	roomService := service.NewRoomService(roomRepo, messageRepo)
	messageService := service.NewMessageService(messageRepo)
	authService := service.NewAuthService(userService, oauthConfig, cfg.OAuth.JWTSecret, redisClient, cfg.OAuth.AccessTokenExpiry, cfg.OAuth.RefreshTokenExpiry)
	queueService := service.NewQueueService(roomService, roomRepo)
	moderationService := service.NewModerationService(bannedWordRepo, reportRepo, userRepo)
	if err := moderationService.LoadWords(context.Background()); err != nil {
		slog.Warn("Failed to load banned words at startup", "error", err)
	}

	// WebSocket hub
	wsHub := ws.NewHub(queueService, messageService, roomService, moderationService, redisClient)
	go wsHub.Run()
	slog.Info("WebSocket hub started")

	queueService.SetMatchNotifier(wsHub)

	// Handlers
	authHandler := handler.NewAuthHandler(authService, oauthConfig, cfg)
	userHandler := handler.NewUserHandler(userService, roomService, queueService, roomRepo, messageRepo, cfg)
	queueHandler := handler.NewQueueHandler(queueService, cfg)
	wsHandler := handler.NewWebSocketHandler(wsHub, authService, cfg)
	moderationHandler := handler.NewModerationHandler(moderationService, messageRepo)
	devHandler := handler.NewDevHandler(db)

	authMiddleware := middleware.AuthMiddleware(authService, userRepo, cfg)

	// Router
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	router.Use(middleware.CORSMiddleware(cfg.ClientURL))
	router.Use(middleware.RateLimitMiddleware(redisClient, cfg.Security.RateLimit, cfg.Security.RateLimit*2))
	slog.Info("Rate limiting enabled", "rate", cfg.Security.RateLimit, "burst", cfg.Security.RateLimit*2)

	// Routes
	router.GET("/auth/google", authHandler.GoogleLogin)
	router.GET("/auth/callback", authHandler.GoogleCallback)
	router.POST("/auth/refresh", authHandler.RefreshToken)
	router.POST("/auth/logout", authHandler.Logout)

	protected := router.Group("/")
	protected.Use(authMiddleware)
	{
		protected.GET("/user/state", userHandler.GetUserState)
		protected.PUT("/profile", userHandler.UpdateProfile)
		protected.POST("/room/leave", userHandler.LeaveCurrentRoom)

		protected.POST("/queue/join", queueHandler.JoinQueue)
		protected.POST("/queue/leave", queueHandler.LeaveQueue)

		protected.GET("/ws", wsHandler.HandleWebSocket)

		protected.POST("/report", moderationHandler.CreateReport)

		admin := protected.Group("/admin")
		{
			admin.GET("/words", moderationHandler.ListWords)
			admin.POST("/words", moderationHandler.AddWord)
			admin.PUT("/words/:id", moderationHandler.UpdateWord)
			admin.DELETE("/words/:id", moderationHandler.DeleteWord)
			admin.GET("/reports", moderationHandler.ListReports)
			admin.POST("/users/:id/ban", moderationHandler.BanUser)
			admin.POST("/users/:id/unban", moderationHandler.UnbanUser)
			admin.GET("/users/banned", moderationHandler.ListBannedUsers)
			admin.GET("/rooms/:id/messages", moderationHandler.ListRoomMessages)
		}
	}

	if !cfg.IsProduction() {
		router.POST("/dev/reset", devHandler.ResetDB)
		slog.Info("Dev routes enabled: POST /dev/reset")
	}

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/healthz", healthHandler())

	return &Server{
		Router:       router,
		QueueService: queueService,
	}
}

func healthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
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
	}
}
