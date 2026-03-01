package dto

// UpdateProfileRequest is the body for PUT /profile.
type UpdateProfileRequest struct {
	Age      *int    `json:"age"`
	City     *string `json:"city"`
	IsMale   *bool   `json:"is_male"`
	IsHidden *bool   `json:"is_hidden"`
}

// UserStateResponse is returned by GET /user/state.
type UserStateResponse struct {
	User      UserDTO      `json:"user"`
	Room      *RoomDTO     `json:"room"`
	Messages  []MessageDTO `json:"messages"`
	IsNewUser bool         `json:"is_new_user"`
	InQueue   bool         `json:"in_queue"`
}

// UserDTO is the user data in API responses.
type UserDTO struct {
	ID        string      `json:"id"`
	Email     *string     `json:"email"`
	Name      *string     `json:"name"`
	AvatarURL *string     `json:"avatar_url"`
	Profile   *ProfileDTO `json:"profile,omitempty"`
}

// ProfileDTO is the profile data in API responses.
type ProfileDTO struct {
	Age      *int    `json:"age"`
	City     *string `json:"city"`
	IsMale   *bool   `json:"is_male"`
	IsHidden bool    `json:"is_hidden"`
}

// RoomDTO is the room data in API responses.
type RoomDTO struct {
	ID      string   `json:"id"`
	User1ID string   `json:"user1_id"`
	User2ID string   `json:"user2_id"`
	Partner *UserDTO `json:"partner,omitempty"`
}

// MessageDTO is the message data in API responses.
type MessageDTO struct {
	ID        string `json:"id"`
	RoomID    string `json:"room_id"`
	SenderID  string `json:"sender_id"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"`
}
