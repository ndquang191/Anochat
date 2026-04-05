package ws

import (
	"context"
	"encoding/json"
	"log/slog"
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
