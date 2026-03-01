package matching

import (
	"time"

	"github.com/google/uuid"
)

// QueueEntry represents a user waiting in the queue.
type QueueEntry struct {
	UserID   uuid.UUID
	JoinedAt time.Time
	MatchChan chan *MatchResult
}

// MatchResult represents the result of a successful match.
type MatchResult struct {
	RoomID  uuid.UUID
	User1ID uuid.UUID
	User2ID uuid.UUID
	Error   error
}

// MatchNotifier is an interface for notifying when matches are found.
type MatchNotifier interface {
	NotifyMatch(user1ID, user2ID, roomID uuid.UUID)
}
