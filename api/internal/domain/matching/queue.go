package matching

import (
	"time"

	"github.com/google/uuid"
)

// QueueEntry represents a user waiting in the queue.
type QueueEntry struct {
	UserID   uuid.UUID
	JoinedAt time.Time
}

// MatchNotifier is an interface for notifying when matches are found.
type MatchNotifier interface {
	NotifyMatch(user1ID, user2ID, roomID uuid.UUID)
}
