package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/dto"
	"github.com/ndquang191/Anochat/api/internal/service"
	"github.com/ndquang191/Anochat/api/pkg/config"
)

type QueueHandler struct {
	queueService *service.QueueService
	config       *config.Config
}

func NewQueueHandler(queueService *service.QueueService, cfg *config.Config) *QueueHandler {
	return &QueueHandler{queueService: queueService, config: cfg}
}

func (h *QueueHandler) JoinQueue(c *gin.Context) {
	userID := getUserID(c)
	if userID == uuid.Nil {
		signOutAndRedirect(c, h.config)
		return
	}

	if err := h.queueService.JoinQueue(c.Request.Context(), userID); err != nil {
		dto.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	dto.OKWithMessage(c, "Successfully joined queue", nil)
}

func (h *QueueHandler) LeaveQueue(c *gin.Context) {
	userID := getUserID(c)
	if userID == uuid.Nil {
		signOutAndRedirect(c, h.config)
		return
	}

	if err := h.queueService.LeaveQueue(c.Request.Context(), userID); err != nil {
		dto.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	dto.OKWithMessage(c, "Successfully left queue", nil)
}
