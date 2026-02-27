package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/chat"
	"github.com/ndquang191/Anochat/api/internal/repository"
)

type MessageService struct {
	messageRepo repository.MessageRepository
	roomRepo    repository.RoomRepository
}

func NewMessageService(messageRepo repository.MessageRepository, roomRepo repository.RoomRepository) *MessageService {
	return &MessageService{messageRepo: messageRepo, roomRepo: roomRepo}
}

func (s *MessageService) CreateMessage(ctx context.Context, roomID, senderID uuid.UUID, content string) (*chat.Message, error) {
	msg := &chat.Message{
		RoomID:   roomID,
		SenderID: senderID,
		Content:  content,
	}
	if err := s.messageRepo.Create(ctx, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (s *MessageService) GetMessagesByRoomID(ctx context.Context, roomID uuid.UUID) ([]*chat.Message, error) {
	return s.messageRepo.FindByRoomID(ctx, roomID)
}

func (s *MessageService) GetUnreadMessagesCount(ctx context.Context, roomID, userID uuid.UUID) (int64, error) {
	room, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return 0, err
	}

	var lastReadMessageID *uuid.UUID
	if room.User1ID == userID {
		lastReadMessageID = room.User1LastReadMessageID
	} else if room.User2ID == userID {
		lastReadMessageID = room.User2LastReadMessageID
	}

	return s.messageRepo.CountUnread(ctx, roomID, userID, lastReadMessageID)
}
