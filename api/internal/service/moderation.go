package service

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/moderation"
	"github.com/ndquang191/Anochat/api/internal/repository"
)

type ModerationService struct {
	bannedWordRepo repository.BannedWordRepository
	reportRepo     repository.ReportRepository
	userRepo       repository.UserRepository

	mu          sync.RWMutex
	cachedWords []string // lowercase
}

func NewModerationService(
	bannedWordRepo repository.BannedWordRepository,
	reportRepo repository.ReportRepository,
	userRepo repository.UserRepository,
) *ModerationService {
	return &ModerationService{
		bannedWordRepo: bannedWordRepo,
		reportRepo:     reportRepo,
		userRepo:       userRepo,
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
func (s *ModerationService) AddWord(ctx context.Context, word string, createdBy uuid.UUID) (*moderation.BannedWord, error) {
	w := &moderation.BannedWord{
		Word:      strings.TrimSpace(word),
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

// CreateReport persists a new moderation report.
// Pass uuid.Nil as reporterID for auto/system-triggered reports.
func (s *ModerationService) CreateReport(ctx context.Context, reporterID, reportedUserID, roomID uuid.UUID) error {
	r := &moderation.Report{
		ReporterID:     reporterID,
		ReportedUserID: reportedUserID,
		RoomID:         roomID,
		Status:         "pending",
	}
	return s.reportRepo.Create(ctx, r)
}

// ListReports returns all reports ordered by newest first.
func (s *ModerationService) ListReports(ctx context.Context) ([]*moderation.Report, error) {
	return s.reportRepo.FindAll(ctx)
}

// BanUser sets is_active = false for a user.
func (s *ModerationService) BanUser(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	user.IsActive = false
	return s.userRepo.Update(ctx, user)
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
