package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/identity"
	"github.com/ndquang191/Anochat/api/internal/repository"
)

var (
	ErrUserNotFound = errors.New("user not found")
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
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*identity.User, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetOrCreateUser(ctx context.Context, email, name, avatarURL string) (*identity.User, error) {
	user, err := s.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return s.CreateUser(ctx, email, name, avatarURL)
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
			return nil, errors.New("profile not found or hidden")
		}
		return nil, err
	}
	return profile, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, isMale *bool, age *int, city *string, isHidden *bool) (*identity.Profile, error) {
	profile, err := s.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	if isMale != nil {
		profile.IsMale = isMale
	}
	if age != nil {
		profile.Age = age
	}
	if city != nil {
		profile.City = city
	}
	if isHidden != nil {
		profile.IsHidden = *isHidden
	}
	profile.UpdatedAt = time.Now()

	if err := s.profileRepo.Update(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}
