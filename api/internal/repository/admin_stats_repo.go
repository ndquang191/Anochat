package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminStats struct {
	TotalUsers       int64
	MaleUsers        int64
	FemaleUsers      int64
	UnspecifiedUsers int64
	ActiveRooms      int64
}

type DailyOverviewMetric struct {
	Date        string
	Matches     int64
	TotalUsers  int64
	ActiveRooms int64
}

type AdminStatsRepository interface {
	GetOverview(ctx context.Context) (*AdminStats, error)
	GetGenderBreakdown(ctx context.Context, userIDs []uuid.UUID) (male, female int64, err error)
	ListSevenDayMetrics(ctx context.Context) ([]DailyOverviewMetric, error)
}

type adminStatsRepo struct {
	db *gorm.DB
}

func NewAdminStatsRepository(db *gorm.DB) AdminStatsRepository {
	return &adminStatsRepo{db: db}
}

func (r *adminStatsRepo) GetOverview(ctx context.Context) (*AdminStats, error) {
	var stats AdminStats
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			(SELECT COUNT(*) FROM users WHERE is_deleted = false) AS total_users,
			(
				SELECT COUNT(*)
				FROM profiles
				JOIN users ON users.id = profiles.user_id
				WHERE users.is_deleted = false AND profiles.is_male = true
			) AS male_users,
			(
				SELECT COUNT(*)
				FROM profiles
				JOIN users ON users.id = profiles.user_id
				WHERE users.is_deleted = false AND profiles.is_male = false
			) AS female_users,
			(
				SELECT COUNT(*)
				FROM users
				LEFT JOIN profiles ON profiles.user_id = users.id
				WHERE users.is_deleted = false AND profiles.is_male IS NULL
			) AS unspecified_users,
			(SELECT COUNT(*) FROM rooms WHERE ended_at IS NULL) AS active_rooms
	`).Scan(&stats).Error
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (r *adminStatsRepo) GetGenderBreakdown(
	ctx context.Context,
	userIDs []uuid.UUID,
) (male, female int64, err error) {
	if len(userIDs) == 0 {
		return 0, 0, nil
	}

	var result struct {
		Male   int64
		Female int64
	}
	err = r.db.WithContext(ctx).
		Table("users").
		Select(`
			COUNT(*) FILTER (WHERE profiles.is_male = true) AS male,
			COUNT(*) FILTER (WHERE profiles.is_male = false) AS female
		`).
		Joins("LEFT JOIN profiles ON profiles.user_id = users.id").
		Where("users.is_deleted = false AND users.id IN ?", userIDs).
		Scan(&result).Error
	if err != nil {
		return 0, 0, err
	}
	return result.Male, result.Female, nil
}

func (r *adminStatsRepo) ListSevenDayMetrics(
	ctx context.Context,
) ([]DailyOverviewMetric, error) {
	var metrics []DailyOverviewMetric
	err := r.db.WithContext(ctx).Raw(`
		WITH days AS (
			SELECT GENERATE_SERIES(
				(NOW() AT TIME ZONE 'Asia/Ho_Chi_Minh')::date - 6,
				(NOW() AT TIME ZONE 'Asia/Ho_Chi_Minh')::date,
				INTERVAL '1 day'
			)::date AS day
		),
		boundaries AS (
			SELECT
				day,
				day::timestamp AT TIME ZONE 'Asia/Ho_Chi_Minh' AS day_start,
				CASE
					WHEN day = (NOW() AT TIME ZONE 'Asia/Ho_Chi_Minh')::date
						THEN NOW()
					ELSE (day + 1)::timestamp AT TIME ZONE 'Asia/Ho_Chi_Minh'
				END AS day_end
			FROM days
		)
		SELECT
			TO_CHAR(day, 'YYYY-MM-DD') AS date,
			(
				SELECT COUNT(*)
				FROM room_sessions
				WHERE matched_at >= day_start AND matched_at < day_end
			) AS matches,
			(
				SELECT COUNT(*)
				FROM users
				WHERE is_deleted = false AND created_at < day_end
			) AS total_users,
			(
				SELECT COUNT(*)
				FROM room_sessions
				WHERE matched_at < day_end
				  AND (ended_at IS NULL OR ended_at > day_end)
			) AS active_rooms
		FROM boundaries
		ORDER BY day
	`).Scan(&metrics).Error
	return metrics, err
}
