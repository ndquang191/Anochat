package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/dto"
	"github.com/ndquang191/Anochat/api/internal/service"
	"github.com/ndquang191/Anochat/api/pkg/apperr"
)

type PushHandler struct {
	service      *service.PushService
	queueService *service.QueueService
}

func NewPushHandler(pushService *service.PushService, queueService *service.QueueService) *PushHandler {
	return &PushHandler{service: pushService, queueService: queueService}
}

func (h *PushHandler) Config(c *gin.Context) {
	dto.OK(c, gin.H{"enabled": h.service.Enabled(), "public_key": h.service.PublicKey()})
}

func (h *PushHandler) Subscribe(c *gin.Context) {
	var request struct {
		Endpoint string `json:"endpoint" binding:"required"`
		Keys     struct {
			P256DH string `json:"p256dh"`
			Auth   string `json:"auth"`
		} `json:"keys" binding:"required"`
		Locale string `json:"locale" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		dto.FailErr(c, apperr.ErrInvalidBody)
		return
	}
	item, err := h.service.Upsert(c.Request.Context(), getUserID(c), service.PushSubscriptionInput{
		Endpoint: request.Endpoint, P256DH: request.Keys.P256DH, Auth: request.Keys.Auth, Locale: request.Locale,
	})
	if err != nil {
		dto.Fail(c, http.StatusBadRequest, "Push subscription không hợp lệ")
		return
	}
	dto.OK(c, gin.H{"id": item.ID})
}

func (h *PushHandler) Delete(c *gin.Context) {
	userID := getUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		dto.FailErr(c, apperr.ErrInvalidID)
		return
	}
	if err := h.service.Delete(c.Request.Context(), id, userID); err != nil {
		dto.FailErr(c, err)
		return
	}
	if h.queueService.BackgroundSubscription(c.Request.Context(), userID) == id {
		_ = h.queueService.LeaveQueue(c.Request.Context(), userID)
		h.queueService.ClearBackgroundSubscriptions(c.Request.Context(), userID)
	}
	dto.OK(c, gin.H{"deleted": true})
}
