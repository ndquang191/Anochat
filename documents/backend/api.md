# API Reference

## Base URL

```
http://localhost:8080
```

---

## Rate Limiting

### HTTP (global)
- **100 req/s** per IP, burst 200 — token bucket via Redis
- Exceeded: `429 Too Many Requests`

### WebSocket messages (per user)
- **10 messages/s** per user — enforced via Redis `INCR`/`EXPIRE`
- Exceeded: WS error event `RATE_LIMIT_EXCEEDED`

---

## Authentication

All protected routes require a valid JWT in the HTTP-only `access_token` cookie.
Include `credentials: "include"` on every frontend fetch.

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/auth/google` | Redirect to Google OAuth |
| GET | `/auth/callback` | OAuth callback — sets cookies, redirects to frontend `/callback` |
| POST | `/auth/refresh` | Exchange `refresh_token` cookie for new access token |
| POST | `/auth/logout` | Revoke refresh token, clear cookies |

---

## Protected Endpoints

### GET `/user/state`

Returns the current user's full state, including any active room and messages.

**Response:**
```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "Display Name",
    "avatar_url": "https://...",
    "profile": {
      "nickname": "anon123",
      "age": 25,
      "is_male": true,
      "is_hidden": false
    }
  },
  "room": null,
  "messages": null,
  "is_new_user": false
}
```

---

### PUT `/profile`

Update profile fields. All fields are optional.

**Request body:**
```json
{
  "nickname": "anon123",
  "age": 25,
  "is_male": true,
  "is_hidden": false
}
```

---

### POST `/queue/join`

Join the matchmaking queue. Match notification is delivered via WebSocket `match_found`.

**Request body:** _(empty)_

**Errors:** `409` if already in queue or already in an active room.

---

### POST `/queue/leave`

Leave the queue without being matched.

---

### POST `/room/leave`

End the current active room and clean up messages.

---

### POST `/report`

Report another user.

**Request body:**
```json
{ "reported_user_id": "uuid" }
```

---

### GET `/ws`

Upgrade to WebSocket. Requires valid `access_token` cookie.

#### Client → Server events

| Type | Payload | Description |
|------|---------|-------------|
| `send_message` | `{ "content": "..." }` | Send a message |
| `join_room` | `{ "room_id": "uuid" }` | Re-join room after reconnect |
| `leave_room` | `{ "room_id": "uuid" }` | Leave current room |
| `typing` | `{ "is_typing": true\|false }` | Typing indicator |

#### Server → Client events

| Type | Payload | Description |
|------|---------|-------------|
| `connected` | `{ user_id, message, timestamp }` | Connection confirmed |
| `room_rejoined` | `{ room_id, timestamp }` | Auto-rejoin on reconnect |
| `match_found` | `{ room_id, timestamp, message }` | Matched with a partner |
| `room_joined` | `{ room_id, timestamp }` | `join_room` confirmed |
| `receive_message` | `{ id, room_id, sender_id, content, created_at }` | Incoming message |
| `partner_typing` | `{ is_typing, user_id }` | Partner typing status |
| `partner_left` | `{ room_id, timestamp, message }` | Partner left |
| `room_left` | `{ room_id, timestamp }` | `leave_room` confirmed |
| `error` | `{ message, code }` | e.g. `RATE_LIMIT_EXCEEDED` |

---

## Admin Endpoints

All require auth (no separate admin role enforced yet).

| Method | Path | Description |
|--------|------|-------------|
| GET | `/admin/words` | List banned words |
| POST | `/admin/words` | Add banned word |
| PUT | `/admin/words/:id` | Update banned word |
| DELETE | `/admin/words/:id` | Delete banned word |
| GET | `/admin/reports` | List reports |
| POST | `/admin/users/:id/ban` | Ban a user |
| POST | `/admin/users/:id/unban` | Unban a user |
| GET | `/admin/users/banned` | List banned users |
| GET | `/admin/rooms/:id/messages` | List messages in a room |

---

## Utility

| Method | Path | Description |
|--------|------|-------------|
| GET | `/healthz` | Health check (DB + Redis status) |
| GET | `/metrics` | Prometheus metrics |

---

## Error Responses

```json
{ "error": "human-readable Vietnamese message", "code": "OPTIONAL_CODE" }
```

| Status | Meaning |
|--------|---------|
| 400 | Bad request / validation error |
| 401 | Missing, invalid, or expired token |
| 409 | Conflict (already in queue, has active room, etc.) |
| 429 | Rate limit exceeded |
| 500 | Server error |
