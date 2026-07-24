package service

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/identity"
	"github.com/ndquang191/Anochat/api/internal/repository"
	"github.com/ndquang191/Anochat/api/pkg/apperr"
	"github.com/ndquang191/Anochat/api/pkg/metrics"
)

const (
	DisplayNameCooldown  = 30 * 24 * time.Hour
	displayNameMinLength = 2
	displayNameMaxLength = 32
)

type UserService struct {
	userRepo    repository.UserRepository
	profileRepo repository.ProfileRepository
}

func NewUserService(userRepo repository.UserRepository, profileRepo repository.ProfileRepository) *UserService {
	return &UserService{userRepo: userRepo, profileRepo: profileRepo}
}

func (s *UserService) CreateUser(ctx context.Context, email, name, avatarURL string) (*identity.User, error) {
	user := &identity.User{
		Email:     &email,
		Name:      &name,
		AvatarURL: &avatarURL,
		IsActive:  true,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetUserByID(ctx context.Context, userID uuid.UUID) (*identity.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperr.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetUserWithProfile(ctx context.Context, userID uuid.UUID) (*identity.User, error) {
	user, err := s.userRepo.FindByIDWithProfile(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperr.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*identity.User, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperr.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetOrCreateUser(ctx context.Context, email, name, avatarURL string) (*identity.User, error) {
	user, err := s.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, apperr.ErrUserNotFound) {
			created, err := s.CreateUser(ctx, email, name, avatarURL)
			if err != nil {
				return nil, err
			}
			metrics.TotalUsers.Inc()
			metrics.NewRegistrations.Inc()
			return created, nil
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetProfile(ctx context.Context, userID uuid.UUID) (*identity.Profile, error) {
	profile, err := s.profileRepo.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			profile = &identity.Profile{
				UserID:   userID,
				IsHidden: false,
			}
			if err := s.profileRepo.Create(ctx, profile); err != nil {
				return nil, err
			}
			return profile, nil
		}
		return nil, err
	}
	return profile, nil
}

func (s *UserService) GetPublicProfile(ctx context.Context, userID uuid.UUID) (*identity.Profile, error) {
	profile, err := s.profileRepo.FindPublicByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperr.ErrProfileNotFound
		}
		return nil, err
	}
	return profile, nil
}

func DisplayNameChangeAvailableAt(profile *identity.Profile, now time.Time) *time.Time {
	if profile == nil || profile.NicknameUpdatedAt == nil {
		return nil
	}
	next := profile.NicknameUpdatedAt.Add(DisplayNameCooldown)
	if !now.Before(next) {
		return nil
	}
	return &next
}

func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, nickname *string, isMale *bool, age *int, isHidden *bool) (*identity.Profile, error) {
	profile, err := s.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	nicknameChanged := false
	if nickname != nil {
		normalized := strings.TrimSpace(*nickname)
		current := ""
		if profile.Nickname != nil {
			current = *profile.Nickname
		}

		if normalized != current {
			length := utf8.RuneCountInString(normalized)
			if normalized != "" && (length < displayNameMinLength || length > displayNameMaxLength) {
				return nil, apperr.ErrInvalidDisplayName
			}
			if DisplayNameChangeAvailableAt(profile, now) != nil {
				return nil, apperr.ErrDisplayNameCooldown
			}

			if normalized == "" {
				profile.Nickname = nil
			} else {
				profile.Nickname = &normalized
			}
			profile.NicknameUpdatedAt = &now
			nicknameChanged = true
		}
	}
	if isMale != nil {
		profile.IsMale = isMale
	}
	if age != nil {
		profile.Age = age
	}
	if isHidden != nil {
		profile.IsHidden = *isHidden
	}
	profile.UpdatedAt = now

	if nicknameChanged {
		err = s.profileRepo.UpdateWithNicknameCooldown(
			ctx,
			profile,
			now.Add(-DisplayNameCooldown),
		)
	} else {
		err = s.profileRepo.Update(ctx, profile)
	}
	if errors.Is(err, repository.ErrNicknameChangeCooldown) {
		return nil, apperr.ErrDisplayNameCooldown
	}
	if err != nil {
		return nil, err
	}
	return profile, nil
}
