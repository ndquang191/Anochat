package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/dto"
	"github.com/ndquang191/Anochat/api/internal/repository"
	"github.com/ndquang191/Anochat/api/internal/service"
)

const AdminUserID = "8d2e7280-bdc8-47b2-8508-8911b5c9f796"

type ModerationHandler struct {
	moderationService *service.ModerationService
	messageRepo       repository.MessageRepository
}

func NewModerationHandler(moderationService *service.ModerationService, messageRepo repository.MessageRepository) *ModerationHandler {
	return &ModerationHandler{moderationService: moderationService, messageRepo: messageRepo}
}

func (h *ModerationHandler) requireAdmin(c *gin.Context) bool {
	id := getUserID(c)
	if id.String() != AdminUserID {
		dto.Fail(c, http.StatusForbidden, "Forbidden")
		c.Abort()
		return false
	}
	return true
}

func (h *ModerationHandler) ListWords(c *gin.Context) {
	if !h.requireAdmin(c) {
		return
	}
	words, err := h.moderationService.ListWords(c.Request.Context())
	if err != nil {
		dto.Fail(c, http.StatusInternalServerError, "Failed to list banned words")
		return
	}

	type wordDTO struct {
		ID        string `json:"id"`
		Word      string `json:"word"`
		CreatedAt int64  `json:"created_at"`
	}
	result := make([]wordDTO, len(words))
	for i, w := range words {
		result[i] = wordDTO{ID: w.ID.String(), Word: w.Word, CreatedAt: w.CreatedAt.Unix()}
	}
	dto.OK(c, result)
}

func (h *ModerationHandler) AddWord(c *gin.Context) {
	if !h.requireAdmin(c) {
		return
	}
	var req dto.AddBannedWordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Fail(c, http.StatusBadRequest, "Invalid request body")
		return
	}
	adminID := getUserID(c)
	word, err := h.moderationService.AddWord(c.Request.Context(), req.Word, adminID)
	if err != nil {
		dto.Fail(c, http.StatusInternalServerError, "Failed to add banned word")
		return
	}
	dto.OKWithMessage(c, "Word added", gin.H{
		"id":         word.ID.String(),
		"word":       word.Word,
		"created_at": word.CreatedAt.Unix(),
	})
}

func (h *ModerationHandler) DeleteWord(c *gin.Context) {
	if !h.requireAdmin(c) {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		dto.Fail(c, http.StatusBadRequest, "Invalid word ID")
		return
	}
	if err := h.moderationService.DeleteWord(c.Request.Context(), id); err != nil {
		dto.Fail(c, http.StatusInternalServerError, "Failed to delete banned word")
		return
	}
	dto.OKWithMessage(c, "Word deleted", nil)
}

func (h *ModerationHandler) ListReports(c *gin.Context) {
	if !h.requireAdmin(c) {
		return
	}
	reports, err := h.moderationService.ListReports(c.Request.Context())
	if err != nil {
		dto.Fail(c, http.StatusInternalServerError, "Failed to list reports")
		return
	}

	type reportDTO struct {
		ID               string  `json:"id"`
		ReporterID       string  `json:"reporter_id"`
		ReportedUserID   string  `json:"reported_user_id"`
		ReportedUserName *string `json:"reported_user_name"`
		RoomID           string  `json:"room_id"`
		Status           string  `json:"status"`
		CreatedAt        int64   `json:"created_at"`
	}
	result := make([]reportDTO, len(reports))
	for i, r := range reports {
		result[i] = reportDTO{
			ID:               r.ID.String(),
			ReporterID:       r.ReporterID.String(),
			ReportedUserID:   r.ReportedUserID.String(),
			ReportedUserName: r.ReportedUserName,
			RoomID:           r.RoomID.String(),
			Status:           r.Status,
			CreatedAt:        r.CreatedAt.Unix(),
		}
	}
	dto.OK(c, result)
}

func (h *ModerationHandler) BanUser(c *gin.Context) {
	if !h.requireAdmin(c) {
		return
	}
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		dto.Fail(c, http.StatusBadRequest, "Invalid user ID")
		return
	}
	if err := h.moderationService.BanUser(c.Request.Context(), userID); err != nil {
		dto.Fail(c, http.StatusInternalServerError, "Failed to ban user")
		return
	}
	dto.OKWithMessage(c, "User banned", nil)
}

func (h *ModerationHandler) ListRoomMessages(c *gin.Context) {
	if !h.requireAdmin(c) {
		return
	}
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		dto.Fail(c, http.StatusBadRequest, "Invalid room ID")
		return
	}
	messages, err := h.messageRepo.FindByRoomID(c.Request.Context(), roomID)
	if err != nil {
		dto.Fail(c, http.StatusInternalServerError, "Failed to fetch messages")
		return
	}
	type msgDTO struct {
		ID        string `json:"id"`
		SenderID  string `json:"sender_id"`
		Content   string `json:"content"`
		CreatedAt int64  `json:"created_at"`
	}
	result := make([]msgDTO, len(messages))
	for i, m := range messages {
		result[i] = msgDTO{
			ID:        m.ID.String(),
			SenderID:  m.SenderID.String(),
			Content:   m.Content,
			CreatedAt: m.CreatedAt.Unix(),
		}
	}
	dto.OK(c, result)
}

func (h *ModerationHandler) CreateReport(c *gin.Context) {
	reporterID := getUserID(c)
	if reporterID == uuid.Nil {
		dto.Fail(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var req dto.CreateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Fail(c, http.StatusBadRequest, "Invalid request body")
		return
	}
	reportedUserID, err := uuid.Parse(req.ReportedUserID)
	if err != nil {
		dto.Fail(c, http.StatusBadRequest, "Invalid reported user ID")
		return
	}
	roomID, err := uuid.Parse(req.RoomID)
	if err != nil {
		dto.Fail(c, http.StatusBadRequest, "Invalid room ID")
		return
	}
	if err := h.moderationService.CreateReport(c.Request.Context(), reporterID, reportedUserID, roomID); err != nil {
		dto.Fail(c, http.StatusInternalServerError, "Failed to submit report")
		return
	}
	dto.OKWithMessage(c, "Report submitted", nil)
}
