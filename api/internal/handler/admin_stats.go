package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/ndquang191/Anochat/api/internal/dto"
	"github.com/ndquang191/Anochat/api/internal/service"
)

type AdminStatsHandler struct {
	adminStatsService *service.AdminStatsService
}

func NewAdminStatsHandler(adminStatsService *service.AdminStatsService) *AdminStatsHandler {
	return &AdminStatsHandler{adminStatsService: adminStatsService}
}

func (h *AdminStatsHandler) GetOverview(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}

	overview, err := h.adminStatsService.GetOverview(c.Request.Context())
	if err != nil {
		dto.FailErr(c, err)
		return
	}

	type dailyMetricDTO struct {
		Date        string `json:"date"`
		Matches     int64  `json:"matches"`
		TotalUsers  int64  `json:"total_users"`
		ActiveRooms int64  `json:"active_rooms"`
	}
	dailyMetrics := make([]dailyMetricDTO, len(overview.DailyMetrics))
	for i, metric := range overview.DailyMetrics {
		dailyMetrics[i] = dailyMetricDTO{
			Date:        metric.Date,
			Matches:     metric.Matches,
			TotalUsers:  metric.TotalUsers,
			ActiveRooms: metric.ActiveRooms,
		}
	}

	dto.OK(c, gin.H{
		"total_users":       overview.TotalUsers,
		"male_users":        overview.MaleUsers,
		"female_users":      overview.FemaleUsers,
		"unspecified_users": overview.UnspecifiedUsers,
		"in_queue":          overview.InQueue,
		"in_queue_male":     overview.InQueueMale,
		"in_queue_female":   overview.InQueueFemale,
		"in_queue_unknown":  overview.InQueueUnknown,
		"active_rooms":      overview.ActiveRooms,
		"daily_metrics":     dailyMetrics,
	})
}
