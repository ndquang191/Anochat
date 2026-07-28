# Chat Matching Flow

## Overview

Three phases:
1. **Queue** — user joins Redis queue, gets atomically reserved and matched
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
   |                   |    clear expired lease|
   |                   |    ZSCORE (already?)  |
   |                   |    ZRANGE (find match)|
   |                   |    reserve or ZADD    |
   |◄──────────────────|                       |
   | 200 OK            |                       |
```

The first Lua script atomically reserves both users with a token and a 30-second
lease. The waiting partner remains in `queue:waiting`, but other match attempts
skip reserved users. This closes the gap between selecting a Redis candidate and
creating the room in PostgreSQL.

**Outcomes:**
- `"waiting"` — added to `queue:waiting` (score = join timestamp ms)
- `"match_in_progress"` — one of the user's requests already owns a live reservation
- `"reserved:<partnerID>:<score>"` — both users reserved while the room is created

After reservation:

1. `CreateRoom(partner, user)` writes the room to PostgreSQL.
2. On success, a second Lua script verifies the ownership token, removes the
   partner from `queue:waiting`, and clears both reservation entries.
3. On failure, the second script only releases the reservation. The waiting
   partner was never removed and can be matched again.
4. If the process stops before finalization, the lease expires and the next
   queue operation atomically clears it.

The ownership token prevents a delayed commit or release from modifying a newer
reservation created after the old lease expired.

The PostgreSQL room trigger also inserts both users into
`active_room_members`. The `user_id` primary key prevents duplicate active
rooms even for direct SQL writes. Redis finalization is retried immediately; a
reconciliation worker also periodically removes queue/reservation state for all
active PostgreSQL rooms.

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
POST /queue/leave → Lua:
  active reservation? → 409 MatchInProgress (keep queue state)
  otherwise           → ZREM queue:waiting {userID}
```

WebSocket disconnect uses the same Lua script, so it cannot remove a partner
while the room transaction is in progress.

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
  | send_message{id}   |                     |                   |                  |
  |───────────────────►|                     |                   |                  |
  |                    | Save to DB          |                   |                  |
  |                    | Publish room:X ─────►────────────────►  |                  |
  |                    |                     |   subscribe room:X|                  |
  |                    |                     |                   | deliver locally  |
  |                    |                     |                   |─────────────────►|
  |◄── message_ack{id}─|                     |                   |                  |
```

The client-generated message ID is persisted before publication. After the
message is saved, it is published to Redis `room:{roomID}` for the partner and
the sender receives `message_ack`. Validation or persistence failures return
`message_failed` with the same ID.

### Typing Indicator

Same path as `send_message` but goes through `room:{roomID}` pub/sub with `partner_typing` payload.

### Leave Room

```
Client A sends leave_room
   → Hub calls RoomService.LeaveRoom (sets ended_at in DB)
   → failure: send room_leave_failed and retain client room state
   → Hub removes client from roomClients
   → Hub publishes partner_left to room:{roomID}
   → Hub sends room_left ACK to Client A
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
| `queue:reservations` | Sorted Set | Reserved users (score = lease expiry time ms) |
| `queue:reservation_tokens` | Hash | User ID → reservation ownership token |
| `room:{roomID}` | Pub/Sub | Broadcast messages to room members |
| `user:{userID}` | Pub/Sub | Direct notifications (match_found) |
| `msgrl:{userID}` | String (TTL 1s) | Per-user message rate limit counter |
| `rl:{ip}` | String (TTL 1s) | Per-IP HTTP rate limit counter |
| `refresh:{hash}` | String (TTL) | Refresh token → userID mapping |

The Hub and reconciliation worker share the server lifecycle context. During
shutdown, HTTP stops accepting requests, WebSocket clients receive a close
frame, Redis pub/sub is closed, and background loops are awaited.
