package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/chat"
	"github.com/ndquang191/Anochat/api/internal/service"
	"github.com/ndquang191/Anochat/api/pkg/apperr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageCursorRoundTrip(t *testing.T) {
	original := &chat.MessageCursor{
		ID:        uuid.New(),
		CreatedAt: time.Date(2026, time.July, 24, 12, 34, 56, 123456789, time.UTC),
	}

	encoded := encodeMessageCursor(original)
	require.NotNil(t, encoded)

	decoded, err := decodeMessageCursor(*encoded)
	require.NoError(t, err)
	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.CreatedAt, decoded.CreatedAt)
}

func TestDecodeMessageCursorRejectsInvalidValues(t *testing.T) {
	tests := []string{
		"not-base64!",
		"dG9vLXNob3J0",
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
	}

	for _, raw := range tests {
		t.Run(raw, func(t *testing.T) {
			cursor, err := decodeMessageCursor(raw)
			assert.Nil(t, cursor)
			assert.True(t, errors.Is(err, apperr.ErrInvalidCursor))
		})
	}
}

func TestParseMessagePageSize(t *testing.T) {
	tests := []struct {
		raw       string
		want      int
		wantError bool
	}{
		{raw: "", want: service.DefaultMessagePageSize},
		{raw: "1", want: 1},
		{raw: "100", want: 100},
		{raw: "0", wantError: true},
		{raw: "101", wantError: true},
		{raw: "abc", wantError: true},
	}

	for _, test := range tests {
		t.Run(test.raw, func(t *testing.T) {
			got, err := parseMessagePageSize(test.raw)
			if test.wantError {
				assert.ErrorIs(t, err, apperr.ErrInvalidPageSize)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestGetRoomMessagesRequiresRoomMembership(t *testing.T) {
	gin.SetMode(gin.TestMode)
	currentUserID := uuid.New()
	room := &chat.Room{ID: uuid.New(), User1ID: uuid.New(), User2ID: uuid.New()}
	messageRepo := &messagePageRepoStub{}
	handler := &UserHandler{
		roomService:    service.NewRoomService(&roomRepoStub{room: room}, messageRepo),
		messageService: service.NewMessageService(messageRepo),
	}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", currentUserID)
		c.Next()
	})
	router.GET("/rooms/:id/messages", handler.GetRoomMessages)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/rooms/"+room.ID.String()+"/messages", nil)
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusForbidden, response.Code)
	assert.False(t, messageRepo.called)
}

func TestGetRoomMessagesReturnsCursorPage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	currentUserID := uuid.New()
	room := &chat.Room{ID: uuid.New(), User1ID: currentUserID, User2ID: uuid.New()}
	now := time.Now().UTC()
	messageRepo := &messagePageRepoStub{
		messages: []*chat.Message{
			{ID: uuid.New(), RoomID: room.ID, SenderID: currentUserID, Content: "new", CreatedAt: now},
			{ID: uuid.New(), RoomID: room.ID, SenderID: room.User2ID, Content: "old", CreatedAt: now.Add(-time.Second)},
		},
	}
	handler := &UserHandler{
		roomService:    service.NewRoomService(&roomRepoStub{room: room}, messageRepo),
		messageService: service.NewMessageService(messageRepo),
	}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", currentUserID)
		c.Next()
	})
	router.GET("/rooms/:id/messages", handler.GetRoomMessages)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/rooms/"+room.ID.String()+"/messages?limit=1", nil)
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, 2, messageRepo.limit)
	assert.Contains(t, response.Body.String(), `"content":"new"`)
	assert.Contains(t, response.Body.String(), `"has_more":true`)
	assert.Contains(t, response.Body.String(), `"next_cursor":`)
}

type roomRepoStub struct {
	room *chat.Room
}

func (r *roomRepoStub) FindByID(context.Context, uuid.UUID) (*chat.Room, error) {
	return r.room, nil
}

func (r *roomRepoStub) FindActiveByUserID(context.Context, uuid.UUID) (*chat.Room, error) {
	return r.room, nil
}

func (r *roomRepoStub) ListActive(context.Context) ([]*chat.Room, error) {
	if r.room == nil {
		return nil, nil
	}
	return []*chat.Room{r.room}, nil
}

func (r *roomRepoStub) Create(context.Context, *chat.Room) error {
	return nil
}

func (r *roomRepoStub) UpdateEndedAt(context.Context, uuid.UUID, time.Time) error {
	return nil
}

func (r *roomRepoStub) UpdateSessionConnection(context.Context, uuid.UUID, bool, time.Time) error {
	return nil
}

func (r *roomRepoStub) Delete(context.Context, uuid.UUID) error {
	return nil
}

type messagePageRepoStub struct {
	messages []*chat.Message
	called   bool
	limit    int
}

func (r *messagePageRepoStub) FindByID(context.Context, uuid.UUID) (*chat.Message, error) {
	return nil, nil
}

func (r *messagePageRepoStub) FindByRoomID(context.Context, uuid.UUID) ([]*chat.Message, error) {
	return nil, nil
}

func (r *messagePageRepoStub) FindByRoomIDWithLimit(context.Context, uuid.UUID, int, int) ([]*chat.Message, error) {
	return nil, nil
}

func (r *messagePageRepoStub) FindPageByRoomID(
	_ context.Context,
	_ uuid.UUID,
	_ *chat.MessageCursor,
	limit int,
) ([]*chat.Message, error) {
	r.called = true
	r.limit = limit
	return r.messages, nil
}

func (r *messagePageRepoStub) Create(context.Context, *chat.Message) error {
	return nil
}

func (r *messagePageRepoStub) DeleteByID(context.Context, uuid.UUID) error {
	return nil
}

func (r *messagePageRepoStub) DeleteByRoomID(context.Context, uuid.UUID) error {
	return nil
}
