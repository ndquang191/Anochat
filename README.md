# Anochat

Anonymous chat application with real-time matching and messaging.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              FRONTEND (Next.js 15)                         │
│                                                                             │
│  ┌──────────┐  ┌──────────────┐  ┌────────────┐  ┌──────────────────────┐  │
│  │  Pages    │  │  Contexts    │  │   Hooks    │  │   Libraries          │  │
│  │ /login    │  │ AuthProvider │  │ useQueue   │  │ api.ts (REST)        │  │
│  │ /callback │  │ AlertDialog  │  │ useWsChat  │  │ websocket.ts (WS)   │  │
│  │ /  (main) │  │ AppProvider  │  │ useUserSt. │  │ query-client.ts (RQ)│  │
│  └──────────┘  └──────────────┘  │ useQueueQ. │  │ cookies.ts          │  │
│                                   └────────────┘  └──────────────────────┘  │
└────────────────────────┬──────────────────┬─────────────────────────────────┘
                         │ REST (fetch)     │ WebSocket
                         ▼                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                            BACKEND (Go / Gin)                               │
│                                                                             │
│  ┌─────────────────────────── Middleware ──────────────────────────────┐    │
│  │  CORS  →  Rate Limit (Redis)  →  Auth (JWT)                       │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌────────────┐  ┌───────────────┐      │
│  │ AuthHandler  │  │ UserHandler │  │QueueHandler│  │  WS Handler   │      │
│  └──────┬──────┘  └──────┬──────┘  └─────┬──────┘  └───────┬───────┘      │
│         │                │               │                  │               │
│  ┌──────▼──────┐  ┌──────▼──────┐  ┌─────▼──────┐  ┌───────▼───────┐      │
│  │ AuthService │  │ UserService │  │QueueService│  │  WebSocket    │      │
│  │ (OAuth+JWT) │  │             │  │(in-memory) │◄─┤  Hub          │      │
│  └──────┬──────┘  └──────┬──────┘  └─────┬──────┘  │  (goroutine)  │      │
│         │                │               │          └───────────────┘      │
│  ┌──────▼──────────────▼───────────────▼────────────────────────┐         │
│  │              Repositories (GORM)                              │         │
│  │  UserRepo  │  ProfileRepo  │  RoomRepo  │  MessageRepo       │         │
│  └──────────────────────────┬───────────────────────────────────┘         │
└─────────────────────────────┼───────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼                               ▼
     ┌────────────────┐              ┌────────────────┐
     │  PostgreSQL     │              │    Redis        │
     │  Users          │              │  rl:{ip}        │
     │  Profiles       │              │  msgrl:{userID} │
     │  Rooms          │              │                 │
     │  Messages       │              │  (fail-open)    │
     └────────────────┘              └────────────────┘
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend | Next.js 15, React 19, React Query, TypeScript |
| Backend | Go, Gin, GORM, Gorilla WebSocket |
| Database | PostgreSQL |
| Cache | Redis (rate limiting, fail-open) |
| Auth | Google OAuth 2.0, JWT (HTTP-only cookie) |
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
| `["queue-stats"]` | on-demand |
| `["match-stats"]` | on-demand |

## Queue Matching

```
User A → POST /queue/join → queues[category][gender]
User B → POST /queue/join → tryMatch() goroutine
                              ├── 1. opposite gender
                              ├── 2. same gender
                              └── 3. unknown gender
                             CreateRoom() → NotifyMatch()
                              ├── WS "match_found" → User A
                              └── WS "match_found" → User B
```

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
