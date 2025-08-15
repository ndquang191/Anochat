# Backend Development Plan - Anonymous Chat

## Phase 1: Foundation Setup

### Đã hoàn thành:

-   [x] Project structure
-   [x] Dependencies installation
-   [x] Basic main.go
-   [x] Environment configuration

#### 1.1 Database Connection (pkg/database/) ✅ HOÀN THÀNH

```go
// connection.go - Setup PostgreSQL connection pool ✅
// migration.go - Database schema creation ✅
// errors.go - Custom database errors ✅
```

#### 1.2 Config Management (pkg/config/) ✅ HOÀN THÀNH

```go
// config.go - Load all environment variables ✅
// chat_config.go - Categories, sensitive keywords ✅
```

## Phase 2: Core Models & Services ✅ HOÀN THÀNH

#### 2.1 Data Models (internal/model/) ✅ HOÀN THÀNH

```go
// models.go - User, Profile, Room, Message structs ✅
// jwt.go - JWT claims struct (cần thêm)
```

#### 2.2 Database Services (internal/service/) ✅ HOÀN THÀNH

```go
// user.go - CRUD operations cho users/profiles ✅
// room.go - Room management logic ✅
// message.go - Message storage/retrieval ✅
```

## Phase 3: Authentication ✅ HOÀN THÀNH (95%)

#### 3.1 Google OAuth (internal/service/auth.go) ✅

```go
// OAuth2 configuration ✅
// Token exchange logic ✅
// JWT generation/validation ✅
// Google user info retrieval ✅
// User creation/login logic ✅
```

#### 3.2 Auth Middleware (internal/middleware/auth.go) ✅

```go
// JWT validation middleware ✅
// User context injection ✅
// Bearer token support ✅
// Cookie token support ✅
```

#### 3.3 Auth Handlers (internal/handler/auth.go) ✅

```go
// GET /auth/google ✅
// GET /auth/callback ✅
// POST /auth/logout ✅ (bonus)
```

#### 3.4 Cần bổ sung (5%):

```go
// State parameter validation (security)
// Enhanced error handling (robustness)
// Token refresh endpoint (optional)
// Rate limiting cho auth endpoints (security)
```

## Phase 4: User Management ✅ HOÀN THÀNH

#### 4.1 User Services (internal/service/user.go) ✅ HOÀN THÀNH

```go
// Profile CRUD operations ✅
// Privacy settings logic ✅
// User state management ✅
// Active room detection ✅
```

#### 4.2 User Handlers (internal/handler/user.go) ✅ HOÀN THÀNH

```go
// GET /user/state - Current user info ✅
// PUT /profile - Update own profile ✅
// Note: GET /profile/:user_id removed (privacy concern)
```

## Phase 5: Matchmaking System ✅ HOÀN THÀNH

#### 5.1 Queue Service (internal/service/queue.go) ✅ HOÀN THÀNH

```go
// In-memory queue management ✅
// Matching algorithm (opposite gender + same category) ✅
// Queue timeout handling ✅
// Queue cleanup goroutine ✅
// Queue statistics ✅
```

#### 5.2 Room Service (internal/service/room.go) ✅ HOÀN THÀNH

```go
// Room creation when matched ✅
// Room cleanup logic ✅
// Sensitive keyword detection ✅
```

#### 5.3 Queue Handlers (internal/handler/queue.go) ✅ HOÀN THÀNH

```go
// POST /queue/join ✅
// DELETE /queue/leave ✅
// GET /queue/status ✅
// GET /queue/wait ✅ (long polling)
// GET /queue/stats ✅ (admin endpoint)
```

## Phase 6: Real-time Chat

#### 6.1 WebSocket Hub (internal/handler/websocket.go)

```go
// WebSocket connection management
// Message broadcasting
// Connection cleanup
```

#### 6.2 WebSocket Events

```go
// join_queue, leave_queue
// send_message, receive_message
// match_found, partner_left
```

#### 6.3 Message Handlers (internal/handler/message.go) ✅ HOÀN THÀNH

```go
// POST /message - Save message ✅
// PUT /message/seen - Mark as seen ✅
```

## Phase 7: Room Management

#### 7.1 Room Handlers (internal/handler/room.go) ✅ HOÀN THÀNH

```go
// POST /room - Create room (internal use) ✅
// GET /room - Get current room ✅
// PUT /room/leave - Leave room ✅
// GET /room/history/:id - Get sensitive room history ✅
```

#### 7.2 Post-Chat Cleanup

```go
// Analyze messages for sensitive keywords
// Delete non-sensitive rooms
// Retain sensitive rooms
```

## Phase 8: Security & Rate Limiting

#### 8.1 Rate Limiting (internal/middleware/)

```go
// ratelimit.go - General API rate limiting
// message_ratelimit.go - Message specific limits
```

#### 8.2 CORS & Security (internal/middleware/) ✅ HOÀN THÀNH (50%)

```go
// CORS configuration ✅ (implemented in main.go)
// Basic security headers ✅
// Note: Need dedicated middleware files for better organization
```

## Phase 9: Testing & Polish

#### 9.1 Integration Testing

```go
// Test complete flow: Auth → Queue → Match → Chat
// Test edge cases: timeout, disconnect, etc.
```

#### 9.2 API Documentation

```go
// Complete API spec with examples
// Error handling standardization
```

#### 9.3 Monitoring & Logging

```go
// Structured logging for all operations
// Health check improvements
// Performance monitoring
```

## Daily Checklist Template

```
□ Code implementation
□ Basic testing
□ Error handling
□ Logging integration
□ Git commit with clear message
□ Update documentation
```

## Priority Order

1. **CRITICAL:** Authentication + User management
2. **HIGH:** Matchmaking + WebSocket
3. **MEDIUM:** Room management + Cleanup
4. **LOW:** Rate limiting + Polish

## 📊 TỔNG KẾT TIẾN ĐỘ

### ✅ ĐÃ HOÀN THÀNH (60%):

-   **Phase 1: Foundation Setup** - 100% ✅
-   **Phase 2: Core Models & Services** - 100% ✅
-   **Phase 3: Authentication** - 95% ✅ (còn 5% security enhancements)
-   **Phase 4: User Management** - 100% ✅
-   **Phase 5: Matchmaking System** - 100% ✅
-   **Phase 6.3: Message Handlers** - 100% ✅
-   **Phase 7.1: Room Handlers** - 100% ✅
-   **Phase 8.2: CORS & Security** - 50% ✅

### 🚧 CẦN LÀM TIẾP (40%):

-   **Phase 6.1: WebSocket Hub** - 0% ❌
-   **Phase 6.2: WebSocket Events** - 0% ❌
-   **Phase 7.2: Post-Chat Cleanup** - 0% ❌
-   **Phase 8.1: Rate Limiting** - 0% ❌
-   **Phase 9: Testing & Polish** - 0% ❌

### 🎯 ƯU TIÊN TIẾP THEO:

1. **Phase 6.1: WebSocket Hub** (Real-time chat) - CRITICAL
2. **Phase 6.2: WebSocket Events** (Real-time events) - CRITICAL
3. **Phase 7.2: Post-Chat Cleanup** - HIGH
4. **Phase 8.1: Rate Limiting** - MEDIUM
5. **Phase 3.4: Auth Security Enhancements** (5% còn lại) - MEDIUM
