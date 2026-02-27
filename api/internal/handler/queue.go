package handler

import (
	"net/http"
	"time"

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

	var req dto.JoinQueueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Fail(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	_, err := h.queueService.JoinQueue(c.Request.Context(), userID, req.Category)
	if err != nil {
		dto.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	status, _ := h.queueService.GetQueueStatus(c.Request.Context(), userID)
	if status != nil && status.IsInQueue {
		dto.OKWithMessage(c, "Successfully joined queue", dto.QueueStatusResponse{
			IsInQueue: status.IsInQueue,
			Position:  status.Position,
			Category:  status.Category,
			JoinedAt:  status.JoinedAt.Format(time.RFC3339),
			ExpiresAt: status.ExpiresAt.Format(time.RFC3339),
		})
	} else {
		dto.OKWithMessage(c, "Successfully joined queue", nil)
	}
}

func (h *QueueHandler) LeaveQueue(c *gin.Context) {
	userID := getUserID(c)
	if userID == uuid.Nil {
		signOutAndRedirect(c, h.config)
		return
	}

	err := h.queueService.LeaveQueue(c.Request.Context(), userID)
	if err != nil {
		dto.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	dto.OKWithMessage(c, "Successfully left queue", nil)
}

func (h *QueueHandler) GetQueueStatus(c *gin.Context) {
	userID := getUserID(c)
	if userID == uuid.Nil {
		signOutAndRedirect(c, h.config)
		return
	}

	status, err := h.queueService.GetQueueStatus(c.Request.Context(), userID)
	if err != nil {
		dto.Fail(c, http.StatusInternalServerError, "Failed to get queue status: "+err.Error())
		return
	}

	response := dto.QueueStatusResponse{
		IsInQueue: status.IsInQueue,
		Position:  status.Position,
		Category:  status.Category,
	}
	if status.IsInQueue {
		response.JoinedAt = status.JoinedAt.Format(time.RFC3339)
		response.ExpiresAt = status.ExpiresAt.Format(time.RFC3339)
	}

	dto.OK(c, response)
}

func (h *QueueHandler) GetQueueStats(c *gin.Context) {
	category := c.Query("category")
	stats := h.queueService.GetQueueStats()

	if category != "" {
		if categoryStats, exists := stats[category]; exists {
			dto.OK(c, categoryStats)
			return
		}
		dto.Fail(c, http.StatusNotFound, "Category not found")
		return
	}

	dto.OK(c, stats)
}

func (h *QueueHandler) GetMatchStats(c *gin.Context) {
	stats := h.queueService.GetMatchStatistics()

	response := dto.MatchStatsResponse{
		TotalMatches:   stats["total_matches"].(int64),
		MaleWaitTime:   stats["male_wait_time"].(time.Duration).String(),
		FemaleWaitTime: stats["female_wait_time"].(time.Duration).String(),
		LastMatchTime:  stats["last_match_time"].(time.Time).Format(time.RFC3339),
	}

	dto.OK(c, response)
}
