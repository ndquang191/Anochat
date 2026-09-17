package repository

import (
	"context"
	"time"

	"github.com/ndquang191/Anochat/api/internal/domain/matching"
	"github.com/ndquang191/Anochat/api/internal/model"
	"gorm.io/gorm"
)

type MatchSettingsRepository interface {
	Get(ctx context.Context) (matching.Settings, error)
	Update(ctx context.Context, settings matching.Settings) error
}

type matchSettingsRepo struct{ db *gorm.DB }

func NewMatchSettingsRepository(db *gorm.DB) MatchSettingsRepository {
	return &matchSettingsRepo{db: db}
}

func (r *matchSettingsRepo) Get(ctx context.Context) (matching.Settings, error) {
	var item model.MatchSettings
	if err := r.db.WithContext(ctx).Where("id = ?", true).First(&item).Error; err != nil {
		return matching.Settings{}, err
	}
	return matching.Settings{
		DefaultMode: item.DefaultMode, AllowUserChoice: item.AllowUserChoice,
		RematchCooldownSeconds: item.RematchCooldownSeconds,
		QueueDisplayMode: item.QueueDisplayMode, QueueCountMinimum: item.QueueCountMinimum,
		QueueMessageVI: item.QueueMessageVI, QueueMessageEN: item.QueueMessageEN,
	}, nil
}

func (r *matchSettingsRepo) Update(ctx context.Context, settings matching.Settings) error {
	return r.db.WithContext(ctx).Model(&model.MatchSettings{}).Where("id = ?", true).Updates(map[string]any{
		"default_mode": settings.DefaultMode, "allow_user_choice": settings.AllowUserChoice,
		"rematch_cooldown_seconds": settings.RematchCooldownSeconds,
		"queue_display_mode": settings.QueueDisplayMode, "queue_count_minimum": settings.QueueCountMinimum,
		"queue_message_vi": settings.QueueMessageVI, "queue_message_en": settings.QueueMessageEN,
		"updated_at": time.Now().UTC(),
	}).Error
}
