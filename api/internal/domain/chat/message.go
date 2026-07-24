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

// MessageCursor uniquely identifies a position in a room's message timeline.
// ID is the tie-breaker for messages with the same creation timestamp.
type MessageCursor struct {
	ID        uuid.UUID
	CreatedAt time.Time
}
