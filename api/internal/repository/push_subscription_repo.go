package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PushSubscriptionRepository interface {
	Upsert(ctx context.Context, subscription *model.PushSubscription) error
	FindOwned(ctx context.Context, id, userID uuid.UUID) (*model.PushSubscription, error)
	DeleteOwned(ctx context.Context, id, userID uuid.UUID) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
	MarkSuccess(ctx context.Context, id uuid.UUID, at time.Time) error
}

type pushSubscriptionRepo struct{ db *gorm.DB }

func NewPushSubscriptionRepository(db *gorm.DB) PushSubscriptionRepository {
	return &pushSubscriptionRepo{db: db}
}

func (r *pushSubscriptionRepo) Upsert(ctx context.Context, item *model.PushSubscription) error {
	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "endpoint"}},
		DoUpdates: clause.AssignmentColumns([]string{"user_id", "p256dh", "auth", "locale", "updated_at"}),
	}).Create(item).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Where("endpoint = ?", item.Endpoint).First(item).Error
}

func (r *pushSubscriptionRepo) FindOwned(ctx context.Context, id, userID uuid.UUID) (*model.PushSubscription, error) {
	var item model.PushSubscription
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &item, err
}

func (r *pushSubscriptionRepo) DeleteOwned(ctx context.Context, id, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&model.PushSubscription{}).Error
}

func (r *pushSubscriptionRepo) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.PushSubscription{}, "id = ?", id).Error
}

func (r *pushSubscriptionRepo) MarkSuccess(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.db.WithContext(ctx).Model(&model.PushSubscription{}).Where("id = ?", id).Update("last_success_at", at).Error
}
