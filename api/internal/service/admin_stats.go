package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/repository"
	"github.com/redis/go-redis/v9"
)

type AdminOverview struct {
	TotalUsers       int64
	MaleUsers        int64
	FemaleUsers      int64
	UnspecifiedUsers int64
	InQueue          int64
	InQueueMale      int64
	InQueueFemale    int64
	InQueueUnknown   int64
	ActiveRooms      int64
	DailyMetrics     []DailyOverviewMetric
}

type DailyOverviewMetric struct {
	Date        string
	Matches     int64
	TotalUsers  int64
	ActiveRooms int64
}

type AdminStatsService struct {
	statsRepo repository.AdminStatsRepository
	rdb       *redis.Client
}

func NewAdminStatsService(
	statsRepo repository.AdminStatsRepository,
	rdb *redis.Client,
) *AdminStatsService {
	return &AdminStatsService{statsRepo: statsRepo, rdb: rdb}
}

func (s *AdminStatsService) GetOverview(ctx context.Context) (*AdminOverview, error) {
	stats, err := s.statsRepo.GetOverview(ctx)
	if err != nil {
		return nil, fmt.Errorf("get admin overview from database: %w", err)
	}

	queueMembers, err := s.rdb.ZRange(ctx, queueKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("get matching queue users: %w", err)
	}

	queueUserIDs := make([]uuid.UUID, 0, len(queueMembers))
	for _, member := range queueMembers {
		if userID, parseErr := uuid.Parse(member); parseErr == nil {
			queueUserIDs = append(queueUserIDs, userID)
		}
	}

	queueMale, queueFemale, err := s.statsRepo.GetGenderBreakdown(ctx, queueUserIDs)
	if err != nil {
		return nil, fmt.Errorf("get matching queue gender breakdown: %w", err)
	}

	dailyMetrics, err := s.statsRepo.ListSevenDayMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("get seven-day overview metrics: %w", err)
	}
	metricItems := make([]DailyOverviewMetric, len(dailyMetrics))
	for i, metric := range dailyMetrics {
		metricItems[i] = DailyOverviewMetric{
			Date:        metric.Date,
			Matches:     metric.Matches,
			TotalUsers:  metric.TotalUsers,
			ActiveRooms: metric.ActiveRooms,
		}
	}

	inQueue := int64(len(queueMembers))
	return &AdminOverview{
		TotalUsers:       stats.TotalUsers,
		MaleUsers:        stats.MaleUsers,
		FemaleUsers:      stats.FemaleUsers,
		UnspecifiedUsers: stats.UnspecifiedUsers,
		InQueue:          inQueue,
		InQueueMale:      queueMale,
		InQueueFemale:    queueFemale,
		InQueueUnknown:   max(0, inQueue-queueMale-queueFemale),
		ActiveRooms:      stats.ActiveRooms,
		DailyMetrics:     metricItems,
	}, nil
}
