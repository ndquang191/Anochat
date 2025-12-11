package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/model"
	"github.com/ndquang191/Anochat/api/pkg/config"
	"gorm.io/gorm"
)

// RoomService handles room operations
type RoomService struct {
	db              *gorm.DB
	messageAnalyzer *config.MessageAnalyzer
}

// NewRoomService creates a new room service
func NewRoomService(db *gorm.DB) *RoomService {
	return &RoomService{
		db:              db,
		messageAnalyzer: config.NewMessageAnalyzer(),
	}
}

// CreateRoom creates a new room between two users
func (s *RoomService) CreateRoom(ctx context.Context, user1ID, user2ID uuid.UUID, category string) (*model.Room, error) {
	room := &model.Room{
		User1ID:  user1ID,
		User2ID:  user2ID,
		Category: category,
	}

	if err := s.db.WithContext(ctx).Create(room).Error; err != nil {
		return nil, err
	}

	return room, nil
}

// GetRoomByID retrieves a room by ID
func (s *RoomService) GetRoomByID(ctx context.Context, roomID uuid.UUID) (*model.Room, error) {
	var room model.Room
	if err := s.db.WithContext(ctx).
		Preload("User1").
		Preload("User2").
		Where("id = ? AND is_deleted = false", roomID).
		First(&room).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("room not found")
		}
		return nil, err
	}
	return &room, nil
}

// GetActiveRoomByUserID gets the active room for a user
func (s *RoomService) GetActiveRoomByUserID(ctx context.Context, userID uuid.UUID) (*model.Room, error) {
	var room model.Room
	if err := s.db.WithContext(ctx).
		Preload("User1").
		Preload("User2").
		Where("(user1_id = ? OR user2_id = ?) AND ended_at IS NULL AND is_deleted = false", userID, userID).
		First(&room).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("no active room found")
		}
		return nil, err
	}
	return &room, nil
}

// LeaveRoom marks a room as ended
func (s *RoomService) LeaveRoom(ctx context.Context, roomID, userID uuid.UUID) error {
	room, err := s.GetRoomByID(ctx, roomID)
	if err != nil {
		return err
	}

	// Verify user is part of the room
	if room.User1ID != userID && room.User2ID != userID {
		return errors.New("user not part of this room")
	}

	// Check if room is already ended
	if room.EndedAt != nil {
		// Room already ended, return success
		return nil
	}

	// Mark room as ended
	now := time.Now()
	room.EndedAt = &now

	// Use Model().Update for immediate database update
	if err := s.db.WithContext(ctx).Model(&model.Room{}).
		Where("id = ?", roomID).
		Update("ended_at", now).Error; err != nil {
		return err
	}

	// Trigger post-chat cleanup in background
	go s.cleanupRoom(context.Background(), roomID)

	return nil
}

// UpdateRoom updates room status (e.g., mark as sensitive)
func (s *RoomService) UpdateRoom(ctx context.Context, roomID uuid.UUID, isSensitive *bool) (*model.Room, error) {
	room, err := s.GetRoomByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	if isSensitive != nil {
		room.IsSensitive = *isSensitive
	}

	if err := s.db.WithContext(ctx).Save(room).Error; err != nil {
		return nil, err
	}

	return room, nil
}

// UpdateLastReadMessage updates the last read message for a user in a room
func (s *RoomService) UpdateLastReadMessage(ctx context.Context, roomID, userID, messageID uuid.UUID) error {
	room, err := s.GetRoomByID(ctx, roomID)
	if err != nil {
		return err
	}

	// Determine which user field to update
	if room.User1ID == userID {
		room.User1LastReadMessageID = &messageID
	} else if room.User2ID == userID {
		room.User2LastReadMessageID = &messageID
	} else {
		return errors.New("user not part of this room")
	}

	return s.db.WithContext(ctx).Save(room).Error
}

// GetRoomHistory retrieves messages for a sensitive room
func (s *RoomService) GetRoomHistory(ctx context.Context, roomID uuid.UUID) ([]model.Message, error) {
	room, err := s.GetRoomByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	// Only allow access to sensitive rooms
	if !room.IsSensitive {
		return nil, errors.New("room history not available for non-sensitive rooms")
	}

	var messages []model.Message
	if err := s.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		return nil, err
	}

	return messages, nil
}

// cleanupRoom performs post-chat cleanup logic
func (s *RoomService) cleanupRoom(ctx context.Context, roomID uuid.UUID) {
	// Get all messages in the room
	var messages []model.Message
	if err := s.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Find(&messages).Error; err != nil {
		return
	}

	// Analyze messages for sensitive content
	var analyses []*config.MessageAnalysis
	for _, msg := range messages {
		analysis := s.messageAnalyzer.AnalyzeMessage(msg.Content)
		analyses = append(analyses, analysis)
	}

	// Check if room should be retained
	if s.messageAnalyzer.ShouldRetainRoom(analyses) {
		// Mark room as sensitive
		s.UpdateRoom(ctx, roomID, &[]bool{true}[0])
	} else {
		// Delete all messages and delete the room completely
		s.db.WithContext(ctx).Where("room_id = ?", roomID).Delete(&model.Message{})
		s.db.WithContext(ctx).Where("id = ?", roomID).Delete(&model.Room{})
	}
}

// GetRoomsByUserID gets all rooms for a user (for admin purposes)
func (s *RoomService) GetRoomsByUserID(ctx context.Context, userID uuid.UUID) ([]model.Room, error) {
	var rooms []model.Room
	if err := s.db.WithContext(ctx).
		Preload("User1").
		Preload("User2").
		Where("(user1_id = ? OR user2_id = ?) AND is_deleted = false", userID, userID).
		Order("created_at DESC").
		Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}
