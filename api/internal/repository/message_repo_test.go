package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/chat"
)

func TestMessageDomainToModelPreservesClientGeneratedFields(t *testing.T) {
	message := &chat.Message{
		ID:        uuid.New(),
		RoomID:    uuid.New(),
		SenderID:  uuid.New(),
		Content:   "hello",
		CreatedAt: time.Unix(123, 456).UTC(),
	}

	model := messageDomainToModel(message)

	if model.ID != message.ID {
		t.Errorf("message id = %s, want %s", model.ID, message.ID)
	}
	if !model.CreatedAt.Equal(message.CreatedAt) {
		t.Errorf("created_at = %s, want %s", model.CreatedAt, message.CreatedAt)
	}
}
