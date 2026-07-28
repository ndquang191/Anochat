package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/chat"
	"github.com/ndquang191/Anochat/api/internal/domain/identity"
	"github.com/ndquang191/Anochat/api/internal/domain/moderation"
	"github.com/stretchr/testify/mock"
)

// --- UserRepository mock ---

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*identity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*identity.User), args.Error(1)
}

func (m *mockUserRepo) FindByIDWithProfile(ctx context.Context, id uuid.UUID) (*identity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*identity.User), args.Error(1)
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*identity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*identity.User), args.Error(1)
}

func (m *mockUserRepo) Create(ctx context.Context, user *identity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) Update(ctx context.Context, user *identity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockUserRepo) FindBannedPage(
	ctx context.Context,
	query string,
	before *identity.BannedUserCursor,
	limit int,
) ([]*identity.User, error) {
	args := m.Called(ctx, query, before, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*identity.User), args.Error(1)
}

func (m *mockUserRepo) CountBanned(ctx context.Context, query string) (int64, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(int64), args.Error(1)
}

// --- ProfileRepository mock ---

type mockProfileRepo struct{ mock.Mock }

func (m *mockProfileRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*identity.Profile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*identity.Profile), args.Error(1)
}

func (m *mockProfileRepo) FindPublicByUserID(ctx context.Context, userID uuid.UUID) (*identity.Profile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*identity.Profile), args.Error(1)
}

func (m *mockProfileRepo) Create(ctx context.Context, profile *identity.Profile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *mockProfileRepo) Update(ctx context.Context, profile *identity.Profile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *mockProfileRepo) UpdateWithNicknameCooldown(ctx context.Context, profile *identity.Profile, cutoff time.Time) error {
	args := m.Called(ctx, profile, cutoff)
	return args.Error(0)
}

// --- BannedWordRepository mock ---

type mockBannedWordRepo struct{ mock.Mock }

func (m *mockBannedWordRepo) FindAll(ctx context.Context) ([]*moderation.BannedWord, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*moderation.BannedWord), args.Error(1)
}

func (m *mockBannedWordRepo) Create(ctx context.Context, word *moderation.BannedWord) error {
	args := m.Called(ctx, word)
	return args.Error(0)
}

func (m *mockBannedWordRepo) Update(ctx context.Context, word *moderation.BannedWord) error {
	args := m.Called(ctx, word)
	return args.Error(0)
}

func (m *mockBannedWordRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// --- ReportRepository mock ---

type mockReportRepo struct{ mock.Mock }

func (m *mockReportRepo) CreateWithSnapshot(
	ctx context.Context,
	report *moderation.Report,
	triggerMessage *chat.Message,
	limit int,
	requireRoom bool,
) error {
	args := m.Called(ctx, report, triggerMessage, limit, requireRoom)
	return args.Error(0)
}

func (m *mockReportRepo) FindGroupedPage(
	ctx context.Context,
	status, query string,
	before *moderation.ReportGroupCursor,
	limit int,
) ([]*moderation.ReportGroup, error) {
	args := m.Called(ctx, status, query, before, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*moderation.ReportGroup), args.Error(1)
}

func (m *mockReportRepo) FindMessages(ctx context.Context, reportID uuid.UUID) ([]*moderation.ReportMessage, error) {
	args := m.Called(ctx, reportID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*moderation.ReportMessage), args.Error(1)
}

func (m *mockReportRepo) MarkReviewedByUser(ctx context.Context, reportedUserID uuid.UUID) error {
	args := m.Called(ctx, reportedUserID)
	return args.Error(0)
}

func (m *mockReportRepo) FindLatestReportForUsers(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	args := m.Called(ctx, userIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID]uuid.UUID), args.Error(1)
}

// --- RoomRepository mock ---

type mockRoomRepo struct{ mock.Mock }

func (m *mockRoomRepo) FindByID(ctx context.Context, id uuid.UUID) (*chat.Room, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*chat.Room), args.Error(1)
}

func (m *mockRoomRepo) FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*chat.Room, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*chat.Room), args.Error(1)
}

func (m *mockRoomRepo) ListActive(ctx context.Context) ([]*chat.Room, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*chat.Room), args.Error(1)
}

func (m *mockRoomRepo) Create(ctx context.Context, room *chat.Room) error {
	args := m.Called(ctx, room)
	return args.Error(0)
}

func (m *mockRoomRepo) UpdateEndedAt(ctx context.Context, roomID uuid.UUID, endedAt time.Time) error {
	args := m.Called(ctx, roomID, endedAt)
	return args.Error(0)
}

func (m *mockRoomRepo) UpdateSessionConnection(ctx context.Context, roomID uuid.UUID, connected bool, changedAt time.Time) error {
	args := m.Called(ctx, roomID, connected, changedAt)
	return args.Error(0)
}

func (m *mockRoomRepo) Delete(ctx context.Context, roomID uuid.UUID) error {
	args := m.Called(ctx, roomID)
	return args.Error(0)
}

// --- MessageRepository mock ---

type mockMessageRepo struct{ mock.Mock }

func (m *mockMessageRepo) FindByID(ctx context.Context, id uuid.UUID) (*chat.Message, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*chat.Message), args.Error(1)
}

func (m *mockMessageRepo) FindByRoomID(ctx context.Context, roomID uuid.UUID) ([]*chat.Message, error) {
	args := m.Called(ctx, roomID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*chat.Message), args.Error(1)
}

func (m *mockMessageRepo) FindByRoomIDWithLimit(ctx context.Context, roomID uuid.UUID, limit, offset int) ([]*chat.Message, error) {
	args := m.Called(ctx, roomID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*chat.Message), args.Error(1)
}

func (m *mockMessageRepo) FindPageByRoomID(ctx context.Context, roomID uuid.UUID, before *chat.MessageCursor, limit int) ([]*chat.Message, error) {
	args := m.Called(ctx, roomID, before, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*chat.Message), args.Error(1)
}

func (m *mockMessageRepo) Create(ctx context.Context, msg *chat.Message) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *mockMessageRepo) DeleteByID(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockMessageRepo) DeleteByRoomID(ctx context.Context, roomID uuid.UUID) error {
	args := m.Called(ctx, roomID)
	return args.Error(0)
}

// --- MatchNotifier mock ---

type mockMatchNotifier struct{ mock.Mock }

func (m *mockMatchNotifier) NotifyMatch(user1ID, user2ID, roomID uuid.UUID) {
	m.Called(user1ID, user2ID, roomID)
}
