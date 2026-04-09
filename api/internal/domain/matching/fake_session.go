package matching

import (
	"time"

	"github.com/google/uuid"
)

// FakeSession represents a temporary synthetic match for low-traffic periods.
type FakeSession struct {
	UserID        uuid.UUID
	RoomID        uuid.UUID
	PartnerID     uuid.UUID
	GreetingID    uuid.UUID
	PartnerName   string
	PartnerIsMale bool
	PartnerAge    int
	Greeting      string
	CreatedAt     time.Time
}
