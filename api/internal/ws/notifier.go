package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/matching"
)

func (h *Hub) notifyPartnerLeft(roomID uuid.UUID, leaverID uuid.UUID) {
	notification := WSMessage{
		Type: "partner_left",
		Payload: map[string]interface{}{
			"room_id":   roomID.String(),
			"timestamp": time.Now().Unix(),
			"message":   "Your chat partner has left the room",
		},
	}

	msgBytes, err := json.Marshal(notification)
	if err != nil {
		slog.Error("Failed to marshal partner_left notification", "error", err, "room_id", roomID)
		return
	}
	h.broadcast <- &BroadcastMessage{
		RoomID:  roomID,
		Message: msgBytes,
		Exclude: leaverID,
	}
}

// NotifyMatch publishes match_found events to each user's Redis channel.
// The receiving Hub instance (wherever each user is connected) will deliver
// the message locally and add the client to the room.
func (h *Hub) NotifyMatch(user1ID, user2ID, roomID uuid.UUID) {
	event := userPubMsg{
		Type:   "match_found",
		RoomID: roomID.String(),
	}
	data, err := json.Marshal(event)
	if err != nil {
		slog.Error("Failed to marshal match_found event", "error", err)
		return
	}

	ctx := context.Background()
	for _, userID := range []uuid.UUID{user1ID, user2ID} {
		if err := h.rdb.Publish(ctx, "user:"+userID.String(), data).Err(); err != nil {
			slog.Error("Failed to publish match_found to user channel", "user_id", userID, "error", err)
		}
	}
}

func (h *Hub) NotifyFakeMatch(session *matching.FakeSession) {
	matchMsg := WSMessage{
		Type: "match_found",
		Payload: map[string]interface{}{
			"room_id":   session.RoomID.String(),
			"timestamp": time.Now().Unix(),
			"message":   "Match found! You are now connected.",
		},
	}
	matchBytes, _ := json.Marshal(matchMsg)

	h.mutex.RLock()
	client := h.clients[session.UserID]
	h.mutex.RUnlock()
	if client == nil {
		return
	}

	client.Send <- matchBytes
	h.AddClientToRoom(session.UserID, session.RoomID)

	greetingMsg := WSMessage{
		Type: "receive_message",
		Payload: map[string]interface{}{
			"id":         session.GreetingID.String(),
			"room_id":    session.RoomID.String(),
			"sender_id":  session.PartnerID.String(),
			"content":    session.Greeting,
			"created_at": session.CreatedAt.Unix(),
		},
	}
	greetingBytes, _ := json.Marshal(greetingMsg)
	client.Send <- greetingBytes
}

func (h *Hub) NotifyFakePartnerLeft(userID, roomID uuid.UUID) {
	notification := WSMessage{
		Type: "partner_left",
		Payload: map[string]interface{}{
			"room_id":   roomID.String(),
			"timestamp": time.Now().Unix(),
			"message":   "Your chat partner has left the room",
		},
	}

	msgBytes, _ := json.Marshal(notification)

	h.mutex.RLock()
	client := h.clients[userID]
	h.mutex.RUnlock()
	if client != nil {
		client.RoomID = nil
		client.Send <- msgBytes
	}

	h.roomMutex.Lock()
	if roomUsers, exists := h.roomClients[roomID]; exists {
		delete(roomUsers, userID)
		if len(roomUsers) == 0 {
			delete(h.roomClients, roomID)
		}
	}
	h.roomMutex.Unlock()
}
