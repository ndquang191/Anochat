package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/chat"
	"github.com/ndquang191/Anochat/api/internal/repository"
)

const (
	DefaultMessagePageSize = 50
	MaxMessagePageSize     = 100
)

type MessagePage struct {
	Messages   []*chat.Message
	NextCursor *chat.MessageCursor
	HasMore    bool
}

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

// GetMessagePage returns messages in chronological order. The repository reads
// newest-first so it can efficiently seek backwards from the supplied cursor.
func (s *MessageService) GetMessagePage(
	ctx context.Context,
	roomID uuid.UUID,
	before *chat.MessageCursor,
	limit int,
) (*MessagePage, error) {
	if limit <= 0 {
		limit = DefaultMessagePageSize
	}
	if limit > MaxMessagePageSize {
		limit = MaxMessagePageSize
	}

	messages, err := s.messageRepo.FindPageByRoomID(ctx, roomID, before, limit+1)
	if err != nil {
		return nil, err
	}

	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}

	var nextCursor *chat.MessageCursor
	if hasMore && len(messages) > 0 {
		oldest := messages[len(messages)-1]
		nextCursor = &chat.MessageCursor{ID: oldest.ID, CreatedAt: oldest.CreatedAt}
	}

	for left, right := 0, len(messages)-1; left < right; left, right = left+1, right-1 {
		messages[left], messages[right] = messages[right], messages[left]
	}

	return &MessagePage{
		Messages:   messages,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}
