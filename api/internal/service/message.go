package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/chat"
	"github.com/ndquang191/Anochat/api/internal/repository"
)

type MessageService struct {
	messageRepo repository.MessageRepository
}

func NewMessageService(messageRepo repository.MessageRepository) *MessageService {
	return &MessageService{messageRepo: messageRepo}
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
