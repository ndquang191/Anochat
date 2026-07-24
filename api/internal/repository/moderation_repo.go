package repository

import (
	"context"
	"sort"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/chat"
	"github.com/ndquang191/Anochat/api/internal/domain/moderation"
	"github.com/ndquang191/Anochat/api/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	CreateWithSnapshot(ctx context.Context, report *moderation.Report, triggerMessage *chat.Message, limit int, requireRoom bool) error
	FindAll(ctx context.Context) ([]*moderation.Report, error)
	FindMessages(ctx context.Context, reportID uuid.UUID) ([]*moderation.ReportMessage, error)
	MarkReviewedByUser(ctx context.Context, reportedUserID uuid.UUID) error
	FindLatestRoomForUsers(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]uuid.UUID, error)
	FindLatestReportForUsers(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]uuid.UUID, error)
}

type reportRepo struct{ db *gorm.DB }

func NewReportRepository(db *gorm.DB) ReportRepository {
	return &reportRepo{db: db}
}

func (r *reportRepo) CreateWithSnapshot(
	ctx context.Context,
	report *moderation.Report,
	triggerMessage *chat.Message,
	limit int,
	requireRoom bool,
) error {
	if limit <= 0 {
		limit = 1
	}

	reportModel := reportDomainToModel(report)
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if requireRoom {
			var room model.Room
			if err := tx.
				Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("id = ?", report.RoomID).
				First(&room).Error; err != nil {
				return err
			}
		}

		if err := tx.Create(reportModel).Error; err != nil {
			return err
		}

		var sourceMessages []model.Message
		if err := tx.
			Where("room_id = ?", report.RoomID).
			Order("created_at DESC").
			Limit(limit).
			Find(&sourceMessages).Error; err != nil {
			return err
		}

		evidenceByOriginalID := make(map[uuid.UUID]model.ReportMessage, len(sourceMessages)+1)
		for i := range sourceMessages {
			message := sourceMessages[i]
			evidenceByOriginalID[message.ID] = model.ReportMessage{
				ReportID:          reportModel.ID,
				OriginalMessageID: message.ID,
				SenderID:          message.SenderID,
				Content:           message.Content,
				CreatedAt:         message.CreatedAt,
			}
		}
		if triggerMessage != nil {
			evidenceByOriginalID[triggerMessage.ID] = model.ReportMessage{
				ReportID:          reportModel.ID,
				OriginalMessageID: triggerMessage.ID,
				SenderID:          triggerMessage.SenderID,
				Content:           triggerMessage.Content,
				CreatedAt:         triggerMessage.CreatedAt,
			}
		}

		evidence := make([]model.ReportMessage, 0, len(evidenceByOriginalID))
		for _, message := range evidenceByOriginalID {
			evidence = append(evidence, message)
		}
		sort.Slice(evidence, func(i, j int) bool {
			return evidence[i].CreatedAt.Before(evidence[j].CreatedAt)
		})
		if len(evidence) > limit {
			evidence = evidence[len(evidence)-limit:]
		}
		if len(evidence) > 0 {
			if err := tx.Create(&evidence).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	report.ID = reportModel.ID
	report.CreatedAt = reportModel.CreatedAt
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

func (r *reportRepo) FindMessages(ctx context.Context, reportID uuid.UUID) ([]*moderation.ReportMessage, error) {
	var models []model.ReportMessage
	if err := r.db.WithContext(ctx).
		Where("report_id = ?", reportID).
		Order("created_at ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]*moderation.ReportMessage, len(models))
	for i := range models {
		result[i] = reportMessageModelToDomain(&models[i])
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

func (r *reportRepo) FindLatestReportForUsers(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	if len(userIDs) == 0 {
		return map[uuid.UUID]uuid.UUID{}, nil
	}
	type row struct {
		ReportedUserID uuid.UUID
		ReportID       uuid.UUID
	}
	var rows []row
	err := r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT ON (reported_user_id) reported_user_id, id AS report_id
		FROM reports
		WHERE reported_user_id = ANY(?)
		ORDER BY reported_user_id, created_at DESC
	`, userIDs).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]uuid.UUID, len(rows))
	for _, row := range rows {
		result[row.ReportedUserID] = row.ReportID
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

func reportMessageModelToDomain(m *model.ReportMessage) *moderation.ReportMessage {
	return &moderation.ReportMessage{
		ID:                m.ID,
		ReportID:          m.ReportID,
		OriginalMessageID: m.OriginalMessageID,
		SenderID:          m.SenderID,
		Content:           m.Content,
		CreatedAt:         m.CreatedAt,
	}
}
