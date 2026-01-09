# API Documentation

## Base URL

```
http://localhost:8080
```

## Rate Limiting

### Global Rate Limit
-   **100 requests per second** per IP address
-   **Burst capacity:** 200 requests
-   Algorithm: Token bucket with automatic refill
-   **Response on limit exceeded:** 429 Too Many Requests

### Message Rate Limit (WebSocket)
-   **10 messages per second** per user
-   Applied to WebSocket `send_message` events
-   **Response on limit exceeded:** Error message via WebSocket
-   Prevents message spam and abuse

### Rate Limit Headers
Rate limiting is transparent and automatic. Clients don't need to track limits manually.

## Authentication

### Google OAuth Flow

1. **GET** `/auth/google` - Redirect to Google OAuth
2. **GET** `/auth/callback` - OAuth callback (handled by backend)
3. **POST** `/auth/logout` - Logout and clear cookies

### Authentication Method

-   Uses HTTP-only cookies for JWT token storage
-   Frontend includes `credentials: "include"` in requests
-   No need for Authorization header in most cases

## Protected Endpoints

### User State 

**GET** `/user/state`

Returns current user state including profile information.

**Response:**

```json
{
	"user": {
		"id": "uuid",
		"email": "user@example.com",
		"name": "User Name",
		"avatar_url": "https://...",
		"profile": {
			"age": 25,
			"city": "Ho Chi Minh City",
			"is_male": true,
			"is_hidden": false
		}
	},
	"room": null,
	"messages": null,
	"is_new_user": false
}
```

### Update Profile

**PUT** `/profile`

Update user profile information.

**Request Body:**

```json
{
	"age": 25, // optional, number
	"city": "Hanoi", // optional, string
	"is_male": true, // optional, boolean
	"is_hidden": false // optional, boolean
}
```

**Response:**

```json
{
	"message": "Profile updated successfully",
	"profile": {
		"age": 25,
		"city": "Hanoi",
		"is_male": true,
		"is_hidden": false
	}
}
```

### Queue Management

**POST** `/queue/join`

Join the matchmaking queue for a specific category.

**Request Body:**

```json
{
	"category": "polite" // required, string
}
```

**Response:**

```json
{
	"is_in_queue": true,
	"position": 1,
	"category": "polite"
}
```

---

**POST** `/queue/leave`

Leave the matchmaking queue.

**Response:**

```json
{
	"message": "Successfully left queue"
}
```

---

**GET** `/queue/status`

Get current queue status for the user.

**Response:**

```json
{
	"is_in_queue": true,
	"position": 1,
	"category": "polite",
	"estimated_wait_time": 10
}
```

---

**GET** `/queue/stats`

Get queue statistics (admin/debugging endpoint).

**Response:**

```json
{
	"categories": {
		"polite": {
			"male": 5,
			"female": 3,
			"total": 8
		}
	},
	"total_users": 8
}
```

---

**GET** `/queue/match-stats`

Get matching statistics.

**Response:**

```json
{
	"total_matches": 150,
	"matches_today": 25,
	"average_wait_time": 8.5
}
```

### WebSocket Endpoints

**GET** `/ws`

Establish WebSocket connection for real-time chat.

**Authentication:** Requires JWT token in cookies

**Events (Client → Server):**

| Event | Payload | Description |
|-------|---------|-------------|
| `send_message` | `{content: string}` | Send a chat message |
| `join_room` | `{room_id: string}` | Join a chat room |
| `leave_room` | `{room_id: string}` | Leave current room |
| `typing` | `{is_typing: boolean}` | Send typing indicator |

**Events (Server → Client):**

| Event | Payload | Description |
|-------|---------|-------------|
| `connected` | `{user_id, message, timestamp}` | Connection established |
| `match_found` | `{room_id, category, timestamp}` | Matched with another user |
| `room_joined` | `{room_id, timestamp}` | Successfully joined room |
| `receive_message` | `{id, room_id, sender_id, content, created_at}` | New message received |
| `partner_left` | `{room_id, timestamp, message}` | Partner left the room |
| `partner_typing` | `{is_typing, user_id}` | Partner typing status |
| `room_left` | `{room_id, timestamp}` | Left room confirmation |
| `error` | `{message, code}` | Error message (e.g., rate limit) |

## Health Check

**GET** `/healthz`

Returns server health status.

**Response:**

```json
{
	"status": "ok",
	"message": "Anonymous Chat API is running",
	"database": "connected"
}
```

## Error Responses

### Unauthorized (401)

```json
{
	"error": "Authorization required"
}
```

### Bad Request (400)

```json
{
	"error": "Invalid request body"
}
```

### Internal Server Error (500)

```json
{
	"error": "Error message"
}
```

### Rate Limit Exceeded (429)

```json
{
	"error": "Rate limit exceeded",
	"message": "Too many requests. Please try again later."
}
```

**For WebSocket message rate limiting:**

```json
{
	"type": "error",
	"payload": {
		"message": "You are sending messages too quickly. Please slow down.",
		"code": "RATE_LIMIT_EXCEEDED"
	}
}
```

## Best Practices

### Rate Limiting
-   Implement exponential backoff on 429 responses
-   Don't retry immediately after rate limit
-   Recommended retry delay: 1 second for first retry, then exponential

### WebSocket Connection
-   Maintain single WebSocket connection per user
-   Implement automatic reconnection with exponential backoff
-   Handle `error` events gracefully

### Authentication
-   Always include `credentials: "include"` in fetch requests
-   Handle 401 errors by redirecting to login
-   JWT token expires after 7 days

### Message Sending
-   Validate message content before sending
-   Keep messages under 1000 characters
-   Handle rate limit errors gracefully with user feedback
