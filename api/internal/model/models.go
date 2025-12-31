package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

// User represents the users table
type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email     *string   `gorm:"type:text" json:"email"`
	Name      *string   `gorm:"type:text" json:"name"`
	AvatarURL *string   `gorm:"type:text;column:avatar_url" json:"avatar_url"`
	IsActive  bool      `gorm:"default:false" json:"is_active"`
	IsDeleted bool      `gorm:"default:false" json:"is_deleted"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relationships
	Profile  *Profile  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"profile,omitempty"`
	Messages []Message `gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE" json:"messages,omitempty"`
}

// Profile represents the profiles table
type Profile struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	IsMale    *bool     `gorm:"type:boolean" json:"is_male"`
	Age       *int      `gorm:"type:integer" json:"age"`
	City      *string   `gorm:"type:text" json:"city"`
	IsHidden  bool      `gorm:"default:false" json:"is_hidden"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// Room represents the rooms table
type Room struct {
	ID                     uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	User1ID                uuid.UUID  `gorm:"type:uuid;not null" json:"user1_id"`
	User2ID                uuid.UUID  `gorm:"type:uuid;not null" json:"user2_id"`
	Category               string     `gorm:"type:text;default:polite" json:"category"`
	CreatedAt              time.Time  `gorm:"autoCreateTime" json:"created_at"`
	EndedAt                *time.Time `gorm:"type:timestamp" json:"ended_at"`
	IsSensitive            bool       `gorm:"default:false" json:"is_sensitive"`
	User1LastReadMessageID *uuid.UUID `gorm:"type:uuid" json:"user1_last_read_message_id"`
	User2LastReadMessageID *uuid.UUID `gorm:"type:uuid" json:"user2_last_read_message_id"`
	IsDeleted              bool       `gorm:"default:false" json:"is_deleted"`

	// Relationships
	User1    *User     `gorm:"foreignKey:User1ID;constraint:OnDelete:CASCADE" json:"user1,omitempty"`
	User2    *User     `gorm:"foreignKey:User2ID;constraint:OnDelete:CASCADE" json:"user2,omitempty"`
	Messages []Message `gorm:"foreignKey:RoomID;constraint:OnDelete:CASCADE" json:"messages,omitempty"`
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

func (Message) TableName() string {
	return "messages"
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
