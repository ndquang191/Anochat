package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/ndquang191/Anochat/api/internal/service"
	"github.com/ndquang191/Anochat/api/internal/ws"
	"github.com/ndquang191/Anochat/api/pkg/config"
)

type WebSocketHandler struct {
	hub         *ws.Hub
	authService *service.AuthService
	config      *config.Config
}

func NewWebSocketHandler(hub *ws.Hub, authService *service.AuthService, cfg *config.Config) *WebSocketHandler {
	return &WebSocketHandler{hub: hub, authService: authService, config: cfg}
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	userID := getUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			return origin == h.config.ClientURL
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("Failed to upgrade connection", "error", err, "user_id", userID)
		return
	}

	client := ws.NewClient(h.hub, conn, userID)
	h.hub.Register() <- client

	go client.WritePump()
	go client.ReadPump()

	slog.Info("WebSocket connection established", "user_id", userID, "client_id", client.ID)
}
