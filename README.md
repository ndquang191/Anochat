# Anochat

Anonymous chat application with real-time matchmaking, WebSocket messaging,
moderation, and horizontal backend coordination through Redis.

## Features

- Google OAuth 2.0 with short-lived access tokens and rotating refresh tokens
  stored in HTTP-only cookies
- Anonymous profiles with nickname, age, gender, and visibility controls
- Redis-backed matchmaking with atomic reservations and stale-state
  reconciliation
- PostgreSQL-enforced invariant preventing one user from joining two active
  rooms
- Distributed WebSocket delivery through Redis pub/sub
- Optimistic messages with server acknowledgements and cursor-based history
- Reporting, banned-word management, user suspension, review requests, and an
  administrator dashboard
- Vietnamese and English UI, responsive layout, and theme preferences

## Architecture

```
Browser
  ├── HTTPS / REST ────────┐
  └── Secure WebSocket ────┤
                           ▼
                 Next.js frontend
                           │
                           ▼
                  Go API (Gin)
             handler → service → repository
                    │
          ┌─────────┴──────────┐
          ▼                    ▼
       Redis               PostgreSQL
       queue               users and profiles
       reservations        rooms and messages
       rate limits         moderation evidence
       refresh sessions    room lifecycle records
       pub/sub
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend | Next.js 15, React 19, TanStack Query 5, TypeScript |
| Backend | Go, Gin, GORM, Gorilla WebSocket |
| Database | PostgreSQL |
| Coordination | Redis (queue, reservations, rate limits, refresh sessions, pub/sub) |
| Auth | Google OAuth 2.0, JWT access and rotating refresh cookies |
| Logging | Zap (via slog interface) |

## Data Flow

```
QueryClientProvider
  └── ErrorBoundary
       └── AuthProvider ← useUserState() ← GET /user/state (cached)
            └── AlertDialogProvider
                 ├── AppShellSidebar (reads the shared query)
                 ├── AppHeader       (account and navigation actions)
                 └── Page        (useQueue + useWebSocketChat)
```

The `["user-state"]` query has a 30-second `staleTime` and is shared by the auth
provider, sidebar, and chat page. Queue actions invalidate this query; match
notifications update the client through WebSocket.

## Queue Matching

The queue is stored in a Redis sorted set (`queue:waiting`, score = join timestamp ms).
Matching uses token-owned Redis reservations with a 30-second lease. Lua scripts
reserve and finalize a pair atomically across backend instances, while PostgreSQL
remains the source of truth for room creation.

```
User A → POST /queue/join
           └── Lua script: ZRANGE queue:waiting → no match → ZADD

User B → POST /queue/join
           └── Lua script: reserve User A + User B (User A remains in queue)
                  └── CreateRoom(A, B)
                       ├── success → commit reservation + ZREM User A
                       │              └── Publish "match_found" → Redis user:A, user:B
                       └── failure → release reservation; User A can be matched again
```

PostgreSQL also writes one `active_room_members` row per participant through a
room trigger in the same transaction. Its primary key on `user_id` is the final
guard against two active rooms for one user, including direct SQL writes. A
background reconciliation pass removes stale Redis queue/reservation entries
for users PostgreSQL already marks as active.

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
Users ── Profiles (1:1)
  ├── Rooms ── ActiveRoomMembers (unique user_id)
  │     └── Messages
  ├── Reports ── ReportMessages
  └── BannedWords (admin-managed)
```

`RoomSessions` preserve non-message lifecycle statistics independently of rooms.

## Getting Started

### Prerequisites

- Go 1.26.5
- Bun 1.3.14
- Docker with Compose v2 (recommended for local PostgreSQL and Redis)

### Quick Start

`start.sh` starts local PostgreSQL and Redis, creates development env files when
missing, applies migrations, and runs both applications:

```bash
./start.sh
```

Open `http://localhost:3000`. The example backend configuration enables the
development-only `Dev A` and `Dev B` logins, so Google OAuth credentials are not
required for local testing.

Stop the application processes with `Ctrl+C`, then stop local infrastructure:

```bash
./end.sh
```

### Manual Backend

```bash
docker compose up -d --wait postgres redis
cd api
cp .env.example .env    # edit with your credentials
go run ./cmd/migrate
go run ./cmd/server
```

Migrations are embedded, versioned SQL files under `api/pkg/database/migrations`.
The production Compose stack runs the one-shot `migrate` service successfully
before starting the API.

### Manual Frontend

```bash
cd frontend
cp .env.example .env.local
bun install
bun run dev
```

### Environment Variables

See `api/.env.example` and `frontend/.env.example` for the complete schemas.
The backend loads `api/.env` for local development; deployed environments should
inject the same variables through their secret/configuration manager.

`SERVER_ENV` must be `development`, `test`, or `production`. Existing process
environment variables take precedence over `api/.env`. Next.js exposes
`NEXT_PUBLIC_*` values at build time, so production frontend values must be set
before `bun run build`.

### Docker Compose

`docker-compose.yml` runs only local PostgreSQL and Redis and binds their ports
to localhost:

```bash
docker compose up -d --wait postgres redis
```

Production uses external PostgreSQL and Redis services:

```bash
cp .env.production.example .env.production
docker compose --env-file .env.production -f docker-compose.prod.yml up -d --build
```

The production Compose file builds and deploys the backend only. Deploy the
Next.js frontend separately with production `NEXT_PUBLIC_API_URL`,
`NEXT_PUBLIC_SITE_URL`, and `NEXT_PUBLIC_DEV_AUTH_ENABLED=false`.

The one-shot `migrate` service must complete before the API starts. Confirm the
deployment through the public reverse proxy:

```bash
curl https://api.example.com/healthz
```

The endpoint returns `503` if PostgreSQL or Redis is unavailable.

### Reset the Local Database

Database reset is available only as a local CLI command. Set
`SERVER_ENV=development` and `ALLOW_DATABASE_RESET=true` in `api/.env`, then run:

```bash
./reset-db.sh
```

There is no HTTP endpoint for resetting the database.

### Grant Administrator Access

Administrator access is stored in the database. After the user has signed in at
least once, grant the role explicitly:

```sql
UPDATE users SET is_admin = true WHERE email = 'admin@example.com';
```

Reload the application after changing the role.

## Quality Checks

```bash
# Backend
cd api
go test ./...
go vet ./...
go build -o /tmp/anochat-server ./cmd/server
go build -o /tmp/anochat-migrate ./cmd/migrate

# Frontend
cd frontend
bun install --frozen-lockfile
bun run test
bun run lint
bun run build
```

GitHub Actions runs these checks for every push and pull request. It also
validates both Compose files and builds the API Docker image. The repository
currently provides CI only; deployment to a VPS is still a manual operation.

## Documentation

General, code-adjacent documentation remains in [`documents/`](documents/):

- REST and WebSocket API reference
- Database schema
- Backend and frontend development rules
- Monitoring and testing guides

## Scaling

Each backend instance owns its local WebSocket connections. Redis coordinates
the queue, reservations, rate limits, refresh sessions, and cross-instance
notifications; PostgreSQL remains the source of truth for rooms, messages, and
moderation data. Multiple API instances can therefore run behind a load
balancer while sharing PostgreSQL and Redis.
