package handler

import (
	"encoding/base64"
	"encoding/binary"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/identity"
	"github.com/ndquang191/Anochat/api/internal/domain/moderation"
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
	limit, err := parseAdminPageSize(c.Query("limit"))
	if err != nil {
		dto.FailErr(c, err)
		return
	}
	before, err := decodeReportGroupCursor(c.Query("before"))
	if err != nil {
		dto.FailErr(c, err)
		return
	}
	status := c.DefaultQuery("status", "pending")
	if status != "pending" && status != "reviewed" {
		dto.FailErr(c, apperr.ErrInvalidBody)
		return
	}
	query := strings.TrimSpace(c.Query("query"))
	if len(query) > 100 {
		dto.FailErr(c, apperr.ErrInvalidBody)
		return
	}
	page, err := h.moderationService.ListReportGroups(
		c.Request.Context(),
		status,
		query,
		before,
		limit,
	)
	if err != nil {
		dto.FailErr(c, err)
		return
	}

	type reportGroupDTO struct {
		ReportedUserID   string  `json:"reported_user_id"`
		ReportedUserName *string `json:"reported_user_name"`
		ReportCount      int64   `json:"report_count"`
		AutoCount        int64   `json:"auto_count"`
		ManualCount      int64   `json:"manual_count"`
		LatestReportID   string  `json:"latest_report_id"`
	}
	groups := make([]reportGroupDTO, len(page.Groups))
	for i, group := range page.Groups {
		groups[i] = reportGroupDTO{
			ReportedUserID:   group.ReportedUserID.String(),
			ReportedUserName: group.ReportedUserName,
			ReportCount:      group.ReportCount,
			AutoCount:        group.AutoCount,
			ManualCount:      group.ManualCount,
			LatestReportID:   group.LatestReportID.String(),
		}
	}
	var nextCursor *string
	if page.NextCursor != nil {
		encoded := encodeReportGroupCursor(page.NextCursor)
		nextCursor = &encoded
	}
	dto.OK(c, gin.H{
		"groups":      groups,
		"next_cursor": nextCursor,
		"has_more":    page.HasMore,
	})
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
	limit, err := parseAdminPageSize(c.Query("limit"))
	if err != nil {
		dto.FailErr(c, err)
		return
	}
	before, err := decodeBannedUserCursor(c.Query("before"))
	if err != nil {
		dto.FailErr(c, err)
		return
	}
	query := strings.TrimSpace(c.Query("query"))
	if len(query) > 100 {
		dto.FailErr(c, apperr.ErrInvalidBody)
		return
	}
	page, err := h.moderationService.ListBannedUsers(c.Request.Context(), query, before, limit)
	if err != nil {
		dto.FailErr(c, err)
		return
	}

	userIDs := make([]uuid.UUID, len(page.Users))
	for i, u := range page.Users {
		userIDs[i] = u.ID
	}
	reportMap, err := h.moderationService.GetLatestReportsForUsers(c.Request.Context(), userIDs)
	if err != nil {
		dto.FailErr(c, err)
		return
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
		LastReportID       *string `json:"last_report_id"`
	}
	users := make([]bannedUserDTO, len(page.Users))
	for i, u := range page.Users {
		var lastReportID *string
		var bannedAt *int64
		if rid, ok := reportMap[u.ID]; ok {
			s := rid.String()
			lastReportID = &s
		}
		if u.BannedAt != nil {
			ts := u.BannedAt.Unix()
			bannedAt = &ts
		}
		users[i] = bannedUserDTO{
			ID:                 u.ID.String(),
			Name:               u.Name,
			Email:              u.Email,
			CreatedAt:          u.CreatedAt.Unix(),
			BannedAt:           bannedAt,
			BanCount:           u.BanCount,
			ReviewRequestCount: u.ReviewRequestCount,
			ReviewRequested:    u.ReviewRequested,
			LastReportID:       lastReportID,
		}
	}
	var nextCursor *string
	if page.NextCursor != nil {
		encoded := encodeBannedUserCursor(page.NextCursor)
		nextCursor = &encoded
	}
	dto.OK(c, gin.H{
		"users":       users,
		"next_cursor": nextCursor,
		"has_more":    page.HasMore,
		"total":       page.Total,
	})
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

func parseAdminPageSize(raw string) (int, error) {
	if raw == "" {
		return service.DefaultAdminPageSize, nil
	}
	limit, err := strconv.Atoi(raw)
	if err != nil || limit < 1 || limit > service.MaxAdminPageSize {
		return 0, apperr.ErrInvalidPageSize
	}
	return limit, nil
}

func encodeReportGroupCursor(cursor *moderation.ReportGroupCursor) string {
	buf := make([]byte, 24)
	binary.BigEndian.PutUint64(buf[:8], uint64(cursor.ReportCount))
	copy(buf[8:], cursor.ReportedUserID[:])
	return base64.RawURLEncoding.EncodeToString(buf)
}

func decodeReportGroupCursor(raw string) (*moderation.ReportGroupCursor, error) {
	if raw == "" {
		return nil, nil
	}
	buf, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil || len(buf) != 24 {
		return nil, apperr.ErrInvalidCursor
	}
	reportCount := binary.BigEndian.Uint64(buf[:8])
	if reportCount > uint64(^uint64(0)>>1) {
		return nil, apperr.ErrInvalidCursor
	}
	id, err := uuid.FromBytes(buf[8:])
	if err != nil || id == uuid.Nil {
		return nil, apperr.ErrInvalidCursor
	}
	return &moderation.ReportGroupCursor{
		ReportCount:    int64(reportCount),
		ReportedUserID: id,
	}, nil
}

func encodeBannedUserCursor(cursor *identity.BannedUserCursor) string {
	buf := make([]byte, 25)
	if cursor.ReviewRequested {
		buf[0] = 1
	}
	binary.BigEndian.PutUint64(buf[1:9], uint64(cursor.SortAt.UnixNano()))
	copy(buf[9:], cursor.ID[:])
	return base64.RawURLEncoding.EncodeToString(buf)
}

func decodeBannedUserCursor(raw string) (*identity.BannedUserCursor, error) {
	if raw == "" {
		return nil, nil
	}
	buf, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil || len(buf) != 25 || buf[0] > 1 {
		return nil, apperr.ErrInvalidCursor
	}
	id, err := uuid.FromBytes(buf[9:])
	if err != nil || id == uuid.Nil {
		return nil, apperr.ErrInvalidCursor
	}
	sortAt := time.Unix(0, int64(binary.BigEndian.Uint64(buf[1:9]))).UTC()
	if sortAt.IsZero() {
		return nil, apperr.ErrInvalidCursor
	}
	return &identity.BannedUserCursor{
		ReviewRequested: buf[0] == 1,
		SortAt:          sortAt,
		ID:              id,
	}, nil
}
