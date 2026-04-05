# Anochat

Anonymous chat application with real-time matching and messaging.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              FRONTEND (Next.js 15)                          │
│                                                                             │
│  ┌──────────┐  ┌──────────────┐  ┌────────────┐  ┌──────────────────────┐  │
│  │  Pages    │  │  Contexts    │  │   Hooks    │  │   Libraries          │  │
│  │ /login    │  │ AuthProvider │  │ useQueue   │  │ api.ts (REST)        │  │
│  │ /callback │  │ AlertDialog  │  │ useWsChat  │  │ websocket.ts (WS)    │  │
│  │ /  (main) │  │ AppProvider  │  │ useUserSt. │  │ query-client.ts (RQ) │  │
│  └──────────┘  └──────────────┘  └────────────┘  └──────────────────────┘  │
└────────────────────────┬──────────────────┬─────────────────────────────────┘
                         │ REST (fetch)     │ WebSocket
                         ▼                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                            BACKEND (Go / Gin)                               │
│                                                                             │
│  ┌─────────────────────────── Middleware ──────────────────────────────┐    │
│  │  CORS  →  Rate Limit (Redis)  →  Auth (JWT cookie)                 │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌────────────┐  ┌───────────────┐      │
│  │ AuthHandler  │  │ UserHandler │  │QueueHandler│  │  WS Handler   │      │
│  └──────┬──────┘  └──────┬──────┘  └─────┬──────┘  └───────┬───────┘      │
│         │                │               │                  │               │
│  ┌──────▼──────┐  ┌──────▼──────┐  ┌─────▼──────┐  ┌───────▼───────┐      │
│  │ AuthService │  │ UserService │  │QueueService│  │  WebSocket    │      │
│  │ (OAuth+JWT) │  │             │  │  (Redis)   │◄─┤  Hub          │      │
│  └──────┬──────┘  └──────┬──────┘  └─────┬──────┘  │  (goroutine)  │      │
│         │                │               │          └───────┬───────┘      │
│  ┌──────▼──────────────▼───────────────▼──────────────────▼────────┐       │
│  │              Repositories (GORM)                                 │       │
│  │  UserRepo  │  ProfileRepo  │  RoomRepo  │  MessageRepo          │       │
│  └──────────────────────────┬────────────────────────────────────--┘       │
└─────────────────────────────┼───────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼                               ▼
     ┌────────────────┐              ┌────────────────────────────┐
     │  PostgreSQL     │              │  Redis                     │
     │  Users          │              │  rate limit: rl:{ip}       │
     │  Profiles       │              │  msg rate:  msgrl:{userID} │
     │  Rooms          │              │  refresh:   refresh:{hash} │
     │  Messages       │              │  queue:     queue:waiting  │
     └────────────────┘              │  pub/sub:   room:{roomID}  │
                                     │             user:{userID}  │
                                     └────────────────────────────┘
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend | Next.js 15, React 19, React Query, TypeScript |
| Backend | Go, Gin, GORM, Gorilla WebSocket |
| Database | PostgreSQL |
| Cache / Broker | Redis (rate limiting, auth tokens, queue, pub/sub) |
| Auth | Google OAuth 2.0, JWT (HTTP-only `access_token` cookie) |
| Logging | Zap (via slog interface) |

## Data Flow

```
QueryClientProvider
  └── ErrorBoundary
       └── AuthProvider ← useUserState() ← GET /user/state (cached)
            └── AlertDialogProvider
                 ├── Sidebar     (reads same cached query)
                 ├── Header      (invalidates on actions)
                 └── Page        (useQueue + useWebSocketChat)
```

**React Query cache keys:**

| Key | Behavior |
|-----|----------|
| `["user-state"]` | staleTime 30s, shared by AuthProvider + Sidebar |
| `["queue-status"]` | refetchInterval 5s while in queue |

## Queue Matching

The queue is stored in a Redis sorted set (`queue:waiting`, score = join timestamp ms).
Matching uses a Lua script for atomicity — safe across multiple backend instances.

```
User A → POST /queue/join
           └── Lua script: ZRANGE queue:waiting → no match → ZADD

User B → POST /queue/join
           └── Lua script: ZRANGE queue:waiting → finds User A → ZREM User A
                              └── CreateRoom(A, B)
                                   └── Publish "match_found" → Redis user:A, user:B
                                        ├── Hub (instance 1) → WS to User A
                                        └── Hub (instance 2) → WS to User B
```

## WebSocket Hub (Distributed)

Each backend instance runs a Hub goroutine. Instances communicate via Redis pub/sub —
a single Redis node handles this at any reasonable scale.

```
Instance 1                  Redis               Instance 2
  User A in room X  ──publish room:X──►  ──subscribe room:X──►  User B in room X
                    ◄─subscribe room:X──  ◄─publish room:X──────
```

**Channel naming:**
- `room:{roomID}` — broadcast to all clients in a room (excludes sender)
- `user:{userID}` — direct notification to a specific user (match_found, etc.)

Subscriptions are managed per-instance: a room channel is subscribed when the first
local client joins and unsubscribed when the last local client leaves.

## Database Schema

```
Users ──┐     Profiles (1:1)
        ├──── Rooms (user1_id, user2_id)
        └──── Messages (sender_id) ──── Rooms (room_id)
```

## Getting Started

### Prerequisites

- Go 1.24+
- Node.js 20+
- PostgreSQL
- Redis

### Backend

```bash
cd api
cp .env.example .env    # edit with your credentials
go run ./cmd/server
```

### Frontend

```bash
cd frontend
cp .env.example .env.local
npm install
npm run dev
```

### Environment Variables

See `api/.env.example` and `frontend/.env.example` for required configuration.

## Scaling

The backend is stateless with respect to WebSocket sessions — all shared state
(queue, room membership, rate limits, auth tokens) lives in Redis. You can run
multiple backend instances behind a load balancer with a single Redis node.
Redis Cluster or Sentinel is only needed if Redis itself becomes a bottleneck.
