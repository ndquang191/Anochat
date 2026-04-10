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

	if c.Hub.moderationService.ContainsBannedWord(content) {
		slog.Info("Auto-reporting user for banned word", "user_id", c.UserID, "room_id", *c.RoomID)
		go func(roomID uuid.UUID) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := c.Hub.moderationService.CreateReport(ctx, uuid.Nil, c.UserID, roomID); err != nil {
				slog.Error("Failed to create auto-report", "error", err, "user_id", c.UserID)
			}
		}(*c.RoomID)
	}

	msgID := uuid.New()
	now := time.Now().UTC()
	roomID := *c.RoomID
	isFakeRoom := c.Hub.fakeMatchService.IsParticipant(c.UserID, roomID)

	broadcastMsg := WSMessage{
		Type: "receive_message",
		Payload: map[string]interface{}{
			"id":         msgID.String(),
			"room_id":    roomID.String(),
			"sender_id":  c.UserID.String(),
			"content":    content,
			"created_at": now.Unix(),
		},
	}

	msgBytes, err := json.Marshal(broadcastMsg)
	if err != nil {
		slog.Error("Failed to marshal broadcast message", "error", err, "user_id", c.UserID, "room_id", roomID)
		return
	}
	c.Hub.broadcast <- &BroadcastMessage{
		RoomID:  roomID,
		Message: msgBytes,
		Exclude: c.UserID,
	}

	if isFakeRoom {
		reply := c.Hub.fakeMatchService.AdvanceConversation(c.UserID, roomID)
		c.Hub.NotifyFakeMessage(c.UserID, reply)
		slog.Info("Fake room message handled", "user_id", c.UserID, "room_id", roomID)
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := c.Hub.messageService.SaveMessage(ctx, msgID, roomID, c.UserID, content, now); err != nil {
			slog.Error("Failed to save message", "error", err, "user_id", c.UserID, "message_id", msgID)
		}
	}()

	slog.Info("Message sent", "user_id", c.UserID, "room_id", roomID, "message_id", msgID)
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
		if !c.Hub.fakeMatchService.IsParticipant(c.UserID, roomID) {
			slog.Error("Room not found", "room_id", roomID, "user_id", c.UserID)
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
		slog.Info("User joined fake room via WebSocket", "user_id", c.UserID, "room_id", roomID)
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

	if c.Hub.fakeMatchService.IsParticipant(c.UserID, roomID) {
		c.Hub.fakeMatchService.EndSession(c.UserID)
	} else if err := c.Hub.roomService.LeaveRoom(context.Background(), roomID, c.UserID); err != nil {
		slog.Error("Failed to leave room in database", "error", err, "user_id", c.UserID, "room_id", roomID)
	}

	c.Hub.removeClientFromRoom(c, roomID)
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
