# Monitoring & Logging Guide

## Logging

**Library:** `log/slog` (structured), backed by Zap
**Format:** JSON in production, console in development
**Default level:** Info (production), Debug (development)

| Level | When to use |
|-------|-------------|
| Info | Normal operations: login, match found, room created |
| Warn | Recoverable issues: rate limit hit, Redis error (fail-open) |
| Error | Requires investigation: DB error, WS failure |
| Debug | Verbose queue and reconnect state |

### What to log

```go
// Auth
slog.Info("User logged in", "user_id", userID)
slog.Warn("Authentication failed", "reason", "invalid_token")

// Queue & matching
slog.Info("User joined queue", "user_id", userID)
slog.Info("Match found", "room_id", room.ID,
    "user1_id", partnerID, "user2_id", userID, "wait_seconds", waitSeconds)
slog.Info("User left queue", "user_id", userID)

// WebSocket
slog.Info("Client registered", "user_id", userID, "client_id", clientID)
slog.Info("Client unregistered", "user_id", userID, "was_latest", isLatest)
slog.Info("Subscribed to room channel", "room_id", roomID)
slog.Info("Unsubscribed from room channel", "room_id", roomID)

// Rate limiting
slog.Warn("Message rate limit exceeded", "user_id", userID)
```

### What NOT to log

- JWT tokens, OAuth tokens, refresh tokens
- Message content
- Full user profiles
- Email addresses in high-volume logs

---

## Prometheus Metrics

Exposed at `GET /metrics`.

| Metric | Type | Description |
|--------|------|-------------|
| `anochat_active_ws_connections` | Gauge | Live WebSocket connections |
| `anochat_users_total` | Gauge | Total registered users |
| `anochat_new_registrations_total` | Counter | New registrations since start |
| `anochat_queue_size` | Gauge | Current `queue:waiting` size (from Redis `ZCARD`) |
| `anochat_match_duration_seconds` | Histogram | Time from joining queue to being matched |

---

## Health Check

`GET /healthz` checks PostgreSQL and Redis connectivity:

```json
{
  "status": "ok",
  "database": "connected",
  "redis": "connected"
}
```

Returns `503` if either dependency is down.
