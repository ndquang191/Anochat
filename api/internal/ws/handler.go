package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

func (c *Client) handleMessage(message []byte) {
	var wsMsg WSMessage
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

func (c *Client) handleSendMessage(payload map[string]interface{}) {
	if c.RoomID == nil {
		slog.Warn("User tried to send message without room", "user_id", c.UserID)
		return
	}

	content, ok := payload["content"].(string)
	if !ok || content == "" {
		return
	}

	if !c.Hub.CheckMessageRateLimit(c.UserID) {
		slog.Warn("Message rate limit exceeded", "user_id", c.UserID)
		errorMsg := WSMessage{
			Type: "error",
			Payload: map[string]interface{}{
				"message": "You are sending messages too quickly. Please slow down.",
				"code":    "RATE_LIMIT_EXCEEDED",
			},
		}
		c.SendJSON(errorMsg)
		return
	}

	msg, err := c.Hub.messageService.CreateMessage(context.Background(), *c.RoomID, c.UserID, content)
	if err != nil {
		slog.Error("Failed to save message", "error", err, "user_id", c.UserID)
		return
	}

	broadcastMsg := WSMessage{
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
		Exclude: uuid.Nil,
	}

	slog.Info("Message sent", "user_id", c.UserID, "room_id", c.RoomID, "message_id", msg.ID)
}

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

	room, err := c.Hub.roomService.GetRoomByID(context.Background(), roomID)
	if err != nil || room == nil {
		slog.Error("Room not found", "room_id", roomID, "user_id", c.UserID)
		return
	}

	if !room.HasUser(c.UserID) {
		slog.Warn("User not authorized for room", "room_id", roomID, "user_id", c.UserID)
		return
	}

	c.Hub.AddClientToRoom(c.UserID, roomID)

	confirmation := WSMessage{
		Type: "room_joined",
		Payload: map[string]interface{}{
			"room_id":   roomID.String(),
			"timestamp": time.Now().Unix(),
		},
	}
	c.SendJSON(confirmation)

	slog.Info("User joined room via WebSocket", "user_id", c.UserID, "room_id", roomID)
}

func (c *Client) handleLeaveRoom(payload map[string]interface{}) {
	if c.RoomID == nil {
		return
	}

	roomID := *c.RoomID

	if err := c.Hub.roomService.LeaveRoom(context.Background(), roomID, c.UserID); err != nil {
		slog.Error("Failed to leave room in database", "error", err, "user_id", c.UserID, "room_id", roomID)
	}

	c.Hub.removeClientFromRoom(c.UserID, roomID)
	c.Hub.notifyPartnerLeft(roomID, c.UserID)
	c.RoomID = nil

	confirmation := WSMessage{
		Type: "room_left",
		Payload: map[string]interface{}{
			"room_id":   roomID.String(),
			"timestamp": time.Now().Unix(),
		},
	}
	c.SendJSON(confirmation)

	slog.Info("User left room", "user_id", c.UserID, "room_id", roomID)
}

func (c *Client) handleTyping(payload map[string]interface{}) {
	if c.RoomID == nil {
		return
	}

	isTyping, ok := payload["is_typing"].(bool)
	if !ok {
		return
	}

	typingMsg := WSMessage{
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
