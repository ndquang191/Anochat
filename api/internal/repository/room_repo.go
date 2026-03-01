package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/chat"
	"github.com/ndquang191/Anochat/api/internal/model"
	"gorm.io/gorm"
)

// RoomRepository defines data access for rooms.
type RoomRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*chat.Room, error)
	FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*chat.Room, error)
	Create(ctx context.Context, room *chat.Room) error
	UpdateEndedAt(ctx context.Context, roomID uuid.UUID, endedAt time.Time) error
	Delete(ctx context.Context, roomID uuid.UUID) error
}

type roomRepo struct{ db *gorm.DB }

func NewRoomRepository(db *gorm.DB) RoomRepository {
	return &roomRepo{db: db}
}

func (r *roomRepo) FindByID(ctx context.Context, id uuid.UUID) (*chat.Room, error) {
	var m model.Room
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return roomModelToDomain(&m), nil
}

func (r *roomRepo) FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*chat.Room, error) {
	var m model.Room
	if err := r.db.WithContext(ctx).
		Where("(user1_id = ? OR user2_id = ?) AND ended_at IS NULL", userID, userID).
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return roomModelToDomain(&m), nil
}

func (r *roomRepo) Create(ctx context.Context, room *chat.Room) error {
	m := roomDomainToModel(room)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	room.ID = m.ID
	room.CreatedAt = m.CreatedAt
	return nil
}

func (r *roomRepo) UpdateEndedAt(ctx context.Context, roomID uuid.UUID, endedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&model.Room{}).Where("id = ?", roomID).Update("ended_at", endedAt).Error
}

func (r *roomRepo) Delete(ctx context.Context, roomID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", roomID).Delete(&model.Room{}).Error
}

// --- mapping helpers ---

func roomModelToDomain(m *model.Room) *chat.Room {
	return &chat.Room{
		ID:        m.ID,
		User1ID:   m.User1ID,
		User2ID:   m.User2ID,
		CreatedAt: m.CreatedAt,
		EndedAt:   m.EndedAt,
	}
}

func roomDomainToModel(r *chat.Room) *model.Room {
	return &model.Room{
		ID:        r.ID,
		User1ID:   r.User1ID,
		User2ID:   r.User2ID,
		CreatedAt: r.CreatedAt,
		EndedAt:   r.EndedAt,
	}
}
