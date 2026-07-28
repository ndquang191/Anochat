package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/chat"
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
	messageIDRaw, ok := payload["id"].(string)
	if !ok {
		c.sendMessageFailed("", "INVALID_MESSAGE_ID", "Message ID is required.")
		return
	}
	messageID, err := uuid.Parse(messageIDRaw)
	if err != nil {
		c.sendMessageFailed(messageIDRaw, "INVALID_MESSAGE_ID", "Message ID is invalid.")
		return
	}

	if c.RoomID == nil {
		slog.Warn("User tried to send message without room", "user_id", c.UserID)
		c.sendMessageFailed(messageIDRaw, "NOT_IN_ROOM", "You are not in a chat room.")
		return
	}

	content, ok := payload["content"].(string)
	if !ok || strings.TrimSpace(content) == "" {
		c.sendMessageFailed(messageIDRaw, "INVALID_MESSAGE", "Message content is required.")
		return
	}
	if messageExceedsLimit(content, c.Hub.maxMessageLength) {
		c.sendMessageFailed(messageIDRaw, "MESSAGE_TOO_LONG", "Message is too long.")
		return
	}

	if !c.Hub.CheckMessageRateLimit(c.UserID) {
		slog.Warn("Message rate limit exceeded", "user_id", c.UserID)
		c.sendMessageFailed(
			messageIDRaw,
			"RATE_LIMIT_EXCEEDED",
			"You are sending messages too quickly. Please slow down.",
		)
		return
	}

	now := time.Now().UTC()
	roomID := *c.RoomID
	isSensitive := c.Hub.moderationService.ContainsBannedWord(content)
	triggerMessage := &chat.Message{
		ID:        messageID,
		RoomID:    roomID,
		SenderID:  c.UserID,
		Content:   content,
		CreatedAt: now,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.Hub.messageService.SaveMessage(ctx, messageID, roomID, c.UserID, content, now); err != nil {
		slog.Error("Failed to save message", "error", err, "user_id", c.UserID, "message_id", messageID)
		c.sendMessageFailed(messageIDRaw, "MESSAGE_SAVE_FAILED", "Message could not be saved.")
		return
	}

	broadcastMsg := WSMessage{
		Type: "receive_message",
		Payload: map[string]interface{}{
			"id":         messageID.String(),
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

	c.sendMessageAck(messageID, now)

	if isSensitive {
		go c.createAutoReport(triggerMessage)
	}

	slog.Info("Message sent", "user_id", c.UserID, "room_id", roomID, "message_id", messageID)
}

func (c *Client) sendMessageFailed(id, code, message string) {
	c.SendJSON(WSMessage{
		Type: "message_failed",
		Payload: map[string]interface{}{
			"id":      id,
			"code":    code,
			"message": message,
		},
	})
}

func (c *Client) sendMessageAck(id uuid.UUID, createdAt time.Time) {
	c.SendJSON(WSMessage{
		Type: "message_ack",
		Payload: map[string]interface{}{
			"id":         id.String(),
			"created_at": createdAt.Unix(),
		},
	})
}

func (c *Client) createAutoReport(triggerMessage *chat.Message) {
	slog.Info(
		"Auto-reporting user for banned word",
		"user_id", c.UserID,
		"room_id", triggerMessage.RoomID,
		"message_id", triggerMessage.ID,
	)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.Hub.moderationService.CreateAutoReport(
		ctx,
		c.UserID,
		triggerMessage.RoomID,
		triggerMessage,
	); err != nil {
		slog.Error("Failed to create auto-report", "error", err, "user_id", c.UserID)
	}
}

func messageExceedsLimit(content string, maxLength int) bool {
	return utf8.RuneCountInString(content) > maxLength
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
		c.sendRoomLeaveFailed("", "NOT_IN_ROOM", "You are not in a chat room.")
		return
	}

	roomID := *c.RoomID

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.Hub.roomService.LeaveRoom(ctx, roomID, c.UserID); err != nil {
		slog.Error("Failed to leave room in database", "error", err, "user_id", c.UserID, "room_id", roomID)
		c.sendRoomLeaveFailed(roomID.String(), "ROOM_LEAVE_FAILED", "The chat room could not be left.")
		return
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

func (c *Client) sendRoomLeaveFailed(roomID, code, message string) {
	c.SendJSON(WSMessage{
		Type: "room_leave_failed",
		Payload: map[string]interface{}{
			"room_id": roomID,
			"code":    code,
			"message": message,
		},
	})
}
