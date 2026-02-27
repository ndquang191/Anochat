package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/identity"
	"github.com/ndquang191/Anochat/api/internal/model"
	"gorm.io/gorm"
)

// UserRepository defines data access for users.
type UserRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*identity.User, error)
	FindByEmail(ctx context.Context, email string) (*identity.User, error)
	Create(ctx context.Context, user *identity.User) error
	Update(ctx context.Context, user *identity.User) error
}

type userRepo struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) FindByID(ctx context.Context, id uuid.UUID) (*identity.User, error) {
	var m model.User
	if err := r.db.WithContext(ctx).Where("id = ? AND is_deleted = false", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return userModelToDomain(&m), nil
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (*identity.User, error) {
	var m model.User
	if err := r.db.WithContext(ctx).Where("email = ? AND is_deleted = false", email).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return userModelToDomain(&m), nil
}

func (r *userRepo) Create(ctx context.Context, user *identity.User) error {
	m := userDomainToModel(user)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	user.ID = m.ID
	user.CreatedAt = m.CreatedAt
	return nil
}

func (r *userRepo) Update(ctx context.Context, user *identity.User) error {
	m := userDomainToModel(user)
	return r.db.WithContext(ctx).Save(m).Error
}

// --- mapping helpers ---

func userModelToDomain(m *model.User) *identity.User {
	u := &identity.User{
		ID:        m.ID,
		Email:     m.Email,
		Name:      m.Name,
		AvatarURL: m.AvatarURL,
		IsActive:  m.IsActive,
		IsDeleted: m.IsDeleted,
		CreatedAt: m.CreatedAt,
	}
	if m.Profile != nil {
		u.Profile = profileModelToDomain(m.Profile)
	}
	return u
}

func userDomainToModel(u *identity.User) *model.User {
	return &model.User{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		AvatarURL: u.AvatarURL,
		IsActive:  u.IsActive,
		IsDeleted: u.IsDeleted,
		CreatedAt: u.CreatedAt,
	}
}
