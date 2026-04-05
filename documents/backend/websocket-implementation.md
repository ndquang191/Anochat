# WebSocket Implementation

## Overview

Each backend instance runs a single `Hub` goroutine that manages all WebSocket connections on that instance. Instances communicate via Redis pub/sub — messages and match notifications are broadcast through Redis so any instance can deliver to any user.

---

## Hub Architecture

```
                    Redis pub/sub
                    ┌───────────┐
                    │  room:X   │
                    │  user:A   │
                    │  user:B   │
                    └─────┬─────┘
          ┌───────────────┼───────────────┐
          ▼               │               ▼
   Hub (instance 1)       │       Hub (instance 2)
   ┌─────────────┐        │       ┌─────────────┐
   │ clients map │        │       │ clients map │
   │ roomClients │        │       │ roomClients │
   │ Run() loop  │        │       │ Run() loop  │
   └──────┬──────┘        │       └──────┬──────┘
          │               │              │
    User A (WS)                    User B (WS)
```

### Hub channels

| Channel | Type | Purpose |
|---------|------|---------|
| `register` | `chan *Client` | New WS connection |
| `unregister` | `chan *Client` | Disconnected client |
| `broadcast` | `chan *BroadcastMessage` | Publish to a room |
| `deliver` | `chan *redis.Message` | Incoming Redis pub/sub message |

---

## Redis Pub/Sub Channels

### `room:{roomID}`

Published when a message is broadcast to a room. Payload:
```json
{ "exclude": "<senderUserID>", "payload": <raw WS message bytes> }
```

Each Hub subscribes to a room's channel when the **first local client** joins and unsubscribes when the **last local client** leaves.

### `user:{userID}`

Published for direct user notifications. Current message types:

```json
{ "type": "match_found", "room_id": "<uuid>" }
```

Each Hub subscribes to `user:{userID}` when that user connects and unsubscribes on disconnect.

---

## Client Lifecycle

```
WS connect
  → Hub.register channel
    → registerClient()
      → subscribe user:{userID}
      → check DB for active room → AddClientToRoom() if found

WS disconnect
  → Hub.unregister channel
    → unregisterClient()
      → unsubscribe user:{userID}
      → removeClientFromRoom() → unsubscribe room:{roomID} if last local client
      → QueueService.UserDisconnected()
```

---

## Message Flow (send_message)

```go
// 1. Client sends WS frame
client.handleSendMessage(payload)

// 2. Rate limit check via Redis
hub.CheckMessageRateLimit(userID)  // INCR msgrl:{userID}, EXPIRE 1s

// 3. Moderation check
hub.moderationService.ContainsBannedWord(content)  // → auto-report if hit

// 4. Publish to Redis
hub.broadcast <- &BroadcastMessage{RoomID, Message, Exclude}
  → hub.publishToRedis()
    → rdb.Publish("room:{roomID}", {exclude, payload})

// 5. Redis fan-out to all subscribed Hub instances
//    Each Hub calls handleRoomMessage() → delivers to local clients in that room

// 6. Save to DB (async goroutine)
messageService.SaveMessage(ctx, msgID, roomID, senderID, content, now)
```

---

## Auto-Rejoin on Reconnect

When a client reconnects, `registerClient` queries the DB for an active room.
If found, `AddClientToRoom` is called automatically and a `room_rejoined` WS event is sent.

---

## Files

| File | Purpose |
|------|---------|
| `api/internal/ws/hub.go` | Hub struct, Run loop, Redis pub/sub management |
| `api/internal/ws/client.go` | Client struct, ReadPump, WritePump |
| `api/internal/ws/handler.go` | WS message routing (send_message, join_room, leave_room, typing) |
| `api/internal/ws/notifier.go` | notifyPartnerLeft, NotifyMatch (publishes to Redis user channels) |
| `api/internal/handler/ws_handler.go` | HTTP upgrade handler, auth check |

---

## Security

- WS route protected by `AuthMiddleware` — requires valid `access_token` cookie
- `join_room` verifies the user is a member of the room before adding to `roomClients`
- Rate limiting per user via Redis prevents message spam

---

## Scaling

One Redis is sufficient. Add backend instances freely — all share the same Redis pub/sub broker.
Redis Cluster is only needed if Redis itself becomes a throughput bottleneck.
