package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/matching"
	"github.com/ndquang191/Anochat/api/internal/dto"
	"github.com/ndquang191/Anochat/api/internal/service"
	"github.com/ndquang191/Anochat/api/pkg/apperr"
)

type MatchSettingsHandler struct{ service *service.MatchSettingsService }

func NewMatchSettingsHandler(service *service.MatchSettingsService) *MatchSettingsHandler {
	return &MatchSettingsHandler{service: service}
}

func matchSettingsDTO(view service.MatchSettingsView) dto.MatchSettingsDTO {
	return dto.MatchSettingsDTO{
		DefaultMode: view.DefaultMode, AllowUserChoice: view.AllowUserChoice,
		UserPreference: view.UserPreference, EffectiveMode: view.EffectiveMode,
		QueueDisplayMode: view.QueueDisplayMode, QueueCountMinimum: view.QueueCountMinimum,
		QueueMessageVI: view.QueueMessageVI, QueueMessageEN: view.QueueMessageEN,
	}
}

func (h *MatchSettingsHandler) UpdateUser(c *gin.Context) {
	userID := getUserID(c)
	if userID == uuid.Nil {
		dto.FailErr(c, apperr.ErrUnauthenticated)
		return
	}
	var req dto.UpdateMatchPreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.FailErr(c, apperr.ErrInvalidBody)
		return
	}
	view, err := h.service.UpdateUserPreference(c.Request.Context(), userID, req.Mode)
	if err != nil {
		dto.FailErr(c, err)
		return
	}
	dto.OK(c, matchSettingsDTO(view))
}

func (h *MatchSettingsHandler) GetAdmin(c *gin.Context) {
	settings, err := h.service.GetAdmin(c.Request.Context())
	if err != nil {
		dto.FailErr(c, err)
		return
	}
	dto.OK(c, adminMatchSettingsResponse(settings))
}

func adminMatchSettingsResponse(settings matching.Settings) gin.H {
	return gin.H{
		"default_mode": settings.DefaultMode, "allow_user_choice": settings.AllowUserChoice,
		"rematch_cooldown_seconds": settings.RematchCooldownSeconds,
		"queue_display_mode": settings.QueueDisplayMode, "queue_count_minimum": settings.QueueCountMinimum,
		"queue_message_vi": settings.QueueMessageVI, "queue_message_en": settings.QueueMessageEN,
	}
}

func (h *MatchSettingsHandler) UpdateAdmin(c *gin.Context) {
	var req dto.UpdateAdminMatchSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.FailErr(c, apperr.ErrInvalidBody)
		return
	}
	settings := matching.Settings{
		DefaultMode: req.DefaultMode, AllowUserChoice: *req.AllowUserChoice,
		RematchCooldownSeconds: req.RematchCooldownSeconds,
		QueueDisplayMode: req.QueueDisplayMode, QueueCountMinimum: req.QueueCountMinimum,
		QueueMessageVI: req.QueueMessageVI, QueueMessageEN: req.QueueMessageEN,
	}
	if err := h.service.UpdateAdmin(c.Request.Context(), settings); err != nil {
		dto.FailErr(c, err)
		return
	}
	dto.OK(c, adminMatchSettingsResponse(settings))
}
