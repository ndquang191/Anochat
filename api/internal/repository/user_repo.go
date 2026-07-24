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
	FindByIDWithProfile(ctx context.Context, id uuid.UUID) (*identity.User, error)
	FindByEmail(ctx context.Context, email string) (*identity.User, error)
	Create(ctx context.Context, user *identity.User) error
	Update(ctx context.Context, user *identity.User) error
	Count(ctx context.Context) (int64, error)
	FindBannedPage(ctx context.Context, query string, before *identity.BannedUserCursor, limit int) ([]*identity.User, error)
	CountBanned(ctx context.Context, query string) (int64, error)
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

func (r *userRepo) FindByIDWithProfile(ctx context.Context, id uuid.UUID) (*identity.User, error) {
	var m model.User
	if err := r.db.WithContext(ctx).Preload("Profile").Where("id = ? AND is_deleted = false", id).First(&m).Error; err != nil {
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

func (r *userRepo) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).Where("is_deleted = false").Count(&count).Error
	return count, err
}

func (r *userRepo) FindBannedPage(
	ctx context.Context,
	query string,
	before *identity.BannedUserCursor,
	limit int,
) ([]*identity.User, error) {
	var models []model.User
	db := r.db.WithContext(ctx).
		Select("id", "email", "name", "is_active", "ban_count", "review_request_count",
			"review_requested", "banned_at", "review_requested_at", "created_at").
		Where("is_active = false AND is_deleted = false").
		Order("review_requested DESC").
		Order("COALESCE(review_requested_at, banned_at, created_at) DESC").
		Order("id DESC")
	if query != "" {
		pattern := "%" + query + "%"
		db = db.Where("name ILIKE ? OR email ILIKE ? OR id::text ILIKE ?", pattern, pattern, pattern)
	}
	if before != nil {
		reviewRequestedRank := 0
		if before.ReviewRequested {
			reviewRequestedRank = 1
		}
		db = db.Where(`
			CASE WHEN review_requested THEN 1 ELSE 0 END < ?
			OR (
				CASE WHEN review_requested THEN 1 ELSE 0 END = ?
				AND (
					COALESCE(review_requested_at, banned_at, created_at) < ?
					OR (
						COALESCE(review_requested_at, banned_at, created_at) = ?
						AND id < ?
					)
				)
			)`,
			reviewRequestedRank,
			reviewRequestedRank,
			before.SortAt,
			before.SortAt,
			before.ID,
		)
	}
	if err := db.Limit(limit).Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]*identity.User, len(models))
	for i := range models {
		result[i] = userModelToDomain(&models[i])
	}
	return result, nil
}

func (r *userRepo) CountBanned(ctx context.Context, query string) (int64, error) {
	db := r.db.WithContext(ctx).Model(&model.User{}).
		Where("is_active = false AND is_deleted = false")
	if query != "" {
		pattern := "%" + query + "%"
		db = db.Where("name ILIKE ? OR email ILIKE ? OR id::text ILIKE ?", pattern, pattern, pattern)
	}
	var count int64
	err := db.Count(&count).Error
	return count, err
}

// --- mapping helpers ---

func userModelToDomain(m *model.User) *identity.User {
	u := &identity.User{
		ID:                 m.ID,
		Email:              m.Email,
		Name:               m.Name,
		AvatarURL:          m.AvatarURL,
		IsActive:           m.IsActive,
		IsAdmin:            m.IsAdmin,
		IsDeleted:          m.IsDeleted,
		BanCount:           m.BanCount,
		ReviewRequestCount: m.ReviewRequestCount,
		ReviewRequested:    m.ReviewRequested,
		BannedAt:           m.BannedAt,
		ReviewRequestedAt:  m.ReviewRequestedAt,
		CreatedAt:          m.CreatedAt,
	}
	if m.Profile != nil {
		u.Profile = profileModelToDomain(m.Profile)
	}
	return u
}

func userDomainToModel(u *identity.User) *model.User {
	return &model.User{
		ID:                 u.ID,
		Email:              u.Email,
		Name:               u.Name,
		AvatarURL:          u.AvatarURL,
		IsActive:           u.IsActive,
		IsAdmin:            u.IsAdmin,
		IsDeleted:          u.IsDeleted,
		BanCount:           u.BanCount,
		ReviewRequestCount: u.ReviewRequestCount,
		ReviewRequested:    u.ReviewRequested,
		BannedAt:           u.BannedAt,
		ReviewRequestedAt:  u.ReviewRequestedAt,
		CreatedAt:          u.CreatedAt,
	}
}
