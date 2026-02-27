package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/identity"
	"github.com/ndquang191/Anochat/api/internal/model"
	"gorm.io/gorm"
)

// ProfileRepository defines data access for profiles.
type ProfileRepository interface {
	FindByUserID(ctx context.Context, userID uuid.UUID) (*identity.Profile, error)
	FindPublicByUserID(ctx context.Context, userID uuid.UUID) (*identity.Profile, error)
	Create(ctx context.Context, profile *identity.Profile) error
	Update(ctx context.Context, profile *identity.Profile) error
}

type profileRepo struct{ db *gorm.DB }

func NewProfileRepository(db *gorm.DB) ProfileRepository {
	return &profileRepo{db: db}
}

func (r *profileRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*identity.Profile, error) {
	var m model.Profile
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return profileModelToDomain(&m), nil
}

func (r *profileRepo) FindPublicByUserID(ctx context.Context, userID uuid.UUID) (*identity.Profile, error) {
	var m model.Profile
	if err := r.db.WithContext(ctx).Where("user_id = ? AND is_hidden = false", userID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return profileModelToDomain(&m), nil
}

func (r *profileRepo) Create(ctx context.Context, profile *identity.Profile) error {
	m := profileDomainToModel(profile)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	profile.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *profileRepo) Update(ctx context.Context, profile *identity.Profile) error {
	m := profileDomainToModel(profile)
	return r.db.WithContext(ctx).Save(m).Error
}

// --- mapping helpers ---

func profileModelToDomain(m *model.Profile) *identity.Profile {
	return &identity.Profile{
		UserID:    m.UserID,
		IsMale:    m.IsMale,
		Age:       m.Age,
		City:      m.City,
		IsHidden:  m.IsHidden,
		UpdatedAt: m.UpdatedAt,
	}
}

func profileDomainToModel(p *identity.Profile) *model.Profile {
	return &model.Profile{
		UserID:    p.UserID,
		IsMale:    p.IsMale,
		Age:       p.Age,
		City:      p.City,
		IsHidden:  p.IsHidden,
		UpdatedAt: p.UpdatedAt,
	}
}
