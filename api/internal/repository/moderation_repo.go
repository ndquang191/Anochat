package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/moderation"
	"github.com/ndquang191/Anochat/api/internal/model"
	"gorm.io/gorm"
)

// BannedWordRepository defines data access for the banned word list.
type BannedWordRepository interface {
	FindAll(ctx context.Context) ([]*moderation.BannedWord, error)
	Create(ctx context.Context, word *moderation.BannedWord) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type bannedWordRepo struct{ db *gorm.DB }

func NewBannedWordRepository(db *gorm.DB) BannedWordRepository {
	return &bannedWordRepo{db: db}
}

func (r *bannedWordRepo) FindAll(ctx context.Context) ([]*moderation.BannedWord, error) {
	var models []model.BannedWord
	if err := r.db.WithContext(ctx).Order("created_at ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]*moderation.BannedWord, len(models))
	for i := range models {
		result[i] = bannedWordModelToDomain(&models[i])
	}
	return result, nil
}

func (r *bannedWordRepo) Create(ctx context.Context, word *moderation.BannedWord) error {
	m := bannedWordDomainToModel(word)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	word.ID = m.ID
	word.CreatedAt = m.CreatedAt
	return nil
}

func (r *bannedWordRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.BannedWord{}).Error
}

// --- mapping helpers ---

func bannedWordModelToDomain(m *model.BannedWord) *moderation.BannedWord {
	return &moderation.BannedWord{
		ID:        m.ID,
		Word:      m.Word,
		CreatedBy: m.CreatedBy,
		CreatedAt: m.CreatedAt,
	}
}

func bannedWordDomainToModel(w *moderation.BannedWord) *model.BannedWord {
	return &model.BannedWord{
		ID:        w.ID,
		Word:      w.Word,
		CreatedBy: w.CreatedBy,
	}
}

// ReportRepository defines data access for moderation reports.
type ReportRepository interface {
	Create(ctx context.Context, report *moderation.Report) error
	FindAll(ctx context.Context) ([]*moderation.Report, error)
}

type reportRepo struct{ db *gorm.DB }

func NewReportRepository(db *gorm.DB) ReportRepository {
	return &reportRepo{db: db}
}

func (r *reportRepo) Create(ctx context.Context, report *moderation.Report) error {
	m := reportDomainToModel(report)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	report.ID = m.ID
	report.CreatedAt = m.CreatedAt
	return nil
}

func (r *reportRepo) FindAll(ctx context.Context) ([]*moderation.Report, error) {
	var models []model.Report
	if err := r.db.WithContext(ctx).
		Preload("ReportedUser").
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]*moderation.Report, len(models))
	for i := range models {
		result[i] = reportModelToDomain(&models[i])
	}
	return result, nil
}

// --- mapping helpers ---

func reportModelToDomain(m *model.Report) *moderation.Report {
	r := &moderation.Report{
		ID:             m.ID,
		ReporterID:     m.ReporterID,
		ReportedUserID: m.ReportedUserID,
		RoomID:         m.RoomID,
		Status:         m.Status,
		CreatedAt:      m.CreatedAt,
	}
	if m.ReportedUser != nil && m.ReportedUser.Name != nil {
		r.ReportedUserName = m.ReportedUser.Name
	}
	return r
}

func reportDomainToModel(r *moderation.Report) *model.Report {
	return &model.Report{
		ID:             r.ID,
		ReporterID:     r.ReporterID,
		ReportedUserID: r.ReportedUserID,
		RoomID:         r.RoomID,
		Status:         r.Status,
	}
}
