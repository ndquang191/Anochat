package service

import (
	"context"
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/repository"
	"github.com/redis/go-redis/v9"
)

type stubAdminStatsRepo struct {
	stats       *repository.AdminStats
	male        int64
	female      int64
	daily       []repository.DailyOverviewMetric
	err         error
	genderError error
	dailyError  error
}

func (r *stubAdminStatsRepo) GetOverview(context.Context) (*repository.AdminStats, error) {
	return r.stats, r.err
}

func (r *stubAdminStatsRepo) GetGenderBreakdown(
	context.Context,
	[]uuid.UUID,
) (int64, int64, error) {
	return r.male, r.female, r.genderError
}

func (r *stubAdminStatsRepo) ListSevenDayMetrics(
	context.Context,
) ([]repository.DailyOverviewMetric, error) {
	return r.daily, r.dailyError
}

func TestAdminStatsServiceGetOverview(t *testing.T) {
	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	ctx := context.Background()
	user1 := uuid.New()
	user2 := uuid.New()
	user3 := uuid.New()
	if err := redisClient.ZAdd(ctx, queueKey,
		redis.Z{Score: 1, Member: user1.String()},
		redis.Z{Score: 2, Member: user2.String()},
		redis.Z{Score: 3, Member: user3.String()},
	).Err(); err != nil {
		t.Fatalf("seed queue: %v", err)
	}

	service := NewAdminStatsService(&stubAdminStatsRepo{
		stats: &repository.AdminStats{
			TotalUsers:       12,
			MaleUsers:        5,
			FemaleUsers:      6,
			UnspecifiedUsers: 1,
			ActiveRooms:      3,
		},
		male:   1,
		female: 1,
		daily: []repository.DailyOverviewMetric{
			{Date: "2026-07-24", Matches: 8, TotalUsers: 12, ActiveRooms: 3},
		},
	}, redisClient)

	got, err := service.GetOverview(ctx)
	if err != nil {
		t.Fatalf("GetOverview() error = %v", err)
	}

	if got.TotalUsers != 12 ||
		got.MaleUsers != 5 ||
		got.FemaleUsers != 6 ||
		got.UnspecifiedUsers != 1 ||
		got.InQueue != 3 ||
		got.InQueueMale != 1 ||
		got.InQueueFemale != 1 ||
		got.InQueueUnknown != 1 ||
		got.ActiveRooms != 3 ||
		len(got.DailyMetrics) != 1 ||
		got.DailyMetrics[0].Matches != 8 {
		t.Fatalf("GetOverview() = %+v", got)
	}
}

func TestAdminStatsServiceReturnsRepositoryError(t *testing.T) {
	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	service := NewAdminStatsService(
		&stubAdminStatsRepo{err: errors.New("database unavailable")},
		redisClient,
	)

	if _, err := service.GetOverview(context.Background()); err == nil {
		t.Fatal("GetOverview() error = nil, want repository error")
	}
}
