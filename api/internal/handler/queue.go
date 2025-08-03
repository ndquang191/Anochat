package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/service"
)

// QueueHandler handles queue-related HTTP requests
type QueueHandler struct {
	queueService *service.QueueService
}

// NewQueueHandler creates a new queue handler
func NewQueueHandler(queueService *service.QueueService) *QueueHandler {
	return &QueueHandler{
		queueService: queueService,
	}
}

// JoinQueueRequest represents the request to join a queue
type JoinQueueRequest struct {
	Category string `json:"category" binding:"required"`
}

// JoinQueueResponse represents the response when joining a queue
type JoinQueueResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Position int    `json:"position,omitempty"`
}

// QueueStatusResponse represents the queue status response
type QueueStatusResponse struct {
	IsInQueue bool   `json:"is_in_queue"`
	Position  int    `json:"position"`
	Category  string `json:"category"`
	JoinedAt  string `json:"joined_at"`
}

// QueueStatsResponse represents the queue statistics response
type QueueStatsResponse struct {
	Category          string `json:"category"`
	MaleCount         int    `json:"male_count"`
	FemaleCount       int    `json:"female_count"`
	TotalCount        int    `json:"total_count"`
	EstimatedWaitTime string `json:"estimated_wait_time"`
}

// MatchStatsResponse represents the match statistics response
type MatchStatsResponse struct {
	TotalMatches   int64  `json:"total_matches"`
	MaleWaitTime   string `json:"male_wait_time"`
	FemaleWaitTime string `json:"female_wait_time"`
	LastMatchTime  string `json:"last_match_time"`
}

// JoinQueue handles the request to join a queue
func (h *QueueHandler) JoinQueue(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	// Parse request body
	var req JoinQueueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	// Join queue
	entry, err := h.queueService.JoinQueue(c.Request.Context(), userID, req.Category)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Get position from entry
	position := 0
	if entry != nil {
		// Find position in the appropriate queue
		status, _ := h.queueService.GetQueueStatus(c.Request.Context(), userID)
		if status != nil {
			position = status.Position
		}
	}

	c.JSON(http.StatusOK, JoinQueueResponse{
		Success:  true,
		Message:  "Successfully joined queue",
		Position: position,
	})
}

// LeaveQueue handles the request to leave a queue
func (h *QueueHandler) LeaveQueue(c *gin.Context) {
	// Get user ID from context
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	// Leave queue
	err = h.queueService.LeaveQueue(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Successfully left queue",
	})
}

// GetQueueStatus handles the request to get queue status
func (h *QueueHandler) GetQueueStatus(c *gin.Context) {
	// Get user ID from context
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	// Get queue status
	status, err := h.queueService.GetQueueStatus(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get queue status: " + err.Error(),
		})
		return
	}

	response := QueueStatusResponse{
		IsInQueue: status.IsInQueue,
		Position:  status.Position,
		Category:  status.Category,
		JoinedAt:  status.JoinedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetQueueStats handles the request to get queue statistics
func (h *QueueHandler) GetQueueStats(c *gin.Context) {
	// Get category from query parameter (optional)
	category := c.Query("category")

	// Get queue stats
	stats := h.queueService.GetQueueStats()

	// If category is specified, return only that category
	if category != "" {
		if categoryStats, exists := stats[category]; exists {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    categoryStats,
			})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Category not found",
		})
		return
	}

	// Return all categories
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetMatchStats handles the request to get match statistics
func (h *QueueHandler) GetMatchStats(c *gin.Context) {
	// Get match statistics
	stats := h.queueService.GetMatchStatistics()

	// Convert to response format
	response := MatchStatsResponse{
		TotalMatches:   stats["total_matches"].(int64),
		MaleWaitTime:   stats["male_wait_time"].(time.Duration).String(),
		FemaleWaitTime: stats["female_wait_time"].(time.Duration).String(),
		LastMatchTime:  stats["last_match_time"].(time.Time).Format("2006-01-02T15:04:05Z07:00"),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// UserDisconnected handles user disconnection (called by WebSocket or other connection handlers)
func (h *QueueHandler) UserDisconnected(userID uuid.UUID) {
	h.queueService.UserDisconnected(userID)
}

// SetupRoutes sets up the queue routes
func (h *QueueHandler) SetupRoutes(router *gin.RouterGroup) {
	queue := router.Group("/queue")
	{
		queue.POST("/join", h.JoinQueue)
		queue.POST("/leave", h.LeaveQueue)
		queue.GET("/status", h.GetQueueStatus)
		queue.GET("/stats", h.GetQueueStats)
		queue.GET("/match-stats", h.GetMatchStats)
	}
}
