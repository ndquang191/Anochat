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
	existing, exists := h.clients[client.UserID]
	isLatest := exists && existing == client
	if isLatest {
		delete(h.clients, client.UserID)
	}
	h.mutex.Unlock()

	if client.RoomID != nil {
		h.removeClientFromRoom(client, *client.RoomID)
	}

	// Only clean up queue state if no newer connection has replaced this one
	if isLatest {
		h.queueService.UserDisconnected(client.UserID)
	}
	close(client.Send)
	slog.Info("Client unregistered", "user_id", client.UserID, "client_id", client.ID, "was_latest", isLatest)
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
		slog.Info("Client added to room", "user_id", userID, "room_id", roomID, "room_size", len(h.roomClients[roomID]))
	} else {
		slog.Warn("AddClientToRoom: client not found in hub", "user_id", userID, "room_id", roomID)
	}
}

func (h *Hub) removeClientFromRoom(client *Client, roomID uuid.UUID) {
	h.roomMutex.Lock()
	defer h.roomMutex.Unlock()

	if roomUsers, exists := h.roomClients[roomID]; exists {
		// Only remove if the client in the map is the same pointer (not a newer connection)
		if roomUsers[client.UserID] == client {
			delete(roomUsers, client.UserID)
			if len(roomUsers) == 0 {
				delete(h.roomClients, roomID)
			}
			slog.Info("Client removed from room", "user_id", client.UserID, "room_id", roomID)
		}
	}
}

func (h *Hub) broadcastToRoom(msg *BroadcastMessage) {
	h.roomMutex.RLock()
	roomUsers := h.roomClients[msg.RoomID]
	// Copy clients under lock to avoid data race during iteration
	targets := make([]*Client, 0, len(roomUsers))
	var targetIDs []uuid.UUID
	for userID, client := range roomUsers {
		if userID != msg.Exclude {
			targets = append(targets, client)
			targetIDs = append(targetIDs, userID)
		}
	}
	h.roomMutex.RUnlock()

	slog.Info("Broadcasting to room", "room_id", msg.RoomID, "total_in_room", len(roomUsers), "targets", len(targets), "exclude", msg.Exclude)

	var staleClients []*Client
	for i, client := range targets {
		select {
		case client.Send <- msg.Message:
			slog.Info("Broadcast delivered to client", "user_id", targetIDs[i], "room_id", msg.RoomID)
		default:
			slog.Warn("Client send channel full, marking for unregister", "user_id", client.UserID)
			staleClients = append(staleClients, client)
		}
	}

	// Unregister stale clients outside the iteration to avoid deadlock
	// (h.unregister is unbuffered and Run() is the only reader)
	for _, client := range staleClients {
		go func(c *Client) {
			h.unregister <- c
		}(client)
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
