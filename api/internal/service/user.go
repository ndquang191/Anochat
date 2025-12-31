package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/model"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

// UserService handles user and profile operations
type UserService struct {
	db *gorm.DB
}

// NewUserService creates a new user service
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// CreateUser creates a new user from Google OAuth data
func (s *UserService) CreateUser(ctx context.Context, email, name, avatarURL string) (*model.User, error) {
	user := &model.User{
		Email:     &email,
		Name:      &name,
		AvatarURL: &avatarURL,
		IsActive:  true,
	}

	if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	var user model.User
	if err := s.db.WithContext(ctx).Where("id = ? AND is_deleted = false", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	if err := s.db.WithContext(ctx).Where("email = ? AND is_deleted = false", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetOrCreateUser gets existing user or creates new one
func (s *UserService) GetOrCreateUser(ctx context.Context, email, name, avatarURL string) (*model.User, error) {
	user, err := s.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return s.CreateUser(ctx, email, name, avatarURL)
		}
		return nil, err
	}
	return user, nil
}

// Profile operations

// GetProfile retrieves a user's profile
func (s *UserService) GetProfile(ctx context.Context, userID uuid.UUID) (*model.Profile, error) {
	var profile model.Profile
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create default profile if not exists
			profile = model.Profile{
				UserID:   userID,
				IsHidden: false,
			}
			if err := s.db.WithContext(ctx).Create(&profile).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return &profile, nil
}

// GetPublicProfile retrieves a public profile (respecting is_hidden setting)
func (s *UserService) GetPublicProfile(ctx context.Context, userID uuid.UUID) (*model.Profile, error) {
	var profile model.Profile
	if err := s.db.WithContext(ctx).Where("user_id = ? AND is_hidden = false", userID).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("profile not found or hidden")
		}
		return nil, err
	}
	return &profile, nil
}

// UpdateProfile updates a user's profile
func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, isMale *bool, age *int, city *string, isHidden *bool) (*model.Profile, error) {
	profile, err := s.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
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

	if err := s.db.WithContext(ctx).Save(profile).Error; err != nil {
		return nil, err
	}

	return profile, nil
}

// GetUserWithProfile retrieves user with profile information
func (s *UserService) GetUserWithProfile(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	var user model.User
	if err := s.db.WithContext(ctx).
		Preload("Profile").
		Where("id = ? AND is_deleted = false", userID).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetActiveRoom gets the active room for a user
func (s *UserService) GetActiveRoom(ctx context.Context, userID uuid.UUID) (*model.Room, error) {
	var room model.Room
	if err := s.db.WithContext(ctx).
		Preload("User1").
		Preload("User2").
		Where("(user1_id = ? OR user2_id = ?) AND ended_at IS NULL AND is_deleted = false", userID, userID).
		First(&room).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Not an error - user simply has no active room
			return nil, nil
		}
		return nil, err
	}
	return &room, nil
}

// GetRoomMessages gets all messages for a room
func (s *UserService) GetRoomMessages(ctx context.Context, roomID uuid.UUID) ([]model.Message, error) {
	var messages []model.Message
	if err := s.db.WithContext(ctx).
		Preload("Sender").
		Where("room_id = ?", roomID).
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}
