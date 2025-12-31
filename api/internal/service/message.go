package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/model"
	"gorm.io/gorm"
)

// MessageService handles message operations
type MessageService struct {
	db *gorm.DB
}

// NewMessageService creates a new message service
func NewMessageService(db *gorm.DB) *MessageService {
	return &MessageService{db: db}
}

// CreateMessage creates a new message
func (s *MessageService) CreateMessage(ctx context.Context, roomID, senderID uuid.UUID, content string) (*model.Message, error) {
	message := &model.Message{
		RoomID:   roomID,
		SenderID: senderID,
		Content:  content,
	}

	if err := s.db.WithContext(ctx).Create(message).Error; err != nil {
		return nil, err
	}

	return message, nil
}

// GetMessageByID retrieves a message by ID
func (s *MessageService) GetMessageByID(ctx context.Context, messageID uuid.UUID) (*model.Message, error) {
	var message model.Message
	if err := s.db.WithContext(ctx).
		Preload("Sender").
		Where("id = ?", messageID).
		First(&message).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("message not found")
		}
		return nil, err
	}
	return &message, nil
}

// GetMessagesByRoomID retrieves all messages for a room
func (s *MessageService) GetMessagesByRoomID(ctx context.Context, roomID uuid.UUID) ([]model.Message, error) {
	var messages []model.Message
	if err := s.db.WithContext(ctx).
		Preload("Sender").
		Where("room_id = ?", roomID).
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

// GetMessagesByRoomIDWithLimit retrieves messages with pagination
func (s *MessageService) GetMessagesByRoomIDWithLimit(ctx context.Context, roomID uuid.UUID, limit, offset int) ([]model.Message, error) {
	var messages []model.Message
	if err := s.db.WithContext(ctx).
		Preload("Sender").
		Where("room_id = ?", roomID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

// GetUnreadMessagesCount gets count of unread messages for a user in a room
func (s *MessageService) GetUnreadMessagesCount(ctx context.Context, roomID, userID uuid.UUID) (int64, error) {
	// Get the last read message ID for the user
	var room model.Room
	if err := s.db.WithContext(ctx).Where("id = ?", roomID).First(&room).Error; err != nil {
		return 0, err
	}

	var lastReadMessageID *uuid.UUID
	if room.User1ID == userID {
		lastReadMessageID = room.User1LastReadMessageID
	} else if room.User2ID == userID {
		lastReadMessageID = room.User2LastReadMessageID
	} else {
		return 0, errors.New("user not part of this room")
	}

	var count int64
	query := s.db.WithContext(ctx).Model(&model.Message{}).Where("room_id = ? AND sender_id != ?", roomID, userID)

	if lastReadMessageID != nil {
		query = query.Where("created_at > (SELECT created_at FROM messages WHERE id = ?)", lastReadMessageID)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// GetUnreadMessages gets unread messages for a user in a room
func (s *MessageService) GetUnreadMessages(ctx context.Context, roomID, userID uuid.UUID) ([]model.Message, error) {
	// Get the last read message ID for the user
	var room model.Room
	if err := s.db.WithContext(ctx).Where("id = ?", roomID).First(&room).Error; err != nil {
		return nil, err
	}

	var lastReadMessageID *uuid.UUID
	if room.User1ID == userID {
		lastReadMessageID = room.User1LastReadMessageID
	} else if room.User2ID == userID {
		lastReadMessageID = room.User2LastReadMessageID
	} else {
		return nil, errors.New("user not part of this room")
	}

	var messages []model.Message
	query := s.db.WithContext(ctx).
		Preload("Sender").
		Where("room_id = ? AND sender_id != ?", roomID, userID)

	if lastReadMessageID != nil {
		query = query.Where("created_at > (SELECT created_at FROM messages WHERE id = ?)", lastReadMessageID)
	}

	if err := query.Order("created_at ASC").Find(&messages).Error; err != nil {
		return nil, err
	}

	return messages, nil
}

// DeleteMessage deletes a message (for admin purposes)
func (s *MessageService) DeleteMessage(ctx context.Context, messageID uuid.UUID) error {
	return s.db.WithContext(ctx).Where("id = ?", messageID).Delete(&model.Message{}).Error
}

// DeleteMessagesByRoomID deletes all messages in a room
func (s *MessageService) DeleteMessagesByRoomID(ctx context.Context, roomID uuid.UUID) error {
	return s.db.WithContext(ctx).Where("room_id = ?", roomID).Delete(&model.Message{}).Error
}

// GetMessageStats gets message statistics for a room
func (s *MessageService) GetMessageStats(ctx context.Context, roomID uuid.UUID) (map[string]interface{}, error) {
	var totalMessages int64
	var firstMessage model.Message
	var lastMessage model.Message

	// Get total count
	if err := s.db.WithContext(ctx).Model(&model.Message{}).Where("room_id = ?", roomID).Count(&totalMessages).Error; err != nil {
		return nil, err
	}

	// Get first message
	if err := s.db.WithContext(ctx).Where("room_id = ?", roomID).Order("created_at ASC").First(&firstMessage).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// No messages in room
			return map[string]interface{}{
				"total_messages": 0,
				"first_message":  nil,
				"last_message":   nil,
			}, nil
		}
		return nil, err
	}

	// Get last message
	if err := s.db.WithContext(ctx).Where("room_id = ?", roomID).Order("created_at DESC").First(&lastMessage).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_messages": totalMessages,
		"first_message":  firstMessage,
		"last_message":   lastMessage,
	}, nil
}
