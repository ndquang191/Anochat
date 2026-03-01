package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/chat"
	"github.com/ndquang191/Anochat/api/internal/repository"
)

type RoomService struct {
	roomRepo    repository.RoomRepository
	messageRepo repository.MessageRepository
}

func NewRoomService(roomRepo repository.RoomRepository, messageRepo repository.MessageRepository) *RoomService {
	return &RoomService{
		roomRepo:    roomRepo,
		messageRepo: messageRepo,
	}
}

func (s *RoomService) CreateRoom(ctx context.Context, user1ID, user2ID uuid.UUID) (*chat.Room, error) {
	room := &chat.Room{
		User1ID: user1ID,
		User2ID: user2ID,
	}
	if err := s.roomRepo.Create(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *RoomService) GetRoomByID(ctx context.Context, roomID uuid.UUID) (*chat.Room, error) {
	return s.roomRepo.FindByID(ctx, roomID)
}

func (s *RoomService) GetActiveRoomByUserID(ctx context.Context, userID uuid.UUID) (*chat.Room, error) {
	room, err := s.roomRepo.FindActiveByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, errors.New("no active room found")
		}
		return nil, err
	}
	return room, nil
}

func (s *RoomService) LeaveRoom(ctx context.Context, roomID, userID uuid.UUID) error {
	room, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return err
	}

	if !room.HasUser(userID) {
		return errors.New("user not part of this room")
	}

	if room.EndedAt != nil {
		return nil
	}

	now := time.Now()
	if err := s.roomRepo.UpdateEndedAt(ctx, roomID, now); err != nil {
		return err
	}

	go s.cleanupRoom(context.Background(), roomID)
	return nil
}

func (s *RoomService) LeaveCurrentRoom(ctx context.Context, userID uuid.UUID) error {
	room, err := s.GetActiveRoomByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("no active room found: %w", err)
	}
	return s.LeaveRoom(ctx, room.ID, userID)
}

func (s *RoomService) cleanupRoom(ctx context.Context, roomID uuid.UUID) {
	slog.Info("Starting room cleanup", "room_id", roomID)

	if err := s.messageRepo.DeleteByRoomID(ctx, roomID); err != nil {
		slog.Error("Failed to delete messages", "room_id", roomID, "error", err)
		return
	}
	if err := s.roomRepo.Delete(ctx, roomID); err != nil {
		slog.Error("Failed to delete room", "room_id", roomID, "error", err)
	}
}
