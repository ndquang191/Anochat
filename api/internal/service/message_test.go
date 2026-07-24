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

func TestGetMessagePage(t *testing.T) {
	repo := new(mockMessageRepo)
	svc := NewMessageService(repo)
	roomID := uuid.New()
	now := time.Now().UTC()

	newest := &chat.Message{ID: uuid.New(), RoomID: roomID, Content: "newest", CreatedAt: now}
	middle := &chat.Message{ID: uuid.New(), RoomID: roomID, Content: "middle", CreatedAt: now.Add(-time.Second)}
	oldestReturned := &chat.Message{ID: uuid.New(), RoomID: roomID, Content: "oldest returned", CreatedAt: now.Add(-2 * time.Second)}
	extra := &chat.Message{ID: uuid.New(), RoomID: roomID, Content: "extra", CreatedAt: now.Add(-3 * time.Second)}

	repo.On("FindPageByRoomID", mock.Anything, roomID, (*chat.MessageCursor)(nil), 4).
		Return([]*chat.Message{newest, middle, oldestReturned, extra}, nil)

	page, err := svc.GetMessagePage(context.Background(), roomID, nil, 3)

	assert.NoError(t, err)
	assert.True(t, page.HasMore)
	assert.Equal(t, []string{"oldest returned", "middle", "newest"}, []string{
		page.Messages[0].Content,
		page.Messages[1].Content,
		page.Messages[2].Content,
	})
	assert.Equal(t, oldestReturned.ID, page.NextCursor.ID)
	assert.Equal(t, oldestReturned.CreatedAt, page.NextCursor.CreatedAt)
	repo.AssertExpectations(t)
}

func TestGetMessagePage_FinalPageHasNoCursor(t *testing.T) {
	repo := new(mockMessageRepo)
	svc := NewMessageService(repo)
	roomID := uuid.New()
	before := &chat.MessageCursor{ID: uuid.New(), CreatedAt: time.Now().UTC()}
	message := &chat.Message{ID: uuid.New(), RoomID: roomID, Content: "last page"}

	repo.On("FindPageByRoomID", mock.Anything, roomID, before, 3).
		Return([]*chat.Message{message}, nil)

	page, err := svc.GetMessagePage(context.Background(), roomID, before, 2)

	assert.NoError(t, err)
	assert.False(t, page.HasMore)
	assert.Nil(t, page.NextCursor)
	assert.Equal(t, []*chat.Message{message}, page.Messages)
	repo.AssertExpectations(t)
}
