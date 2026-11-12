package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	pongWait                        = 60 * time.Second
	pingPeriod                      = 54 * time.Second
	writeWait                       = 10 * time.Second
	maxJSONBytesPerMessageCharacter = 6
	webSocketPayloadOverhead        = 1024
)

type WSMessage struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

type Client struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Conn         *websocket.Conn
	Hub          *Hub
	Send         chan []byte
	roomMutex    sync.RWMutex
	RoomIDs      map[uuid.UUID]struct{}
	LastActivity time.Time
	ctx          context.Context
}

func NewClient(hub *Hub, conn *websocket.Conn, userID uuid.UUID) *Client {
	return &Client{
		ID:           uuid.New(),
		UserID:       userID,
		Conn:         conn,
		Hub:          hub,
		Send:         make(chan []byte, 256),
		RoomIDs:      make(map[uuid.UUID]struct{}),
		LastActivity: time.Now(),
		ctx:          context.Background(),
	}
}

func (c *Client) addRoom(roomID uuid.UUID) bool {
	c.roomMutex.Lock()
	defer c.roomMutex.Unlock()
	if _, exists := c.RoomIDs[roomID]; exists {
		return false
	}
	c.RoomIDs[roomID] = struct{}{}
	return true
}

func (c *Client) removeRoom(roomID uuid.UUID) {
	c.roomMutex.Lock()
	delete(c.RoomIDs, roomID)
	c.roomMutex.Unlock()
}

func (c *Client) hasRoom(roomID uuid.UUID) bool {
	c.roomMutex.RLock()
	_, exists := c.RoomIDs[roomID]
	c.roomMutex.RUnlock()
	return exists
}

func (c *Client) roomIDs() []uuid.UUID {
	c.roomMutex.RLock()
	rooms := make([]uuid.UUID, 0, len(c.RoomIDs))
	for roomID := range c.RoomIDs {
		rooms = append(rooms, roomID)
	}
	c.roomMutex.RUnlock()
	return rooms
}

func (c *Client) SendJSON(msg WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("Failed to marshal message", "error", err)
		return
	}
	select {
	case c.Send <- data:
	default:
		slog.Warn("Client send channel full", "user_id", c.UserID)
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxInboundMessageSize(c.Hub.maxMessageLength))
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		c.LastActivity = time.Now()
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("WebSocket error", "error", err, "user_id", c.UserID)
			}
			break
		}
		c.LastActivity = time.Now()
		c.handleMessage(message)
	}
}

func maxInboundMessageSize(maxMessageLength int) int64 {
	return int64(maxMessageLength*maxJSONBytesPerMessageCharacter + webSocketPayloadOverhead)
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
