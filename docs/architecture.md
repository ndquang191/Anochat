# Anochat Architecture

## System Overview

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

---

## Frontend Data Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                    QueryClientProvider                           │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │                    ErrorBoundary                           │  │
│  │  ┌─────────────────────────────────────────────────────┐  │  │
│  │  │                  AuthProvider                        │  │  │
│  │  │                                                      │  │  │
│  │  │  useUserState() ─── queryKey: ["user-state"]        │  │  │
│  │  │       │              queryFn: GET /user/state        │  │  │
│  │  │       ▼                                              │  │  │
│  │  │  { user, room, messages, loading }                   │  │  │
│  │  │       │                                              │  │  │
│  │  │  ┌────▼─────────────────────────────────────────┐   │  │  │
│  │  │  │            AlertDialogProvider                │   │  │  │
│  │  │  │                                               │   │  │  │
│  │  │  │  ┌──────────┐ ┌──────────┐ ┌──────────────┐ │   │  │  │
│  │  │  │  │  Sidebar  │ │  Header  │ │   Page       │ │   │  │  │
│  │  │  │  │useUserSt()│ │ActionBtn │ │  useQueue()  │ │   │  │  │
│  │  │  │  │(cached!)  │ │invalidate│ │  useWsChat() │ │   │  │  │
│  │  │  │  └──────────┘ └──────────┘ └──────────────┘ │   │  │  │
│  │  │  └───────────────────────────────────────────────┘   │  │  │
│  │  └──────────────────────────────────────────────────────┘  │  │
│  └────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

---

## React Query Cache Strategy

```
                    ┌─────────────────────────┐
                    │    React Query Cache     │
                    │                          │
                    │  ["user-state"]           │
                    │    staleTime: 30s         │
                    │    source: GET /user/state│
                    │    consumers:             │
                    │      - AuthProvider       │
                    │      - AppSidebar         │
                    │                          │
                    │  ["queue-status"]         │
                    │    refetchInterval: 5s    │
                    │    (when isInQueue)       │
                    │                          │
                    │  ["queue-stats"]          │
                    │  ["match-stats"]          │
                    └────────────┬─────────────┘
                                 │
              ┌──────────────────┼──────────────────┐
              │                  │                   │
     invalidateQueries()  invalidateQueries()  removeQueries()
              │                  │                   │
     ┌────────▼──────┐  ┌───────▼───────┐  ┌───────▼───────┐
     │ WS events:    │  │ User actions: │  │ logout()      │
     │ match_found   │  │ leaveRoom     │  │ clears cache  │
     │ partner_left  │  │ updateProfile │  │               │
     │ room_left     │  │ joinQueue     │  │               │
     └───────────────┘  └───────────────┘  └───────────────┘
```

---

## Authentication Flow

```
  Browser                  Backend                Google
    │                        │                      │
    │  GET /auth/google      │                      │
    ├───────────────────────►│                      │
    │  302 → google.com      │                      │
    │◄───────────────────────┤                      │
    │                        │                      │
    │  User consents ────────────────────────────►  │
    │                        │                      │
    │  GET /auth/callback?code=xxx                  │
    ├───────────────────────►│  Exchange code       │
    │                        ├─────────────────────►│
    │                        │◄─ access_token ──────┤
    │                        │                      │
    │                        │  Fetch user info     │
    │                        ├─────────────────────►│
    │                        │◄─ email,name,pic ────┤
    │                        │                      │
    │                        │  GetOrCreateUser()   │
    │                        │  generateJWT()       │
    │                        │                      │
    │  Set-Cookie: jwt_token │                      │
    │  Set-Cookie: temp_data │                      │
    │  302 → /callback       │                      │
    │◄───────────────────────┤                      │
    │                        │                      │
    │  /callback page:       │                      │
    │  read temp_data cookie │                      │
    │  login(user)           │                      │
    │  invalidate query      │                      │
    │  redirect → /          │                      │
    │                        │                      │
    │  GET /user/state       │                      │
    │  Cookie: jwt_token     │                      │
    ├───────────────────────►│  Validate JWT        │
    │◄───────────────────────┤  Return user state   │
```

---

## Queue Matching & WebSocket Flow

```
  User A (Frontend)         Backend                    User B (Frontend)
    │                         │                           │
    │ POST /queue/join        │                           │
    ├────────────────────────►│                           │
    │                         │  Add to queue             │
    │                         │  queues[cat][gender]      │
    │                         │                           │
    │                         │         POST /queue/join  │
    │                         │◄──────────────────────────┤
    │                         │  Add to queue             │
    │                         │  tryMatch() goroutine     │
    │                         │                           │
    │                         │  ┌─────────────────────┐  │
    │                         │  │ findMatch():        │  │
    │                         │  │ 1. opposite gender  │  │
    │                         │  │ 2. same gender      │  │
    │                         │  │ 3. unknown gender   │  │
    │                         │  └─────────┬───────────┘  │
    │                         │            │              │
    │                         │  CreateRoom()             │
    │                         │  NotifyMatch()            │
    │                         │            │              │
    │  WS: "match_found"      │            │              │
    │◄────────────────────────┤────────────┘              │
    │                         ├──────────────────────────►│
    │                         │     WS: "match_found"     │
    │                         │                           │
    │  invalidateUserState()  │                           │
    │  WS: "join_room"        │                           │
    ├────────────────────────►│                           │
    │                         │                           │
    │  WS: "send_message"     │                           │
    ├────────────────────────►│                           │
    │                         │  Redis rate limit check   │
    │                         │  Save to DB               │
    │                         │  WS: "receive_message"    │
    │                         ├──────────────────────────►│
    │                         │                           │
    │  WS: "typing"           │                           │
    ├────────────────────────►│  WS: "partner_typing"     │
    │                         ├──────────────────────────►│
```

---

## Backend Dependency Graph

```
main.go
  ├── config.Load()
  ├── zap.NewDevelopment() / zap.NewProduction()
  │     └── slog.SetDefault(zapslog handler)
  ├── database.InitDatabase(cfg)
  │     └── gorm.Open(postgres)
  ├── cache.InitRedis(cfg)
  │     └── redis.NewClient()
  │
  ├── Repositories
  │   ├── UserRepository(db)
  │   ├── ProfileRepository(db)
  │   ├── RoomRepository(db)
  │   └── MessageRepository(db)
  │
  ├── Services
  │   ├── UserService(userRepo, profileRepo)
  │   ├── RoomService(roomRepo, messageRepo)
  │   ├── MessageService(messageRepo, roomRepo)
  │   ├── AuthService(userService, oauth, jwt, roomRepo, msgRepo)
  │   └── QueueService(roomService, userService, roomRepo, cfg)
  │         └── SetMatchNotifier(hub)
  │
  ├── WebSocket
  │   └── Hub(queueService, messageService, roomService, redisClient)
  │         ├── Client.ReadPump → handleMessage()
  │         ├── Client.WritePump → ping/pong
  │         └── Run() goroutine (register/unregister/broadcast)
  │
  ├── Handlers
  │   ├── AuthHandler(authService, oauth, cfg)
  │   ├── UserHandler(userService, roomService, roomRepo, msgRepo, cfg)
  │   ├── QueueHandler(queueService, cfg)
  │   └── WebSocketHandler(hub, authService, cfg)
  │
  └── Middleware
      ├── CORSMiddleware(clientURL)
      ├── RateLimitMiddleware(redisClient, rate, burst)
      └── AuthMiddleware(authService, cfg)
```

---

## Database Schema

```
┌──────────────┐       ┌──────────────┐
│    Users     │       │   Profiles   │
├──────────────┤       ├──────────────┤
│ id       UUID├──┐    │ id       UUID│
│ email  STRING│  │    │ user_id  UUID├──► Users.id
│ name   STRING│  │    │ age       INT│
│ avatar STRING│  │    │ city   STRING│
│ is_active    │  │    │ is_male  BOOL│
│ is_deleted   │  │    │ is_hidden    │
│ created_at   │  │    │ created_at   │
└──────────────┘  │    └──────────────┘
                  │
┌──────────────┐  │    ┌──────────────┐
│    Rooms     │  │    │   Messages   │
├──────────────┤  │    ├──────────────┤
│ id       UUID│  │    │ id       UUID│
│ user1_id UUID├──┤    │ room_id  UUID├──► Rooms.id
│ user2_id UUID├──┘    │ sender_id    ├──► Users.id
│ category     │       │ content      │
│ is_sensitive │       │ created_at   │
│ ended_at     │       └──────────────┘
│ created_at   │
└──────────────┘
```

---

## Redis Keys

| Key Pattern | Purpose | TTL | Fail Behavior |
|---|---|---|---|
| `rl:{ip}` | HTTP rate limit counter | 1s window | Open (allow) |
| `msgrl:{userID}` | Message rate limit (10/sec) | 1s window | Open (allow) |
