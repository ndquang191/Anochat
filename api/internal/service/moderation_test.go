package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/identity"
	"github.com/ndquang191/Anochat/api/internal/domain/moderation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newModerationService(bwRepo *mockBannedWordRepo, rRepo *mockReportRepo, uRepo *mockUserRepo) *ModerationService {
	return NewModerationService(bwRepo, rRepo, uRepo)
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
	svc := newModerationService(new(mockBannedWordRepo), reportRepo, new(mockUserRepo))

	reportRepo.On("Create", mock.Anything, mock.MatchedBy(func(r *moderation.Report) bool {
		return r.Status == "pending"
	})).Return(nil)

	err := svc.CreateReport(context.Background(), uuid.New(), uuid.New(), uuid.New())
	assert.NoError(t, err)
	reportRepo.AssertExpectations(t)
}

func TestBanUser(t *testing.T) {
	userRepo := new(mockUserRepo)
	reportRepo := new(mockReportRepo)
	svc := newModerationService(new(mockBannedWordRepo), reportRepo, userRepo)
	userID := uuid.New()

	user := &identity.User{ID: userID, IsActive: true}
	userRepo.On("FindByID", mock.Anything, userID).Return(user, nil)
	userRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *identity.User) bool {
		return u.ID == userID && !u.IsActive
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

	user := &identity.User{ID: userID, IsActive: false}
	userRepo.On("FindByID", mock.Anything, userID).Return(user, nil)
	userRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *identity.User) bool {
		return u.ID == userID && u.IsActive
	})).Return(nil)

	err := svc.UnbanUser(context.Background(), userID)
	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}
