package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/chat"
	"github.com/ndquang191/Anochat/api/internal/domain/identity"
	"github.com/ndquang191/Anochat/api/internal/domain/moderation"
	"github.com/ndquang191/Anochat/api/pkg/apperr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newModerationService(bwRepo *mockBannedWordRepo, rRepo *mockReportRepo, uRepo *mockUserRepo) *ModerationService {
	return NewModerationService(bwRepo, rRepo, uRepo, new(mockRoomRepo))
}

func TestContainsBannedWord(t *testing.T) {
	svc := &ModerationService{
		cachedWords: []string{"badword", "spam"},
	}

	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{"no match", "hello world", false},
		{"exact match", "this is badword here", true},
		{"case insensitive", "This is BADWORD here", true},
		{"partial match (contains)", "mybadwordfoo", true},
		{"second word match", "this is spam", true},
		{"empty content", "", false},
		{"empty cache", "anything", false},
	}

	// Test with empty cache separately
	emptySvc := &ModerationService{cachedWords: []string{}}
	assert.False(t, emptySvc.ContainsBannedWord("anything"))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "empty cache" {
				return // tested above
			}
			assert.Equal(t, tt.expected, svc.ContainsBannedWord(tt.content))
		})
	}
}

func TestLoadWords(t *testing.T) {
	bwRepo := new(mockBannedWordRepo)
	svc := newModerationService(bwRepo, new(mockReportRepo), new(mockUserRepo))

	words := []*moderation.BannedWord{
		{ID: uuid.New(), Word: "BadWord"},
		{ID: uuid.New(), Word: "SPAM"},
	}
	bwRepo.On("FindAll", mock.Anything).Return(words, nil)

	err := svc.LoadWords(context.Background())
	assert.NoError(t, err)
	assert.True(t, svc.ContainsBannedWord("badword"))
	assert.True(t, svc.ContainsBannedWord("spam"))
	assert.False(t, svc.ContainsBannedWord("hello"))
	bwRepo.AssertExpectations(t)
}

func TestAddWord(t *testing.T) {
	bwRepo := new(mockBannedWordRepo)
	svc := newModerationService(bwRepo, new(mockReportRepo), new(mockUserRepo))
	createdBy := uuid.New()

	bwRepo.On("Create", mock.Anything, mock.AnythingOfType("*moderation.BannedWord")).Return(nil)
	bwRepo.On("FindAll", mock.Anything).Return([]*moderation.BannedWord{
		{Word: "newword"},
	}, nil)

	word, err := svc.AddWord(context.Background(), " newword ", "Profanity", createdBy)
	assert.NoError(t, err)
	assert.Equal(t, "newword", word.Word)
	assert.Equal(t, "Profanity", word.Category)
	bwRepo.AssertExpectations(t)
}

func TestAddWord_DefaultCategory(t *testing.T) {
	bwRepo := new(mockBannedWordRepo)
	svc := newModerationService(bwRepo, new(mockReportRepo), new(mockUserRepo))

	bwRepo.On("Create", mock.Anything, mock.MatchedBy(func(w *moderation.BannedWord) bool {
		return w.Category == "General"
	})).Return(nil)
	bwRepo.On("FindAll", mock.Anything).Return([]*moderation.BannedWord{}, nil)

	_, err := svc.AddWord(context.Background(), "test", "", uuid.New())
	assert.NoError(t, err)
	bwRepo.AssertExpectations(t)
}

func TestDeleteWord(t *testing.T) {
	bwRepo := new(mockBannedWordRepo)
	svc := newModerationService(bwRepo, new(mockReportRepo), new(mockUserRepo))
	wordID := uuid.New()

	bwRepo.On("Delete", mock.Anything, wordID).Return(nil)
	bwRepo.On("FindAll", mock.Anything).Return([]*moderation.BannedWord{}, nil)

	err := svc.DeleteWord(context.Background(), wordID)
	assert.NoError(t, err)
	bwRepo.AssertExpectations(t)
}

func TestCreateReport(t *testing.T) {
	reportRepo := new(mockReportRepo)
	roomRepo := new(mockRoomRepo)
	svc := NewModerationService(new(mockBannedWordRepo), reportRepo, new(mockUserRepo), roomRepo)
	reporterID := uuid.New()
	reportedUserID := uuid.New()
	roomID := uuid.New()

	roomRepo.On("FindByID", mock.Anything, roomID).Return(&chat.Room{
		ID:      roomID,
		User1ID: reporterID,
		User2ID: reportedUserID,
	}, nil)
	reportRepo.On(
		"CreateWithSnapshot",
		mock.Anything,
		mock.MatchedBy(func(r *moderation.Report) bool {
			return r.Status == "pending" &&
				r.ReporterID == reporterID &&
				r.ReportedUserID == reportedUserID
		}),
		(*chat.Message)(nil),
		reportSnapshotMessageLimit,
		true,
	).Return(nil)

	err := svc.CreateReport(context.Background(), reporterID, reportedUserID, roomID)
	assert.NoError(t, err)
	roomRepo.AssertExpectations(t)
	reportRepo.AssertExpectations(t)
}

func TestCreateAutoReportIncludesTriggerMessage(t *testing.T) {
	reportRepo := new(mockReportRepo)
	svc := newModerationService(new(mockBannedWordRepo), reportRepo, new(mockUserRepo))
	trigger := &chat.Message{
		ID:        uuid.New(),
		RoomID:    uuid.New(),
		SenderID:  uuid.New(),
		Content:   "banned content",
		CreatedAt: time.Now(),
	}

	reportRepo.On(
		"CreateWithSnapshot",
		mock.Anything,
		mock.MatchedBy(func(r *moderation.Report) bool {
			return r.ReporterID == uuid.Nil &&
				r.ReportedUserID == trigger.SenderID &&
				r.RoomID == trigger.RoomID
		}),
		trigger,
		reportSnapshotMessageLimit,
		false,
	).Return(nil)

	err := svc.CreateAutoReport(context.Background(), trigger.SenderID, trigger.RoomID, trigger)
	assert.NoError(t, err)
	reportRepo.AssertExpectations(t)
}

func TestCreateReportRejectsNonParticipant(t *testing.T) {
	reportRepo := new(mockReportRepo)
	roomRepo := new(mockRoomRepo)
	svc := NewModerationService(new(mockBannedWordRepo), reportRepo, new(mockUserRepo), roomRepo)
	roomID := uuid.New()
	reportedUserID := uuid.New()

	roomRepo.On("FindByID", mock.Anything, roomID).Return(&chat.Room{
		ID:      roomID,
		User1ID: uuid.New(),
		User2ID: reportedUserID,
	}, nil)

	err := svc.CreateReport(context.Background(), uuid.New(), reportedUserID, roomID)
	assert.ErrorIs(t, err, apperr.ErrForbidden)
	reportRepo.AssertNotCalled(t, "CreateWithSnapshot", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestBanUser(t *testing.T) {
	userRepo := new(mockUserRepo)
	reportRepo := new(mockReportRepo)
	svc := newModerationService(new(mockBannedWordRepo), reportRepo, userRepo)
	userID := uuid.New()

	user := &identity.User{ID: userID, IsActive: true}
	userRepo.On("FindByID", mock.Anything, userID).Return(user, nil)
	userRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *identity.User) bool {
		return u.ID == userID &&
			!u.IsActive &&
			u.BanCount == 1 &&
			u.BannedAt != nil &&
			!u.ReviewRequested
	})).Return(nil)
	reportRepo.On("MarkReviewedByUser", mock.Anything, userID).Return(nil)

	err := svc.BanUser(context.Background(), userID)
	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
	reportRepo.AssertExpectations(t)
}

func TestUnbanUser(t *testing.T) {
	userRepo := new(mockUserRepo)
	svc := newModerationService(new(mockBannedWordRepo), new(mockReportRepo), userRepo)
	userID := uuid.New()

	now := time.Now()
	user := &identity.User{
		ID:                 userID,
		IsActive:           false,
		BanCount:           2,
		ReviewRequestCount: 1,
		ReviewRequested:    true,
		ReviewRequestedAt:  &now,
	}
	userRepo.On("FindByID", mock.Anything, userID).Return(user, nil)
	userRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *identity.User) bool {
		return u.ID == userID &&
			u.IsActive &&
			u.BanCount == 2 &&
			u.ReviewRequestCount == 1 &&
			!u.ReviewRequested &&
			u.ReviewRequestedAt == nil
	})).Return(nil)

	err := svc.UnbanUser(context.Background(), userID)
	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func TestRequestBanReview(t *testing.T) {
	userRepo := new(mockUserRepo)
	svc := newModerationService(new(mockBannedWordRepo), new(mockReportRepo), userRepo)
	userID := uuid.New()

	user := &identity.User{ID: userID, IsActive: false, ReviewRequestCount: 2}
	userRepo.On("FindByID", mock.Anything, userID).Return(user, nil)
	userRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *identity.User) bool {
		return u.ReviewRequested &&
			u.ReviewRequestedAt != nil &&
			u.ReviewRequestCount == 3
	})).Return(nil)

	err := svc.RequestBanReview(context.Background(), userID)
	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func TestRequestBanReviewIsIdempotentForCurrentBan(t *testing.T) {
	userRepo := new(mockUserRepo)
	svc := newModerationService(new(mockBannedWordRepo), new(mockReportRepo), userRepo)
	userID := uuid.New()

	user := &identity.User{
		ID:                 userID,
		IsActive:           false,
		ReviewRequested:    true,
		ReviewRequestCount: 1,
	}
	userRepo.On("FindByID", mock.Anything, userID).Return(user, nil)

	err := svc.RequestBanReview(context.Background(), userID)
	assert.NoError(t, err)
	userRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestListReportGroupsReturnsCursorForNextPage(t *testing.T) {
	reportRepo := new(mockReportRepo)
	svc := newModerationService(new(mockBannedWordRepo), reportRepo, new(mockUserRepo))
	groups := []*moderation.ReportGroup{
		{ReportedUserID: uuid.New(), ReportCount: 5},
		{ReportedUserID: uuid.New(), ReportCount: 3},
		{ReportedUserID: uuid.New(), ReportCount: 1},
	}
	reportRepo.On("FindGroupedPage", mock.Anything, "pending", "alice", (*moderation.ReportGroupCursor)(nil), 3).
		Return(groups, nil)

	page, err := svc.ListReportGroups(context.Background(), "pending", " alice ", nil, 2)

	assert.NoError(t, err)
	assert.Len(t, page.Groups, 2)
	assert.True(t, page.HasMore)
	assert.Equal(t, groups[1].ReportCount, page.NextCursor.ReportCount)
	assert.Equal(t, groups[1].ReportedUserID, page.NextCursor.ReportedUserID)
	reportRepo.AssertExpectations(t)
}

func TestListBannedUsersReturnsCursorAndTotal(t *testing.T) {
	userRepo := new(mockUserRepo)
	svc := newModerationService(new(mockBannedWordRepo), new(mockReportRepo), userRepo)
	now := time.Now().UTC()
	bannedAt := now.Add(-time.Hour)
	reviewedAt := now
	users := []*identity.User{
		{ID: uuid.New(), ReviewRequested: true, ReviewRequestedAt: &reviewedAt, CreatedAt: now.Add(-2 * time.Hour)},
		{ID: uuid.New(), BannedAt: &bannedAt, CreatedAt: now.Add(-3 * time.Hour)},
		{ID: uuid.New(), CreatedAt: now.Add(-4 * time.Hour)},
	}
	userRepo.On("CountBanned", mock.Anything, "bob").Return(int64(7), nil)
	userRepo.On("FindBannedPage", mock.Anything, "bob", (*identity.BannedUserCursor)(nil), 3).
		Return(users, nil)

	page, err := svc.ListBannedUsers(context.Background(), " bob ", nil, 2)

	assert.NoError(t, err)
	assert.Len(t, page.Users, 2)
	assert.True(t, page.HasMore)
	assert.Equal(t, int64(7), page.Total)
	assert.Equal(t, users[1].ID, page.NextCursor.ID)
	assert.Equal(t, bannedAt, page.NextCursor.SortAt)
	userRepo.AssertExpectations(t)
}
