package initialize

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

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
	Hub          *ws.Hub
	background   sync.WaitGroup
}

func (s *Server) Start(ctx context.Context) {
	s.background.Add(2)
	go func() {
		defer s.background.Done()
		s.Hub.Run(ctx)
	}()
	go func() {
		defer s.background.Done()
		s.QueueService.Run(ctx)
	}()
}

func (s *Server) Wait(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		s.background.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func Router(cfg *config.Config, db *gorm.DB, redisClient *redis.Client) *Server {
	userRepo := repository.NewUserRepository(db)
	profileRepo := repository.NewProfileRepository(db)
	roomRepo := repository.NewRoomRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	bannedWordRepo := repository.NewBannedWordRepository(db)
	reportRepo := repository.NewReportRepository(db)
	adminStatsRepo := repository.NewAdminStatsRepository(db)

	if count, err := userRepo.Count(context.Background()); err == nil {
		metrics.TotalUsers.Set(float64(count))
	}

	oauthConfig := &oauth2.Config{
		ClientID:     cfg.OAuth.GoogleClientID,
		ClientSecret: cfg.OAuth.GoogleClientSecret,
		RedirectURL:  cfg.OAuth.RedirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}

	userService := service.NewUserService(userRepo, profileRepo)
	roomService := service.NewRoomService(roomRepo, messageRepo)
	messageService := service.NewMessageService(messageRepo)
	authService := service.NewAuthService(userService, oauthConfig, cfg.OAuth.JWTSecret, redisClient, cfg.OAuth.AccessTokenExpiry, cfg.OAuth.RefreshTokenExpiry)
	queueService := service.NewQueueService(roomService, roomRepo, redisClient)
	moderationService := service.NewModerationService(bannedWordRepo, reportRepo, userRepo, roomRepo)
	adminStatsService := service.NewAdminStatsService(adminStatsRepo, redisClient)
	if err := moderationService.LoadWords(context.Background()); err != nil {
		slog.Warn("Failed to load banned words at startup", "error", err)
	}

	wsHub := ws.NewHub(
		queueService,
		messageService,
		roomService,
		moderationService,
		redisClient,
		cfg.Chat.MessageRateLimit,
		cfg.Chat.MaxMessageLength,
	)
	queueService.SetMatchNotifier(wsHub)

	authHandler := handler.NewAuthHandler(authService, oauthConfig, cfg)
	userHandler := handler.NewUserHandler(userService, roomService, messageService, queueService, roomRepo, cfg)
	queueHandler := handler.NewQueueHandler(queueService, cfg)
	wsHandler := handler.NewWebSocketHandler(wsHub, authService, cfg)
	moderationHandler := handler.NewModerationHandler(moderationService)
	adminStatsHandler := handler.NewAdminStatsHandler(adminStatsService)

	authMiddleware := middleware.AuthMiddleware(authService, userRepo, cfg)

	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	router.Use(middleware.CORSMiddleware(cfg.ClientURL))
	router.Use(middleware.CSRFMiddleware(cfg.ClientURL))
	router.Use(middleware.RateLimitMiddleware(redisClient, cfg.Security.RateLimit, cfg.Security.RateLimit*2))
	slog.Info("Rate limiting enabled", "rate", cfg.Security.RateLimit, "burst", cfg.Security.RateLimit*2)

	router.GET("/auth/google", authHandler.GoogleLogin)
	router.GET("/auth/callback", authHandler.GoogleCallback)
	if cfg.DevAuthEnabled {
		router.POST("/auth/dev", authHandler.DevLogin)
		slog.Warn("Development login is enabled")
	}
	router.POST("/auth/refresh", authHandler.RefreshToken)
	router.POST("/auth/logout", authHandler.Logout)

	protected := router.Group("/")
	protected.Use(authMiddleware)
	{
		protected.GET("/user/state", userHandler.GetUserState)
		protected.POST("/user/ban-review", moderationHandler.RequestBanReview)

		active := protected.Group("/")
		active.Use(middleware.RequireActive())
		{
			active.GET("/metrics", middleware.RequireAdmin(), gin.WrapH(promhttp.Handler()))
			active.PUT("/profile", userHandler.UpdateProfile)
			active.POST("/room/leave", userHandler.LeaveCurrentRoom)
			active.GET("/rooms/:id/messages", userHandler.GetRoomMessages)

			active.POST("/queue/join", queueHandler.JoinQueue)
			active.POST("/queue/leave", queueHandler.LeaveQueue)

			active.GET("/ws", wsHandler.HandleWebSocket)

			active.POST("/report", moderationHandler.CreateReport)

			admin := active.Group("/admin")
			{
				admin.GET("/overview", adminStatsHandler.GetOverview)
				admin.GET("/words", moderationHandler.ListWords)
				admin.POST("/words", moderationHandler.AddWord)
				admin.PUT("/words/:id", moderationHandler.UpdateWord)
				admin.DELETE("/words/:id", moderationHandler.DeleteWord)
				admin.GET("/reports", moderationHandler.ListReports)
				admin.POST("/users/:id/ban", moderationHandler.BanUser)
				admin.POST("/users/:id/unban", moderationHandler.UnbanUser)
				admin.GET("/users/banned", moderationHandler.ListBannedUsers)
				admin.GET("/reports/:id/messages", moderationHandler.ListReportMessages)
			}
		}
	}

	router.GET("/healthz", healthHandler())

	return &Server{
		Router:       router,
		QueueService: queueService,
		Hub:          wsHub,
	}
}

func healthHandler() gin.HandlerFunc {
	return healthHandlerWithChecks(database.HealthCheck, cache.HealthCheck)
}

func healthHandlerWithChecks(databaseCheck, redisCheck func() error) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := gin.H{
			"database": "connected",
			"redis":    "connected",
		}
		healthy := true

		if err := databaseCheck(); err != nil {
			healthy = false
			status["database"] = "unavailable"
			slog.Warn("Database health check failed", "error", err)
		}
		if err := redisCheck(); err != nil {
			healthy = false
			status["redis"] = "unavailable"
			slog.Warn("Redis health check failed", "error", err)
		}

		if !healthy {
			status["status"] = "error"
			c.JSON(http.StatusServiceUnavailable, status)
			return
		}

		status["status"] = "ok"
		c.JSON(http.StatusOK, status)
	}
}
