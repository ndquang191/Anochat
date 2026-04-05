package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/chat"
	"github.com/ndquang191/Anochat/api/internal/repository"
	"github.com/ndquang191/Anochat/api/pkg/apperr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newRoomServiceWithMocks() (*RoomService, *mockRoomRepo, *mockMessageRepo) {
	roomRepo := new(mockRoomRepo)
	msgRepo := new(mockMessageRepo)
	return NewRoomService(roomRepo, msgRepo), roomRepo, msgRepo
}

func TestCreateRoom(t *testing.T) {
	svc, roomRepo, _ := newRoomServiceWithMocks()
	user1 := uuid.New()
	user2 := uuid.New()

	roomRepo.On("Create", mock.Anything, mock.MatchedBy(func(r *chat.Room) bool {
		return r.User1ID == user1 && r.User2ID == user2
	})).Return(nil)

	room, err := svc.CreateRoom(context.Background(), user1, user2)
	assert.NoError(t, err)
	assert.Equal(t, user1, room.User1ID)
	assert.Equal(t, user2, room.User2ID)
	roomRepo.AssertExpectations(t)
}

func TestGetRoomByID(t *testing.T) {
	svc, roomRepo, _ := newRoomServiceWithMocks()
	roomID := uuid.New()
	expected := &chat.Room{ID: roomID, User1ID: uuid.New(), User2ID: uuid.New()}

	roomRepo.On("FindByID", mock.Anything, roomID).Return(expected, nil)

	room, err := svc.GetRoomByID(context.Background(), roomID)
	assert.NoError(t, err)
	assert.Equal(t, roomID, room.ID)
	roomRepo.AssertExpectations(t)
}

func TestGetActiveRoomByUserID(t *testing.T) {
	svc, roomRepo, _ := newRoomServiceWithMocks()
	userID := uuid.New()
	expected := &chat.Room{ID: uuid.New(), User1ID: userID, User2ID: uuid.New()}

	roomRepo.On("FindActiveByUserID", mock.Anything, userID).Return(expected, nil)

	room, err := svc.GetActiveRoomByUserID(context.Background(), userID)
	assert.NoError(t, err)
	assert.Equal(t, userID, room.User1ID)
	roomRepo.AssertExpectations(t)
}

func TestGetActiveRoomByUserID_NotFound(t *testing.T) {
	svc, roomRepo, _ := newRoomServiceWithMocks()
	userID := uuid.New()

	roomRepo.On("FindActiveByUserID", mock.Anything, userID).Return(nil, repository.ErrNotFound)

	room, err := svc.GetActiveRoomByUserID(context.Background(), userID)
	assert.ErrorIs(t, err, apperr.ErrNoActiveRoom)
	assert.Nil(t, room)
	roomRepo.AssertExpectations(t)
}

func TestLeaveRoom(t *testing.T) {
	svc, roomRepo, msgRepo := newRoomServiceWithMocks()
	userID := uuid.New()
	roomID := uuid.New()
	room := &chat.Room{ID: roomID, User1ID: userID, User2ID: uuid.New()}

	roomRepo.On("FindByID", mock.Anything, roomID).Return(room, nil)
	roomRepo.On("UpdateEndedAt", mock.Anything, roomID, mock.AnythingOfType("time.Time")).Return(nil)
	// cleanup goroutine
	msgRepo.On("DeleteByRoomID", mock.Anything, roomID).Return(nil)
	roomRepo.On("Delete", mock.Anything, roomID).Return(nil)

	err := svc.LeaveRoom(context.Background(), roomID, userID)
	assert.NoError(t, err)
	roomRepo.AssertCalled(t, "UpdateEndedAt", mock.Anything, roomID, mock.AnythingOfType("time.Time"))

	// Wait briefly for cleanup goroutine
	time.Sleep(50 * time.Millisecond)
	roomRepo.AssertExpectations(t)
	msgRepo.AssertExpectations(t)
}

func TestLeaveRoom_NotInRoom(t *testing.T) {
	svc, roomRepo, _ := newRoomServiceWithMocks()
	roomID := uuid.New()
	otherUser := uuid.New()
	room := &chat.Room{ID: roomID, User1ID: uuid.New(), User2ID: uuid.New()}

	roomRepo.On("FindByID", mock.Anything, roomID).Return(room, nil)

	err := svc.LeaveRoom(context.Background(), roomID, otherUser)
	assert.ErrorIs(t, err, apperr.ErrNotRoomMember)
}

func TestLeaveRoom_AlreadyEnded(t *testing.T) {
	svc, roomRepo, _ := newRoomServiceWithMocks()
	userID := uuid.New()
	roomID := uuid.New()
	ended := time.Now()
	room := &chat.Room{ID: roomID, User1ID: userID, User2ID: uuid.New(), EndedAt: &ended}

	roomRepo.On("FindByID", mock.Anything, roomID).Return(room, nil)

	err := svc.LeaveRoom(context.Background(), roomID, userID)
	assert.NoError(t, err)
	// Should not call UpdateEndedAt
	roomRepo.AssertNotCalled(t, "UpdateEndedAt", mock.Anything, mock.Anything, mock.Anything)
}

func TestLeaveCurrentRoom(t *testing.T) {
	svc, roomRepo, msgRepo := newRoomServiceWithMocks()
	userID := uuid.New()
	roomID := uuid.New()
	room := &chat.Room{ID: roomID, User1ID: userID, User2ID: uuid.New()}

	roomRepo.On("FindActiveByUserID", mock.Anything, userID).Return(room, nil)
	roomRepo.On("FindByID", mock.Anything, roomID).Return(room, nil)
	roomRepo.On("UpdateEndedAt", mock.Anything, roomID, mock.AnythingOfType("time.Time")).Return(nil)
	msgRepo.On("DeleteByRoomID", mock.Anything, roomID).Return(nil)
	roomRepo.On("Delete", mock.Anything, roomID).Return(nil)

	err := svc.LeaveCurrentRoom(context.Background(), userID)
	assert.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	roomRepo.AssertExpectations(t)
}

func TestLeaveCurrentRoom_NoActiveRoom(t *testing.T) {
	svc, roomRepo, _ := newRoomServiceWithMocks()
	userID := uuid.New()

	roomRepo.On("FindActiveByUserID", mock.Anything, userID).Return(nil, repository.ErrNotFound)

	err := svc.LeaveCurrentRoom(context.Background(), userID)
	assert.Error(t, err)
}
