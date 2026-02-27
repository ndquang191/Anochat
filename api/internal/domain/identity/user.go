package identity

import (
	"time"

	"github.com/google/uuid"
)

// User represents a registered user.
type User struct {
	ID        uuid.UUID
	Email     *string
	Name      *string
	AvatarURL *string
	IsActive  bool
	IsDeleted bool
	CreatedAt time.Time
	Profile   *Profile
}

// Profile represents user demographic information.
type Profile struct {
	UserID    uuid.UUID
	IsMale    *bool
	Age       *int
	City      *string
	IsHidden  bool
	UpdatedAt time.Time
}
