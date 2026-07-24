package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

// User represents the users table
type User struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email              *string    `gorm:"type:text" json:"email"`
	Name               *string    `gorm:"type:text" json:"name"`
	AvatarURL          *string    `gorm:"type:text;column:avatar_url" json:"avatar_url"`
	IsActive           bool       `gorm:"default:false" json:"is_active"`
	IsAdmin            bool       `gorm:"default:false;not null" json:"is_admin"`
	IsDeleted          bool       `gorm:"default:false" json:"is_deleted"`
	BanCount           int        `gorm:"default:0;not null" json:"ban_count"`
	ReviewRequestCount int        `gorm:"default:0;not null" json:"review_request_count"`
	ReviewRequested    bool       `gorm:"default:false;not null" json:"review_requested"`
	BannedAt           *time.Time `gorm:"type:timestamp" json:"banned_at"`
	ReviewRequestedAt  *time.Time `gorm:"type:timestamp" json:"review_requested_at"`
	CreatedAt          time.Time  `gorm:"autoCreateTime" json:"created_at"`

	// Relationships
	Profile  *Profile  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"profile,omitempty"`
	Messages []Message `gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE" json:"messages,omitempty"`
}

// Profile represents the profiles table
type Profile struct {
	UserID            uuid.UUID  `gorm:"type:uuid;primaryKey" json:"user_id"`
	Nickname          *string    `gorm:"type:text" json:"nickname"`
	NicknameUpdatedAt *time.Time `gorm:"type:timestamptz" json:"nickname_updated_at"`
	IsMale            *bool      `gorm:"type:boolean" json:"is_male"`
	Age               *int       `gorm:"type:integer" json:"age"`
	IsHidden          bool       `gorm:"default:false" json:"is_hidden"`
	UpdatedAt         time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// Room represents the rooms table
type Room struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	User1ID   uuid.UUID  `gorm:"type:uuid;not null" json:"user1_id"`
	User2ID   uuid.UUID  `gorm:"type:uuid;not null" json:"user2_id"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	EndedAt   *time.Time `gorm:"type:timestamp" json:"ended_at"`

	// Relationships
	User1    *User     `gorm:"foreignKey:User1ID;constraint:OnDelete:CASCADE" json:"user1,omitempty"`
	User2    *User     `gorm:"foreignKey:User2ID;constraint:OnDelete:CASCADE" json:"user2,omitempty"`
	Messages []Message `gorm:"foreignKey:RoomID;constraint:OnDelete:CASCADE" json:"messages,omitempty"`
}

// RoomSession keeps non-sensitive lifecycle data after the live room is deleted.
type RoomSession struct {
	RoomID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"room_id"`
	Status         string     `gorm:"type:varchar(16);not null;default:'matched'" json:"status"`
	MatchedAt      time.Time  `gorm:"type:timestamptz;not null" json:"matched_at"`
	ConnectedAt    *time.Time `gorm:"type:timestamptz" json:"connected_at"`
	DisconnectedAt *time.Time `gorm:"type:timestamptz" json:"disconnected_at"`
	EndedAt        *time.Time `gorm:"type:timestamptz" json:"ended_at"`
}

// Message represents the messages table
type Message struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RoomID    uuid.UUID `gorm:"type:uuid;not null" json:"room_id"`
	SenderID  uuid.UUID `gorm:"type:uuid;not null" json:"sender_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relationships
	Room   *Room `gorm:"foreignKey:RoomID;constraint:OnDelete:CASCADE" json:"room,omitempty"`
	Sender *User `gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE" json:"sender,omitempty"`
}

// BannedWord stores words to filter from chat messages.
type BannedWord struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Word      string    `gorm:"type:text;not null;uniqueIndex"`
	Category  string    `gorm:"type:text;not null;default:'General'"`
	CreatedBy uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// Report stores a report made by one user (or system) about another.
type Report struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ReporterID     uuid.UUID `gorm:"type:uuid;not null"`
	ReportedUserID uuid.UUID `gorm:"type:uuid;not null"`
	RoomID         uuid.UUID `gorm:"type:uuid;not null"`
	Status         string    `gorm:"type:text;not null;default:'pending'"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`

	ReportedUser *User `gorm:"foreignKey:ReportedUserID;constraint:OnDelete:CASCADE"`
}

// ReportMessage is an immutable snapshot of a chat message kept as moderation
// evidence after the original room and its messages have been deleted.
type ReportMessage struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ReportID          uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_report_original_message"`
	OriginalMessageID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_report_original_message"`
	SenderID          uuid.UUID `gorm:"type:uuid;not null"`
	Content           string    `gorm:"type:text;not null"`
	CreatedAt         time.Time `gorm:"type:timestamp;not null"`

	Report *Report `gorm:"foreignKey:ReportID;constraint:OnDelete:CASCADE"`
}

// TableName methods to ensure correct table names
func (User) TableName() string {
	return "users"
}

func (Profile) TableName() string {
	return "profiles"
}

func (Room) TableName() string {
	return "rooms"
}

func (RoomSession) TableName() string {
	return "room_sessions"
}

func (Message) TableName() string {
	return "messages"
}

func (BannedWord) TableName() string {
	return "banned_words"
}

func (Report) TableName() string {
	return "reports"
}

func (ReportMessage) TableName() string {
	return "report_messages"
}

// BeforeCreate hooks for setting UUIDs if needed
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (r *Room) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

func (b *BannedWord) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

func (r *Report) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

func (m *ReportMessage) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
