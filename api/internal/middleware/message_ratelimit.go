package middleware

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MessageRateLimiter manages rate limiting for chat messages per user
type MessageRateLimiter struct {
	users   map[uuid.UUID]*MessageVisitor
	mu      sync.RWMutex
	rate    int           // messages per second
	burst   int           // maximum burst size
	cleanup time.Duration // cleanup interval
}

// MessageVisitor represents a user with message rate limiting
type MessageVisitor struct {
	limiter  *TokenBucket
	lastSeen time.Time
}

// NewMessageRateLimiter creates a new message rate limiter
func NewMessageRateLimiter(messagesPerSecond, burst int) *MessageRateLimiter {
	mrl := &MessageRateLimiter{
		users:   make(map[uuid.UUID]*MessageVisitor),
		rate:    messagesPerSecond,
		burst:   burst,
		cleanup: 10 * time.Minute,
	}

	// Start cleanup goroutine
	go mrl.cleanupStaleUsers()

	return mrl
}

// getUser retrieves or creates a message visitor for a user
func (mrl *MessageRateLimiter) getUser(userID uuid.UUID) *MessageVisitor {
	mrl.mu.Lock()
	defer mrl.mu.Unlock()

	v, exists := mrl.users[userID]
	if !exists {
		limiter := NewTokenBucket(float64(mrl.burst), float64(mrl.rate))
		v = &MessageVisitor{
			limiter:  limiter,
			lastSeen: time.Now(),
		}
		mrl.users[userID] = v
	} else {
		v.lastSeen = time.Now()
	}

	return v
}

// cleanupStaleUsers removes inactive users periodically
func (mrl *MessageRateLimiter) cleanupStaleUsers() {
	ticker := time.NewTicker(mrl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		mrl.mu.Lock()
		for userID, v := range mrl.users {
			if time.Since(v.lastSeen) > mrl.cleanup {
				delete(mrl.users, userID)
			}
		}
		mrl.mu.Unlock()
	}
}

// MessageRateLimitMiddleware creates a message rate limiting middleware
// This should be applied to message-sending endpoints
func MessageRateLimitMiddleware(messagesPerSecond, burst int) gin.HandlerFunc {
	limiter := NewMessageRateLimiter(messagesPerSecond, burst)

	return func(c *gin.Context) {
		// Get user ID from context (set by auth middleware)
		userIDValue, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not authenticated",
			})
			c.Abort()
			return
		}

		userID, ok := userIDValue.(uuid.UUID)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid user ID format",
			})
			c.Abort()
			return
		}

		visitor := limiter.getUser(userID)

		if !visitor.limiter.Allow() {
			slog.Warn("Message rate limit exceeded", "user_id", userID)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Message rate limit exceeded",
				"message": "You are sending messages too quickly. Please slow down.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
