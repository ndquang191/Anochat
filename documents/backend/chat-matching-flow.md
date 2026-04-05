# Chat Matching Flow

## Overview

Three phases:
1. **Queue** — user joins Redis queue, gets atomically matched
2. **Chat** — real-time messaging via WebSocket + Redis pub/sub
3. **Leave** — room ends, messages are cleaned up

---

## 1. Queue Flow

### Join Queue

```
Frontend          Backend API          QueueService (Redis)
   |                   |                       |
   | POST /queue/join  |                       |
   |──────────────────►|                       |
   |                   | 1. Check active room  |
   |                   | 2. Run Lua script ───►|
   |                   |    ZSCORE (already?)  |
   |                   |    ZRANGE (find match)|
   |                   |    ZADD or ZREM       |
   |◄──────────────────|                       |
   | 200 OK            |                       |
```

The Lua script is atomic — no two instances can race to match the same user.

**Outcomes:**
- `"waiting"` — added to `queue:waiting` (score = join timestamp ms)
- `"matched:<partnerID>:<score>"` — partner removed, room created, `match_found` published to both user channels

### Match Notification (via Redis pub/sub)

```
QueueService           Redis              Hub (instance A)    Hub (instance B)
     |                   |                      |                    |
     | Publish           |                      |                    |
     | user:<userA>  ───►| ──subscribe──────────►| deliver to userA  |
     | user:<userB>  ───►| ──subscribe────────────────────────────►  | deliver to userB
     |                   |                      |                    |
     |                   |              AddClientToRoom         AddClientToRoom
```

Each Hub instance is subscribed to `user:{userID}` channels for all its locally connected users.

### Leave Queue

```
POST /queue/leave → ZREM queue:waiting {userID}
```

---

## 2. Chat Flow

### WebSocket Connection

On connect, the Hub:
1. Registers client in memory
2. Subscribes to `user:{userID}` Redis channel
3. Checks DB for active room → auto-rejoins if found

### Send Message

```
Client A             Hub A             Redis (room:X)         Hub B             Client B
  |                    |                     |                   |                  |
  | send_message       |                     |                   |                  |
  |───────────────────►|                     |                   |                  |
  |                    | Publish room:X ─────►────────────────►  |                  |
  |                    |                     |   subscribe room:X|                  |
  |                    |                     |                   | deliver locally  |
  |                    |                     |                   |─────────────────►|
  |                    | Save to DB (async)  |                   |                  |
```

Message is published to Redis `room:{roomID}`. Every Hub instance subscribed to that channel (because it has a local client in that room) delivers it locally.

### Typing Indicator

Same path as `send_message` but goes through `room:{roomID}` pub/sub with `partner_typing` payload.

### Leave Room

```
Client A sends leave_room
   → Hub calls RoomService.LeaveRoom (sets ended_at in DB)
   → Hub removes client from roomClients
   → Hub publishes partner_left to room:{roomID}
   → Hub B delivers partner_left to Client B
   → Room cleanup goroutine deletes messages after a delay
```

---

## 3. Cleanup

When a room ends, `RoomService` starts a goroutine that:
1. Deletes all messages for the room
2. Deletes the room record

This happens asynchronously to avoid blocking the leave operation.

---

## Redis Keys Used

| Key | Type | Purpose |
|-----|------|---------|
| `queue:waiting` | Sorted Set | Matchmaking queue (score = join time ms) |
| `room:{roomID}` | Pub/Sub | Broadcast messages to room members |
| `user:{userID}` | Pub/Sub | Direct notifications (match_found) |
| `msgrl:{userID}` | String (TTL 1s) | Per-user message rate limit counter |
| `rl:{ip}` | String (TTL 1s) | Per-IP HTTP rate limit counter |
| `refresh:{hash}` | String (TTL) | Refresh token → userID mapping |
