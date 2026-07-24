package dto

// CreateReportRequest is the body for POST /report.
type CreateReportRequest struct {
	ReportedUserID string `json:"reported_user_id" binding:"required"`
	RoomID         string `json:"room_id"          binding:"required"`
}

// AddBannedWordRequest is the body for POST /admin/words.
type AddBannedWordRequest struct {
	Word     string `json:"word"     binding:"required"`
	Category string `json:"category"`
}

// UpdateBannedWordRequest is the body for PUT /admin/words/:id.
type UpdateBannedWordRequest struct {
	Word     string `json:"word"     binding:"required"`
	Category string `json:"category"`
}

// UpdateProfileRequest is the body for PUT /profile.
type UpdateProfileRequest struct {
	Nickname *string `json:"nickname"`
	Age      *int    `json:"age"`
	IsMale   *bool   `json:"is_male"`
	IsHidden *bool   `json:"is_hidden"`
}

// UserStateResponse is returned by GET /user/state.
type UserStateResponse struct {
	Profile            *ProfileDTO  `json:"profile,omitempty"`
	Room               *RoomDTO     `json:"room"`
	Messages           []MessageDTO `json:"messages"`
	InQueue            bool         `json:"in_queue"`
	IsAdmin            bool         `json:"is_admin"`
	IsBanned           bool         `json:"is_banned"`
	BanCount           int          `json:"ban_count"`
	ReviewRequestCount int          `json:"review_request_count"`
	ReviewRequested    bool         `json:"review_requested"`
}

// UserDTO is the user data in API responses.
type UserDTO struct {
	ID       string      `json:"id"`
	Email    *string     `json:"email"`
	Name     *string     `json:"name"`
	Nickname *string     `json:"nickname,omitempty"`
	Profile  *ProfileDTO `json:"profile,omitempty"`
	IsAdmin  bool        `json:"is_admin,omitempty"`
}

// ProfileDTO is the profile data in API responses.
type ProfileDTO struct {
	Nickname                  *string `json:"nickname,omitempty"`
	NicknameChangeAvailableAt *int64  `json:"nickname_change_available_at,omitempty"`
	Age                       *int    `json:"age"`
	IsMale                    *bool   `json:"is_male"`
	IsHidden                  bool    `json:"is_hidden"`
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
	SenderID  string `json:"sender_id"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"`
}
