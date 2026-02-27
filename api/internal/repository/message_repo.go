package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/chat"
	"github.com/ndquang191/Anochat/api/internal/model"
	"gorm.io/gorm"
)

// MessageRepository defines data access for messages.
type MessageRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*chat.Message, error)
	FindByRoomID(ctx context.Context, roomID uuid.UUID) ([]*chat.Message, error)
	FindByRoomIDWithLimit(ctx context.Context, roomID uuid.UUID, limit, offset int) ([]*chat.Message, error)
	Create(ctx context.Context, msg *chat.Message) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
	DeleteByRoomID(ctx context.Context, roomID uuid.UUID) error
	CountUnread(ctx context.Context, roomID, userID uuid.UUID, lastReadMessageID *uuid.UUID) (int64, error)
}

type messageRepo struct{ db *gorm.DB }

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepo{db: db}
}

func (r *messageRepo) FindByID(ctx context.Context, id uuid.UUID) (*chat.Message, error) {
	var m model.Message
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return messageModelToDomain(&m), nil
}

func (r *messageRepo) FindByRoomID(ctx context.Context, roomID uuid.UUID) ([]*chat.Message, error) {
	var models []model.Message
	if err := r.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Order("created_at ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	return messageModelsToDomain(models), nil
}

func (r *messageRepo) FindByRoomIDWithLimit(ctx context.Context, roomID uuid.UUID, limit, offset int) ([]*chat.Message, error) {
	var models []model.Message
	if err := r.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error; err != nil {
		return nil, err
	}
	return messageModelsToDomain(models), nil
}

func (r *messageRepo) Create(ctx context.Context, msg *chat.Message) error {
	m := messageDomainToModel(msg)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	msg.ID = m.ID
	msg.CreatedAt = m.CreatedAt
	return nil
}

func (r *messageRepo) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Message{}).Error
}

func (r *messageRepo) DeleteByRoomID(ctx context.Context, roomID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("room_id = ?", roomID).Delete(&model.Message{}).Error
}

func (r *messageRepo) CountUnread(ctx context.Context, roomID, userID uuid.UUID, lastReadMessageID *uuid.UUID) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.Message{}).Where("room_id = ? AND sender_id != ?", roomID, userID)
	if lastReadMessageID != nil {
		query = query.Where("created_at > (SELECT created_at FROM messages WHERE id = ?)", lastReadMessageID)
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// --- mapping helpers ---

func messageModelToDomain(m *model.Message) *chat.Message {
	return &chat.Message{
		ID:        m.ID,
		RoomID:    m.RoomID,
		SenderID:  m.SenderID,
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
	}
}

func messageDomainToModel(msg *chat.Message) *model.Message {
	return &model.Message{
		ID:       msg.ID,
		RoomID:   msg.RoomID,
		SenderID: msg.SenderID,
		Content:  msg.Content,
	}
}

func messageModelsToDomain(models []model.Message) []*chat.Message {
	result := make([]*chat.Message, len(models))
	for i := range models {
		result[i] = messageModelToDomain(&models[i])
	}
	return result
}
