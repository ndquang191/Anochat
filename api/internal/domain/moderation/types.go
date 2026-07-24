package moderation

import (
	"time"

	"github.com/google/uuid"
)

type BannedWord struct {
	ID        uuid.UUID
	Word      string
	Category  string
	CreatedBy uuid.UUID
	CreatedAt time.Time
}

// Report stores a moderation report. ReporterID == uuid.Nil means system-triggered (auto-report).
type Report struct {
	ID               uuid.UUID
	ReporterID       uuid.UUID
	ReportedUserID   uuid.UUID
	RoomID           uuid.UUID
	Status           string
	CreatedAt        time.Time
	ReportedUserName *string
}

type ReportMessage struct {
	ID                uuid.UUID
	ReportID          uuid.UUID
	OriginalMessageID uuid.UUID
	SenderID          uuid.UUID
	Content           string
	CreatedAt         time.Time
}
