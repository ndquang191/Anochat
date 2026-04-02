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
	Update(ctx context.Context, word *moderation.BannedWord) error
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

func (r *bannedWordRepo) Update(ctx context.Context, word *moderation.BannedWord) error {
	category := word.Category
	if category == "" {
		category = "General"
	}
	return r.db.WithContext(ctx).Model(&model.BannedWord{}).
		Where("id = ?", word.ID).
		Updates(map[string]any{"word": word.Word, "category": category}).Error
}

func (r *bannedWordRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.BannedWord{}).Error
}

// --- mapping helpers ---

func bannedWordModelToDomain(m *model.BannedWord) *moderation.BannedWord {
	return &moderation.BannedWord{
		ID:        m.ID,
		Word:      m.Word,
		Category:  m.Category,
		CreatedBy: m.CreatedBy,
		CreatedAt: m.CreatedAt,
	}
}

func bannedWordDomainToModel(w *moderation.BannedWord) *model.BannedWord {
	category := w.Category
	if category == "" {
		category = "General"
	}
	return &model.BannedWord{
		ID:        w.ID,
		Word:      w.Word,
		Category:  category,
		CreatedBy: w.CreatedBy,
	}
}

// ReportRepository defines data access for moderation reports.
type ReportRepository interface {
	Create(ctx context.Context, report *moderation.Report) error
	FindAll(ctx context.Context) ([]*moderation.Report, error)
	MarkReviewedByUser(ctx context.Context, reportedUserID uuid.UUID) error
	FindLatestRoomForUsers(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]uuid.UUID, error)
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

func (r *reportRepo) FindLatestRoomForUsers(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	if len(userIDs) == 0 {
		return map[uuid.UUID]uuid.UUID{}, nil
	}
	type row struct {
		ReportedUserID uuid.UUID
		RoomID         uuid.UUID
	}
	var rows []row
	err := r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT ON (reported_user_id) reported_user_id, room_id
		FROM reports
		WHERE reported_user_id = ANY(?)
		ORDER BY reported_user_id, created_at DESC
	`, userIDs).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]uuid.UUID, len(rows))
	for _, row := range rows {
		result[row.ReportedUserID] = row.RoomID
	}
	return result, nil
}

func (r *reportRepo) MarkReviewedByUser(ctx context.Context, reportedUserID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.Report{}).
		Where("reported_user_id = ? AND status = 'pending'", reportedUserID).
		Update("status", "reviewed").Error
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
