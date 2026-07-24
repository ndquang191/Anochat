package service

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/chat"
	"github.com/ndquang191/Anochat/api/internal/domain/identity"
	"github.com/ndquang191/Anochat/api/internal/domain/moderation"
	"github.com/ndquang191/Anochat/api/internal/repository"
	"github.com/ndquang191/Anochat/api/pkg/apperr"
)

const reportSnapshotMessageLimit = 50

type ModerationService struct {
	bannedWordRepo repository.BannedWordRepository
	reportRepo     repository.ReportRepository
	userRepo       repository.UserRepository
	roomRepo       repository.RoomRepository

	mu          sync.RWMutex
	cachedWords []string // lowercase
}

func NewModerationService(
	bannedWordRepo repository.BannedWordRepository,
	reportRepo repository.ReportRepository,
	userRepo repository.UserRepository,
	roomRepo repository.RoomRepository,
) *ModerationService {
	return &ModerationService{
		bannedWordRepo: bannedWordRepo,
		reportRepo:     reportRepo,
		userRepo:       userRepo,
		roomRepo:       roomRepo,
	}
}

// LoadWords fetches all banned words from DB and populates the in-memory cache.
// Call once at startup.
func (s *ModerationService) LoadWords(ctx context.Context) error {
	words, err := s.bannedWordRepo.FindAll(ctx)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cachedWords = make([]string, len(words))
	for i, w := range words {
		s.cachedWords[i] = strings.ToLower(w.Word)
	}
	slog.Info("Banned words loaded", "count", len(s.cachedWords))
	return nil
}

// ContainsBannedWord returns true if content contains any banned word (case-insensitive).
func (s *ModerationService) ContainsBannedWord(content string) bool {
	lower := strings.ToLower(content)
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, w := range s.cachedWords {
		if strings.Contains(lower, w) {
			return true
		}
	}
	return false
}

// AddWord persists a new banned word and refreshes the cache.
func (s *ModerationService) AddWord(ctx context.Context, word, category string, createdBy uuid.UUID) (*moderation.BannedWord, error) {
	if category == "" {
		category = "General"
	}
	w := &moderation.BannedWord{
		Word:      strings.TrimSpace(word),
		Category:  category,
		CreatedBy: createdBy,
	}
	if err := s.bannedWordRepo.Create(ctx, w); err != nil {
		return nil, err
	}
	if err := s.refreshCache(ctx); err != nil {
		slog.Warn("Failed to refresh banned word cache after add", "error", err)
	}
	return w, nil
}

// UpdateWord changes a word's text and category, then refreshes cache.
func (s *ModerationService) UpdateWord(ctx context.Context, id uuid.UUID, word, category string) error {
	if category == "" {
		category = "General"
	}
	w := &moderation.BannedWord{ID: id, Word: strings.TrimSpace(word), Category: category}
	if err := s.bannedWordRepo.Update(ctx, w); err != nil {
		return err
	}
	return s.refreshCache(ctx)
}

// DeleteWord removes a banned word by ID and refreshes the cache.
func (s *ModerationService) DeleteWord(ctx context.Context, id uuid.UUID) error {
	if err := s.bannedWordRepo.Delete(ctx, id); err != nil {
		return err
	}
	if err := s.refreshCache(ctx); err != nil {
		slog.Warn("Failed to refresh banned word cache after delete", "error", err)
	}
	return nil
}

// ListWords returns all banned words from DB.
func (s *ModerationService) ListWords(ctx context.Context) ([]*moderation.BannedWord, error) {
	return s.bannedWordRepo.FindAll(ctx)
}

// CreateReport validates both participants and snapshots the latest room
// messages before the original chat can be cleaned up.
func (s *ModerationService) CreateReport(ctx context.Context, reporterID, reportedUserID, roomID uuid.UUID) error {
	room, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return apperr.ErrForbidden
	}
	if reporterID == reportedUserID || !room.HasUser(reporterID) || !room.HasUser(reportedUserID) {
		return apperr.ErrForbidden
	}
	return s.createReportWithSnapshot(ctx, reporterID, reportedUserID, roomID, nil, true)
}

// CreateAutoReport snapshots the offending message explicitly. This guarantees
// that moderation evidence contains it even when message persistence and room
// cleanup happen concurrently.
func (s *ModerationService) CreateAutoReport(
	ctx context.Context,
	reportedUserID, roomID uuid.UUID,
	triggerMessage *chat.Message,
) error {
	return s.createReportWithSnapshot(ctx, uuid.Nil, reportedUserID, roomID, triggerMessage, false)
}

func (s *ModerationService) createReportWithSnapshot(
	ctx context.Context,
	reporterID, reportedUserID, roomID uuid.UUID,
	triggerMessage *chat.Message,
	requireRoom bool,
) error {
	r := &moderation.Report{
		ReporterID:     reporterID,
		ReportedUserID: reportedUserID,
		RoomID:         roomID,
		Status:         "pending",
	}
	return s.reportRepo.CreateWithSnapshot(ctx, r, triggerMessage, reportSnapshotMessageLimit, requireRoom)
}

// ListReports returns all reports ordered by newest first.
func (s *ModerationService) ListReports(ctx context.Context) ([]*moderation.Report, error) {
	return s.reportRepo.FindAll(ctx)
}

func (s *ModerationService) GetReportMessages(ctx context.Context, reportID uuid.UUID) ([]*moderation.ReportMessage, error) {
	return s.reportRepo.FindMessages(ctx, reportID)
}

// BanUser suspends an active user, records a new ban occurrence, and clears any
// review request left over from a previous suspension.
func (s *ModerationService) BanUser(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if !user.IsActive {
		return nil
	}
	now := time.Now().UTC()
	user.IsActive = false
	user.BanCount++
	user.BannedAt = &now
	user.ReviewRequested = false
	user.ReviewRequestedAt = nil
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}
	return s.reportRepo.MarkReviewedByUser(ctx, userID)
}

// UnbanUser reactivates a user while preserving their lifetime counters.
func (s *ModerationService) UnbanUser(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	user.IsActive = true
	user.ReviewRequested = false
	user.ReviewRequestedAt = nil
	return s.userRepo.Update(ctx, user)
}

// RequestBanReview records at most one review request for the current ban.
func (s *ModerationService) RequestBanReview(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.IsActive {
		return apperr.ErrForbidden
	}
	if user.ReviewRequested {
		return nil
	}
	now := time.Now().UTC()
	user.ReviewRequested = true
	user.ReviewRequestedAt = &now
	user.ReviewRequestCount++
	return s.userRepo.Update(ctx, user)
}

// ListBannedUsers returns all users with is_active = false.
func (s *ModerationService) ListBannedUsers(ctx context.Context) ([]*identity.User, error) {
	return s.userRepo.FindBanned(ctx)
}

// GetLatestRoomsForUsers returns a map of userID -> latest roomID from reports.
func (s *ModerationService) GetLatestRoomsForUsers(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	return s.reportRepo.FindLatestRoomForUsers(ctx, userIDs)
}

func (s *ModerationService) GetLatestReportsForUsers(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	return s.reportRepo.FindLatestReportForUsers(ctx, userIDs)
}

func (s *ModerationService) refreshCache(ctx context.Context) error {
	words, err := s.bannedWordRepo.FindAll(ctx)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cachedWords = make([]string, len(words))
	for i, w := range words {
		s.cachedWords[i] = strings.ToLower(w.Word)
	}
	return nil
}
