package ws

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMessageExceedsLimitCountsUnicodeCharacters(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		maxLength  int
		wantExceed bool
	}{
		{name: "at limit", content: "hello", maxLength: 5, wantExceed: false},
		{name: "over limit", content: "hello!", maxLength: 5, wantExceed: true},
		{name: "unicode at limit", content: "xin chào 👋", maxLength: 10, wantExceed: false},
		{name: "unicode over limit", content: "xin chào 👋!", maxLength: 10, wantExceed: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := messageExceedsLimit(tt.content, tt.maxLength); got != tt.wantExceed {
				t.Errorf("messageExceedsLimit(%q, %d) = %v, want %v", tt.content, tt.maxLength, got, tt.wantExceed)
			}
		})
	}
}

func TestMaxInboundMessageSizeIncludesJSONEncodingOverhead(t *testing.T) {
	if got, minimum := maxInboundMessageSize(1000), int64(6000); got <= minimum {
		t.Errorf("maxInboundMessageSize(1000) = %d, want greater than %d", got, minimum)
	}
}

func TestSendMessageFailedIncludesClientMessageID(t *testing.T) {
	client := &Client{Send: make(chan []byte, 1)}

	client.sendMessageFailed("client-message-id", "MESSAGE_SAVE_FAILED", "Message could not be saved.")

	var message WSMessage
	if err := json.Unmarshal(<-client.Send, &message); err != nil {
		t.Fatalf("unmarshal message: %v", err)
	}
	if message.Type != "message_failed" {
		t.Fatalf("message type = %q, want message_failed", message.Type)
	}
	if got := message.Payload["id"]; got != "client-message-id" {
		t.Errorf("message id = %v, want client-message-id", got)
	}
	if got := message.Payload["code"]; got != "MESSAGE_SAVE_FAILED" {
		t.Errorf("message code = %v, want MESSAGE_SAVE_FAILED", got)
	}
}

func TestHandleSendMessageRejectsInvalidClientMessageID(t *testing.T) {
	client := &Client{Send: make(chan []byte, 1)}

	client.handleSendMessage(map[string]interface{}{
		"id":      "not-a-uuid",
		"content": "hello",
	})

	var message WSMessage
	if err := json.Unmarshal(<-client.Send, &message); err != nil {
		t.Fatalf("unmarshal message: %v", err)
	}
	if message.Type != "message_failed" {
		t.Fatalf("message type = %q, want message_failed", message.Type)
	}
	if got := message.Payload["id"]; got != "not-a-uuid" {
		t.Errorf("message id = %v, want not-a-uuid", got)
	}
	if got := message.Payload["code"]; got != "INVALID_MESSAGE_ID" {
		t.Errorf("message code = %v, want INVALID_MESSAGE_ID", got)
	}
}

func TestSendMessageAckReturnsPersistedClientMessageID(t *testing.T) {
	client := &Client{Send: make(chan []byte, 1)}
	messageID := uuid.New()
	createdAt := time.Unix(123, 0)

	client.sendMessageAck(messageID, createdAt)

	var message WSMessage
	if err := json.Unmarshal(<-client.Send, &message); err != nil {
		t.Fatalf("unmarshal message: %v", err)
	}
	if message.Type != "message_ack" {
		t.Fatalf("message type = %q, want message_ack", message.Type)
	}
	if got := message.Payload["id"]; got != messageID.String() {
		t.Errorf("message id = %v, want %s", got, messageID)
	}
	if got := message.Payload["created_at"]; got != float64(createdAt.Unix()) {
		t.Errorf("created_at = %v, want %d", got, createdAt.Unix())
	}
}

func TestSendRoomLeaveFailedIncludesRoomAndCode(t *testing.T) {
	client := &Client{Send: make(chan []byte, 1)}
	roomID := uuid.NewString()

	client.sendRoomLeaveFailed(roomID, "ROOM_LEAVE_FAILED", "leave failed")

	var message WSMessage
	if err := json.Unmarshal(<-client.Send, &message); err != nil {
		t.Fatalf("unmarshal message: %v", err)
	}
	if message.Type != "room_leave_failed" {
		t.Fatalf("message type = %q, want room_leave_failed", message.Type)
	}
	if got := message.Payload["room_id"]; got != roomID {
		t.Errorf("room id = %v, want %s", got, roomID)
	}
	if got := message.Payload["code"]; got != "ROOM_LEAVE_FAILED" {
		t.Errorf("code = %v, want ROOM_LEAVE_FAILED", got)
	}
}
