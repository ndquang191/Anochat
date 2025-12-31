package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/ndquang191/Anochat/api/internal/service"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client represents a WebSocket client connection
type Client struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Conn         *websocket.Conn
	Hub          *Hub
	Send         chan []byte
	RoomID       *uuid.UUID
	LastActivity time.Time
}

// Hub manages all WebSocket connections
type Hub struct {
	clients        map[uuid.UUID]*Client               // userID -> client
	roomClients    map[uuid.UUID]map[uuid.UUID]*Client // roomID -> userID -> client
	register       chan *Client
	unregister     chan *Client
	broadcast      chan *BroadcastMessage
	queueService   *service.QueueService
	messageService *service.MessageService
	roomService    *service.RoomService
	mutex          sync.RWMutex
	roomMutex      sync.RWMutex
}

// BroadcastMessage represents a message to broadcast
type BroadcastMessage struct {
	RoomID  uuid.UUID
	Message []byte
	Exclude uuid.UUID // Exclude this user from broadcast
}

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

// NewHub creates a new WebSocket hub
func NewHub(queueService *service.QueueService, messageService *service.MessageService, roomService *service.RoomService) *Hub {
	return &Hub{
		clients:        make(map[uuid.UUID]*Client),
		roomClients:    make(map[uuid.UUID]map[uuid.UUID]*Client),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		broadcast:      make(chan *BroadcastMessage, 256),
		queueService:   queueService,
		messageService: messageService,
		roomService:    roomService,
	}
}

// Run starts the hub's main loop
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

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mutex.Lock()
	h.clients[client.UserID] = client
	h.mutex.Unlock()

	slog.Info("Client registered", "user_id", client.UserID, "client_id", client.ID)

	// Send connection success message
	successMsg := WebSocketMessage{
		Type: "connected",
		Payload: map[string]interface{}{
			"user_id":   client.UserID.String(),
			"message":   "Connected to WebSocket",
			"timestamp": time.Now().Unix(),
		},
	}
	client.SendJSON(successMsg)
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.mutex.Lock()
	if _, exists := h.clients[client.UserID]; exists {
		delete(h.clients, client.UserID)
	}
	h.mutex.Unlock()

	// Remove from room
	if client.RoomID != nil {
		// Update database - mark room as ended
		if err := h.roomService.LeaveRoom(context.Background(), *client.RoomID, client.UserID); err != nil {
			slog.Error("Failed to leave room in database on disconnect", "error", err, "user_id", client.UserID, "room_id", *client.RoomID)
			// Continue with cleanup even if DB update fails
		}

		h.removeClientFromRoom(client.UserID, *client.RoomID)

		// Notify partner that user left
		h.notifyPartnerLeft(*client.RoomID, client.UserID)
	}

	// Notify queue service
	h.queueService.UserDisconnected(client.UserID)

	close(client.Send)
	slog.Info("Client unregistered", "user_id", client.UserID, "client_id", client.ID)
}

// addClientToRoom adds a client to a room
func (h *Hub) addClientToRoom(userID uuid.UUID, roomID uuid.UUID) {
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

// removeClientFromRoom removes a client from a room
func (h *Hub) removeClientFromRoom(userID uuid.UUID, roomID uuid.UUID) {
	h.roomMutex.Lock()
	defer h.roomMutex.Unlock()

	if roomUsers, exists := h.roomClients[roomID]; exists {
		delete(roomUsers, userID)

		// Clean up empty rooms
		if len(roomUsers) == 0 {
			delete(h.roomClients, roomID)
		}

		slog.Info("Client removed from room", "user_id", userID, "room_id", roomID)
	}
}

// broadcastToRoom broadcasts a message to all clients in a room
func (h *Hub) broadcastToRoom(msg *BroadcastMessage) {
	h.roomMutex.RLock()
	roomUsers := h.roomClients[msg.RoomID]
	h.roomMutex.RUnlock()

	for userID, client := range roomUsers {
		if userID != msg.Exclude {
			select {
			case client.Send <- msg.Message:
			default:
				// Client's send channel is full, unregister
				h.unregister <- client
			}
		}
	}
}

// notifyPartnerLeft notifies the partner that the user left
func (h *Hub) notifyPartnerLeft(roomID uuid.UUID, leaverID uuid.UUID) {
	notification := WebSocketMessage{
		Type: "partner_left",
		Payload: map[string]interface{}{
			"room_id":   roomID.String(),
			"timestamp": time.Now().Unix(),
			"message":   "Your chat partner has left the room",
		},
	}

	msgBytes, _ := json.Marshal(notification)
	h.broadcast <- &BroadcastMessage{
		RoomID:  roomID,
		Message: msgBytes,
		Exclude: leaverID,
	}
}

// NotifyMatch notifies users when they are matched
func (h *Hub) NotifyMatch(user1ID, user2ID, roomID uuid.UUID, category string) {
	matchMsg := WebSocketMessage{
		Type: "match_found",
		Payload: map[string]interface{}{
			"room_id":   roomID.String(),
			"category":  category,
			"timestamp": time.Now().Unix(),
			"message":   "Match found! You are now connected.",
		},
	}

	msgBytes, _ := json.Marshal(matchMsg)

	// Send to both users
	h.mutex.RLock()
	if client1 := h.clients[user1ID]; client1 != nil {
		client1.Send <- msgBytes
	}
	if client2 := h.clients[user2ID]; client2 != nil {
		client2.Send <- msgBytes
	}
	h.mutex.RUnlock()

	// Add both users to the room
	h.addClientToRoom(user1ID, roomID)
	h.addClientToRoom(user2ID, roomID)
}

// SendJSON sends a JSON message to the client
func (c *Client) SendJSON(msg WebSocketMessage) {
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

// readPump reads messages from the WebSocket connection
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
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

// writePump writes messages to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage handles incoming WebSocket messages
func (c *Client) handleMessage(message []byte) {
	var wsMsg WebSocketMessage
	if err := json.Unmarshal(message, &wsMsg); err != nil {
		slog.Error("Failed to unmarshal message", "error", err, "user_id", c.UserID)
		return
	}

	switch wsMsg.Type {
	case "send_message":
		c.handleSendMessage(wsMsg.Payload)
	case "join_room":
		c.handleJoinRoom(wsMsg.Payload)
	case "leave_room":
		c.handleLeaveRoom(wsMsg.Payload)
	case "typing":
		c.handleTyping(wsMsg.Payload)
	default:
		slog.Warn("Unknown message type", "type", wsMsg.Type, "user_id", c.UserID)
	}
}

// handleSendMessage handles sending a chat message
func (c *Client) handleSendMessage(payload map[string]interface{}) {
	if c.RoomID == nil {
		slog.Warn("User tried to send message without room", "user_id", c.UserID)
		return
	}

	content, ok := payload["content"].(string)
	if !ok || content == "" {
		return
	}

	// Save message to database
	msg, err := c.Hub.messageService.CreateMessage(context.Background(), *c.RoomID, c.UserID, content)
	if err != nil {
		slog.Error("Failed to save message", "error", err, "user_id", c.UserID)
		return
	}

	// Broadcast to room
	broadcastMsg := WebSocketMessage{
		Type: "receive_message",
		Payload: map[string]interface{}{
			"id":         msg.ID.String(),
			"room_id":    msg.RoomID.String(),
			"sender_id":  msg.SenderID.String(),
			"content":    msg.Content,
			"created_at": msg.CreatedAt.Unix(),
		},
	}

	msgBytes, _ := json.Marshal(broadcastMsg)
	c.Hub.broadcast <- &BroadcastMessage{
		RoomID:  *c.RoomID,
		Message: msgBytes,
		Exclude: uuid.Nil, // Send to everyone including sender for confirmation
	}

	slog.Info("Message sent", "user_id", c.UserID, "room_id", c.RoomID, "message_id", msg.ID)
}

// handleJoinRoom handles joining a room
func (c *Client) handleJoinRoom(payload map[string]interface{}) {
	roomIDStr, ok := payload["room_id"].(string)
	if !ok {
		return
	}

	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		slog.Error("Invalid room ID", "error", err, "user_id", c.UserID)
		return
	}

	// Verify user is in this room
	room, err := c.Hub.roomService.GetRoomByID(context.Background(), roomID)
	if err != nil || room == nil {
		slog.Error("Room not found", "room_id", roomID, "user_id", c.UserID)
		return
	}

	if room.User1ID != c.UserID && room.User2ID != c.UserID {
		slog.Warn("User not authorized for room", "room_id", roomID, "user_id", c.UserID)
		return
	}

	c.Hub.addClientToRoom(c.UserID, roomID)

	// Send room joined confirmation
	confirmation := WebSocketMessage{
		Type: "room_joined",
		Payload: map[string]interface{}{
			"room_id":   roomID.String(),
			"timestamp": time.Now().Unix(),
		},
	}
	c.SendJSON(confirmation)

	slog.Info("User joined room via WebSocket", "user_id", c.UserID, "room_id", roomID)
}

// handleLeaveRoom handles leaving a room
func (c *Client) handleLeaveRoom(payload map[string]interface{}) {
	if c.RoomID == nil {
		return
	}

	roomID := *c.RoomID

	// Update database - mark room as ended
	if err := c.Hub.roomService.LeaveRoom(context.Background(), roomID, c.UserID); err != nil {
		slog.Error("Failed to leave room in database", "error", err, "user_id", c.UserID, "room_id", roomID)
		// Continue with WebSocket cleanup even if DB update fails
	}

	c.Hub.removeClientFromRoom(c.UserID, roomID)
	c.Hub.notifyPartnerLeft(roomID, c.UserID)
	c.RoomID = nil

	// Send leave confirmation
	confirmation := WebSocketMessage{
		Type: "room_left",
		Payload: map[string]interface{}{
			"room_id":   roomID.String(),
			"timestamp": time.Now().Unix(),
		},
	}
	c.SendJSON(confirmation)

	slog.Info("User left room", "user_id", c.UserID, "room_id", roomID)
}

// handleTyping handles typing indicator
func (c *Client) handleTyping(payload map[string]interface{}) {
	if c.RoomID == nil {
		return
	}

	isTyping, ok := payload["is_typing"].(bool)
	if !ok {
		return
	}

	typingMsg := WebSocketMessage{
		Type: "partner_typing",
		Payload: map[string]interface{}{
			"is_typing": isTyping,
			"user_id":   c.UserID.String(),
		},
	}

	msgBytes, _ := json.Marshal(typingMsg)
	c.Hub.broadcast <- &BroadcastMessage{
		RoomID:  *c.RoomID,
		Message: msgBytes,
		Exclude: c.UserID,
	}
}

// WebSocketHandler handles WebSocket connection upgrades
type WebSocketHandler struct {
	hub         *Hub
	authService *service.AuthService
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(hub *Hub, authService *service.AuthService) *WebSocketHandler {
	return &WebSocketHandler{
		hub:         hub,
		authService: authService,
	}
}

// HandleWebSocket handles WebSocket connection requests
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("Failed to upgrade connection", "error", err, "user_id", userID)
		return
	}

	// Create new client
	client := &Client{
		ID:           uuid.New(),
		UserID:       userID,
		Conn:         conn,
		Hub:          h.hub,
		Send:         make(chan []byte, 256),
		RoomID:       nil,
		LastActivity: time.Now(),
	}

	// Register client
	h.hub.register <- client

	// Start read and write pumps
	go client.writePump()
	go client.readPump()

	slog.Info("WebSocket connection established", "user_id", userID, "client_id", client.ID)
}
