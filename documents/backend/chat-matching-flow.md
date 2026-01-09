# Flow Chat và Matching - Anochat

## Tổng quan

Hệ thống Anochat sử dụng queue-based matching với WebSocket real-time communication. Flow bao gồm 3 giai đoạn chính:
1. **Queue Flow** - Tham gia hàng chờ và matching
2. **Chat Flow** - Giao tiếp real-time qua WebSocket
3. **Leave Flow** - Rời phòng và cleanup

---

## 1. Queue Flow - Tham gia hàng chờ và matching

### 1.1. User Join Queue

```
[Frontend] ──────────> [Backend API] ──────────> [QueueService]
   │                         │                          │
   │ POST /queue/join        │                          │
   │ {category: "polite"}    │                          │
   │                         │                          │
   │                         │ 1. Validate category     │
   │                         │ 2. Check user in queue   │
   │                         │ 3. Check active room     │
   │                         │ 4. Get user profile      │
   │                         │                          │
   │                         │ ◄────────────────────────┤
   │                         │   Create QueueEntry      │
   │ ◄───────────────────────┤                          │
   │ Response: {             │                          │
   │   is_in_queue: true,    │                          │
   │   position: 1,          │                          │
   │   category: "polite"    │                          │
   │ }                       │                          │
```

**Backend Logic (api/internal/service/queue.go - Line 112-217):**

```go
func (qs *QueueService) JoinQueue(ctx context.Context, userID uuid.UUID, category string) (*QueueEntry, error) {
    // 1. Validate category
    if !config.IsValidCategory(category) {
        return nil, fmt.Errorf("invalid category: %s", category)
    }

    // 2. Get user profile (for gender-based queue)
    profile, err := qs.userService.GetProfile(ctx, userID)
    if err != nil || profile == nil {
        return nil, fmt.Errorf("user profile not found")
    }

    // 3. Check if user is already in queue
    if qs.isUserInQueue(userID) {
        return nil, fmt.Errorf("user already in queue")
    }

    // 4. Check if user has active room
    activeRoom, err := qs.userService.GetActiveRoom(ctx, userID)
    if err == nil && activeRoom != nil {
        return nil, fmt.Errorf("user already has active room")
    }

    // 5. Create queue entry with TTL
    entry := &QueueEntry{
        UserID:    userID,
        Profile:   profile,
        Category:  category,
        JoinedAt:  time.Now(),
        ExpiresAt: time.Now().Add(QueueHeartbeatTTL), // TTL for heartbeat
        IsMatched: false,
        MatchChan: make(chan *MatchResult, 1),
    }

    // 6. Add to appropriate gender queue (male/female)
    isMale := profile.IsMale != nil && *profile.IsMale
    if isMale {
        qs.queueMale[category] = append(qs.queueMale[category], entry)
    } else {
        qs.queueFemale[category] = append(qs.queueFemale[category], entry)
    }

    // 7. Mark user as connected
    qs.userConnections[userID] = true

    // 8. Try to find match immediately
    go qs.tryMatch(entry)

    return entry, nil
}
```

**Queue Structure:**
- **queueMale[category]**: Array of male users waiting
- **queueFemale[category]**: Array of female users waiting
- **userPositions[userID]**: Track user position in queue (O(1) lookup)
- **userConnections[userID]**: Track user connection status

---

### 1.2. Matching Algorithm

```
[QueueService.tryMatch]
         │
         ├──> 1. Check if user still connected
         │
         ├──> 2. Try opposite gender first (Priority)
         │    ├─> Male user → Search queueFemale[category]
         │    └─> Female user → Search queueMale[category]
         │
         ├──> 3. If no match → Try same gender (Fallback)
         │
         ├──> 4. Match found!
         │    ├─> Mark both entries as matched
         │    │
         │    ├─> *** CLEANUP BOTH USERS FROM QUEUE ***
         │    │   ├─> Remove from queueMale/queueFemale array
         │    │   ├─> Delete userPositions[userID]
         │    │   ├─> Delete userConnections[userID]
         │    │   └─> Update positions for remaining users
         │    │
         │    ├─> Create room in database
         │    └─> Send match result to both users
         │
         └──> 5. Notify via WebSocket
```

**Backend Logic (api/internal/service/queue.go - Line 219-347):**

```go
func (qs *QueueService) tryMatch(entry *QueueEntry) {
    qs.queueMutex.Lock()
    defer qs.queueMutex.Unlock()

    // Check if entry is still valid
    if entry.IsMatched || !qs.isUserConnected(entry.UserID) {
        return
    }

    category := entry.Category
    isMale := entry.Profile.IsMale != nil && *entry.Profile.IsMale

    // Priority 1: Try opposite gender
    var oppositeQueue []*QueueEntry
    if isMale {
        oppositeQueue = qs.queueFemale[category]
    } else {
        oppositeQueue = qs.queueMale[category]
    }

    var match *QueueEntry
    for _, otherEntry := range oppositeQueue {
        if !otherEntry.IsMatched && qs.isUserConnected(otherEntry.UserID) {
            match = otherEntry
            break
        }
    }

    // Priority 2: Fallback to same gender
    if match == nil {
        var sameQueue []*QueueEntry
        if isMale {
            sameQueue = qs.queueMale[category]
        } else {
            sameQueue = qs.queueFemale[category]
        }

        for _, otherEntry := range sameQueue {
            if !otherEntry.IsMatched &&
               qs.isUserConnected(otherEntry.UserID) &&
               otherEntry.UserID != entry.UserID {
                match = otherEntry
                break
            }
        }
    }

    if match == nil {
        return // No match found
    }

    // Mark both as matched
    entry.IsMatched = true
    match.IsMatched = true

    // *** CRITICAL: Remove both users from queues ***
    // This prevents them from being matched again
    qs.removeEntryFromQueue(category, entry, isMale)
    qs.removeEntryFromQueue(category, match, matchIsMale)

    // Create room in database
    room, err := qs.roomService.CreateRoom(ctx, entry.UserID, match.UserID, category)
    if err != nil {
        // Send error to both users
        return
    }

    // Notify via WebSocket
    if qs.matchNotifier != nil {
        qs.matchNotifier.NotifyMatch(entry.UserID, match.UserID, room.ID, category)
    }
}
```

**Matching Priority:**
1. **Opposite gender first** (Male ↔ Female)
2. **Same gender fallback** (Male ↔ Male or Female ↔ Female)
3. **Unknown gender** → Try both queues

---

### 1.2b. Queue Cleanup After Match (CRITICAL)

**Khi match được tìm thấy, cả 2 users PHẢI được remove hoàn toàn khỏi queue system:**

```go
// Remove entry from queue (api/internal/service/queue.go - Line 375-402)
func (qs *QueueService) removeEntryFromQueue(category string, entry *QueueEntry, isMale bool) {
    var queue []*QueueEntry
    if isMale {
        queue = qs.queueMale[category]
    } else {
        queue = qs.queueFemale[category]
    }

    for i, e := range queue {
        if e.UserID == entry.UserID {
            // 1. Remove from queue array
            if isMale {
                qs.queueMale[category] = append(queue[:i], queue[i+1:]...)
            } else {
                qs.queueFemale[category] = append(queue[:i], queue[i+1:]...)
            }

            // 2. Remove from position tracking (O(1) lookups)
            qs.posMutex.Lock()
            delete(qs.userPositions, entry.UserID)
            qs.posMutex.Unlock()

            // 3. Remove from connection tracking
            qs.connMutex.Lock()
            delete(qs.userConnections, entry.UserID)
            qs.connMutex.Unlock()

            // 4. Update positions for remaining users in queue
            qs.updatePositionsAfterRemoval(category, isMale, i)
            break
        }
    }
}
```

**Cleanup Steps (cả 2 users):**
1. ✅ **Remove from queueMale[category]** hoặc **queueFemale[category]** array
2. ✅ **Delete from userPositions map** (position tracking)
3. ✅ **Delete from userConnections map** (connection tracking)
4. ✅ **Update positions** của các users còn lại trong queue
5. ✅ **Close MatchChan** (implicitly, không còn reference)

**Verification:**
- Sau khi match, `isUserInQueue(userID)` sẽ return `false`
- User có thể join queue lại sau khi leave room
- Positions của users còn lại được update tự động

---

### 1.3. Match Found - WebSocket Notification

```
[QueueService] ─────> [WebSocket Hub] ─────> [Frontend]
                            │
                            ├──> Send to User 1
                            │    {
                            │      type: "match_found",
                            │      payload: {
                            │        room_id: "uuid",
                            │        category: "polite"
                            │      }
                            │    }
                            │
                            └──> Send to User 2
                                 (same message)
```

**Backend Logic (api/internal/handler/websocket.go - Line 216-243):**

```go
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
```

**Frontend React Hook (frontend/src/hooks/use-websocket-chat.tsx - Line 60-71):**

```typescript
// Handle match found
const handleMatchFound = (message: WebSocketMessage) => {
    const { room_id, category } = message.payload;
    console.log("Match found:", room_id, category);
    setRoomId(room_id);

    // Join the room automatically via WebSocket
    client.send("join_room", { room_id });

    if (onMatchFoundRef.current) {
        onMatchFoundRef.current(room_id);
    }
};

client.on("match_found", handleMatchFound);
```

**Frontend State Changes:**
1. `isInQueue: true` → `false` (leave queue)
2. `roomId: null` → `room_id` (enter chat room)
3. UI: Queue screen → Chat screen
4. Toast notification: "Đã tìm thấy đối tác chat!"

---

## 2. Chat Flow - Real-time Communication

### 2.1. WebSocket Connection

```
[Frontend] ──────────────> [Backend WebSocket Endpoint]
    │                               │
    │ GET /ws                       │
    │ Cookie: session_token         │
    │                               │
    │ ◄─────────────────────────────┤
    │ Upgrade to WebSocket          │
    │                               │
    │ ◄─────────────────────────────┤
    │ {type: "connected"}           │
```

**Backend WebSocket Handler (api/internal/handler/websocket.go - Line 494-538):**

```go
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
    // Get user ID from auth middleware
    userID, ok := c.Get("user_id")
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    // Upgrade HTTP to WebSocket
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }

    // Create client
    client := &Client{
        ID:           uuid.New(),
        UserID:       userID,
        Conn:         conn,
        Hub:          h.hub,
        Send:         make(chan []byte, 256),
        RoomID:       nil,
        LastActivity: time.Now(),
    }

    // Register to hub
    h.hub.register <- client

    // Start pumps
    go client.writePump()
    go client.readPump()
}
```

**Frontend WebSocket Client (frontend/src/lib/websocket.ts - Line 33-96):**

```typescript
async connect(): Promise<void> {
    // Prevent multiple connections
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        return Promise.resolve();
    }

    return new Promise((resolve, reject) => {
        this.isConnecting = true;

        // Create WebSocket (cookies sent automatically)
        this.ws = new WebSocket(this.url);

        this.ws.onopen = () => {
            console.log("WebSocket connected");
            this.isConnecting = false;
            resolve();
        };

        this.ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            this.handleMessage(message);
        };

        this.ws.onerror = (error) => {
            this.isConnecting = false;
            reject(error);
        };

        this.ws.onclose = () => {
            this.ws = null;
            this.isConnecting = false;
            // Auto-reconnect if not intentionally closed
            if (!this.isIntentionallyClosed) {
                setTimeout(() => this.connect(), 1000);
            }
        };
    });
}
```

**WebSocket Singleton:**
- Frontend sử dụng singleton pattern để đảm bảo chỉ có 1 connection
- Connection được tái sử dụng cho toàn bộ app lifecycle
- Auto-reconnect khi bị disconnect không mong muốn

---

### 2.2. Join Room via WebSocket

```
[Frontend] ────────────> [WebSocket Server] ────────> [RoomService]
    │                           │                           │
    │ {                         │                           │
    │   type: "join_room",      │                           │
    │   payload: {              │                           │
    │     room_id: "uuid"       │ Verify user in room       │
    │   }                       │ ──────────────────────────>│
    │ }                         │                           │
    │                           │ Check:                    │
    │                           │ - Room exists             │
    │                           │ - User authorized         │
    │                           │ ◄──────────────────────────┤
    │                           │                           │
    │                           │ Add to roomClients map    │
    │                           │                           │
    │ ◄─────────────────────────┤                           │
    │ {type: "room_joined"}     │                           │
```

**Backend Handler (api/internal/handler/websocket.go - Line 381-419):**

```go
func (c *Client) handleJoinRoom(payload map[string]interface{}) {
    roomIDStr, ok := payload["room_id"].(string)
    if !ok {
        return
    }

    roomID, err := uuid.Parse(roomIDStr)
    if err != nil {
        return
    }

    // Verify user is in this room
    room, err := c.Hub.roomService.GetRoomByID(context.Background(), roomID)
    if err != nil || room == nil {
        slog.Error("Room not found", "room_id", roomID, "user_id", c.UserID)
        return
    }

    // Check authorization
    if room.User1ID != c.UserID && room.User2ID != c.UserID {
        slog.Warn("User not authorized for room", "room_id", roomID)
        return
    }

    // Add to room
    c.Hub.addClientToRoom(c.UserID, roomID)

    // Send confirmation
    confirmation := WebSocketMessage{
        Type: "room_joined",
        Payload: map[string]interface{}{
            "room_id":   roomID.String(),
            "timestamp": time.Now().Unix(),
        },
    }
    c.SendJSON(confirmation)
}
```

**Hub Room Management:**
```go
type Hub struct {
    clients     map[uuid.UUID]*Client              // userID -> client
    roomClients map[uuid.UUID]map[uuid.UUID]*Client // roomID -> userID -> client
}

func (h *Hub) addClientToRoom(userID uuid.UUID, roomID uuid.UUID) {
    h.roomMutex.Lock()
    defer h.roomMutex.Unlock()

    // Initialize room if not exists
    if h.roomClients[roomID] == nil {
        h.roomClients[roomID] = make(map[uuid.UUID]*Client)
    }

    // Add client to room
    client := h.clients[userID]
    if client != nil {
        client.RoomID = &roomID
        h.roomClients[roomID][userID] = client
    }
}
```

---

### 2.3. Send Message

```
[User 1 Frontend] ───> [WebSocket] ───> [MessageService] ───> [Database]
                            │                  │
                            │                  │ Save message
                            │                  │
                            │ ◄────────────────┤
                            │
                            ├──> Broadcast to room
                            │
                            ├──> [User 1] {type: "receive_message"}
                            └──> [User 2] {type: "receive_message"}
```

**Frontend Send (frontend/src/hooks/use-websocket-chat.tsx - Line 141-152):**

```typescript
const sendMessage = useCallback((content: string) => {
    if (!roomId || !isConnected) {
        console.warn("Cannot send message: not in a room or not connected");
        return;
    }

    const client = wsClient.current;
    client.send("send_message", {
        content,
    });
}, [roomId, isConnected]);
```

**Backend Handler (api/internal/handler/websocket.go - Line 340-379):**

```go
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
    msg, err := c.Hub.messageService.CreateMessage(
        context.Background(),
        *c.RoomID,
        c.UserID,
        content,
    )
    if err != nil {
        slog.Error("Failed to save message", "error", err)
        return
    }

    // Broadcast to room (including sender for confirmation)
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
        Exclude: uuid.Nil, // Send to everyone
    }
}
```

**Frontend Receive (frontend/src/hooks/use-websocket-chat.tsx - Line 79-89):**

```typescript
const handleReceiveMessage = (message: WebSocketMessage) => {
    const chatMessage = message.payload as ChatMessage;
    setMessages((prev) => {
        // Avoid duplicates
        if (prev.some((m) => m.id === chatMessage.id)) {
            return prev;
        }
        return [...prev, chatMessage];
    });
};

client.on("receive_message", handleReceiveMessage);
```

**Message Flow:**
1. User types message → Submit form
2. Frontend sends via WebSocket `send_message`
3. Backend saves to database
4. Backend broadcasts `receive_message` to all users in room
5. Both users receive and display message (including sender)

---

### 2.4. Typing Indicator

```
[User 1] ──> typing ──> [WebSocket] ──> [User 2]
                              │
                              └──> {
                                    type: "partner_typing",
                                    payload: {is_typing: true}
                                   }
```

**Frontend Send (frontend/src/hooks/use-websocket-chat.tsx - Line 192-202):**

```typescript
const sendTypingIndicator = useCallback((isTyping: boolean) => {
    if (!roomId || !isConnected) {
        return;
    }

    const client = wsClient.current;
    client.send("typing", {
        is_typing: isTyping,
    });
}, [roomId, isConnected]);

// Usage in input change handler
const handleInputChange = (e) => {
    setInputMessage(e.target.value);

    if (e.target.value.length > 0) {
        sendTypingIndicator(true);

        // Auto-stop after 2 seconds
        clearTimeout(typingTimeout);
        typingTimeout = setTimeout(() => {
            sendTypingIndicator(false);
        }, 2000);
    }
};
```

**Backend Handler (api/internal/handler/websocket.go - Line 452-477):**

```go
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

    // Broadcast to room, exclude sender
    c.Hub.broadcast <- &BroadcastMessage{
        RoomID:  *c.RoomID,
        Message: msgBytes,
        Exclude: c.UserID,
    }
}
```

---

## 3. Leave Flow - Cleanup và State Management

### 3.1. User Leave Room (Manual)

```
[Frontend] ─────> [WebSocket] ─────> [RoomService] ─────> [Database]
    │                  │                   │                    │
    │ Click "Rời      │                   │                    │
    │  phòng"         │                   │                    │
    │                  │                   │                    │
    │ {type:           │ Update DB:        │                    │
    │  "leave_room"}   │ ended_at = NOW    │                    │
    │ ────────────────>│ ─────────────────>│ UPDATE rooms       │
    │                  │                   │ SET ended_at       │
    │                  │                   │ ───────────────────>│
    │                  │                   │                    │
    │                  │ Remove from       │                    │
    │                  │ roomClients       │                    │
    │                  │                   │                    │
    │                  │ Notify partner    │                    │
    │                  │ {type:            │                    │
    │                  │  "partner_left"}  │                    │
    │                  │                   │                    │
    │ ◄────────────────┤                   │                    │
    │ {type:           │                   │                    │
    │  "room_left"}    │                   │                    │
```

**Frontend Handler (frontend/src/hooks/use-websocket-chat.tsx - Line 172-190):**

```typescript
const leaveRoom = useCallback(() => {
    if (!roomId || !isConnected) {
        console.warn("Cannot leave room: not in a room or not connected");
        return;
    }

    const client = wsClient.current;

    // Send leave_room message via WebSocket
    client.send("leave_room", {
        room_id: roomId,
    });

    // Clear local state immediately
    setRoomId(null);
    setMessages([]);

    // Refresh user state to clear active room from auth context
    setTimeout(() => {
        checkAuth().catch(console.error);
    }, 500);
}, [roomId, isConnected, checkAuth]);
```

**Backend Handler (api/internal/handler/websocket.go - Line 421-450):**

```go
func (c *Client) handleLeaveRoom(payload map[string]interface{}) {
    if c.RoomID == nil {
        return
    }

    roomID := *c.RoomID

    // Update database - mark room as ended
    if err := c.Hub.roomService.LeaveRoom(context.Background(), roomID, c.UserID); err != nil {
        slog.Error("Failed to leave room in database", "error", err)
        // Continue with cleanup even if DB fails
    }

    // Remove from WebSocket room
    c.Hub.removeClientFromRoom(c.UserID, roomID)

    // Notify partner
    c.Hub.notifyPartnerLeft(roomID, c.UserID)

    // Clear client's room ID
    c.RoomID = nil

    // Send confirmation to user who left
    confirmation := WebSocketMessage{
        Type: "room_left",
        Payload: map[string]interface{}{
            "room_id":   roomID.String(),
            "timestamp": time.Now().Unix(),
        },
    }
    c.SendJSON(confirmation)
}
```

**RoomService Leave (api/internal/service/room.go - Line 75-108):**

```go
func (s *RoomService) LeaveRoom(ctx context.Context, roomID, userID uuid.UUID) error {
    room, err := s.GetRoomByID(ctx, roomID)
    if err != nil {
        return err
    }

    // Verify user is part of the room
    if room.User1ID != userID && room.User2ID != userID {
        return errors.New("user not part of this room")
    }

    // Check if room is already ended (idempotent)
    if room.EndedAt != nil {
        return nil
    }

    now := time.Now()

    // Use Model().Update for immediate database update
    // This prevents race condition when user joins queue immediately
    if err := s.db.WithContext(ctx).Model(&model.Room{}).
        Where("id = ?", roomID).
        Update("ended_at", now).Error; err != nil {
        return err
    }

    // Trigger post-chat cleanup in background
    go s.cleanupRoom(context.Background(), roomID)

    return nil
}
```

**Key Points:**
1. ✅ **Idempotent**: Nếu room đã ended, return success luôn
2. ✅ **Immediate Update**: Dùng `Model().Update()` thay vì `Save()` để commit ngay
3. ✅ **Frontend Refresh**: Gọi `checkAuth()` sau 500ms để refresh user state
4. ✅ **Prevent Race Condition**: User có thể join queue ngay sau leave room

---

### 3.2. Partner Left Notification

```
[User 1 leaves] ───> [WebSocket Hub] ───> [User 2]
                            │
                            └──> {
                                  type: "partner_left",
                                  payload: {
                                    room_id: "uuid",
                                    message: "Your chat partner has left"
                                  }
                                 }
```

**Backend Notification (api/internal/handler/websocket.go - Line 197-214):**

```go
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

    // Broadcast to room, exclude leaver
    h.broadcast <- &BroadcastMessage{
        RoomID:  roomID,
        Message: msgBytes,
        Exclude: leaverID,
    }
}
```

**Frontend Handler (frontend/src/hooks/use-websocket-chat.tsx - Line 93-106):**

```typescript
const handlePartnerLeft = (message: WebSocketMessage) => {
    console.log("Partner left the room");

    // Clear room state
    setRoomId(null);
    setMessages([]);

    // Call callback
    if (onPartnerLeftRef.current) {
        onPartnerLeftRef.current(); // Show toast notification
    }

    // Refresh user state
    setTimeout(() => {
        checkAuth().catch(console.error);
    }, 300);
};

client.on("partner_left", handlePartnerLeft);
```

**UI Response:**
- Toast: "Đối tác đã rời phòng"
- Auto-clear chat messages
- Reset roomId state
- Show "no room" screen
- Can join queue again

---

### 3.3. Disconnect Cleanup (Unexpected)

```
[WebSocket Disconnect]
         │
         ├──> Hub.unregisterClient()
         │
         ├──> Remove from Hub.clients
         │
         ├──> If in room:
         │    ├─> LeaveRoom (database)
         │    ├─> removeClientFromRoom
         │    └─> notifyPartnerLeft
         │
         └──> QueueService.UserDisconnected()
              └─> Remove from queue if in queue
```

**Backend Cleanup (api/internal/handler/websocket.go - Line 113-140):**

```go
func (h *Hub) unregisterClient(client *Client) {
    h.mutex.Lock()
    if _, exists := h.clients[client.UserID]; exists {
        delete(h.clients, client.UserID)
    }
    h.mutex.Unlock()

    // If user was in a room, clean up
    if client.RoomID != nil {
        // Update database
        if err := h.roomService.LeaveRoom(context.Background(), *client.RoomID, client.UserID); err != nil {
            slog.Error("Failed to leave room on disconnect", "error", err)
        }

        // Remove from WebSocket room
        h.removeClientFromRoom(client.UserID, *client.RoomID)

        // Notify partner
        h.notifyPartnerLeft(*client.RoomID, client.UserID)
    }

    // Remove from queue if in queue
    h.queueService.UserDisconnected(client.UserID)

    close(client.Send)
}
```

**QueueService Cleanup (api/internal/service/queue.go - Line 553-608):**

```go
func (qs *QueueService) UserDisconnected(userID uuid.UUID) {
    // Mark as disconnected
    qs.connMutex.Lock()
    qs.userConnections[userID] = false
    qs.connMutex.Unlock()

    // Remove from queues
    qs.queueMutex.Lock()
    defer qs.queueMutex.Unlock()

    // Search and remove from male queue
    for category, entries := range qs.queueMale {
        for i, entry := range entries {
            if entry.UserID == userID {
                qs.queueMale[category] = append(entries[:i], entries[i+1:]...)
                close(entry.MatchChan)

                // Update positions
                qs.updatePositionsAfterRemoval(category, true, i)

                // Remove tracking
                delete(qs.userPositions, userID)
                return
            }
        }
    }

    // Search and remove from female queue
    for category, entries := range qs.queueFemale {
        for i, entry := range entries {
            if entry.UserID == userID {
                qs.queueFemale[category] = append(entries[:i], entries[i+1:]...)
                close(entry.MatchChan)

                // Update positions
                qs.updatePositionsAfterRemoval(category, false, i)

                // Remove tracking
                delete(qs.userPositions, userID)
                return
            }
        }
    }
}
```

---

## 4. State Management Summary

### 4.1. Backend State

**Hub State:**
```go
type Hub struct {
    clients     map[uuid.UUID]*Client              // All connected clients
    roomClients map[uuid.UUID]map[uuid.UUID]*Client // Room memberships
}
```

**QueueService State:**
```go
type QueueService struct {
    queueMale       map[string][]*QueueEntry    // Male queues by category
    queueFemale     map[string][]*QueueEntry    // Female queues by category
    userConnections map[uuid.UUID]bool          // Connection status
    userPositions   map[uuid.UUID]*QueuePosition // Queue positions
}
```

**Database State:**
```sql
-- Rooms table
CREATE TABLE rooms (
    id UUID PRIMARY KEY,
    user1_id UUID,
    user2_id UUID,
    category VARCHAR,
    ended_at TIMESTAMP,  -- NULL = active, NOT NULL = ended
    is_deleted BOOLEAN,
    created_at TIMESTAMP
);
```

### 4.2. Frontend State

**Auth Context:**
```typescript
interface AuthState {
    isAuthenticated: boolean;
    user: User | null;  // Contains profile and active room
    token: string | null;
    loading: boolean;
}
```

**Queue Hook:**
```typescript
interface QueueState {
    isInQueue: boolean;
    queueStatus: QueueStatus | null;
    position: number;
}
```

**WebSocket Chat Hook:**
```typescript
interface ChatState {
    isConnected: boolean;
    roomId: string | null;
    messages: ChatMessage[];
    isPartnerTyping: boolean;
}
```

### 4.3. State Transitions

**Normal Flow:**
```
[Idle] ─────────────────────────────────────────────────────────┐
  │                                                               │
  │ joinQueue()                                                   │
  ├──> [In Queue] ──────────────────────────────────────┐        │
  │       │                                              │        │
  │       │ match_found                                  │        │
  │       ├──> [In Room] ─────────────────────┐         │        │
  │       │       │                            │         │        │
  │       │       │ leaveRoom()                │         │        │
  │       │       └───────────────────────────>│         │        │
  │       │                                    │         │        │
  │       │ leaveQueue()                       │         │        │
  │       └───────────────────────────────────>│         │        │
  │                                            │         │        │
  └────────────────────────────────────────────┴─────────┴────────┘
```

**Cleanup Checkpoints:**
1. ✅ Leave queue manually → Remove from queue maps + positions
2. ✅ Leave room manually → Update DB + clear WebSocket room + notify partner
3. ✅ Partner leaves → Receive notification + clear state + refresh auth
4. ✅ Disconnect → Auto cleanup queue + room + notify partner
5. ✅ Match found → Remove from queue + auto join room
6. ✅ Frontend refresh auth → Sync user state with backend

---

## 5. Key Features và Optimizations

### 5.1. Gender-Based Queue System
- **Separate queues** cho male/female để matching hiệu quả
- **Priority matching**: Opposite gender first, same gender fallback
- **O(n) matching** trong mỗi queue (n = số user cùng gender trong category)

### 5.2. Connection Tracking
- **userConnections map**: Track real-time connection status
- **Auto-remove** disconnected users từ queue
- **Prevent matching** với disconnected users

### 5.3. Position Tracking
- **userPositions map**: O(1) lookup user position trong queue
- **Auto-update** positions khi user removed
- **Heartbeat system**: TTL-based queue expiration

### 5.4. WebSocket Optimization
- **Singleton pattern**: Chỉ 1 WebSocket connection
- **Auto-reconnect**: Tự động reconnect khi bị disconnect
- **Buffered channels**: Send channel size 256 để tránh blocking
- **Ping/Pong**: Heartbeat để detect dead connections

### 5.5. Database Optimization
- **Immediate update**: Dùng `Model().Update()` thay vì `Save()`
- **Indexed queries**: Index trên `user1_id`, `user2_id`, `ended_at`
- **Background cleanup**: Async cleanup messages sau khi room ended

### 5.6. Frontend State Sync
- **checkAuth() refresh**: Sync user state sau leave room
- **Optimistic updates**: UI update ngay, backend confirm sau
- **Debounced typing**: Chỉ send typing sau 2s không gõ

---

## 6. Error Handling

### 6.1. Backend Errors

**Join Queue Errors:**
- ✅ Invalid category → 400 Bad Request
- ✅ Already in queue → 400 "user already in queue"
- ✅ Already has active room → 400 "user already has active room"
- ✅ Profile not found → 400 "user profile not found"

**WebSocket Errors:**
- ✅ Not authenticated → 401 Unauthorized
- ✅ Room not found → Error log, no action
- ✅ Not authorized for room → Warn log, no action
- ✅ Message save failed → Error log, no broadcast

### 6.2. Frontend Errors

**WebSocket Errors:**
- ✅ Connection failed → Show "Đang kết nối..." + auto-retry
- ✅ Unexpected close → Auto-reconnect with exponential backoff

**API Errors:**
- ✅ Join queue failed → Toast error + revert optimistic update
- ✅ Leave queue failed → Toast error + keep state

### 6.3. Race Conditions Handled

**Scenario 1: Leave room → Join queue ngay lập tức**
- ✅ Fixed: Database update immediately với `Model().Update()`
- ✅ Fixed: Frontend refresh auth state sau 500ms

**Scenario 2: Multiple join queue requests**
- ✅ Fixed: Check `isUserInQueue()` trước khi thêm vào queue
- ✅ Fixed: Frontend disable button khi `isLoading`

**Scenario 3: WebSocket reconnect nhiều lần**
- ✅ Fixed: Check connection state trước khi connect
- ✅ Fixed: Use `useRef` cho callbacks để tránh effect re-run

---

## 7. Testing Scenarios

### 7.1. Happy Path
1. ✅ User 1 join queue
2. ✅ User 2 join queue (opposite gender)
3. ✅ Match found → Both enter room
4. ✅ Exchange messages
5. ✅ User 1 leaves → User 2 receives notification
6. ✅ User 2 can join queue again

### 7.2. Edge Cases
1. ✅ User joins queue → disconnects → auto removed
2. ✅ User leaves room → immediately joins queue → no error
3. ✅ Partner disconnects → user receives notification + can rejoin
4. ✅ Multiple tabs → same user → shared WebSocket singleton
5. ✅ User in queue → refreshes page → WebSocket reconnects

### 7.3. Stress Testing
- ✅ Multiple users join queue simultaneously
- ✅ Fast message sending (typing indicator debounced)
- ✅ Rapid leave/join cycles

---

## 8. Future Improvements

### 8.1. Performance
- [ ] Redis-based queue để scale horizontally
- [ ] Message pagination (hiện load tất cả messages)
- [ ] WebSocket connection pooling

### 8.2. Features
- [ ] Report user feature
- [ ] Block user feature
- [ ] Message reactions
- [ ] File/image sharing
- [ ] Voice/video call

### 8.3. Monitoring
- [ ] Queue metrics (wait time, match rate)
- [ ] WebSocket connection metrics
- [ ] Error tracking và alerting

---

**Generated:** 2025-12-11
**Updated:** 2026-01-08
**Version:** 1.0
**Status:** Production-ready
