package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/chat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSaveMessage(t *testing.T) {
	repo := new(mockMessageRepo)
	svc := NewMessageService(repo)

	msgID := uuid.New()
	roomID := uuid.New()
	senderID := uuid.New()
	now := time.Now()

	repo.On("Create", mock.Anything, mock.MatchedBy(func(m *chat.Message) bool {
		return m.ID == msgID && m.RoomID == roomID && m.SenderID == senderID && m.Content == "hello"
	})).Return(nil)

	err := svc.SaveMessage(context.Background(), msgID, roomID, senderID, "hello", now)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestGetMessagesByRoomID(t *testing.T) {
	repo := new(mockMessageRepo)
	svc := NewMessageService(repo)
	roomID := uuid.New()

	expected := []*chat.Message{
		{ID: uuid.New(), RoomID: roomID, Content: "msg1"},
		{ID: uuid.New(), RoomID: roomID, Content: "msg2"},
	}
	repo.On("FindByRoomID", mock.Anything, roomID).Return(expected, nil)

	messages, err := svc.GetMessagesByRoomID(context.Background(), roomID)
	assert.NoError(t, err)
	assert.Len(t, messages, 2)
	assert.Equal(t, "msg1", messages[0].Content)
	repo.AssertExpectations(t)
}

func TestGetMessagesByRoomID_Empty(t *testing.T) {
	repo := new(mockMessageRepo)
	svc := NewMessageService(repo)
	roomID := uuid.New()

	repo.On("FindByRoomID", mock.Anything, roomID).Return([]*chat.Message{}, nil)

	messages, err := svc.GetMessagesByRoomID(context.Background(), roomID)
	assert.NoError(t, err)
	assert.Empty(t, messages)
	repo.AssertExpectations(t)
}
