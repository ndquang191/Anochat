package chat

import (
	"time"

	"github.com/google/uuid"
)

// Room represents a 1-on-1 chat session.
type Room struct {
	ID                     uuid.UUID
	User1ID                uuid.UUID
	User2ID                uuid.UUID
	Category               string
	CreatedAt              time.Time
	EndedAt                *time.Time
	IsSensitive            bool
	User1LastReadMessageID *uuid.UUID
	User2LastReadMessageID *uuid.UUID
	IsDeleted              bool
}

// IsActive returns true if the room has not ended and is not deleted.
func (r *Room) IsActive() bool {
	return r.EndedAt == nil && !r.IsDeleted
}

// HasUser returns true if the given user is a participant.
func (r *Room) HasUser(userID uuid.UUID) bool {
	return r.User1ID == userID || r.User2ID == userID
}
