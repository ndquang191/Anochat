package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/chat"
	"github.com/ndquang191/Anochat/api/internal/repository"
	"github.com/ndquang191/Anochat/api/pkg/apperr"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	return redis.NewClient(&redis.Options{Addr: mr.Addr()})
}

func newQueueServiceWithMocks(t *testing.T) (*QueueService, *mockRoomRepo, *mockMessageRepo, *mockMatchNotifier) {
	roomRepo := new(mockRoomRepo)
	msgRepo := new(mockMessageRepo)
	roomService := NewRoomService(roomRepo, msgRepo)
	notifier := new(mockMatchNotifier)
	rdb := newTestRedis(t)
	qs := NewQueueService(roomService, roomRepo, rdb)
	qs.SetMatchNotifier(notifier)
	return qs, roomRepo, msgRepo, notifier
}

func TestJoinQueue(t *testing.T) {
	qs, roomRepo, _, _ := newQueueServiceWithMocks(t)
	userID := uuid.New()

	roomRepo.On("FindActiveByUserID", mock.Anything, userID).Return(nil, repository.ErrNotFound)

	err := qs.JoinQueue(context.Background(), userID)
	require.NoError(t, err)
	assert.True(t, qs.IsInQueue(userID))
}

func TestJoinQueue_AlreadyInQueue(t *testing.T) {
	qs, roomRepo, _, _ := newQueueServiceWithMocks(t)
	userID := uuid.New()

	roomRepo.On("FindActiveByUserID", mock.Anything, userID).Return(nil, repository.ErrNotFound)

	err := qs.JoinQueue(context.Background(), userID)
	require.NoError(t, err)

	err = qs.JoinQueue(context.Background(), userID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "đã ở trong hàng chờ")
}

func TestJoinQueue_HasActiveRoom(t *testing.T) {
	qs, roomRepo, _, _ := newQueueServiceWithMocks(t)
	userID := uuid.New()

	roomRepo.On("FindActiveByUserID", mock.Anything, userID).Return(&chat.Room{ID: uuid.New()}, nil)

	err := qs.JoinQueue(context.Background(), userID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "đang có phòng chat")
}

func TestJoinQueue_Match(t *testing.T) {
	qs, roomRepo, _, notifier := newQueueServiceWithMocks(t)
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

func TestJoinQueue_CreateRoomFailureReleasesReservation(t *testing.T) {
	qs, roomRepo, _, _ := newQueueServiceWithMocks(t)
	partnerID := uuid.New()
	userID := uuid.New()
	createErr := errors.New("database unavailable")

	roomRepo.On("FindActiveByUserID", mock.Anything, partnerID).Return(nil, repository.ErrNotFound)
	roomRepo.On("FindActiveByUserID", mock.Anything, userID).Return(nil, repository.ErrNotFound)
	roomRepo.On("Create", mock.Anything, mock.AnythingOfType("*chat.Room")).Return(createErr)

	require.NoError(t, qs.JoinQueue(context.Background(), partnerID))

	err := qs.JoinQueue(context.Background(), userID)
	require.Error(t, err)
	assert.ErrorIs(t, err, createErr)
	assert.True(t, qs.IsInQueue(partnerID), "partner must remain available after room creation fails")
	assert.False(t, qs.IsInQueue(userID))

	reservationCount, redisErr := qs.rdb.ZCard(context.Background(), queueReservationsKey).Result()
	require.NoError(t, redisErr)
	assert.Zero(t, reservationCount)
	tokenCount, redisErr := qs.rdb.HLen(context.Background(), queueReservationTokensKey).Result()
	require.NoError(t, redisErr)
	assert.Zero(t, tokenCount)
}

func TestJoinQueue_ConcurrentDuplicateOnlyCreatesOneRoom(t *testing.T) {
	qs, roomRepo, _, notifier := newQueueServiceWithMocks(t)
	partner1 := uuid.New()
	partner2 := uuid.New()
	userID := uuid.New()
	createStarted := make(chan struct{})
	allowCreate := make(chan struct{})

	require.NoError(t, qs.rdb.ZAdd(
		context.Background(),
		queueKey,
		redis.Z{Score: 1, Member: partner1.String()},
		redis.Z{Score: 2, Member: partner2.String()},
	).Err())

	roomRepo.On("FindActiveByUserID", mock.Anything, userID).Return(nil, repository.ErrNotFound)
	roomRepo.On("Create", mock.Anything, mock.AnythingOfType("*chat.Room")).
		Run(func(mock.Arguments) {
			close(createStarted)
			<-allowCreate
		}).
		Return(nil).
		Once()
	notifier.On("NotifyMatch", partner1, userID, mock.AnythingOfType("uuid.UUID")).Return().Once()

	firstResult := make(chan error, 1)
	go func() {
		firstResult <- qs.JoinQueue(context.Background(), userID)
	}()

	select {
	case <-createStarted:
	case <-time.After(time.Second):
		t.Fatal("first join did not reach room creation")
	}

	secondErr := qs.JoinQueue(context.Background(), userID)
	assert.ErrorIs(t, secondErr, apperr.ErrMatchInProgress)

	close(allowCreate)
	select {
	case err := <-firstResult:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("first join did not finish")
	}

	assert.False(t, qs.IsInQueue(partner1))
	assert.True(t, qs.IsInQueue(partner2))
	roomRepo.AssertNumberOfCalls(t, "Create", 1)
	notifier.AssertExpectations(t)
}

func TestJoinQueue_ExpiredReservationDoesNotBlockCandidate(t *testing.T) {
	qs, roomRepo, _, notifier := newQueueServiceWithMocks(t)
	partnerID := uuid.New()
	staleUserID := uuid.New()
	userID := uuid.New()
	staleToken := uuid.NewString()

	require.NoError(t, qs.rdb.ZAdd(
		context.Background(),
		queueKey,
		redis.Z{Score: 1, Member: partnerID.String()},
	).Err())
	require.NoError(t, qs.rdb.ZAdd(
		context.Background(),
		queueReservationsKey,
		redis.Z{Score: float64(time.Now().Add(-time.Second).UnixMilli()), Member: partnerID.String()},
		redis.Z{Score: float64(time.Now().Add(-time.Second).UnixMilli()), Member: staleUserID.String()},
	).Err())
	require.NoError(t, qs.rdb.HSet(
		context.Background(),
		queueReservationTokensKey,
		partnerID.String(),
		staleToken,
		staleUserID.String(),
		staleToken,
	).Err())

	roomRepo.On("FindActiveByUserID", mock.Anything, userID).Return(nil, repository.ErrNotFound)
	roomRepo.On("Create", mock.Anything, mock.AnythingOfType("*chat.Room")).Return(nil).Once()
	notifier.On("NotifyMatch", partnerID, userID, mock.AnythingOfType("uuid.UUID")).Return().Once()

	require.NoError(t, qs.JoinQueue(context.Background(), userID))
	assert.False(t, qs.IsInQueue(partnerID))

	reservationCount, err := qs.rdb.ZCard(context.Background(), queueReservationsKey).Result()
	require.NoError(t, err)
	assert.Zero(t, reservationCount)
	tokenCount, err := qs.rdb.HLen(context.Background(), queueReservationTokensKey).Result()
	require.NoError(t, err)
	assert.Zero(t, tokenCount)
}

func TestFinishMatchReservation_StaleTokenCannotClearNewReservation(t *testing.T) {
	qs, _, _, _ := newQueueServiceWithMocks(t)
	partnerID := uuid.New()
	userID := uuid.New()
	currentToken := uuid.NewString()

	require.NoError(t, qs.rdb.ZAdd(
		context.Background(),
		queueKey,
		redis.Z{Score: 1, Member: partnerID.String()},
	).Err())
	require.NoError(t, qs.rdb.ZAdd(
		context.Background(),
		queueReservationsKey,
		redis.Z{Score: float64(time.Now().Add(time.Second).UnixMilli()), Member: partnerID.String()},
		redis.Z{Score: float64(time.Now().Add(time.Second).UnixMilli()), Member: userID.String()},
	).Err())
	require.NoError(t, qs.rdb.HSet(
		context.Background(),
		queueReservationTokensKey,
		partnerID.String(),
		currentToken,
		userID.String(),
		currentToken,
	).Err())

	committed, err := qs.finishMatchReservation(userID, partnerID, uuid.NewString(), "commit")
	require.NoError(t, err)
	assert.False(t, committed)
	assert.True(t, qs.IsInQueue(partnerID))

	reservationCount, err := qs.rdb.ZCard(context.Background(), queueReservationsKey).Result()
	require.NoError(t, err)
	assert.EqualValues(t, 2, reservationCount)
}

func TestLeaveQueue(t *testing.T) {
	qs, roomRepo, _, _ := newQueueServiceWithMocks(t)
	userID := uuid.New()

	roomRepo.On("FindActiveByUserID", mock.Anything, userID).Return(nil, repository.ErrNotFound)

	err := qs.JoinQueue(context.Background(), userID)
	require.NoError(t, err)

	err = qs.LeaveQueue(context.Background(), userID)
	assert.NoError(t, err)
	assert.False(t, qs.IsInQueue(userID))
}

func TestLeaveQueue_NotInQueue(t *testing.T) {
	qs, _, _, _ := newQueueServiceWithMocks(t)
	userID := uuid.New()

	err := qs.LeaveQueue(context.Background(), userID)
	assert.ErrorIs(t, err, apperr.ErrNotInQueue)
}

func TestLeaveQueue_MatchReservationPreventsRemoval(t *testing.T) {
	qs, _, _, _ := newQueueServiceWithMocks(t)
	partnerID := uuid.New()
	userID := uuid.New()
	token := uuid.NewString()

	require.NoError(t, qs.rdb.ZAdd(
		context.Background(),
		queueKey,
		redis.Z{Score: 1, Member: partnerID.String()},
	).Err())
	require.NoError(t, qs.rdb.ZAdd(
		context.Background(),
		queueReservationsKey,
		redis.Z{Score: float64(time.Now().Add(time.Second).UnixMilli()), Member: partnerID.String()},
		redis.Z{Score: float64(time.Now().Add(time.Second).UnixMilli()), Member: userID.String()},
	).Err())
	require.NoError(t, qs.rdb.HSet(
		context.Background(),
		queueReservationTokensKey,
		partnerID.String(),
		token,
		userID.String(),
		token,
	).Err())

	err := qs.LeaveQueue(context.Background(), partnerID)
	assert.ErrorIs(t, err, apperr.ErrMatchInProgress)
	assert.True(t, qs.IsInQueue(partnerID))
}

func TestReconcileRoomsRemovesStaleQueueAndReservationState(t *testing.T) {
	qs, _, _, _ := newQueueServiceWithMocks(t)
	user1 := uuid.New()
	user2 := uuid.New()
	token := uuid.NewString()
	roomCreatedAt := time.Now()
	room := &chat.Room{
		ID:        uuid.New(),
		User1ID:   user1,
		User2ID:   user2,
		CreatedAt: roomCreatedAt,
	}

	require.NoError(t, qs.rdb.ZAdd(
		context.Background(),
		queueKey,
		redis.Z{Score: float64(roomCreatedAt.Add(-time.Second).UnixMilli()), Member: user1.String()},
		redis.Z{Score: float64(roomCreatedAt.Add(-time.Second).UnixMilli()), Member: user2.String()},
	).Err())
	require.NoError(t, qs.rdb.ZAdd(
		context.Background(),
		queueReservationsKey,
		redis.Z{Score: float64(roomCreatedAt.Add(matchReservationTTL - time.Second).UnixMilli()), Member: user1.String()},
		redis.Z{Score: float64(roomCreatedAt.Add(matchReservationTTL - time.Second).UnixMilli()), Member: user2.String()},
	).Err())
	require.NoError(t, qs.rdb.HSet(
		context.Background(),
		queueReservationTokensKey,
		user1.String(),
		token,
		user2.String(),
		token,
	).Err())

	require.NoError(t, qs.reconcileRooms(context.Background(), []*chat.Room{room}))
	assert.False(t, qs.IsInQueue(user1))
	assert.False(t, qs.IsInQueue(user2))
	reservations, err := qs.rdb.ZCard(context.Background(), queueReservationsKey).Result()
	require.NoError(t, err)
	assert.Zero(t, reservations)
	tokens, err := qs.rdb.HLen(context.Background(), queueReservationTokensKey).Result()
	require.NoError(t, err)
	assert.Zero(t, tokens)
}

func TestReconcileRoomsPreservesQueueStateNewerThanRoom(t *testing.T) {
	qs, _, _, _ := newQueueServiceWithMocks(t)
	user1 := uuid.New()
	user2 := uuid.New()
	roomCreatedAt := time.Now().Add(-time.Minute)
	room := &chat.Room{
		ID:        uuid.New(),
		User1ID:   user1,
		User2ID:   user2,
		CreatedAt: roomCreatedAt,
	}

	require.NoError(t, qs.rdb.ZAdd(
		context.Background(),
		queueKey,
		redis.Z{Score: float64(time.Now().UnixMilli()), Member: user1.String()},
	).Err())

	require.NoError(t, qs.reconcileRooms(context.Background(), []*chat.Room{room}))
	assert.True(t, qs.IsInQueue(user1), "a queue entry created after the room must survive stale reconciliation")
}

func TestIsInQueue(t *testing.T) {
	qs, roomRepo, _, _ := newQueueServiceWithMocks(t)
	userID := uuid.New()

	assert.False(t, qs.IsInQueue(userID))

	roomRepo.On("FindActiveByUserID", mock.Anything, userID).Return(nil, repository.ErrNotFound)
	_ = qs.JoinQueue(context.Background(), userID)

	assert.True(t, qs.IsInQueue(userID))
}

func TestUserDisconnected(t *testing.T) {
	qs, roomRepo, _, _ := newQueueServiceWithMocks(t)
	userID := uuid.New()

	roomRepo.On("FindActiveByUserID", mock.Anything, userID).Return(nil, repository.ErrNotFound)
	_ = qs.JoinQueue(context.Background(), userID)
	assert.True(t, qs.IsInQueue(userID))

	qs.UserDisconnected(userID)
	assert.False(t, qs.IsInQueue(userID))
}

func TestUserDisconnected_NotInQueue(t *testing.T) {
	qs, _, _, _ := newQueueServiceWithMocks(t)
	// Should not panic
	qs.UserDisconnected(uuid.New())
}

func TestRunStopsWhenContextIsCancelled(t *testing.T) {
	qs, roomRepo, _, _ := newQueueServiceWithMocks(t)
	roomRepo.On("ListActive", mock.Anything).Return([]*chat.Room{}, nil).Once()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		qs.Run(ctx)
		close(done)
	}()
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("queue reconciler did not stop after context cancellation")
	}
	roomRepo.AssertExpectations(t)
}
