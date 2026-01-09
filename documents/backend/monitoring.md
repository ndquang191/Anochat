# Monitoring & Logging Guide - Anochat Backend

## Overview

This document outlines monitoring, logging, and observability practices for the Anochat backend system. It provides guidelines for production monitoring and debugging.

---

## Logging Strategy

### Current Implementation

**Logging Library:** `log/slog` (structured logging)
**Format:** JSON for production, console for development
**Level:** Info (production), Debug (development)

### Log Levels

| Level | Usage | Examples |
|-------|-------|----------|
| **Info** | Normal operations, important events | User login, room created, match found |
| **Warn** | Recoverable issues, rate limits | Rate limit exceeded, invalid input |
| **Error** | Errors requiring investigation | Database errors, WebSocket failures |
| **Debug** | Detailed debugging info | Request/response details, queue state |

---

## What to Log

### ✅ DO Log

#### Authentication Events
```go
slog.Info("User logged in", "user_id", userID, "email", email)
slog.Info("User logged out", "user_id", userID)
slog.Warn("Authentication failed", "reason", "invalid_token")
```

#### Queue & Matching Events
```go
slog.Info("User joined queue",
    "user_id", userID,
    "category", category,
    "position", position)

slog.Info("Match found",
    "user1_id", user1ID,
    "user2_id", user2ID,
    "category", category,
    "wait_time_ms", waitTime)

slog.Info("User left queue",
    "user_id", userID,
    "time_in_queue_ms", duration)
```

#### Room & Message Events
```go
slog.Info("Room created",
    "room_id", roomID,
    "user1_id", user1ID,
    "user2_id", user2ID)

slog.Info("Message sent",
    "user_id", userID,
    "room_id", roomID,
    "message_id", msgID)

slog.Info("Room ended",
    "room_id", roomID,
    "duration_seconds", duration,
    "message_count", msgCount)
```

#### WebSocket Events
```go
slog.Info("Client registered",
    "user_id", userID,
    "client_id", clientID)

slog.Info("Client disconnected",
    "user_id", userID,
    "reason", reason,
    "session_duration_seconds", duration)

slog.Warn("WebSocket error",
    "user_id", userID,
    "error", err)
```

#### Rate Limiting
```go
slog.Warn("Rate limit exceeded",
    "ip", clientIP,
    "endpoint", endpoint,
    "limit", rateLimit)

slog.Warn("Message rate limit exceeded",
    "user_id", userID,
    "limit", messageRateLimit)
```

### ❌ DO NOT Log

**Security Sensitive:**
- ❌ JWT tokens
- ❌ OAuth access tokens
- ❌ User passwords
- ❌ API secrets

**Privacy Sensitive:**
- ❌ Message content
- ❌ Email addresses in public logs
- ❌ IP addresses in public dashboards
- ❌ Full user profiles

---

## Monitoring Metrics

### Application Metrics

#### Connection Metrics
- **Active WebSocket Connections**
- **Connection Rate**
- **Disconnection Rate**
- **Average Session Duration**

#### Queue Metrics
- **Queue Size by Category**
- **Average Wait Time**
- **Match Rate**
- **Queue Abandonment Rate**

#### Message Metrics
- **Messages Per Second**
- **Message Latency**
- **Failed Messages**

#### API Metrics
- **Request Rate**
- **Response Time (P50, P95, P99)**
- **Error Rate**
- **Rate Limit Hits**

---

**Document Version:** 1.0
**Last Updated:** 2026-01-08
**Status:** Production Guidelines
