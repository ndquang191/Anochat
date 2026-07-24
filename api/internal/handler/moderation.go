package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/dto"
	"github.com/ndquang191/Anochat/api/internal/service"
	"github.com/ndquang191/Anochat/api/pkg/apperr"
)

type ModerationHandler struct {
	moderationService *service.ModerationService
}

func NewModerationHandler(moderationService *service.ModerationService) *ModerationHandler {
	return &ModerationHandler{moderationService: moderationService}
}

func requireAdmin(c *gin.Context) bool {
	if !c.GetBool("is_admin") {
		dto.FailErr(c, apperr.ErrForbidden)
		c.Abort()
		return false
	}
	return true
}

func (h *ModerationHandler) requireAdmin(c *gin.Context) bool {
	return requireAdmin(c)
}

func (h *ModerationHandler) ListWords(c *gin.Context) {
	if !h.requireAdmin(c) {
		return
	}
	words, err := h.moderationService.ListWords(c.Request.Context())
	if err != nil {
		dto.FailErr(c, err)
		return
	}

	type wordDTO struct {
		ID        string `json:"id"`
		Word      string `json:"word"`
		Category  string `json:"category"`
		CreatedAt int64  `json:"created_at"`
	}
	result := make([]wordDTO, len(words))
	for i, w := range words {
		result[i] = wordDTO{ID: w.ID.String(), Word: w.Word, Category: w.Category, CreatedAt: w.CreatedAt.Unix()}
	}
	dto.OK(c, result)
}

func (h *ModerationHandler) AddWord(c *gin.Context) {
	if !h.requireAdmin(c) {
		return
	}
	var req dto.AddBannedWordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.FailErr(c, apperr.ErrInvalidBody)
		return
	}
	adminID := getUserID(c)
	word, err := h.moderationService.AddWord(c.Request.Context(), req.Word, req.Category, adminID)
	if err != nil {
		dto.FailErr(c, err)
		return
	}
	dto.OKWithMessage(c, "Word added", gin.H{
		"id":         word.ID.String(),
		"word":       word.Word,
		"category":   word.Category,
		"created_at": word.CreatedAt.Unix(),
	})
}

func (h *ModerationHandler) UpdateWord(c *gin.Context) {
	if !h.requireAdmin(c) {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		dto.FailErr(c, apperr.ErrInvalidID)
		return
	}
	var req dto.UpdateBannedWordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.FailErr(c, apperr.ErrInvalidBody)
		return
	}
	if err := h.moderationService.UpdateWord(c.Request.Context(), id, req.Word, req.Category); err != nil {
		dto.FailErr(c, err)
		return
	}
	dto.OKWithMessage(c, "Word updated", nil)
}

func (h *ModerationHandler) DeleteWord(c *gin.Context) {
	if !h.requireAdmin(c) {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		dto.FailErr(c, apperr.ErrInvalidID)
		return
	}
	if err := h.moderationService.DeleteWord(c.Request.Context(), id); err != nil {
		dto.FailErr(c, err)
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
		dto.FailErr(c, err)
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
		dto.FailErr(c, apperr.ErrInvalidID)
		return
	}
	if err := h.moderationService.BanUser(c.Request.Context(), userID); err != nil {
		dto.FailErr(c, err)
		return
	}
	dto.OKWithMessage(c, "User banned", nil)
}

func (h *ModerationHandler) ListReportMessages(c *gin.Context) {
	if !h.requireAdmin(c) {
		return
	}
	reportID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		dto.FailErr(c, apperr.ErrInvalidID)
		return
	}
	messages, err := h.moderationService.GetReportMessages(c.Request.Context(), reportID)
	if err != nil {
		dto.FailErr(c, err)
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

func (h *ModerationHandler) ListBannedUsers(c *gin.Context) {
	if !h.requireAdmin(c) {
		return
	}
	users, err := h.moderationService.ListBannedUsers(c.Request.Context())
	if err != nil {
		dto.FailErr(c, err)
		return
	}

	userIDs := make([]uuid.UUID, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}
	roomMap, err := h.moderationService.GetLatestRoomsForUsers(c.Request.Context(), userIDs)
	if err != nil {
		roomMap = map[uuid.UUID]uuid.UUID{} // non-fatal
	}
	reportMap, err := h.moderationService.GetLatestReportsForUsers(c.Request.Context(), userIDs)
	if err != nil {
		reportMap = map[uuid.UUID]uuid.UUID{} // non-fatal
	}

	type bannedUserDTO struct {
		ID                 string  `json:"id"`
		Name               *string `json:"name"`
		Email              *string `json:"email"`
		CreatedAt          int64   `json:"created_at"`
		BannedAt           *int64  `json:"banned_at"`
		BanCount           int     `json:"ban_count"`
		ReviewRequestCount int     `json:"review_request_count"`
		ReviewRequested    bool    `json:"review_requested"`
		ReviewRequestedAt  *int64  `json:"review_requested_at"`
		LastRoomID         *string `json:"last_room_id"`
		LastReportID       *string `json:"last_report_id"`
	}
	result := make([]bannedUserDTO, len(users))
	for i, u := range users {
		var lastRoomID *string
		var lastReportID *string
		var bannedAt *int64
		var reviewRequestedAt *int64
		if rid, ok := roomMap[u.ID]; ok {
			s := rid.String()
			lastRoomID = &s
		}
		if rid, ok := reportMap[u.ID]; ok {
			s := rid.String()
			lastReportID = &s
		}
		if u.BannedAt != nil {
			ts := u.BannedAt.Unix()
			bannedAt = &ts
		}
		if u.ReviewRequestedAt != nil {
			ts := u.ReviewRequestedAt.Unix()
			reviewRequestedAt = &ts
		}
		result[i] = bannedUserDTO{
			ID:                 u.ID.String(),
			Name:               u.Name,
			Email:              u.Email,
			CreatedAt:          u.CreatedAt.Unix(),
			BannedAt:           bannedAt,
			BanCount:           u.BanCount,
			ReviewRequestCount: u.ReviewRequestCount,
			ReviewRequested:    u.ReviewRequested,
			ReviewRequestedAt:  reviewRequestedAt,
			LastRoomID:         lastRoomID,
			LastReportID:       lastReportID,
		}
	}
	dto.OK(c, result)
}

func (h *ModerationHandler) UnbanUser(c *gin.Context) {
	if !h.requireAdmin(c) {
		return
	}
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		dto.FailErr(c, apperr.ErrInvalidID)
		return
	}
	if err := h.moderationService.UnbanUser(c.Request.Context(), userID); err != nil {
		dto.FailErr(c, err)
		return
	}
	dto.OKWithMessage(c, "User unbanned", nil)
}

func (h *ModerationHandler) RequestBanReview(c *gin.Context) {
	userID := getUserID(c)
	if userID == uuid.Nil {
		dto.FailErr(c, apperr.ErrUnauthenticated)
		return
	}
	if err := h.moderationService.RequestBanReview(c.Request.Context(), userID); err != nil {
		dto.FailErr(c, err)
		return
	}
	dto.OKWithMessage(c, "Review request submitted", nil)
}

func (h *ModerationHandler) CreateReport(c *gin.Context) {
	reporterID := getUserID(c)
	if reporterID == uuid.Nil {
		dto.FailErr(c, apperr.ErrUnauthenticated)
		return
	}
	var req dto.CreateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.FailErr(c, apperr.ErrInvalidBody)
		return
	}
	reportedUserID, err := uuid.Parse(req.ReportedUserID)
	if err != nil {
		dto.FailErr(c, apperr.ErrInvalidID)
		return
	}
	roomID, err := uuid.Parse(req.RoomID)
	if err != nil {
		dto.FailErr(c, apperr.ErrInvalidID)
		return
	}
	if err := h.moderationService.CreateReport(c.Request.Context(), reporterID, reportedUserID, roomID); err != nil {
		dto.FailErr(c, err)
		return
	}
	dto.OKWithMessage(c, "Report submitted", nil)
}
