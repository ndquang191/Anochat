package service

import (
	"context"
	"time"

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

func (s *MessageService) SaveMessage(ctx context.Context, id, roomID, senderID uuid.UUID, content string, createdAt time.Time) error {
	msg := &chat.Message{
		ID:        id,
		RoomID:    roomID,
		SenderID:  senderID,
		Content:   content,
		CreatedAt: createdAt,
	}
	return s.messageRepo.Create(ctx, msg)
}

func (s *MessageService) GetMessagesByRoomID(ctx context.Context, roomID uuid.UUID) ([]*chat.Message, error) {
	return s.messageRepo.FindByRoomID(ctx, roomID)
}
