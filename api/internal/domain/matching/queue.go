package matching

import (
	"time"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/identity"
)

// Gender represents user gender for queue matching.
type Gender int

const (
	GenderUnknown Gender = iota
	GenderMale
	GenderFemale
)

// GenderFromProfile derives the Gender from a profile's IsMale field.
func GenderFromProfile(p *identity.Profile) Gender {
	if p == nil || p.IsMale == nil {
		return GenderUnknown
	}
	if *p.IsMale {
		return GenderMale
	}
	return GenderFemale
}

func (g Gender) String() string {
	switch g {
	case GenderMale:
		return "male"
	case GenderFemale:
		return "female"
	default:
		return "unknown"
	}
}

// QueueEntry represents a user waiting in the queue.
type QueueEntry struct {
	UserID    uuid.UUID
	Profile   *identity.Profile
	Gender    Gender
	Category  string
	JoinedAt  time.Time
	ExpiresAt time.Time
	IsMatched bool
	MatchChan chan *MatchResult
}

// MatchResult represents the result of a successful match.
type MatchResult struct {
	RoomID   uuid.UUID
	User1ID  uuid.UUID
	User2ID  uuid.UUID
	Category string
	Error    error
}

// QueueStatus represents the status of a user in the queue.
type QueueStatus struct {
	IsInQueue bool      `json:"is_in_queue"`
	Position  int       `json:"position"`
	Category  string    `json:"category"`
	JoinedAt  time.Time `json:"joined_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// QueuePosition tracks user position in queue.
type QueuePosition struct {
	UserID    uuid.UUID
	Position  int
	Category  string
	Gender    Gender
	JoinedAt  time.Time
	ExpiresAt time.Time
}

// MatchNotifier is an interface for notifying when matches are found.
type MatchNotifier interface {
	NotifyMatch(user1ID, user2ID, roomID uuid.UUID, category string)
}
