package ws

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
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

	msgBytes, _ := json.Marshal(notification)
	h.broadcast <- &BroadcastMessage{
		RoomID:  roomID,
		Message: msgBytes,
		Exclude: leaverID,
	}
}

func (h *Hub) NotifyMatch(user1ID, user2ID, roomID uuid.UUID) {
	matchMsg := WSMessage{
		Type: "match_found",
		Payload: map[string]interface{}{
			"room_id":   roomID.String(),
			"timestamp": time.Now().Unix(),
			"message":   "Match found! You are now connected.",
		},
	}

	msgBytes, _ := json.Marshal(matchMsg)

	h.mutex.RLock()
	if client1 := h.clients[user1ID]; client1 != nil {
		client1.Send <- msgBytes
	}
	if client2 := h.clients[user2ID]; client2 != nil {
		client2.Send <- msgBytes
	}
	h.mutex.RUnlock()

	h.AddClientToRoom(user1ID, roomID)
	h.AddClientToRoom(user2ID, roomID)
}
