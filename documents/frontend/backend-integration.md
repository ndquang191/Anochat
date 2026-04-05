# Frontend ↔ Backend Integration

## Authentication

The backend uses HTTP-only cookies — no Authorization headers needed.
Include `credentials: "include"` on every fetch.

### Flow

1. User clicks Login → frontend redirects to `/auth/google`
2. Google OAuth handled by backend
3. Backend sets `access_token` + `refresh_token` cookies, redirects to `/callback`
4. Frontend reads `temp_user_data` cookie (60 s TTL) to get user info
5. Frontend calls `GET /user/state` for full user state

### Token refresh

When a request returns `401 { "code": "token_expired" }`:
- Call `POST /auth/refresh` (uses the `refresh_token` cookie automatically)
- Backend sets a new `access_token` cookie
- Retry the original request

---

## API Calls

```typescript
// All calls need credentials: "include"
const response = await fetch(`${API_URL}/user/state`, {
  credentials: "include",
});
```

### Get user state
```typescript
GET /user/state
// returns { user, room, messages, is_new_user }
```

### Update profile
```typescript
PUT /profile
body: {
  nickname?: string,
  age?: number,
  is_male?: boolean,
  is_hidden?: boolean,
}
```

### Queue
```typescript
POST /queue/join   // join queue, match comes via WebSocket
POST /queue/leave  // leave queue
```

### Leave room
```typescript
POST /room/leave
```

### Logout
```typescript
POST /auth/logout
```

---

## WebSocket

Connect to `GET /ws` — the `access_token` cookie is checked on upgrade.

### Key events to handle (server → client)

| Event | When |
|-------|------|
| `connected` | WS connection established |
| `match_found` | Matched with a partner, contains `room_id` |
| `room_rejoined` | Auto-rejoin after reconnect |
| `receive_message` | New message from partner |
| `partner_typing` | Partner started/stopped typing |
| `partner_left` | Partner left the room |
| `error` | e.g. `RATE_LIMIT_EXCEEDED` |

### Events to send (client → server)

| Event | When |
|-------|------|
| `send_message` | User sends a message |
| `typing` | User starts/stops typing |
| `leave_room` | User leaves the room |
| `join_room` | Explicit room join (usually not needed — auto-rejoin handles it) |

---

## Error Handling

| HTTP Status | Action |
|-------------|--------|
| 401 `no_token` | Redirect to `/login` |
| 401 `token_expired` | Refresh token, retry |
| 401 `account_suspended` | Show ban message, redirect to `/login` |
| 429 | Exponential backoff, show message to user |
| 5xx | Show generic error, log to console |

---

## CORS

Backend allows the configured `CLIENT_URL` with `credentials: true`.
`SameSite=None; Secure` cookies are used in production. `SameSite=Lax` in development.
