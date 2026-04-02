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
	Nickname  *string
	IsMale    *bool
	Age       *int
	IsHidden  bool
	UpdatedAt time.Time
}
