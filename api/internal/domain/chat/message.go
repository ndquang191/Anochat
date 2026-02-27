package chat

import (
	"time"

	"github.com/google/uuid"
)

// Message represents a chat message within a room.
type Message struct {
	ID        uuid.UUID
	RoomID    uuid.UUID
	SenderID  uuid.UUID
	Content   string
	CreatedAt time.Time
}
