package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/service"
	"github.com/ndquang191/Anochat/api/pkg/metrics"
	"github.com/redis/go-redis/v9"
)

// roomPubMsg is the message published to a room's Redis channel.
type roomPubMsg struct {
	Exclude string `json:"exclude"`
	Payload []byte `json:"payload"`
}

// userPubMsg is the message published to a user's Redis channel.
type userPubMsg struct {
	Type   string `json:"type"`
	RoomID string `json:"room_id,omitempty"`
}

type Hub struct {
	clients     map[uuid.UUID]*Client
	roomClients map[uuid.UUID]map[uuid.UUID]*Client
	register    chan *Client
	unregister  chan *Client
	broadcast   chan *BroadcastMessage
	deliver     chan *redis.Message

	queueService      *service.QueueService
	messageService    *service.MessageService
	roomService       *service.RoomService
	moderationService *service.ModerationService
	rdb               *redis.Client
	pubsub            *redis.PubSub

	mutex          sync.RWMutex
	roomMutex      sync.RWMutex
	roomLocalCount map[uuid.UUID]int // number of local clients per room
}

type BroadcastMessage struct {
	RoomID  uuid.UUID
	Message []byte
	Exclude uuid.UUID
}

func NewHub(queueService *service.QueueService, messageService *service.MessageService, roomService *service.RoomService, moderationService *service.ModerationService, rdb *redis.Client) *Hub {
	pubsub := rdb.Subscribe(context.Background()) // empty subscription; channels added dynamically
	return &Hub{
		clients:           make(map[uuid.UUID]*Client),
		roomClients:       make(map[uuid.UUID]map[uuid.UUID]*Client),
		register:          make(chan *Client, 64),
		unregister:        make(chan *Client, 64),
		broadcast:         make(chan *BroadcastMessage, 256),
		deliver:           make(chan *redis.Message, 256),
		queueService:      queueService,
		messageService:    messageService,
		roomService:       roomService,
		moderationService: moderationService,
		rdb:               rdb,
		pubsub:            pubsub,
		roomLocalCount:    make(map[uuid.UUID]int),
	}
}

func (h *Hub) Run() {
	go h.runPubSubListener()
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)
		case client := <-h.unregister:
			h.unregisterClient(client)
		case message := <-h.broadcast:
			h.publishToRedis(message)
		case msg := <-h.deliver:
			h.handleRedisMessage(msg)
		}
	}
}

// runPubSubListener forwards incoming Redis pub/sub messages into the deliver channel.
func (h *Hub) runPubSubListener() {
	ch := h.pubsub.Channel()
	for msg := range ch {
		select {
		case h.deliver <- msg:
		default:
			slog.Warn("Hub deliver channel full, dropping Redis message", "channel", msg.Channel)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.mutex.Lock()
	h.clients[client.UserID] = client
	h.mutex.Unlock()

	// Each Hub instance subscribes to its local users' channels for direct notifications.
	if err := h.pubsub.Subscribe(context.Background(), "user:"+client.UserID.String()); err != nil {
		slog.Error("Failed to subscribe to user channel", "user_id", client.UserID, "error", err)
	}

	metrics.ActiveConnections.Inc()
	slog.Info("Client registered", "user_id", client.UserID, "client_id", client.ID)

	client.SendJSON(WSMessage{
		Type: "connected",
		Payload: map[string]interface{}{
			"user_id":   client.UserID.String(),
			"message":   "Connected to WebSocket",
			"timestamp": time.Now().Unix(),
		},
	})

	// Auto-rejoin active room if exists.
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		room, err := h.roomService.GetActiveRoomByUserID(ctx, client.UserID)
		if err != nil || room == nil {
			return
		}

		h.AddClientToRoom(client.UserID, room.ID)
		slog.Info("Client auto-rejoined room", "user_id", client.UserID, "room_id", room.ID)

		client.SendJSON(WSMessage{
			Type: "room_rejoined",
			Payload: map[string]interface{}{
				"room_id":   room.ID.String(),
				"timestamp": time.Now().Unix(),
			},
		})
	}()
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

	if isLatest {
		if err := h.pubsub.Unsubscribe(context.Background(), "user:"+client.UserID.String()); err != nil {
			slog.Error("Failed to unsubscribe from user channel", "user_id", client.UserID, "error", err)
		}
		h.queueService.UserDisconnected(client.UserID)
		metrics.ActiveConnections.Dec()
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

	if client == nil {
		slog.Warn("AddClientToRoom: client not found in hub", "user_id", userID, "room_id", roomID)
		return
	}

	client.RoomID = &roomID
	h.roomClients[roomID][userID] = client

	// Subscribe to the room channel when the first local client joins.
	prev := h.roomLocalCount[roomID]
	h.roomLocalCount[roomID]++
	if prev == 0 {
		if err := h.pubsub.Subscribe(context.Background(), "room:"+roomID.String()); err != nil {
			slog.Error("Failed to subscribe to room channel", "room_id", roomID, "error", err)
		} else {
			slog.Info("Subscribed to room channel", "room_id", roomID)
		}
	}

	slog.Info("Client added to room", "user_id", userID, "room_id", roomID, "room_size", len(h.roomClients[roomID]))
}

func (h *Hub) removeClientFromRoom(client *Client, roomID uuid.UUID) {
	h.roomMutex.Lock()
	defer h.roomMutex.Unlock()

	roomUsers, exists := h.roomClients[roomID]
	if !exists {
		return
	}

	// Only remove if this pointer matches (not a newer reconnection).
	if roomUsers[client.UserID] != client {
		return
	}

	delete(roomUsers, client.UserID)
	h.roomLocalCount[roomID]--

	if h.roomLocalCount[roomID] <= 0 {
		delete(h.roomLocalCount, roomID)
		delete(h.roomClients, roomID)
		if err := h.pubsub.Unsubscribe(context.Background(), "room:"+roomID.String()); err != nil {
			slog.Error("Failed to unsubscribe from room channel", "room_id", roomID, "error", err)
		} else {
			slog.Info("Unsubscribed from room channel", "room_id", roomID)
		}
	}

	slog.Info("Client removed from room", "user_id", client.UserID, "room_id", roomID)
}

// publishToRedis publishes a broadcast message to the Redis room channel.
// All Hub instances subscribed to that room will receive it and deliver locally.
func (h *Hub) publishToRedis(msg *BroadcastMessage) {
	pub := roomPubMsg{
		Exclude: msg.Exclude.String(),
		Payload: msg.Message,
	}
	data, err := json.Marshal(pub)
	if err != nil {
		slog.Error("Failed to marshal room pub message", "error", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	if err := h.rdb.Publish(ctx, "room:"+msg.RoomID.String(), data).Err(); err != nil {
		slog.Error("Failed to publish room message to Redis", "error", err, "room_id", msg.RoomID)
	}
}

// handleRedisMessage routes an incoming Redis pub/sub message to the correct handler.
func (h *Hub) handleRedisMessage(msg *redis.Message) {
	switch {
	case strings.HasPrefix(msg.Channel, "room:"):
		h.handleRoomMessage(msg.Channel, []byte(msg.Payload))
	case strings.HasPrefix(msg.Channel, "user:"):
		h.handleUserMessage(msg.Channel, []byte(msg.Payload))
	default:
		slog.Warn("Unknown Redis pub/sub channel", "channel", msg.Channel)
	}
}

// handleRoomMessage delivers a room broadcast to local clients in that room.
func (h *Hub) handleRoomMessage(channel string, data []byte) {
	roomIDStr := strings.TrimPrefix(channel, "room:")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		slog.Error("Invalid room ID in Redis channel", "channel", channel)
		return
	}

	var pub roomPubMsg
	if err := json.Unmarshal(data, &pub); err != nil {
		slog.Error("Failed to unmarshal room pub message", "error", err)
		return
	}
	excludeID, _ := uuid.Parse(pub.Exclude)

	h.roomMutex.RLock()
	roomUsers := h.roomClients[roomID]
	targets := make([]*Client, 0, len(roomUsers))
	targetIDs := make([]uuid.UUID, 0, len(roomUsers))
	for userID, client := range roomUsers {
		if userID != excludeID {
			targets = append(targets, client)
			targetIDs = append(targetIDs, userID)
		}
	}
	h.roomMutex.RUnlock()

	slog.Info("Delivering room message from Redis", "room_id", roomID, "targets", len(targets))

	var staleClients []*Client
	for i, client := range targets {
		select {
		case client.Send <- pub.Payload:
			slog.Info("Message delivered to client", "user_id", targetIDs[i], "room_id", roomID)
		default:
			slog.Warn("Client send channel full, marking for unregister", "user_id", client.UserID)
			staleClients = append(staleClients, client)
		}
	}
	for _, client := range staleClients {
		select {
		case h.unregister <- client:
		default:
			go func(c *Client) { h.unregister <- c }(client)
		}
	}
}

// handleUserMessage handles direct notifications delivered to a specific user via Redis.
func (h *Hub) handleUserMessage(channel string, data []byte) {
	userIDStr := strings.TrimPrefix(channel, "user:")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		slog.Error("Invalid user ID in Redis channel", "channel", channel)
		return
	}

	var pub userPubMsg
	if err := json.Unmarshal(data, &pub); err != nil {
		slog.Error("Failed to unmarshal user pub message", "error", err)
		return
	}

	switch pub.Type {
	case "match_found":
		roomID, err := uuid.Parse(pub.RoomID)
		if err != nil {
			slog.Error("Invalid room ID in match_found event", "error", err)
			return
		}

		matchMsg := WSMessage{
			Type: "match_found",
			Payload: map[string]interface{}{
				"room_id":   pub.RoomID,
				"timestamp": time.Now().Unix(),
				"message":   "Match found! You are now connected.",
			},
		}
		msgBytes, _ := json.Marshal(matchMsg)

		h.mutex.RLock()
		client := h.clients[userID]
		h.mutex.RUnlock()

		if client != nil {
			select {
			case client.Send <- msgBytes:
			default:
				slog.Warn("Client send channel full for match_found", "user_id", userID)
			}
			h.AddClientToRoom(userID, roomID)
		}
	default:
		slog.Warn("Unknown user pub/sub message type", "type", pub.Type, "user_id", userID)
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
