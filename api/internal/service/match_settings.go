package service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/matching"
	"github.com/ndquang191/Anochat/api/internal/repository"
	"github.com/ndquang191/Anochat/api/pkg/apperr"
)

type MatchSettingsView struct {
	DefaultMode       string
	AllowUserChoice   bool
	UserPreference    *string
	EffectiveMode     string
	QueueDisplayMode  string
	QueueCountMinimum int
	QueueMessageVI    string
	QueueMessageEN    string
}

type MatchSettingsService struct {
	repo        repository.MatchSettingsRepository
	userService *UserService
	queue       *QueueService
}

func NewMatchSettingsService(repo repository.MatchSettingsRepository, userService *UserService, queue *QueueService) *MatchSettingsService {
	return &MatchSettingsService{repo: repo, userService: userService, queue: queue}
}

func (s *MatchSettingsService) GetForUser(ctx context.Context, userID uuid.UUID) (MatchSettingsView, error) {
	settings, err := s.repo.Get(ctx)
	if err != nil {
		return MatchSettingsView{}, err
	}
	profile, err := s.userService.GetProfile(ctx, userID)
	if err != nil {
		return MatchSettingsView{}, err
	}
	return MatchSettingsView{
		DefaultMode:          settings.DefaultMode,
		AllowUserChoice:      settings.AllowUserChoice,
		UserPreference:       profile.MatchPreference,
		EffectiveMode:        settings.EffectiveMode(profile.MatchPreference),
		QueueDisplayMode:     settings.QueueDisplayMode,
		QueueCountMinimum:    settings.QueueCountMinimum,
		QueueMessageVI:       settings.QueueMessageVI,
		QueueMessageEN:       settings.QueueMessageEN,
	}, nil
}

func (s *MatchSettingsService) GetAdmin(ctx context.Context) (matching.Settings, error) {
	return s.repo.Get(ctx)
}

func (s *MatchSettingsService) UpdateUserPreference(ctx context.Context, userID uuid.UUID, mode string) (MatchSettingsView, error) {
	if !matching.ValidMode(mode) {
		return MatchSettingsView{}, apperr.ErrInvalidMatchMode
	}
	settings, err := s.repo.Get(ctx)
	if err != nil {
		return MatchSettingsView{}, err
	}
	if !settings.AllowUserChoice {
		return MatchSettingsView{}, apperr.ErrMatchChoiceDisabled
	}
	profile, err := s.userService.GetProfile(ctx, userID)
	if err != nil {
		return MatchSettingsView{}, err
	}
	profile.MatchPreference = &mode
	if err := s.userService.profileRepo.UpdateMatchPreference(ctx, userID, &mode); err != nil {
		return MatchSettingsView{}, err
	}
	if s.queue != nil {
		s.queue.RefreshQueuedUser(ctx, userID, profile)
	}
	return MatchSettingsView{
		DefaultMode:       settings.DefaultMode,
		AllowUserChoice:   true,
		UserPreference:    &mode,
		EffectiveMode:     mode,
		QueueDisplayMode:  settings.QueueDisplayMode,
		QueueCountMinimum: settings.QueueCountMinimum,
		QueueMessageVI:    settings.QueueMessageVI,
		QueueMessageEN:    settings.QueueMessageEN,
	}, nil
}

func (s *MatchSettingsService) UpdateAdmin(ctx context.Context, settings matching.Settings) error {
	if !matching.ValidMode(settings.DefaultMode) {
		return apperr.ErrInvalidMatchMode
	}
	if !matching.ValidRematchCooldownSeconds(settings.RematchCooldownSeconds) ||
		!matching.ValidQueueDisplayMode(settings.QueueDisplayMode) || settings.QueueCountMinimum < 1 ||
		settings.QueueCountMinimum > 1000 || len([]rune(settings.QueueMessageVI)) > 500 ||
		len([]rune(settings.QueueMessageEN)) > 500 {
		return apperr.ErrInvalidBody
	}
	settings.QueueMessageVI = strings.TrimSpace(settings.QueueMessageVI)
	settings.QueueMessageEN = strings.TrimSpace(settings.QueueMessageEN)
	if err := s.repo.Update(ctx, settings); err != nil {
		return err
	}
	if s.queue != nil {
		return s.queue.ReprocessQueue(ctx)
	}
	return nil
}
