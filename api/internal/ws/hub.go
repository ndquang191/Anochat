package ws

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/service"
	"github.com/redis/go-redis/v9"
)

type Hub struct {
	clients     map[uuid.UUID]*Client
	roomClients map[uuid.UUID]map[uuid.UUID]*Client
	register    chan *Client
	unregister  chan *Client
	broadcast   chan *BroadcastMessage

	queueService   *service.QueueService
	messageService *service.MessageService
	roomService    *service.RoomService
	rdb            *redis.Client

	mutex     sync.RWMutex
	roomMutex sync.RWMutex
}

type BroadcastMessage struct {
	RoomID  uuid.UUID
	Message []byte
	Exclude uuid.UUID
}

func NewHub(queueService *service.QueueService, messageService *service.MessageService, roomService *service.RoomService, rdb *redis.Client) *Hub {
	return &Hub{
		clients:        make(map[uuid.UUID]*Client),
		roomClients:    make(map[uuid.UUID]map[uuid.UUID]*Client),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		broadcast:      make(chan *BroadcastMessage, 256),
		queueService:   queueService,
		messageService: messageService,
		roomService:    roomService,
		rdb:            rdb,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)
		case client := <-h.unregister:
			h.unregisterClient(client)
		case message := <-h.broadcast:
			h.broadcastToRoom(message)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.mutex.Lock()
	h.clients[client.UserID] = client
	h.mutex.Unlock()

	slog.Info("Client registered", "user_id", client.UserID, "client_id", client.ID)

	successMsg := WSMessage{
		Type: "connected",
		Payload: map[string]interface{}{
			"user_id":   client.UserID.String(),
			"message":   "Connected to WebSocket",
			"timestamp": time.Now().Unix(),
		},
	}
	client.SendJSON(successMsg)
}

func (h *Hub) unregisterClient(client *Client) {
	h.mutex.Lock()
	if _, exists := h.clients[client.UserID]; exists {
		delete(h.clients, client.UserID)
	}
	h.mutex.Unlock()

	if client.RoomID != nil {
		if err := h.roomService.LeaveRoom(client.ctx, *client.RoomID, client.UserID); err != nil {
			slog.Error("Failed to leave room on disconnect", "error", err, "user_id", client.UserID, "room_id", *client.RoomID)
		}
		h.removeClientFromRoom(client.UserID, *client.RoomID)
		h.notifyPartnerLeft(*client.RoomID, client.UserID)
	}

	h.queueService.UserDisconnected(client.UserID)
	close(client.Send)
	slog.Info("Client unregistered", "user_id", client.UserID, "client_id", client.ID)
}

func (h *Hub) AddClientToRoom(userID uuid.UUID, roomID uuid.UUID) {
	h.roomMutex.Lock()
	defer h.roomMutex.Unlock()

	if h.roomClients[roomID] == nil {
		h.roomClients[roomID] = make(map[uuid.UUID]*Client)
	}

	h.mutex.RLock()
	client := h.clients[userID]
	h.mutex.RUnlock()

	if client != nil {
		client.RoomID = &roomID
		h.roomClients[roomID][userID] = client
		slog.Info("Client added to room", "user_id", userID, "room_id", roomID)
	}
}

func (h *Hub) removeClientFromRoom(userID uuid.UUID, roomID uuid.UUID) {
	h.roomMutex.Lock()
	defer h.roomMutex.Unlock()

	if roomUsers, exists := h.roomClients[roomID]; exists {
		delete(roomUsers, userID)
		if len(roomUsers) == 0 {
			delete(h.roomClients, roomID)
		}
		slog.Info("Client removed from room", "user_id", userID, "room_id", roomID)
	}
}

func (h *Hub) broadcastToRoom(msg *BroadcastMessage) {
	h.roomMutex.RLock()
	roomUsers := h.roomClients[msg.RoomID]
	h.roomMutex.RUnlock()

	for userID, client := range roomUsers {
		if userID != msg.Exclude {
			select {
			case client.Send <- msg.Message:
			default:
				h.unregister <- client
			}
		}
	}
}

func (h *Hub) Register() chan<- *Client {
	return h.register
}

func (h *Hub) CheckMessageRateLimit(userID uuid.UUID) bool {
	key := fmt.Sprintf("msgrl:%s", userID.String())

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	count, err := h.rdb.Incr(ctx, key).Result()
	if err != nil {
		slog.Warn("Message rate limiter Redis error, allowing message", "error", err, "user_id", userID)
		return true
	}

	if count == 1 {
		h.rdb.Expire(ctx, key, time.Second)
	}

	return count <= 10
}
