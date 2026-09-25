package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/dto"
	"github.com/ndquang191/Anochat/api/internal/service"
	"github.com/ndquang191/Anochat/api/pkg/config"
)

type QueueHandler struct {
	queueService *service.QueueService
	pushService  *service.PushService
	config       *config.Config
}

func NewQueueHandler(queueService *service.QueueService, pushService *service.PushService, cfg *config.Config) *QueueHandler {
	return &QueueHandler{queueService: queueService, pushService: pushService, config: cfg}
}

func (h *QueueHandler) JoinQueue(c *gin.Context) {
	userID := getUserID(c)
	if userID == uuid.Nil {
		signOutAndRedirect(c, h.config)
		return
	}

	var request struct {
		PushSubscriptionID string `json:"push_subscription_id"`
	}
	_ = c.ShouldBindJSON(&request)
	subscriptionID := uuid.Nil
	if parsed, err := uuid.Parse(request.PushSubscriptionID); err == nil && h.pushService.IsOwned(c.Request.Context(), parsed, userID) {
		subscriptionID = parsed
	}

	if err := h.queueService.JoinQueueWithSubscription(c.Request.Context(), userID, subscriptionID); err != nil {
		dto.FailErr(c, err)
		return
	}

	dto.OKWithMessage(c, "Successfully joined queue", nil)
}

func (h *QueueHandler) Heartbeat(c *gin.Context) {
	if err := h.queueService.Heartbeat(c.Request.Context(), getUserID(c)); err != nil {
		dto.FailErr(c, err)
		return
	}
	dto.OKWithMessage(c, "Queue lease renewed", nil)
}

func (h *QueueHandler) LeaveQueue(c *gin.Context) {
	userID := getUserID(c)
	if userID == uuid.Nil {
		signOutAndRedirect(c, h.config)
		return
	}

	if err := h.queueService.LeaveQueue(c.Request.Context(), userID); err != nil {
		dto.FailErr(c, err)
		return
	}

	dto.OKWithMessage(c, "Successfully left queue", nil)
}
