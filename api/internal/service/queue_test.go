package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/chat"
	"github.com/ndquang191/Anochat/api/internal/repository"
	"github.com/ndquang191/Anochat/api/pkg/apperr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newQueueServiceWithMocks() (*QueueService, *mockRoomRepo, *mockMessageRepo, *mockMatchNotifier) {
	roomRepo := new(mockRoomRepo)
	msgRepo := new(mockMessageRepo)
	roomService := NewRoomService(roomRepo, msgRepo)
	notifier := new(mockMatchNotifier)
	qs := NewQueueService(roomService, roomRepo)
	qs.SetMatchNotifier(notifier)
	return qs, roomRepo, msgRepo, notifier
}

func TestJoinQueue(t *testing.T) {
	qs, roomRepo, _, _ := newQueueServiceWithMocks()
	userID := uuid.New()

	roomRepo.On("FindActiveByUserID", mock.Anything, userID).Return(nil, repository.ErrNotFound)

	err := qs.JoinQueue(context.Background(), userID)
	require.NoError(t, err)
	assert.True(t, qs.IsInQueue(userID))
}

func TestJoinQueue_AlreadyInQueue(t *testing.T) {
	qs, roomRepo, _, _ := newQueueServiceWithMocks()
	userID := uuid.New()

	roomRepo.On("FindActiveByUserID", mock.Anything, userID).Return(nil, repository.ErrNotFound)

	err := qs.JoinQueue(context.Background(), userID)
	require.NoError(t, err)

	err = qs.JoinQueue(context.Background(), userID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "đã ở trong hàng chờ")
}

func TestJoinQueue_HasActiveRoom(t *testing.T) {
	qs, roomRepo, _, _ := newQueueServiceWithMocks()
	userID := uuid.New()

	roomRepo.On("FindActiveByUserID", mock.Anything, userID).Return(&chat.Room{ID: uuid.New()}, nil)

	err := qs.JoinQueue(context.Background(), userID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "đang có phòng chat")
}

func TestJoinQueue_Match(t *testing.T) {
	qs, roomRepo, _, notifier := newQueueServiceWithMocks()
	user1 := uuid.New()
	user2 := uuid.New()

	roomRepo.On("FindActiveByUserID", mock.Anything, user1).Return(nil, repository.ErrNotFound)
	roomRepo.On("FindActiveByUserID", mock.Anything, user2).Return(nil, repository.ErrNotFound)
	roomRepo.On("Create", mock.Anything, mock.AnythingOfType("*chat.Room")).Return(nil)
	notifier.On("NotifyMatch", user1, user2, mock.AnythingOfType("uuid.UUID")).Return()

	// User 1 joins - no match
	err := qs.JoinQueue(context.Background(), user1)
	require.NoError(t, err)
	assert.True(t, qs.IsInQueue(user1))

	// User 2 joins - should match with user 1
	err = qs.JoinQueue(context.Background(), user2)
	require.NoError(t, err)
	assert.False(t, qs.IsInQueue(user1))
	assert.False(t, qs.IsInQueue(user2))

	notifier.AssertCalled(t, "NotifyMatch", user1, user2, mock.AnythingOfType("uuid.UUID"))
}

func TestLeaveQueue(t *testing.T) {
	qs, roomRepo, _, _ := newQueueServiceWithMocks()
	userID := uuid.New()

	roomRepo.On("FindActiveByUserID", mock.Anything, userID).Return(nil, repository.ErrNotFound)

	err := qs.JoinQueue(context.Background(), userID)
	require.NoError(t, err)

	err = qs.LeaveQueue(context.Background(), userID)
	assert.NoError(t, err)
	assert.False(t, qs.IsInQueue(userID))
}

func TestLeaveQueue_NotInQueue(t *testing.T) {
	qs, _, _, _ := newQueueServiceWithMocks()
	userID := uuid.New()

	err := qs.LeaveQueue(context.Background(), userID)
	assert.ErrorIs(t, err, apperr.ErrNotInQueue)
}

func TestIsInQueue(t *testing.T) {
	qs, roomRepo, _, _ := newQueueServiceWithMocks()
	userID := uuid.New()

	assert.False(t, qs.IsInQueue(userID))

	roomRepo.On("FindActiveByUserID", mock.Anything, userID).Return(nil, repository.ErrNotFound)
	_ = qs.JoinQueue(context.Background(), userID)

	assert.True(t, qs.IsInQueue(userID))
}

func TestUserDisconnected(t *testing.T) {
	qs, roomRepo, _, _ := newQueueServiceWithMocks()
	userID := uuid.New()

	roomRepo.On("FindActiveByUserID", mock.Anything, userID).Return(nil, repository.ErrNotFound)
	_ = qs.JoinQueue(context.Background(), userID)
	assert.True(t, qs.IsInQueue(userID))

	qs.UserDisconnected(userID)
	assert.False(t, qs.IsInQueue(userID))
}

func TestUserDisconnected_NotInQueue(t *testing.T) {
	qs, _, _, _ := newQueueServiceWithMocks()
	// Should not panic
	qs.UserDisconnected(uuid.New())
}

func TestStop(t *testing.T) {
	qs, roomRepo, _, _ := newQueueServiceWithMocks()
	user1 := uuid.New()

	roomRepo.On("FindActiveByUserID", mock.Anything, user1).Return(nil, repository.ErrNotFound)
	_ = qs.JoinQueue(context.Background(), user1)
	assert.True(t, qs.IsInQueue(user1))

	qs.Stop()
	assert.False(t, qs.IsInQueue(user1))
}
